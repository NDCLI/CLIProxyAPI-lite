package management

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"golang.org/x/crypto/bcrypt"
)

func newLocalSettingsTestHandler(t *testing.T) (*Handler, string, string) {
	t.Helper()
	root := t.TempDir()
	authDir := filepath.Join(root, "auths")
	if err := os.MkdirAll(authDir, 0o700); err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("123456"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	configPath := filepath.Join(root, "config.yaml")
	content := fmt.Sprintf("host: 127.0.0.1\nport: 8317\nremote-management:\n  secret-key: %q\nauth-dir: %q\n", hash, authDir)
	if err = os.WriteFile(configPath, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg, err := config.LoadConfig(configPath)
	if err != nil {
		t.Fatal(err)
	}
	return NewHandler(cfg, configPath, nil), configPath, authDir
}

func localSettingsRequest(t *testing.T, handler gin.HandlerFunc, method, path string, body []byte, remoteAddr string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(method, path, strings.NewReader(string(body)))
	ctx.Request.RemoteAddr = remoteAddr
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler(ctx)
	return recorder
}

func TestChangeManagementPasswordPersistsHash(t *testing.T) {
	h, configPath, _ := newLocalSettingsTestHandler(t)
	bad := localSettingsRequest(t, h.PatchManagementPassword, http.MethodPatch, "/management-password", []byte(`{"current_password":"wrong","new_password":"new-secret"}`), "127.0.0.1:1234")
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("wrong current password status = %d", bad.Code)
	}
	good := localSettingsRequest(t, h.PatchManagementPassword, http.MethodPatch, "/management-password", []byte(`{"current_password":"123456","new_password":"new-secret"}`), "127.0.0.1:1234")
	if good.Code != http.StatusOK {
		t.Fatalf("change password status = %d: %s", good.Code, good.Body.String())
	}
	data, err := os.ReadFile(configPath)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(data), "new-secret") || bcrypt.CompareHashAndPassword([]byte(h.cfg.RemoteManagement.SecretKey), []byte("new-secret")) != nil {
		t.Fatal("new password was not persisted as a valid hash")
	}
	if allowed, _, _ := h.AuthenticateManagementKey("127.0.0.1", true, "123456"); allowed {
		t.Fatal("old password remained valid")
	}
	if allowed, _, _ := h.AuthenticateManagementKey("127.0.0.1", true, "new-secret"); !allowed {
		t.Fatal("new password was not accepted")
	}
}

func TestLocalBackupRestoresSettingsAndAccounts(t *testing.T) {
	h, configPath, authDir := newLocalSettingsTestHandler(t)
	root := filepath.Dir(configPath)
	for path, data := range map[string]string{
		filepath.Join(root, ".env"):                        "SAMPLE_LOCAL_SETTING=enabled\n",
		filepath.Join(root, "combos.json"):                 `[{"name":"saved"}]`,
		filepath.Join(root, "proxy-pools.json"):            `[]`,
		filepath.Join(root, "certs", "mitm-mappings.json"): `{}`,
		filepath.Join(authDir, "provider-account.json"):    `{"type":"antigravity","access_token":"saved-token"}`,
	} {
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(data), 0o600); err != nil {
			t.Fatal(err)
		}
	}
	exported := localSettingsRequest(t, h.ExportLocalBackup, http.MethodGet, "/local-backup", nil, "127.0.0.1:1234")
	if exported.Code != http.StatusOK || exported.Header().Get("Cache-Control") != "no-store" {
		t.Fatalf("export status = %d", exported.Code)
	}
	var backup localBackup
	if err := json.Unmarshal(exported.Body.Bytes(), &backup); err != nil {
		t.Fatal(err)
	}
	if string(backup.Files["auth/provider-account.json"]) != `{"type":"antigravity","access_token":"saved-token"}` {
		t.Fatal("signed-in provider account was missing from export")
	}
	if string(backup.Files[".env"]) != "SAMPLE_LOCAL_SETTING=enabled\n" {
		t.Fatal("local environment settings were missing from export")
	}
	backup.ConfigYAML = strings.Replace(backup.ConfigYAML, fmt.Sprintf("auth-dir: %q", authDir), `auth-dir: "C:\\old-machine\\auths"`, 1)
	differentHash, err := bcrypt.GenerateFromPassword([]byte("backup-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	backup.ConfigYAML = regexp.MustCompile(`(?m)^  secret-key:.*$`).ReplaceAllStringFunc(backup.ConfigYAML, func(string) string { return `  secret-key: "` + string(differentHash) + `"` })
	importBody, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(root, "combos.json"), []byte(`[{"name":"changed"}]`), 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(authDir, "extra.json"), []byte(`{"type":"codex"}`), 0o600); err != nil {
		t.Fatal(err)
	}
	imported := localSettingsRequest(t, h.ImportLocalBackup, http.MethodPost, "/local-backup", importBody, "127.0.0.1:1234")
	if imported.Code != http.StatusOK {
		t.Fatalf("import status = %d: %s", imported.Code, imported.Body.String())
	}
	combos, err := os.ReadFile(filepath.Join(root, "combos.json"))
	if err != nil || string(combos) != `[{"name":"saved"}]` {
		t.Fatalf("combos were not restored: %v", err)
	}
	if _, err := os.Stat(filepath.Join(authDir, "extra.json")); !os.IsNotExist(err) {
		t.Fatal("account absent from backup was not removed")
	}
	if h.cfg.AuthDir != authDir {
		t.Fatalf("auth directory = %q, want current machine directory %q", h.cfg.AuthDir, authDir)
	}
	if bcrypt.CompareHashAndPassword([]byte(h.cfg.RemoteManagement.SecretKey), []byte("123456")) != nil {
		t.Fatal("import changed the current management password without opt-in")
	}
	var result struct {
		SafetyBackup string `json:"safety_backup"`
	}
	if err := json.Unmarshal(imported.Body.Bytes(), &result); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(result.SafetyBackup); err != nil {
		t.Fatalf("pre-import safety backup missing: %v", err)
	}
	safetyData, err := os.ReadFile(result.SafetyBackup)
	if err != nil {
		t.Fatal(err)
	}
	var safety localBackup
	if err = json.Unmarshal(safetyData, &safety); err != nil {
		t.Fatal(err)
	}
	if _, preserved := safety.Files["auth/extra.json"]; !preserved {
		t.Fatal("safety backup did not preserve the account removed during import")
	}
}

func TestLocalBackupRequiresKnownPasswordBeforeRestoringIt(t *testing.T) {
	h, configPath, authDir := newLocalSettingsTestHandler(t)
	backup, err := collectLocalBackup(configPath, authDir)
	if err != nil {
		t.Fatal(err)
	}
	hash, err := bcrypt.GenerateFromPassword([]byte("backup-password"), bcrypt.MinCost)
	if err != nil {
		t.Fatal(err)
	}
	backup.ConfigYAML = regexp.MustCompile(`(?m)^  secret-key:.*$`).ReplaceAllStringFunc(backup.ConfigYAML, func(string) string { return `  secret-key: "` + string(hash) + `"` })
	backup.RestorePassword = true
	backup.BackupPassword = "wrong"
	badBody, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	bad := localSettingsRequest(t, h.ImportLocalBackup, http.MethodPost, "/local-backup", badBody, "127.0.0.1:1234")
	if bad.Code != http.StatusBadRequest {
		t.Fatalf("wrong backup password status = %d", bad.Code)
	}
	if bcrypt.CompareHashAndPassword([]byte(h.cfg.RemoteManagement.SecretKey), []byte("123456")) != nil {
		t.Fatal("wrong backup password changed the current password")
	}
	backup.BackupPassword = "backup-password"
	goodBody, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	good := localSettingsRequest(t, h.ImportLocalBackup, http.MethodPost, "/local-backup", goodBody, "127.0.0.1:1234")
	if good.Code != http.StatusOK {
		t.Fatalf("restore backup password status = %d: %s", good.Code, good.Body.String())
	}
	if bcrypt.CompareHashAndPassword([]byte(h.cfg.RemoteManagement.SecretKey), []byte("backup-password")) != nil {
		t.Fatal("backup password was not restored")
	}
}

func TestLocalBackupRejectsUnsafeFilesAndRemoteClients(t *testing.T) {
	h, _, _ := newLocalSettingsTestHandler(t)
	remote := localSettingsRequest(t, h.ExportLocalBackup, http.MethodGet, "/local-backup", nil, "198.51.100.5:1234")
	if remote.Code != http.StatusForbidden {
		t.Fatalf("remote export status = %d", remote.Code)
	}
	backup, err := collectLocalBackup(h.configFilePath, h.cfg.AuthDir)
	if err != nil {
		t.Fatal(err)
	}
	backup.Files["auth/../escape.json"] = []byte(`{"type":"codex"}`)
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	unsafe := localSettingsRequest(t, h.ImportLocalBackup, http.MethodPost, "/local-backup", data, "127.0.0.1:1234")
	if unsafe.Code != http.StatusBadRequest {
		t.Fatalf("unsafe import status = %d: %s", unsafe.Code, unsafe.Body.String())
	}
}

func TestLocalBackupImportWaitsForMITMDNS(t *testing.T) {
	h, configPath, authDir := newLocalSettingsTestHandler(t)
	backup, err := collectLocalBackup(configPath, authDir)
	if err != nil {
		t.Fatal(err)
	}
	data, err := json.Marshal(backup)
	if err != nil {
		t.Fatal(err)
	}
	h.mitm.dnsMu.Lock()
	h.mitm.dnsDesired["antigravity"] = true
	h.mitm.dnsMu.Unlock()
	result := localSettingsRequest(t, h.ImportLocalBackup, http.MethodPost, "/local-backup", data, "127.0.0.1:1234")
	if result.Code != http.StatusConflict {
		t.Fatalf("import with MITM DNS enabled status = %d", result.Code)
	}
}

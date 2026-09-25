package management

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"golang.org/x/crypto/bcrypt"
)

const (
	localBackupFormat  = "cliproxy-local-backup"
	localBackupVersion = 1
	maxLocalBackupSize = 32 << 20
)

var errBackupPasswordIncorrect = errors.New("backup password is incorrect")

var localBackupFiles = []string{
	".env",
	"combos.json",
	"proxy-pools.json",
	"certs/mitm-mappings.json",
	"certs/ca.crt",
	"certs/ca.key",
}

type localBackup struct {
	Format          string            `json:"format"`
	Version         int               `json:"version"`
	ExportedAt      time.Time         `json:"exported_at"`
	ConfigYAML      string            `json:"config_yaml"`
	Files           map[string][]byte `json:"files"`
	RestorePassword bool              `json:"restore_password,omitempty"`
	BackupPassword  string            `json:"backup_password,omitempty"`
}

func requireLocalManagement(c *gin.Context) bool {
	host, _, err := net.SplitHostPort(c.Request.RemoteAddr)
	if err == nil {
		if ip := net.ParseIP(host); ip != nil && ip.IsLoopback() {
			return true
		}
	}
	c.JSON(http.StatusForbidden, gin.H{"error": gin.H{"code": "local_only", "message": "This operation is available only from this computer"}})
	return false
}

func (h *Handler) ExportLocalBackup(c *gin.Context) {
	if !requireLocalManagement(c) {
		return
	}
	h.mu.Lock()
	if h.cfg == nil || h.configFilePath == "" || h.cfg.AuthDir == "" {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "local configuration is unavailable"})
		return
	}
	backup, err := collectLocalBackup(h.configFilePath, h.cfg.AuthDir)
	h.mu.Unlock()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to read local settings"})
		return
	}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to encode backup"})
		return
	}
	if len(data) > maxLocalBackupSize {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"error": "backup is larger than 32 MiB"})
		return
	}
	c.Header("Cache-Control", "no-store")
	c.Header("X-Content-Type-Options", "nosniff")
	c.Header("Content-Disposition", fmt.Sprintf("attachment; filename=\"cliproxy-backup-%s.json\"", time.Now().UTC().Format("20060102-150405")))
	c.Data(http.StatusOK, "application/json; charset=utf-8", data)
}

func collectLocalBackup(configPath, authDir string) (localBackup, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return localBackup{}, fmt.Errorf("read config: %w", err)
	}
	backup := localBackup{
		Format: localBackupFormat, Version: localBackupVersion, ExportedAt: time.Now().UTC(),
		ConfigYAML: string(data), Files: make(map[string][]byte),
	}
	baseDir := filepath.Dir(configPath)
	for _, name := range localBackupFiles {
		path := filepath.Join(baseDir, filepath.FromSlash(name))
		info, errStat := os.Lstat(path)
		if errors.Is(errStat, os.ErrNotExist) {
			continue
		}
		if errStat != nil || info.Mode()&os.ModeSymlink != 0 {
			return localBackup{}, fmt.Errorf("cannot safely read %s", name)
		}
		content, errRead := os.ReadFile(path)
		if errRead != nil {
			return localBackup{}, fmt.Errorf("read %s: %w", name, errRead)
		}
		backup.Files[name] = content
	}
	entries, err := os.ReadDir(authDir)
	if errors.Is(err, os.ErrNotExist) {
		return backup, nil
	}
	if err != nil {
		return localBackup{}, fmt.Errorf("read auth directory: %w", err)
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}
		if isUnsafeBackupAuthName(name) || entry.Type()&os.ModeSymlink != 0 {
			return localBackup{}, fmt.Errorf("invalid auth file name")
		}
		content, errRead := os.ReadFile(filepath.Join(authDir, name))
		if errRead != nil {
			return localBackup{}, fmt.Errorf("read auth file: %w", errRead)
		}
		backup.Files["auth/"+name] = content
	}
	return backup, nil
}

func (h *Handler) ImportLocalBackup(c *gin.Context) {
	if !requireLocalManagement(c) {
		return
	}
	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxLocalBackupSize)
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()
	var backup localBackup
	if err := decoder.Decode(&backup); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_backup", "message": "Invalid backup JSON or file is too large"}})
		return
	}
	var extra any
	if decoder.Decode(&extra) != io.EOF {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "invalid_backup", "message": "Backup must contain one JSON object"}})
		return
	}
	if backup.Format != localBackupFormat || backup.Version != localBackupVersion {
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": "unsupported_backup", "message": "Unsupported backup format or version"}})
		return
	}

	h.mu.Lock()
	if h.cfg == nil || h.configFilePath == "" || h.cfg.AuthDir == "" {
		h.mu.Unlock()
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "local configuration is unavailable"})
		return
	}
	if backup.RestorePassword && (h.envSecret != "" || h.localPassword != "") {
		h.mu.Unlock()
		c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "password_override", "message": "Remove the runtime password override before restoring a backup password"}})
		return
	}
	if h.mitm != nil {
		h.mitm.mu.RLock()
		running := h.mitm.server != nil
		h.mitm.mu.RUnlock()
		h.mitm.dnsMu.RLock()
		dnsActive := len(h.mitm.dnsDesired) > 0
		h.mitm.dnsMu.RUnlock()
		if running || dnsActive {
			h.mu.Unlock()
			c.JSON(http.StatusConflict, gin.H{"error": gin.H{"code": "mitm_running", "message": "Stop the MITM server and disable its DNS before importing settings"}})
			return
		}
	}
	configPath, authDir := h.configFilePath, h.cfg.AuthDir
	current, errCurrent := collectLocalBackup(configPath, authDir)
	if errCurrent != nil {
		h.mu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to capture current settings"})
		return
	}
	if _, hasCA := backup.Files["certs/ca.crt"]; !hasCA {
		if cert, existingCA := current.Files["certs/ca.crt"]; existingCA {
			if backup.Files == nil {
				backup.Files = make(map[string][]byte)
			}
			backup.Files["certs/ca.crt"] = cert
			backup.Files["certs/ca.key"] = current.Files["certs/ca.key"]
		}
	}
	validatedConfig, newCfg, errValidate := validateLocalBackup(backup, configPath, authDir, h.cfg.RemoteManagement.SecretKey)
	if errValidate != nil {
		h.mu.Unlock()
		code := "invalid_backup"
		if errors.Is(errValidate, errBackupPasswordIncorrect) {
			code = "backup_password_wrong"
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": gin.H{"code": code, "message": errValidate.Error()}})
		return
	}
	safetyPath, errSafety := savePreImportBackup(configPath, current)
	if errSafety != nil {
		h.mu.Unlock()
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create safety backup"})
		return
	}
	if errApply := applyLocalBackup(configPath, authDir, backup.Files, validatedConfig); errApply != nil {
		errRollback := applyLocalBackup(configPath, authDir, current.Files, []byte(current.ConfigYAML))
		h.mu.Unlock()
		message := "Import failed; current settings were restored"
		if errRollback != nil {
			message = "Import and automatic rollback failed; use the safety backup to recover"
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": gin.H{"code": "import_failed", "message": message}, "safety_backup": safetyPath})
		return
	}
	h.cfg = newCfg
	snapshot := h.reloadSnapshotConfigLocked()
	h.mu.Unlock()
	h.reloadConfigAfterManagementSaveAsync(c.Request.Context(), snapshot)
	c.JSON(http.StatusOK, gin.H{"ok": true, "account_count": countBackupAuthFiles(backup.Files), "safety_backup": safetyPath, "restart_required": true})
}

func validateLocalBackup(backup localBackup, configPath, authDir, currentPasswordHash string) ([]byte, *config.Config, error) {
	if strings.TrimSpace(backup.ConfigYAML) == "" {
		return nil, nil, fmt.Errorf("backup has no config.yaml")
	}
	seenTargets := make(map[string]bool, len(backup.Files))
	for name, data := range backup.Files {
		path, err := localBackupTarget(configPath, authDir, name)
		if err != nil {
			return nil, nil, err
		}
		folded := strings.ToLower(filepath.Clean(path))
		if seenTargets[folded] {
			return nil, nil, fmt.Errorf("backup contains conflicting file names")
		}
		seenTargets[folded] = true
		if strings.HasSuffix(strings.ToLower(name), ".json") && !json.Valid(data) {
			return nil, nil, fmt.Errorf("backup contains invalid JSON in %s", name)
		}
	}
	cert, hasCert := backup.Files["certs/ca.crt"]
	key, hasKey := backup.Files["certs/ca.key"]
	if hasCert != hasKey {
		return nil, nil, fmt.Errorf("MITM certificate and key must both be present")
	}
	if hasCert {
		if _, err := tls.X509KeyPair(cert, key); err != nil {
			return nil, nil, fmt.Errorf("invalid MITM certificate and key")
		}
	}
	tmp, err := os.CreateTemp(filepath.Dir(configPath), ".backup-validate-*.yaml")
	if err != nil {
		return nil, nil, fmt.Errorf("cannot validate config")
	}
	tmpPath := tmp.Name()
	defer os.Remove(tmpPath)
	if _, err = tmp.WriteString(backup.ConfigYAML); err != nil {
		_ = tmp.Close()
		return nil, nil, fmt.Errorf("cannot validate config")
	}
	if err = tmp.Close(); err != nil {
		return nil, nil, fmt.Errorf("cannot validate config")
	}
	if err = config.SaveConfigPreserveCommentsUpdateNestedScalar(tmpPath, []string{"auth-dir"}, authDir); err != nil {
		return nil, nil, fmt.Errorf("invalid config.yaml")
	}
	if !backup.RestorePassword {
		if err = config.SaveConfigPreserveCommentsUpdateNestedScalar(tmpPath, []string{"remote-management", "secret-key"}, currentPasswordHash); err != nil {
			return nil, nil, fmt.Errorf("cannot preserve current management password")
		}
	}
	newCfg, err := config.LoadConfig(tmpPath)
	if err != nil || newCfg.RemoteManagement.SecretKey == "" {
		return nil, nil, fmt.Errorf("invalid config.yaml or missing management password")
	}
	if _, err = bcrypt.Cost([]byte(newCfg.RemoteManagement.SecretKey)); err != nil {
		return nil, nil, fmt.Errorf("invalid management password hash")
	}
	if backup.RestorePassword && bcrypt.CompareHashAndPassword([]byte(newCfg.RemoteManagement.SecretKey), []byte(backup.BackupPassword)) != nil {
		return nil, nil, errBackupPasswordIncorrect
	}
	validated, err := os.ReadFile(tmpPath)
	if err != nil {
		return nil, nil, fmt.Errorf("cannot read validated config")
	}
	return validated, newCfg, nil
}

func localBackupTarget(configPath, authDir, name string) (string, error) {
	for _, allowed := range localBackupFiles {
		if name == allowed {
			return filepath.Join(filepath.Dir(configPath), filepath.FromSlash(name)), nil
		}
	}
	if strings.HasPrefix(name, "auth/") {
		file := strings.TrimPrefix(name, "auth/")
		if !isUnsafeBackupAuthName(file) && strings.HasSuffix(strings.ToLower(file), ".json") {
			return filepath.Join(authDir, file), nil
		}
	}
	return "", fmt.Errorf("backup contains unsupported file name")
}

func isUnsafeBackupAuthName(name string) bool {
	return isUnsafeAuthFileName(name) || !filepath.IsLocal(name) || filepath.Base(name) != name || len(name) > 255
}

func savePreImportBackup(configPath string, backup localBackup) (string, error) {
	dir := filepath.Join(filepath.Dir(configPath), "backups")
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", err
	}
	data, err := json.MarshalIndent(backup, "", "  ")
	if err != nil {
		return "", err
	}
	f, err := os.CreateTemp(dir, "before-import-*.json")
	if err != nil {
		return "", err
	}
	path := f.Name()
	if _, err = f.Write(data); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		_ = os.Remove(path)
		return "", err
	}
	if err = f.Close(); err != nil {
		_ = os.Remove(path)
		return "", err
	}
	return path, nil
}

func applyLocalBackup(configPath, authDir string, files map[string][]byte, configYAML []byte) error {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	sort.Strings(names)
	for _, name := range names {
		path, err := localBackupTarget(configPath, authDir, name)
		if err != nil {
			return err
		}
		if err = writeLocalBackupFile(path, files[name]); err != nil {
			return err
		}
	}
	for _, name := range localBackupFiles {
		if _, present := files[name]; present {
			continue
		}
		path, _ := localBackupTarget(configPath, authDir, name)
		if err := os.Remove(path); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	entries, err := os.ReadDir(authDir)
	if err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	for _, entry := range entries {
		name := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(strings.ToLower(name), ".json") {
			continue
		}
		if _, present := files["auth/"+name]; present {
			continue
		}
		if err := os.Remove(filepath.Join(authDir, name)); err != nil && !errors.Is(err, os.ErrNotExist) {
			return err
		}
	}
	return writeLocalBackupFile(configPath, configYAML)
}

func writeLocalBackupFile(path string, data []byte) error {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return err
	}
	f, err := os.CreateTemp(dir, ".restore-*")
	if err != nil {
		return err
	}
	defer os.Remove(f.Name())
	if _, err = io.Copy(f, bytes.NewReader(data)); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		_ = f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	return os.Rename(f.Name(), path)
}

func countBackupAuthFiles(files map[string][]byte) int {
	count := 0
	for name := range files {
		if strings.HasPrefix(name, "auth/") {
			count++
		}
	}
	return count
}

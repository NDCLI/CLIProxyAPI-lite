package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func TestImport9RouterAccountsConvertsSupportedProviders(t *testing.T) {
	gin.SetMode(gin.TestMode)
	authDir := t.TempDir()
	h := NewHandlerWithoutConfigFilePath(&config.Config{AuthDir: authDir}, coreauth.NewManager(nil, nil, nil))
	payload := `{"providerConnections":[
		{"id":"codex-1","provider":"codex","authType":"oauth","email":"user@example.com","accessToken":"codex-secret","refreshToken":"refresh-secret","providerSpecificData":{"chatgptAccountId":"acct-1"}},
		{"id":"ag-1","provider":"antigravity","accessToken":"ag-secret","refreshToken":"ag-refresh","projectId":"project-1"},
		{"id":"ignored","provider":"github","accessToken":"must-not-import"}
	]}`

	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	request := httptest.NewRequest(http.MethodPost, "/v0/management/import/9router", strings.NewReader(payload))
	request.Header.Set("Content-Type", "application/json")
	ctx.Request = request
	h.Import9RouterAccounts(ctx)

	if recorder.Code != http.StatusMultiStatus {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusMultiStatus, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "secret") {
		t.Fatalf("response leaked imported token: %s", recorder.Body.String())
	}
	var response struct {
		Imported int   `json:"imported"`
		Failed   []any `json:"failed"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &response); errDecode != nil {
		t.Fatalf("decode response: %v", errDecode)
	}
	if response.Imported != 2 || len(response.Failed) != 1 {
		t.Fatalf("unexpected result: imported=%d failed=%d", response.Imported, len(response.Failed))
	}

	codexData, errRead := os.ReadFile(filepath.Join(authDir, "9router-codex-codex-1.json"))
	if errRead != nil {
		t.Fatalf("read imported Codex account: %v", errRead)
	}
	if !strings.Contains(string(codexData), `"account_id": "acct-1"`) || !strings.Contains(string(codexData), `"access_token": "codex-secret"`) {
		t.Fatalf("Codex account was not converted correctly: %s", codexData)
	}
	antigravityData, errRead := os.ReadFile(filepath.Join(authDir, "9router-antigravity-ag-1.json"))
	if errRead != nil {
		t.Fatalf("read imported Antigravity account: %v", errRead)
	}
	if !strings.Contains(string(antigravityData), `"project_id": "project-1"`) {
		t.Fatalf("Antigravity account was not converted correctly: %s", antigravityData)
	}
}

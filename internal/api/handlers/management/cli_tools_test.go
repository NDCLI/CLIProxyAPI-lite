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
)

func TestCLIToolsReportConfiguredWithoutSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	path := filepath.Join(home, ".claude", "settings.json")
	if errWrite := os.MkdirAll(filepath.Dir(path), 0o700); errWrite != nil {
		t.Fatal(errWrite)
	}
	if errWrite := os.WriteFile(path, []byte(`{"env":{"ANTHROPIC_BASE_URL":"http://127.0.0.1:8317/v1","ANTHROPIC_AUTH_TOKEN":"secret"}}`), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/cli-tools", nil)
	(&Handler{}).GetCLITools(ctx)
	if recorder.Code != http.StatusOK || string(recorder.Body.Bytes()) == "" {
		t.Fatalf("response = %d %s", recorder.Code, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "secret") || strings.Contains(recorder.Body.String(), path) {
		t.Fatalf("response leaked sensitive data: %s", recorder.Body.String())
	}
	var response struct {
		Items []cliToolItem `json:"items"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &response); errDecode != nil {
		t.Fatal(errDecode)
	}
	if len(response.Items) == 0 || !response.Items[0].Configured {
		t.Fatalf("items = %#v", response.Items)
	}
	seen := make(map[string]struct{}, len(response.Items))
	for _, item := range response.Items {
		if _, exists := seen[item.ID]; exists {
			t.Fatalf("duplicate CLI tool id %q", item.ID)
		}
		seen[item.ID] = struct{}{}
	}
}

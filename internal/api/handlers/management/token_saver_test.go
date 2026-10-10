package management

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/tokensaver"
)

func TestTokenSaverSettingsPersistAndApplyToLiveService(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("host: 127.0.0.1\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	handler := NewHandler(&config.Config{Host: "127.0.0.1"}, path, nil)
	service := tokensaver.NewService(config.TokenSaverConfig{})
	handler.SetTokenSaver(service)
	call := func(method, route, payload string, action gin.HandlerFunc) *httptest.ResponseRecorder {
		t.Helper()
		recorder := httptest.NewRecorder()
		ctx, _ := gin.CreateTestContext(recorder)
		ctx.Request = httptest.NewRequest(method, route, bytes.NewBufferString(payload))
		ctx.Request.Header.Set("Content-Type", "application/json")
		action(ctx)
		return recorder
	}

	updated := call(http.MethodPatch, "/v0/management/token-saver", `{"rtk_enabled":true}`, handler.PatchTokenSaver)
	if updated.Code != http.StatusOK || !service.Snapshot().Config.RTKEnabled {
		t.Fatalf("settings did not apply live: %d %s", updated.Code, updated.Body.String())
	}
	saved, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(saved), "rtk-enabled: true") {
		t.Fatalf("settings not persisted: %v %s", err, saved)
	}
	got := call(http.MethodGet, "/v0/management/token-saver", "", handler.GetTokenSaver)
	if got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"rtk_enabled":true`) {
		t.Fatalf("GET settings = %d %s", got.Code, got.Body.String())
	}
}

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
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/health" {
			t.Errorf("unexpected Headroom path %s", r.URL.Path)
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	}))
	defer upstream.Close()

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

	updated := call(http.MethodPatch, "/v0/management/token-saver", `{"rtk_enabled":true,"headroom_enabled":true,"headroom_url":"`+upstream.URL+`","headroom_timeout_ms":1500}`, handler.PatchTokenSaver)
	if updated.Code != http.StatusOK || !service.Snapshot().Config.RTKEnabled || !service.Snapshot().Config.HeadroomEnabled {
		t.Fatalf("settings did not apply live: %d %s", updated.Code, updated.Body.String())
	}
	saved, err := os.ReadFile(path)
	if err != nil || !strings.Contains(string(saved), "rtk-enabled: true") || !strings.Contains(string(saved), upstream.URL) {
		t.Fatalf("settings not persisted: %v %s", err, saved)
	}
	got := call(http.MethodGet, "/v0/management/token-saver", "", handler.GetTokenSaver)
	if got.Code != http.StatusOK || !strings.Contains(got.Body.String(), `"rtk_enabled":true`) {
		t.Fatalf("GET settings = %d %s", got.Code, got.Body.String())
	}
	health := call(http.MethodPost, "/v0/management/token-saver/headroom/test", "{}", handler.TestHeadroom)
	if health.Code != http.StatusOK || !strings.Contains(health.Body.String(), `"ok":true`) {
		t.Fatalf("Headroom test = %d %s", health.Code, health.Body.String())
	}
	invalid := call(http.MethodPatch, "/v0/management/token-saver", `{"headroom_url":"file:///etc/passwd"}`, handler.PatchTokenSaver)
	if invalid.Code != http.StatusBadRequest || service.Snapshot().Config.HeadroomURL != upstream.URL {
		t.Fatalf("invalid settings changed service: %d %s", invalid.Code, invalid.Body.String())
	}
}

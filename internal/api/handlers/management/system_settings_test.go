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
)

func TestSystemSettingsAreSecretFreeAndPersisted(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if errWrite := os.WriteFile(path, []byte("host: 127.0.0.1\n"), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	handler := NewHandler(&config.Config{Host: "127.0.0.1", Port: 8317, SDKConfig: config.SDKConfig{APIKeys: []string{"secret-key"}}}, path, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/system-settings", nil)
	handler.GetSystemSettings(ctx)
	if recorder.Code != http.StatusOK || strings.Contains(recorder.Body.String(), "secret-key") {
		t.Fatalf("settings response = %d %s", recorder.Code, recorder.Body.String())
	}
	recorder = httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/system-settings", bytes.NewBufferString(`{"logging_to_file":true,"request_retry":2}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PatchSystemSettings(ctx)
	if recorder.Code != http.StatusOK || !handler.cfg.LoggingToFile || handler.cfg.RequestRetry != 2 {
		t.Fatalf("patch response = %d %s", recorder.Code, recorder.Body.String())
	}
}

func TestSystemSettingsRejectNegativeRetries(t *testing.T) {
	gin.SetMode(gin.TestMode)
	path := filepath.Join(t.TempDir(), "config.yaml")
	if errWrite := os.WriteFile(path, []byte("host: 127.0.0.1\n"), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	handler := NewHandler(&config.Config{}, path, nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/system-settings", bytes.NewBufferString(`{"request_retry":-1}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PatchSystemSettings(ctx)
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusBadRequest)
	}
}

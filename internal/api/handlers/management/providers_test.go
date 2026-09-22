package management

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
)

func TestProvidersAreNormalizedAndSecretFree(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := coreauth.NewManager(nil, nil, nil)
	if _, errRegister := manager.Register(context.Background(), &coreauth.Auth{
		ID: "claude-credential", Provider: "claude", Label: "Work Claude", Status: coreauth.StatusActive,
		Attributes: map[string]string{coreauth.AttributeAPIKey: "secret-provider-key"},
	}); errRegister != nil {
		t.Fatal(errRegister)
	}
	if _, errRegister := manager.Register(context.Background(), &coreauth.Auth{
		ID: "codex-credential", Provider: "codex", FileName: "C:/private/auth/codex.json", Disabled: true, Status: coreauth.StatusDisabled,
		Metadata: map[string]any{"access_token": "secret-oauth-token", "email": "user@example.com"},
	}); errRegister != nil {
		t.Fatal(errRegister)
	}
	handler := NewHandlerWithoutConfigFilePath(&config.Config{}, manager)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/providers", nil)
	handler.GetProviders(ctx)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	if strings.Contains(recorder.Body.String(), "secret-") || strings.Contains(recorder.Body.String(), "user@example.com") || strings.Contains(recorder.Body.String(), "C:/private") {
		t.Fatalf("provider response leaked credential material: %s", recorder.Body.String())
	}
	var response struct {
		SchemaVersion int            `json:"schema_version"`
		Items         []providerItem `json:"items"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &response); errDecode != nil {
		t.Fatal(errDecode)
	}
	if response.SchemaVersion != 1 || len(response.Items) != 2 {
		t.Fatalf("unexpected response: %#v", response)
	}
	if response.Items[0].ID != "claude-credential" || response.Items[0].AuthType != coreauth.AuthKindAPIKey || !response.Items[0].Enabled {
		t.Fatalf("unexpected API-key provider: %#v", response.Items[0])
	}
	if response.Items[1].Label != "codex.json" || response.Items[1].Enabled || response.Items[1].Status != string(coreauth.StatusDisabled) {
		t.Fatalf("unexpected OAuth provider: %#v", response.Items[1])
	}
}

func TestProviderRoutesRequireManagementAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := &Handler{cfg: &config.Config{RemoteManagement: config.RemoteManagement{SecretKey: "configured"}}, failedAttempts: make(map[string]*attemptInfo), localPassword: "test-management-key"}
	router := gin.New()
	management := router.Group("/v0/management")
	management.Use(handler.Middleware())
	management.GET("/providers", handler.GetProviders)

	request := httptest.NewRequest(http.MethodGet, "/v0/management/providers", bytes.NewBuffer(nil))
	request.RemoteAddr = "127.0.0.1:12345"
	unauthorized := httptest.NewRecorder()
	router.ServeHTTP(unauthorized, request)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}
}

func TestPatchProviderChangesEnabledState(t *testing.T) {
	gin.SetMode(gin.TestMode)
	manager := coreauth.NewManager(nil, nil, nil)
	if _, errRegister := manager.Register(context.Background(), &coreauth.Auth{ID: "provider-1", Provider: "claude", Status: coreauth.StatusActive}); errRegister != nil {
		t.Fatal(errRegister)
	}
	handler := NewHandlerWithoutConfigFilePath(&config.Config{}, manager)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Params = gin.Params{{Key: "id", Value: "provider-1"}}
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/providers/provider-1", strings.NewReader(`{"enabled":false}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PatchProvider(ctx)
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d: %s", recorder.Code, http.StatusOK, recorder.Body.String())
	}
	updated, ok := manager.GetByID("provider-1")
	if !ok || !updated.Disabled || updated.Status != coreauth.StatusDisabled {
		t.Fatalf("provider state = %#v, want disabled", updated)
	}
}

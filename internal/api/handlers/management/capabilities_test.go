package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestGetCapabilitiesContract(t *testing.T) {
	gin.SetMode(gin.TestMode)

	recorder := httptest.NewRecorder()
	ginContext, _ := gin.CreateTestContext(recorder)
	ginContext.Request = httptest.NewRequest(http.MethodGet, "/v0/management/capabilities", nil)

	(&Handler{}).GetCapabilities(ginContext)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var response managementCapabilitiesResponse
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &response); errDecode != nil {
		t.Fatalf("decode response: %v", errDecode)
	}
	if response.SchemaVersion != managementCapabilitiesSchemaVersion {
		t.Fatalf("schema_version = %d, want %d", response.SchemaVersion, managementCapabilitiesSchemaVersion)
	}

	wantStates := map[string]string{
		"endpoint_keys":   "ready",
		"providers":       "ready",
		"combos":          "ready",
		"usage":           "ready",
		"quota":           "ready",
		"token_saver":     "ready",
		"cli_tools":       "ready",
		"logs":            "ready",
		"system_settings": "ready",
		"system_info":     "ready",
		"proxy_pools":     "partial",
	}
	if len(response.Capabilities) != len(wantStates) {
		t.Fatalf("capabilities count = %d, want %d", len(response.Capabilities), len(wantStates))
	}

	seen := make(map[string]bool, len(response.Capabilities))
	for _, capability := range response.Capabilities {
		if seen[capability.ID] {
			t.Fatalf("duplicate capability %q", capability.ID)
		}
		seen[capability.ID] = true
		wantState, exists := wantStates[capability.ID]
		if !exists {
			t.Fatalf("unexpected capability %q", capability.ID)
		}
		if capability.State != wantState {
			t.Errorf("capability %q state = %q, want %q", capability.ID, capability.State, wantState)
		}
		if capability.State != "ready" && capability.ReasonCode == "" {
			t.Errorf("capability %q has state %q without reason_code", capability.ID, capability.State)
		}
	}
}

func TestGetCapabilitiesUsesManagementAuthentication(t *testing.T) {
	gin.SetMode(gin.TestMode)

	handler := &Handler{
		cfg:            &config.Config{RemoteManagement: config.RemoteManagement{SecretKey: "configured"}},
		failedAttempts: make(map[string]*attemptInfo),
		localPassword:  "test-management-key",
	}
	router := gin.New()
	management := router.Group("/v0/management")
	management.Use(handler.Middleware())
	management.GET("/capabilities", handler.GetCapabilities)

	unauthorized := httptest.NewRecorder()
	unauthorizedRequest := httptest.NewRequest(http.MethodGet, "/v0/management/capabilities", nil)
	unauthorizedRequest.RemoteAddr = "127.0.0.1:12345"
	router.ServeHTTP(unauthorized, unauthorizedRequest)
	if unauthorized.Code != http.StatusUnauthorized {
		t.Fatalf("unauthorized status = %d, want %d", unauthorized.Code, http.StatusUnauthorized)
	}

	authorized := httptest.NewRecorder()
	authorizedRequest := httptest.NewRequest(http.MethodGet, "/v0/management/capabilities", nil)
	authorizedRequest.RemoteAddr = "127.0.0.1:12345"
	authorizedRequest.Header.Set("Authorization", "Bearer test-management-key")
	router.ServeHTTP(authorized, authorizedRequest)
	if authorized.Code != http.StatusOK {
		t.Fatalf("authorized status = %d, want %d body=%s", authorized.Code, http.StatusOK, authorized.Body.String())
	}
}

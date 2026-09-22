package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
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
		"endpoint_keys":   "partial",
		"providers":       "partial",
		"combos":          "planned",
		"usage":           "partial",
		"quota":           "ready",
		"token_saver":     "unavailable",
		"cli_tools":       "partial",
		"logs":            "ready",
		"system_settings": "ready",
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

package management

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestDiscoverCompatibleProviderModels(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer test-key" {
			t.Errorf("unexpected provider request: path=%q authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"xpiki-model"},{"id":"bad\nmodel"}]}`))
	}))
	defer upstream.Close()
	gin.SetMode(gin.TestMode)
	body := []byte(`{"base_url":"` + upstream.URL + `/v1","api_key":"test-key"}`)
	request := httptest.NewRequest(http.MethodPost, "/provider-configs/discover-models", bytes.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	(&Handler{}).DiscoverCompatibleProviderModels(ctx)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"xpiki-model"`) || strings.Contains(response.Body.String(), "bad") {
		t.Fatalf("unexpected model discovery: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTestCompatibleProviderAPIKeyUsesSelectedSavedKey(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer stored-secret" {
			t.Errorf("unexpected provider request: path=%q authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"data":[{"id":"model-a"},{"id":"model-b"}]}`))
	}))
	defer upstream.Close()
	gin.SetMode(gin.TestMode)
	providerID := compatibleProviderID("test-provider")
	body := `{"id":"` + providerID + `","base_url":"` + upstream.URL + `/v1","api_key_index":1}`
	request := httptest.NewRequest(http.MethodPost, "/provider-configs/test-api-key", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	h := &Handler{cfg: &config.Config{OpenAICompatibility: []config.OpenAICompatibility{{
		Name: "test-provider", BaseURL: upstream.URL + "/v1",
		APIKeyEntries: []config.OpenAICompatibilityAPIKey{{APIKey: "first-secret"}, {APIKey: "stored-secret"}},
	}}}}
	h.TestCompatibleProviderAPIKey(ctx)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ok":true`) || !strings.Contains(response.Body.String(), `"model_count":2`) {
		t.Fatalf("unexpected API key test response: status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "stored-secret") {
		t.Fatal("API key test response exposed the saved key")
	}
}

func TestTestCompatibleProviderAPIKeyDoesNotExposeInvalidKey(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte(`{"error":"bad secret"}`))
	}))
	defer upstream.Close()
	gin.SetMode(gin.TestMode)
	body := `{"base_url":"` + upstream.URL + `/v1","api_key":"invalid-secret"}`
	request := httptest.NewRequest(http.MethodPost, "/provider-configs/test-api-key", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	(&Handler{}).TestCompatibleProviderAPIKey(ctx)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ok":false`) {
		t.Fatalf("unexpected failed API key test response: status=%d body=%s", response.Code, response.Body.String())
	}
	if strings.Contains(response.Body.String(), "invalid-secret") || strings.Contains(response.Body.String(), "bad secret") {
		t.Fatal("API key test response exposed credentials or upstream error data")
	}
}

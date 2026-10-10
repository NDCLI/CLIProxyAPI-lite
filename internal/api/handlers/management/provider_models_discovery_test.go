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

func TestTestCompatibleProviderModelUsesSavedKeyAndValidatesResponse(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/chat/completions" || r.Header.Get("Authorization") != "Bearer model-secret" {
			t.Errorf("unexpected model request: path=%q authorization=%q", r.URL.Path, r.Header.Get("Authorization"))
		}
		_, _ = w.Write([]byte(`{"choices":[{"message":{"content":"OK"}}]}`))
	}))
	defer upstream.Close()
	gin.SetMode(gin.TestMode)
	body := `{"model":"model-a","api_key_index":1}`
	request := httptest.NewRequest(http.MethodPost, "/provider-configs/"+compatibleProviderID("model-provider")+"/test-model", strings.NewReader(body))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	ctx.Params = gin.Params{{Key: "id", Value: compatibleProviderID("model-provider")}}
	h := &Handler{cfg: &config.Config{OpenAICompatibility: []config.OpenAICompatibility{{Name: "model-provider", BaseURL: upstream.URL + "/v1", Models: []config.OpenAICompatibilityModel{{Name: "model-a", Alias: "public-model"}}, APIKeyEntries: []config.OpenAICompatibilityAPIKey{{APIKey: "first-secret"}, {APIKey: "model-secret"}}}}}}
	h.TestCompatibleProviderModel(ctx)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), `"ok":true`) || strings.Contains(response.Body.String(), "model-secret") {
		t.Fatalf("unexpected model test response: status=%d body=%s", response.Code, response.Body.String())
	}
}

func TestTestCompatibleProviderModelRejectsEmptyChoices(t *testing.T) {
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer upstream.Close()
	gin.SetMode(gin.TestMode)
	request := httptest.NewRequest(http.MethodPost, "/provider-configs/"+compatibleProviderID("empty-provider")+"/test-model", strings.NewReader(`{"model":"model-a","api_key_index":0}`))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = request
	ctx.Params = gin.Params{{Key: "id", Value: compatibleProviderID("empty-provider")}}
	h := &Handler{cfg: &config.Config{OpenAICompatibility: []config.OpenAICompatibility{{Name: "empty-provider", BaseURL: upstream.URL, Models: []config.OpenAICompatibilityModel{{Name: "model-a", Alias: "model-a"}}, APIKeyEntries: []config.OpenAICompatibilityAPIKey{{APIKey: "secret"}}}}}}
	h.TestCompatibleProviderModel(ctx)
	if response.Code != http.StatusOK || strings.Contains(response.Body.String(), `"ok":true`) {
		t.Fatalf("empty choices accepted: status=%d body=%s", response.Code, response.Body.String())
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

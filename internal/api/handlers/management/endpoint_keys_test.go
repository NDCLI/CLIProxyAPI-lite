package management

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func newEndpointKeyTestHandler(t *testing.T, keys ...string) *Handler {
	t.Helper()
	configPath := filepath.Join(t.TempDir(), "config.yaml")
	if errWrite := os.WriteFile(configPath, []byte("api-keys: []\n"), 0o600); errWrite != nil {
		t.Fatalf("write config: %v", errWrite)
	}
	return NewHandler(&config.Config{SDKConfig: config.SDKConfig{APIKeys: append([]string(nil), keys...)}}, configPath, nil)
}

func endpointKeyRequest(t *testing.T, handler gin.HandlerFunc, method, path, body, id string) *httptest.ResponseRecorder {
	t.Helper()
	recorder := httptest.NewRecorder()
	context, _ := gin.CreateTestContext(recorder)
	context.Request = httptest.NewRequest(method, path, bytes.NewBufferString(body))
	context.Request.Header.Set("Content-Type", "application/json")
	if id != "" {
		context.Params = gin.Params{{Key: "id", Value: id}}
	}
	handler(context)
	return recorder
}

func TestEndpointKeysNeverExposeStoredSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	const secret = "sk-super-secret-value"
	handler := newEndpointKeyTestHandler(t, secret)
	recorder := endpointKeyRequest(t, handler.GetEndpointKeys, http.MethodGet, "/v0/management/endpoint-keys", "", "")
	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}
	if strings.Contains(recorder.Body.String(), secret) {
		t.Fatal("endpoint key response exposed the stored secret")
	}
	var response struct {
		SchemaVersion int               `json:"schema_version"`
		Items         []endpointKeyItem `json:"items"`
	}
	if errDecode := json.Unmarshal(recorder.Body.Bytes(), &response); errDecode != nil {
		t.Fatalf("decode response: %v", errDecode)
	}
	if response.SchemaVersion != 1 || len(response.Items) != 1 || response.Items[0].Mask != "••••alue" {
		t.Fatalf("unexpected response: %#v", response)
	}
}

func TestEndpointKeyCreateRotateDelete(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := newEndpointKeyTestHandler(t)

	created := endpointKeyRequest(t, handler.PostEndpointKey, http.MethodPost, "/v0/management/endpoint-keys", `{}`, "")
	if created.Code != http.StatusCreated {
		t.Fatalf("create status = %d, want %d body=%s", created.Code, http.StatusCreated, created.Body.String())
	}
	var createResponse struct {
		Secret string          `json:"secret"`
		Item   endpointKeyItem `json:"item"`
	}
	if errDecode := json.Unmarshal(created.Body.Bytes(), &createResponse); errDecode != nil {
		t.Fatalf("decode create response: %v", errDecode)
	}
	if !strings.HasPrefix(createResponse.Secret, "sk-") || createResponse.Item.ID == "" {
		t.Fatalf("unexpected create response: %#v", createResponse)
	}
	if len(handler.cfg.APIKeys) != 1 || handler.cfg.APIKeys[0] != createResponse.Secret {
		t.Fatal("generated endpoint key was not persisted")
	}

	rotated := endpointKeyRequest(t, handler.PatchEndpointKey, http.MethodPatch, "/v0/management/endpoint-keys/"+createResponse.Item.ID, `{"revision":"`+createResponse.Item.Revision+`"}`, createResponse.Item.ID)
	if rotated.Code != http.StatusOK {
		t.Fatalf("rotate status = %d, want %d body=%s", rotated.Code, http.StatusOK, rotated.Body.String())
	}
	var rotateResponse struct {
		Secret string          `json:"secret"`
		Item   endpointKeyItem `json:"item"`
	}
	if errDecode := json.Unmarshal(rotated.Body.Bytes(), &rotateResponse); errDecode != nil {
		t.Fatalf("decode rotate response: %v", errDecode)
	}
	if rotateResponse.Secret == createResponse.Secret || rotateResponse.Item.ID == createResponse.Item.ID {
		t.Fatal("rotate did not replace the endpoint key")
	}

	deleted := endpointKeyRequest(t, handler.DeleteEndpointKey, http.MethodDelete, "/v0/management/endpoint-keys/"+rotateResponse.Item.ID+"?revision="+rotateResponse.Item.Revision, "", rotateResponse.Item.ID)
	if deleted.Code != http.StatusNoContent {
		t.Fatalf("delete status = %d, want %d body=%s", deleted.Code, http.StatusNoContent, deleted.Body.String())
	}
	if len(handler.cfg.APIKeys) != 0 {
		t.Fatal("endpoint key was not deleted")
	}
}

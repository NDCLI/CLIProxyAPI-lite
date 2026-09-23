package management

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
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

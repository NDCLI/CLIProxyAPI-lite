package management

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestComboCRUDPersistsValidatedTargets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&config.Config{}, t.TempDir()+"\\config.yaml", nil)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v0/management/combos", strings.NewReader(`{"name":"Fast","model":"fast","targets":[{"provider":"codex","model":"gpt-5"},{"provider":"claude","model":"sonnet"}]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PutCombo(ctx)
	if recorder.Code != http.StatusCreated || !strings.Contains(recorder.Body.String(), "claude") {
		t.Fatalf("create = %d %s", recorder.Code, recorder.Body.String())
	}
	list := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(list)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/combos", nil)
	handler.GetCombos(ctx)
	if list.Code != http.StatusOK || !strings.Contains(list.Body.String(), "gpt-5") {
		t.Fatalf("list = %d %s", list.Code, list.Body.String())
	}
}

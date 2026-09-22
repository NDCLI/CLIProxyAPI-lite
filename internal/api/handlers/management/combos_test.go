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

func TestComboPatchAndValidate(t *testing.T) {
	gin.SetMode(gin.TestMode)
	handler := NewHandler(&config.Config{}, t.TempDir()+"\\config.yaml", nil)
	created := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(created)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v0/management/combos", strings.NewReader(`{"id":"fast","name":"Fast","model":"fast","enabled":false,"targets":[{"provider":"codex","model":"gpt-5"}]}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PutCombo(ctx)
	if created.Code != http.StatusCreated || !strings.Contains(created.Body.String(), `"enabled":false`) {
		t.Fatalf("create = %d %s", created.Code, created.Body.String())
	}
	patched := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(patched)
	ctx.Params = gin.Params{{Key: "id", Value: "fast"}}
	ctx.Request = httptest.NewRequest(http.MethodPatch, "/v0/management/combos/fast", strings.NewReader(`{"enabled":false,"vision":true}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.PatchCombo(ctx)
	if patched.Code != http.StatusOK || !strings.Contains(patched.Body.String(), `"vision":true`) {
		t.Fatalf("patch = %d %s", patched.Code, patched.Body.String())
	}
	validated := httptest.NewRecorder()
	ctx, _ = gin.CreateTestContext(validated)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/v0/management/combos/validate", strings.NewReader(`{"name":"invalid","model":"invalid"}`))
	ctx.Request.Header.Set("Content-Type", "application/json")
	handler.ValidateCombo(ctx)
	if validated.Code != http.StatusBadRequest {
		t.Fatalf("validate = %d %s", validated.Code, validated.Body.String())
	}
}

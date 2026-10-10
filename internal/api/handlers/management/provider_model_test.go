package management

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
	sdktranslator "github.com/router-for-me/CLIProxyAPI/v7/sdk/translator"
	"github.com/tidwall/gjson"
)

type modelProbeExecutor struct {
	refreshRecordExecutor
	called bool
	empty  bool
}

func (e *modelProbeExecutor) Execute(_ context.Context, auth *coreauth.Auth, req coreexecutor.Request, opts coreexecutor.Options) (coreexecutor.Response, error) {
	e.called = true
	if opts.SourceFormat != sdktranslator.FormatOpenAI || gjson.GetBytes(opts.OriginalRequest, "messages.0.content").String() != "Reply with OK." || auth.ID != "model-probe-auth" {
		return coreexecutor.Response{}, fmt.Errorf("incorrect model probe protocol or credential")
	}
	if e.empty {
		return coreexecutor.Response{Payload: []byte(`{"choices":[]}`)}, nil
	}
	return coreexecutor.Response{Payload: []byte(`{"choices":[{"message":{"content":"OK"}}]}`)}, nil
}

func TestProviderModelSuppliesTranslationSourceAndPinnedCredential(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	executor := &modelProbeExecutor{refreshRecordExecutor: refreshRecordExecutor{provider: "codex"}}
	manager.RegisterExecutor(executor)
	provider := executor.Identifier()
	auth := &coreauth.Auth{ID: "model-probe-auth", Provider: provider, Status: coreauth.StatusActive}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	registry.GetGlobalRegistry().RegisterClient(auth.ID, provider, []*registry.ModelInfo{{ID: "test-probe-model"}})
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	h := NewHandlerWithoutConfigFilePath(&config.Config{}, manager)
	response := endpointKeyRequest(t, h.TestProviderModel, http.MethodPost, "/providers/model-probe-auth/test-model", `{"model":"test-probe-model"}`, auth.ID)
	if !executor.called || response.Code != http.StatusOK || !gjson.GetBytes(response.Body.Bytes(), "ok").Bool() {
		t.Fatalf("probe failed: %s", response.Body.String())
	}
}

func TestProviderModelRejectsEmptyChoices(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	executor := &modelProbeExecutor{refreshRecordExecutor: refreshRecordExecutor{provider: "codex"}, empty: true}
	manager.RegisterExecutor(executor)
	provider := executor.Identifier()
	auth := &coreauth.Auth{ID: "model-probe-auth", Provider: provider, Status: coreauth.StatusActive}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	registry.GetGlobalRegistry().RegisterClient(auth.ID, provider, []*registry.ModelInfo{{ID: "test-probe-model"}})
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	h := NewHandlerWithoutConfigFilePath(&config.Config{}, manager)
	response := endpointKeyRequest(t, h.TestProviderModel, http.MethodPost, "/providers/model-probe-auth/test-model", `{"model":"test-probe-model"}`, auth.ID)
	if !executor.called || response.Code != http.StatusOK || gjson.GetBytes(response.Body.Bytes(), "ok").Bool() {
		t.Fatalf("probe incorrectly accepted empty response: %s", response.Body.String())
	}
}

func TestTestProviderModelRequiresExecutionManager(t *testing.T) {
	gin.SetMode(gin.TestMode)
	recorder := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(recorder)
	ctx.Request = httptest.NewRequest(http.MethodPost, "/providers/provider-1/test-model", nil)
	ctx.Params = gin.Params{{Key: "id", Value: "provider-1"}}

	(&Handler{}).TestProviderModel(ctx)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

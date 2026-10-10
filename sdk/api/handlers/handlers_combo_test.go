package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	coreexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

type comboResolverTest struct{}

func (comboResolverTest) ResolveCombo(model string) ([]ComboTarget, bool) {
	if model != "fast" {
		return nil, false
	}
	return []ComboTarget{{Provider: "codex", Model: "gpt-5"}, {Provider: "claude", Model: "sonnet"}}, true
}

func (comboResolverTest) ListComboModels() []ComboModel {
	return []ComboModel{{ID: "fast", Name: "Fast"}}
}

type orderedComboResolver struct{ targets []ComboTarget }

func (r orderedComboResolver) ResolveCombo(model string) ([]ComboTarget, bool) {
	return append([]ComboTarget(nil), r.targets...), model == "fast"
}

func TestComboResolutionUsesOrderedProvidersAndProviderModel(t *testing.T) {
	handler := NewBaseAPIHandlers(nil, nil)
	handler.SetComboResolver(comboResolverTest{})
	providers, model, errMsg := handler.providersForExecution("fast", "fast", false, modelRouteDecision{}, modelExecutionOptions{})
	if errMsg != nil || model != "gpt-5" || len(providers) != 2 || providers[1] != "claude" {
		t.Fatalf("resolution = %#v %q %v", providers, model, errMsg)
	}
	var opts coreexecutor.Options
	opts.Metadata = map[string]any{}
	handler.installComboInterceptor(&opts, "fast")
	response := opts.RequestAfterAuthInterceptor(context.Background(), coreexecutor.RequestAfterAuthInterceptRequest{Metadata: map[string]any{coreexecutor.ComboProviderMetadataKey: "claude"}, Model: "gpt-5"})
	if response.Model != "sonnet" {
		t.Fatalf("provider model = %q, want sonnet", response.Model)
	}
}

func TestComboModelAppearsInAvailableModels(t *testing.T) {
	handler := NewBaseAPIHandlers(nil, nil)
	handler.SetComboResolver(comboResolverTest{})
	models := handler.AppendComboModels([]map[string]any{{"id": "existing", "object": "model"}}, "openai")
	if len(models) != 2 || models[1]["id"] != "fast" || models[1]["display_name"] != "Fast" {
		t.Fatalf("combo model list = %#v", models)
	}
}

func TestComboExecutesPrimaryBeforeFallback(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	manager.SetRetryConfig(0, 0, 1)
	var calls []string
	primary := &modelExecutionCaptureExecutor{provider: "combo-primary", execute: func(_ context.Context, _ *coreauth.Auth, req coreexecutor.Request, _ coreexecutor.Options) (coreexecutor.Response, error) {
		calls = append(calls, req.Model)
		return coreexecutor.Response{}, fmt.Errorf("primary failed")
	}}
	fallback := &modelExecutionCaptureExecutor{provider: "combo-fallback", execute: func(_ context.Context, _ *coreauth.Auth, req coreexecutor.Request, _ coreexecutor.Options) (coreexecutor.Response, error) {
		calls = append(calls, req.Model)
		return coreexecutor.Response{Payload: []byte("fallback response")}, nil
	}}
	manager.RegisterExecutor(primary)
	manager.RegisterExecutor(fallback)
	for _, entry := range []struct {
		provider string
		model    string
	}{{"combo-primary", "combo-main-model"}, {"combo-fallback", "combo-backup-model"}} {
		auth := &coreauth.Auth{ID: "combo-auth-" + entry.provider, Provider: entry.provider, Status: coreauth.StatusActive}
		if _, err := manager.Register(context.Background(), auth); err != nil {
			t.Fatal(err)
		}
		registry.GetGlobalRegistry().RegisterClient(auth.ID, entry.provider, []*registry.ModelInfo{{ID: entry.model}})
		t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	}
	handler := NewBaseAPIHandlers(nil, manager)
	handler.SetComboResolver(orderedComboResolver{targets: []ComboTarget{{Provider: "combo-primary", Model: "combo-main-model"}, {Provider: "combo-fallback", Model: "combo-backup-model"}}})
	body, _, errMsg := handler.executeWithAuthManagerFormats(context.Background(), "openai", "openai", "fast", []byte(`{"model":"fast"}`), "", false, modelExecutionOptions{})
	if errMsg != nil || string(body) != "fallback response" {
		t.Fatalf("combo response = %q, error = %v", body, errMsg)
	}
	if len(calls) != 2 || calls[0] != "combo-main-model" || calls[1] != "combo-backup-model" {
		t.Fatalf("attempt order = %v", calls)
	}
}

func TestComboRetriesAnotherModelOnSameProvider(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	manager.SetRetryConfig(0, 0, 1)
	var calls []string
	executor := &modelExecutionCaptureExecutor{provider: "combo-shared-provider", execute: func(_ context.Context, _ *coreauth.Auth, req coreexecutor.Request, _ coreexecutor.Options) (coreexecutor.Response, error) {
		calls = append(calls, req.Model)
		if req.Model == "combo-main-model" {
			return coreexecutor.Response{}, fmt.Errorf("primary failed")
		}
		return coreexecutor.Response{Payload: []byte("fallback response")}, nil
	}}
	manager.RegisterExecutor(executor)
	auth := &coreauth.Auth{ID: "combo-shared-auth", Provider: executor.Identifier(), Status: coreauth.StatusActive}
	if _, err := manager.Register(context.Background(), auth); err != nil {
		t.Fatal(err)
	}
	registry.GetGlobalRegistry().RegisterClient(auth.ID, executor.Identifier(), []*registry.ModelInfo{{ID: "combo-main-model"}, {ID: "combo-backup-model"}})
	t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	handler := NewBaseAPIHandlers(nil, manager)
	handler.SetComboResolver(orderedComboResolver{targets: []ComboTarget{{Provider: executor.Identifier(), Model: "combo-main-model"}, {Provider: executor.Identifier(), Model: "combo-backup-model"}}})
	body, _, errMsg := handler.executeWithAuthManagerFormats(context.Background(), "openai", "openai", "fast", []byte(`{"model":"fast"}`), "", false, modelExecutionOptions{})
	if errMsg != nil || string(body) != "fallback response" {
		t.Fatalf("combo response = %q, error = %v", body, errMsg)
	}
	if len(calls) != 2 || calls[0] != "combo-main-model" || calls[1] != "combo-backup-model" {
		t.Fatalf("attempt order = %v", calls)
	}
}

func TestComboStreamFallsBackBeforeFirstChunk(t *testing.T) {
	manager := coreauth.NewManager(nil, nil, nil)
	manager.SetRetryConfig(0, 0, 1)
	primary := &modelExecutionCaptureExecutor{provider: "combo-stream-primary", stream: func(context.Context, *coreauth.Auth, coreexecutor.Request, coreexecutor.Options) (*coreexecutor.StreamResult, error) {
		return nil, fmt.Errorf("primary stream failed")
	}}
	fallback := &modelExecutionCaptureExecutor{provider: "combo-stream-fallback", stream: func(context.Context, *coreauth.Auth, coreexecutor.Request, coreexecutor.Options) (*coreexecutor.StreamResult, error) {
		chunks := make(chan coreexecutor.StreamChunk, 1)
		chunks <- coreexecutor.StreamChunk{Payload: []byte("fallback chunk")}
		close(chunks)
		return &coreexecutor.StreamResult{Chunks: chunks}, nil
	}}
	manager.RegisterExecutor(primary)
	manager.RegisterExecutor(fallback)
	for _, entry := range []struct {
		provider string
		model    string
	}{{"combo-stream-primary", "combo-stream-main"}, {"combo-stream-fallback", "combo-stream-backup"}} {
		auth := &coreauth.Auth{ID: "combo-auth-" + entry.provider, Provider: entry.provider, Status: coreauth.StatusActive}
		if _, err := manager.Register(context.Background(), auth); err != nil {
			t.Fatal(err)
		}
		registry.GetGlobalRegistry().RegisterClient(auth.ID, entry.provider, []*registry.ModelInfo{{ID: entry.model}})
		t.Cleanup(func() { registry.GetGlobalRegistry().UnregisterClient(auth.ID) })
	}
	handler := NewBaseAPIHandlers(nil, manager)
	handler.SetComboResolver(orderedComboResolver{targets: []ComboTarget{{Provider: "combo-stream-primary", Model: "combo-stream-main"}, {Provider: "combo-stream-fallback", Model: "combo-stream-backup"}}})
	data, _, errors := handler.executeStreamWithAuthManagerFormats(context.Background(), "openai", "openai", "fast", []byte(`{"model":"fast"}`), "", false, modelExecutionOptions{})
	if data == nil {
		t.Fatal("combo stream returned no data channel")
	}
	if got := string(<-data); got != "fallback chunk" {
		t.Fatalf("first stream chunk = %q", got)
	}
	for errMsg := range errors {
		if errMsg != nil {
			t.Fatalf("combo stream error = %v", errMsg)
		}
	}
	if _, ok := <-data; ok {
		t.Fatal("expected fallback stream to close after its first chunk")
	}
}

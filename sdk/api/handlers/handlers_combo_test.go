package handlers

import (
	"context"
	"testing"

	coreexecutor "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/executor"
)

type comboResolverTest struct{}

func (comboResolverTest) ResolveCombo(model string) ([]ComboTarget, bool) {
	if model != "fast" {
		return nil, false
	}
	return []ComboTarget{{Provider: "codex", Model: "gpt-5"}, {Provider: "claude", Model: "sonnet"}}, true
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

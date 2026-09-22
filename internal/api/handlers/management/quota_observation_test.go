package management

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	coreauth "github.com/router-for-me/CLIProxyAPI/v7/sdk/cliproxy/auth"
	"github.com/router-for-me/CLIProxyAPI/v7/sdk/pluginapi"
)

func TestQuotaObservationResponseNormalizesUsedPercent(t *testing.T) {
	auth := &coreauth.Auth{Provider: "codex", Quota: coreauth.QuotaState{
		ObservedAt: time.Unix(10, 0),
		Signals:    map[string]string{"X-Codex-Primary-Used-Percent": "25"},
	}}
	response := quotaObservationResponse(auth)
	groups, ok := response["groups"].([]pluginapi.QuotaGroup)
	if !ok || len(groups) != 1 || groups[0].Buckets[0].RemainingFraction != 0.75 {
		t.Fatalf("normalized response = %#v", response)
	}
}

func TestQuotaObservationResponseIncludesAntigravityCredits(t *testing.T) {
	auth := &coreauth.Auth{ID: "antigravity-test", Provider: "antigravity"}
	coreauth.SetAntigravityCreditsHint(auth.ID, coreauth.AntigravityCreditsHint{Known: true, Available: true, CreditAmount: 12.5, MinCreditAmount: 1})
	response := quotaObservationResponse(auth)
	if response["credits_available"] != true {
		t.Fatalf("credits_available = %#v", response["credits_available"])
	}
	if summary, ok := response["summary"].([]pluginapi.QuotaMetric); !ok || len(summary) != 2 || summary[0].Value != 12.5 {
		t.Fatalf("credits summary = %#v", response["summary"])
	}
}

func TestParseAntigravityNativeQuotaUsesVisibleModels(t *testing.T) {
	data, _ := json.Marshal(map[string]any{"models": map[string]any{
		"gemini-3.7-flash":      map[string]any{"displayName": "Gemini 3.7 Flash", "quotaInfo": map[string]any{"remainingFraction": 0.9, "resetTime": "2026-01-01T00:00:00Z"}},
		"gemini-3.8-flash-high": map[string]any{"displayName": "Gemini 3.8 Flash (High)", "quotaInfo": map[string]any{"remainingFraction": 0.9, "resetTime": "2026-01-01T00:00:00Z"}},
		"gemini-3.1-pro-low":    map[string]any{"displayName": "Gemini 3.1 Pro (Low)", "quotaInfo": map[string]any{"remainingFraction": 0.8}},
		"claude-opus":           map[string]any{"displayName": "Claude Opus 4.6 (Thinking)", "quotaInfo": map[string]any{"remainingFraction": 0.7}},
		"gemini-image":          map[string]any{"displayName": "Gemini 3.1 Flash Image", "quotaInfo": map[string]any{"remainingFraction": 0.6}},
		"chat-model":            map[string]any{"displayName": "chat_23310", "quotaInfo": map[string]any{"remainingFraction": 0.5}},
		"tab-preview":           map[string]any{"displayName": "tab_flash_lite_preview", "quotaInfo": map[string]any{"remainingFraction": 0.4}},
	}})
	allowed := map[string]string{
		"gemini 3.7 flash":           "Gemini 3.7 Flash",
		"gemini 3.8 flash":           "Gemini 3.8 Flash",
		"gemini 3.1 pro (high)":      "Gemini 3.1 Pro (High)",
		"gemini 3.1 pro (low)":       "Gemini 3.1 Pro (Low)",
		"claude opus 4.6 (thinking)": "Claude Opus 4.6 (Thinking)",
		"gemini 3.1 flash image":     "Gemini 3.1 Flash Image",
		"chat_23310":                 "chat_23310",
		"tab_flash_lite_preview":     "tab_flash_lite_preview",
	}
	response := parseAntigravityNativeQuota(data, allowed)
	if len(response.Groups) != 5 {
		t.Fatalf("visible model response = %#v", response)
	}
	for _, group := range response.Groups {
		if strings.Contains(group.DisplayName, "Image") || strings.Contains(group.DisplayName, "chat_") || strings.Contains(group.DisplayName, "tab_") {
			t.Fatalf("filtered model leaked into response = %#v", response)
		}
	}
}

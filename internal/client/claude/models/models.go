// Package models builds model catalogs for Anthropic clients.
package models

import (
	"sort"
	"strings"
)

const claudeDDModelPrefix = "claude-fable-5-dd-"

// BuildResponse builds an Anthropic model response from available models.
func BuildResponse(availableModels []map[string]any, disableCloaking bool) map[string]any {
	return BuildResponseWithProviders(availableModels, nil, disableCloaking)
}

// BuildResponseWithProviders builds an Anthropic model response and labels duplicate display names with their providers.
func BuildResponseWithProviders(availableModels []map[string]any, providersForModel func(string) []string, disableCloaking bool) map[string]any {
	models := make([]map[string]any, len(availableModels))
	for i, model := range availableModels {
		models[i] = cloneModel(model)
	}
	addProviderLabelsToDuplicateNames(models, providersForModel)
	if !disableCloaking {
		for i := range models {
			if id, ok := models[i]["id"].(string); ok {
				models[i]["id"] = EnsureClaudeModelIDPrefix(id)
			}
		}
	}

	sort.SliceStable(models, func(i, j int) bool {
		displayNameI, _ := models[i]["display_name"].(string)
		displayNameJ, _ := models[j]["display_name"].(string)
		if displayNameI != displayNameJ {
			return displayNameI < displayNameJ
		}
		idI, _ := models[i]["id"].(string)
		idJ, _ := models[j]["id"].(string)
		return idI < idJ
	})

	firstID := ""
	lastID := ""
	if len(models) > 0 {
		firstID, _ = models[0]["id"].(string)
		lastID, _ = models[len(models)-1]["id"].(string)
	}

	return map[string]any{
		"data":     models,
		"has_more": false,
		"first_id": firstID,
		"last_id":  lastID,
	}
}

func addProviderLabelsToDuplicateNames(models []map[string]any, providersForModel func(string) []string) {
	if providersForModel == nil {
		return
	}
	type modelLabel struct{ id, displayName, key, provider string }
	labels := make([]modelLabel, len(models))
	nameCounts := map[string]int{}
	providerCounts := map[string]map[string]int{}
	for i, model := range models {
		id, _ := model["id"].(string)
		displayName, _ := model["display_name"].(string)
		displayName = strings.TrimSpace(displayName)
		if id == "" || displayName == "" {
			continue
		}
		key := strings.ToLower(displayName)
		provider := strings.Join(providersForModel(id), ", ")
		labels[i] = modelLabel{id: id, displayName: displayName, key: key, provider: provider}
		nameCounts[key]++
		if provider != "" {
			if providerCounts[key] == nil {
				providerCounts[key] = map[string]int{}
			}
			providerCounts[key][provider]++
		}
	}
	for i, label := range labels {
		if label.id == "" || nameCounts[label.key] < 2 {
			continue
		}
		provider := label.provider
		if provider == "" {
			provider = label.id
		} else if providerCounts[label.key][provider] > 1 {
			if prefix, _, ok := strings.Cut(label.id, "/"); ok && prefix != "" {
				provider += " / " + prefix
			} else {
				provider += " / " + label.id
			}
		}
		models[i]["display_name"] = label.displayName + " (" + provider + ")"
	}
}

// EnsureClaudeModelIDPrefix rewrites model IDs for Anthropic model listings.
// IDs that already start with "claude-" are returned unchanged; all other IDs
// become "claude-fable-5-dd-" plus the original ID with its characters reversed.
func EnsureClaudeModelIDPrefix(id string) string {
	if id == "" || strings.HasPrefix(id, "claude-") {
		return id
	}
	return claudeDDModelPrefix + reverseModelID(id)
}

// ResolveClaudeModelIDPrefix reverses EnsureClaudeModelIDPrefix for request routing.
// Optional thinking suffixes in model(value) form are preserved.
func ResolveClaudeModelIDPrefix(id string) string {
	if id == "" {
		return id
	}
	base, suffix, hasSuffix := splitModelThinkingSuffix(id)
	if !strings.HasPrefix(base, claudeDDModelPrefix) {
		return id
	}
	encoded := base[len(claudeDDModelPrefix):]
	if encoded == "" {
		return id
	}
	resolved := reverseModelID(encoded)
	if hasSuffix {
		return resolved + "(" + suffix + ")"
	}
	return resolved
}

func cloneModel(model map[string]any) map[string]any {
	cloned := make(map[string]any, len(model))
	for key, value := range model {
		cloned[key] = value
	}
	return cloned
}

func splitModelThinkingSuffix(model string) (base, suffix string, hasSuffix bool) {
	lastOpen := strings.LastIndex(model, "(")
	if lastOpen == -1 || !strings.HasSuffix(model, ")") {
		return model, "", false
	}
	return model[:lastOpen], model[lastOpen+1 : len(model)-1], true
}

func reverseModelID(id string) string {
	runes := []rune(id)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

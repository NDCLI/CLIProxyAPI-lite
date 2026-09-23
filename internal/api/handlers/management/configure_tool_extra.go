package management

import (
	"fmt"
	"path/filepath"
	"strings"
)

const extraToolProvider = "cliproxyapi-lite"

func extraCLITools() []cliToolItem {
	home := homeDir()
	return []cliToolItem{
		{ID: "opencode", Label: "OpenCode", ConfigPath: filepath.Join(home, ".config", "opencode", "opencode.json"), Format: "json", Command: "opencode", ModelRequired: true, CanReset: true, Capabilities: []string{"subagent"}},
		{ID: "openclaw", Label: "OpenClaw", ConfigPath: filepath.Join(home, ".openclaw", "openclaw.json"), Format: "json", Command: "openclaw", ModelRequired: true, CanReset: true},
		{ID: "droid", Label: "Factory Droid", ConfigPath: filepath.Join(home, ".factory", "settings.json"), Format: "json", Command: "droid", ModelRequired: true, CanReset: true},
		{ID: "hermes", Label: "Hermes Agent", ConfigPath: filepath.Join(home, ".hermes", "config.yaml"), Format: "yaml", Command: "hermes", ModelRequired: true, CanReset: true},
		{ID: "kilo", Label: "Kilo Code CLI", ConfigPath: filepath.Join(home, ".local", "share", "kilo", "auth.json"), Format: "json", Command: "kilo", ModelRequired: true, CanReset: true},
		{ID: "deepseek-tui", Label: "DeepSeek TUI", ConfigPath: filepath.Join(home, ".deepseek", "config.toml"), Format: "toml", Command: "deepseek", ModelRequired: true, CanReset: true},
		{ID: "grok-build", Label: "Grok Build", ConfigPath: filepath.Join(home, ".grok", "config.toml"), Format: "toml", Command: "grok", ModelRequired: true, CanReset: true, Capabilities: []string{"subagent"}},
	}
}

// extraToolConfig follows each tool's native configuration shape. File validation,
// secret-safe previews, persistence and selective reset are owned by the caller.
func extraToolConfig(tool, baseURL, apiKey, model, subagentModel string, existing map[string]any) (map[string]any, error) {
	if strings.TrimSpace(model) == "" {
		return nil, fmt.Errorf("model is required")
	}
	if existing == nil {
		existing = make(map[string]any)
	}
	switch tool {
	case "opencode":
		provider, errMap := extraToolObject(existing, "provider", extraToolProvider)
		if errMap != nil {
			return nil, errMap
		}
		provider["npm"] = "@ai-sdk/openai-compatible"
		provider["name"] = "CLIProxyAPI-lite"
		options, errMap := extraToolObject(provider, "options")
		if errMap != nil {
			return nil, errMap
		}
		options["baseURL"], options["apiKey"] = baseURL, apiKey
		models, errMap := extraToolObject(provider, "models")
		if errMap != nil {
			return nil, errMap
		}
		entry, errMap := extraToolObject(models, model)
		if errMap != nil {
			return nil, errMap
		}
		entry["name"] = model
		existing["model"] = extraToolProvider + "/" + model
		if subagentModel != "" {
			subModel, errMap := extraToolObject(models, subagentModel)
			if errMap != nil {
				return nil, errMap
			}
			subModel["name"] = subagentModel
			explorer, errMap := extraToolObject(existing, "agent", "explorer")
			if errMap != nil {
				return nil, errMap
			}
			explorer["mode"], explorer["model"] = "subagent", extraToolProvider+"/"+subagentModel
		}
	case "openclaw":
		provider, errMap := extraToolObject(existing, "models", "providers", extraToolProvider)
		if errMap != nil {
			return nil, errMap
		}
		provider["baseUrl"], provider["apiKey"], provider["api"] = baseURL, apiKey, "openai-completions"
		providerModels, errModels := extraToolModelList(provider["models"], "models.providers."+extraToolProvider+".models")
		if errModels != nil {
			return nil, errModels
		}
		found := false
		for _, value := range providerModels {
			entry := value.(map[string]any)
			if entry["id"] == model {
				found = true
			}
		}
		if !found {
			providerModels = append(providerModels, map[string]any{"id": model, "name": model})
		}
		provider["models"] = providerModels
		defaults, errMap := extraToolObject(existing, "agents", "defaults")
		if errMap != nil {
			return nil, errMap
		}
		// OpenClaw accepts both a legacy string and an object with fallbacks.
		if _, legacy := defaults["model"].(string); legacy {
			defaults["model"] = map[string]any{"primary": defaults["model"]}
		}
		defaultModel, errMap := extraToolObject(defaults, "model")
		if errMap != nil {
			return nil, errMap
		}
		defaultModel["primary"] = extraToolProvider + "/" + model
		allowlist, errMap := extraToolObject(defaults, "models")
		if errMap != nil {
			return nil, errMap
		}
		if _, errMap = extraToolObject(allowlist, extraToolProvider+"/"+model); errMap != nil {
			return nil, errMap
		}
	case "droid":
		entries, errModels := extraToolModelList(existing["customModels"], "customModels")
		if errModels != nil {
			return nil, errModels
		}
		managedID := "custom:CLIProxyAPI-lite-0"
		managed := map[string]any{"id": managedID}
		updated := make([]any, 0, len(entries)+1)
		found := false
		for _, value := range entries {
			entry := value.(map[string]any)
			if entry["id"] == managedID {
				managed = entry
				found = true
				managed["model"], managed["baseUrl"], managed["apiKey"] = model, baseURL, apiKey
				managed["displayName"], managed["provider"] = model, "openai"
				updated = append(updated, managed)
			} else {
				updated = append(updated, entry)
			}
		}
		if !found {
			managed["model"], managed["baseUrl"], managed["apiKey"] = model, baseURL, apiKey
			managed["displayName"], managed["provider"] = model, "openai"
			updated = append(updated, managed)
		}
		existing["customModels"] = updated
	case "hermes":
		settings, errMap := extraToolObject(existing, "model")
		if errMap != nil {
			return nil, errMap
		}
		settings["default"], settings["provider"] = model, "custom"
		settings["base_url"], settings["api_key"] = baseURL, apiKey
	case "kilo":
		settings, errMap := extraToolObject(existing, "openai-compatible")
		if errMap != nil {
			return nil, errMap
		}
		settings["type"], settings["apiKey"], settings["baseUrl"], settings["model"] = "api-key", apiKey, baseURL, model
	case "deepseek-tui":
		settings, errMap := extraToolObject(existing, "providers", "openai")
		if errMap != nil {
			return nil, errMap
		}
		settings["base_url"], settings["api_key"], settings["model"] = baseURL, apiKey, model
		existing["provider"] = "openai"
	case "grok-build":
		settings, errMap := extraToolObject(existing, "model", extraToolProvider)
		if errMap != nil {
			return nil, errMap
		}
		settings["model"], settings["base_url"], settings["api_key"] = model, baseURL, apiKey
		settings["name"], settings["api_backend"] = "CLIProxyAPI-lite", "chat_completions"
		models, errMap := extraToolObject(existing, "models")
		if errMap != nil {
			return nil, errMap
		}
		models["default"] = extraToolProvider
		if subagentModel != "" {
			subagents, errMap := extraToolObject(existing, "subagents", "models")
			if errMap != nil {
				return nil, errMap
			}
			for _, role := range []string{"general-purpose", "explore", "plan"} {
				slot := extraToolProvider + "-" + role
				subModel, errMap := extraToolObject(existing, "model", slot)
				if errMap != nil {
					return nil, errMap
				}
				subModel["model"], subModel["base_url"], subModel["api_key"] = subagentModel, baseURL, apiKey
				subModel["name"], subModel["api_backend"] = "CLIProxyAPI-lite "+role, "chat_completions"
				subagents[role] = slot
			}
		}
	default:
		return nil, fmt.Errorf("unsupported tool: %s", tool)
	}
	return existing, nil
}

func extraToolObject(values map[string]any, path ...string) (map[string]any, error) {
	current := values
	for index, key := range path {
		value, exists := current[key]
		if !exists {
			value = make(map[string]any)
			current[key] = value
		}
		next, ok := value.(map[string]any)
		if !ok || next == nil {
			return nil, fmt.Errorf("invalid tool configuration: %s must be an object", strings.Join(path[:index+1], "."))
		}
		current = next
	}
	return current, nil
}

func extraToolModelList(value any, field string) ([]any, error) {
	if value == nil {
		return []any{}, nil
	}
	values, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("invalid tool configuration: %s must be an array", field)
	}
	for _, entry := range values {
		if object, valid := entry.(map[string]any); !valid || object == nil {
			return nil, fmt.Errorf("invalid tool configuration: %s entries must be objects", field)
		}
	}
	return values, nil
}

func extraToolCurrent(tool string, config map[string]any) cliToolCurrent {
	current := cliToolCurrent{}
	object := func(path ...string) map[string]any {
		result := config
		for _, key := range path {
			if result == nil {
				return nil
			}
			result, _ = result[key].(map[string]any)
		}
		return result
	}
	str := func(values map[string]any, key string) string {
		value, _ := values[key].(string)
		return value
	}
	switch tool {
	case "opencode":
		current.Model = strings.TrimPrefix(str(config, "model"), extraToolProvider+"/")
		current.BaseURL = str(object("provider", extraToolProvider, "options"), "baseURL")
		current.SubagentModel = strings.TrimPrefix(str(object("agent", "explorer"), "model"), extraToolProvider+"/")
	case "openclaw":
		defaults := object("agents", "defaults")
		current.Model = str(defaults, "model")
		if current.Model == "" {
			current.Model = str(object("agents", "defaults", "model"), "primary")
		}
		current.Model = strings.TrimPrefix(current.Model, extraToolProvider+"/")
		current.BaseURL = str(object("models", "providers", extraToolProvider), "baseUrl")
	case "droid":
		models, _ := extraToolModelList(config["customModels"], "customModels")
		for _, value := range models {
			entry := value.(map[string]any)
			if entry["id"] == "custom:CLIProxyAPI-lite-0" {
				current.Model, current.BaseURL = str(entry, "model"), str(entry, "baseUrl")
				break
			}
		}
	case "hermes":
		settings := object("model")
		current.Model, current.BaseURL = str(settings, "default"), str(settings, "base_url")
	case "kilo":
		settings := object("openai-compatible")
		current.Model, current.BaseURL = str(settings, "model"), str(settings, "baseUrl")
	case "deepseek-tui":
		settings := object("providers", "openai")
		current.Model, current.BaseURL = str(settings, "model"), str(settings, "base_url")
	case "grok-build":
		settings := object("model", extraToolProvider)
		current.Model, current.BaseURL = str(settings, "model"), str(settings, "base_url")
		current.SubagentModel = str(object("model", extraToolProvider+"-general-purpose"), "model")
	}
	return current
}

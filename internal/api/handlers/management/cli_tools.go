package management

import (
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/registry"
)

type cliToolCurrent struct {
	BaseURL           string            `json:"base_url,omitempty"`
	Model             string            `json:"model,omitempty"`
	Models            map[string]string `json:"models,omitempty"`
	SubagentModel     string            `json:"subagent_model,omitempty"`
	AutoCompactWindow int               `json:"auto_compact_window,omitempty"`
}

type cliToolItem struct {
	ID            string         `json:"id"`
	Label         string         `json:"label"`
	Description   string         `json:"description,omitempty"`
	Category      string         `json:"category"`
	Format        string         `json:"format,omitempty"`
	Command       string         `json:"command,omitempty"`
	Configured    bool           `json:"configured"`
	Installed     bool           `json:"installed"`
	ConfigExists  bool           `json:"config_exists"`
	ConfigError   bool           `json:"config_error,omitempty"`
	ModelRequired bool           `json:"model_required,omitempty"`
	Capabilities  []string       `json:"capabilities,omitempty"`
	Current       cliToolCurrent `json:"current"`
	ConfigPath    string         `json:"-"`
	CanReset      bool           `json:"can_reset"`
}

type cliToolModel struct {
	ID       string `json:"id"`
	Label    string `json:"label"`
	Provider string `json:"provider,omitempty"`
	Kind     string `json:"kind"`
}

func cliTools() []cliToolItem {
	home := homeDir()
	items := []cliToolItem{
		{ID: "claude-code", Label: "Claude Code", Category: "auto", Format: "json", Command: "claude", ConfigPath: filepath.Join(home, ".claude", "settings.json"), Capabilities: []string{"models", "auto_compact", "reset"}},
		{ID: "codex-cli", Label: "OpenAI Codex", Category: "auto", Format: "toml", Command: "codex", ConfigPath: filepath.Join(home, ".codex", "config.toml"), ModelRequired: true, Capabilities: []string{"subagent", "reset"}},
		{ID: "copilot", Label: "GitHub Copilot", Category: "guide", Command: "code"},
		{ID: "cursor", Label: "Cursor", Category: "guide", Command: "cursor"},
		{ID: "cline", Label: "Cline", Category: "guide", Command: "code"},
		{ID: "continue", Label: "Continue", Category: "guide", Command: "code"},
		{ID: "roo", Label: "Roo Code", Category: "guide", Command: "code"},
		{ID: "amp", Label: "Amp CLI", Category: "guide", Command: "amp"},
		{ID: "qwen-code", Label: "Qwen Code", Category: "guide", Command: "qwen"},
		{ID: "opendesign", Label: "OpenDesign", Category: "guide"},
	}
	return append(items, extraCLITools()...)
}

func findCLITool(id string) (cliToolItem, bool) {
	id = strings.TrimSpace(strings.ToLower(id))
	for _, item := range cliTools() {
		if item.ID == id {
			return item, true
		}
	}
	return cliToolItem{}, false
}

func inspectCLITool(item *cliToolItem) {
	if item == nil {
		return
	}
	if item.Command != "" {
		_, errLookPath := exec.LookPath(item.Command)
		item.Installed = errLookPath == nil
	}
	if item.Format == "" {
		return
	}
	if item.ConfigPath != "" {
		item.CanReset = fileExists(cliToolBackupPath(item.ConfigPath))
	}
	if item.ConfigPath == "" {
		return
	}
	if _, errStat := os.Stat(item.ConfigPath); errStat != nil {
		return
	}
	item.ConfigExists = true
	config, errRead := readToolConfig(*item)
	if errRead != nil {
		item.ConfigError = true
		return
	}
	item.Current = currentCLIToolConfig(item.ID, config)
	item.Configured = strings.TrimSpace(item.Current.BaseURL) != ""
	item.Installed = item.Installed || item.ConfigExists
}

func (h *Handler) GetCLITools(c *gin.Context) {
	items := cliTools()
	for index := range items {
		inspectCLITool(&items[index])
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}

func (h *Handler) GetCLITool(c *gin.Context) {
	item, ok := findCLITool(c.Param("id"))
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "cli_tool_not_found", "message": "CLI tool not found"}})
		return
	}
	inspectCLITool(&item)
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": item})
}

func (h *Handler) GetCLIToolModels(c *gin.Context) {
	models := make([]cliToolModel, 0)
	seen := make(map[string]struct{})
	for _, format := range []string{"openai", "claude", "gemini"} {
		for _, item := range registry.GetGlobalRegistry().GetAvailableModels(format) {
			id, _ := item["id"].(string)
			if id == "" {
				continue
			}
			if _, exists := seen[id]; exists {
				continue
			}
			seen[id] = struct{}{}
			label, _ := item["display_name"].(string)
			if label == "" {
				label, _ = item["name"].(string)
			}
			if label == "" {
				label = id
			}
			models = append(models, cliToolModel{ID: id, Label: label, Provider: format, Kind: "model"})
		}
	}
	if h != nil && h.combos != nil {
		for _, definition := range h.combos.List() {
			if !definition.Enabled || strings.TrimSpace(definition.Model) == "" {
				continue
			}
			if _, exists := seen[definition.Model]; exists {
				continue
			}
			seen[definition.Model] = struct{}{}
			models = append(models, cliToolModel{ID: definition.Model, Label: definition.Name, Kind: "combo"})
		}
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": models, "next_cursor": nil})
}

func cliToolBackupPath(configPath string) string {
	return configPath + ".cliproxyapi-backup"
}

func currentCLIToolConfig(tool string, config map[string]any) cliToolCurrent {
	current := cliToolCurrent{}
	str := func(values map[string]any, key string) string { value, _ := values[key].(string); return value }
	object := func(values map[string]any, key string) map[string]any {
		value, _ := values[key].(map[string]any)
		return value
	}
	switch tool {
	case "claude-code":
		env := object(config, "env")
		current.BaseURL, current.Model = str(env, "ANTHROPIC_BASE_URL"), str(env, "ANTHROPIC_MODEL")
		current.Models = map[string]string{}
		for slot, key := range map[string]string{"fable": "ANTHROPIC_DEFAULT_FABLE_MODEL", "opus": "ANTHROPIC_DEFAULT_OPUS_MODEL", "sonnet": "ANTHROPIC_DEFAULT_SONNET_MODEL", "haiku": "ANTHROPIC_DEFAULT_HAIKU_MODEL"} {
			if value := str(env, key); value != "" {
				current.Models[slot] = value
			}
		}
		if value := str(env, "CLAUDE_CODE_AUTO_COMPACT_WINDOW"); value != "" {
			_, _ = fmt.Sscan(value, &current.AutoCompactWindow)
		}
	case "codex-cli":
		current.Model, _ = config["model"].(string)
		provider := object(object(config, "model_providers"), "cliproxyapi-lite")
		current.BaseURL = str(provider, "base_url")
		agents := object(config, "agents")
		current.SubagentModel = str(agents, "default_subagent_model")
	default:
		current = extraToolCurrent(tool, config)
	}
	return current
}

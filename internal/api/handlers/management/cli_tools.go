package management

import (
	"net/http"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
)

type cliToolItem struct {
	ID         string `json:"id"`
	Label      string `json:"label"`
	Configured bool   `json:"configured"`
	ConfigPath string `json:"-"`
	CanReset   bool   `json:"can_reset"`
}

func cliTools() []cliToolItem {
	home := homeDir()
	appData := os.Getenv("APPDATA")
	vscodePath := filepath.Join(home, ".config", "Code", "User", "settings.json")
	cursorPath := filepath.Join(home, ".config", "Cursor", "User", "settings.json")
	clinePath := filepath.Join(home, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings.json")
	if runtime.GOOS == "windows" && appData != "" {
		vscodePath = filepath.Join(appData, "Code", "User", "settings.json")
		cursorPath = filepath.Join(appData, "Cursor", "User", "settings.json")
		clinePath = filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev", "settings.json")
	}
	return []cliToolItem{
		{ID: "claude-code", Label: "Claude Code", ConfigPath: filepath.Join(home, ".claude", "settings.json"), CanReset: true},
		{ID: "codex-cli", Label: "Codex CLI", ConfigPath: filepath.Join(home, ".codex", "config.toml"), CanReset: true},
		{ID: "continue", Label: "Continue.dev", ConfigPath: filepath.Join(home, ".continue", "config.json")},
		{ID: "cline", Label: "Cline", ConfigPath: clinePath},
		{ID: "vscode", Label: "VS Code", ConfigPath: vscodePath},
		{ID: "cursor", Label: "Cursor", ConfigPath: cursorPath},
		{ID: "env-openai", Label: "OpenAI environment"},
		{ID: "env-anthropic", Label: "Anthropic environment"},
	}
}

func cliToolConfigured(item *cliToolItem) {
	if item == nil || item.ConfigPath == "" {
		return
	}
	body, errRead := os.ReadFile(item.ConfigPath)
	if errRead != nil {
		return
	}
	item.Configured = strings.Contains(string(body), "CLIProxyAPI-lite") || strings.Contains(string(body), "cliproxyapi-lite") || strings.Contains(string(body), "127.0.0.1:")
}

// GetCLITools reports safe local configuration status without exposing credentials.
func (h *Handler) GetCLITools(c *gin.Context) {
	items := cliTools()
	for index := range items {
		cliToolConfigured(&items[index])
	}
	c.JSON(http.StatusOK, gin.H{"schema_version": 1, "items": items, "next_cursor": nil})
}

func (h *Handler) GetCLITool(c *gin.Context) {
	for _, item := range cliTools() {
		if item.ID == strings.TrimSpace(c.Param("id")) {
			cliToolConfigured(&item)
			c.JSON(http.StatusOK, gin.H{"schema_version": 1, "item": item})
			return
		}
	}
	c.JSON(http.StatusNotFound, gin.H{"error": gin.H{"code": "cli_tool_not_found", "message": "CLI tool not found"}})
}

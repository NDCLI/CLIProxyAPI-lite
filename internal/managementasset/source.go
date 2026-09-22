package managementasset

import (
	"embed"
	"strings"
)

//go:embed web/*
var sourceManagementAssets embed.FS

// SourceManagementAsset returns an allow-listed asset for the source-owned management UI.
func SourceManagementAsset(name string) ([]byte, string, bool) {
	name = strings.TrimPrefix(strings.TrimSpace(name), "/")
	contentType := ""
	switch name {
	case "index.html":
		contentType = "text/html; charset=utf-8"
	case "app.css":
		contentType = "text/css; charset=utf-8"
	case "app.js":
		contentType = "text/javascript; charset=utf-8"
	case "i18n/en.json", "i18n/vi.json":
		contentType = "application/json; charset=utf-8"
	case "providers/antigravity.png", "providers/codex.png", "providers/claude.png", "providers/gemini.png", "providers/qwen.png", "providers/kimi.png", "providers/openai.png", "providers/cursor.png", "providers/cline.png", "providers/continue.png", "providers/iflow.png", "providers/github.png":
		contentType = "image/png"
	default:
		return nil, "", false
	}
	body, errRead := sourceManagementAssets.ReadFile("web/" + name)
	if errRead != nil {
		return nil, "", false
	}
	return body, contentType, true
}

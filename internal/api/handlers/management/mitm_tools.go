package management

type mitmSourceModel struct {
	ID    string `json:"id"`
	Label string `json:"label"`
}

type mitmToolInfo struct {
	ID           string            `json:"id"`
	Label        string            `json:"label"`
	Hosts        []string          `json:"hosts"`
	SourceModels []mitmSourceModel `json:"source_models"`
}

// These are IDE source slots, not gateway provider models. Users select the target separately.
func mitmTools() []mitmToolInfo {
	return []mitmToolInfo{
		{ID: "antigravity", Label: "Antigravity", Hosts: mitmToolHosts["antigravity"], SourceModels: []mitmSourceModel{
			{"gemini-3.8-flash-high", "Gemini 3.8 Flash (High)"}, {"gemini-3.8-flash-medium", "Gemini 3.8 Flash (Medium)"}, {"gemini-3.8-flash-low", "Gemini 3.8 Flash (Low)"},
			{"gemini-3.7-flash-high", "Gemini 3.7 Flash (High)"}, {"gemini-3.7-flash-medium", "Gemini 3.7 Flash (Medium)"}, {"gemini-3.7-flash-low", "Gemini 3.7 Flash (Low)"},
			{"gemini-3.6-flash-high", "Gemini 3.6 Flash (High)"}, {"gemini-3.6-flash-medium", "Gemini 3.6 Flash (Medium)"}, {"gemini-3.6-flash-low", "Gemini 3.6 Flash (Low)"},
			{"gemini-3.5-flash-low", "Gemini 3.5 Flash (Medium / Default)"}, {"gemini-3-flash-agent", "Gemini 3.5 Flash (High)"}, {"gemini-3.5-flash-extra-low", "Gemini 3.5 Flash (Low)"},
			{"gemini-3.1-pro-low", "Gemini 3.1 Pro (Low)"}, {"gemini-pro-agent", "Gemini 3.1 Pro (High)"},
			{"claude-sonnet-4-6", "Claude Sonnet 4.6 (Thinking)"}, {"claude-opus-4-6-thinking", "Claude Opus 4.6 (Thinking)"},
			{"gpt-oss-120b-medium", "GPT-OSS 120B (Medium)"}, {"gemini-3-flash", "Gemini 3 Flash (Command)"},
		}},
		{ID: "kiro", Label: "Kiro", Hosts: mitmToolHosts["kiro"], SourceModels: []mitmSourceModel{
			{"auto", "Auto (Kiro Agent)"}, {"claude-sonnet-5", "Claude Sonnet 5"}, {"claude-sonnet-4.5", "Claude Sonnet 4.5"}, {"claude-sonnet-4", "Claude Sonnet 4"}, {"claude-haiku-4.5", "Claude Haiku 4.5"},
			{"deepseek-3.2", "DeepSeek 3.2"}, {"minimax-m2.1", "MiniMax M2.1"}, {"gpt-5.6-sol", "GPT 5.6 Sol"}, {"gpt-5.6-terra", "GPT 5.6 Terra"}, {"gpt-5.6-luna", "GPT 5.6 Luna"}, {"simple-task", "Qwen3 Coder Next"},
		}},
		{ID: "copilot", Label: "GitHub Copilot", Hosts: mitmToolHosts["copilot"], SourceModels: []mitmSourceModel{}},
	}
}

var mitmAntigravityAliases = map[string]string{
	"gemini-default":          "gemini-3.5-flash-low",
	"gemini-3.5-flash-high":   "gemini-3-flash-agent",
	"gemini-3.5-flash-medium": "gemini-3.5-flash-low",
	"gemini-3.8-flash":        "gemini-3.8-flash-medium",
	"gemini-3.1-pro-high":     "gemini-pro-agent",
	"gemini-3-pro-high":       "gemini-pro-agent",
	"gemini-3-pro-low":        "gemini-3.1-pro-low",
}

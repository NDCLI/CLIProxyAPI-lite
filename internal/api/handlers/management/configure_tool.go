package management

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	log "github.com/sirupsen/logrus"
)

type configureToolRequest struct {
	Tool   string `json:"tool"`
	APIKey string `json:"api_key"`
	Model  string `json:"model,omitempty"`
	Action string `json:"action,omitempty"`
}

type configureToolResponse struct {
	Status  string `json:"status"`
	Tool    string `json:"tool"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
}

func (h *Handler) ConfigureTool(c *gin.Context) {
	var req configureToolRequest
	if errBind := c.ShouldBindJSON(&req); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	req.Tool = strings.TrimSpace(strings.ToLower(req.Tool))
	req.APIKey = strings.TrimSpace(req.APIKey)
	req.Action = strings.TrimSpace(strings.ToLower(req.Action))
	if req.Action == "reset" {
		var resp configureToolResponse
		var errReset error
		switch req.Tool {
		case "claude-code":
			resp, errReset = resetClaudeCode()
		case "codex-cli":
			resp, errReset = resetCodexCLI()
		default:
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("reset is not supported for tool: %s", req.Tool)})
			return
		}
		if errReset != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": errReset.Error()})
			return
		}
		c.JSON(http.StatusOK, resp)
		return
	}
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_key is required"})
		return
	}

	serverAddr := fmt.Sprintf("http://127.0.0.1:%d", h.cfg.Port)

	var resp configureToolResponse
	var errConfigure error

	switch req.Tool {
	case "claude-code":
		resp, errConfigure = configureClaudeCode(serverAddr, req.APIKey, req.Model)
	case "codex-cli":
		resp, errConfigure = configureCodexCLI(serverAddr, req.APIKey, req.Model)
	case "continue":
		resp, errConfigure = configureContinueDev(serverAddr, req.APIKey, req.Model)
	case "cline":
		resp, errConfigure = configureCline(serverAddr, req.APIKey, req.Model)
	case "vscode":
		resp, errConfigure = configureVSCode(serverAddr, req.APIKey, req.Model)
	case "cursor":
		resp, errConfigure = configureCursor(serverAddr, req.APIKey, req.Model)
	case "env-openai":
		resp, errConfigure = configureEnvOpenAI(serverAddr, req.APIKey, req.Model)
	case "env-anthropic":
		resp, errConfigure = configureEnvAnthropic(serverAddr, req.APIKey, req.Model)
	default:
		c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("unsupported tool: %s", req.Tool)})
		return
	}

	if errConfigure != nil {
		log.WithError(errConfigure).WithField("tool", req.Tool).Error("failed to configure tool")
		c.JSON(http.StatusInternalServerError, gin.H{"error": errConfigure.Error()})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func homeDir() string {
	if h, errHome := os.UserHomeDir(); errHome == nil {
		return h
	}
	return ""
}

func writeJSONFile(path string, data any) error {
	if errDir := os.MkdirAll(filepath.Dir(path), 0o700); errDir != nil {
		return errDir
	}
	content, errMarshal := json.MarshalIndent(data, "", "  ")
	if errMarshal != nil {
		return errMarshal
	}
	return os.WriteFile(path, content, 0o600)
}

func mergeJSONFile(path string, merge map[string]any) error {
	existing := make(map[string]any)
	if data, errRead := os.ReadFile(path); errRead == nil {
		_ = json.Unmarshal(data, &existing)
	}
	for k, v := range merge {
		existing[k] = v
	}
	return writeJSONFile(path, existing)
}

func configureClaudeCode(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	settings := make(map[string]any)
	if data, errRead := os.ReadFile(settingsPath); errRead == nil {
		_ = json.Unmarshal(data, &settings)
	}

	envVars, ok := settings["env"].(map[string]any)
	if !ok {
		envVars = make(map[string]any)
	}
	envVars["ANTHROPIC_BASE_URL"] = serverAddr + "/v1"
	envVars["ANTHROPIC_AUTH_TOKEN"] = apiKey
	delete(envVars, "ANTHROPIC_API_KEY")
	if model != "" {
		envVars["ANTHROPIC_MODEL"] = model
	}
	settings["env"] = envVars
	settings["hasCompletedOnboarding"] = true

	if errWrite := writeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	msg := fmt.Sprintf("Claude Code configured → %s", serverAddr)
	if model != "" {
		msg += fmt.Sprintf(" (model: %s)", model)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "claude-code",
		Message: msg,
		Path:    settingsPath,
	}, nil
}

func configureCodexCLI(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}
	if model == "" {
		return configureToolResponse{}, fmt.Errorf("model is required for Codex CLI")
	}
	codexDir := filepath.Join(home, ".codex")
	configPath := filepath.Join(codexDir, "config.toml")
	authPath := filepath.Join(codexDir, "auth.json")
	configData := make(map[string]any)
	if existing, errRead := os.ReadFile(configPath); errRead == nil {
		if errParse := toml.Unmarshal(existing, &configData); errParse != nil {
			return configureToolResponse{}, fmt.Errorf("parse %s: %w", configPath, errParse)
		}
	}
	configData["model"] = model
	configData["model_provider"] = "cliproxyapi-lite"
	providers := nestedMap(configData, "model_providers")
	providers["cliproxyapi-lite"] = map[string]any{
		"name":     "CLIProxyAPI-lite",
		"base_url": serverAddr + "/v1",
		"wire_api": "responses",
	}
	configData["model_providers"] = providers
	agents := nestedMap(configData, "agents")
	agents["subagent"] = map[string]any{"model": model}
	configData["agents"] = agents
	encodedConfig, errMarshal := toml.Marshal(configData)
	if errMarshal != nil {
		return configureToolResponse{}, fmt.Errorf("encode Codex config: %w", errMarshal)
	}
	if errDir := os.MkdirAll(codexDir, 0o700); errDir != nil {
		return configureToolResponse{}, errDir
	}
	if errWrite := os.WriteFile(configPath, encodedConfig, 0o600); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("write %s: %w", configPath, errWrite)
	}
	authData := make(map[string]any)
	if existing, errRead := os.ReadFile(authPath); errRead == nil {
		_ = json.Unmarshal(existing, &authData)
	}
	authData["OPENAI_API_KEY"] = apiKey
	authData["auth_mode"] = "apikey"
	if errWrite := writeJSONFile(authPath, authData); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("write %s: %w", authPath, errWrite)
	}

	msg := fmt.Sprintf("Codex CLI configured → %s/v1", serverAddr)
	if model != "" {
		msg += fmt.Sprintf(" [model: %s]", model)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "codex-cli",
		Message: msg,
		Path:    configPath,
	}, nil
}

func resetClaudeCode() (configureToolResponse, error) {
	settingsPath := filepath.Join(homeDir(), ".claude", "settings.json")
	settings := make(map[string]any)
	data, errRead := os.ReadFile(settingsPath)
	if os.IsNotExist(errRead) {
		return configureToolResponse{Status: "ok", Tool: "claude-code", Message: "Claude Code has no CLIProxyAPI-lite settings"}, nil
	}
	if errRead != nil || json.Unmarshal(data, &settings) != nil {
		return configureToolResponse{}, fmt.Errorf("read %s", settingsPath)
	}
	if envVars, ok := settings["env"].(map[string]any); ok {
		for _, key := range []string{"ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY", "ANTHROPIC_MODEL"} {
			delete(envVars, key)
		}
		if len(envVars) == 0 {
			delete(settings, "env")
		}
	}
	if errWrite := writeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, errWrite
	}
	return configureToolResponse{Status: "ok", Tool: "claude-code", Message: "CLIProxyAPI-lite settings removed", Path: settingsPath}, nil
}

func resetCodexCLI() (configureToolResponse, error) {
	home := homeDir()
	configPath := filepath.Join(home, ".codex", "config.toml")
	authPath := filepath.Join(home, ".codex", "auth.json")
	configData := make(map[string]any)
	if existing, errRead := os.ReadFile(configPath); errRead == nil {
		if errParse := toml.Unmarshal(existing, &configData); errParse != nil {
			return configureToolResponse{}, fmt.Errorf("parse %s: %w", configPath, errParse)
		}
		managedModel, _ := configData["model"].(string)
		if configData["model_provider"] == "cliproxyapi-lite" {
			delete(configData, "model")
			delete(configData, "model_provider")
		}
		if providers, ok := configData["model_providers"].(map[string]any); ok {
			delete(providers, "cliproxyapi-lite")
			if len(providers) == 0 {
				delete(configData, "model_providers")
			}
		}
		if agents, ok := configData["agents"].(map[string]any); ok {
			if subagent, ok := agents["subagent"].(map[string]any); ok && subagent["model"] == managedModel {
				delete(agents, "subagent")
			}
			if len(agents) == 0 {
				delete(configData, "agents")
			}
		}
		encoded, errMarshal := toml.Marshal(configData)
		if errMarshal != nil {
			return configureToolResponse{}, errMarshal
		}
		if errWrite := os.WriteFile(configPath, encoded, 0o600); errWrite != nil {
			return configureToolResponse{}, errWrite
		}
	}
	if data, errRead := os.ReadFile(authPath); errRead == nil {
		authData := make(map[string]any)
		if json.Unmarshal(data, &authData) == nil {
			delete(authData, "OPENAI_API_KEY")
			if authData["auth_mode"] == "apikey" {
				delete(authData, "auth_mode")
			}
			if errWrite := writeJSONFile(authPath, authData); errWrite != nil {
				return configureToolResponse{}, errWrite
			}
		}
	}
	return configureToolResponse{Status: "ok", Tool: "codex-cli", Message: "CLIProxyAPI-lite settings removed", Path: configPath}, nil
}

func nestedMap(values map[string]any, key string) map[string]any {
	if nested, ok := values[key].(map[string]any); ok {
		return nested
	}
	return make(map[string]any)
}

func configureContinueDev(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	configPath := filepath.Join(home, ".continue", "config.json")
	existing := make(map[string]any)
	if data, errRead := os.ReadFile(configPath); errRead == nil {
		_ = json.Unmarshal(data, &existing)
	}

	openaiModel := "gpt-4o"
	anthropicModel := "claude-sonnet-4-20250514"
	if model != "" {
		if strings.Contains(model, "claude") || strings.Contains(model, "anthropic") {
			anthropicModel = model
		} else {
			openaiModel = model
		}
	}

	models := []map[string]any{
		{
			"title":    "CLIProxyAPI-lite (OpenAI)",
			"provider": "openai",
			"model":    openaiModel,
			"apiBase":  serverAddr + "/v1",
			"apiKey":   apiKey,
		},
		{
			"title":    "CLIProxyAPI-lite (Anthropic)",
			"provider": "anthropic",
			"model":    anthropicModel,
			"apiBase":  serverAddr,
			"apiKey":   apiKey,
		},
	}
	existing["models"] = models

	if errWrite := writeJSONFile(configPath, existing); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", configPath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "continue",
		Message: "Continue.dev configured with OpenAI + Anthropic providers",
		Path:    configPath,
	}, nil
}

func configureCline(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	if model == "" {
		model = "gpt-4o"
	}

	var settingsDir string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return configureToolResponse{}, fmt.Errorf("APPDATA not set")
		}
		settingsDir = filepath.Join(appData, "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	} else {
		settingsDir = filepath.Join(home, ".config", "Code", "User", "globalStorage", "saoudrizwan.claude-dev")
	}

	settingsPath := filepath.Join(settingsDir, "settings.json")
	settings := map[string]any{
		"apiProvider":   "openai-native",
		"openAiBaseUrl": serverAddr + "/v1",
		"openAiApiKey":  apiKey,
		"openAiModelId": model,
	}

	if errWrite := mergeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "cline",
		Message: fmt.Sprintf("Cline configured → model: %s", model),
		Path:    settingsPath,
	}, nil
}

func configureVSCode(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	var settingsPath string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return configureToolResponse{}, fmt.Errorf("APPDATA not set")
		}
		settingsPath = filepath.Join(appData, "Code", "User", "settings.json")
	} else {
		settingsPath = filepath.Join(home, ".config", "Code", "User", "settings.json")
	}

	settings := map[string]any{
		"claude-code.apiBaseUrl": serverAddr,
	}
	if model != "" {
		settings["claude-code.model"] = model
	}

	if errWrite := mergeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	msg := "VS Code Claude Code extension configured"
	if model != "" {
		msg += fmt.Sprintf(" (model: %s)", model)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "vscode",
		Message: msg,
		Path:    settingsPath,
	}, nil
}

func configureCursor(serverAddr, apiKey, model string) (configureToolResponse, error) {
	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	var settingsPath string
	if runtime.GOOS == "windows" {
		appData := os.Getenv("APPDATA")
		if appData == "" {
			return configureToolResponse{}, fmt.Errorf("APPDATA not set")
		}
		settingsPath = filepath.Join(appData, "Cursor", "User", "settings.json")
	} else if runtime.GOOS == "darwin" {
		settingsPath = filepath.Join(home, "Library", "Application Support", "Cursor", "User", "settings.json")
	} else {
		settingsPath = filepath.Join(home, ".config", "Cursor", "User", "settings.json")
	}

	settings := map[string]any{
		"openai.apiBaseUrl": serverAddr + "/v1",
	}
	if model != "" {
		settings["openai.model"] = model
	}

	if errWrite := mergeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	msg := "Cursor configured → OpenAI Base URL set"
	if model != "" {
		msg += fmt.Sprintf(" (model: %s)", model)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "cursor",
		Message: msg,
		Path:    settingsPath,
	}, nil
}

func configureEnvOpenAI(serverAddr, apiKey, model string) (configureToolResponse, error) {
	if runtime.GOOS == "windows" {
		if errSet := setWindowsUserEnv("OPENAI_BASE_URL", serverAddr+"/v1"); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if errSet := setWindowsUserEnv("OPENAI_API_KEY", apiKey); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if model != "" {
			_ = setWindowsUserEnv("OPENAI_MODEL", model)
		}
		msg := "OPENAI_BASE_URL and OPENAI_API_KEY set for current user (restart terminal to apply)"
		if model != "" {
			msg += fmt.Sprintf(" [model: %s]", model)
		}
		return configureToolResponse{
			Status:  "ok",
			Tool:    "env-openai",
			Message: msg,
		}, nil
	}
	msg := fmt.Sprintf("Set OPENAI_BASE_URL=%s/v1 and OPENAI_API_KEY in your shell profile (the key is intentionally omitted)", serverAddr)
	if model != "" {
		msg += fmt.Sprintf("\nexport OPENAI_MODEL=%s", model)
	}
	return configureToolResponse{
		Status:  "ok",
		Tool:    "env-openai",
		Message: msg,
	}, nil
}

func configureEnvAnthropic(serverAddr, apiKey, model string) (configureToolResponse, error) {
	if runtime.GOOS == "windows" {
		if errSet := setWindowsUserEnv("ANTHROPIC_BASE_URL", serverAddr); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if errSet := setWindowsUserEnv("ANTHROPIC_API_KEY", apiKey); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if model != "" {
			_ = setWindowsUserEnv("ANTHROPIC_MODEL", model)
		}
		msg := "ANTHROPIC_BASE_URL and ANTHROPIC_API_KEY set for current user (restart terminal to apply)"
		if model != "" {
			msg += fmt.Sprintf(" [model: %s]", model)
		}
		return configureToolResponse{
			Status:  "ok",
			Tool:    "env-anthropic",
			Message: msg,
		}, nil
	}
	msg := fmt.Sprintf("Set ANTHROPIC_BASE_URL=%s and ANTHROPIC_API_KEY in your shell profile (the key is intentionally omitted)", serverAddr)
	if model != "" {
		msg += fmt.Sprintf("\nexport ANTHROPIC_MODEL=%s", model)
	}
	return configureToolResponse{
		Status:  "ok",
		Tool:    "env-anthropic",
		Message: msg,
	}, nil
}

func setWindowsUserEnv(name, value string) error {
	cmd := exec.Command("cmd", "/C", "setx", name, value)
	output, errRun := cmd.CombinedOutput()
	if errRun != nil {
		return fmt.Errorf("setx %s failed: %s — %w", name, strings.TrimSpace(string(output)), errRun)
	}
	return nil
}

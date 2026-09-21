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
	log "github.com/sirupsen/logrus"
)

type configureToolRequest struct {
	Tool   string `json:"tool"`
	APIKey string `json:"api_key"`
	Model  string `json:"model,omitempty"`
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
	if req.APIKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "api_key is required"})
		return
	}

	serverAddr := fmt.Sprintf("http://127.0.0.1:%d", h.cfg.Port)

	var resp configureToolResponse
	var errConfigure error

	switch req.Tool {
	case "claude-code":
		resp, errConfigure = configureClaudeCode(serverAddr, req.APIKey)
	case "codex-cli":
		resp, errConfigure = configureCodexCLI(serverAddr, req.APIKey)
	case "continue":
		resp, errConfigure = configureContinueDev(serverAddr, req.APIKey, req.Model)
	case "cline":
		resp, errConfigure = configureCline(serverAddr, req.APIKey, req.Model)
	case "vscode":
		resp, errConfigure = configureVSCode(serverAddr, req.APIKey)
	case "cursor":
		resp, errConfigure = configureCursor(serverAddr, req.APIKey)
	case "env-openai":
		resp, errConfigure = configureEnvOpenAI(serverAddr, req.APIKey)
	case "env-anthropic":
		resp, errConfigure = configureEnvAnthropic(serverAddr, req.APIKey)
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

func configureClaudeCode(serverAddr, apiKey string) (configureToolResponse, error) {
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
	envVars["ANTHROPIC_BASE_URL"] = serverAddr
	envVars["ANTHROPIC_API_KEY"] = apiKey
	settings["env"] = envVars

	if errWrite := writeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "claude-code",
		Message: fmt.Sprintf("Claude Code configured → %s", serverAddr),
		Path:    settingsPath,
	}, nil
}

func configureCodexCLI(serverAddr, apiKey string) (configureToolResponse, error) {
	if runtime.GOOS == "windows" {
		if errSet := setWindowsUserEnv("OPENAI_BASE_URL", serverAddr+"/v1"); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if errSet := setWindowsUserEnv("OPENAI_API_KEY", apiKey); errSet != nil {
			return configureToolResponse{}, errSet
		}
		return configureToolResponse{
			Status:  "ok",
			Tool:    "codex-cli",
			Message: "Codex CLI configured via user environment variables (restart terminal to apply)",
		}, nil
	}

	home := homeDir()
	if home == "" {
		return configureToolResponse{}, fmt.Errorf("cannot determine home directory")
	}

	profilePath := filepath.Join(home, ".bashrc")
	lines := []string{
		fmt.Sprintf("\n# CLIProxyAPI-lite — Codex CLI"),
		fmt.Sprintf("export OPENAI_BASE_URL=%s/v1", serverAddr),
		fmt.Sprintf("export OPENAI_API_KEY=%s", apiKey),
	}
	content := strings.Join(lines, "\n") + "\n"

	f, errOpen := os.OpenFile(profilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0o644)
	if errOpen != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", profilePath, errOpen)
	}
	defer func() {
		if errClose := f.Close(); errClose != nil {
			log.Errorf("failed to close %s: %v", profilePath, errClose)
		}
	}()
	if _, errWrite := f.WriteString(content); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", profilePath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "codex-cli",
		Message: fmt.Sprintf("Codex CLI configured → %s/v1 (restart terminal to apply)", serverAddr),
		Path:    profilePath,
	}, nil
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
		"apiProvider":  "openai-native",
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

func configureVSCode(serverAddr, apiKey string) (configureToolResponse, error) {
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

	if errWrite := mergeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "vscode",
		Message: "VS Code Claude Code extension configured",
		Path:    settingsPath,
	}, nil
}

func configureCursor(serverAddr, apiKey string) (configureToolResponse, error) {
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

	if errWrite := mergeJSONFile(settingsPath, settings); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("failed writing %s: %w", settingsPath, errWrite)
	}

	return configureToolResponse{
		Status:  "ok",
		Tool:    "cursor",
		Message: "Cursor configured → OpenAI Base URL set",
		Path:    settingsPath,
	}, nil
}

func configureEnvOpenAI(serverAddr, apiKey string) (configureToolResponse, error) {
	if runtime.GOOS == "windows" {
		if errSet := setWindowsUserEnv("OPENAI_BASE_URL", serverAddr+"/v1"); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if errSet := setWindowsUserEnv("OPENAI_API_KEY", apiKey); errSet != nil {
			return configureToolResponse{}, errSet
		}
		return configureToolResponse{
			Status:  "ok",
			Tool:    "env-openai",
			Message: "OPENAI_BASE_URL and OPENAI_API_KEY set for current user (restart terminal to apply)",
		}, nil
	}
	return configureToolResponse{
		Status:  "ok",
		Tool:    "env-openai",
		Message: fmt.Sprintf("Set these in your shell profile:\nexport OPENAI_BASE_URL=%s/v1\nexport OPENAI_API_KEY=%s", serverAddr, apiKey),
	}, nil
}

func configureEnvAnthropic(serverAddr, apiKey string) (configureToolResponse, error) {
	if runtime.GOOS == "windows" {
		if errSet := setWindowsUserEnv("ANTHROPIC_BASE_URL", serverAddr); errSet != nil {
			return configureToolResponse{}, errSet
		}
		if errSet := setWindowsUserEnv("ANTHROPIC_API_KEY", apiKey); errSet != nil {
			return configureToolResponse{}, errSet
		}
		return configureToolResponse{
			Status:  "ok",
			Tool:    "env-anthropic",
			Message: "ANTHROPIC_BASE_URL and ANTHROPIC_API_KEY set for current user (restart terminal to apply)",
		}, nil
	}
	return configureToolResponse{
		Status:  "ok",
		Tool:    "env-anthropic",
		Message: fmt.Sprintf("Set these in your shell profile:\nexport ANTHROPIC_BASE_URL=%s\nexport ANTHROPIC_API_KEY=%s", serverAddr, apiKey),
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

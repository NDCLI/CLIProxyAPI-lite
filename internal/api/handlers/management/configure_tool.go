package management

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"gopkg.in/yaml.v3"
)

// ponytail: one lock serializes local config edits; split per file if contention becomes measurable.
var cliConfigMu sync.Mutex

type configureToolRequest struct {
	Tool              string            `json:"tool"`
	APIKey            string            `json:"api_key"`
	APIKeyID          string            `json:"api_key_id"`
	BaseURL           string            `json:"base_url"`
	Model             string            `json:"model,omitempty"`
	Models            map[string]string `json:"models,omitempty"`
	SubagentModel     string            `json:"subagent_model,omitempty"`
	AutoCompactWindow *int              `json:"auto_compact_window,omitempty"`
	Action            string            `json:"action,omitempty"`
}

type configureToolResponse struct {
	Status  string `json:"status"`
	Tool    string `json:"tool"`
	Message string `json:"message"`
	Path    string `json:"path,omitempty"`
	Format  string `json:"format,omitempty"`
	Preview string `json:"preview,omitempty"`
	Skipped int    `json:"skipped,omitempty"`
}

func (h *Handler) ConfigureTool(c *gin.Context) {
	var req configureToolRequest
	if errBind := c.ShouldBindJSON(&req); errBind != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}
	req.Tool, req.Action = strings.ToLower(strings.TrimSpace(req.Tool)), strings.ToLower(strings.TrimSpace(req.Action))
	item, ok := findCLITool(req.Tool)
	if !ok || item.Format == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "This tool uses manual configuration; open its setup guide"})
		return
	}
	if req.Action != "" && req.Action != "apply" && req.Action != "preview" && req.Action != "reset" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "action must be apply, preview, or reset"})
		return
	}
	if req.Action != "reset" {
		if req.BaseURL == "" {
			h.mu.Lock()
			if h.cfg == nil {
				h.mu.Unlock()
				c.JSON(http.StatusServiceUnavailable, gin.H{"error": "configuration unavailable"})
				return
			}
			req.BaseURL = fmt.Sprintf("http://127.0.0.1:%d", h.cfg.Port)
			h.mu.Unlock()
		}
		baseURL, errURL := normalizeCLIBaseURL(req.BaseURL)
		if req.Tool == "claude-cowork" {
			baseURL, errURL = normalizeClaudeCoworkBaseURL(req.BaseURL)
		}
		if errURL != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errURL.Error()})
			return
		}
		req.BaseURL = baseURL
		apiKey, errKey := h.resolveCLIEndpointKey(req.BaseURL, req.APIKeyID, req.APIKey)
		if errKey != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errKey.Error()})
			return
		}
		if req.Action != "preview" && apiKey == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "api_key or api_key_id is required"})
			return
		}
		req.APIKey = apiKey
		if errValidate := validateCLIToolRequest(item, req); errValidate != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": errValidate.Error()})
			return
		}
	}
	cliConfigMu.Lock()
	defer cliConfigMu.Unlock()
	var response configureToolResponse
	var errConfigure error
	if req.Tool == "claude-cowork" {
		response, errConfigure = configureClaudeCowork(item, req)
	} else {
		response, errConfigure = configureCLITool(item, req)
	}
	if errConfigure != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": errConfigure.Error()})
		return
	}
	c.JSON(http.StatusOK, response)
}

func normalizeCLIBaseURL(value string) (string, error) {
	parsed, errParse := url.Parse(strings.TrimSpace(value))
	if errParse != nil || parsed == nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("base_url must be an HTTP(S) URL without credentials, query, or fragment")
	}
	parsed.Path = strings.TrimRight(parsed.Path, "/")
	if !strings.HasSuffix(parsed.Path, "/v1") {
		parsed.Path += "/v1"
	}
	parsed.RawPath = ""
	return parsed.String(), nil
}

func (h *Handler) resolveCLIEndpointKey(baseURL, keyID, apiKey string) (string, error) {
	keyID, apiKey = strings.TrimSpace(keyID), strings.TrimSpace(apiKey)
	if keyID == "" {
		if strings.ContainsAny(apiKey, "\r\n\x00") {
			return "", fmt.Errorf("api_key must not contain control characters")
		}
		return apiKey, nil
	}
	if apiKey != "" {
		return "", fmt.Errorf("choose api_key or api_key_id, not both")
	}
	parsed, errParse := url.Parse(baseURL)
	if errParse != nil || parsed == nil || parsed.User != nil || parsed.RawQuery != "" || parsed.Fragment != "" {
		return "", fmt.Errorf("invalid endpoint for saved API key")
	}
	if h == nil {
		return "", fmt.Errorf("configuration unavailable")
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.cfg == nil {
		return "", fmt.Errorf("configuration unavailable")
	}
	host := parsed.Hostname()
	loopback := host == "localhost"
	if ip := net.ParseIP(host); ip != nil {
		loopback = ip.IsLoopback()
	}
	port := parsed.Port()
	if port == "" {
		if parsed.Scheme == "https" {
			port = "443"
		} else if parsed.Scheme == "http" {
			port = "80"
		}
	}
	path := strings.TrimRight(parsed.Path, "/")
	if !loopback || port != strconv.Itoa(h.cfg.Port) || (path != "" && path != "/v1") || (parsed.Scheme != "http" && parsed.Scheme != "https") {
		return "", fmt.Errorf("saved endpoint keys can only be used with this local server; enter a custom key for other endpoints")
	}
	for _, key := range h.cfg.APIKeys {
		if endpointKeyID(key) == keyID {
			if strings.ContainsAny(key, "\r\n\x00") {
				return "", fmt.Errorf("saved endpoint key contains invalid characters")
			}
			return key, nil
		}
	}
	return "", fmt.Errorf("endpoint API key no longer exists; refresh and select a key")
}

func validateCLIToolRequest(item cliToolItem, req configureToolRequest) error {
	if len(req.BaseURL) > 2048 || len(req.APIKey) > 8192 || len(req.Model) > 512 || len(req.SubagentModel) > 512 {
		return fmt.Errorf("endpoint, key, or model value is too long")
	}
	if strings.ContainsAny(req.APIKey, "\r\n\x00") {
		return fmt.Errorf("invalid API key value")
	}
	if item.ModelRequired && strings.TrimSpace(req.Model) == "" {
		return fmt.Errorf("model is required for %s", item.Label)
	}
	for slot, model := range req.Models {
		if slot != "fable" && slot != "opus" && slot != "sonnet" && slot != "haiku" {
			return fmt.Errorf("unknown Claude model slot: %s", slot)
		}
		if strings.ContainsAny(model, "\r\n\x00") {
			return fmt.Errorf("invalid model value")
		}
	}
	if strings.ContainsAny(req.Model+req.SubagentModel, "\r\n\x00") {
		return fmt.Errorf("invalid model value")
	}
	if req.AutoCompactWindow != nil && (*req.AutoCompactWindow < 0 || *req.AutoCompactWindow > 10000000) {
		return fmt.Errorf("auto_compact_window must be between 0 and 10000000")
	}
	return nil
}

func buildCLIToolConfig(req configureToolRequest, config map[string]any) (map[string]any, error) {
	switch req.Tool {
	case "claude-code":
		env, errMap := extraToolObject(config, "env")
		if errMap != nil {
			return nil, errMap
		}
		env["ANTHROPIC_BASE_URL"], env["ANTHROPIC_AUTH_TOKEN"] = req.BaseURL, req.APIKey
		delete(env, "ANTHROPIC_API_KEY")
		if req.Model != "" {
			env["ANTHROPIC_MODEL"] = req.Model
		}
		for slot, model := range req.Models {
			key := "ANTHROPIC_DEFAULT_" + strings.ToUpper(slot) + "_MODEL"
			if strings.TrimSpace(model) == "" {
				delete(env, key)
			} else {
				env[key] = strings.TrimSpace(model)
			}
		}
		if req.AutoCompactWindow != nil {
			if *req.AutoCompactWindow == 0 {
				delete(env, "CLAUDE_CODE_AUTO_COMPACT_WINDOW")
			} else {
				env["CLAUDE_CODE_AUTO_COMPACT_WINDOW"] = strconv.Itoa(*req.AutoCompactWindow)
			}
		}
		config["env"] = env
		config["hasCompletedOnboarding"] = true
	case "codex-cli":
		config["model"], config["model_provider"] = req.Model, "cliproxyapi-lite"
		providers, errMap := extraToolObject(config, "model_providers")
		if errMap != nil {
			return nil, errMap
		}
		provider, errMap := extraToolObject(providers, "cliproxyapi-lite")
		if errMap != nil {
			return nil, errMap
		}
		provider["name"], provider["base_url"], provider["wire_api"] = "Lumina", req.BaseURL, "responses"
		headers, errMap := extraToolObject(provider, "http_headers")
		if errMap != nil {
			return nil, errMap
		}
		headers["Authorization"] = "Bearer " + req.APIKey
		provider["http_headers"] = headers
		provider["requires_openai_auth"] = false
		providers["cliproxyapi-lite"] = provider
		config["model_providers"] = providers
		agents, errMap := extraToolObject(config, "agents")
		if errMap != nil {
			return nil, errMap
		}
		model := req.SubagentModel
		if model == "" {
			model = req.Model
		}
		agents["default_subagent_model"] = model
		config["agents"] = agents
	default:
		return extraToolConfig(req.Tool, req.BaseURL, req.APIKey, req.Model, req.SubagentModel, config)
	}
	return config, nil
}

func homeDir() string {
	if home, errHome := os.UserHomeDir(); errHome == nil {
		return home
	}
	return ""
}

type cliToolBackup struct {
	Exists   bool   `json:"exists"`
	Format   string `json:"format"`
	Original []byte `json:"original,omitempty"`
}

func readToolConfig(item cliToolItem) (map[string]any, error) {
	data, errRead := os.ReadFile(item.ConfigPath)
	if errors.Is(errRead, os.ErrNotExist) {
		return map[string]any{}, nil
	}
	if errRead != nil {
		return nil, fmt.Errorf("read %s: %w", item.Label, errRead)
	}
	return decodeToolConfig(item.Format, data)
}

func decodeToolConfig(format string, data []byte) (map[string]any, error) {
	values := make(map[string]any)
	var errDecode error
	switch format {
	case "json":
		errDecode = json.Unmarshal(data, &values)
	case "toml":
		errDecode = toml.Unmarshal(data, &values)
	case "yaml":
		errDecode = yaml.Unmarshal(data, &values)
	default:
		return nil, fmt.Errorf("unsupported configuration format: %s", format)
	}
	if errDecode != nil {
		return nil, fmt.Errorf("existing %s configuration is invalid: %w", format, errDecode)
	}
	if values == nil {
		values = make(map[string]any)
	}
	return values, nil
}

func encodeToolConfig(format string, values map[string]any) ([]byte, error) {
	switch format {
	case "json":
		return json.MarshalIndent(values, "", "  ")
	case "toml":
		return toml.Marshal(values)
	case "yaml":
		return yaml.Marshal(values)
	default:
		return nil, fmt.Errorf("unsupported configuration format: %s", format)
	}
}

func writeToolFile(path string, data []byte) error {
	directory := filepath.Dir(path)
	if errDir := os.MkdirAll(directory, 0o700); errDir != nil {
		return errDir
	}
	file, errCreate := os.CreateTemp(directory, ".cliproxyapi-config-*")
	if errCreate != nil {
		return errCreate
	}
	temporary := file.Name()
	defer os.Remove(temporary)
	if errChmod := file.Chmod(0o600); errChmod != nil {
		_ = file.Close()
		return errChmod
	}
	if _, errWrite := file.Write(data); errWrite != nil {
		_ = file.Close()
		return errWrite
	}
	if errSync := file.Sync(); errSync != nil {
		_ = file.Close()
		return errSync
	}
	if errClose := file.Close(); errClose != nil {
		return errClose
	}
	return os.Rename(temporary, path)
}

func configureCLITool(item cliToolItem, req configureToolRequest) (configureToolResponse, error) {
	if item.ConfigPath == "" || item.Format == "" {
		return configureToolResponse{}, fmt.Errorf("this tool requires manual setup")
	}
	if req.Action == "reset" {
		backupPath := cliToolBackupPath(item.ConfigPath)
		data, errRead := os.ReadFile(backupPath)
		if errors.Is(errRead, os.ErrNotExist) {
			return configureToolResponse{Status: "ok", Tool: item.ID, Message: "No saved Lumina configuration to reset"}, nil
		}
		if errRead != nil {
			return configureToolResponse{}, errRead
		}
		var backup cliToolBackup
		if errDecode := json.Unmarshal(data, &backup); errDecode != nil || backup.Format != item.Format {
			return configureToolResponse{}, fmt.Errorf("saved backup for %s is invalid", item.Label)
		}
		current, errCurrent := readToolConfig(item)
		if errCurrent != nil {
			return configureToolResponse{}, errCurrent
		}
		if _, errStat := os.Stat(item.ConfigPath); errors.Is(errStat, os.ErrNotExist) {
			if backup.Exists {
				if errWrite := writeToolFile(item.ConfigPath, backup.Original); errWrite != nil {
					return configureToolResponse{}, errWrite
				}
			} else {
				_ = os.Remove(backupPath)
				return configureToolResponse{Status: "ok", Tool: item.ID, Message: "Lumina settings removed"}, nil
			}
		} else {
			original := map[string]any{}
			if backup.Exists {
				original, errCurrent = decodeToolConfig(item.Format, backup.Original)
				if errCurrent != nil {
					return configureToolResponse{}, fmt.Errorf("saved backup for %s is invalid: %w", item.Label, errCurrent)
				}
			}
			if errRestore := restoreCLIToolFields(item.ID, current, original); errRestore != nil {
				return configureToolResponse{}, errRestore
			}
			encoded, errEncode := encodeToolConfig(item.Format, current)
			if errEncode != nil {
				return configureToolResponse{}, errEncode
			}
			if errWrite := writeToolFile(item.ConfigPath, encoded); errWrite != nil {
				return configureToolResponse{}, errWrite
			}
		}
		if errRemove := os.Remove(backupPath); errRemove != nil && !errors.Is(errRemove, os.ErrNotExist) {
			return configureToolResponse{}, errRemove
		}
		return configureToolResponse{Status: "ok", Tool: item.ID, Message: "Lumina settings removed", Path: item.ConfigPath}, nil
	}
	current, errRead := readToolConfig(item)
	if errRead != nil {
		return configureToolResponse{}, errRead
	}
	updated, errBuild := buildCLIToolConfig(req, current)
	if errBuild != nil {
		return configureToolResponse{}, errBuild
	}
	encoded, errEncode := encodeToolConfig(item.Format, updated)
	if errEncode != nil {
		return configureToolResponse{}, errEncode
	}
	if req.Action == "preview" {
		previewReq := req
		previewReq.APIKey = "YOUR_API_KEY"
		previewConfig, errPreview := buildCLIToolConfig(previewReq, current)
		if errPreview != nil {
			return configureToolResponse{}, errPreview
		}
		preview, errPreview := encodeToolConfig(item.Format, previewConfig)
		if errPreview != nil {
			return configureToolResponse{}, errPreview
		}
		return configureToolResponse{Status: "preview", Tool: item.ID, Format: item.Format, Preview: string(preview), Path: item.ConfigPath}, nil
	}
	backupPath := cliToolBackupPath(item.ConfigPath)
	if _, errStat := os.Stat(backupPath); errors.Is(errStat, os.ErrNotExist) {
		original, errOriginal := os.ReadFile(item.ConfigPath)
		exists := errOriginal == nil
		if errors.Is(errOriginal, os.ErrNotExist) {
			original = nil
		} else if errOriginal != nil {
			return configureToolResponse{}, errOriginal
		}
		backup := cliToolBackup{Exists: exists, Format: item.Format, Original: original}
		backupData, errMarshal := json.Marshal(backup)
		if errMarshal != nil {
			return configureToolResponse{}, errMarshal
		}
		if errWrite := writeToolFile(backupPath, backupData); errWrite != nil {
			return configureToolResponse{}, fmt.Errorf("save reset backup: %w", errWrite)
		}
	} else if errStat != nil {
		return configureToolResponse{}, fmt.Errorf("check reset backup: %w", errStat)
	}
	if errWrite := writeToolFile(item.ConfigPath, encoded); errWrite != nil {
		return configureToolResponse{}, fmt.Errorf("write %s: %w", item.Label, errWrite)
	}
	return configureToolResponse{Status: "ok", Tool: item.ID, Message: item.Label + " configured", Path: item.ConfigPath}, nil
}

func cliToolManagedPaths(tool string) [][]string {
	switch tool {
	case "claude-code":
		paths := [][]string{{"hasCompletedOnboarding"}}
		for _, key := range []string{"ANTHROPIC_BASE_URL", "ANTHROPIC_AUTH_TOKEN", "ANTHROPIC_API_KEY", "ANTHROPIC_MODEL", "ANTHROPIC_DEFAULT_FABLE_MODEL", "ANTHROPIC_DEFAULT_OPUS_MODEL", "ANTHROPIC_DEFAULT_SONNET_MODEL", "ANTHROPIC_DEFAULT_HAIKU_MODEL", "CLAUDE_CODE_AUTO_COMPACT_WINDOW"} {
			paths = append(paths, []string{"env", key})
		}
		return paths
	case "codex-cli":
		return [][]string{{"model"}, {"model_provider"}, {"model_providers", "cliproxyapi-lite"}, {"agents", "default_subagent_model"}}
	case "opencode":
		return [][]string{{"provider", "cliproxyapi-lite"}, {"model"}, {"agent", "explorer", "mode"}, {"agent", "explorer", "model"}}
	case "openclaw":
		return [][]string{{"models", "providers", "cliproxyapi-lite"}, {"agents", "defaults", "model"}, {"agents", "defaults", "models", "cliproxyapi-lite"}}
	case "droid":
		return nil
	case "hermes":
		return [][]string{{"model", "default"}, {"model", "provider"}, {"model", "base_url"}, {"model", "api_key"}}
	case "kilo":
		return [][]string{{"openai-compatible"}}
	case "deepseek-tui":
		return [][]string{{"provider"}, {"providers", "openai"}}
	case "grok-build":
		return [][]string{{"model", "cliproxyapi-lite"}, {"model", "cliproxyapi-lite-general-purpose"}, {"model", "cliproxyapi-lite-explore"}, {"model", "cliproxyapi-lite-plan"}, {"models", "default"}, {"subagents", "models", "general-purpose"}, {"subagents", "models", "explore"}, {"subagents", "models", "plan"}}
	default:
		return nil
	}
}

func restoreCLIToolFields(tool string, current, original map[string]any) error {
	if tool == "droid" {
		currentModels, errCurrent := extraToolModelList(current["customModels"], "customModels")
		if errCurrent != nil {
			return errCurrent
		}
		originalModels, errOriginal := extraToolModelList(original["customModels"], "customModels")
		if errOriginal != nil {
			return errOriginal
		}
		const managedID = "custom:CLIProxyAPI-lite-0"
		var prior any
		for _, item := range originalModels {
			if item.(map[string]any)["id"] == managedID {
				prior = item
			}
		}
		merged := make([]any, 0, len(currentModels)+1)
		for _, item := range currentModels {
			if item.(map[string]any)["id"] != managedID {
				merged = append(merged, item)
			}
		}
		if prior != nil {
			merged = append(merged, prior)
		}
		if len(merged) == 0 {
			delete(current, "customModels")
		} else {
			current["customModels"] = merged
		}
		return nil
	}
	for _, path := range cliToolManagedPaths(tool) {
		value, exists := getCLIToolPath(original, path)
		if errSet := setCLIToolPath(current, path, value, exists); errSet != nil {
			return errSet
		}
	}
	return nil
}

func getCLIToolPath(values map[string]any, path []string) (any, bool) {
	current := values
	for index, key := range path {
		value, exists := current[key]
		if !exists {
			return nil, false
		}
		if index == len(path)-1 {
			return value, true
		}
		current, _ = value.(map[string]any)
		if current == nil {
			return nil, false
		}
	}
	return nil, false
}

func setCLIToolPath(values map[string]any, path []string, value any, exists bool) error {
	current := values
	for index, key := range path[:len(path)-1] {
		nested, ok := current[key].(map[string]any)
		if !ok {
			if _, present := current[key]; present {
				return fmt.Errorf("cannot reset configuration: %s changed to a non-object", strings.Join(path[:index+1], "."))
			}
			nested = make(map[string]any)
			current[key] = nested
		}
		current = nested
	}
	key := path[len(path)-1]
	if exists {
		current[key] = value
	} else {
		delete(current, key)
	}
	return nil
}

func configureClaudeCode(serverAddr, apiKey, model string) (configureToolResponse, error) {
	return configureLegacyCLITool("claude-code", serverAddr, apiKey, model)
}

func configureCodexCLI(serverAddr, apiKey, model string) (configureToolResponse, error) {
	return configureLegacyCLITool("codex-cli", serverAddr, apiKey, model)
}

func configureLegacyCLITool(id, serverAddr, apiKey, model string) (configureToolResponse, error) {
	item, ok := findCLITool(id)
	if !ok {
		return configureToolResponse{}, fmt.Errorf("unsupported tool: %s", id)
	}
	baseURL, errURL := normalizeCLIBaseURL(serverAddr)
	if errURL != nil {
		return configureToolResponse{}, errURL
	}
	return configureCLITool(item, configureToolRequest{Tool: id, BaseURL: baseURL, APIKey: apiKey, Model: model, Action: "apply"})
}

func resetCodexCLI() (configureToolResponse, error) {
	item, _ := findCLITool("codex-cli")
	return configureCLITool(item, configureToolRequest{Tool: item.ID, Action: "reset"})
}

func configureContinueDev(_, _, _ string) (configureToolResponse, error) {
	return configureToolResponse{Status: "guide", Tool: "continue", Message: "Use the Continue setup guide to add this gateway manually"}, nil
}

func configureEnvOpenAI(serverAddr, _, model string) (configureToolResponse, error) {
	return configureToolResponse{Status: "guide", Tool: "env-openai", Message: fmt.Sprintf("Set OPENAI_BASE_URL=%s/v1, OPENAI_API_KEY, and optionally OPENAI_MODEL=%s in your shell profile", strings.TrimRight(serverAddr, "/"), model)}, nil
}

func configureEnvAnthropic(serverAddr, _, model string) (configureToolResponse, error) {
	return configureToolResponse{Status: "guide", Tool: "env-anthropic", Message: fmt.Sprintf("Set ANTHROPIC_BASE_URL=%s, ANTHROPIC_API_KEY, and optionally ANTHROPIC_MODEL=%s in your shell profile", strings.TrimRight(serverAddr, "/"), model)}, nil
}

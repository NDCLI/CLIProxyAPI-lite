package management

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestConfigureToolRejectsEmptyAPIKey(t *testing.T) {
	h := NewHandlerWithoutConfigFilePath(&config.Config{Port: 8317}, nil)
	if h == nil {
		t.Fatal("expected handler")
	}
}

func TestConfigureClaudeCodeWritesSettings(t *testing.T) {
	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", home)
	defer os.Setenv("USERPROFILE", origUserProfile)

	resp, errCfg := configureClaudeCode("http://127.0.0.1:8317", "test-key-123", "claude-sonnet-4-20250514")
	if errCfg != nil {
		t.Fatalf("configureClaudeCode failed: %v", errCfg)
	}
	if resp.Status != "ok" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}

	settingsPath := filepath.Join(home, ".claude", "settings.json")
	data, errRead := os.ReadFile(settingsPath)
	if errRead != nil {
		t.Fatalf("settings file not created: %v", errRead)
	}
	content := string(data)
	if !contains(content, "ANTHROPIC_BASE_URL") || !contains(content, "http://127.0.0.1:8317") {
		t.Fatalf("settings file missing expected content: %s", content)
	}
	if !contains(content, "ANTHROPIC_AUTH_TOKEN") || !contains(content, "test-key-123") {
		t.Fatalf("settings file missing API key: %s", content)
	}
	if contains(content, "ANTHROPIC_API_KEY") {
		t.Fatalf("settings file retained legacy ANTHROPIC_API_KEY: %s", content)
	}
}

func TestConfigureAndResetCodexCLIPreservesUnrelatedSettings(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	codexDir := filepath.Join(home, ".codex")
	if errDir := os.MkdirAll(codexDir, 0o700); errDir != nil {
		t.Fatal(errDir)
	}
	if errWrite := os.WriteFile(filepath.Join(codexDir, "config.toml"), []byte("approval_policy = \"never\"\n"), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	if errWrite := os.WriteFile(filepath.Join(codexDir, "auth.json"), []byte(`{"tokens":{"access_token":"keep-me"}}`), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}

	resp, errConfigure := configureCodexCLI("http://127.0.0.1:8317", "gateway-key", "gpt-5.4")
	if errConfigure != nil {
		t.Fatalf("configureCodexCLI failed: %v", errConfigure)
	}
	configData, _ := os.ReadFile(resp.Path)
	if !contains(string(configData), `model_provider = 'cliproxyapi-lite'`) || !contains(string(configData), `approval_policy = 'never'`) {
		t.Fatalf("unexpected Codex config: %s", configData)
	}
	authData, _ := os.ReadFile(filepath.Join(codexDir, "auth.json"))
	if !contains(string(authData), "gateway-key") || !contains(string(authData), "keep-me") {
		t.Fatalf("unexpected Codex auth: %s", authData)
	}

	if _, errReset := resetCodexCLI(); errReset != nil {
		t.Fatalf("resetCodexCLI failed: %v", errReset)
	}
	configData, _ = os.ReadFile(resp.Path)
	if contains(string(configData), "cliproxyapi-lite") || !contains(string(configData), "approval_policy") {
		t.Fatalf("reset removed unrelated config or retained managed config: %s", configData)
	}
	authData, _ = os.ReadFile(filepath.Join(codexDir, "auth.json"))
	if contains(string(authData), "gateway-key") || !contains(string(authData), "keep-me") {
		t.Fatalf("reset removed unrelated auth or retained gateway key: %s", authData)
	}
}

func TestConfigureEnvMessagesDoNotEchoAPIKey(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("environment configuration uses setx on Windows")
	}
	key := "secret-env-key"
	for _, configure := range []func(string, string, string) (configureToolResponse, error){configureEnvOpenAI, configureEnvAnthropic} {
		response, errConfigure := configure("http://127.0.0.1:8317", key, "gpt-5")
		if errConfigure != nil {
			t.Fatalf("configure env failed: %v", errConfigure)
		}
		if contains(response.Message, key) {
			t.Fatalf("environment configuration echoed API key: %s", response.Message)
		}
	}
}

func TestConfigureContinueDevWritesConfig(t *testing.T) {
	home := t.TempDir()
	origHome := os.Getenv("HOME")
	os.Setenv("HOME", home)
	defer os.Setenv("HOME", origHome)

	origUserProfile := os.Getenv("USERPROFILE")
	os.Setenv("USERPROFILE", home)
	defer os.Setenv("USERPROFILE", origUserProfile)

	resp, errCfg := configureContinueDev("http://127.0.0.1:8317", "test-key", "claude-sonnet-4-20250514")
	if errCfg != nil {
		t.Fatalf("configureContinueDev failed: %v", errCfg)
	}
	if resp.Status != "ok" {
		t.Fatalf("unexpected status: %s", resp.Status)
	}

	configPath := filepath.Join(home, ".continue", "config.json")
	data, errRead := os.ReadFile(configPath)
	if errRead != nil {
		t.Fatalf("config file not created: %v", errRead)
	}
	content := string(data)
	if !contains(content, "CLIProxyAPI-lite") {
		t.Fatalf("config missing provider title: %s", content)
	}
	if !contains(content, "claude-sonnet-4-20250514") {
		t.Fatalf("config missing selected model: %s", content)
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsStr(s, substr))
}

func containsStr(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

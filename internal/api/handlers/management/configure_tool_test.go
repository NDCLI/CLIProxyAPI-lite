package management

import (
	"os"
	"path/filepath"
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
	if !contains(content, "ANTHROPIC_API_KEY") || !contains(content, "test-key-123") {
		t.Fatalf("settings file missing API key: %s", content)
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

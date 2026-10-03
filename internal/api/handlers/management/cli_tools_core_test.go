package management

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestCLIToolCatalogCoworkAndMITMOnlyCopilot(t *testing.T) {
	item, ok := findCLITool("claude-cowork")
	if !ok || item.Format != "claude-cowork" || item.ModelRequired {
		t.Fatalf("Cowork must be configurable without a required model: %#v", item)
	}
	if _, ok := findCLITool("copilot"); ok {
		t.Fatal("Copilot must not appear in the regular CLI catalog")
	}
	for _, item := range mitmTools() {
		if item.ID == "copilot" {
			return
		}
	}
	t.Fatal("Copilot must remain in the MITM catalog")
}

func TestCLIToolPreviewHidesKeyAndDoesNotWrite(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	item, _ := findCLITool("claude-code")
	response, errPreview := configureCLITool(item, configureToolRequest{
		Tool: item.ID, BaseURL: "http://127.0.0.1:8317/v1", APIKey: "secret-preview-key", Action: "preview",
	})
	if errPreview != nil {
		t.Fatal(errPreview)
	}
	if response.Status != "preview" || !strings.Contains(response.Preview, "YOUR_API_KEY") || strings.Contains(response.Preview, "secret-preview-key") {
		t.Fatalf("preview = %#v", response)
	}
	if _, errStat := os.Stat(item.ConfigPath); !os.IsNotExist(errStat) {
		t.Fatalf("preview unexpectedly wrote settings: %v", errStat)
	}
}

func TestCLIToolApplyRejectsMalformedConfigWithoutChangingIt(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	item, _ := findCLITool("claude-code")
	if errDir := os.MkdirAll(filepath.Dir(item.ConfigPath), 0o700); errDir != nil {
		t.Fatal(errDir)
	}
	malformed := []byte(`{"env":`)
	if errWrite := os.WriteFile(item.ConfigPath, malformed, 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	if _, errApply := configureLegacyCLITool(item.ID, "http://127.0.0.1:8317", "key", "claude-sonnet"); errApply == nil {
		t.Fatal("malformed settings were accepted")
	}
	actual, errRead := os.ReadFile(item.ConfigPath)
	if errRead != nil || string(actual) != string(malformed) {
		t.Fatalf("malformed settings changed: %q, %v", actual, errRead)
	}
	if fileExists(cliToolBackupPath(item.ConfigPath)) {
		t.Fatal("malformed settings produced a reset backup")
	}
}

func TestCLIToolResetRestoresManagedFieldsAndPreservesLaterChanges(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	item, _ := findCLITool("claude-code")
	if errDir := os.MkdirAll(filepath.Dir(item.ConfigPath), 0o700); errDir != nil {
		t.Fatal(errDir)
	}
	original := `{"theme":"dark","env":{"ANTHROPIC_BASE_URL":"https://previous.example/v1","ANTHROPIC_AUTH_TOKEN":"previous-key","USER_SETTING":"before"}}`
	if errWrite := os.WriteFile(item.ConfigPath, []byte(original), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	if _, errApply := configureLegacyCLITool(item.ID, "http://127.0.0.1:8317", "gateway-key", "claude-sonnet"); errApply != nil {
		t.Fatal(errApply)
	}
	var current map[string]any
	data, _ := os.ReadFile(item.ConfigPath)
	if errDecode := json.Unmarshal(data, &current); errDecode != nil {
		t.Fatal(errDecode)
	}
	current["theme"] = "changed-after-apply"
	current["user_option"] = "preserve"
	encoded, _ := json.Marshal(current)
	if errWrite := os.WriteFile(item.ConfigPath, encoded, 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	if _, errReset := configureCLITool(item, configureToolRequest{Tool: item.ID, Action: "reset"}); errReset != nil {
		t.Fatal(errReset)
	}
	data, _ = os.ReadFile(item.ConfigPath)
	if errDecode := json.Unmarshal(data, &current); errDecode != nil {
		t.Fatal(errDecode)
	}
	env := current["env"].(map[string]any)
	if env["ANTHROPIC_BASE_URL"] != "https://previous.example/v1" || env["ANTHROPIC_AUTH_TOKEN"] != "previous-key" || env["USER_SETTING"] != "before" {
		t.Fatalf("managed fields not restored: %#v", env)
	}
	if current["theme"] != "changed-after-apply" || current["user_option"] != "preserve" {
		t.Fatalf("later unrelated changes were lost: %#v", current)
	}
	if _, errStat := os.Stat(cliToolBackupPath(item.ConfigPath)); !os.IsNotExist(errStat) {
		t.Fatalf("backup was not removed after reset: %v", errStat)
	}
}

func TestCLIEndpointKeyCannotBeUsedForExternalEndpoint(t *testing.T) {
	key := "sk-local-endpoint"
	handler := &Handler{cfg: &config.Config{Port: 8317, SDKConfig: config.SDKConfig{APIKeys: []string{key}}}}
	resolved, errResolve := handler.resolveCLIEndpointKey("http://127.0.0.1:8317/v1", endpointKeyID(key), "")
	if errResolve != nil || resolved != key {
		t.Fatalf("local endpoint key = %q, %v", resolved, errResolve)
	}
	if _, errResolve = handler.resolveCLIEndpointKey("https://api.example.com/v1", endpointKeyID(key), ""); errResolve == nil {
		t.Fatal("saved local key was accepted for an external endpoint")
	}
}

func TestExtraCLIToolAdaptersPreserveUnrelatedSettings(t *testing.T) {
	for _, tool := range []string{"opencode", "openclaw", "droid", "hermes", "kilo", "deepseek-tui", "grok-build"} {
		t.Run(tool, func(t *testing.T) {
			config := map[string]any{"user_setting": "keep"}
			updated, errUpdate := extraToolConfig(tool, "http://127.0.0.1:8317/v1", "test-key", "test-main", "test-subagent", config)
			if errUpdate != nil {
				t.Fatal(errUpdate)
			}
			if updated["user_setting"] != "keep" {
				t.Fatalf("unrelated setting changed: %#v", updated)
			}
			current := extraToolCurrent(tool, updated)
			if current.Model != "test-main" || current.BaseURL != "http://127.0.0.1:8317/v1" {
				t.Fatalf("saved endpoint/model were not readable: %#v", current)
			}
			if (tool == "opencode" || tool == "grok-build") && current.SubagentModel != "test-subagent" {
				t.Fatalf("saved subagent model = %q", current.SubagentModel)
			}
		})
	}
}

func TestGetCLIToolsDoesNotLeakConfigurationPathsOrSecrets(t *testing.T) {
	gin.SetMode(gin.TestMode)
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	path := filepath.Join(home, ".claude", "settings.json")
	if errDir := os.MkdirAll(filepath.Dir(path), 0o700); errDir != nil {
		t.Fatal(errDir)
	}
	if errWrite := os.WriteFile(path, []byte(`{"env":{"ANTHROPIC_BASE_URL":"http://127.0.0.1:8317/v1","ANTHROPIC_AUTH_TOKEN":"private"}}`), 0o600); errWrite != nil {
		t.Fatal(errWrite)
	}
	response := httptest.NewRecorder()
	ctx, _ := gin.CreateTestContext(response)
	ctx.Request = httptest.NewRequest(http.MethodGet, "/v0/management/cli-tools", nil)
	(&Handler{}).GetCLITools(ctx)
	if strings.Contains(response.Body.String(), "private") || strings.Contains(response.Body.String(), path) {
		t.Fatalf("CLI status exposed local secrets or paths: %s", response.Body.String())
	}
}

package managementasset

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestSourceManagementAssets(t *testing.T) {
	for _, name := range []string{
		"index.html", "app.css", "app.js", "tools.js", "i18n/en.json", "i18n/vi.json",
		"providers/antigravity.png", "providers/codex.png", "providers/claude.png",
		"providers/gemini.png", "providers/openai.png", "providers/opencode.png",
		"providers/kiro.png", "providers/hermes.png", "providers/droid.png",
	} {
		body, contentType, ok := SourceManagementAsset(name)
		if !ok || len(body) == 0 || contentType == "" {
			t.Fatalf("asset %q unavailable", name)
		}
	}
	if _, _, ok := SourceManagementAsset("../config.yaml"); ok {
		t.Fatal("unexpected path traversal asset")
	}
}

func TestSourceManagementTranslationsHaveMatchingKeys(t *testing.T) {
	read := func(name string) map[string]string {
		t.Helper()
		body, _, ok := SourceManagementAsset(name)
		if !ok {
			t.Fatalf("translation %q unavailable", name)
		}
		var values map[string]string
		if errDecode := json.Unmarshal(body, &values); errDecode != nil {
			t.Fatalf("decode %q: %v", name, errDecode)
		}
		return values
	}

	english := read("i18n/en.json")
	vietnamese := read("i18n/vi.json")
	for key := range english {
		if _, ok := vietnamese[key]; !ok {
			t.Errorf("Vietnamese translation missing %q", key)
		}
	}
	for key := range vietnamese {
		if _, ok := english[key]; !ok {
			t.Errorf("English translation missing %q", key)
		}
	}

	for _, scriptName := range []string{"app.js", "tools.js"} {
		script, _, ok := SourceManagementAsset(scriptName)
		if !ok {
			t.Fatalf("%s unavailable", scriptName)
		}
		keyPattern := regexp.MustCompile(`(?:^|[^[:alnum:]_$])t\("([^"]+)"\)`)
		for _, match := range keyPattern.FindAllSubmatch(script, -1) {
			key := string(match[1])
			if _, exists := english[key]; !exists {
				t.Errorf("%s references missing translation %q", scriptName, key)
			}
		}
	}
	for _, key := range []string{
		"tools.guide.copilot", "tools.guide.cursor", "tools.guide.cline", "tools.guide.continue",
		"tools.guide.continue-dev", "tools.guide.roo", "tools.guide.amp", "tools.guide.qwen-code", "tools.guide.opendesign",
		"tools.slot.fable", "tools.slot.opus", "tools.slot.sonnet", "tools.slot.haiku",
	} {
		if _, exists := english[key]; !exists {
			t.Errorf("missing dynamic tools translation %q", key)
		}
		if _, exists := vietnamese[key]; !exists {
			t.Errorf("Vietnamese translation missing dynamic tools key %q", key)
		}
	}
}

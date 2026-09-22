package managementasset

import (
	"encoding/json"
	"regexp"
	"testing"
)

func TestSourceManagementAssets(t *testing.T) {
	for _, name := range []string{
		"index.html", "app.css", "app.js", "i18n/en.json", "i18n/vi.json",
		"providers/antigravity.png", "providers/codex.png", "providers/claude.png",
		"providers/gemini.png", "providers/openai.png",
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

	script, _, ok := SourceManagementAsset("app.js")
	if !ok {
		t.Fatal("app.js unavailable")
	}
	keyPattern := regexp.MustCompile(`(?:^|[^[:alnum:]_$])t\("([^"]+)"\)`)
	for _, match := range keyPattern.FindAllSubmatch(script, -1) {
		key := string(match[1])
		if _, exists := english[key]; !exists {
			t.Errorf("app.js references missing translation %q", key)
		}
	}
}

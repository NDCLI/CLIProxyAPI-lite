package management

import (
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestMutateCompatibleProviderAPIKey(t *testing.T) {
	weight := 4
	tests := []struct {
		name      string
		action    string
		index     *int
		key       string
		wantKeys  []config.OpenAICompatibilityAPIKey
		wantError bool
	}{
		{
			name:   "replace selected key and preserve its settings",
			action: "replace",
			index:  intPointer(0),
			key:    "replacement",
			wantKeys: []config.OpenAICompatibilityAPIKey{
				{APIKey: "replacement", Weight: &weight, ProxyURL: "http://proxy.local"},
				{APIKey: "second"},
			},
		},
		{
			name:   "append explicitly",
			action: "append",
			key:    "third",
			wantKeys: []config.OpenAICompatibilityAPIKey{
				{APIKey: "first", Weight: &weight, ProxyURL: "http://proxy.local"},
				{APIKey: "second"},
				{APIKey: "third"},
			},
		},
		{
			name: "preserve legacy append behavior",
			key:  "third",
			wantKeys: []config.OpenAICompatibilityAPIKey{
				{APIKey: "first", Weight: &weight, ProxyURL: "http://proxy.local"},
				{APIKey: "second"},
				{APIKey: "third"},
			},
		},
		{
			name:   "delete selected key",
			action: "delete",
			index:  intPointer(0),
			wantKeys: []config.OpenAICompatibilityAPIKey{
				{APIKey: "second"},
			},
		},
		{name: "disable first key only", action: "disable", index: intPointer(0), wantKeys: []config.OpenAICompatibilityAPIKey{{APIKey: "first", Weight: &weight, ProxyURL: "http://proxy.local", Disabled: true}, {APIKey: "second"}}},
		{name: "enable selected key", action: "enable", index: intPointer(1), wantKeys: []config.OpenAICompatibilityAPIKey{{APIKey: "first", Weight: &weight, ProxyURL: "http://proxy.local"}, {APIKey: "second"}}},
		{name: "reject toggle with secret", action: "disable", index: intPointer(0), key: "secret", wantError: true},
		{name: "reject toggle without index", action: "disable", wantError: true},
		{name: "reject replace without a key", action: "replace", index: intPointer(0), wantError: true},
		{name: "reject invalid index", action: "delete", index: intPointer(2), wantError: true},
		{name: "reject unknown action", action: "rotate", index: intPointer(0), key: "new", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entry := config.OpenAICompatibility{APIKeyEntries: []config.OpenAICompatibilityAPIKey{
				{APIKey: "first", Weight: &weight, ProxyURL: "http://proxy.local"},
				{APIKey: "second"},
			}}
			if err := mutateCompatibleProviderAPIKey(&entry, test.action, test.index, test.key); (err != nil) != test.wantError {
				t.Fatalf("mutateCompatibleProviderAPIKey() error = %v, wantError %v", err, test.wantError)
			}
			if test.wantError {
				return
			}
			if len(entry.APIKeyEntries) != len(test.wantKeys) {
				t.Fatalf("APIKeyEntries len = %d, want %d", len(entry.APIKeyEntries), len(test.wantKeys))
			}
			for index := range test.wantKeys {
				got, want := entry.APIKeyEntries[index], test.wantKeys[index]
				if got.APIKey != want.APIKey || got.ProxyURL != want.ProxyURL || got.Disabled != want.Disabled {
					t.Errorf("APIKeyEntries[%d] = %+v, want %+v", index, got, want)
				}
				if (got.Weight == nil) != (want.Weight == nil) || (got.Weight != nil && *got.Weight != *want.Weight) {
					t.Errorf("APIKeyEntries[%d].Weight = %v, want %v", index, got.Weight, want.Weight)
				}
			}
		})
	}
}

func intPointer(value int) *int { return &value }

func TestCompatibleProviderKeyTogglePersistsWithoutChangingSiblings(t *testing.T) {
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte("openai-compatibility: []\n"), 0o600); err != nil {
		t.Fatal(err)
	}
	cfg := &config.Config{OpenAICompatibility: []config.OpenAICompatibility{{Name: "test", BaseURL: "https://example.test/v1", APIKeyEntries: []config.OpenAICompatibilityAPIKey{{APIKey: "first"}, {APIKey: "second"}}}}}
	h := NewHandler(cfg, path, nil)
	for _, action := range []string{"disable", "enable"} {
		response := endpointKeyRequest(t, h.PatchCompatibleProvider, http.MethodPatch, "/provider-configs/test", `{"api_key_action":"`+action+`","api_key_index":0}`, compatibleProviderID("test"))
		if response.Code != http.StatusOK {
			t.Fatalf("toggle status = %d: %s", response.Code, response.Body.String())
		}
		var result struct {
			Item compatibleProviderItem `json:"item"`
		}
		if err := json.Unmarshal(response.Body.Bytes(), &result); err != nil {
			t.Fatal(err)
		}
		if len(result.Item.APIKeyEnabled) != 2 || result.Item.APIKeyEnabled[0] != (action == "enable") || !result.Item.APIKeyEnabled[1] {
			t.Fatalf("wrong key states: %v", result.Item.APIKeyEnabled)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		loaded, err := config.ParseConfigBytes(data)
		if err != nil {
			t.Fatal(err)
		}
		keys := loaded.OpenAICompatibility[0].APIKeyEntries
		if len(keys) != 2 || keys[0].Disabled != (action == "disable") || keys[1].Disabled || keys[0].APIKey != "first" || keys[1].APIKey != "second" {
			t.Fatal("saved toggle changed keys or sibling state")
		}
	}
}

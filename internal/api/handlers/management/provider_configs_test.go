package management

import (
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
				if got.APIKey != want.APIKey || got.ProxyURL != want.ProxyURL {
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

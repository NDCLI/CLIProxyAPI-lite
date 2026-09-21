package management

import "testing"

func TestParseAntigravityCallbackPayload(t *testing.T) {
	tests := []struct {
		name        string
		payload     map[string]string
		expected    string
		wantCode    string
		wantError   bool
	}{
		{
			name:      "empty state",
			payload:   map[string]string{"code": "test-code"},
			expected:  "test-state",
			wantError: true,
		},
		{
			name:      "mismatched state",
			payload:   map[string]string{"state": "other-state", "code": "test-code"},
			expected:  "test-state",
			wantError: true,
		},
		{
			name:      "valid callback",
			payload:   map[string]string{"state": "test-state", "code": "test-code"},
			expected:  "test-state",
			wantCode:  "test-code",
		},
		{
			name:      "missing authorization code",
			payload:   map[string]string{"state": "test-state"},
			expected:  "test-state",
			wantError: true,
		},
		{
			name:      "provider error",
			payload:   map[string]string{"state": "test-state", "error": "access_denied"},
			expected:  "test-state",
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			code, errParse := parseAntigravityCallbackPayload(tt.payload, tt.expected)
			if tt.wantError {
				if errParse == nil {
					t.Fatal("expected callback payload to be rejected")
				}
				return
			}
			if errParse != nil {
				t.Fatalf("expected callback payload to be accepted: %v", errParse)
			}
			if code != tt.wantCode {
				t.Fatalf("unexpected authorization code: got %q want %q", code, tt.wantCode)
			}
		})
	}
}

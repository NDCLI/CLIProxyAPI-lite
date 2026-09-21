package config

import (
	"path/filepath"
	"testing"
)

func TestConfigHostDefaultsToLoopback(t *testing.T) {
	tests := []struct {
		name string
		yaml string
	}{
		{name: "missing host", yaml: "port: 8317\n"},
		{name: "explicit empty host", yaml: "host: \"\"\n"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg, errParse := ParseConfigBytes([]byte(tt.yaml))
			if errParse != nil {
				t.Fatalf("ParseConfigBytes() error = %v", errParse)
			}
			if cfg.Host != DefaultHost {
				t.Fatalf("unexpected host: got %q want %q", cfg.Host, DefaultHost)
			}
		})
	}
}

func TestLoadConfigOptionalHostDefaultsToLoopback(t *testing.T) {
	missingPath := filepath.Join(t.TempDir(), "missing-config.yaml")
	cfg, errLoad := LoadConfigOptional(missingPath, true)
	if errLoad != nil {
		t.Fatalf("LoadConfigOptional() error = %v", errLoad)
	}
	if cfg.Host != DefaultHost {
		t.Fatalf("unexpected host: got %q want %q", cfg.Host, DefaultHost)
	}
}

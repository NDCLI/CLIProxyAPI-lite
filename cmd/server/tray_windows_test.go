//go:build windows

package main

import (
	"encoding/base64"
	"encoding/binary"
	"testing"
	"unicode/utf16"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestEncodePowerShell(t *testing.T) {
	encoded, err := base64.StdEncoding.DecodeString(encodePowerShell(trayPowerShellScript))
	if err != nil {
		t.Fatalf("decode encoded script: %v", err)
	}
	if len(encoded)%2 != 0 {
		t.Fatalf("encoded script has odd UTF-16 byte length: %d", len(encoded))
	}
	units := make([]uint16, len(encoded)/2)
	for i := range units {
		units[i] = binary.LittleEndian.Uint16(encoded[i*2:])
	}
	if got := string(utf16.Decode(units)); got != trayPowerShellScript {
		t.Fatal("PowerShell command encoding did not round-trip")
	}
}

func TestTrayManagementURL(t *testing.T) {
	cases := []struct {
		name string
		cfg  *config.Config
		want string
	}{
		{name: "default host", cfg: &config.Config{Port: 8317}, want: "http://127.0.0.1:8317/management.html"},
		{name: "wildcard host", cfg: &config.Config{Host: "0.0.0.0", Port: 9000}, want: "http://127.0.0.1:9000/management.html"},
		{name: "ipv6 tls", cfg: &config.Config{Host: "::1", Port: 443, TLS: config.TLSConfig{Enable: true}}, want: "https://[::1]:443/management.html"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := trayManagementURL(tc.cfg); got != tc.want {
				t.Fatalf("trayManagementURL() = %q, want %q", got, tc.want)
			}
		})
	}
}

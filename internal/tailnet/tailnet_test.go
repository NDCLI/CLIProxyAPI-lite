package tailnet

import (
	"net/url"
	"testing"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"tailscale.com/tsnet"
)

func TestIsOwnEndpointRequiresActiveListener(t *testing.T) {
	m := &Manager{
		srv: &tsnet.Server{}, cfg: config.TailscaleConfig{Hostname: "proxy"},
		dnsName: "proxy.example.ts.net", ips: []string{"100.64.0.1", "fd7a:115c:a1e0::1"},
		urls: []string{"http://proxy.example.ts.net"},
	}
	for _, tc := range []struct {
		value string
		want  bool
	}{
		{"http://proxy.example.ts.net/v1", true},
		{"http://PROXY.EXAMPLE.TS.NET:80", true},
		{"http://proxy", true},
		{"http://100.64.0.1/v1", true},
		{"http://[fd7a:115c:a1e0::1]/v1", true},
		{"http://proxy.example.ts.net:9999/v1", false},
		{"https://proxy.example.ts.net/v1", false},
		{"https://proxy.example.ts.net:80/v1", false},
		{"ftp://proxy.example.ts.net", false},
		{"http://another.example.ts.net/v1", false},
		{"http://proxy.example.ts.net.attacker.test/v1", false},
	} {
		t.Run(tc.value, func(t *testing.T) {
			parsed, errParse := url.Parse(tc.value)
			if errParse != nil {
				t.Fatal(errParse)
			}
			if got := m.IsOwnEndpoint(parsed); got != tc.want {
				t.Fatalf("IsOwnEndpoint(%q) = %v; want %v", tc.value, got, tc.want)
			}
		})
	}
	parsed, _ := url.Parse("https://proxy.example.ts.net:443/v1")
	m.urls = append(m.urls, "https://proxy.example.ts.net")
	if !m.IsOwnEndpoint(parsed) {
		t.Fatal("active HTTPS listener was rejected")
	}
	m.srv = nil
	if m.IsOwnEndpoint(parsed) {
		t.Fatal("stopped node was accepted")
	}
}

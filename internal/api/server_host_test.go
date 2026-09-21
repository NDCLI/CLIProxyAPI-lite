package api

import (
	"testing"

	proxyconfig "github.com/router-for-me/CLIProxyAPI/v7/internal/config"
)

func TestNewServerDefaultsEmptyHostToLoopback(t *testing.T) {
	server := newTestServer(t)
	if server.server == nil {
		t.Fatal("expected HTTP server to be initialized")
	}
	wantAddr := proxyconfig.DefaultHost + ":0"
	if server.server.Addr != wantAddr {
		t.Fatalf("unexpected server address: got %q want %q", server.server.Addr, wantAddr)
	}
}

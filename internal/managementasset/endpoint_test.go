package managementasset

import (
	"strings"
	"testing"
)

func TestAddEndpointCard(t *testing.T) {
	got, ok := AddEndpointCard([]byte("<html><body><main>panel</main></body></html>"))
	if !ok || !strings.Contains(string(got), "cliproxy-endpoint-card") || !strings.Contains(string(got), "onConfigTab") {
		t.Fatalf("endpoint card was not added: %s", got)
	}
}

func TestAddEndpointCardIsIdempotent(t *testing.T) {
	body := []byte("<body></body>")
	first, ok := AddEndpointCard(body)
	if !ok {
		t.Fatal("first injection should succeed")
	}
	second, ok := AddEndpointCard(first)
	if ok || string(second) != string(first) {
		t.Fatal("second injection should be a no-op")
	}
}

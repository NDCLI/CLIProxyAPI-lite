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
	first, ok := AddEndpointCard([]byte("<body></body>"))
	if !ok {
		t.Fatal("first injection should succeed")
	}
	second, ok := AddEndpointCard(first)
	if ok || string(second) != string(first) {
		t.Fatal("second injection should be a no-op")
	}
}

func TestAddRouterNavigation(t *testing.T) {
	got, ok := AddRouterNavigation([]byte("<html><head></head><body></body></html>"))
	if !ok || !strings.Contains(string(got), "cliproxy-router-nav") || !strings.Contains(string(got), "Combo & Vision Adapter") {
		t.Fatalf("router navigation was not added: %s", got)
	}
}

func TestAddRouterNavigationIsIdempotent(t *testing.T) {
	first, ok := AddRouterNavigation([]byte("<head></head>"))
	if !ok {
		t.Fatal("first injection should succeed")
	}
	second, ok := AddRouterNavigation(first)
	if ok || string(second) != string(first) {
		t.Fatal("second injection should be a no-op")
	}
}

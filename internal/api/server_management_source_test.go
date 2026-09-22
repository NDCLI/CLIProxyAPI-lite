package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSourceManagementRoutes(t *testing.T) {
	server := newTestServer(t)

	tests := []struct {
		path        string
		contentType string
		contains    string
	}{
		{path: "/management-next.html", contentType: "text/html", contains: `id="app"`},
		{path: "/management-next/app.css", contentType: "text/css", contains: ".app-shell"},
		{path: "/management-next/app.js", contentType: "text/javascript", contains: "loadCapabilities"},
		{path: "/management-next/i18n/en.json", contentType: "application/json", contains: `"app.name"`},
		{path: "/management-next/i18n/vi.json", contentType: "application/json", contains: `"app.name"`},
	}

	for _, test := range tests {
		t.Run(test.path, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, test.path, nil)
			server.engine.ServeHTTP(recorder, request)
			if recorder.Code != http.StatusOK {
				t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
			}
			if got := recorder.Header().Get("Content-Type"); !strings.HasPrefix(got, test.contentType) {
				t.Fatalf("Content-Type = %q, want prefix %q", got, test.contentType)
			}
			if !strings.Contains(recorder.Body.String(), test.contains) {
				t.Fatalf("response does not contain %q", test.contains)
			}
		})
	}
}

func TestSourceManagementRejectsUnknownAsset(t *testing.T) {
	server := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/management-next/unknown.js", nil)
	server.engine.ServeHTTP(recorder, request)
	if recorder.Code != http.StatusNotFound {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusNotFound)
	}
}

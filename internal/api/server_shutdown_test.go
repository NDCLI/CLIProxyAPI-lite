package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestManagementShutdownRequiresAuthentication(t *testing.T) {
	t.Setenv("MANAGEMENT_PASSWORD", "shutdown-test-key")
	server := newTestServer(t)
	for _, key := range []string{"", "shutdown-test-key", "shutdown-test-key"} {
		req := httptest.NewRequest(http.MethodPost, "/v0/management/shutdown", nil)
		req.RemoteAddr = "127.0.0.1:1234"
		if key != "" {
			req.Header.Set("Authorization", "Bearer "+key)
		}
		recorder := httptest.NewRecorder()
		server.engine.ServeHTTP(recorder, req)
		if key == "" {
			if recorder.Code != http.StatusUnauthorized {
				t.Fatalf("unauthenticated shutdown status = %d", recorder.Code)
			}
			select {
			case <-server.ShutdownRequested():
				t.Fatal("unauthenticated request triggered shutdown")
			default:
			}
			continue
		}
		if recorder.Code != http.StatusAccepted || !recorder.Flushed {
			t.Fatalf("shutdown response = %d, flushed = %t", recorder.Code, recorder.Flushed)
		}
		select {
		case <-server.ShutdownRequested():
		default:
			t.Fatal("authenticated request did not trigger shutdown")
		}
	}
}

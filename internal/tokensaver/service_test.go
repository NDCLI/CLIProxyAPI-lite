package tokensaver

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
)

func longToolOutput() string { return strings.Repeat("repeated tool output line\n", 65) }

func TestRTKCompressesToolResultsAcrossRequestFormats(t *testing.T) {
	toolText := longToolOutput()
	tests := []struct {
		name     string
		payload  map[string]any
		toolPath string
	}{
		{name: "OpenAI Chat", payload: map[string]any{"messages": []any{map[string]any{"role": "user", "content": toolText}, map[string]any{"role": "tool", "content": toolText}}}, toolPath: "messages.1.content"},
		{name: "Anthropic Messages", payload: map[string]any{"messages": []any{map[string]any{"role": "user", "content": []any{map[string]any{"type": "tool_result", "tool_use_id": "call-1", "content": toolText}, map[string]any{"type": "tool_result", "tool_use_id": "call-2", "is_error": true, "content": toolText}}}}}, toolPath: "messages.0.content.0.content"},
		{name: "OpenAI Responses", payload: map[string]any{"input": []any{map[string]any{"type": "function_call_output", "call_id": "call-1", "output": toolText}}}, toolPath: "input.0.output"},
		{name: "Gemini", payload: map[string]any{"contents": []any{map[string]any{"role": "user", "parts": []any{map[string]any{"functionResponse": map[string]any{"name": "read", "response": map[string]any{"output": toolText}}}}}}}, toolPath: "contents.0.parts.0.functionResponse.response.output"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			payload, errMarshal := json.Marshal(test.payload)
			if errMarshal != nil {
				t.Fatal(errMarshal)
			}
			service := NewService(config.TokenSaverConfig{RTKEnabled: true})
			compressed := service.Process(context.Background(), http.Header{}, payload)
			if len(compressed) >= len(payload) || !strings.Contains(gjson.GetBytes(compressed, test.toolPath).String(), "duplicate lines omitted") {
				t.Fatalf("tool result was not compacted: before=%d after=%d", len(payload), len(compressed))
			}
			if got := service.Snapshot().Statistics.RTKHits; got != 1 {
				t.Fatalf("RTK hits = %d, want 1", got)
			}
			if test.name == "OpenAI Chat" && gjson.GetBytes(compressed, "messages.0.content").String() != toolText {
				t.Fatal("user content changed")
			}
			if test.name == "Anthropic Messages" && gjson.GetBytes(compressed, "messages.0.content.1.content").String() != toolText {
				t.Fatal("error trace changed")
			}
		})
	}
}

func TestMiddlewareHonorsRequestOptOutAndKeepsContentLength(t *testing.T) {
	gin.SetMode(gin.TestMode)
	service := NewService(config.TokenSaverConfig{RTKEnabled: true})
	router := gin.New()
	router.POST("/v1/chat/completions", service.Middleware(), func(c *gin.Context) {
		body, _ := io.ReadAll(c.Request.Body)
		if c.Request.ContentLength != int64(len(body)) {
			t.Errorf("content length = %d, body length = %d", c.Request.ContentLength, len(body))
		}
		c.Data(http.StatusOK, "application/json", body)
	})
	input, _ := json.Marshal(map[string]any{"model": "example", "messages": []any{map[string]any{"role": "tool", "content": longToolOutput()}}})
	request := func(bypass bool) []byte {
		t.Helper()
		recorder := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodPost, "/v1/chat/completions", bytes.NewReader(input))
		req.Header.Set("Content-Type", "application/json")
		if bypass {
			req.Header.Set("X-CLIProxy-Token-Saver", "off")
		}
		router.ServeHTTP(recorder, req)
		if recorder.Code != http.StatusOK {
			t.Fatalf("status = %d", recorder.Code)
		}
		return recorder.Body.Bytes()
	}
	if result := request(false); len(result) >= len(input) {
		t.Fatalf("RTK did not reduce request body: %d >= %d", len(result), len(input))
	}
	if result := request(true); !bytes.Equal(result, input) {
		t.Fatal("per-request opt-out changed the body")
	}
	if service.Snapshot().Statistics.PerRequestBypasses != 1 {
		t.Fatal("per-request bypass was not counted")
	}
}

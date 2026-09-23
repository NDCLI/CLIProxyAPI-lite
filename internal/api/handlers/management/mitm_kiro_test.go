package management

import (
	"encoding/binary"
	"encoding/json"
	"hash/crc32"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranslateKiroMITMRequestPreservesHistoryToolsAndMapping(t *testing.T) {
	input := `{"conversationState":{"history":[{"userInputMessage":{"content":"inspect"}},{"assistantResponseMessage":{"content":"checking"}}],"currentMessage":{"userInputMessage":{"content":"open it","userInputMessageContext":{"tools":[{"toolSpecification":{"name":"read_file","description":"Read a file","inputSchema":{"json":{"type":"object"}}}}]}}}}}`
	data, errConvert := translateKiroMITMRequest([]byte(input), "openai/gpt-5")
	if errConvert != nil {
		t.Fatal(errConvert)
	}
	var body struct {
		Model    string           `json:"model"`
		Stream   bool             `json:"stream"`
		Messages []map[string]any `json:"messages"`
		Tools    []map[string]any `json:"tools"`
	}
	if errDecode := json.Unmarshal(data, &body); errDecode != nil {
		t.Fatal(errDecode)
	}
	if body.Model != "openai/gpt-5" || !body.Stream || len(body.Messages) != 3 || body.Messages[0]["role"] != "user" || body.Messages[1]["role"] != "assistant" || body.Messages[2]["content"] != "open it" || len(body.Tools) != 1 {
		t.Fatalf("converted Kiro request = %#v", body)
	}
}

func TestKiroEventFrameHasValidLengthsAndChecksums(t *testing.T) {
	frame := kiroEventFrame("assistantResponseEvent", map[string]string{"content": "hello"}, "application/json")
	total := int(binary.BigEndian.Uint32(frame[:4]))
	headerSize := int(binary.BigEndian.Uint32(frame[4:8]))
	if total != len(frame) || 12+headerSize+4 > len(frame) {
		t.Fatalf("invalid event frame sizes: total=%d headers=%d bytes=%d", total, headerSize, len(frame))
	}
	if got, want := binary.BigEndian.Uint32(frame[8:12]), crc32.ChecksumIEEE(frame[:8]); got != want {
		t.Fatalf("prelude crc = %08x, want %08x", got, want)
	}
	if got, want := binary.BigEndian.Uint32(frame[len(frame)-4:]), crc32.ChecksumIEEE(frame[:len(frame)-4]); got != want {
		t.Fatalf("message crc = %08x, want %08x", got, want)
	}
}

func TestProxyKiroMITMResponseWrapsOpenAITextInEventStream(t *testing.T) {
	response := &http.Response{StatusCode: http.StatusOK, Header: make(http.Header), Body: http.NoBody}
	response.Body = io.NopCloser(strings.NewReader("data: {\"model\":\"model-a\",\"choices\":[{\"delta\":{\"content\":\"hello\"},\"finish_reason\":null}]}\n\ndata: [DONE]\n\n"))
	recorder := httptest.NewRecorder()
	if errProxy := proxyKiroMITMResponse(recorder, response, "model-a"); errProxy != nil {
		t.Fatal(errProxy)
	}
	body := recorder.Body.Bytes()
	if recorder.Header().Get("Content-Type") != "application/vnd.amazon.eventstream" || len(body) < 16 || !strings.Contains(string(body), "assistantResponseEvent") || !strings.Contains(string(body), "messageStopEvent") {
		t.Fatalf("Kiro response content-type=%q bytes=%d", recorder.Header().Get("Content-Type"), len(body))
	}
}

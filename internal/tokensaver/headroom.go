package tokensaver

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strconv"
	"strings"
	"time"

	"github.com/router-for-me/CLIProxyAPI/v7/internal/config"
	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

const maxHeadroomResponseBytes = 16 << 20

type Health struct {
	OK        bool   `json:"ok"`
	Status    string `json:"status"`
	LatencyMS int64  `json:"latency_ms"`
}

func (s *Service) CheckHeadroom(ctx context.Context) Health {
	value := s.currentConfig()
	started := time.Now()
	checkCtx, cancel := context.WithTimeout(ctx, time.Duration(value.HeadroomTimeoutMS)*time.Millisecond)
	defer cancel()
	request, errRequest := http.NewRequestWithContext(checkCtx, http.MethodGet, value.HeadroomURL+"/health", nil)
	if errRequest != nil {
		s.setHeadroomStatus("invalid_url")
		return Health{Status: "invalid_url"}
	}
	response, errDo := s.client.Do(request)
	if errDo != nil {
		status := headroomErrorCode(errDo)
		s.setHeadroomStatus(status)
		return Health{Status: status, LatencyMS: time.Since(started).Milliseconds()}
	}
	defer response.Body.Close()
	status := "ready"
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		status = "http_" + strconv.Itoa(response.StatusCode)
	}
	s.setHeadroomStatus(status)
	return Health{OK: status == "ready", Status: status, LatencyMS: time.Since(started).Milliseconds()}
}

func (s *Service) compressHeadroom(ctx context.Context, value config.TokenSaverConfig, body []byte) ([]byte, string) {
	messages := gjson.GetBytes(body, "messages")
	model := strings.TrimSpace(gjson.GetBytes(body, "model").String())
	if !messages.IsArray() || model == "" || len(messages.Raw) < minToolTextBytes {
		return body, ""
	}
	payload, errMarshal := json.Marshal(struct {
		Messages json.RawMessage `json:"messages"`
		Model    string          `json:"model"`
		Config   struct {
			Mode                 string `json:"mode"`
			CompressUserMessages bool   `json:"compress_user_messages"`
		} `json:"config"`
	}{Messages: json.RawMessage(messages.Raw), Model: model, Config: struct {
		Mode                 string `json:"mode"`
		CompressUserMessages bool   `json:"compress_user_messages"`
	}{Mode: "lossy_inline"}})
	if errMarshal != nil {
		return body, "invalid_request"
	}

	// Headroom is a short preprocessing call; bound it so a stopped sidecar cannot hold an LLM request.
	compressCtx, cancel := context.WithTimeout(ctx, time.Duration(value.HeadroomTimeoutMS)*time.Millisecond)
	defer cancel()
	request, errRequest := http.NewRequestWithContext(compressCtx, http.MethodPost, value.HeadroomURL+"/v1/compress", bytes.NewReader(payload))
	if errRequest != nil {
		return body, "invalid_url"
	}
	request.Header.Set("Content-Type", "application/json")
	response, errDo := s.client.Do(request)
	if errDo != nil {
		return body, headroomErrorCode(errDo)
	}
	defer response.Body.Close()
	if response.StatusCode < http.StatusOK || response.StatusCode >= http.StatusMultipleChoices {
		return body, "http_" + strconv.Itoa(response.StatusCode)
	}
	responseBody, errRead := io.ReadAll(io.LimitReader(response.Body, maxHeadroomResponseBytes+1))
	if errRead != nil || len(responseBody) > maxHeadroomResponseBytes || !json.Valid(responseBody) {
		return body, "invalid_response"
	}
	if gjson.GetBytes(responseBody, "compression_skipped").Bool() {
		return body, ""
	}
	compressed := gjson.GetBytes(responseBody, "messages")
	if !compatibleHeadroomMessages(messages, compressed) {
		return body, "invalid_response"
	}
	updated, errSet := sjson.SetRawBytes(body, "messages", []byte(compressed.Raw))
	if errSet != nil {
		return body, "invalid_response"
	}
	if len(updated) >= len(body) {
		return body, ""
	}
	return updated, ""
}

func headroomErrorCode(err error) string {
	if errors.Is(err, context.DeadlineExceeded) {
		return "timeout"
	}
	if errors.Is(err, context.Canceled) {
		return "canceled"
	}
	return "unreachable"
}

func compatibleHeadroomMessages(original, candidate gjson.Result) bool {
	if !original.IsArray() || !candidate.IsArray() {
		return false
	}
	before, after := original.Array(), candidate.Array()
	if len(before) != len(after) {
		return false
	}
	for index := range before {
		if !compatibleHeadroomMessage(before[index], after[index]) {
			return false
		}
	}
	return true
}

func compatibleHeadroomMessage(original, candidate gjson.Result) bool {
	role := original.Get("role").String()
	if role == "" || candidate.Get("role").String() != role {
		return false
	}
	var before, after map[string]any
	if err := json.Unmarshal([]byte(original.Raw), &before); err != nil {
		return false
	}
	if err := json.Unmarshal([]byte(candidate.Raw), &after); err != nil {
		return false
	}
	beforeContent, afterContent := before["content"], after["content"]
	delete(before, "content")
	delete(after, "content")
	if !reflect.DeepEqual(before, after) {
		return false
	}
	if role == "system" || role == "developer" {
		return reflect.DeepEqual(beforeContent, afterContent)
	}
	if text, ok := beforeContent.(string); ok {
		_, sameType := afterContent.(string)
		return sameType && (role != "user" || text == afterContent)
	}
	oldBlocks, oldArray := beforeContent.([]any)
	newBlocks, newArray := afterContent.([]any)
	if !oldArray || !newArray {
		return reflect.DeepEqual(beforeContent, afterContent)
	}
	if len(oldBlocks) != len(newBlocks) {
		return false
	}
	for index := range oldBlocks {
		oldBlock, oldMap := oldBlocks[index].(map[string]any)
		newBlock, newMap := newBlocks[index].(map[string]any)
		if !oldMap || !newMap {
			if !reflect.DeepEqual(oldBlocks[index], newBlocks[index]) {
				return false
			}
			continue
		}
		kind, _ := oldBlock["type"].(string)
		if kind != newBlock["type"] {
			return false
		}
		allowText := kind == "tool_result" && oldBlock["is_error"] != true || (role == "assistant" || role == "tool") && (kind == "text" || kind == "input_text")
		if !allowText {
			if !reflect.DeepEqual(oldBlock, newBlock) {
				return false
			}
			continue
		}
		field := "text"
		if kind == "tool_result" {
			field = "content"
		}
		oldText, newText := oldBlock[field], newBlock[field]
		delete(oldBlock, field)
		delete(newBlock, field)
		if !reflect.DeepEqual(oldBlock, newBlock) || !sameTextShape(oldText, newText) {
			return false
		}
	}
	return true
}

func sameTextShape(before, after any) bool {
	if _, ok := before.(string); ok {
		_, valid := after.(string)
		return valid
	}
	oldParts, oldArray := before.([]any)
	newParts, newArray := after.([]any)
	if !oldArray || !newArray || len(oldParts) != len(newParts) {
		return reflect.DeepEqual(before, after)
	}
	for index := range oldParts {
		oldPart, oldMap := oldParts[index].(map[string]any)
		newPart, newMap := newParts[index].(map[string]any)
		if !oldMap || !newMap {
			if !reflect.DeepEqual(oldParts[index], newParts[index]) {
				return false
			}
			continue
		}
		if oldPart["type"] != "text" || newPart["type"] != "text" {
			if !reflect.DeepEqual(oldPart, newPart) {
				return false
			}
			continue
		}
		if _, ok := oldPart["text"].(string); !ok {
			return false
		}
		if _, ok := newPart["text"].(string); !ok {
			return false
		}
		delete(oldPart, "text")
		delete(newPart, "text")
		if !reflect.DeepEqual(oldPart, newPart) {
			return false
		}
	}
	return true
}

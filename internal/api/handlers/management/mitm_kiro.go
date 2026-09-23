package management

import (
	"bufio"
	"bytes"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"hash/crc32"
	"net/http"
	"strings"
)

type kiroRequest struct {
	ConversationState struct {
		History        []map[string]json.RawMessage `json:"history"`
		CurrentMessage map[string]json.RawMessage   `json:"currentMessage"`
	} `json:"conversationState"`
}

type kiroUserMessage struct {
	Content string `json:"content"`
	Context struct {
		ToolResults []struct {
			ToolUseID string `json:"toolUseId"`
			Content   []struct {
				Text string `json:"text"`
			} `json:"content"`
		} `json:"toolResults"`
		Tools []struct {
			ToolSpecification struct {
				Name        string `json:"name"`
				Description string `json:"description"`
				InputSchema struct {
					JSON any `json:"json"`
				} `json:"inputSchema"`
			} `json:"toolSpecification"`
		} `json:"tools"`
	} `json:"userInputMessageContext"`
}

type kiroAssistantMessage struct {
	Content  string `json:"content"`
	ToolUses []struct {
		ID    string `json:"toolUseId"`
		Name  string `json:"name"`
		Input any    `json:"input"`
	} `json:"toolUses"`
}

func translateKiroMITMRequest(original []byte, model string) ([]byte, error) {
	var request kiroRequest
	if errDecode := json.Unmarshal(original, &request); errDecode != nil {
		return nil, fmt.Errorf("decode Kiro request: %w", errDecode)
	}
	messages := make([]map[string]any, 0, len(request.ConversationState.History)+1)
	var tools any
	appendUser := func(raw json.RawMessage) error {
		var user kiroUserMessage
		if errDecode := json.Unmarshal(raw, &user); errDecode != nil {
			return errDecode
		}
		for _, result := range user.Context.ToolResults {
			parts := make([]string, 0, len(result.Content))
			for _, part := range result.Content {
				if part.Text != "" {
					parts = append(parts, part.Text)
				}
			}
			messages = append(messages, map[string]any{"role": "tool", "tool_call_id": result.ToolUseID, "content": strings.Join(parts, "\n")})
		}
		if user.Content != "" || len(user.Context.ToolResults) == 0 {
			messages = append(messages, map[string]any{"role": "user", "content": user.Content})
		}
		if len(user.Context.Tools) > 0 && tools == nil {
			converted := make([]map[string]any, 0, len(user.Context.Tools))
			for _, entry := range user.Context.Tools {
				spec := entry.ToolSpecification
				if spec.Name == "" {
					continue
				}
				parameters := spec.InputSchema.JSON
				if parameters == nil {
					parameters = map[string]any{"type": "object", "properties": map[string]any{}}
				}
				converted = append(converted, map[string]any{"type": "function", "function": map[string]any{"name": spec.Name, "description": spec.Description, "parameters": parameters}})
			}
			if len(converted) > 0 {
				tools = converted
			}
		}
		return nil
	}
	for _, entry := range request.ConversationState.History {
		if raw := entry["userInputMessage"]; len(raw) > 0 {
			if errAppend := appendUser(raw); errAppend != nil {
				return nil, errAppend
			}
			continue
		}
		if raw := entry["assistantResponseMessage"]; len(raw) > 0 {
			var assistant kiroAssistantMessage
			if errDecode := json.Unmarshal(raw, &assistant); errDecode != nil {
				return nil, errDecode
			}
			message := map[string]any{"role": "assistant", "content": assistant.Content}
			if len(assistant.ToolUses) > 0 {
				calls := make([]map[string]any, 0, len(assistant.ToolUses))
				for _, use := range assistant.ToolUses {
					args, errMarshal := json.Marshal(use.Input)
					if errMarshal != nil {
						return nil, errMarshal
					}
					calls = append(calls, map[string]any{"id": use.ID, "type": "function", "function": map[string]any{"name": use.Name, "arguments": string(args)}})
				}
				message["tool_calls"] = calls
			}
			messages = append(messages, message)
		}
	}
	if raw := request.ConversationState.CurrentMessage["userInputMessage"]; len(raw) > 0 {
		if errAppend := appendUser(raw); errAppend != nil {
			return nil, errAppend
		}
	}
	if len(messages) == 0 {
		return nil, fmt.Errorf("Kiro request contains no conversation messages")
	}
	body := map[string]any{"model": model, "messages": messages, "stream": true}
	if tools != nil {
		body["tools"], body["tool_choice"] = tools, "auto"
	}
	return json.Marshal(body)
}

func proxyKiroMITMResponse(w http.ResponseWriter, response *http.Response, model string) error {
	w.Header().Set("Content-Type", "application/vnd.amazon.eventstream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(response.StatusCode)
	flusher, _ := w.(http.Flusher)
	write := func(frame []byte) error {
		if _, errWrite := w.Write(frame); errWrite != nil {
			return errWrite
		}
		if flusher != nil {
			flusher.Flush()
		}
		return nil
	}
	if errWrite := write(kiroEventFrame("initial-response", map[string]any{"conversationId": ""}, "application/x-amz-json-1.0")); errWrite != nil {
		return errWrite
	}
	toolCalls := map[int]struct{ id, name string }{}
	terminal := false
	var streamErr error
	process := func(data []byte) error {
		var chunk struct {
			Model   string `json:"model"`
			Choices []struct {
				Delta struct {
					Content   any    `json:"content"`
					Reasoning string `json:"reasoning_content"`
					ToolCalls []struct {
						Index    int    `json:"index"`
						ID       string `json:"id"`
						Function struct {
							Name      string `json:"name"`
							Arguments string `json:"arguments"`
						} `json:"function"`
					} `json:"tool_calls"`
				} `json:"delta"`
				FinishReason string `json:"finish_reason"`
			} `json:"choices"`
			Usage struct {
				Input  int `json:"prompt_tokens"`
				Output int `json:"completion_tokens"`
			} `json:"usage"`
		}
		if errDecode := json.Unmarshal(data, &chunk); errDecode != nil {
			return errDecode
		}
		for _, choice := range chunk.Choices {
			delta := choice.Delta
			if delta.Reasoning != "" {
				if errWrite := write(kiroEventFrame("reasoningContentEvent", map[string]any{"content": delta.Reasoning, "modelId": model}, "application/json")); errWrite != nil {
					return errWrite
				}
			}
			if delta.Content != nil {
				content, _ := delta.Content.(string)
				if content != "" {
					if errWrite := write(kiroEventFrame("assistantResponseEvent", map[string]any{"content": content, "modelId": model}, "application/json")); errWrite != nil {
						return errWrite
					}
				}
			}
			for _, call := range delta.ToolCalls {
				known := toolCalls[call.Index]
				if call.ID != "" || call.Function.Name != "" {
					if call.ID != "" {
						known.id = call.ID
					}
					if call.Function.Name != "" {
						known.name = call.Function.Name
					}
					toolCalls[call.Index] = known
					if known.id != "" && known.name != "" {
						if errWrite := write(kiroEventFrame("toolUseEvent", map[string]any{"name": known.name, "toolUseId": known.id}, "application/json")); errWrite != nil {
							return errWrite
						}
					}
				}
				if call.Function.Arguments != "" {
					if errWrite := write(kiroEventFrame("toolUseEvent", map[string]any{"input": call.Function.Arguments, "name": known.name, "toolUseId": known.id}, "application/json")); errWrite != nil {
						return errWrite
					}
				}
			}
			if choice.FinishReason != "" && !terminal {
				if len(toolCalls) > 0 {
					for _, call := range toolCalls {
						if errWrite := write(kiroEventFrame("toolUseEvent", map[string]any{"name": call.name, "toolUseId": call.id, "stop": true}, "application/json")); errWrite != nil {
							return errWrite
						}
					}
				} else if errWrite := write(kiroEventFrame("messageStopEvent", map[string]any{}, "application/json")); errWrite != nil {
					return errWrite
				}
				terminal = true
			}
		}
		if chunk.Usage.Input > 0 || chunk.Usage.Output > 0 {
			return write(kiroEventFrame("usageEvent", map[string]any{"inputTokens": chunk.Usage.Input, "outputTokens": chunk.Usage.Output}, "application/json"))
		}
		return nil
	}
	scanner := bufio.NewScanner(response.Body)
	scanner.Buffer(make([]byte, 64<<10), 8<<20)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "data:") {
			continue
		}
		data := strings.TrimSpace(strings.TrimPrefix(line, "data:"))
		if data == "" || data == "[DONE]" {
			continue
		}
		if errProcess := process([]byte(data)); errProcess != nil {
			streamErr = errProcess
			break
		}
	}
	if streamErr != nil {
		return streamErr
	}
	if errScan := scanner.Err(); errScan != nil {
		return errScan
	}
	if !terminal {
		return write(kiroEventFrame("messageStopEvent", map[string]any{}, "application/json"))
	}
	return nil
}

func kiroEventFrame(event string, payload any, contentType string) []byte {
	body, _ := json.Marshal(payload)
	var headers bytes.Buffer
	for _, header := range [][2]string{{":message-type", "event"}, {":event-type", event}, {":content-type", contentType}} {
		_ = headers.WriteByte(byte(len(header[0])))
		_, _ = headers.WriteString(header[0])
		_ = headers.WriteByte(7)
		_ = binary.Write(&headers, binary.BigEndian, uint16(len(header[1])))
		_, _ = headers.WriteString(header[1])
	}
	total := 12 + headers.Len() + len(body) + 4
	frame := make([]byte, total)
	binary.BigEndian.PutUint32(frame[0:4], uint32(total))
	binary.BigEndian.PutUint32(frame[4:8], uint32(headers.Len()))
	binary.BigEndian.PutUint32(frame[8:12], crc32.ChecksumIEEE(frame[:8]))
	copy(frame[12:], headers.Bytes())
	copy(frame[12+headers.Len():], body)
	binary.BigEndian.PutUint32(frame[total-4:], crc32.ChecksumIEEE(frame[:total-4]))
	return frame
}

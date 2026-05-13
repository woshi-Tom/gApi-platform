package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

type ClaudeAdapter struct {
	client *http.Client
}

func NewClaudeAdapter() *ClaudeAdapter {
	return &ClaudeAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *ClaudeAdapter) GetName() string {
	return "Claude (Anthropic)"
}

func (a *ClaudeAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/messages", baseURL)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	payload := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": maxTokens,
	}

	// Extract system messages — Anthropic requires system as a top-level field, not in messages
	var systemParts []string
	messages := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m["role"] == "system" {
			systemParts = append(systemParts, m["content"])
		} else {
			messages = append(messages, map[string]string{
				"role":    m["role"],
				"content": m["content"],
			})
		}
	}
	if len(systemParts) > 0 {
		payload["system"] = strings.Join(systemParts, "\n")
	}
	payload["messages"] = messages

	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", channel.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")

	resp, err := DoChannelRequest(a.client, httpReq, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var claudeResp struct {
		ID         string `json:"id"`
		Type       string `json:"type"`
		Role       string `json:"role"`
		Content    []struct {
			Type string `json:"type"`
			Text string `json:"text"`
		} `json:"content"`
		Model       string `json:"model"`
		StopReason string `json:"stop_reason"`
		StopSequence interface{} `json:"stop_sequence"`
		Usage       struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
		Error *struct {
			Type    string `json:"type"`
			Message string `json:"message"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &claudeResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if claudeResp.Error != nil {
			return nil, fmt.Errorf("Claude API error: %s (status %d)", claudeResp.Error.Message, resp.StatusCode)
		}
		return nil, fmt.Errorf("Claude API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	content := ""
	for _, c := range claudeResp.Content {
		if c.Type == "text" {
			content = c.Text
			break
		}
	}

	return &ChatResponse{
		ID:      claudeResp.ID,
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   claudeResp.Model,
		Choices: []struct {
			Index        int `json:"index"`
			Message      struct {
				Role    string `json:"role"`
				Content string `json:"content"`
			} `json:"message"`
			FinishReason string `json:"finish_reason"`
		}{
			{
				Message: struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				}{
					Role:    "assistant",
					Content: content,
				},
				FinishReason: claudeResp.StopReason,
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     claudeResp.Usage.InputTokens,
			CompletionTokens: claudeResp.Usage.OutputTokens,
			TotalTokens:      claudeResp.Usage.InputTokens + claudeResp.Usage.OutputTokens,
		},
	}, nil
}

func (a *ClaudeAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.anthropic.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")
	url := fmt.Sprintf("%s/v1/messages", baseURL)

	maxTokens := req.MaxTokens
	if maxTokens == 0 {
		maxTokens = 1024
	}

	payload := map[string]interface{}{
		"model":      req.Model,
		"max_tokens": maxTokens,
		"stream":     true,
	}

	// Extract system messages (same as Chat method)
	var systemParts []string
	messages := make([]map[string]string, 0, len(req.Messages))
	for _, m := range req.Messages {
		if m["role"] == "system" {
			systemParts = append(systemParts, m["content"])
		} else {
			messages = append(messages, map[string]string{
				"role":    m["role"],
				"content": m["content"],
			})
		}
	}
	if len(systemParts) > 0 {
		payload["system"] = strings.Join(systemParts, "\n")
	}
	payload["messages"] = messages

	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("x-api-key", channel.APIKey)
	httpReq.Header.Set("anthropic-version", "2023-06-01")
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := DoChannelRequest(a.client, httpReq, channel)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Claude API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	ch := make(chan ChatStreamResponse, 100)
	go a.readStream(resp.Body, ch, resp)
	return ch, nil
}

func (a *ClaudeAdapter) readStream(body io.Reader, ch chan<- ChatStreamResponse, resp *http.Response) {
	defer close(ch)
	defer resp.Body.Close()

	reader := NewSSEReader(body)
	var currentEvent string

	for {
		line, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				ch <- ChatStreamResponse{
					Error: &StreamError{Message: fmt.Sprintf("stream error: %v", err)},
				}
			}
			return
		}

		trimmed := strings.TrimSpace(line)

		// Parse event type
		if strings.HasPrefix(trimmed, "event: ") {
			currentEvent = strings.TrimPrefix(trimmed, "event: ")
			continue
		}

		// Parse data
		if strings.HasPrefix(trimmed, "data: ") {
			data := strings.TrimPrefix(trimmed, "data: ")

			switch currentEvent {
			case "content_block_delta":
				var event struct {
					Delta struct {
						Type string `json:"type"`
						Text string `json:"text"`
					} `json:"delta"`
				}
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					if event.Delta.Type == "text_delta" && event.Delta.Text != "" {
						ch <- ChatStreamResponse{
							Choices: []struct {
								Index        int    `json:"index"`
								Delta        struct {
									Role    string `json:"role,omitempty"`
									Content string `json:"content,omitempty"`
								} `json:"delta"`
								FinishReason string `json:"finish_reason,omitempty"`
							}{
								{
									Delta: struct {
										Role    string `json:"role,omitempty"`
										Content string `json:"content,omitempty"`
									}{
										Content: event.Delta.Text,
									},
								},
							},
						}
					}
				}

			case "message_delta":
				var event struct {
					Delta struct {
						StopReason string `json:"stop_reason"`
					} `json:"delta"`
					Usage struct {
						OutputTokens int `json:"output_tokens"`
					} `json:"usage"`
				}
				if err := json.Unmarshal([]byte(data), &event); err == nil {
					finishReason := mapStopReasonReverseStream(event.Delta.StopReason)
					ch <- ChatStreamResponse{
						Choices: []struct {
							Index        int    `json:"index"`
							Delta        struct {
								Role    string `json:"role,omitempty"`
								Content string `json:"content,omitempty"`
							} `json:"delta"`
							FinishReason string `json:"finish_reason,omitempty"`
						}{
							{
								FinishReason: finishReason,
							},
						},
						Usage: &struct {
							PromptTokens     int `json:"prompt_tokens"`
							CompletionTokens int `json:"completion_tokens"`
							TotalTokens      int `json:"total_tokens"`
						}{
							CompletionTokens: event.Usage.OutputTokens,
						},
					}
				}

			case "message_stop":
				ch <- ChatStreamResponse{Done: true}
				return

			case "error":
				var errBody struct {
					Type    string `json:"type"`
					Message string `json:"message"`
				}
				if err := json.Unmarshal([]byte(data), &errBody); err == nil {
					ch <- ChatStreamResponse{
						Error: &StreamError{Message: errBody.Message},
					}
				}
				return
			}

			// Reset event type after processing data
			currentEvent = ""
		}

		// Empty line signals end of an SSE event block
		if trimmed == "" {
			currentEvent = ""
		}
	}
}

func mapStopReasonReverseStream(anthropicReason string) string {
	switch anthropicReason {
	case "end_turn":
		return "stop"
	case "max_tokens":
		return "length"
	case "stop_sequence":
		return "stop"
	case "tool_use":
		return "tool_calls"
	default:
		return "stop"
	}
}

func (a *ClaudeAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("Claude embeddings not supported")
}

func (a *ClaudeAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
	return &ModelsResponse{
		Object: "list",
		Data: []struct {
			ID         string `json:"id"`
			Object     string `json:"object"`
			Created    int    `json:"created"`
			OwnedBy    string `json:"owned_by"`
			Permission []struct {
				ID        string `json:"id"`
				Object    string `json:"object"`
				Created   int    `json:"created"`
				AllowCreateEngine bool `json:"allow_create_engine"`
				AllowSampling bool `json:"allow_sampling"`
				AllowLogprobs bool `json:"allow_logprobs"`
				AllowSearchIndices bool `json:"allow_search_indices"`
				AllowView bool `json:"allow_view"`
				AllowFineTuning bool `json:"allow_fine_tuning"`
				Organization string `json:"organization"`
				Group interface{} `json:"group"`
				IsBlocking bool `json:"is_blocking"`
			} `json:"permission,omitempty"`
		}{
			{ID: "claude-3-opus-20240229", Object: "model", Created: 1709594900, OwnedBy: "anthropic"},
			{ID: "claude-3-sonnet-20240229", Object: "model", Created: 1709594900, OwnedBy: "anthropic"},
			{ID: "claude-3-haiku-20240307", Object: "model", Created: 1709594900, OwnedBy: "anthropic"},
			{ID: "claude-2.1", Object: "model", Created: 1698375610, OwnedBy: "anthropic"},
			{ID: "claude-2.0", Object: "model", Created: 1689975900, OwnedBy: "anthropic"},
			{ID: "claude-instant-1.2", Object: "model", Created: 1689975900, OwnedBy: "anthropic"},
		},
	}, nil
}

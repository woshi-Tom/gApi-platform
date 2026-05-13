package handler

import (
	"encoding/json"
	"fmt"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/adapter"
)

// anthropicToOpenAIChatRequest converts an Anthropic Messages API request to an internal ChatRequest.
func anthropicToOpenAIChatRequest(req *model.AnthropicMessagesRequest) *adapter.ChatRequest {
	messages := make([]map[string]string, 0, len(req.Messages)+1)

	// Extract system prompt and prepend as role:"system" message
	if req.System != nil {
		systemText := extractAnthropicSystemText(req.System)
		if systemText != "" {
			messages = append(messages, map[string]string{
				"role":    "system",
				"content": systemText,
			})
		}
	}

	// Convert messages
	for _, msg := range req.Messages {
		content := extractAnthropicContentText(msg.Content)
		messages = append(messages, map[string]string{
			"role":    msg.Role,
			"content": content,
		})
	}

	chatReq := &adapter.ChatRequest{
		Model:     req.Model,
		Messages:  messages,
		MaxTokens: req.MaxTokens,
		Stream:    req.Stream,
	}

	if req.Temperature != nil {
		chatReq.Temperature = *req.Temperature
	}
	if req.TopP != nil {
		chatReq.TopP = *req.TopP
	}

	return chatReq
}

// extractAnthropicSystemText extracts text from Anthropic system field.
func extractAnthropicSystemText(system interface{}) string {
	switch s := system.(type) {
	case string:
		return s
	case []interface{}:
		text := ""
		for _, block := range s {
			if b, ok := block.(map[string]interface{}); ok {
				if b["type"] == "text" {
					if t, ok := b["text"].(string); ok {
						if text != "" {
							text += "\n"
						}
						text += t
					}
				}
			}
		}
		return text
	default:
		return fmt.Sprintf("%v", system)
	}
}

// extractAnthropicContentText extracts text from Anthropic message content.
func extractAnthropicContentText(content interface{}) string {
	switch c := content.(type) {
	case string:
		return c
	case []interface{}:
		text := ""
		for _, block := range c {
			if b, ok := block.(map[string]interface{}); ok {
				if b["type"] == "text" {
					if t, ok := b["text"].(string); ok {
						if text != "" {
							text += "\n"
						}
						text += t
					}
				}
			}
		}
		return text
	default:
		return fmt.Sprintf("%v", content)
	}
}

// openAIToAnthropicResponse converts an OpenAI ChatResponse to Anthropic format.
func openAIToAnthropicResponse(resp *adapter.ChatResponse) *model.AnthropicMessagesResponse {
	content := make([]model.AnthropicResponseContent, 0, len(resp.Choices))
	var stopReason string

	for _, choice := range resp.Choices {
		if choice.Message.Content != "" {
			content = append(content, model.AnthropicResponseContent{
				Type: "text",
				Text: choice.Message.Content,
			})
		}
		stopReason = mapStopReason(choice.FinishReason)
	}

	if len(content) == 0 {
		content = []model.AnthropicResponseContent{
			{Type: "text", Text: ""},
		}
	}

	var stopReasonPtr *string
	if stopReason != "" {
		stopReasonPtr = &stopReason
	}

	return &model.AnthropicMessagesResponse{
		ID:         resp.ID,
		Type:       "message",
		Role:       "assistant",
		Content:    content,
		Model:      resp.Model,
		StopReason: stopReasonPtr,
		Usage: model.AnthropicUsage{
			InputTokens:  resp.Usage.PromptTokens,
			OutputTokens: resp.Usage.CompletionTokens,
		},
	}
}

// mapStopReason maps OpenAI finish_reason to Anthropic stop_reason.
func mapStopReason(openaiReason string) string {
	switch openaiReason {
	case "stop":
		return "end_turn"
	case "length":
		return "max_tokens"
	case "tool_calls":
		return "tool_use"
	case "end_turn":
		return "end_turn" // pass through if already Anthropic format
	case "max_tokens":
		return "max_tokens"
	default:
		if openaiReason != "" {
			return "end_turn"
		}
		return ""
	}
}

// mapStopReasonReverse maps Anthropic stop_reason to OpenAI finish_reason.
func mapStopReasonReverse(anthropicReason string) string {
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

// writeAnthropicSSEvent writes an Anthropic-format SSE event to the response writer.
// Unlike Gin's c.SSEvent() which uses "event: message", this writes a custom event type.
func writeAnthropicSSEvent(w interface{ Write([]byte) (int, error) }, eventType string, data interface{}) {
	jsonBytes, _ := json.Marshal(data)
	fmt.Fprintf(w, "event: %s\n", eventType)
	fmt.Fprintf(w, "data: %s\n\n", string(jsonBytes))
}

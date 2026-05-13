package handler

import (
	"testing"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/adapter"
)

// --- anthropicToOpenAIChatRequest ---

func TestAnthropicToOpenAI_SimpleText(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 1024,
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hello"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if result.Model != "claude-3-sonnet" {
		t.Errorf("expected model 'claude-3-sonnet', got '%s'", result.Model)
	}
	if result.MaxTokens != 1024 {
		t.Errorf("expected max_tokens 1024, got %d", result.MaxTokens)
	}
	if len(result.Messages) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result.Messages))
	}
	if result.Messages[0]["role"] != "user" {
		t.Errorf("expected role 'user', got '%s'", result.Messages[0]["role"])
	}
	if result.Messages[0]["content"] != "hello" {
		t.Errorf("expected content 'hello', got '%s'", result.Messages[0]["content"])
	}
}

func TestAnthropicToOpenAI_SystemPromptString(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 1024,
		System:    "You are a helpful assistant.",
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hi"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if len(result.Messages) != 2 {
		t.Fatalf("expected 2 messages (system + user), got %d", len(result.Messages))
	}
	if result.Messages[0]["role"] != "system" {
		t.Errorf("expected first message role 'system', got '%s'", result.Messages[0]["role"])
	}
	if result.Messages[0]["content"] != "You are a helpful assistant." {
		t.Errorf("expected system content 'You are a helpful assistant.', got '%s'", result.Messages[0]["content"])
	}
	if result.Messages[1]["role"] != "user" {
		t.Errorf("expected second message role 'user', got '%s'", result.Messages[1]["role"])
	}
}

func TestAnthropicToOpenAI_SystemPromptArray(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 1024,
		System: []interface{}{
			map[string]interface{}{"type": "text", "text": "Part 1"},
			map[string]interface{}{"type": "text", "text": "Part 2"},
		},
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hi"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if result.Messages[0]["role"] != "system" {
		t.Errorf("expected system role, got '%s'", result.Messages[0]["role"])
	}
	if result.Messages[0]["content"] != "Part 1\nPart 2" {
		t.Errorf("expected 'Part 1\\nPart 2', got '%s'", result.Messages[0]["content"])
	}
}

func TestAnthropicToOpenAI_ContentArray(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 1024,
		Messages: []model.AnthropicMessage{
			{
				Role: "user",
				Content: []interface{}{
					map[string]interface{}{"type": "text", "text": "describe this"},
					map[string]interface{}{"type": "image", "source": map[string]interface{}{"type": "base64"}},
				},
			},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if result.Messages[0]["content"] != "describe this" {
		t.Errorf("expected 'describe this', got '%s'", result.Messages[0]["content"])
	}
}

func TestAnthropicToOpenAI_MultiTurn(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 512,
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hi"},
			{Role: "assistant", Content: "hello!"},
			{Role: "user", Content: "how are you?"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if len(result.Messages) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result.Messages))
	}
	if result.Messages[2]["content"] != "how are you?" {
		t.Errorf("expected 'how are you?', got '%s'", result.Messages[2]["content"])
	}
}

func TestAnthropicToOpenAI_TemperatureTopP(t *testing.T) {
	temp := 0.7
	topP := 0.9
	req := &model.AnthropicMessagesRequest{
		Model:       "claude-3-sonnet",
		MaxTokens:   1024,
		Temperature: &temp,
		TopP:        &topP,
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hi"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if result.Temperature != 0.7 {
		t.Errorf("expected temperature 0.7, got %f", result.Temperature)
	}
	if result.TopP != 0.9 {
		t.Errorf("expected top_p 0.9, got %f", result.TopP)
	}
}

func TestAnthropicToOpenAI_StreamFlag(t *testing.T) {
	req := &model.AnthropicMessagesRequest{
		Model:     "claude-3-sonnet",
		MaxTokens: 1024,
		Stream:    true,
		Messages: []model.AnthropicMessage{
			{Role: "user", Content: "hi"},
		},
	}
	result := anthropicToOpenAIChatRequest(req)

	if !result.Stream {
		t.Error("expected stream=true")
	}
}

// --- extractAnthropicContentText ---

func TestExtractAnthropicContentText_String(t *testing.T) {
	result := extractAnthropicContentText("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestExtractAnthropicContentText_ArrayWithImages(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "describe this"},
		map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/img.jpg"}},
	}
	result := extractAnthropicContentText(input)
	if result != "describe this" {
		t.Errorf("expected 'describe this', got '%s'", result)
	}
}

func TestExtractAnthropicContentText_MultipleTextBlocks(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "first"},
		map[string]interface{}{"type": "text", "text": "second"},
	}
	result := extractAnthropicContentText(input)
	if result != "first\nsecond" {
		t.Errorf("expected 'first\\nsecond', got '%s'", result)
	}
}

func TestExtractAnthropicContentText_EmptyArray(t *testing.T) {
	result := extractAnthropicContentText([]interface{}{})
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

// --- openAIToAnthropicResponse ---

func TestOpenAIToAnthropicResponse_Basic(t *testing.T) {
	resp := &adapter.ChatResponse{
		ID:      "chatcmpl-123",
		Object:  "chat.completion",
		Model:   "gpt-4",
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
					Content: "Hello there!",
				},
				FinishReason: "stop",
			},
		},
		Usage: struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		}{
			PromptTokens:     10,
			CompletionTokens: 5,
		},
	}

	result := openAIToAnthropicResponse(resp)

	if result.ID != "chatcmpl-123" {
		t.Errorf("expected ID 'chatcmpl-123', got '%s'", result.ID)
	}
	if result.Type != "message" {
		t.Errorf("expected type 'message', got '%s'", result.Type)
	}
	if result.Role != "assistant" {
		t.Errorf("expected role 'assistant', got '%s'", result.Role)
	}
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}
	if result.Content[0].Type != "text" {
		t.Errorf("expected content type 'text', got '%s'", result.Content[0].Type)
	}
	if result.Content[0].Text != "Hello there!" {
		t.Errorf("expected 'Hello there!', got '%s'", result.Content[0].Text)
	}
	if result.StopReason == nil || *result.StopReason != "end_turn" {
		t.Errorf("expected stop_reason 'end_turn', got %v", result.StopReason)
	}
	if result.Usage.InputTokens != 10 {
		t.Errorf("expected input_tokens 10, got %d", result.Usage.InputTokens)
	}
	if result.Usage.OutputTokens != 5 {
		t.Errorf("expected output_tokens 5, got %d", result.Usage.OutputTokens)
	}
}

func TestOpenAIToAnthropicResponse_LengthFinish(t *testing.T) {
	resp := &adapter.ChatResponse{
		ID:    "chatcmpl-456",
		Model: "gpt-4",
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
				}{Content: "truncated..."},
				FinishReason: "length",
			},
		},
	}

	result := openAIToAnthropicResponse(resp)
	if result.StopReason == nil || *result.StopReason != "max_tokens" {
		t.Errorf("expected stop_reason 'max_tokens', got %v", result.StopReason)
	}
}

func TestOpenAIToAnthropicResponse_EmptyContent(t *testing.T) {
	resp := &adapter.ChatResponse{
		ID:    "chatcmpl-789",
		Model: "gpt-4",
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
				}{Content: ""},
				FinishReason: "stop",
			},
		},
	}

	result := openAIToAnthropicResponse(resp)
	if len(result.Content) != 1 {
		t.Fatalf("expected 1 content block, got %d", len(result.Content))
	}
	if result.Content[0].Text != "" {
		t.Errorf("expected empty text, got '%s'", result.Content[0].Text)
	}
}

// --- mapStopReason ---

func TestMapStopReason(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"stop", "end_turn"},
		{"length", "max_tokens"},
		{"tool_calls", "tool_use"},
		{"end_turn", "end_turn"},
		{"max_tokens", "max_tokens"},
		{"", ""},
		{"unknown", "end_turn"},
	}
	for _, tt := range tests {
		result := mapStopReason(tt.input)
		if result != tt.expected {
			t.Errorf("mapStopReason(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// --- mapStopReasonReverse ---

func TestMapStopReasonReverse(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"end_turn", "stop"},
		{"max_tokens", "length"},
		{"stop_sequence", "stop"},
		{"tool_use", "tool_calls"},
		{"unknown", "stop"},
	}
	for _, tt := range tests {
		result := mapStopReasonReverse(tt.input)
		if result != tt.expected {
			t.Errorf("mapStopReasonReverse(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

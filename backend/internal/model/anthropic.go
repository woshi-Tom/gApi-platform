package model

// AnthropicMessagesRequest represents the inbound Anthropic Messages API request.
type AnthropicMessagesRequest struct {
	Model       string             `json:"model" binding:"required"`
	MaxTokens   int                `json:"max_tokens" binding:"required"`
	Messages    []AnthropicMessage `json:"messages" binding:"required"`
	System      interface{}        `json:"system,omitempty"`    // string or []content block
	Stream      bool               `json:"stream,omitempty"`
	Temperature *float64           `json:"temperature,omitempty"`
	TopP        *float64           `json:"top_p,omitempty"`
	TopK        *int               `json:"top_k,omitempty"`
	Metadata    *AnthropicMetadata `json:"metadata,omitempty"`
}

// AnthropicMessage represents a single message in the Anthropic format.
type AnthropicMessage struct {
	Role    string      `json:"role"`    // "user" or "assistant"
	Content interface{} `json:"content"` // string or []content block
}

// AnthropicMetadata contains optional request metadata.
type AnthropicMetadata struct {
	UserID string `json:"user_id,omitempty"`
}

// AnthropicMessagesResponse represents the outbound Anthropic Messages API response.
type AnthropicMessagesResponse struct {
	ID         string                     `json:"id"`
	Type       string                     `json:"type"` // "message"
	Role       string                     `json:"role"` // "assistant"
	Content    []AnthropicResponseContent `json:"content"`
	Model      string                     `json:"model"`
	StopReason *string                    `json:"stop_reason,omitempty"`
	Usage      AnthropicUsage             `json:"usage"`
}

// AnthropicResponseContent represents a content block in the response.
type AnthropicResponseContent struct {
	Type string `json:"type"` // "text"
	Text string `json:"text,omitempty"`
}

// AnthropicUsage represents token usage in Anthropic format.
type AnthropicUsage struct {
	InputTokens  int `json:"input_tokens"`
	OutputTokens int `json:"output_tokens"`
}

// AnthropicErrorResponse represents an Anthropic error response.
type AnthropicErrorResponse struct {
	Type  string            `json:"type"` // "error"
	Error AnthropicErrorBody `json:"error"`
}

// AnthropicErrorBody contains error details.
type AnthropicErrorBody struct {
	Type    string `json:"type"`    // "invalid_request_error", "authentication_error", "api_error"
	Message string `json:"message"`
}

// AnthropicStreamEvent represents a streaming event in Anthropic SSE format.
type AnthropicStreamEvent struct {
	Type         string                     `json:"type"`
	Message      *AnthropicMessagesResponse `json:"message,omitempty"`
	Index        int                        `json:"index,omitempty"`
	ContentBlock *AnthropicResponseContent  `json:"content_block,omitempty"`
	Delta        *AnthropicStreamDelta      `json:"delta,omitempty"`
	Usage        *AnthropicUsage            `json:"usage,omitempty"`
}

// AnthropicStreamDelta represents the delta in streaming events.
type AnthropicStreamDelta struct {
	Type         string  `json:"type,omitempty"`
	Text         string  `json:"text,omitempty"`
	StopReason   *string `json:"stop_reason,omitempty"`
	StopSequence *string `json:"stop_sequence,omitempty"`
}

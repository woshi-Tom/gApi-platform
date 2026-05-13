package adapter

import (
	"io"
	"net/http"
	"strings"
	"testing"
)

// --- System message extraction (tested via Chat request building logic) ---

func TestClaudeAdapter_SystemMessageExtraction(t *testing.T) {
	// Simulate the system extraction logic from Chat method
	messages := []map[string]string{
		{"role": "system", "content": "You are helpful."},
		{"role": "user", "content": "hi"},
		{"role": "assistant", "content": "hello!"},
		{"role": "system", "content": "Be concise."},
	}

	var systemParts []string
	var filtered []map[string]string
	for _, m := range messages {
		if m["role"] == "system" {
			systemParts = append(systemParts, m["content"])
		} else {
			filtered = append(filtered, m)
		}
	}

	if len(systemParts) != 2 {
		t.Fatalf("expected 2 system parts, got %d", len(systemParts))
	}
	if systemParts[0] != "You are helpful." {
		t.Errorf("expected 'You are helpful.', got '%s'", systemParts[0])
	}
	if systemParts[1] != "Be concise." {
		t.Errorf("expected 'Be concise.', got '%s'", systemParts[1])
	}
	if len(filtered) != 2 {
		t.Fatalf("expected 2 non-system messages, got %d", len(filtered))
	}
	if filtered[0]["role"] != "user" {
		t.Errorf("expected first non-system role 'user', got '%s'", filtered[0]["role"])
	}
}

func TestClaudeAdapter_NoSystemMessage(t *testing.T) {
	messages := []map[string]string{
		{"role": "user", "content": "hi"},
		{"role": "assistant", "content": "hello!"},
	}

	var systemParts []string
	var filtered []map[string]string
	for _, m := range messages {
		if m["role"] == "system" {
			systemParts = append(systemParts, m["content"])
		} else {
			filtered = append(filtered, m)
		}
	}

	if len(systemParts) != 0 {
		t.Errorf("expected 0 system parts, got %d", len(systemParts))
	}
	if len(filtered) != 2 {
		t.Errorf("expected 2 messages, got %d", len(filtered))
	}
}

// newMockSSE creates a mock http.Response with SSE body for testing readStream.
func newMockSSE(body string) *http.Response {
	return &http.Response{
		Body: io.NopCloser(strings.NewReader(body)),
	}
}

// --- Anthropic SSE stream parsing ---

func TestClaudeAdapter_ReadStream_TextDelta(t *testing.T) {
	sseData := "event: content_block_start\ndata: {\"type\":\"content_block_start\",\"index\":0,\"content_block\":{\"type\":\"text\",\"text\":\"\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\"Hello\"}}\n\nevent: content_block_delta\ndata: {\"type\":\"content_block_delta\",\"index\":0,\"delta\":{\"type\":\"text_delta\",\"text\":\" world\"}}\n\nevent: content_block_stop\ndata: {\"type\":\"content_block_stop\",\"index\":0}\n\nevent: message_delta\ndata: {\"type\":\"message_delta\",\"delta\":{\"stop_reason\":\"end_turn\"},\"usage\":{\"output_tokens\":5}}\n\nevent: message_stop\ndata: {\"type\":\"message_stop\"}\n"

	a := &ClaudeAdapter{}
	ch := make(chan ChatStreamResponse, 10)
	resp := newMockSSE(sseData)

	go a.readStream(resp.Body, ch, resp)

	var texts []string
	var gotDone bool
	for r := range ch {
		if r.Done {
			gotDone = true
			break
		}
		if r.Error != nil {
			t.Fatalf("unexpected error: %s", r.Error.Message)
		}
		for _, choice := range r.Choices {
			if choice.Delta.Content != "" {
				texts = append(texts, choice.Delta.Content)
			}
			if choice.FinishReason != "" && choice.FinishReason != "stop" {
				t.Errorf("expected finish_reason 'stop', got '%s'", choice.FinishReason)
			}
		}
	}

	if !gotDone {
		t.Error("expected Done signal")
	}
	if len(texts) != 2 {
		t.Fatalf("expected 2 text chunks, got %d", len(texts))
	}
	if texts[0] != "Hello" {
		t.Errorf("expected 'Hello', got '%s'", texts[0])
	}
	if texts[1] != " world" {
		t.Errorf("expected ' world', got '%s'", texts[1])
	}
}

func TestClaudeAdapter_ReadStream_Error(t *testing.T) {
	sseData := "event: error\ndata: {\"type\":\"error\",\"type\":\"api_error\",\"message\":\"rate limit exceeded\"}\n"

	a := &ClaudeAdapter{}
	ch := make(chan ChatStreamResponse, 10)
	resp := newMockSSE(sseData)

	go a.readStream(resp.Body, ch, resp)

	var errMsg string
	for r := range ch {
		if r.Error != nil {
			errMsg = r.Error.Message
		}
	}

	if errMsg == "" {
		t.Error("expected error message")
	}
}

func TestClaudeAdapter_ReadStream_EmptyBody(t *testing.T) {
	a := &ClaudeAdapter{}
	ch := make(chan ChatStreamResponse, 10)
	resp := newMockSSE("")

	go a.readStream(resp.Body, ch, resp)

	var count int
	for range ch {
		count++
	}
	if count != 0 {
		t.Errorf("expected 0 events from empty body, got %d", count)
	}
}

func TestMapStopReasonReverseStream(t *testing.T) {
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
		result := mapStopReasonReverseStream(tt.input)
		if result != tt.expected {
			t.Errorf("mapStopReasonReverseStream(%q) = %q, want %q", tt.input, result, tt.expected)
		}
	}
}

// Ensure unused import is consumed
var _ io.Reader = strings.NewReader("")

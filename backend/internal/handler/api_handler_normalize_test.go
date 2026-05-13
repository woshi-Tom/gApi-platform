package handler

import (
	"testing"
)

func TestExtractTextContent_String(t *testing.T) {
	result := extractTextContent("hello")
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestExtractTextContent_EmptyString(t *testing.T) {
	result := extractTextContent("")
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestExtractTextContent_SingleTextBlock(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "hello"},
	}
	result := extractTextContent(input)
	if result != "hello" {
		t.Errorf("expected 'hello', got '%s'", result)
	}
}

func TestExtractTextContent_MultipleTextBlocks(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "first"},
		map[string]interface{}{"type": "text", "text": "second"},
	}
	result := extractTextContent(input)
	if result != "first\nsecond" {
		t.Errorf("expected 'first\\nsecond', got '%s'", result)
	}
}

func TestExtractTextContent_NonTextBlocksIgnored(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{"type": "text", "text": "describe this"},
		map[string]interface{}{"type": "image_url", "image_url": map[string]interface{}{"url": "https://example.com/img.jpg"}},
	}
	result := extractTextContent(input)
	if result != "describe this" {
		t.Errorf("expected 'describe this', got '%s'", result)
	}
}

func TestExtractTextContent_EmptyArray(t *testing.T) {
	input := []interface{}{}
	result := extractTextContent(input)
	if result != "" {
		t.Errorf("expected empty string, got '%s'", result)
	}
}

func TestExtractTextContent_Nil(t *testing.T) {
	result := extractTextContent(nil)
	// nil falls through to default case: fmt.Sprintf("%v", nil) = "<nil>"
	if result != "<nil>" {
		t.Errorf("expected '<nil>', got '%s'", result)
	}
}

func TestExtractTextContent_IntFallback(t *testing.T) {
	result := extractTextContent(42)
	if result != "42" {
		t.Errorf("expected '42', got '%s'", result)
	}
}

func TestNormalizeMessages_StringContent(t *testing.T) {
	msgs := []map[string]interface{}{
		{"role": "user", "content": "hello"},
	}
	result := normalizeMessages(msgs)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0]["role"] != "user" {
		t.Errorf("expected role 'user', got '%s'", result[0]["role"])
	}
	if result[0]["content"] != "hello" {
		t.Errorf("expected content 'hello', got '%s'", result[0]["content"])
	}
}

func TestNormalizeMessages_ArrayContent(t *testing.T) {
	msgs := []map[string]interface{}{
		{
			"role": "user",
			"content": []interface{}{
				map[string]interface{}{"type": "text", "text": "hello"},
			},
		},
	}
	result := normalizeMessages(msgs)
	if len(result) != 1 {
		t.Fatalf("expected 1 message, got %d", len(result))
	}
	if result[0]["content"] != "hello" {
		t.Errorf("expected content 'hello', got '%s'", result[0]["content"])
	}
}

func TestNormalizeMessages_MixedMessages(t *testing.T) {
	msgs := []map[string]interface{}{
		{"role": "user", "content": "plain text"},
		{
			"role": "assistant",
			"content": []interface{}{
				map[string]interface{}{"type": "text", "text": "array response"},
			},
		},
		{"role": "user", "content": "follow up"},
	}
	result := normalizeMessages(msgs)
	if len(result) != 3 {
		t.Fatalf("expected 3 messages, got %d", len(result))
	}
	if result[0]["content"] != "plain text" {
		t.Errorf("msg[0] expected 'plain text', got '%s'", result[0]["content"])
	}
	if result[1]["content"] != "array response" {
		t.Errorf("msg[1] expected 'array response', got '%s'", result[1]["content"])
	}
	if result[2]["content"] != "follow up" {
		t.Errorf("msg[2] expected 'follow up', got '%s'", result[2]["content"])
	}
}

func TestNormalizeMessages_EmptyInput(t *testing.T) {
	result := normalizeMessages([]map[string]interface{}{})
	if len(result) != 0 {
		t.Errorf("expected empty result, got %d messages", len(result))
	}
}

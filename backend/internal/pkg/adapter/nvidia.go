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

const (
	NVIDIABaseURL     = "https://integrate.api.nvidia.com/v1"
	NVIDIAModelsURL   = "https://integrate.api.nvidia.com/v1/models"
)

type NVIDIAAdapter struct {
	client *http.Client
}

func NewNVIDIAAdapter() *NVIDIAAdapter {
	return &NVIDIAAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *NVIDIAAdapter) GetName() string {
	return "NVIDIA NIM"
}

func (a *NVIDIAAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = NVIDIABaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
	}
	if req.TopP > 0 {
		payload["top_p"] = req.TopP
	}
	if req.Stream {
		payload["stream"] = true
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	httpReq.Header.Set("Accept", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ChatResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		if result.Error != nil {
			return &result, fmt.Errorf("NVIDIA API error: %s (status %d)", result.Error.Message, resp.StatusCode)
		}
		return &result, fmt.Errorf("NVIDIA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return &result, nil
}

func (a *NVIDIAAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = NVIDIABaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		payload["max_tokens"] = req.MaxTokens
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("NVIDIA API error: status %d, body: %s", resp.StatusCode, string(body))
	}

	ch := make(chan ChatStreamResponse, 100)
	go a.readStream(resp.Body, ch, resp)
	return ch, nil
}

func (a *NVIDIAAdapter) readStream(body io.Reader, ch chan<- ChatStreamResponse, resp *http.Response) {
	defer close(ch)
	defer resp.Body.Close()

	reader := NewSSEReader(body)
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

		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				ch <- ChatStreamResponse{Done: true}
				return
			}

			var chunk ChatStreamResponse
			if err := json.Unmarshal([]byte(data), &chunk); err != nil {
				continue
			}
			ch <- chunk
		}
	}
}

func (a *NVIDIAAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = NVIDIABaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/embeddings", baseURL)

	payload := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
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
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result EmbeddingsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

func (a *NVIDIAAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = NVIDIABaseURL
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/models", baseURL)

	httpReq, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var result ModelsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	return &result, nil
}

type SSEReader struct {
	reader io.Reader
}

func NewSSEReader(r io.Reader) *SSEReader {
	return &SSEReader{reader: r}
}

func (s *SSEReader) Read() (string, error) {
	var line []byte
	buf := make([]byte, 1)
	for {
		n, err := s.reader.Read(buf)
		if err != nil {
			return string(line), err
		}
		if n > 0 {
			if buf[0] == '\n' {
				return string(line), nil
			}
			line = append(line, buf[0])
		}
	}
}

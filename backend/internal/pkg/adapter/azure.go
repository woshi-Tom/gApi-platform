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

type AzureAdapter struct {
	client *http.Client
}

func NewAzureAdapter() *AzureAdapter {
	return &AzureAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *AzureAdapter) GetName() string {
	return "Azure OpenAI"
}

func (a *AzureAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("Azure OpenAI requires base_url")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	deployment := channel.ModelMapping["deployment"]
	if deployment == "" {
		deployment = req.Model
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", baseURL, deployment)

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
	httpReq.Header.Set("api-key", channel.APIKey)

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
			return &result, fmt.Errorf("Azure OpenAI error: %s (status %d)", result.Error.Message, resp.StatusCode)
		}
		return &result, fmt.Errorf("Azure OpenAI error: status %d, body: %s", resp.StatusCode, string(body))
	}

	return &result, nil
}

func (a *AzureAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("Azure OpenAI requires base_url")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	deployment := channel.ModelMapping["deployment"]
	if deployment == "" {
		deployment = req.Model
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/chat/completions?api-version=2024-02-15-preview", baseURL, deployment)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}
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
	httpReq.Header.Set("api-key", channel.APIKey)
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		resp.Body.Close()
		return nil, fmt.Errorf("Azure OpenAI error: status %d, body: %s", resp.StatusCode, string(body))
	}

	ch := make(chan ChatStreamResponse, 100)
	go a.readStream(resp.Body, ch, resp)
	return ch, nil
}

func (a *AzureAdapter) readStream(body io.Reader, ch chan<- ChatStreamResponse, resp *http.Response) {
	defer close(ch)
		defer resp.Body.Close()

	reader := NewSSEReader(body)
	for {
		line, err := reader.Read()
		if err != nil {
			if err != io.EOF {
				ch <- ChatStreamResponse{Error: &StreamError{Message: fmt.Sprintf("stream error: %v", err)}}
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

func (a *AzureAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		return nil, fmt.Errorf("Azure OpenAI requires base_url")
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	deployment := channel.ModelMapping["deployment"]
	if deployment == "" {
		deployment = req.Model
	}

	url := fmt.Sprintf("%s/openai/deployments/%s/embeddings?api-version=2024-02-15-preview", baseURL, deployment)

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
	httpReq.Header.Set("api-key", channel.APIKey)

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

func (a *AzureAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
	return &ModelsResponse{
		Object: "list",
		Data: []struct {
			ID         string `json:"id"`
			Object     string `json:"object"`
			Created    int    `json:"created"`
			OwnedBy    string `json:"owned_by"`
			Permission []struct {
				ID                  string `json:"id"`
				Object              string `json:"object"`
				Created             int    `json:"created"`
				AllowCreateEngine   bool   `json:"allow_create_engine"`
				AllowSampling       bool   `json:"allow_sampling"`
				AllowLogprobs       bool   `json:"allow_logprobs"`
				AllowSearchIndices  bool   `json:"allow_search_indices"`
				AllowView           bool   `json:"allow_view"`
				AllowFineTuning     bool   `json:"allow_fine_tuning"`
				Organization        string `json:"organization"`
				Group               interface{} `json:"group"`
				IsBlocking          bool   `json:"is_blocking"`
			} `json:"permission,omitempty"`
		}{
			{ID: "gpt-35-turbo", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "azure"},
			{ID: "gpt-4", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "azure"},
		},
	}, nil
}

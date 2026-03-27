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

type DeepSeekAdapter struct {
	client *http.Client
}

func NewDeepSeekAdapter() *DeepSeekAdapter {
	return &DeepSeekAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *DeepSeekAdapter) GetName() string {
	return "DeepSeek"
}

func (a *DeepSeekAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
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
	if req.Stream {
		payload["stream"] = true
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ChatResponse
	json.Unmarshal(body, &result)

	if resp.StatusCode != http.StatusOK {
		return &result, fmt.Errorf("DeepSeek API error: status %d", resp.StatusCode)
	}

	return &result, nil
}

func (a *DeepSeekAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
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

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	httpReq.Header.Set("Accept", "text/event-stream")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan ChatStreamResponse, 100)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		reader := NewSSEReader(resp.Body)
		for {
			line, err := reader.Read()
			if err != nil {
				return
			}
			if strings.HasPrefix(line, "data: ") {
				data := strings.TrimPrefix(line, "data: ")
				if data == "[DONE]" {
					ch <- ChatStreamResponse{Done: true}
					return
				}
				var chunk ChatStreamResponse
				json.Unmarshal([]byte(data), &chunk)
				ch <- chunk
			}
		}
	}()
	return ch, nil
}

func (a *DeepSeekAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.deepseek.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/embeddings", baseURL)

	payload := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result EmbeddingsResponse
	json.Unmarshal(body, &result)
	return &result, nil
}

func (a *DeepSeekAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "deepseek-chat", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "deepseek"},
			{ID: "deepseek-coder", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "deepseek"},
		},
	}, nil
}

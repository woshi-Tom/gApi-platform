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

type GeminiAdapter struct {
	client *http.Client
}

func NewGeminiAdapter() *GeminiAdapter {
	return &GeminiAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *GeminiAdapter) GetName() string {
	return "Google Gemini"
}

func (a *GeminiAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	model := req.Model
	if strings.HasPrefix(model, "models/") {
		model = strings.TrimPrefix(model, "models/")
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", baseURL, model, channel.APIKey)

	contents := make([]map[string]interface{}, 0)

	for _, m := range req.Messages {
		role := m["role"]
		if role == "assistant" {
			role = "model"
		} else if role == "system" {
			role = "user"
		}
		
		if len(contents) == 0 || contents[len(contents)-1]["role"] != role {
			if len(contents) > 0 {
				contents = append(contents, map[string]interface{}{
					"role": role,
					"parts": []map[string]string{},
				})
			}
		}
		
		lastIdx := len(contents) - 1
		parts := contents[lastIdx]["parts"].([]map[string]string)
		parts = append(parts, map[string]string{"text": m["content"]})
		contents[lastIdx]["parts"] = parts
	}

	payload := map[string]interface{}{
		"contents": contents,
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
	}
	if req.MaxTokens > 0 {
		payload["maxOutputTokens"] = req.MaxTokens
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

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var geminiResp struct {
		Candidates []struct {
			Content struct {
				Parts []struct {
					Text string `json:"text"`
				} `json:"parts"`
			} `json:"content"`
			FinishReason string `json:"finishReason"`
		} `json:"candidates"`
		Error *struct {
			Message string `json:"message"`
			Code    int    `json:"code"`
		} `json:"error"`
	}

	if err := json.Unmarshal(body, &geminiResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	if resp.StatusCode != http.StatusOK || geminiResp.Error != nil {
		errMsg := "Gemini API error"
		if geminiResp.Error != nil {
			errMsg = geminiResp.Error.Message
		}
		return nil, fmt.Errorf("%s (status %d)", errMsg, resp.StatusCode)
	}

	if len(geminiResp.Candidates) == 0 {
		return nil, fmt.Errorf("no candidates in response")
	}

	content := ""
	for _, part := range geminiResp.Candidates[0].Content.Parts {
		content += part.Text
	}

	return &ChatResponse{
		ID:      fmt.Sprintf("gemini-%d", time.Now().UnixNano()),
		Object:  "chat.completion",
		Created: time.Now().Unix(),
		Model:   req.Model,
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
					Role:    "model",
					Content: content,
				},
				FinishReason: geminiResp.Candidates[0].FinishReason,
			},
		},
	}, nil
}

func (a *GeminiAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	return nil, fmt.Errorf("Gemini streaming not implemented yet")
}

func (a *GeminiAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://generativelanguage.googleapis.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	model := req.Model
	if strings.HasPrefix(model, "models/") {
		model = strings.TrimPrefix(model, "models/")
	}

	url := fmt.Sprintf("%s/v1beta/models/%s:batchEmbedContents?key=%s", baseURL, model, channel.APIKey)

	var inputs []string
	switch v := req.Input.(type) {
	case string:
		inputs = []string{v}
	case []string:
		inputs = v
	case []interface{}:
		for _, item := range v {
			if s, ok := item.(string); ok {
				inputs = append(inputs, s)
			}
		}
	}

	payload := map[string]interface{}{
		"requests": inputs,
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

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var embeddingsResp struct {
		Embeddings []struct {
			Values []float64 `json:"values"`
		} `json:"embeddings"`
	}

	if err := json.Unmarshal(body, &embeddingsResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	result := &EmbeddingsResponse{
		Object: "list",
		Data:   []struct {
			Object    string    `json:"object"`
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		}{},
	}

	for i, emb := range embeddingsResp.Embeddings {
		result.Data = append(result.Data, struct {
			Object    string    `json:"object"`
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			Object:    "embedding",
			Embedding: emb.Values,
			Index:     i,
		})
	}

	return result, nil
}

func (a *GeminiAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "gemini-pro", Object: "model", Created: 1709608000, OwnedBy: "google"},
			{ID: "gemini-pro-vision", Object: "model", Created: 1709608000, OwnedBy: "google"},
			{ID: "gemini-1.5-pro", Object: "model", Created: 1719608000, OwnedBy: "google"},
			{ID: "gemini-1.5-flash", Object: "model", Created: 1719608000, OwnedBy: "google"},
		},
	}, nil
}

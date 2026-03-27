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

type ZhipuAdapter struct {
	client *http.Client
}

func NewZhipuAdapter() *ZhipuAdapter {
	return &ZhipuAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *ZhipuAdapter) GetName() string {
	return "Zhipu (智谱)"
}

func (a *ZhipuAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
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
		return &result, fmt.Errorf("Zhipu API error: status %d", resp.StatusCode)
	}

	return &result, nil
}

func (a *ZhipuAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
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

func (a *ZhipuAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://open.bigmodel.cn/api/paas/v4"
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

func (a *ZhipuAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "glm-4", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "zhipu"},
			{ID: "glm-3-turbo", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "zhipu"},
		},
	}, nil
}

type BaiduAdapter struct {
	client *http.Client
}

func NewBaiduAdapter() *BaiduAdapter {
	return &BaiduAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *BaiduAdapter) GetName() string {
	return "Baidu Qianfan (百度千帆)"
}

func (a *BaiduAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://qianfan.baidubce.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v2/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
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
		return &result, fmt.Errorf("Baidu API error: status %d", resp.StatusCode)
	}

	return &result, nil
}

func (a *BaiduAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	return nil, fmt.Errorf("Baidu streaming not implemented")
}

func (a *BaiduAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("Baidu embeddings not implemented")
}

func (a *BaiduAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "ernie-bot", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "baidu"},
			{ID: "ernie-bot-turbo", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "baidu"},
		},
	}, nil
}

type YiAdapter struct {
	client *http.Client
}

func NewYiAdapter() *YiAdapter {
	return &YiAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *YiAdapter) GetName() string {
	return "Yi (零一万物)"
}

func (a *YiAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.lingyiwanwu.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
	}
	if req.Temperature > 0 {
		payload["temperature"] = req.Temperature
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
	return &result, nil
}

func (a *YiAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	return nil, fmt.Errorf("Yi streaming not implemented")
}

func (a *YiAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("Yi embeddings not implemented")
}

func (a *YiAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "yi-large", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "yi"},
			{ID: "yi-medium", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "yi"},
		},
	}, nil
}

type YiAPIAdapter struct {
	client *http.Client
}

func NewYiAPIAdapter() *YiAPIAdapter {
	return &YiAPIAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *YiAPIAdapter) GetName() string {
	return "Yi API"
}

func (a *YiAPIAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	return nil, fmt.Errorf("Yi API not implemented")
}

func (a *YiAPIAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	return nil, fmt.Errorf("Yi API streaming not implemented")
}

func (a *YiAPIAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("Yi API embeddings not implemented")
}

func (a *YiAPIAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "yi-large", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "yi"},
			{ID: "yi-medium", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "yi"},
		},
	}, nil
}

type OllamaAdapter struct {
	client *http.Client
}

func NewOllamaAdapter() *OllamaAdapter {
	return &OllamaAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *OllamaAdapter) GetName() string {
	return "Ollama"
}

func (a *OllamaAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/api/chat", baseURL)

	messages := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]string{
			"role":    m["role"],
			"content": m["content"],
		}
	}

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   false,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ollamaResp struct {
		Message struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		Done    bool `json:"done"`
		TotalDuration int64 `json:"total_duration"`
	}

	json.Unmarshal(body, &ollamaResp)

	return &ChatResponse{
		ID:      fmt.Sprintf("ollama-%d", time.Now().UnixNano()),
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
					Role:    ollamaResp.Message.Role,
					Content: ollamaResp.Message.Content,
				},
				FinishReason: "stop",
			},
		},
	}, nil
}

func (a *OllamaAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/api/chat", baseURL)

	messages := make([]map[string]string, len(req.Messages))
	for i, m := range req.Messages {
		messages[i] = map[string]string{
			"role":    m["role"],
			"content": m["content"],
		}
	}

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": messages,
		"stream":   true,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}

	ch := make(chan ChatStreamResponse, 100)
	go func() {
		defer close(ch)
		defer resp.Body.Close()
		decoder := json.NewDecoder(resp.Body)
		for decoder.More() {
			var ollamaResp struct {
				Message struct {
					Role    string `json:"role"`
					Content string `json:"content"`
				} `json:"message"`
				Done bool `json:"done"`
			}
			if err := decoder.Decode(&ollamaResp); err != nil {
				return
			}
			ch <- ChatStreamResponse{
				Choices: []struct {
					Index        int `json:"index"`
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
							Role:    ollamaResp.Message.Role,
							Content: ollamaResp.Message.Content,
						},
					},
				},
				Done: ollamaResp.Done,
			}
		}
	}()
	return ch, nil
}

func (a *OllamaAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/api/embeddings", baseURL)

	payload := map[string]interface{}{
		"model": req.Model,
		"prompt": req.Input,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ollamaResp struct {
		Embedding []float64 `json:"embedding"`
	}
	json.Unmarshal(body, &ollamaResp)

	return &EmbeddingsResponse{
		Object: "list",
		Data: []struct {
			Object    string    `json:"object"`
			Embedding []float64 `json:"embedding"`
			Index     int       `json:"index"`
		}{
			{
				Object:    "embedding",
				Embedding: ollamaResp.Embedding,
				Index:     0,
			},
		},
	}, nil
}

func (a *OllamaAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:11434"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/api/tags", baseURL)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var ollamaResp struct {
		Models []struct {
			Name string `json:"name"`
		} `json:"models"`
	}
	json.Unmarshal(body, &ollamaResp)

	result := &ModelsResponse{Object: "list", Data: []struct {
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
	}{}}

	for _, m := range ollamaResp.Models {
		result.Data = append(result.Data, struct {
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
			ID:      m.Name,
			Object:  "model",
			OwnedBy: "ollama",
		})
	}

	return result, nil
}

type LocalAIAdapter struct {
	client *http.Client
}

func NewLocalAIAdapter() *LocalAIAdapter {
	return &LocalAIAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *LocalAIAdapter) GetName() string {
	return "LocalAI"
}

func (a *LocalAIAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)

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
	if channel.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ChatResponse
	json.Unmarshal(body, &result)
	return &result, nil
}

func (a *LocalAIAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")
	if channel.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	}
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

func (a *LocalAIAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/embeddings", baseURL)

	payload := map[string]interface{}{
		"model": req.Model,
		"input": req.Input,
	}

	jsonPayload, _ := json.Marshal(payload)

	httpReq, _ := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(jsonPayload))
	httpReq.Header.Set("Content-Type", "application/json")
	if channel.APIKey != "" {
		httpReq.Header.Set("Authorization", fmt.Sprintf("Bearer %s", channel.APIKey))
	}

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

func (a *LocalAIAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "http://localhost:8080"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/v1/models", baseURL)

	httpReq, _ := http.NewRequestWithContext(ctx, "GET", url, nil)

	resp, err := a.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)

	var result ModelsResponse
	json.Unmarshal(body, &result)
	return &result, nil
}

type GroqAdapter struct {
	client *http.Client
}

func NewGroqAdapter() *GroqAdapter {
	return &GroqAdapter{
		client: &http.Client{Timeout: 120 * time.Second},
	}
}

func (a *GroqAdapter) GetName() string {
	return "Groq"
}

func (a *GroqAdapter) Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/openai/v1/chat/completions", baseURL)

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
	return &result, nil
}

func (a *GroqAdapter) ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error) {
	baseURL := channel.BaseURL
	if baseURL == "" {
		baseURL = "https://api.groq.com"
	}
	baseURL = strings.TrimSuffix(baseURL, "/")

	url := fmt.Sprintf("%s/openai/v1/chat/completions", baseURL)

	payload := map[string]interface{}{
		"model":    req.Model,
		"messages": req.Messages,
		"stream":   true,
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

func (a *GroqAdapter) Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error) {
	return nil, fmt.Errorf("Groq embeddings not supported")
}

func (a *GroqAdapter) ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error) {
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
			{ID: "llama-3.3-70b-versatile", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "groq"},
			{ID: "llama-3.1-8b-instant", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "groq"},
			{ID: "mixtral-8x7b-32768", Object: "model", Created: int(time.Now().Unix()), OwnedBy: "groq"},
		},
	}, nil
}

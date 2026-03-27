package adapter

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"
)

type Channel struct {
	ID           uint
	Type         string
	Name         string
	BaseURL      string
	APIKey       string
	Models       []string
	ModelMapping map[string]string
	Timeout      int
}

type ChatRequest struct {
	Model       string                   `json:"model"`
	Messages    []map[string]string      `json:"messages"`
	Temperature float64                  `json:"temperature,omitempty"`
	MaxTokens   int                      `json:"max_tokens,omitempty"`
	TopP        float64                  `json:"top_p,omitempty"`
	Stream      bool                     `json:"stream,omitempty"`
	User        string                   `json:"user,omitempty"`
	Functions   []map[string]interface{} `json:"functions,omitempty"`
}

type ChatResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Message      struct {
			Role    string `json:"role"`
			Content string `json:"content"`
		} `json:"message"`
		FinishReason string `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Code    string `json:"code"`
		Type    string `json:"type"`
	} `json:"error,omitempty"`
}

type EmbeddingsRequest struct {
	Model string   `json:"model"`
	Input interface{} `json:"input"`
}

type EmbeddingsResponse struct {
	Object string `json:"object"`
	Data   []struct {
		Object    string    `json:"object"`
		Embedding []float64 `json:"embedding"`
		Index     int       `json:"index"`
	} `json:"data"`
	Usage struct {
		PromptTokens int `json:"prompt_tokens"`
		TotalTokens  int `json:"total_tokens"`
	} `json:"usage"`
}

type ModelsResponse struct {
	Object string `json:"object"`
	Data   []struct {
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
	} `json:"data"`
}

type Adapter interface {
	Chat(ctx context.Context, channel *Channel, req *ChatRequest) (*ChatResponse, error)
	ChatStream(ctx context.Context, channel *Channel, req *ChatRequest) (<-chan ChatStreamResponse, error)
	Embeddings(ctx context.Context, channel *Channel, req *EmbeddingsRequest) (*EmbeddingsResponse, error)
	ListModels(ctx context.Context, channel *Channel) (*ModelsResponse, error)
	GetName() string
}

type StreamError struct {
	Message string `json:"message"`
}

type ChatStreamResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int `json:"index"`
		Delta        struct {
			Role    string `json:"role,omitempty"`
			Content string `json:"content,omitempty"`
		} `json:"delta"`
		FinishReason string `json:"finish_reason,omitempty"`
	} `json:"choices"`
	Usage *struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage,omitempty"`
	Error *StreamError `json:"error,omitempty"`
	Done  bool        `json:"-"`
}

func GetHTTPClient(timeout int) *http.Client {
	if timeout <= 0 {
		timeout = 60
	}
	return &http.Client{
		Timeout: time.Duration(timeout) * time.Second,
	}
}

func NewRequest(method, url string, body interface{}, headers map[string]string) (*http.Request, error) {
	var bodyReader io.Reader
	if body != nil {
		if b, ok := body.([]byte); ok {
			bodyReader = bytes.NewReader(b)
		} else {
			jsonBytes, err := json.Marshal(body)
			if err != nil {
				return nil, err
			}
			bodyReader = bytes.NewReader(jsonBytes)
		}
	}

	req, err := http.NewRequest(method, url, bodyReader)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")
	for k, v := range headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

func DoRequest(client *http.Client, req *http.Request) (*http.Response, error) {
	return client.Do(req)
}

type TestResult struct {
	Success        bool
	StatusCode    int
	ResponseTimeMs int64
	Models        []string
	Error         string
}

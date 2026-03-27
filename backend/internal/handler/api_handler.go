package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/adapter"
	"gapi-platform/internal/pkg/crypto"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	tokenService   *service.TokenService
	channelService *service.ChannelService
	userRepo       *repository.UserRepository
}

func NewAPIHandler(tokenService *service.TokenService, channelService *service.ChannelService, userRepo *repository.UserRepository) *APIHandler {
	return &APIHandler{
		tokenService:   tokenService,
		channelService: channelService,
		userRepo:       userRepo,
	}
}

func (h *APIHandler) ChatCompletions(c *gin.Context) {
	var req model.ChatCompletionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "MISSING_MODEL",
				Message: "model is required",
			},
		})
		return
	}

	selectedChannel, err := h.channelService.SelectChannel(req.Model)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "NO_CHANNEL",
				Message: "no available channel for model: " + req.Model,
			},
		})
		return
	}

	apiKey, err := crypto.Decrypt(selectedChannel.APIKeyEncrypted)
	if err != nil {
		apiKey = selectedChannel.APIKeyEncrypted
	}

	channel := &adapter.Channel{
		ID:           selectedChannel.ID,
		Type:         selectedChannel.Type,
		Name:         selectedChannel.Name,
		BaseURL:      selectedChannel.BaseURL,
		APIKey:       apiKey,
		Models:       selectedChannel.GetModels(),
		ModelMapping: selectedChannel.GetModelMapping(),
		Timeout:      120,
	}

	chatReq := &adapter.ChatRequest{
		Model:       req.Model,
		Messages:    req.Messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		Stream:      req.Stream,
		User:        req.User,
	}

	chatAdapter, err := adapter.GetAdapter(selectedChannel.Type)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "ADAPTER_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	if req.Stream {
		h.handleStream(ctx, c, chatAdapter, channel, chatReq, selectedChannel.ID)
		return
	}

	resp, err := chatAdapter.Chat(ctx, channel, chatReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "UPSTREAM_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) handleStream(ctx context.Context, c *gin.Context, chatAdapter adapter.Adapter, channel *adapter.Channel, chatReq *adapter.ChatRequest, channelID uint) {
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Transfer-Encoding", "chunked")

	streamCh, err := chatAdapter.ChatStream(ctx, channel, chatReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "STREAM_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "STREAM_NOT_SUPPORTED",
				Message: "streaming not supported",
			},
		})
		return
	}

	c.Stream(func(w io.Writer) bool {
		select {
		case <-ctx.Done():
			return false
		case chunk, ok := <-streamCh:
			if !ok {
				return false
			}
			if chunk.Done {
				return false
			}
			if chunk.Error != nil {
				c.SSEvent("error", chunk.Error.Message)
				flusher.Flush()
				return false
			}
			data, _ := json.Marshal(chunk)
			c.SSEvent("message", string(data))
			flusher.Flush()
			return true
		}
	})
}

func (h *APIHandler) ListModels(c *gin.Context) {
	channels, err := h.channelService.GetActiveChannels()
	if err != nil || len(channels) == 0 {
		c.JSON(http.StatusOK, adapter.ModelsResponse{
			Object: "list",
			Data: []struct {
				ID         string `json:"id"`
				Object     string `json:"object"`
				Created    int    `json:"created"`
				OwnedBy    string `json:"owned_by"`
				Permission []struct {
					ID                  string      `json:"id"`
					Object              string      `json:"object"`
					Created             int         `json:"created"`
					AllowCreateEngine   bool        `json:"allow_create_engine"`
					AllowSampling       bool        `json:"allow_sampling"`
					AllowLogprobs       bool        `json:"allow_logprobs"`
					AllowSearchIndices  bool        `json:"allow_search_indices"`
					AllowView           bool        `json:"allow_view"`
					AllowFineTuning     bool        `json:"allow_fine_tuning"`
					Organization        string      `json:"organization"`
					Group               interface{} `json:"group"`
					IsBlocking          bool        `json:"is_blocking"`
				} `json:"permission,omitempty"`
			}{},
		})
		return
	}

	modelMap := make(map[string]struct {
		id       string
		ownedBy  string
		created  int
	})
	now := int(time.Now().Unix())

	for _, ch := range channels {
		for _, m := range ch.GetModels() {
			if _, exists := modelMap[m]; !exists {
				chatAdapter, err := adapter.GetAdapter(ch.Type)
				if err != nil {
					modelMap[m] = struct {
						id       string
						ownedBy  string
						created  int
					}{id: m, ownedBy: ch.Type, created: now}
					continue
				}
				modelMap[m] = struct {
					id       string
					ownedBy  string
					created  int
				}{id: m, ownedBy: chatAdapter.GetName(), created: now}
			}
		}
	}

	models := make([]struct {
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
	}, 0, len(modelMap))

	for m, info := range modelMap {
		models = append(models, struct {
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
			ID:      m,
			Object:  "model",
			Created: info.created,
			OwnedBy: info.ownedBy,
		})
	}

	c.JSON(http.StatusOK, map[string]interface{}{
		"object": "list",
		"data":   models,
	})
}

func (h *APIHandler) Embeddings(c *gin.Context) {
	var req model.EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "INVALID_REQUEST",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}

	selectedChannel, err := h.channelService.SelectChannel(req.Model)
	if err != nil {
		c.JSON(http.StatusServiceUnavailable, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "NO_CHANNEL",
				Message: "no available channel",
			},
		})
		return
	}

	apiKey, err := crypto.Decrypt(selectedChannel.APIKeyEncrypted)
	if err != nil {
		apiKey = selectedChannel.APIKeyEncrypted
	}

	channel := &adapter.Channel{
		ID:        selectedChannel.ID,
		Type:      selectedChannel.Type,
		BaseURL:   selectedChannel.BaseURL,
		APIKey:    apiKey,
		Timeout:   120,
	}

	chatAdapter, err := adapter.GetAdapter(selectedChannel.Type)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "ADAPTER_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	embedReq := &adapter.EmbeddingsRequest{
		Model: req.Model,
		Input: req.Input,
	}

	resp, err := chatAdapter.Embeddings(ctx, channel, embedReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "UPSTREAM_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) logUsage(c *gin.Context, modelName string, channelID uint, resp *http.Response) {
	var usage struct {
		Usage struct {
			PromptTokens     int `json:"prompt_tokens"`
			CompletionTokens int `json:"completion_tokens"`
			TotalTokens      int `json:"total_tokens"`
		} `json:"usage"`
	}

	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &usage)

	_ = modelName
	_ = channelID
	_ = usage
}

func parseSSEStream(body io.Reader, handler func(string)) error {
	scanner := bufio.NewScanner(body)
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "data: ") {
			data := strings.TrimPrefix(line, "data: ")
			if data == "[DONE]" {
				handler(data)
				return nil
			}
			handler(data)
		}
	}
	return scanner.Err()
}

func getChannelID(c *gin.Context) uint {
	if id, exists := c.Get("channel_id"); exists {
		return id.(uint)
	}
	return 0
}

func setChannelID(c *gin.Context, id uint) {
	c.Set("channel_id", id)
}

func getModelName(req interface{}) string {
	switch v := req.(type) {
	case *model.ChatCompletionsRequest:
		return v.Model
	case map[string]interface{}:
		if model, ok := v["model"].(string); ok {
			return model
		}
	}
	return ""
}

func getUserID(c *gin.Context) uint {
	if id, exists := c.Get("user_id"); exists {
		return id.(uint)
	}
	return 0
}

func getTokenID(c *gin.Context) uint {
	if id, exists := c.Get("token_id"); exists {
		return id.(uint)
	}
	return 0
}

func parseIntParam(params map[string]string, key string, defaultVal int) int {
	if val, exists := params[key]; exists {
		if intVal, err := strconv.Atoi(val); err == nil {
			return intVal
		}
	}
	return defaultVal
}

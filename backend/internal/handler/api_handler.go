package handler

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/adapter"
	"gapi-platform/internal/pkg/crypto"
	"gapi-platform/internal/pkg/logger"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type APIHandler struct {
	tokenService    *service.TokenService
	channelService  *service.ChannelService
	userRepo        *repository.UserRepository
	modelGroupSvc   *service.ModelGroupService
	billingService  *service.BillingService
}

func NewAPIHandler(tokenService *service.TokenService, channelService *service.ChannelService, userRepo *repository.UserRepository, modelGroupSvc *service.ModelGroupService, billingService *service.BillingService) *APIHandler {
	return &APIHandler{
		tokenService:    tokenService,
		channelService:  channelService,
		userRepo:        userRepo,
		modelGroupSvc:   modelGroupSvc,
		billingService:  billingService,
	}
}

const maxChannelRetries = 3

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

	// T-08: Set model in context for APIAccessLog middleware
	c.Set("request_model", req.Model)

	// Pre-check quota
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			estimatedTokens := h.estimateChatTokens(req.Messages, req.MaxTokens)
			if err := h.billingService.PreConsumeQuota(userID, tokenID, req.Model, estimatedTokens); err != nil {
				if err == service.ErrQuotaInsufficient {
					c.JSON(http.StatusPaymentRequired, model.APIErrorResponse{
						Error: &model.APIError{
							Code:    "QUOTA_INSUFFICIENT",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Code:    "TOKEN_INVALID",
							Message: err.Error(),
						},
					})
					return
				}
			}
		}
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

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	if req.Stream {
		h.handleStreamWithFailover(ctx, c, chatReq)
		return
	}

	resp, err := h.chatWithFailover(ctx, chatReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "UPSTREAM_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// T-04: Post-consume quota for non-streaming response
	if chatResp, ok := resp.(*adapter.ChatResponse); ok && h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			h.billingService.PostConsumeQuota(userID, tokenID, req.Model,
				chatResp.Usage.PromptTokens, chatResp.Usage.CompletionTokens)
			h.billingService.LogUsage(&service.UsageRecord{
				UserID:           userID,
				TokenID:          tokenID,
				Model:            req.Model,
				PromptTokens:     chatResp.Usage.PromptTokens,
				CompletionTokens: chatResp.Usage.CompletionTokens,
				TotalTokens:      chatResp.Usage.TotalTokens,
			})
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) chatWithFailover(ctx context.Context, chatReq *adapter.ChatRequest) (interface{}, error) {
	var lastErr error
	attemptedChannels := make(map[uint]bool)

	for attempt := 0; attempt < maxChannelRetries; attempt++ {
		selectedChannel, err := h.channelService.SelectChannel(chatReq.Model)
		if err != nil {
			return nil, fmt.Errorf("no available channel for model: %s", chatReq.Model)
		}

		if attemptedChannels[selectedChannel.ID] {
			continue
		}
		attemptedChannels[selectedChannel.ID] = true

		apiKey, err := crypto.Decrypt(selectedChannel.APIKeyEncrypted)
		if err != nil {
			logger.Errorf("failed to decrypt API key for channel %d: %v", selectedChannel.ID, err)
			continue
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

		chatAdapter, err := adapter.GetAdapter(selectedChannel.Type)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := chatAdapter.Chat(ctx, channel, chatReq)
		if err != nil {
			h.channelService.IncrementFailureCount(selectedChannel.ID)
			lastErr = err
			continue
		}

		h.channelService.ResetFailureCount(selectedChannel.ID)
		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all channels failed after %d attempts: %v", maxChannelRetries, lastErr)
	}
	return nil, fmt.Errorf("no available channel")
}

func (h *APIHandler) handleStreamWithFailover(ctx context.Context, c *gin.Context, chatReq *adapter.ChatRequest) {
	attemptedChannels := make(map[uint]bool)

	for attempt := 0; attempt < maxChannelRetries; attempt++ {
		selectedChannel, err := h.channelService.SelectChannel(chatReq.Model)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, model.APIErrorResponse{
				Error: &model.APIError{
					Code:    "NO_CHANNEL",
					Message: "no available channel for model: " + chatReq.Model,
				},
			})
			return
		}

		if attemptedChannels[selectedChannel.ID] {
			continue
		}
		attemptedChannels[selectedChannel.ID] = true

		apiKey, err := crypto.Decrypt(selectedChannel.APIKeyEncrypted)
		if err != nil {
			logger.Errorf("failed to decrypt API key for channel %d: %v", selectedChannel.ID, err)
			continue
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

		chatAdapter, err := adapter.GetAdapter(selectedChannel.Type)
		if err != nil {
			continue
		}

		// Test the channel connection before streaming
		streamCh, err := chatAdapter.ChatStream(ctx, channel, chatReq)
		if err != nil {
			h.channelService.IncrementFailureCount(selectedChannel.ID)
			continue
		}

		// Channel works, set headers and stream
		c.Header("Content-Type", "text/event-stream")
		c.Header("Cache-Control", "no-cache")
		c.Header("Connection", "keep-alive")
		c.Header("Transfer-Encoding", "chunked")

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

		// T-03: Track usage for streaming quota consumption
		var (
			upstreamPromptTokens     int
			upstreamCompletionTokens int
			upstreamTotalTokens      int
			hasUpstreamUsage         bool
			completionContentLen     int
		)

		c.Stream(func(w io.Writer) bool {
			select {
			case <-ctx.Done():
				return false
			case chunk, ok := <-streamCh:
				if !ok {
					return false
				}
				if chunk.Done {
					// Capture usage from final usage-only chunk (if upstream sends it)
					if chunk.Usage != nil {
						upstreamPromptTokens = chunk.Usage.PromptTokens
						upstreamCompletionTokens = chunk.Usage.CompletionTokens
						upstreamTotalTokens = chunk.Usage.TotalTokens
						hasUpstreamUsage = true
					}
					return false
				}
				if chunk.Error != nil {
					c.SSEvent("error", chunk.Error.Message)
					flusher.Flush()
					return false
				}
				// Accumulate streamed content length for fallback estimation
				for _, choice := range chunk.Choices {
					completionContentLen += len(choice.Delta.Content)
				}
				data, _ := json.Marshal(chunk)
				c.SSEvent("message", string(data))
				flusher.Flush()
				return true
			}
		})

		// T-03: Post-consume quota after stream completes
		if h.billingService != nil {
			userID := getUserID(c)
			tokenID := getTokenID(c)
			if userID > 0 && tokenID > 0 {
				var promptTokens, completionTokens int
				if hasUpstreamUsage {
					promptTokens = upstreamPromptTokens
					completionTokens = upstreamCompletionTokens
				} else {
					// Fallback: estimate from request content and accumulated output
					promptTokens = h.estimateChatTokens(chatReq.Messages, 0)
					completionTokens = completionContentLen/4 + 1
				}
				h.billingService.PostConsumeQuota(userID, tokenID, chatReq.Model, promptTokens, completionTokens)
				h.billingService.LogUsage(&service.UsageRecord{
					UserID:           userID,
					TokenID:          tokenID,
					ChannelID:        selectedChannel.ID,
					Model:            chatReq.Model,
					PromptTokens:     promptTokens,
					CompletionTokens: completionTokens,
					TotalTokens:      promptTokens + completionTokens,
				})
			}
		}

		h.channelService.ResetFailureCount(selectedChannel.ID)
		return
	}

	c.JSON(http.StatusBadGateway, model.APIErrorResponse{
		Error: &model.APIError{
			Code:    "UPSTREAM_ERROR",
			Message: "all channels failed after retries",
		},
	})
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
	userID, _ := c.Get("user_id")
	var channels []model.Channel
	var err error

	if h.modelGroupSvc != nil && userID != nil {
		uid := userID.(uint)
		channelIDs, err := h.modelGroupSvc.GetAvailableChannelIDsForUser(uid)
		if err != nil || len(channelIDs) == 0 {
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": []interface{}{}})
			return
		}
		channels, err = h.channelService.GetActiveChannelsByIDs(channelIDs)
		if err != nil || len(channels) == 0 {
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": []interface{}{}})
			return
		}
	} else {
		channels, err = h.channelService.GetActiveChannels()
		if err != nil || len(channels) == 0 {
			c.JSON(http.StatusOK, gin.H{"object": "list", "data": []interface{}{}})
			return
		}
	}

	modelMap := make(map[string]struct {
		id      string
		ownedBy string
		created int
	})
	now := int(time.Now().Unix())

	for _, ch := range channels {
		for _, m := range ch.GetModels() {
			if _, exists := modelMap[m]; !exists {
				chatAdapter, err := adapter.GetAdapter(ch.Type)
				if err != nil {
					modelMap[m] = struct {
						id      string
						ownedBy string
						created int
					}{id: m, ownedBy: ch.Type, created: now}
					continue
				}
				modelMap[m] = struct {
					id      string
					ownedBy string
					created int
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
			ID                 string      `json:"id"`
			Object             string      `json:"object"`
			Created            int         `json:"created"`
			AllowCreateEngine  bool        `json:"allow_create_engine"`
			AllowSampling      bool        `json:"allow_sampling"`
			AllowLogprobs      bool        `json:"allow_logprobs"`
			AllowSearchIndices bool        `json:"allow_search_indices"`
			AllowView          bool        `json:"allow_view"`
			AllowFineTuning    bool        `json:"allow_fine_tuning"`
			Organization       string      `json:"organization"`
			Group              interface{} `json:"group"`
			IsBlocking         bool        `json:"is_blocking"`
		} `json:"permission,omitempty"`
	}, 0, len(modelMap))

	for m, info := range modelMap {
		models = append(models, struct {
			ID         string `json:"id"`
			Object     string `json:"object"`
			Created    int    `json:"created"`
			OwnedBy    string `json:"owned_by"`
			Permission []struct {
				ID                 string      `json:"id"`
				Object             string      `json:"object"`
				Created            int         `json:"created"`
				AllowCreateEngine  bool        `json:"allow_create_engine"`
				AllowSampling      bool        `json:"allow_sampling"`
				AllowLogprobs      bool        `json:"allow_logprobs"`
				AllowSearchIndices bool        `json:"allow_search_indices"`
				AllowView          bool        `json:"allow_view"`
				AllowFineTuning    bool        `json:"allow_fine_tuning"`
				Organization       string      `json:"organization"`
				Group              interface{} `json:"group"`
				IsBlocking         bool        `json:"is_blocking"`
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

	// T-08: Set model in context for APIAccessLog middleware
	c.Set("request_model", req.Model)

	// Pre-check quota
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			estimatedTokens := h.estimateEmbeddingTokens(req.Input)
			if err := h.billingService.PreConsumeQuota(userID, tokenID, req.Model, estimatedTokens); err != nil {
				if err == service.ErrQuotaInsufficient {
					c.JSON(http.StatusPaymentRequired, model.APIErrorResponse{
						Error: &model.APIError{
							Code:    "QUOTA_INSUFFICIENT",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Code:    "TOKEN_INVALID",
							Message: err.Error(),
						},
					})
					return
				}
			}
		}
	}

	embedReq := &adapter.EmbeddingsRequest{
		Model: req.Model,
		Input: req.Input,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	resp, err := h.embeddingsWithFailover(ctx, embedReq)
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Code:    "UPSTREAM_ERROR",
				Message: err.Error(),
			},
		})
		return
	}

	// Post-consume quota for embeddings
	if embedResp, ok := resp.(*adapter.EmbeddingsResponse); ok && h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			h.billingService.PostConsumeQuota(userID, tokenID, req.Model,
				embedResp.Usage.PromptTokens, 0)
			h.billingService.LogUsage(&service.UsageRecord{
				UserID:           userID,
				TokenID:          tokenID,
				Model:            req.Model,
				PromptTokens:     embedResp.Usage.PromptTokens,
				CompletionTokens: 0,
				TotalTokens:      embedResp.Usage.TotalTokens,
			})
		}
	}

	c.JSON(http.StatusOK, resp)
}

func (h *APIHandler) embeddingsWithFailover(ctx context.Context, embedReq *adapter.EmbeddingsRequest) (interface{}, error) {
	var lastErr error
	attemptedChannels := make(map[uint]bool)

	for attempt := 0; attempt < maxChannelRetries; attempt++ {
		selectedChannel, err := h.channelService.SelectChannel(embedReq.Model)
		if err != nil {
			return nil, fmt.Errorf("no available channel for model: %s", embedReq.Model)
		}

		if attemptedChannels[selectedChannel.ID] {
			continue
		}
		attemptedChannels[selectedChannel.ID] = true

		apiKey, err := crypto.Decrypt(selectedChannel.APIKeyEncrypted)
		if err != nil {
			logger.Errorf("failed to decrypt API key for channel %d: %v", selectedChannel.ID, err)
			continue
		}

		channel := &adapter.Channel{
			ID:      selectedChannel.ID,
			Type:    selectedChannel.Type,
			BaseURL: selectedChannel.BaseURL,
			APIKey:  apiKey,
			Timeout: 120,
		}

		chatAdapter, err := adapter.GetAdapter(selectedChannel.Type)
		if err != nil {
			lastErr = err
			continue
		}

		resp, err := chatAdapter.Embeddings(ctx, channel, embedReq)
		if err != nil {
			h.channelService.IncrementFailureCount(selectedChannel.ID)
			lastErr = err
			continue
		}

		h.channelService.ResetFailureCount(selectedChannel.ID)
		return resp, nil
	}

	if lastErr != nil {
		return nil, fmt.Errorf("all channels failed after %d attempts: %v", maxChannelRetries, lastErr)
	}
	return nil, fmt.Errorf("no available channel")
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
	if err := json.Unmarshal(body, &usage); err != nil {
		logger.Warnf("Failed to parse usage from response: %v", err)
		return
	}

	logger.Infof("API usage logged: model=%s channel_id=%d prompt_tokens=%d completion_tokens=%d total_tokens=%d",
		modelName, channelID, usage.Usage.PromptTokens, usage.Usage.CompletionTokens, usage.Usage.TotalTokens)
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

func (h *APIHandler) estimateChatTokens(messages []map[string]string, maxTokens int) int {
	totalChars := 0
	for _, m := range messages {
		totalChars += len(m["content"]) + len(m["role"]) + 10
	}
	estimated := totalChars / 4
	if maxTokens > 0 {
		estimated += maxTokens
	}
	if estimated < 100 {
		estimated = 100
	}
	return estimated
}

func (h *APIHandler) estimateEmbeddingTokens(input interface{}) int {
	switch v := input.(type) {
	case string:
		return len(v)/4 + 1
	case []interface{}:
		total := 0
		for _, item := range v {
			if s, ok := item.(string); ok {
				total += len(s)/4 + 1
			}
		}
		if total < 1 {
			total = 1
		}
		return total
	default:
		return 100
	}
}

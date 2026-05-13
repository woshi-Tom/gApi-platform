package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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
				Type:    "invalid_request_error",
				Code:    "invalid_request",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "invalid_request_error",
				Code:    "missing_model",
				Message: "model is required",
			},
		})
		return
	}

	// Set model in context for APIAccessLog middleware
	c.Set("request_model", req.Model)

	// Normalize messages: convert content arrays to strings (OpenAI spec compatibility)
	normalizedMessages := normalizeMessages(req.Messages)

	// Pre-check quota
	preCheckStart := time.Now()
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			estimatedTokens := h.estimateChatTokens(normalizedMessages, req.MaxTokens)
			if err := h.billingService.PreConsumeQuota(userID, tokenID, req.Model, estimatedTokens); err != nil {
				if err == service.ErrQuotaInsufficient {
					c.JSON(http.StatusPaymentRequired, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "quota_insufficient",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "token_invalid",
							Message: err.Error(),
						},
					})
					return
				}
			}
		}
	}
	c.Set("pre_check_duration_ms", int(time.Since(preCheckStart).Milliseconds()))

	chatReq := &adapter.ChatRequest{
		Model:       req.Model,
		Messages:    normalizedMessages,
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

	upstreamStart := time.Now()
	resp, err := h.chatWithFailover(ctx, chatReq)
	c.Set("upstream_duration_ms", int(time.Since(upstreamStart).Milliseconds()))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "server_error",
				Code:    "upstream_error",
				Message: err.Error(),
			},
		})
		return
	}

	// Post-consume quota for non-streaming response
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

// Messages handles Anthropic Messages API requests (POST /api/v1/messages)
func (h *APIHandler) Messages(c *gin.Context) {
	var req model.AnthropicMessagesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.AnthropicErrorResponse{
			Type: "error",
			Error: model.AnthropicErrorBody{
				Type:    "invalid_request_error",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, model.AnthropicErrorResponse{
			Type: "error",
			Error: model.AnthropicErrorBody{
				Type:    "invalid_request_error",
				Message: "model is required",
			},
		})
		return
	}

	// Set model in context for APIAccessLog middleware
	c.Set("request_model", req.Model)

	// Convert Anthropic request to internal ChatRequest
	chatReq := anthropicToOpenAIChatRequest(&req)

	// Pre-check quota
	preCheckStart := time.Now()
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			estimatedTokens := h.estimateChatTokens(chatReq.Messages, req.MaxTokens)
			if err := h.billingService.PreConsumeQuota(userID, tokenID, req.Model, estimatedTokens); err != nil {
				if err == service.ErrQuotaInsufficient {
					c.JSON(http.StatusPaymentRequired, model.AnthropicErrorResponse{
						Type: "error",
						Error: model.AnthropicErrorBody{
							Type:    "invalid_request_error",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.AnthropicErrorResponse{
						Type: "error",
						Error: model.AnthropicErrorBody{
							Type:    "authentication_error",
							Message: err.Error(),
						},
					})
					return
				}
			}
		}
	}
	c.Set("pre_check_duration_ms", int(time.Since(preCheckStart).Milliseconds()))

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	if req.Stream {
		h.handleMessagesStream(ctx, c, chatReq)
		return
	}

	// Non-streaming: reuse chatWithFailover
	upstreamStart := time.Now()
	resp, err := h.chatWithFailover(ctx, chatReq)
	c.Set("upstream_duration_ms", int(time.Since(upstreamStart).Milliseconds()))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.AnthropicErrorResponse{
			Type: "error",
			Error: model.AnthropicErrorBody{
				Type:    "api_error",
				Message: err.Error(),
			},
		})
		return
	}

	chatResp, ok := resp.(*adapter.ChatResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.AnthropicErrorResponse{
			Type: "error",
			Error: model.AnthropicErrorBody{
				Type:    "api_error",
				Message: "invalid response type",
			},
		})
		return
	}

	// Post-consume quota
	if h.billingService != nil {
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

	c.JSON(http.StatusOK, openAIToAnthropicResponse(chatResp))
}

// handleMessagesStream handles streaming for Anthropic Messages API.
func (h *APIHandler) handleMessagesStream(ctx context.Context, c *gin.Context, chatReq *adapter.ChatRequest) {
	// Phase 1: Send SSE headers and initial events immediately
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.AnthropicErrorResponse{
			Type: "error",
			Error: model.AnthropicErrorBody{
				Type:    "api_error",
				Message: "streaming not supported",
			},
		})
		return
	}

	msgID := fmt.Sprintf("msg_%016x", time.Now().UnixNano())
	estimatedInput := h.estimateChatTokens(chatReq.Messages, 0)
	writeAnthropicSSEvent(c.Writer, "message_start", model.AnthropicStreamEvent{
		Type: "message_start",
		Message: &model.AnthropicMessagesResponse{
			ID:      msgID,
			Type:    "message",
			Role:    "assistant",
			Content: []model.AnthropicResponseContent{},
			Model:   chatReq.Model,
			Usage:   model.AnthropicUsage{InputTokens: estimatedInput},
		},
	})
	writeAnthropicSSEvent(c.Writer, "content_block_start", model.AnthropicStreamEvent{
		Type:  "content_block_start",
		Index: 0,
		ContentBlock: &model.AnthropicResponseContent{
			Type: "text",
			Text: "",
		},
	})
	flusher.Flush()

	// Phase 2: Retry loop with channel failover
	attemptedChannels := make(map[uint]bool)

	for attempt := 0; attempt < maxChannelRetries; attempt++ {
		selectedChannel, err := h.channelService.SelectChannel(chatReq.Model)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, model.AnthropicErrorResponse{
				Type: "error",
				Error: model.AnthropicErrorBody{
					Type:    "api_error",
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

		streamCh, err := chatAdapter.ChatStream(ctx, channel, chatReq)
		if err != nil {
			h.channelService.IncrementFailureCount(selectedChannel.ID)
			continue
		}

		// Track usage
		var (
			upstreamPromptTokens     int
			upstreamCompletionTokens int
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
					if chunk.Usage != nil {
						upstreamPromptTokens = chunk.Usage.PromptTokens
						upstreamCompletionTokens = chunk.Usage.CompletionTokens
						hasUpstreamUsage = true
					}
					return false
				}
				if chunk.Error != nil {
					writeAnthropicSSEvent(w, "error", model.AnthropicErrorBody{
						Type:    "api_error",
						Message: chunk.Error.Message,
					})
					flusher.Flush()
					return false
				}
				for _, choice := range chunk.Choices {
					if choice.Delta.Content != "" {
						completionContentLen += len(choice.Delta.Content)
						writeAnthropicSSEvent(w, "content_block_delta", model.AnthropicStreamEvent{
							Type:  "content_block_delta",
							Index: 0,
							Delta: &model.AnthropicStreamDelta{
								Type: "text_delta",
								Text: choice.Delta.Content,
							},
						})
						flusher.Flush()
					}
					if choice.FinishReason != "" {
						stopReason := mapStopReason(choice.FinishReason)
						writeAnthropicSSEvent(w, "content_block_stop", model.AnthropicStreamEvent{
							Type:  "content_block_stop",
							Index: 0,
						})
						outputTokens := upstreamCompletionTokens
						if !hasUpstreamUsage {
							outputTokens = completionContentLen/4 + 1
						}
						writeAnthropicSSEvent(w, "message_delta", model.AnthropicStreamEvent{
							Type: "message_delta",
							Delta: &model.AnthropicStreamDelta{
								Type:       "message_delta",
								StopReason: &stopReason,
							},
							Usage: &model.AnthropicUsage{
								OutputTokens: outputTokens,
							},
						})
						flusher.Flush()
					}
				}
				return true
			}
		})

		// Emit message_stop
		writeAnthropicSSEvent(c.Writer, "message_stop", model.AnthropicStreamEvent{Type: "message_stop"})

		// Post-consume quota after stream completes
		if h.billingService != nil {
			userID := getUserID(c)
			tokenID := getTokenID(c)
			if userID > 0 && tokenID > 0 {
				var promptTokens, completionTokens int
				if hasUpstreamUsage {
					promptTokens = upstreamPromptTokens
					completionTokens = upstreamCompletionTokens
				} else {
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

	// All channels failed — notify via SSE since headers are already sent
	writeAnthropicSSEvent(c.Writer, "error", model.AnthropicErrorBody{
		Type:    "api_error",
		Message: "all channels failed after retries",
	})
	flusher.Flush()
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
	// Phase 1: Send SSE headers immediately
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")

	flusher, ok := c.Writer.(http.Flusher)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "server_error",
				Code:    "internal_error",
				Message: "streaming not supported",
			},
		})
		return
	}

	// SSE comment triggers header flush so client knows connection is established
	fmt.Fprintf(c.Writer, ":ok\n\n")
	flusher.Flush()

	// Phase 2: Retry loop with channel failover
	attemptedChannels := make(map[uint]bool)

	for attempt := 0; attempt < maxChannelRetries; attempt++ {
		selectedChannel, err := h.channelService.SelectChannel(chatReq.Model)
		if err != nil {
			c.JSON(http.StatusServiceUnavailable, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "server_error",
					Code:    "no_available_channel",
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

		// Track usage for streaming quota consumption
		var (
			upstreamPromptTokens     int
			upstreamCompletionTokens int
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
					if chunk.Usage != nil {
						upstreamPromptTokens = chunk.Usage.PromptTokens
						upstreamCompletionTokens = chunk.Usage.CompletionTokens
						hasUpstreamUsage = true
					}
					return false
				}
				if chunk.Error != nil {
					c.SSEvent("error", chunk.Error.Message)
					flusher.Flush()
					return false
				}
				for _, choice := range chunk.Choices {
					completionContentLen += len(choice.Delta.Content)
				}
				data, _ := json.Marshal(chunk)
				c.SSEvent("message", string(data))
				flusher.Flush()
				return true
			}
		})

		// Post-consume quota after stream completes
		if h.billingService != nil {
			userID := getUserID(c)
			tokenID := getTokenID(c)
			if userID > 0 && tokenID > 0 {
				var promptTokens, completionTokens int
				if hasUpstreamUsage {
					promptTokens = upstreamPromptTokens
					completionTokens = upstreamCompletionTokens
				} else {
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

	// All channels failed — notify via SSE since headers are already sent
	c.SSEvent("error", "all channels failed after retries")
	flusher.Flush()
}

func (h *APIHandler) ListModels(c *gin.Context) {
	userID, _ := c.Get("user_id")
	var channels []model.Channel
	var err error

	if h.modelGroupSvc != nil && userID != nil {
		uid, ok := userID.(uint)
		if !ok || uid == 0 {
			channels, err = h.channelService.GetActiveChannels()
		} else {
			channelIDs, err := h.modelGroupSvc.GetAvailableChannelIDsForUser(uid)
			if err != nil || len(channelIDs) == 0 {
				c.JSON(http.StatusOK, gin.H{"object": "list", "data": []interface{}{}})
				return
			}
			channels, err = h.channelService.GetActiveChannelsByIDs(channelIDs)
		}
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

// Completions handles OpenAI text completions (legacy API)
func (h *APIHandler) Completions(c *gin.Context) {
	preCheckStart := time.Now()
	var req model.CompletionsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "invalid_request_error",
				Code:    "invalid_request",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "invalid_request_error",
				Code:    "missing_model",
				Message: "model is required",
			},
		})
		return
	}

	// Convert prompt to messages format
	var promptStr string
	switch p := req.Prompt.(type) {
	case string:
		promptStr = p
	case []interface{}:
		for _, item := range p {
			if s, ok := item.(string); ok {
				promptStr += s
			}
		}
	}

	messages := []map[string]string{
		{"role": "user", "content": promptStr},
	}

	// Pre-check quota
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 {
			estimatedTokens := h.estimateChatTokens(messages, req.MaxTokens)
			if err := h.billingService.PreConsumeQuota(userID, tokenID, req.Model, estimatedTokens); err != nil {
				if err == service.ErrQuotaInsufficient {
					c.JSON(http.StatusPaymentRequired, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "quota_insufficient",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "token_invalid",
							Message: err.Error(),
						},
					})
					return
				}
			}
		}
	}
	c.Set("pre_check_duration_ms", int(time.Since(preCheckStart).Milliseconds()))

	chatReq := &adapter.ChatRequest{
		Model:       req.Model,
		Messages:    messages,
		Temperature: req.Temperature,
		MaxTokens:   req.MaxTokens,
		TopP:        req.TopP,
		User:        req.User,
	}

	ctx, cancel := context.WithTimeout(c.Request.Context(), 120*time.Second)
	defer cancel()

	upstreamStart := time.Now()
	resp, err := h.chatWithFailover(ctx, chatReq)
	c.Set("upstream_duration_ms", int(time.Since(upstreamStart).Milliseconds()))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "server_error",
				Code:    "upstream_error",
				Message: err.Error(),
			},
		})
		return
	}

	// Convert chat response to completions format
	chatResp, ok := resp.(*adapter.ChatResponse)
	if !ok {
		c.JSON(http.StatusInternalServerError, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "server_error",
				Code:    "internal_error",
				Message: "invalid response type",
			},
		})
		return
	}

	type completionChoice struct {
		Text         string `json:"text"`
		Index        int    `json:"index"`
		Logprobs     *int   `json:"logprobs"`
		FinishReason string `json:"finish_reason"`
	}

	choices := make([]completionChoice, len(chatResp.Choices))
	for i, choice := range chatResp.Choices {
		choices[i] = completionChoice{
			Text:         choice.Message.Content,
			Index:        i,
			Logprobs:     nil,
			FinishReason: choice.FinishReason,
		}
	}

	completionsResp := map[string]interface{}{
		"id":      chatResp.ID,
		"object":  "text_completion",
		"created": chatResp.Created,
		"model":   chatResp.Model,
		"choices": choices,
	}
	if chatResp.Usage.TotalTokens > 0 {
		completionsResp["usage"] = chatResp.Usage
	}

	// Post-consume quota
	if h.billingService != nil {
		userID := getUserID(c)
		tokenID := getTokenID(c)
		if userID > 0 && tokenID > 0 && chatResp.Usage.TotalTokens > 0 {
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

	c.JSON(http.StatusOK, completionsResp)
}

func (h *APIHandler) Embeddings(c *gin.Context) {
	var req model.EmbeddingsRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "invalid_request_error",
				Code:    "invalid_request",
				Message: err.Error(),
			},
		})
		return
	}

	if req.Model == "" {
		req.Model = "text-embedding-ada-002"
	}

	// Set model in context for APIAccessLog middleware
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
							Type:    "invalid_request_error",
							Code:    "quota_insufficient",
							Message: "insufficient quota",
						},
					})
					return
				}
				if err == service.ErrTokenDisabled || err == service.ErrTokenExpired {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "token_invalid",
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

	upstreamStart := time.Now()
	resp, err := h.embeddingsWithFailover(ctx, embedReq)
	c.Set("upstream_duration_ms", int(time.Since(upstreamStart).Milliseconds()))
	if err != nil {
		c.JSON(http.StatusBadGateway, model.APIErrorResponse{
			Error: &model.APIError{
				Type:    "server_error",
				Code:    "upstream_error",
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

// normalizeMessages converts OpenAI content format (string or content block array)
// to the flat string format used internally by adapters.
func normalizeMessages(msgs []map[string]interface{}) []map[string]string {
	result := make([]map[string]string, 0, len(msgs))
	for _, m := range msgs {
		normalized := make(map[string]string)
		for k, v := range m {
			if k == "content" {
				normalized[k] = extractTextContent(v)
			} else if s, ok := v.(string); ok {
				normalized[k] = s
			} else {
				normalized[k] = fmt.Sprintf("%v", v)
			}
		}
		result = append(result, normalized)
	}
	return result
}

// extractTextContent extracts text from either a string or an OpenAI content block array.
func extractTextContent(v interface{}) string {
	switch val := v.(type) {
	case string:
		return val
	case []interface{}:
		text := ""
		for _, part := range val {
			if block, ok := part.(map[string]interface{}); ok {
				if block["type"] == "text" {
					if t, ok := block["text"].(string); ok {
						if text != "" {
							text += "\n"
						}
						text += t
					}
				}
			}
		}
		return text
	default:
		return fmt.Sprintf("%v", v)
	}
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

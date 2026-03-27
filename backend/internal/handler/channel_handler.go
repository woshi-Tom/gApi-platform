package handler

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strconv"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// ChannelHandler handles channel-related endpoints
type ChannelHandler struct {
	channelService *service.ChannelService
	auditRepo      *repository.AuditRepository
}

// NewChannelHandler creates a new channel handler
func NewChannelHandler(channelService *service.ChannelService, auditRepo *repository.AuditRepository) *ChannelHandler {
	return &ChannelHandler{
		channelService: channelService,
		auditRepo:      auditRepo,
	}
}

// List returns all channels
func (h *ChannelHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	channelType := c.Query("type")
	status := c.Query("status")
	group := c.Query("group")
	keyword := c.Query("keyword")

	channels, total, err := h.channelService.List(page, pageSize, channelType, status, group, keyword)
	if err != nil {
		response.InternalError(c, "failed to list channels")
		return
	}

	response.Paginated(c, channels, page, pageSize, total)
}

// Create creates a new channel
func (h *ChannelHandler) Create(c *gin.Context) {
	var req struct {
		Name     string   `json:"name" binding:"required"`
		Type     string   `json:"type" binding:"required"`
		BaseURL  string   `json:"base_url" binding:"required"`
		APIKey   string   `json:"api_key" binding:"required"`
		Models   []string `json:"models"`
		Weight   int      `json:"weight"`
		Priority int      `json:"priority"`
		Group    string   `json:"group_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	// Encrypt API key
	encryptedKey, err := encryptAPIKey(req.APIKey)
	if err != nil {
		response.InternalError(c, "failed to encrypt api key")
		return
	}

	channel := &model.Channel{
		Name:            req.Name,
		Type:            req.Type,
		BaseURL:         req.BaseURL,
		APIKeyEncrypted: encryptedKey,
		Weight:          req.Weight,
		Priority:        req.Priority,
		GroupName:       req.Group,
		Status:          1,
		IsHealthy:       true,
	}

	if len(req.Models) > 0 {
		modelsJSON, _ := json.Marshal(req.Models)
		channel.Models = string(modelsJSON)
	}

	if err := h.channelService.Create(channel); err != nil {
		response.Fail(c, "CHANNEL_CREATE_FAILED", err.Error())
		return
	}

	response.Created(c, channel)
}

// Update updates a channel
func (h *ChannelHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid channel id")
		return
	}

	channel, err := h.channelService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "channel not found")
		return
	}

	var req struct {
		Name     string   `json:"name"`
		BaseURL  string   `json:"base_url"`
		APIKey   string   `json:"api_key"`
		Models   []string `json:"models"`
		Weight   int      `json:"weight"`
		Priority int      `json:"priority"`
		Group    string   `json:"group_name"`
		Status   int      `json:"status"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.BaseURL != "" {
		channel.BaseURL = req.BaseURL
	}
	if req.APIKey != "" {
		encryptedKey, err := encryptAPIKey(req.APIKey)
		if err != nil {
			response.InternalError(c, "failed to encrypt api key")
			return
		}
		channel.APIKeyEncrypted = encryptedKey
	}
	if len(req.Models) > 0 {
		modelsJSON, _ := json.Marshal(req.Models)
		channel.Models = string(modelsJSON)
	}
	if req.Weight > 0 {
		channel.Weight = req.Weight
	}
	if req.Group != "" {
		channel.GroupName = req.Group
	}
	channel.Priority = req.Priority
	channel.Status = req.Status

	if err := h.channelService.Update(channel); err != nil {
		response.InternalError(c, "failed to update channel")
		return
	}

	response.SuccessWithMessage(c, channel, "channel updated")
}

// Delete deletes a channel
func (h *ChannelHandler) Delete(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid channel id")
		return
	}

	if err := h.channelService.Delete(uint(id)); err != nil {
		response.InternalError(c, "failed to delete channel")
		return
	}

	response.SuccessWithMessage(c, nil, "channel deleted")
}

// Test tests a channel connection
func (h *ChannelHandler) Test(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid channel id")
		return
	}

	channel, err := h.channelService.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "channel not found")
		return
	}

	var req model.ChannelTestRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	// Decrypt API key
	apiKey, err := decryptAPIKey(channel.APIKeyEncrypted)
	if err != nil {
		response.InternalError(c, "failed to decrypt api key")
		return
	}

	startTime := time.Now()
	var result model.ChannelTestResponse

	switch req.TestType {
	case "models":
		result = testModels(channel.BaseURL, apiKey)
	case "chat":
		result = testChat(channel.BaseURL, apiKey, &req)
	case "embeddings":
		result = testEmbeddings(channel.BaseURL, apiKey, &req)
	default:
		response.Fail(c, "INVALID_PARAMETER", "unsupported test type")
		return
	}

	result.ResponseTimeMs = time.Since(startTime).Milliseconds()

	// Save test history
	go h.saveTestHistory(channel.ID, c.GetUint("user_id"), &req, &result)

	response.Success(c, result)
}

func testModels(baseURL, apiKey string) model.ChannelTestResponse {
	client := &http.Client{Timeout: 10 * time.Second}
	req, err := http.NewRequest("GET", baseURL+"/v1/models", nil)
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)

	resp, err := client.Do(req)
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		body, _ := io.ReadAll(resp.Body)
		return model.ChannelTestResponse{
			Success:    false,
			StatusCode: resp.StatusCode,
			Error:      string(body),
		}
	}

	var result struct {
		Data []struct {
			ID string `json:"id"`
		} `json:"data"`
	}
	body, _ := io.ReadAll(resp.Body)
	json.Unmarshal(body, &result)

	models := make([]string, 0, len(result.Data))
	for _, m := range result.Data {
		models = append(models, m.ID)
	}

	return model.ChannelTestResponse{
		Success:    true,
		StatusCode: 200,
		Models:     models,
	}
}

func testChat(baseURL, apiKey string, testReq *model.ChannelTestRequest) model.ChannelTestResponse {
	body := map[string]interface{}{
		"model":    testReq.Model,
		"messages": testReq.Messages,
	}
	if testReq.Temperature > 0 {
		body["temperature"] = testReq.Temperature
	}
	if testReq.MaxTokens > 0 {
		body["max_tokens"] = testReq.MaxTokens
	}

	bodyBytes, _ := json.Marshal(body)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/v1/chat/completions", bytes.NewReader(bodyBytes))
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return model.ChannelTestResponse{
			Success:    false,
			StatusCode: resp.StatusCode,
			Error:      string(respBody),
		}
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
		Usage model.Usage `json:"usage"`
	}
	json.Unmarshal(respBody, &result)

	content := ""
	if len(result.Choices) > 0 {
		content = result.Choices[0].Message.Content
	}

	return model.ChannelTestResponse{
		Success:    true,
		StatusCode: 200,
		Content:    content,
		Usage:      &result.Usage,
	}
}

func testEmbeddings(baseURL, apiKey string, testReq *model.ChannelTestRequest) model.ChannelTestResponse {
	body := map[string]interface{}{
		"model": testReq.Model,
		"input": testReq.Input,
	}

	bodyBytes, _ := json.Marshal(body)
	client := &http.Client{Timeout: 30 * time.Second}
	req, err := http.NewRequest("POST", baseURL+"/v1/embeddings", bytes.NewReader(bodyBytes))
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	req.Header.Set("Content-Type", "application/json")

	resp, err := client.Do(req)
	if err != nil {
		return model.ChannelTestResponse{
			Success: false,
			Error:   err.Error(),
		}
	}
	defer resp.Body.Close()

	respBody, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != 200 {
		return model.ChannelTestResponse{
			Success:    false,
			StatusCode: resp.StatusCode,
			Error:      string(respBody),
		}
	}

	var result struct {
		Data []struct {
			Embedding []float64 `json:"embedding"`
		} `json:"data"`
	}
	json.Unmarshal(respBody, &result)

	var embedding []float64
	if len(result.Data) > 0 {
		embedding = result.Data[0].Embedding
	}

	return model.ChannelTestResponse{
		Success:    true,
		StatusCode: 200,
		Embedding:  embedding,
	}
}

func (h *ChannelHandler) saveTestHistory(channelID, userID uint, req *model.ChannelTestRequest, result *model.ChannelTestResponse) {
	reqBody, _ := json.Marshal(req)
	resBody, _ := json.Marshal(result)

	history := &model.ChannelTestHistory{
		ChannelID:      channelID,
		UserID:         userID,
		TestType:       req.TestType,
		Model:          req.Model,
		RequestBody:    string(reqBody),
		StatusCode:     result.StatusCode,
		ResponseBody:   string(resBody),
		ResponseTimeMs: int(result.ResponseTimeMs),
		Success:        result.Success,
		ErrorMessage:   result.Error,
	}

	// Use audit repo for now (in real impl, need test history repo)
	_ = history
	_ = io.EOF
}

// Helper functions for encryption (placeholder - should use pkg/crypto)
func encryptAPIKey(key string) (string, error) {
	// TODO: Use pkg/crypto.Encrypt
	return key, nil
}

func decryptAPIKey(encrypted string) (string, error) {
	// TODO: Use pkg/crypto.Decrypt
	return encrypted, nil
}

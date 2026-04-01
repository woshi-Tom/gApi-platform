package handler

import (
	"strconv"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// TokenHandler handles token-related endpoints
type TokenHandler struct {
	tokenService *service.TokenService
}

// NewTokenHandler creates a new token handler
func NewTokenHandler(tokenService *service.TokenService) *TokenHandler {
	return &TokenHandler{tokenService: tokenService}
}

// List godoc
// @Summary List API tokens
// @Description Get all API tokens for current user with quota info
// @Tags tokens
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/tokens [get]
func (h *TokenHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	tokens, err := h.tokenService.ListByUser(userID)
	if err != nil {
		response.InternalError(c, "failed to list tokens")
		return
	}

	quota, err := h.tokenService.GetUserQuota(userID)
	if err != nil {
		response.InternalError(c, "failed to get quota")
		return
	}

	type TokenWithQuota struct {
		model.Token
		TotalQuota int64 `json:"total_quota"`
	}

	result := make([]TokenWithQuota, len(tokens))
	for i, t := range tokens {
		result[i] = TokenWithQuota{
			Token:      t,
			TotalQuota: quota,
		}
	}

	response.Success(c, result)
}

// Create godoc
// @Summary Create API token
// @Description Create a new API token for current user
// @Tags tokens
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 201 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/tokens [post]
func (h *TokenHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Name          string     `json:"name" binding:"required"`
		AllowedModels []string   `json:"allowed_models"`
		AllowedIPs    []string   `json:"allowed_ips"`
		ExpiresAt     *time.Time `json:"expires_at"`
		RPMLimit      *int       `json:"rpm_limit"`
		TPMLimit      *int       `json:"tpm_limit"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	token, err := h.tokenService.Create(userID, req.Name, req.AllowedModels, req.AllowedIPs, req.ExpiresAt, req.RPMLimit, req.TPMLimit)
	if err != nil {
		if err.Error() == "token limit exceeded" {
			response.Fail(c, "TOKEN_LIMIT_EXCEEDED", "您的API密钥数量已达上限，普通用户仅限1个，VIP用户可创建更多。")
			return
		}
		response.Fail(c, "TOKEN_CREATE_FAILED", err.Error())
		return
	}

	response.Created(c, token)
}

// Delete godoc
// @Summary Delete API token
// @Description Delete an API token
// @Tags tokens
// @Produce json
// @Security BearerAuth
// @Param id path int true "Token ID"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Router /api/v1/tokens/{id} [delete]
func (h *TokenHandler) Delete(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	tokenIDStr := c.Param("id")

	tokenID, err := strconv.ParseUint(tokenIDStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid token id")
		return
	}

	// Verify ownership
	token, err := h.tokenService.GetByID(uint(tokenID))
	if err != nil {
		response.NotFound(c, "token not found")
		return
	}

	if token.UserID != userID {
		response.Forbidden(c, "not your token")
		return
	}

	if err := h.tokenService.Delete(uint(tokenID)); err != nil {
		response.InternalError(c, "failed to delete token")
		return
	}

	response.SuccessWithMessage(c, nil, "token deleted")
}

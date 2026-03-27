package handler

import (
	"strconv"
	"time"

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

// List returns the user's tokens
func (h *TokenHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	tokens, err := h.tokenService.ListByUser(userID)
	if err != nil {
		response.InternalError(c, "failed to list tokens")
		return
	}

	response.Success(c, tokens)
}

// Create creates a new API token
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
		response.Fail(c, "TOKEN_CREATE_FAILED", err.Error())
		return
	}

	response.Created(c, token)
}

// Delete deletes a token
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

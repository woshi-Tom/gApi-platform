package handler

import (
	"errors"

	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func verifyTurnstile(c *gin.Context, turnstileService *service.TurnstileService, token string) bool {
	if turnstileService == nil || !turnstileService.Enabled() {
		return true
	}

	if err := turnstileService.Verify(c.Request.Context(), token, c.ClientIP()); err != nil {
		if errors.Is(err, service.ErrTurnstileNotConfigured) {
			response.InternalError(c, "人机验证配置错误，请联系管理员")
			return false
		}
		response.Fail(c, "TURNSTILE_FAILED", "人机验证失败，请刷新后重试")
		return false
	}

	return true
}

package middleware

import (
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

func InitProtected(settingsSvc *service.SettingsService) gin.HandlerFunc {
	return func(c *gin.Context) {
		enabled, err := settingsSvc.GetInitEnabled()
		if err != nil {
			enabled = true
		}

		if enabled {
			c.Next()
			return
		}

		response.Fail(c, "ENDPOINT_DISABLED", "init endpoint is disabled")
		c.Abort()
	}
}

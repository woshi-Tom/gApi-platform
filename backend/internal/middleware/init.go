package middleware

import (
	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// InitProtection creates middleware to protect init endpoints
// Only allows access when system is not initialized
// After initialization, only /init/status is accessible
func InitProtection(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if system is initialized
		var sc model.SystemConfig
		initialized := false
		if err := db.Where("config_key = ?", "system_initialized").First(&sc).Error; err == nil && sc.ConfigValue == "true" {
			initialized = true
		}

		// Get the full path
		path := c.FullPath()

		// If system is initialized, only allow /init/status
		if initialized {
			if path == "/api/v1/init/status" {
				c.Next()
				return
			}
			response.Forbidden(c, "system already initialized")
			c.Abort()
			return
		}

		// If system not initialized, check if trying to create admin
		// Only allow if no admin exists
		if path == "/api/v1/init/create-admin" {
			var count int64
			if err := db.Model(&model.AdminUser{}).Count(&count).Error; err == nil && count > 0 {
				response.Forbidden(c, "admin already exists")
				c.Abort()
				return
			}
		}

		c.Next()
	}
}
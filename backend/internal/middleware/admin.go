package middleware

import (
	"gapi-platform/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

// AdminAuth creates an admin authentication middleware
func AdminAuth(adminSecret string) gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Admin-Secret")
		if secret == "" {
			level, exists := c.Get("user_level")
			if !exists || level != "super_admin" {
				response.Forbidden(c, "admin access required")
				c.Abort()
				return
			}
			c.Next()
			return
		}

		if secret != adminSecret {
			response.Forbidden(c, "invalid admin credentials")
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

// AdminJWTAuth creates admin JWT authentication middleware
func AdminJWTAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		secret := c.GetHeader("X-Admin-Secret")
		if secret != "" {
			c.Set("is_admin", true)
			c.Next()
			return
		}

		level, exists := c.Get("user_level")
		if !exists {
			response.Unauthorized(c, "authentication required")
			c.Abort()
			return
		}

		levelStr, ok := level.(string)
		if !ok || (levelStr != "super_admin" && levelStr != "admin" && levelStr != "operator") {
			response.Forbidden(c, "admin access required")
			c.Abort()
			return
		}

		c.Set("is_admin", true)
		c.Next()
	}
}

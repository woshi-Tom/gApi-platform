package middleware

import (
	"net/http"
	"strings"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// JWTAuth creates a JWT authentication middleware
func JWTAuth(authService *service.AuthService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Unauthorized(c, "missing authorization header")
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Unauthorized(c, "invalid authorization format")
			c.Abort()
			return
		}

		tokenString := parts[1]

		// Validate token
		claims, err := authService.ValidateToken(tokenString)
		if err != nil {
			response.Unauthorized(c, "invalid or expired token")
			c.Abort()
			return
		}

		// Set user info in context
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("user_level", claims.Level)

		c.Next()
	}
}

// TokenAuth creates an API token authentication middleware (for OpenAI-compatible endpoints)
func TokenAuth(tokenService *service.TokenService) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "MISSING_API_KEY",
					Message: "Missing API key in Authorization header",
				},
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "INVALID_API_KEY",
					Message: "Invalid API key format",
				},
			})
			c.Abort()
			return
		}

		tokenKey := parts[1]

		// Validate API token
		token, err := tokenService.Validate(tokenKey)
		if err != nil {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "INVALID_API_KEY",
					Message: "Invalid API key",
				},
			})
			c.Abort()
			return
		}

		if token == nil {
			c.JSON(http.StatusUnauthorized, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "EXPIRED_API_KEY",
					Message: "API key has expired or been disabled",
				},
			})
			c.Abort()
			return
		}

		// Set token and user info in context
		c.Set("token_id", token.ID)
		c.Set("user_id", token.UserID)
		c.Set("tenant_id", token.TenantID)

		c.Next()
	}
}

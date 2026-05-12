package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net"
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
			c.JSON(http.StatusUnauthorized, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "invalid_request_error",
					Code:    "invalid_api_key",
					Message: "Missing API key in Authorization header",
				},
			})
			c.Abort()
			return
		}

		// Check Bearer prefix
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "invalid_request_error",
					Code:    "invalid_api_key",
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
			c.JSON(http.StatusUnauthorized, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "invalid_request_error",
					Code:    "invalid_api_key",
					Message: "Invalid API key",
				},
			})
			c.Abort()
			return
		}

		if token == nil {
			c.JSON(http.StatusUnauthorized, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "invalid_request_error",
					Code:    "expired_api_key",
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

		// T-01: IP whitelist check
		if allowedIPs := token.GetAllowedIPs(); len(allowedIPs) > 0 {
			clientIP := c.ClientIP()
			if !isIPAllowed(clientIP, allowedIPs) {
				c.JSON(http.StatusForbidden, model.APIErrorResponse{
					Error: &model.APIError{
						Type:    "invalid_request_error",
						Code:    "ip_not_allowed",
						Message: fmt.Sprintf("IP %s is not in the token's allowed list", clientIP),
					},
				})
				c.Abort()
				return
			}
		}

		// T-02: Model permission check
		if modelName := extractModelFromRequest(c); modelName != "" {
			// Check denied_models blacklist first
			if deniedModels := token.GetDeniedModels(); len(deniedModels) > 0 {
				for _, denied := range deniedModels {
					if denied == modelName {
						c.JSON(http.StatusForbidden, model.APIErrorResponse{
							Error: &model.APIError{
								Type:    "invalid_request_error",
								Code:    "model_not_allowed",
								Message: fmt.Sprintf("Model '%s' is in the token's denied list", modelName),
							},
						})
						c.Abort()
						return
					}
				}
			}

			// Check allowed_models whitelist
			if allowedModels := token.GetAllowedModels(); len(allowedModels) > 0 {
				found := false
				for _, allowed := range allowedModels {
					if allowed == modelName {
						found = true
						break
					}
				}
				if !found {
					c.JSON(http.StatusForbidden, model.APIErrorResponse{
						Error: &model.APIError{
							Type:    "invalid_request_error",
							Code:    "model_not_allowed",
							Message: fmt.Sprintf("Model '%s' is not in the token's allowed list", modelName),
						},
					})
					c.Abort()
					return
				}
			}
		}

		c.Next()
	}
}

// isIPAllowed checks if the client IP is in the allowed list (supports CIDR notation).
func isIPAllowed(clientIP string, allowedIPs []string) bool {
	ip := net.ParseIP(clientIP)
	if ip == nil {
		return false
	}
	for _, entry := range allowedIPs {
		entry = strings.TrimSpace(entry)
		if entry == "" {
			continue
		}
		if strings.Contains(entry, "/") {
			_, cidr, err := net.ParseCIDR(entry)
			if err == nil && cidr.Contains(ip) {
				return true
			}
		} else if entry == clientIP {
			return true
		}
	}
	return false
}

// extractModelFromRequest reads the model field from the request body for
// /v1/completions and /v1/embeddings endpoints, then resets the body so
// downstream handlers can read it again.
func extractModelFromRequest(c *gin.Context) string {
	path := c.Request.URL.Path
	if !strings.Contains(path, "/completions") && !strings.Contains(path, "/embeddings") {
		return ""
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return ""
	}
	c.Request.Body = io.NopCloser(bytes.NewBuffer(body))

	var partial struct {
		Model string `json:"model"`
	}
	if err := json.Unmarshal(body, &partial); err != nil {
		return ""
	}
	return partial.Model
}

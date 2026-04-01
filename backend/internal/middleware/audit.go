package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"strings"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

// sensitiveFields contains fields that should be masked in audit logs
var sensitiveFields = map[string]bool{
	"password":      true,
	"password_hash": true,
	"api_key":       true,
	"token":         true,
	"credit_card":   true,
	"bank_account":  true,
	"secret":        true,
	"private_key":   true,
}

// skipPaths contains paths that should not be audited
var skipPaths = map[string]bool{
	"/api/v1/internal/health":      true,
	"/health":                      true,
	"/ping":                        true,
	"/api/v1/admin/logs/operation": true, // 避免审计日志本身的记录形成数据膨胀
	"/api/v1/admin/logs/login":     true,
}

// AuditLog creates an audit logging middleware
func AuditLog(auditRepo *repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Read and restore request body
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			requestBody = string(bodyBytes)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
		}

		// Capture response
		writer := &responseWriter{ResponseWriter: c.Writer, body: bytes.NewBufferString("")}
		c.Writer = writer

		// Process request
		c.Next()

		// Record audit log asynchronously
		go func() {
			// Get user info from context
			var userID *uint
			var username string
			if uid, exists := c.Get("user_id"); exists {
				if id, ok := uid.(uint); ok {
					userID = &id
				}
			}
			if uname, exists := c.Get("username"); exists {
				if name, ok := uname.(string); ok {
					username = name
				}
			}

			// Determine action and group
			action, group := determineAction(c.Request.Method, c.Request.URL.Path)

			// Get resource info
			resourceType, resourceID := determineResource(c.Request.URL.Path)

			statusCode := c.Writer.Status()
			success := statusCode < 400

			log := &model.AuditLog{
				UserID:        userID,
				Username:      username,
				Action:        action,
				ActionGroup:   group,
				ResourceType:  resourceType,
				ResourceID:    resourceID,
				RequestMethod: c.Request.Method,
				RequestPath:   c.Request.URL.Path,
				RequestBody:   maskSensitiveData(requestBody),
				RequestIP:     c.ClientIP(),
				StatusCode:    &statusCode,
				ResponseBody:  maskSensitiveData(writer.body.String()),
				Success:       success,
				UserAgent:     c.Request.UserAgent(),
				CreatedAt:     time.Now(),
			}

			auditRepo.Create(log)
		}()
	}
}

// responseWriter captures response body
type responseWriter struct {
	gin.ResponseWriter
	body *bytes.Buffer
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func determineAction(method, path string) (string, string) {
	path = strings.TrimPrefix(path, "/api/v1/")
	path = strings.TrimPrefix(path, "internal/")
	path = strings.TrimPrefix(path, "admin/")

	// Determine group from path
	group := "system"
	action := method + "." + path

	switch {
	case strings.HasPrefix(path, "user"):
		group = "auth"
		if strings.Contains(path, "register") {
			action = "user.register"
		} else if strings.Contains(path, "login") {
			action = "user.login"
		} else if strings.Contains(path, "password") {
			action = "user.password_change"
		}
	case strings.HasPrefix(path, "token"):
		group = "token"
		switch method {
		case "POST":
			action = "token.create"
		case "DELETE":
			action = "token.delete"
		case "PUT":
			action = "token.update"
		}
	case strings.HasPrefix(path, "channel"):
		group = "channel"
		if strings.Contains(path, "test") {
			action = "channel.test"
		} else {
			switch method {
			case "POST":
				action = "channel.create"
			case "PUT":
				action = "channel.update"
			case "DELETE":
				action = "channel.delete"
			}
		}
	case strings.HasPrefix(path, "order"):
		group = "order"
		switch method {
		case "POST":
			action = "order.create"
		}
	case strings.HasPrefix(path, "payment"):
		group = "payment"
		if strings.Contains(path, "callback") {
			action = "payment.callback"
		} else {
			action = "payment.init"
		}
	case strings.HasPrefix(path, "vip"):
		group = "vip"
	}

	return action, group
}

func determineResource(path string) (string, *uint) {
	parts := strings.Split(strings.TrimPrefix(path, "/api/v1/"), "/")

	// Remove internal/admin prefix
	for i, p := range parts {
		if p == "internal" || p == "admin" {
			parts = parts[i+1:]
			break
		}
	}

	if len(parts) == 0 {
		return "", nil
	}

	resourceType := parts[0]

	// Check if there's an ID
	if len(parts) >= 2 {
		// Try to parse ID (simplified - just check if it's numeric)
		idStr := parts[1]
		if len(idStr) > 0 && idStr[0] >= '0' && idStr[0] <= '9' {
			var id uint
			for _, c := range idStr {
				if c >= '0' && c <= '9' {
					id = id*10 + uint(c-'0')
				}
			}
			return resourceType, &id
		}
	}

	return resourceType, nil
}

func maskSensitiveData(data string) string {
	if data == "" {
		return data
	}

	var obj map[string]interface{}
	if err := json.Unmarshal([]byte(data), &obj); err != nil {
		return data
	}

	for key := range obj {
		if sensitiveFields[strings.ToLower(key)] {
			obj[key] = "***"
		}
	}

	result, _ := json.Marshal(obj)
	return string(result)
}

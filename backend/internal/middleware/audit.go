package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"strconv"
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

// importantGetPrefixes contains GET path prefixes that should still be audited
var importantGetPrefixes = []string{
	"/api/v1/payment/", // 支付回调等重要操作
}

// AuditLog creates an audit logging middleware
func AuditLog(auditRepo *repository.AuditRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Skip audit for certain paths
		if skipPaths[c.Request.URL.Path] {
			c.Next()
			return
		}

		// Skip GET requests to reduce data bloat (unless they are important operations)
		if c.Request.Method == "GET" {
			shouldAudit := false
			for _, prefix := range importantGetPrefixes {
				if strings.HasPrefix(c.Request.URL.Path, prefix) {
					shouldAudit = true
					break
				}
			}
			if !shouldAudit {
				c.Next()
				return
			}
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

		startTime := time.Now()

		// Process request
		c.Next()

		// Snapshot response body after request is done
		responseBodyRaw := make([]byte, writer.body.Len())
		copy(responseBodyRaw, writer.body.Bytes())

		// Record audit log asynchronously
		go func() {
			defer func() {
				if r := recover(); r != nil {
					log.Printf("[audit] panic recovered: %v", r)
				}
			}()
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
			responseTimeMs := int(time.Since(startTime).Milliseconds())

			// Determine log type based on request method
			logType := model.LogTypeOperation
			if c.Request.Method == "GET" {
				logType = model.LogTypeAccess
			}

			// Limit body size to prevent data bloat
			maxBodySize := 50000
			responseBody := maskSensitiveData(string(responseBodyRaw))
			if len(responseBody) > maxBodySize {
				responseBody = responseBody[:maxBodySize] + "...[truncated]"
			}
			requestBodyMasked := maskSensitiveData(requestBody)
			if len(requestBodyMasked) > maxBodySize {
				requestBodyMasked = requestBodyMasked[:maxBodySize] + "...[truncated]"
			}

			log := &model.AuditLog{
				UserID:         userID,
				Username:       username,
				Action:         action,
				ActionGroup:    group,
				ResourceType:   resourceType,
				ResourceID:     resourceID,
				RequestMethod:  c.Request.Method,
				RequestPath:    c.Request.URL.Path,
				RequestBody:    requestBodyMasked,
				RequestIP:      c.ClientIP(),
				StatusCode:     &statusCode,
				ResponseBody:   responseBody,
				Success:        success,
				LogType:        logType,
				ResponseTimeMs: responseTimeMs,
				UserAgent:      c.Request.UserAgent(),
				CreatedAt:      time.Now(),
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
		idStr := parts[1]
		if id, err := strconv.ParseUint(idStr, 10, 64); err == nil {
			uid := uint(id)
			return resourceType, &uid
		}
	}

	return resourceType, nil
}

func maskSensitiveData(data string) string {
	if data == "" {
		return data
	}

	var raw interface{}
	if err := json.Unmarshal([]byte(data), &raw); err != nil {
		return data
	}

	masked := maskValue(raw)
	result, _ := json.Marshal(masked)
	return string(result)
}

func maskValue(v interface{}) interface{} {
	switch val := v.(type) {
	case map[string]interface{}:
		for key := range val {
			if sensitiveFields[strings.ToLower(key)] {
				val[key] = "***"
			} else {
				val[key] = maskValue(val[key])
			}
		}
		return val
	case []interface{}:
		for i, item := range val {
			val[i] = maskValue(item)
		}
		return val
	default:
		return v
	}
}

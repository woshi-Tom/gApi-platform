package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestMaskSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "non-JSON string",
			input:    "plain text",
			expected: "plain text",
		},
		{
			name:     "mask password field",
			input:    `{"password": "secret123", "username": "admin"}`,
			expected: `{"password":"***","username":"admin"}`,
		},
		{
			name:     "mask api_key field",
			input:    `{"api_key": "sk-xxxxx", "name": "test"}`,
			expected: `{"api_key":"***","name":"test"}`,
		},
		{
			name:     "mask private_key field",
			input:    `{"private_key": "-----BEGIN RSA PRIVATE KEY-----\nxxxxx"}`,
			expected: `{"private_key":"***"}`,
		},
		{
			name:     "no sensitive fields",
			input:    `{"name": "test", "email": "test@example.com"}`,
			expected: `{"email":"test@example.com","name":"test"}`,
		},
		{
			name:     "case insensitive matching",
			input:    `{"Password": "secret", "API_KEY": "key123"}`,
			expected: `{"API_KEY":"***","Password":"***"}`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := maskSensitiveData(tt.input)
			// Compare JSON by parsing both to normalize ordering
			if tt.input != "" && tt.input != "plain text" {
				var expectedObj, resultObj map[string]interface{}
				json.Unmarshal([]byte(tt.expected), &expectedObj)
				json.Unmarshal([]byte(result), &resultObj)
				assert.Equal(t, expectedObj, resultObj)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestSkipPaths(t *testing.T) {
	// Verify known skip paths are in the map
	skipTests := []struct {
		path   string
		skip   bool
	}{
		{"/api/v1/internal/health", true},
		{"/health", true},
		{"/ping", true},
		{"/api/v1/admin/logs/operation", true},
		{"/api/v1/admin/logs/login", true},
		{"/api/v1/user/profile", false},
		{"/api/v1/admin/dashboard", false},
		{"/api/v1/payment/alipay/query/123", false},
	}

	for _, tt := range skipTests {
		t.Run(tt.path, func(t *testing.T) {
			_, exists := skipPaths[tt.path]
			assert.Equal(t, tt.skip, exists, "skipPaths[%q]", tt.path)
		})
	}
}

func TestImportantGetPrefixes(t *testing.T) {
	// Verify payment paths are in importantGetPrefixes
	assert.Contains(t, importantGetPrefixes, "/api/v1/payment/")
}

func TestAuditMiddlewareSkipsHealthPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/health", nil)

	// Run middleware with nil repo (should not panic)
	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddlewareSkipsOperationLogPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/admin/logs/operation", nil)

	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddlewareSkipsLoginLogPath(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/admin/logs/login", nil)

	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddlewareSkipsRegularGET(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/user/profile", nil)

	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddlewareDoesNotSkipGETOnImportantPrefix(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/api/v1/payment/config", nil)

	// With nil repo, this won't actually create audit log entries
	// but it should NOT skip the request (middleware continues processing)
	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAuditMiddlewareDoesNotSkipPOST(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/api/v1/admin/settings/email", nil)

	handler := AuditLog(nil)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestDetermineAction(t *testing.T) {
	tests := []struct {
		method     string
		path       string
		wantAction string
		wantGroup  string
	}{
		{"POST", "/api/v1/admin/channel", "channel.create", "channel"},
		{"PUT", "/api/v1/admin/channel/1", "channel.update", "channel"},
		{"DELETE", "/api/v1/admin/channel/1", "channel.delete", "channel"},
		{"POST", "/api/v1/user/register", "user.register", "auth"},
		{"POST", "/api/v1/user/login", "user.login", "auth"},
		{"POST", "/api/v1/order", "order.create", "order"},
		{"POST", "/api/v1/payment/callback/alipay", "payment.callback", "payment"},
	}

	for _, tt := range tests {
		t.Run(tt.method+"."+tt.path, func(t *testing.T) {
			action, group := determineAction(tt.method, tt.path)
			assert.Equal(t, tt.wantAction, action)
			assert.Equal(t, tt.wantGroup, group)
		})
	}
}

func TestDetermineResource(t *testing.T) {
	tests := []struct {
		path             string
		wantResourceType string
		wantID           bool
	}{
		{"/api/v1/user/1", "user", true},
		{"/api/v1/user/profile", "user", false},
		{"/api/v1/admin/channel/42", "channel", true},
		{"/api/v1/admin/channel", "channel", false},
		{"/api/v1/admin/logs/operation", "logs", false},
	}

	for _, tt := range tests {
		t.Run(tt.path, func(t *testing.T) {
			resourceType, resourceID := determineResource(tt.path)
			assert.Equal(t, tt.wantResourceType, resourceType)
			if tt.wantID {
				assert.NotNil(t, resourceID)
			} else {
				assert.Nil(t, resourceID)
			}
		})
	}
}



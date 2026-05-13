package middleware

import (
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func TestAdminAuth(t *testing.T) {
	secret := "test-admin-secret"

	tests := []struct {
		name          string
		headerSecret  string
		userLevel     interface{}
		expectedAbort bool
	}{
		{
			name:          "valid secret header",
			headerSecret:  secret,
			userLevel:     nil,
			expectedAbort: false,
		},
		{
			name:          "invalid secret header",
			headerSecret:  "wrong-secret",
			userLevel:     nil,
			expectedAbort: true,
		},
		{
			name:          "missing secret with super_admin",
			headerSecret:  "",
			userLevel:     "super_admin",
			expectedAbort: false,
		},
		{
			name:          "missing secret with regular user",
			headerSecret:  "",
			userLevel:     "free",
			expectedAbort: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			if tt.headerSecret != "" {
				c.Request.Header.Set("X-Admin-Secret", tt.headerSecret)
			}
			if tt.userLevel != nil {
				c.Set("user_level", tt.userLevel)
			}

			middleware := AdminAuth(secret)
			middleware(c)

			if tt.expectedAbort != c.IsAborted() {
				t.Errorf("expected abort=%v, got abort=%v", tt.expectedAbort, c.IsAborted())
			}
		})
	}
}

func TestAdminJWTAuth(t *testing.T) {
	tests := []struct {
		name          string
		headerSecret  string
		userLevel     interface{}
		expectedAbort bool
	}{
		{
			name:          "valid secret header",
			headerSecret:  "any-secret",
			userLevel:     nil,
			expectedAbort: false,
		},
		{
			name:          "super_admin role",
			headerSecret:  "",
			userLevel:     "super_admin",
			expectedAbort: false,
		},
		{
			name:          "admin role",
			headerSecret:  "",
			userLevel:     "admin",
			expectedAbort: false,
		},
		{
			name:          "operator role",
			headerSecret:  "",
			userLevel:     "operator",
			expectedAbort: false,
		},
		{
			name:          "regular user role",
			headerSecret:  "",
			userLevel:     "free",
			expectedAbort: true,
		},
		{
			name:          "no credentials",
			headerSecret:  "",
			userLevel:     nil,
			expectedAbort: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("GET", "/", nil)

			if tt.headerSecret != "" {
				c.Request.Header.Set("X-Admin-Secret", tt.headerSecret)
			}
			if tt.userLevel != nil {
				c.Set("user_level", tt.userLevel)
			}

			middleware := AdminJWTAuth()
			middleware(c)

			if tt.expectedAbort != c.IsAborted() {
				t.Errorf("expected abort=%v, got abort=%v", tt.expectedAbort, c.IsAborted())
			}
		})
	}
}

func TestJWTAuth_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	middleware := JWTAuth(nil)
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestJWTAuth_InvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat")

	middleware := JWTAuth(nil)
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestTokenAuth_MissingHeader(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)

	middleware := TokenAuth(nil)
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestTokenAuth_InvalidFormat(t *testing.T) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Request.Header.Set("Authorization", "InvalidFormat")

	middleware := TokenAuth(nil)
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected request to be aborted")
	}
}

func TestTokenAuth_XApiKeyHeader_MissingBoth(t *testing.T) {
	// No x-api-key and no Authorization header → should abort
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)

	middleware := TokenAuth(nil)
	middleware(c)

	if !c.IsAborted() {
		t.Error("expected request to be aborted when both headers missing")
	}
}

func TestTokenAuth_XApiKeyHeader_Present(t *testing.T) {
	// x-api-key header present → should get past header extraction
	// (will panic at tokenService.Validate since we pass nil, which confirms header was accepted)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/messages", nil)
	c.Request.Header.Set("x-api-key", "sk-ap-test123")

	defer func() {
		r := recover()
		if r != nil {
			// Panic at Validate() means x-api-key was successfully extracted
			// (if header was missing, we'd get a JSON 401 response, not a panic)
			return
		}
		// No panic means tokenService handled it (shouldn't happen with nil)
		if w.Code == 401 && strings.Contains(w.Body.String(), "Missing API key") {
			t.Error("x-api-key header should have been accepted, but got 'Missing API key' error")
		}
	}()

	middleware := TokenAuth(nil)
	middleware(c)
}

func TestTokenAuth_BearerStillWorks(t *testing.T) {
	// Authorization: Bearer should still work (regression test)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("POST", "/v1/chat/completions", nil)
	c.Request.Header.Set("Authorization", "Bearer sk-ap-test123")

	defer func() {
		r := recover()
		if r != nil {
			return // Panic at Validate() means Bearer was successfully extracted
		}
		if w.Code == 401 && strings.Contains(w.Body.String(), "Missing API key") {
			t.Error("Bearer header should have been accepted, but got 'Missing API key' error")
		}
	}()

	middleware := TokenAuth(nil)
	middleware(c)
}

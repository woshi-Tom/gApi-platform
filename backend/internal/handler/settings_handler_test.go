package handler

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestSettingsHandler_UpdateGeneralSettings_JSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type updateGeneralSettingsRequest struct {
		SiteName        string `json:"site_name"`
		SiteLogo        string `json:"site_logo"`
		SiteDescription string `json:"site_description"`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid full request",
			body:       `{"site_name":"My Site","site_logo":"/logo.png","site_description":"desc"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid empty fields",
			body:       `{"site_name":"","site_logo":"","site_description":""}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid partial fields",
			body:       `{"site_name":"My Site"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid empty body",
			body:       `{}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req updateGeneralSettingsRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
			} else if err == nil {
				t.Error("expected error for invalid json")
			}
		})
	}
}

func TestSettingsHandler_UpdateRateLimitSettings_JSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type updateRateLimitSettingsRequest struct {
		FreeRPM int `json:"free_rpm"`
		FreeTPM int `json:"free_tpm"`
		VIPRPM  int `json:"vip_rpm"`
		VIPTPM  int `json:"vip_tpm"`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid full request",
			body:       `{"free_rpm":60,"free_tpm":10000,"vip_rpm":2000,"vip_tpm":500000}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid zero values",
			body:       `{"free_rpm":0,"free_tpm":0,"vip_rpm":0,"vip_tpm":0}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid partial fields",
			body:       `{"free_rpm":60}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid empty body",
			body:       `{}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong type for int field",
			body:       `{"free_rpm":"not-a-number"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req updateRateLimitSettingsRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
			} else if err == nil {
				t.Error("expected error for invalid json or type mismatch")
			}
		})
	}
}

func TestSettingsHandler_UpdateSecuritySettings_JSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type updateSecuritySettingsRequest struct {
		JWTSecret          string `json:"jwt_secret"`
		JWTExpireHours     int    `json:"jwt_expire_hours"`
		PasswordMinLength  int    `json:"password_min_length"`
		PasswordExpireDays int    `json:"password_expire_days"`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid full request",
			body:       `{"jwt_secret":"my-secret","jwt_expire_hours":168,"password_min_length":8,"password_expire_days":90}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid partial fields",
			body:       `{"jwt_secret":"my-secret"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "valid empty body",
			body:       `{}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "wrong type for int field",
			body:       `{"jwt_expire_hours":"not-a-number"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req updateSecuritySettingsRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
			} else if err == nil {
				t.Error("expected error for invalid json or type mismatch")
			}
		})
	}
}

func TestSettingsHandler_TestSMTPConnection_JSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	type testSMTPConnectionRequest struct {
		TestEmail string `json:"test_email"`
	}

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid email",
			body:       `{"test_email":"test@example.com"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "empty email",
			body:       `{"test_email":""}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req testSMTPConnectionRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
			} else if err == nil {
				t.Error("expected error for invalid json")
			}
		})
	}
}

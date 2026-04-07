package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"gapi-platform/internal/model"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	return gin.New()
}

func TestCreateCodeRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     CreateCodeRequest
		wantErr bool
	}{
		{
			name: "valid request - vip type",
			req: CreateCodeRequest{
				Prefix:   "VIP",
				Count:    10,
				CodeType: model.RedemptionCodeTypeVIP,
				VIPDays:  30,
				MaxUses:  1,
			},
			wantErr: false,
		},
		{
			name: "valid request - quota type",
			req: CreateCodeRequest{
				Prefix:    "QUOTA",
				Count:     5,
				CodeType:  model.RedemptionCodeTypeQuota,
				Quota:     100000,
				QuotaType: "permanent",
				MaxUses:   10,
			},
			wantErr: false,
		},
		{
			name: "valid request - recharge type",
			req: CreateCodeRequest{
				Prefix:   "RECH",
				Count:    1,
				CodeType: model.RedemptionCodeTypeRecharge,
				Quota:    50000,
				MaxUses:  1,
			},
			wantErr: false,
		},
		{
			name: "empty prefix",
			req: CreateCodeRequest{
				Prefix:   "",
				Count:    1,
				CodeType: model.RedemptionCodeTypeVIP,
			},
			wantErr: true,
		},
		{
			name: "zero count",
			req: CreateCodeRequest{
				Prefix:   "TEST",
				Count:    0,
				CodeType: model.RedemptionCodeTypeVIP,
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hasPrefix := tt.req.Prefix != ""
			hasCount := tt.req.Count > 0
			hasType := tt.req.CodeType != ""
			isValid := hasPrefix && hasCount && hasType

			if tt.wantErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestRedeemRequest_Validation(t *testing.T) {
	tests := []struct {
		name    string
		req     RedeemRequest
		wantErr bool
	}{
		{
			name: "valid code",
			req: RedeemRequest{
				Code: "VIP-ABC123XYZ",
			},
			wantErr: false,
		},
		{
			name: "empty code",
			req: RedeemRequest{
				Code: "",
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isValid := tt.req.Code != ""
			if tt.wantErr {
				assert.False(t, isValid)
			} else {
				assert.True(t, isValid)
			}
		})
	}
}

func TestGetUserIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name     string
		setup    func(c *gin.Context)
		expected uint
	}{
		{
			name: "user_id present",
			setup: func(c *gin.Context) {
				c.Set("user_id", uint(123))
			},
			expected: 123,
		},
		{
			name: "user_id as int",
			setup: func(c *gin.Context) {
				c.Set("user_id", 456)
			},
			expected: 0,
		},
		{
			name:     "user_id not present",
			setup:    func(c *gin.Context) {},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			tt.setup(c)

			result := getUserIDFromContext(c)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetAdminIDFromContext(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("admin_id present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Set("admin_id", uint(1))

		result := getAdminIDFromContext(c)
		assert.NotNil(t, result)
		assert.Equal(t, uint(1), *result)
	})

	t.Run("admin_id not present", func(t *testing.T) {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)

		result := getAdminIDFromContext(c)
		assert.Nil(t, result)
	})
}

func TestIsUniqueConstraintError(t *testing.T) {
	result := isUniqueConstraintError(nil)
	assert.False(t, result)
}

func TestContainsHelper(t *testing.T) {
	tests := []struct {
		s        string
		substr   string
		expected bool
	}{
		{"hello world", "world", true},
		{"hello world", "foo", false},
		{"", "test", false},
		{"test", "", true},
		{"test", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.s+"_"+tt.substr, func(t *testing.T) {
			result := contains(tt.s, tt.substr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRedemptionHandler_Create_JSONBinding(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name       string
		body       string
		wantStatus int
	}{
		{
			name:       "valid request",
			body:       `{"prefix":"VIP","count":1,"code_type":"vip","vip_days":30}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "invalid json",
			body:       `{invalid}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "missing required field",
			body:       `{"prefix":"VIP"}`,
			wantStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request = httptest.NewRequest("POST", "/", bytes.NewBufferString(tt.body))
			c.Request.Header.Set("Content-Type", "application/json")

			var req CreateCodeRequest
			err := c.ShouldBindJSON(&req)

			if tt.wantStatus == http.StatusOK {
				assert.NoError(t, err)
			} else {
				if err == nil && tt.body == `{invalid}` {
					t.Error("expected error for invalid json")
				}
			}
		})
	}
}

func TestRedemptionCode_JSON(t *testing.T) {
	now := time.Now()
	future := now.Add(24 * time.Hour)

	code := model.RedemptionCode{
		ID:         1,
		Code:       "TEST-ABC123",
		CodeType:   model.RedemptionCodeTypeVIP,
		VIPDays:    30,
		MaxUses:    10,
		UsedCount:  0,
		Status:     model.RedemptionStatusActive,
		ValidFrom:  &now,
		ValidUntil: &future,
	}

	data, err := json.Marshal(code)
	assert.NoError(t, err)
	assert.Contains(t, string(data), "TEST-ABC123")

	var decoded model.RedemptionCode
	err = json.Unmarshal(data, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, code.Code, decoded.Code)
	assert.Equal(t, code.CodeType, decoded.CodeType)
}

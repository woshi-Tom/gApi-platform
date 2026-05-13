package model

import (
	"time"

	"github.com/golang-jwt/jwt/v5"
)

// 本文件中的类型专用于 OpenAI 兼容 API 响应格式（/v1/chat/completions 等）。
// 业务 API（/api/v1/*、/api/v1/admin/*）请使用 pkg/response 包。

// APIResponse represents a standard API response
type APIResponse struct {
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Message string      `json:"message,omitempty"`
	Error   *APIError   `json:"error,omitempty"`
}

// APIError represents an API error (OpenAI-compatible format)
type APIError struct {
	Type    string      `json:"type,omitempty"`
	Code    string      `json:"code"`
	Message string      `json:"message"`
	Param   interface{} `json:"param,omitempty"`
}

// JWTPayload represents JWT token claims
type JWTPayload struct {
	UserID    uint   `json:"user_id"`
	Username  string `json:"username"`
	Level     string `json:"level"`               // free|vip|enterprise
	TokenType string `json:"token_type"`          // access|refresh
	TokenKey  string `json:"token_key,omitempty"` // For API token auth
	Exp       int64  `json:"exp"`
	Iat       int64  `json:"iat"`
}

func (j JWTPayload) GetExpirationTime() (*jwt.NumericDate, error) {
	if j.Exp == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(j.Exp, 0)), nil
}

func (j JWTPayload) GetIssuedAt() (*jwt.NumericDate, error) {
	if j.Iat == 0 {
		return nil, nil
	}
	return jwt.NewNumericDate(time.Unix(j.Iat, 0)), nil
}

func (j JWTPayload) GetNotBefore() (*jwt.NumericDate, error) {
	return nil, nil
}

func (j JWTPayload) GetIssuer() (string, error) {
	return "", nil
}

func (j JWTPayload) GetSubject() (string, error) {
	return "", nil
}

func (j JWTPayload) GetAudience() (jwt.ClaimStrings, error) {
	return nil, nil
}

// LoginResponse represents login response
type LoginResponse struct {
	Token        string    `json:"token"`
	RefreshToken string    `json:"refresh_token,omitempty"`
	ExpiresAt    time.Time `json:"expires_at"`
	User         *User     `json:"user"`
}

// RegisterResponse represents register response
type RegisterResponse struct {
	UserID       uint   `json:"user_id"`
	Username     string `json:"username"`
	Email        string `json:"email"`
	Quota        int64  `json:"quota"`
	QuotaType    string `json:"quota_type"`
	TrialVIPDays int    `json:"trial_vip_days"`
	NeedVerify   bool   `json:"need_verify"`
}

// TokenResponse represents token creation response
type TokenResponse struct {
	ID            uint       `json:"id"`
	Name          string     `json:"name"`
	TokenKey      string     `json:"token_key"`      // Only shown once
	TokenKeyFull  string     `json:"token_key_full"` // Full key, only shown once
	AllowedModels []string   `json:"allowed_models"`
	AllowedIPs    []string   `json:"allowed_ips"`
	ExpiresAt     *time.Time `json:"expires_at"`
	CreatedAt     time.Time  `json:"created_at"`
	Status        string     `json:"status"`
	RemainQuota   int64      `json:"remain_quota"`
	UsedQuota     int64      `json:"used_quota"`
}

// QuotaInfo represents user quota information
type QuotaInfo struct {
	FreeQuota      int64      `json:"free_quota"`
	FreeExpiredAt  *time.Time `json:"free_expired_at"`
	VIPQuota       int64      `json:"vip_quota"`
	VIPExpiredAt   *time.Time `json:"vip_expired_at"`
	IsVIP          bool       `json:"is_vip"`
	Level          string     `json:"level"`
	UsedQuotaToday int64      `json:"used_quota_today"`
	UsedQuotaMonth int64      `json:"used_quota_month"`
}

// ChannelTestRequest represents channel test request
type ChannelTestRequest struct {
	TestType    string        `json:"test_type"` // models|chat|embeddings
	Model       string        `json:"model"`
	Messages    []ChatMessage `json:"messages"`
	Input       string        `json:"input"`
	Temperature float64       `json:"temperature,omitempty"`
	MaxTokens   int           `json:"max_tokens,omitempty"`
}

// ChatMessage represents a chat message
type ChatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

// ChannelTestResponse represents channel test response
type ChannelTestResponse struct {
	Success        bool      `json:"success"`
	ResponseTimeMs int64     `json:"response_time_ms"`
	StatusCode     int       `json:"status_code"`
	Models         []string  `json:"models,omitempty"`
	Content        string    `json:"content,omitempty"`
	Usage          *Usage    `json:"usage,omitempty"`
	Embedding      []float64 `json:"embedding,omitempty"`
	Error          string    `json:"error,omitempty"`
	ErrorType      string    `json:"error_type,omitempty"`
}

// Usage represents token usage
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// OrderCreateRequest represents order creation request
type OrderCreateRequest struct {
	PackageID     uint   `json:"package_id" binding:"required"`
	PaymentMethod string `json:"payment_method" binding:"required"` // alipay
}

// Product represents a product (for frontend compatibility)
type Product struct {
	ID            uint      `json:"id"`
	Name          string    `json:"name"`
	Description   string    `json:"description"`
	ProductType   string    `json:"product_type"` // recharge|vip|package
	Price         float64   `json:"price"`
	OriginalPrice *float64  `json:"original_price"`
	Quota         int64     `json:"quota"`
	BonusQuota    int64     `json:"bonus_quota"`
	VIPDays       int       `json:"vip_days"`
	VIPQuota      int64     `json:"vip_quota"`
	RPMLimit      int       `json:"rpm_limit"`
	TPMLimit      int       `json:"tpm_limit"`
	SortOrder     int       `json:"sort_order"`
	IsRecommended bool      `json:"is_recommended"`
	IsHot         bool      `json:"is_hot"`
	Status        string    `json:"status"` // draft|active|inactive
	CreatedAt     time.Time `json:"created_at"`
}

// DashboardStats represents dashboard statistics
type DashboardStats struct {
	TotalUsers          int64   `json:"total_users"`
	ActiveUsersToday    int64   `json:"active_users_today"`
	TotalChannels       int64   `json:"total_channels"`
	HealthyChannels     int64   `json:"healthy_channels"`
	TotalOrdersToday    int64   `json:"total_orders_today"`
	TotalRevenueToday   float64 `json:"total_revenue_today"`
	TotalQuotaUsedToday int64   `json:"total_quota_used_today"`
	VIPUsersCount       int64   `json:"vip_users_count"`
}

// AuditStats represents audit log statistics
type AuditStats struct {
	TotalCount   int           `json:"total_count"`
	SuccessCount int           `json:"success_count"`
	FailedCount  int           `json:"failed_count"`
	TopActions   []ActionCount `json:"top_actions"`
	Trend        []DateCount   `json:"trend"`
}

// ActionCount represents action count
type ActionCount struct {
	Action string `json:"action"`
	Count  int    `json:"count"`
}

// DateCount represents date count
type DateCount struct {
	Date  string `json:"date"`
	Count int    `json:"count"`
}

// ChatCompletionsRequest represents OpenAI chat completions request
type ChatCompletionsRequest struct {
	Model       string                   `json:"model"` // handler 显式校验，返回 missing_model
	Messages    []map[string]interface{} `json:"messages" binding:"required"`
	Temperature float64                  `json:"temperature"`
	MaxTokens   int                      `json:"max_tokens"`
	TopP        float64                  `json:"top_p"`
	Stream      bool                     `json:"stream"`
	User        string                   `json:"user"`
	Functions   []map[string]interface{} `json:"functions"`
}

// CompletionsRequest represents OpenAI text completions request
type CompletionsRequest struct {
	Model       string  `json:"model"` // handler 显式校验，返回 missing_model
	Prompt      interface{} `json:"prompt" binding:"required"` // string or []string
	MaxTokens   int     `json:"max_tokens"`
	Temperature float64 `json:"temperature"`
	TopP        float64 `json:"top_p"`
	Stream      bool    `json:"stream"`
	User        string  `json:"user"`
}

// EmbeddingsRequest represents OpenAI embeddings request
type EmbeddingsRequest struct {
	Model string      `json:"model"`
	Input interface{} `json:"input" binding:"required"`
}

// APIErrorResponse represents API error response format
type APIErrorResponse struct {
	Error *APIError `json:"error,omitempty"`
}

package model

import (
	"time"
)

// LogType constants for audit logs
const (
	LogTypeOperation = "operation"
	LogTypeAccess    = "access"
	LogTypeSystem    = "system"
)

// AuditLog represents audit logs
type AuditLog struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TenantID *uint  `json:"tenant_id" gorm:"index"`
	UserID   *uint  `json:"user_id" gorm:"index"`
	Username string `json:"username" gorm:"size:100"`

	// Action
	Action       string `json:"action" gorm:"size:100;not null"`
	ActionGroup  string `json:"action_group" gorm:"size:50;not null"`
	ResourceType string `json:"resource_type" gorm:"size:50"`
	ResourceID   *uint  `json:"resource_id" gorm:"index"`

	// Request
	RequestMethod string `json:"request_method" gorm:"size:10"`
	RequestPath   string `json:"request_path" gorm:"size:500"`
	RequestBody   string `json:"request_body" gorm:"type:text"`
	RequestIP     string `json:"request_ip" gorm:"size:50"`
	RequestUA     string `json:"request_ua" gorm:"size:500"`

	// Response
	StatusCode   *int   `json:"status_code"`
	ResponseBody string `json:"response_body" gorm:"type:text"`

	// Result
	Success      bool   `json:"success" gorm:"default:true"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// Changes
	OldValue string `json:"old_value" gorm:"type:text"`
	NewValue string `json:"new_value" gorm:"type:text"`

	// Metadata
	LogType        string `json:"log_type" gorm:"size:20;default:operation;index"`
	ResponseTimeMs int    `json:"response_time_ms" gorm:"default:0"`
	UserAgent      string `json:"user_agent" gorm:"size:500"`
	SessionID      string `json:"session_id" gorm:"size:100"`
	TraceID        string `json:"trace_id" gorm:"size:64"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

// AuditLogBrief represents brief audit log info for list view
type AuditLogBrief struct {
	ID            uint      `json:"id"`
	Action        string    `json:"action"`
	ActionGroup   string    `json:"action_group"`
	ResourceType  string    `json:"resource_type"`
	ResourceID    *uint     `json:"resource_id"`
	Username      string    `json:"username"`
	RequestMethod string    `json:"request_method"`
	RequestPath   string    `json:"request_path"`
	RequestIP     string    `json:"request_ip"`
	Success       bool      `json:"success"`
	ErrorMessage  string    `json:"error_message,omitempty"`
	LogType       string    `json:"log_type"`
	CreatedAt     time.Time `json:"created_at"`
}

func (AuditLog) TableName() string {
	return "audit_logs"
}

// LoginLog represents login attempts
type LoginLog struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	TenantID  *uint  `json:"tenant_id" gorm:"index"`
	UserID    *uint  `json:"user_id" gorm:"index"`
	Username  string `json:"username" gorm:"size:100"`
	LoginType string `json:"login_type" gorm:"size:20;not null"` // user|admin

	// Login info
	IP         string `json:"ip" gorm:"size:50"`
	IPLocation string `json:"ip_location" gorm:"size:200"`
	UserAgent  string `json:"user_agent" gorm:"size:500"`
	DeviceType string `json:"device_type" gorm:"size:50"` // web|mobile|desktop

	// Result
	Success    bool   `json:"success" gorm:"default:false"`
	FailReason string `json:"fail_reason" gorm:"size:100"`

	// Token info
	Token          string     `json:"token" gorm:"size:500"`
	TokenExpiredAt *time.Time `json:"token_expired_at"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func (LoginLog) TableName() string {
	return "login_logs"
}

// UsageLog represents API usage (partitioned by month)
type UsageLog struct {
	ID        uint  `json:"id" gorm:"primaryKey"`
	TenantID  uint  `json:"tenant_id" gorm:"not null;index"`
	UserID    uint  `json:"user_id" gorm:"not null;index"`
	TokenID   *uint `json:"token_id" gorm:"index"`
	ChannelID *uint `json:"channel_id" gorm:"index"`

	// Request
	RequestID string `json:"request_id" gorm:"size:64"` // Request ID for idempotency
	Model     string `json:"model" gorm:"size:100;not null"`

	// Token usage
	PromptTokens     int `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int `json:"completion_tokens" gorm:"default:0"`
	TotalTokens      int `json:"total_tokens" gorm:"default:0"`

	// Cost
	Cost float64 `json:"cost" gorm:"type:decimal(10,4);default:0"`

	// Response
	StatusCode     *int `json:"status_code"`
	ResponseTimeMs int  `json:"response_time_ms" gorm:"default:0"`

	// Error
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"not null;index"`
}

func (UsageLog) TableName() string {
	return "usage_logs"
}

// APIAccessLog represents user API access logs (for user dashboard)
type APIAccessLog struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	TenantID uint `json:"tenant_id" gorm:"index"`

	// Request info
	Endpoint string `json:"endpoint" gorm:"size:100;not null"` // /api/v1/chat/completions
	Method   string `json:"method" gorm:"size:10;not null"`    // POST
	Model    string `json:"model" gorm:"size:50"`              // gpt-3.5-turbo
	TokenID  *uint  `json:"token_id" gorm:"index"`

	// Usage
	PromptTokens     int `json:"prompt_tokens" gorm:"default:0"`
	CompletionTokens int `json:"completion_tokens" gorm:"default:0"`
	TotalTokens      int `json:"total_tokens" gorm:"default:0"`

	// Response
	StatusCode   int    `json:"status_code" gorm:"default:0"`
	ResponseTime int    `json:"response_time" gorm:"default:0"` // milliseconds
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// Request metadata
	RequestIP string `json:"request_ip" gorm:"size:50"`
	UserAgent string `json:"user_agent" gorm:"size:200"`

	// Timestamp
	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func (APIAccessLog) TableName() string {
	return "api_access_logs"
}

// SystemConfig represents system configuration
type SystemConfig struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	TenantID    *uint  `json:"tenant_id" gorm:"index"` // NULL for global config
	ConfigKey   string `json:"config_key" gorm:"size:100;not null"`
	ConfigValue string `json:"config_value" gorm:"type:text"`
	ValueType   string `json:"value_type" gorm:"size:20;default:'string'"`    // string|number|boolean|json
	ConfigGroup string `json:"config_group" gorm:"size:50;default:'general'"` // general|payment|email|sms|oauth
	Description string `json:"description" gorm:"size:200"`
	IsPublic    bool   `json:"is_public" gorm:"default:false"`    // Can be viewed by users
	IsSensitive bool   `json:"is_sensitive" gorm:"default:false"` // Requires permission

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy *uint     `json:"created_by"`
	UpdatedBy *uint     `json:"updated_by"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}

// SignupConfig represents signup bonus configuration
type SignupConfig struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TenantID uint `json:"tenant_id" gorm:"index"`

	// Bonus config
	Enabled     bool   `json:"enabled" gorm:"default:true"`
	QuotaAmount int64  `json:"quota_amount" gorm:"default:100000"`
	QuotaType   string `json:"quota_type" gorm:"size:10;default:'permanent'"` // permanent|vip

	// VIP trial
	TrialVIPDays int   `json:"trial_vip_days" gorm:"default:0"`
	TrialQuota   int64 `json:"trial_quota" gorm:"default:0"`

	// Limits
	PerIPLimit           int  `json:"per_ip_limit" gorm:"default:3"`
	PerEmailVerification bool `json:"per_email_verification" gorm:"default:true"`

	// Validity
	ValidFrom   *time.Time `json:"valid_from"`
	ValidUntil  *time.Time `json:"valid_until"`
	Description string     `json:"description" gorm:"size:200"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedBy *uint     `json:"created_by"`
}

func (SignupConfig) TableName() string {
	return "signup_configs"
}

// PaymentLog records detailed payment operations for debugging and auditing
type PaymentLog struct {
	ID        uint   `json:"id" gorm:"primaryKey"`
	OrderID   uint   `json:"order_id" gorm:"index"`
	OrderNo   string `json:"order_no" gorm:"size:50;index"`
	PaymentID *uint  `json:"payment_id" gorm:"index"`
	UserID    uint   `json:"user_id" gorm:"index"`

	Action       string `json:"action" gorm:"size:32;not null"` // create|notify|query|cancel|refund|expire
	Status       string `json:"status" gorm:"size:16"`
	RequestData  string `json:"request_data" gorm:"type:text"`
	ResponseData string `json:"response_data" gorm:"type:text"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`
	IPAddress    string `json:"ip_address" gorm:"size:50"`

	CreatedAt time.Time `json:"created_at" gorm:"index"`
}

func (PaymentLog) TableName() string {
	return "payment_logs"
}

// IdempotencyKey stores idempotency keys to prevent duplicate operations
type IdempotencyKey struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Key       string    `json:"key" gorm:"size:64;uniqueIndex"`
	UserID    uint      `json:"user_id" gorm:"index"`
	Action    string    `json:"action" gorm:"size:32"`
	OrderID   *uint     `json:"order_id" gorm:"index"`
	OrderNo   string    `json:"order_no" gorm:"size:50"`
	CreatedAt time.Time `json:"created_at"`
	ExpiresAt time.Time `json:"expires_at" gorm:"index"`
}

func (IdempotencyKey) TableName() string {
	return "idempotency_keys"
}

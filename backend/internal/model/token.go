package model

import (
	"encoding/json"
	"time"
)

// Token represents a user's API token
type Token struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TenantID uint `json:"tenant_id" gorm:"index"`
	UserID   uint `json:"user_id" gorm:"not null;index"`

	// Token info
	Name      string `json:"name" gorm:"size:100;not null"`
	TokenKey  string `json:"token_key" gorm:"size:64;not null;uniqueIndex"` // sk-xxx format
	TokenHash string `json:"token_hash" gorm:"size:64;not null"`            // SHA-256 hash
	KeyPrefix string `json:"key_prefix" gorm:"size:10;default:'sk-ap-'"`    // Key prefix

	// Quota
	RemainQuota int64 `json:"remain_quota" gorm:"default:0"`
	IsVIPQuota  bool  `json:"is_vip_quota" gorm:"default:false"`

	// Access control
	AllowedModels string `json:"allowed_models" gorm:"type:jsonb"` // []string, empty means all
	DeniedModels  string `json:"denied_models" gorm:"type:jsonb"`  // []string
	AllowedIPs    string `json:"allowed_ips" gorm:"type:jsonb"`    // []string, empty means no limit

	// Rate limits (override global)
	RPMLimit *int `json:"rpm_limit"`
	TPMLimit *int `json:"tpm_limit"`

	// Usage limits
	MaxUsagePerDay *int64     `json:"max_usage_per_day"`
	ExpiresAt      *time.Time `json:"expires_at"`
	UnlimitedQuota bool      `json:"unlimited_quota" gorm:"default:false"`

	// Status
	Status    string `json:"status" gorm:"size:20;default:'active'"` // active|disabled|expired
	UsedQuota int64  `json:"used_quota" gorm:"default:0"`

	// Timestamps
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
	LastUsedAt *time.Time `json:"last_used_at"`
	LastUsedIP string     `json:"last_used_ip" gorm:"size:50"`
	DeletedAt  *time.Time `json:"deleted_at" gorm:"index"`
}

func (Token) TableName() string {
	return "tokens"
}

// GetAllowedModels returns allowed models
func (t *Token) GetAllowedModels() []string {
	var models []string
	json.Unmarshal([]byte(t.AllowedModels), &models)
	return models
}

// GetDeniedModels returns denied models
func (t *Token) GetDeniedModels() []string {
	var models []string
	json.Unmarshal([]byte(t.DeniedModels), &models)
	return models
}

// GetAllowedIPs returns allowed IPs
func (t *Token) GetAllowedIPs() []string {
	var ips []string
	json.Unmarshal([]byte(t.AllowedIPs), &ips)
	return ips
}

func (t *Token) IsExpired() bool {
	if t.ExpiresAt == nil {
		return false
	}
	return time.Now().After(*t.ExpiresAt)
}

// QuotaTransaction represents quota change records
type QuotaTransaction struct {
	ID       uint  `json:"id" gorm:"primaryKey"`
	TenantID uint  `json:"tenant_id" gorm:"index"`
	UserID   uint  `json:"user_id" gorm:"not null;index"`
	TokenID  *uint `json:"token_id" gorm:"index"`

	// Transaction type
	Type      string `json:"type" gorm:"size:20;not null"`       // recharge|purchase|vip_grant|usage|refund|adjust|expire
	QuotaType string `json:"quota_type" gorm:"size:10;not null"` // permanent|vip

	// Amount
	ChangeAmount  int64 `json:"change_amount" gorm:"not null"` // Positive = increase, Negative = decrease
	BalanceBefore int64 `json:"balance_before" gorm:"not null"`
	BalanceAfter  int64 `json:"balance_after" gorm:"not null"`

	// Reference
	OrderID   *uint  `json:"order_id" gorm:"index"`
	PackageID *uint  `json:"package_id" gorm:"index"`
	ChannelID *uint  `json:"channel_id" gorm:"index"` // Usage source
	Model     string `json:"model" gorm:"size:100"`

	// Description
	Description string `json:"description" gorm:"size:500"`

	// Timestamp
	CreatedAt time.Time `json:"created_at"`
}

func (QuotaTransaction) TableName() string {
	return "quota_transactions"
}

// VIPPackage represents VIP subscription packages
type VIPPackage struct {
	ID            uint     `json:"id" gorm:"primaryKey"`
	TenantID      uint     `json:"tenant_id" gorm:"index"`
	Name          string   `json:"name" gorm:"size:100;not null"`
	Description   string   `json:"description" gorm:"type:text"`
	Price         float64  `json:"price" gorm:"type:decimal(10,2);not null"`
	OriginalPrice *float64 `json:"original_price" gorm:"type:decimal(10,2)"`

	// Duration
	DurationDays int `json:"duration_days" gorm:"default:30"`

	// Quota config
	Quota           int64 `json:"quota" gorm:"default:1000000"`
	RPMLimit        int   `json:"rpm_limit" gorm:"default:2000"`
	TPMLimit        int   `json:"tpm_limit" gorm:"default:100000"`
	ConcurrentLimit int   `json:"concurrent_limit" gorm:"default:10"`

	// Features
	Features string `json:"features" gorm:"type:jsonb"` // JSON object

	// Display
	SortOrder     int  `json:"sort_order" gorm:"default:0"`
	IsRecommended bool `json:"is_recommended" gorm:"default:false"`
	IsPopular     bool `json:"is_popular" gorm:"default:false"`

	// Status
	Status    string `json:"status" gorm:"size:20;default:'active'"` // active|disabled|deleted
	IsVisible bool   `json:"is_visible" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (VIPPackage) TableName() string {
	return "vip_packages"
}

// RechargePackage represents recharge packages for permanent quota
type RechargePackage struct {
	ID            uint     `json:"id" gorm:"primaryKey"`
	TenantID      uint     `json:"tenant_id" gorm:"index"`
	Name          string   `json:"name" gorm:"size:100;not null"`
	Description   string   `json:"description" gorm:"type:text"`
	Price         float64  `json:"price" gorm:"type:decimal(10,2);not null"`
	OriginalPrice *float64 `json:"original_price" gorm:"type:decimal(10,2)"`

	// Quota
	Quota      int64 `json:"quota" gorm:"not null"`        // tokens
	BonusQuota int64 `json:"bonus_quota" gorm:"default:0"` // bonus tokens

	// Display
	SortOrder     int  `json:"sort_order" gorm:"default:0"`
	IsRecommended bool `json:"is_recommended" gorm:"default:false"`
	IsPopular     bool `json:"is_popular" gorm:"default:false"`

	// Status
	Status    string `json:"status" gorm:"size:20;default:'active'"`
	IsVisible bool   `json:"is_visible" gorm:"default:true"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (RechargePackage) TableName() string {
	return "recharge_packages"
}

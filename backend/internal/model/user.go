package model

import (
	"time"
)

// User represents a user account
type User struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	TenantID     uint   `json:"tenant_id" gorm:"index"`
	Username     string `json:"username" gorm:"size:50;not null"`
	Email        string `json:"email" gorm:"size:100;not null;index"`
	Phone        string `json:"phone" gorm:"size:20"`
	PasswordHash string `json:"-" gorm:"size:255;not null"`

	// Verification
	EmailVerified bool       `json:"email_verified" gorm:"default:false"`
	VerifyToken   string     `json:"-" gorm:"size:100"`
	VerifyExpired *time.Time `json:"verify_expired"`

	// Level & VIP
	Level        string     `json:"level" gorm:"size:20;default:'free'"` // free|premium|vip|enterprise
	VIPExpiredAt *time.Time `json:"vip_expired_at" gorm:"column:v_ip_expired_at"`
	VIPPackageID uint       `json:"vip_package_id" gorm:"column:v_ip_package_id"`

	// Quota
	RemainQuota int64 `json:"remain_quota" gorm:"default:0"`                // Permanent quota
	VIPQuota    int64 `json:"vip_quota" gorm:"column:v_ip_quota;default:0"` // VIP quota (30 days)

	// Status
	Status         string     `json:"status" gorm:"size:20;default:'active'"` // active|disabled|suspended
	DisabledReason string     `json:"disabled_reason" gorm:"size:200"`
	LastLoginAt    *time.Time `json:"last_login_at"`
	LastLoginIP    string     `json:"last_login_ip" gorm:"size:50"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`

	// Relations
	Tokens     []Token            `json:"tokens" gorm:"foreignKey:UserID"`
	Orders     []Order            `json:"orders" gorm:"foreignKey:UserID"`
	QuotaTrans []QuotaTransaction `json:"quota_trans" gorm:"foreignKey:UserID"`
}

func (User) TableName() string {
	return "users"
}

// AdminUser represents an administrator
type AdminUser struct {
	ID           uint   `json:"id" gorm:"primaryKey"`
	TenantID     *uint  `json:"tenant_id" gorm:"index"` // NULL for super admin
	Username     string `json:"username" gorm:"size:50;not null;uniqueIndex"`
	PasswordHash string `json:"-" gorm:"size:255;not null"`
	Email        string `json:"email" gorm:"size:100;not null;uniqueIndex"`
	Phone        string `json:"phone" gorm:"size:20"`
	Avatar       string `json:"avatar" gorm:"size:500"`

	// Role & Permissions
	Role        string `json:"role" gorm:"size:20;default:'admin'"` // super_admin|admin|operator|viewer
	Permissions string `json:"permissions" gorm:"type:jsonb"`       // JSON array of permissions

	// Status
	Status      string     `json:"status" gorm:"size:20;default:'active'"`
	LastLoginAt *time.Time `json:"last_login_at"`
	LastLoginIP string     `json:"last_login_ip" gorm:"size:50"`

	// Password Security
	PasswordChangedAt   time.Time  `json:"password_changed_at" gorm:"default:now()"`
	PasswordExpireDays  int        `json:"password_expire_days" gorm:"default:90"`
	FailedLoginAttempts int        `json:"failed_login_attempts" gorm:"default:0"`
	LockedUntil         *time.Time `json:"locked_until"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	CreatedBy *uint      `json:"created_by"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}

// Tenant represents a tenant
type Tenant struct {
	ID          uint   `json:"id" gorm:"primaryKey"`
	Name        string `json:"name" gorm:"size:100;not null"`
	Code        string `json:"code" gorm:"size:50;not null;uniqueIndex"`
	Description string `json:"description" gorm:"type:text"`

	// Quota limits
	MaxUsers    int `json:"max_users" gorm:"default:100"`
	MaxChannels int `json:"max_channels" gorm:"default:50"`
	MaxTokens   int `json:"max_tokens" gorm:"default:100"`

	// Features
	Features string `json:"features" gorm:"type:jsonb"` // JSON object

	// Status
	Status string `json:"status" gorm:"size:20;default:'active'"`

	// Timestamps
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (Tenant) TableName() string {
	return "tenants"
}

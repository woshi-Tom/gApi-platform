package model

import (
	"time"
)

// OrderStatus constants
const (
	OrderStatusPending   = "pending"
	OrderStatusPaid      = "paid"
	OrderStatusCompleted = "completed"
	OrderStatusCancelled = "cancelled"
	OrderStatusExpired   = "expired"
	OrderStatusRefunded  = "refunded"
)

const (
	RedemptionCodeTypeRecharge = "recharge"
	RedemptionCodeTypeVIP      = "vip"
	RedemptionCodeTypeQuota    = "quota"
)

const (
	QuotaTypePermanent = "permanent"
	QuotaTypeVIP       = "vip"
)

const (
	RedemptionStatusActive   = "active"
	RedemptionStatusDisabled = "disabled"
	RedemptionStatusExpired  = "expired"
	RedemptionStatusUsed     = "used"
)

// Order represents an order
type Order struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TenantID uint `json:"tenant_id" gorm:"index"`
	UserID   uint `json:"user_id" gorm:"not null;index"`

	// Order info
	OrderNo   string `json:"order_no" gorm:"size:50;not null;uniqueIndex"`
	OrderType string `json:"order_type" gorm:"size:20;not null"` // recharge|vip|package

	// Package info
	PackageID   *uint  `json:"package_id" gorm:"index"`
	PackageName string `json:"package_name" gorm:"size:100"`

	// Amount
	TotalAmount    float64 `json:"total_amount" gorm:"type:decimal(10,2);not null"`
	DiscountAmount float64 `json:"discount_amount" gorm:"type:decimal(10,2);default:0"`
	PayAmount      float64 `json:"pay_amount" gorm:"type:decimal(10,2);not null"`

	// Status
	Status       string     `json:"status" gorm:"size:20;default:'pending'"` // pending|paid|completed|cancelled|refunded|expired
	PaidAt       *time.Time `json:"paid_at"`
	CompletedAt  *time.Time `json:"completed_at"`
	CancelReason string     `json:"cancel_reason" gorm:"size:200"`
	RefundReason string     `json:"refund_reason" gorm:"type:text"`
	RefundAmount *float64   `json:"refund_amount" gorm:"type:decimal(10,2)"`

	// Expiry - when the order expires if unpaid
	ExpiresAt *time.Time `json:"expires_at" gorm:"index"` // defaults to 4 hours from creation

	// Alipay
	AlipayTradeNo string     `json:"alipay_trade_no" gorm:"size:64"`
	AlipayQRURL   string     `json:"alipay_qr_url" gorm:"type:text"`
	QRExpireAt    *time.Time `json:"qr_expire_at"`

	// Idempotency
	IdempotencyKey string `json:"idempotency_key" gorm:"size:64;index"`

	// Optimistic locking
	Version int `json:"version" gorm:"default:1"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Payment *Payment `json:"payment" gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "orders"
}

// PaymentStatus constants
const (
	PaymentStatusPending  = "pending"
	PaymentStatusSuccess  = "success"
	PaymentStatusFailed   = "failed"
	PaymentStatusRefunded = "refunded"
)

// Payment represents a payment record
type Payment struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TenantID uint `json:"tenant_id" gorm:"index"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	OrderID  uint `json:"order_id" gorm:"not null;index"`

	PaymentNo     string `json:"payment_no" gorm:"size:100;not null;uniqueIndex"`
	PaymentMethod string `json:"payment_method" gorm:"size:20;not null"`

	Amount float64 `json:"amount" gorm:"type:decimal(10,2);not null"`

	Status string     `json:"status" gorm:"size:20;default:'pending'"`
	PaidAt *time.Time `json:"paid_at"`

	ChannelOrderNo string `json:"channel_order_no" gorm:"size:100"`
	ChannelTradeNo string `json:"channel_trade_no" gorm:"size:100"`
	PaymentURL     string `json:"payment_url" gorm:"type:text"`
	QRCode         string `json:"qr_code" gorm:"type:text"`

	CallbackURL  string     `json:"callback_url" gorm:"size:500"`
	CallbackBody string     `json:"callback_body" gorm:"type:text"`
	CallbackAt   *time.Time `json:"callback_at"`

	ErrorCode    string `json:"error_code" gorm:"size:50"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	IdempotencyKey string     `json:"idempotency_key" gorm:"size:64;index"`
	RetryCount     int        `json:"retry_count" gorm:"default:0"`
	LastRetryAt    *time.Time `json:"last_retry_at"`

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (Payment) TableName() string {
	return "payments"
}

// RedemptionCode represents gift card / redemption codes
type RedemptionCode struct {
	ID       uint   `json:"id" gorm:"primaryKey"`
	TenantID uint   `json:"tenant_id" gorm:"index"`
	Code     string `json:"code" gorm:"size:50;not null;uniqueIndex"`
	CodeType string `json:"code_type" gorm:"size:20;not null"` // recharge|vip|quota

	// Rewards
	Quota       int64  `json:"quota" gorm:"default:0"`
	QuotaType   string `json:"quota_type" gorm:"size:10;default:'permanent'"` // permanent|vip
	VIPDays     int    `json:"vip_days" gorm:"default:0"`
	IsPermanent bool   `json:"is_permanent" gorm:"default:false"` // Permanent VIP

	// Usage limits
	MaxUses    int        `json:"max_uses" gorm:"default:1"`
	UsedCount  int        `json:"used_count" gorm:"default:0"`
	ValidFrom  *time.Time `json:"valid_from"`
	ValidUntil *time.Time `json:"valid_until"`

	// Binding
	BoundUserID *uint      `json:"bound_user_id" gorm:"index"`
	BoundAt     *time.Time `json:"bound_at"`

	// Status
	Status string `json:"status" gorm:"size:20;default:'active'"` // active|disabled|expired|used

	// Batch
	BatchID string `json:"batch_id" gorm:"size:50"`

	// Info
	CreatedBy *uint      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt time.Time  `json:"updated_at"`
	UsedAt    *time.Time `json:"used_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (RedemptionCode) TableName() string {
	return "redemption_codes"
}

type RedemptionUsage struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	CodeID       uint      `json:"code_id" gorm:"not null;index"`
	UserID       uint      `json:"user_id" gorm:"not null;index"`
	QuotaGranted int64     `json:"quota_granted" gorm:"default:0"`
	VIPGranted   bool      `json:"vip_granted" gorm:"column:vip_granted;default:false"`
	VIPDays      int       `json:"vip_days" gorm:"column:vip_days;default:0"`
	RedeemedAt   time.Time `json:"redeemed_at" gorm:"not null;default:now()"`
	IPAddress    string    `json:"ip_address" gorm:"size:50"`
	UserAgent    string    `json:"user_agent" gorm:"size:500"`
}

func (RedemptionUsage) TableName() string {
	return "redemption_usage"
}

func (r *RedemptionCode) IsValid() bool {
	now := time.Now()
	if r.Status != RedemptionStatusActive {
		return false
	}
	if r.UsedCount >= r.MaxUses {
		return false
	}
	if r.ValidFrom != nil && now.Before(*r.ValidFrom) {
		return false
	}
	if r.ValidUntil != nil && now.After(*r.ValidUntil) {
		return false
	}
	return true
}

func GenerateCode(prefix string) string {
	timestamp := time.Now().Format("0601021504")
	suffix := randomString(4)
	return prefix + timestamp + suffix
}

func randomString(length int) string {
	const charset = "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	result := make([]byte, length)
	for i := range result {
		result[i] = charset[time.Now().UnixNano()%int64(len(charset))]
	}
	return string(result)
}

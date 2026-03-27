package model

import (
	"time"
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
	Status       string     `json:"status" gorm:"size:20;default:'pending'"` // pending|paid|cancelled|refunded|expired
	PaidAt       *time.Time `json:"paid_at"`
	CancelReason string     `json:"cancel_reason" gorm:"size:200"`
	RefundReason string     `json:"refund_reason" gorm:"type:text"`
	RefundAmount *float64   `json:"refund_amount" gorm:"type:decimal(10,2)"`

	// Expiry
	ExpireAt *time.Time `json:"expire_at"`

	// Timestamps
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`

	// Relations
	Payment *Payment `json:"payment" gorm:"foreignKey:OrderID"`
}

func (Order) TableName() string {
	return "orders"
}

// Payment represents a payment record
type Payment struct {
	ID       uint `json:"id" gorm:"primaryKey"`
	TenantID uint `json:"tenant_id" gorm:"index"`
	UserID   uint `json:"user_id" gorm:"not null;index"`
	OrderID  uint `json:"order_id" gorm:"not null;index"`

	// Payment info
	PaymentNo     string `json:"payment_no" gorm:"size:100;not null;uniqueIndex"`
	PaymentMethod string `json:"payment_method" gorm:"size:20;not null"` // alipay|wechat|bank

	// Amount
	Amount float64 `json:"amount" gorm:"type:decimal(10,2);not null"`

	// Status
	Status string     `json:"status" gorm:"size:20;default:'pending'"` // pending|success|failed|refunded
	PaidAt *time.Time `json:"paid_at"`

	// Channel info
	ChannelOrderNo string `json:"channel_order_no" gorm:"size:100"` // Alipay/WeChat order number
	ChannelTradeNo string `json:"channel_trade_no" gorm:"size:100"` // Third-party trade number
	PaymentURL     string `json:"payment_url" gorm:"type:text"`     // Payment link/qr code
	QRCode         string `json:"qr_code" gorm:"type:text"`         // QR code base64

	// Callback info
	CallbackURL  string     `json:"callback_url" gorm:"size:500"`
	CallbackBody string     `json:"callback_body" gorm:"type:text"`
	CallbackAt   *time.Time `json:"callback_at"`

	// Error info
	ErrorCode    string `json:"error_code" gorm:"size:50"`
	ErrorMessage string `json:"error_message" gorm:"type:text"`

	// Timestamp
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

	// Info
	CreatedBy *uint      `json:"created_by"`
	CreatedAt time.Time  `json:"created_at"`
	UsedAt    *time.Time `json:"used_at"`
	DeletedAt *time.Time `json:"deleted_at" gorm:"index"`
}

func (RedemptionCode) TableName() string {
	return "redemption_codes"
}

package model

import (
	"time"
)

type EmailVerification struct {
	ID           uint       `json:"id" gorm:"primaryKey"`
	Email        string     `json:"email" gorm:"size:100;not null;index"`
	Code         string     `json:"code" gorm:"size:10;not null"`
	IPAddress    string     `json:"ip_address" gorm:"size:50"`
	UserAgent    string     `json:"user_agent" gorm:"size:500"`
	DeviceHash   string     `json:"device_hash" gorm:"size:64;index"`
	CaptchaToken string     `json:"captcha_token" gorm:"size:100"`
	IsUsed       bool       `json:"is_used" gorm:"default:false"`
	UsedAt       *time.Time `json:"used_at"`
	ExpiresAt    time.Time  `json:"expires_at" gorm:"not null;index"`
	CreatedAt    time.Time  `json:"created_at"`
}

func (EmailVerification) TableName() string {
	return "email_verifications"
}

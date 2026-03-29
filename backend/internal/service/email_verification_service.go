package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"math/big"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

var (
	ErrCodeInvalid     = errors.New("验证码无效")
	ErrCodeExpired     = errors.New("验证码已过期")
	ErrCodeUsed        = errors.New("验证码已使用")
	ErrCodeMaxAttempts = errors.New("验证码尝试次数过多")
	ErrRateLimited     = errors.New("请求过于频繁")
)

type EmailVerificationService struct {
	db          *gorm.DB
	redisRepo   *repository.RedisClient
	cfg         *config.EmailConfig
	settingsSvc *SettingsService
	mailer      *DynamicEmailMailer
}

func NewEmailVerificationService(db *gorm.DB, redisRepo *repository.RedisClient, cfg *config.EmailConfig, settingsSvc *SettingsService) *EmailVerificationService {
	return &EmailVerificationService{
		db:          db,
		redisRepo:   redisRepo,
		cfg:         cfg,
		settingsSvc: settingsSvc,
		mailer:      NewDynamicEmailMailer(settingsSvc),
	}
}

func (s *EmailVerificationService) GenerateCode() (string, error) {
	max := big.NewInt(999999)
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%06d", n.Int64()), nil
}

func (s *EmailVerificationService) HashCode(code string) string {
	hash := sha256.Sum256([]byte(code))
	return hex.EncodeToString(hash[:])
}

func (s *EmailVerificationService) GenerateToken() (string, error) {
	bytes := make([]byte, 32)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (s *EmailVerificationService) HashToken(token string) string {
	hash := sha256.Sum256([]byte(token))
	return hex.EncodeToString(hash[:])
}

func (s *EmailVerificationService) CheckSendLimit(email, ip, deviceHash string) error {
	ctx := context.Background()

	emailKey := fmt.Sprintf("email_send:%s", email)
	count, err := s.redisRepo.Client.Incr(ctx, emailKey).Result()
	if err == nil && count == 1 {
		s.redisRepo.Client.Expire(ctx, emailKey, time.Hour)
	}
	if err == nil && count > int64(s.cfg.RateLimit.PerEmailPerHour) {
		return ErrRateLimited
	}

	ipKey := fmt.Sprintf("email_send:ip:%s", ip)
	ipCount, err := s.redisRepo.Client.Incr(ctx, ipKey).Result()
	if err == nil && ipCount == 1 {
		s.redisRepo.Client.Expire(ctx, ipKey, time.Hour)
	}
	if err == nil && ipCount > int64(s.cfg.RateLimit.PerIPPerHour) {
		return ErrRateLimited
	}

	if deviceHash != "" {
		deviceKey := fmt.Sprintf("email_send:device:%s", deviceHash)
		deviceCount, err := s.redisRepo.Client.Incr(ctx, deviceKey).Result()
		if err == nil && deviceCount == 1 {
			s.redisRepo.Client.Expire(ctx, deviceKey, time.Hour)
		}
		if err == nil && deviceCount > int64(s.cfg.RateLimit.PerDevicePerHour) {
			return ErrRateLimited
		}
	}

	return nil
}

func (s *EmailVerificationService) SendVerificationCode(email, ip, userAgent, deviceHash, captchaToken, purpose string) error {
	code, err := s.GenerateCode()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	codeHash := s.HashCode(code)
	expiresAt := time.Now().Add(time.Duration(s.cfg.CodeExpiresMinutes) * time.Minute)

	verification := &model.EmailVerification{
		Email:        email,
		CodeHash:     codeHash,
		Purpose:      purpose,
		IPAddress:    ip,
		UserAgent:    userAgent,
		DeviceHash:   deviceHash,
		CaptchaToken: captchaToken,
		ExpiresAt:    expiresAt,
	}

	if err := s.db.Create(verification).Error; err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	if err := s.mailer.SendVerificationEmail(email, code, purpose); err != nil {
		fmt.Printf("Failed to send email: %v\n", err)
	}

	fmt.Printf("Verification code for %s (purpose: %s): %s\n", email, purpose, code)
	return nil
}

func (s *EmailVerificationService) VerifyCode(email, code, purpose string) (bool, string, error) {
	codeHash := s.HashCode(code)

	var verification model.EmailVerification
	err := s.db.Where("email = ? AND purpose = ? AND is_used = false AND expires_at > ?",
		email, purpose, time.Now()).Order("created_at DESC").First(&verification).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, "", ErrCodeInvalid
		}
		return false, "", err
	}

	if verification.AttemptCount >= s.cfg.MaxAttempts {
		s.db.Model(&verification).Updates(map[string]interface{}{
			"is_used": true,
			"used_at": time.Now(),
		})
		return false, "", ErrCodeMaxAttempts
	}

	if verification.CodeHash != codeHash {
		s.db.Model(&verification).Update("attempt_count", verification.AttemptCount+1)
		return false, "", ErrCodeInvalid
	}

	s.db.Model(&verification).Updates(map[string]interface{}{
		"is_used": true,
		"used_at": time.Now(),
	})

	token := verification.Email + ":" + fmt.Sprintf("%d", time.Now().UnixNano())
	return true, token, nil
}

func (s *EmailVerificationService) SendPasswordResetEmail(email, ip, userAgent, deviceHash, captchaToken string) error {
	token, err := s.GenerateToken()
	if err != nil {
		return fmt.Errorf("failed to generate token: %w", err)
	}

	tokenHash := s.HashToken(token)
	expiresAt := time.Now().Add(time.Duration(s.cfg.PasswordReset.TokenExpiresMinutes) * time.Minute)

	var userID *uint
	var user model.User
	if err := s.db.Where("email = ?", email).First(&user).Error; err == nil {
		userID = &user.ID
	}

	reset := &model.PasswordReset{
		Email:     email,
		TokenHash: tokenHash,
		UserID:    userID,
		IPAddress: ip,
		ExpiresAt: expiresAt,
	}

	if err := s.db.Create(reset).Error; err != nil {
		return fmt.Errorf("failed to create password reset: %w", err)
	}

	frontendURL := "http://localhost:5173"
	resetLink := fmt.Sprintf("%s/reset-password?token=%s", frontendURL, token)

	if err := s.mailer.SendPasswordResetEmail(email, resetLink); err != nil {
		fmt.Printf("Failed to send password reset email: %v\n", err)
	}

	fmt.Printf("Password reset for %s, token: %s\n", email, token)
	return nil
}

func (s *EmailVerificationService) VerifyResetToken(token string) (*model.PasswordReset, error) {
	tokenHash := s.HashToken(token)

	var reset model.PasswordReset
	err := s.db.Where("token_hash = ? AND is_used = false AND expires_at > ?",
		tokenHash, time.Now()).First(&reset).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, errors.New("重置链接无效或已过期")
		}
		return nil, err
	}

	return &reset, nil
}

func (s *EmailVerificationService) ResetPassword(token, newPassword string) error {
	reset, err := s.VerifyResetToken(token)
	if err != nil {
		return err
	}

	hashedPassword, err := bcryptHash(newPassword)
	if err != nil {
		return fmt.Errorf("failed to hash password: %w", err)
	}

	tx := s.db.Begin()

	if err := tx.Model(&model.User{}).Where("id = ?", reset.UserID).Update("password_hash", hashedPassword).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to update password: %w", err)
	}

	reset.IsUsed = true
	now := time.Now()
	reset.UsedAt = &now
	if err := tx.Save(reset).Error; err != nil {
		tx.Rollback()
		return fmt.Errorf("failed to mark token as used: %w", err)
	}

	tx.Commit()
	return nil
}

func bcryptHash(password string) (string, error) {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

package service

import (
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"gorm.io/gorm"
)

type EmailVerificationService struct {
	db        *gorm.DB
	redisRepo *repository.RedisClient
}

func NewEmailVerificationService(db *gorm.DB, redisRepo *repository.RedisClient) *EmailVerificationService {
	return &EmailVerificationService{
		db:        db,
		redisRepo: redisRepo,
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

func (s *EmailVerificationService) SendVerificationEmail(email, ip, userAgent, deviceHash, captchaToken string) error {
	code, err := s.GenerateCode()
	if err != nil {
		return fmt.Errorf("failed to generate code: %w", err)
	}

	expireMinutes := 10
	expiresAt := time.Now().Add(time.Duration(expireMinutes) * time.Minute)

	verification := &model.EmailVerification{
		Email:        email,
		Code:         code,
		IPAddress:    ip,
		UserAgent:    userAgent,
		DeviceHash:   deviceHash,
		CaptchaToken: captchaToken,
		ExpiresAt:    expiresAt,
	}

	if err := s.db.Create(verification).Error; err != nil {
		return fmt.Errorf("failed to create verification: %w", err)
	}

	fmt.Printf("验证码已发送到 %s: %s\n", email, code)
	return nil
}

func (s *EmailVerificationService) VerifyCode(email, code string) (bool, error) {
	var verification model.EmailVerification
	err := s.db.Where("email = ? AND code = ? AND is_used = false AND expires_at > ?",
		email, code, time.Now()).Order("created_at DESC").First(&verification).Error

	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, nil
		}
		return false, err
	}

	now := time.Now()
	verification.IsUsed = true
	verification.UsedAt = &now

	if err := s.db.Save(&verification).Error; err != nil {
		return false, err
	}

	return true, nil
}

func (s *EmailVerificationService) CheckSendLimit(email, ip, deviceHash string) error {
	emailKey := fmt.Sprintf("email_send:%s", email)
	if s.isRateLimited(emailKey, 5, time.Hour) {
		return fmt.Errorf("邮箱验证码发送过于频繁，请稍后再试")
	}

	ipKey := fmt.Sprintf("email_send:ip:%s", ip)
	if s.isRateLimited(ipKey, 10, time.Hour) {
		return fmt.Errorf("IP验证码发送过于频繁，请稍后再试")
	}

	deviceKey := fmt.Sprintf("email_send:device:%s", deviceHash)
	if s.isRateLimited(deviceKey, 3, time.Hour) {
		return fmt.Errorf("设备验证码发送过于频繁，请稍后再试")
	}

	return nil
}

func (s *EmailVerificationService) isRateLimited(key string, limit int, window time.Duration) bool {
	count, err := s.redisRepo.Client.Incr(s.redisRepo.Client.Context(), key).Result()
	if err != nil {
		return false
	}

	if count == 1 {
		s.redisRepo.Client.Expire(s.redisRepo.Client.Context(), key, window)
	}

	return count > int64(limit)
}

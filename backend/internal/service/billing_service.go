package service

import (
	"errors"
	"math"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
)

var (
	ErrQuotaInsufficient = errors.New("insufficient quota")
	ErrTokenDisabled     = errors.New("token is disabled")
	ErrTokenExpired      = errors.New("token has expired")
)

type BillingService struct {
	userRepo   *repository.UserRepository
	tokenRepo  *repository.TokenRepository
	usageRepo  *repository.UsageLogRepository
	txRepo     *repository.QuotaTransactionRepository
}

func NewBillingService(
	userRepo *repository.UserRepository,
	tokenRepo *repository.TokenRepository,
	usageRepo *repository.UsageLogRepository,
	txRepo *repository.QuotaTransactionRepository,
) *BillingService {
	return &BillingService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		usageRepo: usageRepo,
		txRepo:    txRepo,
	}
}

type UsageRecord struct {
	UserID           uint
	TokenID          uint
	ChannelID        uint
	Model            string
	PromptTokens     int
	CompletionTokens int
	TotalTokens      int
	Cost             float64
}

func (s *BillingService) PreConsumeQuota(userID, tokenID uint, modelName string, estimatedTokens int) error {
	token, err := s.tokenRepo.GetByID(tokenID)
	if err != nil {
		return err
	}

	if token.Status != "active" {
		return ErrTokenDisabled
	}

	if token.IsExpired() {
		return ErrTokenExpired
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	estimatedQuota := s.CalculateQuota(modelName, estimatedTokens, estimatedTokens)

	if !token.IsVIPQuota && user.RemainQuota < int64(estimatedQuota) {
		return ErrQuotaInsufficient
	}

	if token.IsVIPQuota && user.VIPQuota < int64(estimatedQuota) {
		return ErrQuotaInsufficient
	}

	if !token.UnlimitedQuota && token.RemainQuota < int64(estimatedQuota) {
		return ErrQuotaInsufficient
	}

	return nil
}

func (s *BillingService) PostConsumeQuota(userID, tokenID uint, modelName string, promptTokens, completionTokens int) error {
	token, err := s.tokenRepo.GetByID(tokenID)
	if err != nil {
		return err
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	actualQuota := s.CalculateQuota(modelName, promptTokens, completionTokens)
	_ = s.CalculateCost(modelName, promptTokens, completionTokens)

	quotaType := "permanent"
	if user.Level == "vip" || user.Level == "enterprise" {
		quotaType = "vip"
	}

	err = s.txRepo.Create(&model.QuotaTransaction{
		UserID:        userID,
		TokenID:       &tokenID,
		Type:          "usage",
		QuotaType:     quotaType,
		ChangeAmount: -int64(actualQuota),
		BalanceBefore: user.RemainQuota,
		BalanceAfter:  user.RemainQuota - int64(actualQuota),
		Description:   "API usage: " + modelName,
		Model:         modelName,
	})
	if err != nil {
		return err
	}

	user.RemainQuota -= int64(actualQuota)
	if token.IsVIPQuota {
		user.VIPQuota -= int64(actualQuota)
	}
	token.UsedQuota += int64(actualQuota)
	if !token.UnlimitedQuota {
		token.RemainQuota -= int64(actualQuota)
	}

	if err := s.userRepo.Update(user); err != nil {
		return err
	}

	if err := s.tokenRepo.Update(token); err != nil {
		return err
	}

	return nil
}

func (s *BillingService) CalculateQuota(modelName string, promptTokens, completionTokens int) int {
	rate := s.GetModelRate(modelName)
	baseTokens := promptTokens + completionTokens

	multiplier := 1.0
	if containsSubstring(modelName, "gpt-4") {
		multiplier = 10.0
	} else if containsSubstring(modelName, "gpt-3.5-turbo") {
		multiplier = 1.0
	} else if containsSubstring(modelName, "claude-3-opus") {
		multiplier = 15.0
	} else if containsSubstring(modelName, "claude-3-sonnet") {
		multiplier = 3.0
	} else if containsSubstring(modelName, "gemini") {
		multiplier = 1.0
	}

	return int(math.Ceil(float64(baseTokens) * rate * multiplier))
}

func (s *BillingService) CalculateCost(modelName string, promptTokens, completionTokens int) float64 {
	inputCost := float64(promptTokens) / 1000.0 * 0.001
	outputCost := float64(completionTokens) / 1000.0 * 0.002

	return inputCost + outputCost
}

func (s *BillingService) GetModelRate(modelName string) float64 {
	rates := map[string]float64{
		"gpt-3.5-turbo":      1.0,
		"gpt-4":               20.0,
		"gpt-4-turbo":         10.0,
		"gpt-4-32k":           30.0,
		"claude-3-opus":       15.0,
		"claude-3-sonnet":      3.0,
		"claude-3-haiku":       1.0,
		"gemini-pro":           1.0,
		"gemini-1.5-pro":       2.0,
		"deepseek-chat":        1.0,
		"deepseek-coder":       1.0,
	}

	for pattern, rate := range rates {
		if containsSubstring(modelName, pattern) {
			return rate
		}
	}

	return 1.0
}

func containsSubstring(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsSubstringHelper(s, substr))
}

func containsSubstringHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (s *BillingService) AddQuota(userID uint, amount int64, quotaType, description string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	tx := &model.QuotaTransaction{
		UserID:        userID,
		Type:          "recharge",
		QuotaType:     quotaType,
		ChangeAmount:  amount,
		BalanceBefore: user.RemainQuota,
	}

	if quotaType == "vip" {
		user.VIPQuota += amount
		tx.BalanceAfter = user.VIPQuota
	} else {
		user.RemainQuota += amount
		tx.BalanceAfter = user.RemainQuota
	}

	tx.Description = description

	if err := s.txRepo.Create(tx); err != nil {
		return err
	}

	return s.userRepo.Update(user)
}

func (s *BillingService) DeductQuota(userID uint, amount int64, reason string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	if user.RemainQuota < amount {
		return ErrQuotaInsufficient
	}

	tx := &model.QuotaTransaction{
		UserID:        userID,
		Type:          "adjust",
		QuotaType:     "permanent",
		ChangeAmount:  -amount,
		BalanceBefore: user.RemainQuota,
		BalanceAfter:  user.RemainQuota - amount,
		Description:   reason,
	}

	user.RemainQuota -= amount

	if err := s.txRepo.Create(tx); err != nil {
		return err
	}

	return s.userRepo.Update(user)
}

func (s *BillingService) GetUserQuota(userID uint) (*QuotaInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	return &QuotaInfo{
		RemainQuota:    user.RemainQuota,
		VIPQuota:       user.VIPQuota,
		VIPExpiredAt:   user.VIPExpiredAt,
		IsVIP:          user.Level == "vip" || user.Level == "enterprise",
		Level:          user.Level,
		UsedQuotaToday: 0,
		UsedQuotaMonth: 0,
	}, nil
}

type QuotaInfo struct {
	RemainQuota    int64
	VIPQuota       int64
	VIPExpiredAt   *time.Time
	IsVIP          bool
	Level          string
	UsedQuotaToday int64
	UsedQuotaMonth int64
}

func (s *BillingService) LogUsage(record *UsageRecord) error {
	usageLog := &model.UsageLog{
		UserID:           record.UserID,
		TokenID:          &record.TokenID,
		ChannelID:        &record.ChannelID,
		Model:            record.Model,
		PromptTokens:     record.PromptTokens,
		CompletionTokens: record.CompletionTokens,
		TotalTokens:      record.TotalTokens,
		Cost:             record.Cost,
	}

	return s.usageRepo.Create(usageLog)
}

func (s *BillingService) GetUserUsageStats(userID uint, days int) (*repository.UsageStats, error) {
	stats, err := s.usageRepo.GetUserStats(userID, days)
	if err != nil {
		return nil, err
	}
	return stats, nil
}

type TokenQuotaInfo struct {
	RemainQuota    int64
	UsedQuota      int64
	TotalQuota     int64
	UsagePercent   float64
	ResetAt        *interface{}
	IsUnlimited    bool
}

func (s *BillingService) GetTokenQuota(tokenID uint) (*TokenQuotaInfo, error) {
	token, err := s.tokenRepo.GetByID(tokenID)
	if err != nil {
		return nil, err
	}

	totalQuota := token.RemainQuota + token.UsedQuota
	usagePercent := 0.0
	if totalQuota > 0 {
		usagePercent = float64(token.UsedQuota) / float64(totalQuota) * 100
	}

	return &TokenQuotaInfo{
		RemainQuota:  token.RemainQuota,
		UsedQuota:    token.UsedQuota,
		TotalQuota:   totalQuota,
		UsagePercent: usagePercent,
		ResetAt:     nil,
		IsUnlimited:  token.UnlimitedQuota,
	}, nil
}

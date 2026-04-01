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
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	usageRepo *repository.UsageLogRepository
	txRepo    *repository.QuotaTransactionRepository
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

	estimatedQuota := int64(s.CalculateQuota(modelName, estimatedTokens, estimatedTokens))

	totalAvailable := s.GetTotalAvailableQuota(user)

	if totalAvailable >= estimatedQuota {
		return nil
	}

	return ErrQuotaInsufficient
}

func (s *BillingService) GetTotalAvailableQuota(user *model.User) int64 {
	var total int64

	if user.FreeQuota > 0 && (user.FreeExpiredAt == nil || user.FreeExpiredAt.After(time.Now())) {
		total += user.FreeQuota
	}

	if s.userRepo != nil {
		rechargeQuota := s.getActiveRechargeQuota(user.ID)
		total += rechargeQuota
	}

	if user.VIPQuota > 0 && (user.VIPExpiredAt == nil || user.VIPExpiredAt.After(time.Now())) {
		total += user.VIPQuota
	}

	return total
}

func (s *BillingService) getActiveRechargeQuota(userID uint) int64 {
	return 0
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

	remaining := int64(actualQuota)

	if user.FreeQuota > 0 && (user.FreeExpiredAt == nil || user.FreeExpiredAt.After(time.Now())) {
		consume := min(remaining, user.FreeQuota)
		user.FreeQuota -= consume
		remaining -= consume

		err = s.txRepo.Create(&model.QuotaTransaction{
			UserID:        userID,
			TokenID:       &tokenID,
			Type:          "usage",
			QuotaType:     "free",
			ChangeAmount:  -consume,
			BalanceBefore: user.FreeQuota + consume,
			BalanceAfter:  user.FreeQuota,
			Description:   "API usage (free): " + modelName,
			Model:         modelName,
		})
		if err != nil {
			return err
		}
	}

	if remaining > 0 {
		hasRecharge := s.consumeRechargeQuota(userID, tokenID, &remaining, modelName)
		if hasRecharge && remaining > 0 {
			user.VIPQuota -= remaining
			if user.VIPQuota < 0 {
				user.VIPQuota = 0
			}

			err = s.txRepo.Create(&model.QuotaTransaction{
				UserID:        userID,
				TokenID:       &tokenID,
				Type:          "usage",
				QuotaType:     "vip",
				ChangeAmount:  -remaining,
				BalanceBefore: user.VIPQuota + remaining,
				BalanceAfter:  user.VIPQuota,
				Description:   "API usage (VIP): " + modelName,
				Model:         modelName,
			})
			if err != nil {
				return err
			}
		}
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

func (s *BillingService) consumeRechargeQuota(userID, tokenID uint, remaining *int64, modelName string) bool {
	return false
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
		"gpt-3.5-turbo":   1.0,
		"gpt-4":           20.0,
		"gpt-4-turbo":     10.0,
		"gpt-4-32k":       30.0,
		"claude-3-opus":   15.0,
		"claude-3-sonnet": 3.0,
		"claude-3-haiku":  1.0,
		"gemini-pro":      1.0,
		"gemini-1.5-pro":  2.0,
		"deepseek-chat":   1.0,
		"deepseek-coder":  1.0,
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

	var balanceBefore, balanceAfter int64

	switch quotaType {
	case "free":
		balanceBefore = user.FreeQuota
		user.FreeQuota += amount
		balanceAfter = user.FreeQuota
	case "vip":
		balanceBefore = user.VIPQuota
		user.VIPQuota += amount
		balanceAfter = user.VIPQuota
	default:
		return errors.New("invalid quota type")
	}

	tx := &model.QuotaTransaction{
		UserID:        userID,
		Type:          "recharge",
		QuotaType:     quotaType,
		ChangeAmount:  amount,
		BalanceBefore: balanceBefore,
		BalanceAfter:  balanceAfter,
		Description:   description,
	}

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

	totalQuota := s.GetTotalAvailableQuota(user)
	if totalQuota < amount {
		return ErrQuotaInsufficient
	}

	tx := &model.QuotaTransaction{
		UserID:        userID,
		Type:          "adjust",
		QuotaType:     "mixed",
		ChangeAmount:  -amount,
		BalanceBefore: totalQuota,
		BalanceAfter:  totalQuota - amount,
		Description:   reason,
	}

	if err := s.txRepo.Create(tx); err != nil {
		return err
	}

	return nil
}

func (s *BillingService) GetUserQuota(userID uint) (*QuotaInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	isVIP := s.isVIPUser(user)

	return &QuotaInfo{
		FreeQuota:      user.FreeQuota,
		FreeExpiredAt:  user.FreeExpiredAt,
		VIPQuota:       user.VIPQuota,
		VIPExpiredAt:   user.VIPExpiredAt,
		IsVIP:          isVIP,
		Level:          user.Level,
		UsedQuotaToday: 0,
		UsedQuotaMonth: 0,
	}, nil
}

func (s *BillingService) isVIPUser(user *model.User) bool {
	if user == nil {
		return false
	}
	hasLevel := user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"
	if !hasLevel {
		return false
	}
	if user.VIPExpiredAt == nil {
		return false
	}
	return user.VIPExpiredAt.After(time.Now())
}

type QuotaInfo struct {
	FreeQuota      int64
	FreeExpiredAt  *time.Time
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
	RemainQuota  int64
	UsedQuota    int64
	TotalQuota   int64
	UsagePercent float64
	ResetAt      *interface{}
	IsUnlimited  bool
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
		ResetAt:      nil,
		IsUnlimited:  token.UnlimitedQuota,
	}, nil
}

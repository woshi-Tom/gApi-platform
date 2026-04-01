package service

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
)

var ErrTokenLimitExceeded = errors.New("token limit exceeded")

// TokenService handles token operations
type TokenService struct {
	tokenRepo      *repository.TokenRepository
	userRepo       *repository.UserRepository
	vipPackageRepo *repository.VIPPackageRepository
}

// NewTokenService creates a new token service
func NewTokenService(tokenRepo *repository.TokenRepository) *TokenService {
	return &TokenService{tokenRepo: tokenRepo}
}

// SetUserRepo sets the user repository for token limit checking
func (s *TokenService) SetUserRepo(userRepo *repository.UserRepository, vipPackageRepo *repository.VIPPackageRepository) {
	s.userRepo = userRepo
	s.vipPackageRepo = vipPackageRepo
}

// Create creates a new API token
func (s *TokenService) Create(userID uint, name string, allowedModels, allowedIPs []string, expiresAt *time.Time, rpmLimit, tpmLimit *int) (*model.TokenResponse, error) {
	// Check token limit based on user tier
	if err := s.checkTokenLimit(userID); err != nil {
		return nil, err
	}

	// Generate token key
	keyBytes := make([]byte, 32)
	rand.Read(keyBytes)
	tokenKey := "sk-ap-" + hex.EncodeToString(keyBytes)[:32]

	// Generate hash
	hashBytes := make([]byte, 32)
	rand.Read(hashBytes)
	tokenHash := hex.EncodeToString(hashBytes)

	token := &model.Token{
		UserID:      userID,
		Name:        name,
		TokenKey:    tokenKey,
		TokenHash:   tokenHash,
		KeyPrefix:   "sk-ap-",
		RemainQuota: 0,
		Status:      "active",
		ExpiresAt:   expiresAt,
		RPMLimit:    rpmLimit,
		TPMLimit:    tpmLimit,
	}

	if len(allowedModels) > 0 {
		b, _ := json.Marshal(allowedModels)
		token.AllowedModels = string(b)
	} else {
		token.AllowedModels = "[]"
	}
	if len(allowedIPs) > 0 {
		b, _ := json.Marshal(allowedIPs)
		token.AllowedIPs = string(b)
	} else {
		token.AllowedIPs = "[]"
	}
	token.DeniedModels = "[]"

	if err := s.tokenRepo.Create(token); err != nil {
		return nil, err
	}

	return &model.TokenResponse{
		ID:            token.ID,
		Name:          token.Name,
		TokenKey:      tokenKey,
		TokenKeyFull:  tokenKey,
		AllowedModels: allowedModels,
		AllowedIPs:    allowedIPs,
		ExpiresAt:     token.ExpiresAt,
		CreatedAt:     token.CreatedAt,
		Status:        token.Status,
		RemainQuota:   token.RemainQuota,
		UsedQuota:     token.UsedQuota,
	}, nil
}

// ListByUser lists tokens for a user
func (s *TokenService) ListByUser(userID uint) ([]model.Token, error) {
	return s.tokenRepo.ListByUser(userID)
}

// Delete deletes a token
func (s *TokenService) Delete(id uint) error {
	return s.tokenRepo.Delete(id)
}

// GetByID gets a token by ID
func (s *TokenService) GetByID(id uint) (*model.Token, error) {
	return s.tokenRepo.GetByID(id)
}

// Validate validates a token
func (s *TokenService) Validate(tokenKey string) (*model.Token, error) {
	token, err := s.tokenRepo.GetByKey(tokenKey)
	if err != nil {
		return nil, err
	}

	if token.Status != "active" {
		return nil, nil
	}

	if token.ExpiresAt != nil && token.ExpiresAt.Before(time.Now()) {
		token.Status = "expired"
		s.tokenRepo.Update(token)
		return nil, nil
	}

	return token, nil
}

// GetUserQuota returns user's total available quota (permanent + VIP)
func (s *TokenService) GetUserQuota(userID uint) (int64, error) {
	if s.userRepo == nil {
		return 0, nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return 0, err
	}

	return user.FreeQuota + user.VIPQuota, nil
}

func joinStrings(slice []string, sep string) string {
	if len(slice) == 0 {
		return ""
	}
	result := slice[0]
	for i := 1; i < len(slice); i++ {
		result += sep + slice[i]
	}
	return result
}

func (s *TokenService) checkTokenLimit(userID uint) error {
	if s.userRepo == nil || s.vipPackageRepo == nil {
		return nil
	}

	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil
	}

	tokenCount, err := s.tokenRepo.CountByUser(userID)
	if err != nil {
		return err
	}

	var maxTokens int
	isVIP := user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"

	if isVIP {
		if user.VIPPackageID > 0 {
			pkg, err := s.vipPackageRepo.GetByID(user.VIPPackageID)
			if err == nil && pkg != nil && pkg.ConcurrentLimit > 0 {
				maxTokens = pkg.ConcurrentLimit
			} else {
				maxTokens = 3
			}
		} else {
			maxTokens = 3
		}
	} else if user.Level == "enterprise" {
		maxTokens = 5
	} else {
		maxTokens = 1
	}

	if int(tokenCount) >= maxTokens {
		return ErrTokenLimitExceeded
	}
	return nil
}

package service

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AuthService handles authentication
type AuthService struct {
	userRepo  *repository.UserRepository
	tokenRepo *repository.TokenRepository
	jwtCfg    *config.JWTConfig
}

// NewAuthService creates a new auth service
func NewAuthService(userRepo *repository.UserRepository, tokenRepo *repository.TokenRepository, jwtCfg *config.JWTConfig) *AuthService {
	return &AuthService{
		userRepo:  userRepo,
		tokenRepo: tokenRepo,
		jwtCfg:    jwtCfg,
	}
}

// Register errors
var (
	ErrRegistrationClosed    = errors.New("REGISTRATION_CLOSED")
	ErrPasswordTooWeak        = errors.New("PASSWORD_TOO_WEAK")
	ErrEmailDomainNotAllowed = errors.New("EMAIL_DOMAIN_NOT_ALLOWED")
	ErrIPRegistrationLimit   = errors.New("IP_REGISTRATION_LIMIT_EXCEEDED")
)

// Register creates a new user
func (s *AuthService) Register(c *gin.Context, username, email, password string) (*model.RegisterResponse, error) {
	settingsService := NewSettingsService(s.userRepo.GetDB())
	registerSettings, err := settingsService.GetRegisterSettings()
	if err != nil {
		return nil, fmt.Errorf("failed to get register settings: %w", err)
	}

	if !registerSettings.AllowRegister {
		return nil, ErrRegistrationClosed
	}

	if len(password) < registerSettings.MinPasswordLength {
		return nil, ErrPasswordTooWeak
	}

	if registerSettings.AllowedDomains != "" {
		allowedDomains := strings.Split(registerSettings.AllowedDomains, ",")
		userDomain := email[strings.Index(email, "@"):]
		domainAllowed := false
		for _, d := range allowedDomains {
			if strings.TrimSpace(d) == userDomain {
				domainAllowed = true
				break
			}
		}
		if !domainAllowed {
			return nil, ErrEmailDomainNotAllowed
		}
	}

	clientIP := c.ClientIP()
	if clientIP != "" {
		ipRegCount, err := s.userRepo.CountByIPLast24Hours(clientIP)
		if err != nil {
			return nil, fmt.Errorf("failed to check IP: %w", err)
		}
		if ipRegCount >= int64(registerSettings.MaxAccountsPerIP) {
			return nil, ErrIPRegistrationLimit
		}
	}

	// Check if email already exists
	existing, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, fmt.Errorf("failed to check email: %w", err)
	}
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Check if username already exists
	existing, err = s.userRepo.GetByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("failed to check username: %w", err)
	}
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user with signup reward
	user := &model.User{
		Username:          username,
		Email:             email,
		PasswordHash:     string(hash),
		Status:            "active",
		IPRegisteredFrom: clientIP,
	}

	rewardType := registerSettings.SignupRewardType
	rewardAmount := registerSettings.SignupRewardAmount

	switch rewardType {
	case "quota":
		user.Level = "free"
		user.FreeQuota = rewardAmount
		if rewardAmount > 0 {
			t := time.Now().AddDate(0, 0, 7)
			user.FreeExpiredAt = &t
		}
	case "vip":
		user.Level = "vip_bronze"
		user.VIPQuota = rewardAmount
		t := time.Now().AddDate(0, 1, 0)
		user.VIPExpiredAt = &t
	default:
		user.Level = "free"
		user.FreeQuota = 0
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	var quota int64
	var quotaType string
	if user.Level == "vip_bronze" {
		quota = user.VIPQuota
		quotaType = "vip"
	} else {
		quota = user.FreeQuota
		quotaType = "free"
	}

	return &model.RegisterResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Quota:        quota,
		QuotaType:    quotaType,
		TrialVIPDays: 0,
		NeedVerify:   false,
	}, nil
}

// Login authenticates a user
func (s *AuthService) Login(email, password string) (*model.LoginResponse, error) {
	// Find user
	user, err := s.userRepo.GetByEmail(email)
	if err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Verify password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password)); err != nil {
		return nil, errors.New("invalid credentials")
	}

	// Check status
	if user.Status != "active" {
		return nil, errors.New("account is disabled")
	}

	// Generate JWT
	token, expiresAt, err := s.generateToken(user)
	if err != nil {
		return nil, err
	}

	// Update last login
	user.LastLoginAt = &time.Time{}
	now := time.Now()
	user.LastLoginAt = &now
	s.userRepo.Update(user)

	return &model.LoginResponse{
		Token:     token,
		ExpiresAt: expiresAt,
		User:      user,
	}, nil
}

// ChangePassword changes user password
func (s *AuthService) ChangePassword(userID uint, oldPassword, newPassword string) error {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return err
	}

	// Verify old password
	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(oldPassword)); err != nil {
		return errors.New("invalid old password")
	}

	// Hash new password
	hash, err := bcrypt.GenerateFromPassword([]byte(newPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user.PasswordHash = string(hash)
	return s.userRepo.Update(user)
}

func (s *AuthService) generateToken(user *model.User) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtCfg.Expiry())

	claims := &model.JWTPayload{
		UserID:   user.ID,
		Username: user.Username,
		Level:    user.Level,
		Exp:      expiresAt.Unix(),
		Iat:      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// GenerateAdminToken generates a JWT for admin login
func (s *AuthService) GenerateAdminToken(username, role string) (string, time.Time, error) {
	expiresAt := time.Now().Add(s.jwtCfg.Expiry())

	claims := &model.JWTPayload{
		Username: username,
		Level:    role, // Use Level field to store admin role
		Exp:      expiresAt.Unix(),
		Iat:      time.Now().Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.jwtCfg.Secret))
	if err != nil {
		return "", time.Time{}, err
	}

	return tokenString, expiresAt, nil
}

// ValidateToken validates a JWT token
func (s *AuthService) ValidateToken(tokenString string) (*model.JWTPayload, error) {
	claims := &model.JWTPayload{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(s.jwtCfg.Secret), nil
	})
	if err != nil {
		return nil, err
	}

	if !token.Valid {
		return nil, errors.New("invalid token")
	}

	return claims, nil
}

// UserService handles user operations
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService creates a new user service
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

// GetByID gets a user by ID
func (s *UserService) GetByID(id uint) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

func (s *UserService) GetByUsername(username string) (*model.User, error) {
	return s.userRepo.GetByUsername(username)
}

// Update updates a user
func (s *UserService) Update(user *model.User) error {
	return s.userRepo.Update(user)
}

// GetQuota gets user quota info
func (s *UserService) GetQuota(userID uint) (*model.QuotaInfo, error) {
	user, err := s.userRepo.GetByID(userID)
	if err != nil {
		return nil, err
	}

	isVIP := user.Level == "enterprise" ||
		user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"
	if !isVIP || user.VIPExpiredAt == nil || user.VIPExpiredAt.Before(time.Now()) {
		isVIP = false
	}

	return &model.QuotaInfo{
		FreeQuota:     user.FreeQuota,
		FreeExpiredAt: user.FreeExpiredAt,
		VIPQuota:      user.VIPQuota,
		VIPExpiredAt:  user.VIPExpiredAt,
		IsVIP:         isVIP,
		Level:         user.Level,
	}, nil
}

func (s *UserService) GetDB() *gorm.DB {
	return s.userRepo.GetDB()
}

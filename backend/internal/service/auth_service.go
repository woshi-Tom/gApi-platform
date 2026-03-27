package service

import (
	"errors"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
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

// Register creates a new user
func (s *AuthService) Register(username, email, password string) (*model.RegisterResponse, error) {
	// Check if email already exists
	existing, _ := s.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, errors.New("email already registered")
	}

	// Check if username already exists
	existing, _ = s.userRepo.GetByUsername(username)
	if existing != nil {
		return nil, errors.New("username already taken")
	}

	// Hash password
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}

	// Create user
	user := &model.User{
		Username:     username,
		Email:        email,
		PasswordHash: string(hash),
		Level:        "free",
		Status:       "active",
		RemainQuota:  100000, // Default signup bonus
	}

	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	return &model.RegisterResponse{
		UserID:       user.ID,
		Username:     user.Username,
		Email:        user.Email,
		Quota:        user.RemainQuota,
		QuotaType:    "permanent",
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

	isVIP := user.Level == "vip" || user.Level == "enterprise"
	if user.VIPExpiredAt != nil && user.VIPExpiredAt.Before(time.Now()) {
		isVIP = false
	}

	return &model.QuotaInfo{
		RemainQuota:  user.RemainQuota,
		VIPQuota:     user.VIPQuota,
		VIPExpiredAt: user.VIPExpiredAt,
		IsVIP:        isVIP,
		Level:        user.Level,
	}, nil
}

func (s *UserService) GetDB() *gorm.DB {
	return s.userRepo.GetDB()
}

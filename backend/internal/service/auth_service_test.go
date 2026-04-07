package service

import (
	"testing"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"

	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
)

func TestJWTPayload(t *testing.T) {
	payload := model.JWTPayload{
		UserID:   123,
		Username: "testuser",
		Level:    "free",
		Exp:      time.Now().Add(24 * time.Hour).Unix(),
		Iat:      time.Now().Unix(),
	}

	assert.Equal(t, uint(123), payload.UserID)
	assert.Equal(t, "testuser", payload.Username)
	assert.Equal(t, "free", payload.Level)
}

func TestJWTPayload_GetExpirationTime(t *testing.T) {
	t.Run("with expiration", func(t *testing.T) {
		exp := time.Now().Add(time.Hour).Unix()
		payload := model.JWTPayload{Exp: exp}

		result, err := payload.GetExpirationTime()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, exp, result.Unix())
	})

	t.Run("without expiration", func(t *testing.T) {
		payload := model.JWTPayload{Exp: 0}

		result, err := payload.GetExpirationTime()
		assert.NoError(t, err)
		assert.Nil(t, result)
	})
}

func TestJWTPayload_GetIssuedAt(t *testing.T) {
	t.Run("with issued time", func(t *testing.T) {
		iat := time.Now().Unix()
		payload := model.JWTPayload{Iat: iat}

		result, err := payload.GetIssuedAt()
		assert.NoError(t, err)
		assert.NotNil(t, result)
		assert.Equal(t, iat, result.Unix())
	})
}

func TestJWTPayload_GetNotBefore(t *testing.T) {
	payload := model.JWTPayload{}

	result, err := payload.GetNotBefore()
	assert.NoError(t, err)
	assert.Nil(t, result)
}

func TestGenerateToken(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: 24,
	}

	service := &AuthService{jwtCfg: jwtCfg}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Level:    "free",
	}

	tokenString, expiresAt, err := service.generateToken(user)

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.False(t, expiresAt.IsZero())

	assert.True(t, expiresAt.After(time.Now()))
	assert.True(t, expiresAt.Before(time.Now().Add(25*time.Hour)))
}

func TestGenerateAdminToken(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: 24,
	}

	service := &AuthService{jwtCfg: jwtCfg}

	tokenString, expiresAt, err := service.GenerateAdminToken("admin", "super_admin")

	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
	assert.False(t, expiresAt.IsZero())

	_, err = jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(jwtCfg.Secret), nil
	})
	assert.NoError(t, err)
}

func TestValidateToken(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: 24,
	}

	service := &AuthService{jwtCfg: jwtCfg}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Level:    "free",
	}

	tokenString, _, err := service.generateToken(user)
	assert.NoError(t, err)

	claims, err := service.ValidateToken(tokenString)
	assert.NoError(t, err)
	assert.NotNil(t, claims)
	assert.Equal(t, user.ID, claims.UserID)
	assert.Equal(t, user.Username, claims.Username)
	assert.Equal(t, user.Level, claims.Level)
}

func TestValidateToken_Invalid(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: 24,
	}

	service := &AuthService{jwtCfg: jwtCfg}

	_, err := service.ValidateToken("invalid-token")
	assert.Error(t, err)
}

func TestValidateToken_WrongSecret(t *testing.T) {
	jwtCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: 24,
	}

	service := &AuthService{jwtCfg: jwtCfg}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Level:    "free",
	}

	tokenString, _, err := service.generateToken(user)
	assert.NoError(t, err)

	wrongSecretService := &AuthService{
		jwtCfg: &config.JWTConfig{
			Secret:     "different-secret-key-that-is-32-chars",
			ExpireHour: 24,
		},
	}

	_, err = wrongSecretService.ValidateToken(tokenString)
	assert.Error(t, err)
}

func TestValidateToken_Expired(t *testing.T) {
	expiredCfg := &config.JWTConfig{
		Secret:     "test-secret-key-that-is-at-least-32-chars",
		ExpireHour: -1,
	}

	expiredService := &AuthService{jwtCfg: expiredCfg}

	user := &model.User{
		ID:       1,
		Username: "testuser",
		Level:    "free",
	}

	tokenString, _, err := expiredService.generateToken(user)
	assert.NoError(t, err)
	assert.NotEmpty(t, tokenString)
}

func TestQuotaInfo(t *testing.T) {
	now := time.Now()
	info := model.QuotaInfo{
		FreeQuota:     50000,
		FreeExpiredAt: &now,
		VIPQuota:      100000,
		VIPExpiredAt:  &now,
		IsVIP:         true,
		Level:         "vip_gold",
	}

	assert.Equal(t, int64(50000), info.FreeQuota)
	assert.Equal(t, int64(100000), info.VIPQuota)
	assert.True(t, info.IsVIP)
	assert.Equal(t, "vip_gold", info.Level)
}

func TestLoginResponse(t *testing.T) {
	now := time.Now()
	user := &model.User{
		ID:    1,
		Email: "test@example.com",
	}

	resp := model.LoginResponse{
		Token:     "test-token",
		ExpiresAt: now,
		User:      user,
	}

	assert.Equal(t, "test-token", resp.Token)
	assert.Equal(t, now, resp.ExpiresAt)
	assert.Equal(t, user, resp.User)
}

func TestRegisterResponse(t *testing.T) {
	resp := model.RegisterResponse{
		UserID:       1,
		Username:     "testuser",
		Email:        "test@example.com",
		Quota:        50000,
		QuotaType:    "free",
		TrialVIPDays: 0,
		NeedVerify:   false,
	}

	assert.Equal(t, uint(1), resp.UserID)
	assert.Equal(t, "testuser", resp.Username)
	assert.Equal(t, int64(50000), resp.Quota)
}

package middleware

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupRateLimitDB(t *testing.T) (*gorm.DB, *service.TokenService) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	if err := db.AutoMigrate(&model.Token{}, &model.User{}); err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}
	tokenRepo := repository.NewTokenRepository(db)
	tokenSvc := service.NewTokenService(tokenRepo)
	return db, tokenSvc
}

func createTestToken(t *testing.T, db *gorm.DB, id uint, rpm, tpm int) {
	t.Helper()
	token := &model.Token{
		ID:       id,
		UserID:   1,
		Name:     "test-token",
		TokenKey: fmt.Sprintf("sk-ap-test-%d", id),
		Status:   "active",
	}
	if rpm > 0 {
		token.RPMLimit = &rpm
	}
	if tpm > 0 {
		token.TPMLimit = &tpm
	}
	if err := db.Create(token).Error; err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
}

func TestTokenRateLimit_RPM(t *testing.T) {
	db, tokenSvc := setupRateLimitDB(t)

	createTestToken(t, db, 101, 2, 0)

	middleware := TokenRateLimit(tokenSvc)

	// First 2 requests should pass (burst=2)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("token_id", uint(101))

		middleware(c)

		if c.IsAborted() {
			t.Errorf("request %d should pass (within RPM limit), got status %d", i+1, w.Code)
		}
	}

	// 3rd request should be rate limited
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("token_id", uint(101))

	middleware(c)

	if !c.IsAborted() {
		t.Error("request 3 should be rate limited (exceeded RPM)")
	}
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestTokenRateLimit_TPM(t *testing.T) {
	db, tokenSvc := setupRateLimitDB(t)

	// Create token with RPM=1000 (won't hit RPM) and TPM=2
	createTestToken(t, db, 201, 1000, 2)

	middleware := TokenRateLimit(tokenSvc)

	// First 2 requests should pass
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("token_id", uint(201))

		middleware(c)

		if c.IsAborted() {
			t.Errorf("request %d should pass (within TPM limit), got status %d", i+1, w.Code)
		}
	}

	// 3rd request should hit TPM limit
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("token_id", uint(201))

	middleware(c)

	if !c.IsAborted() {
		t.Error("request 3 should be rate limited (exceeded TPM)")
	}
	if w.Code != http.StatusTooManyRequests {
		t.Errorf("expected status 429, got %d", w.Code)
	}
}

func TestTokenRateLimit_DefaultRPM(t *testing.T) {
	db, tokenSvc := setupRateLimitDB(t)

	// Create token with NO custom RPM → should use default (10 RPM)
	createTestToken(t, db, 301, 0, 0)

	middleware := TokenRateLimit(tokenSvc)

	// Fire 11 requests rapidly — first 10 should pass, 11th blocked
	passed := 0
	for i := 0; i < 11; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("token_id", uint(301))

		middleware(c)
		if !c.IsAborted() {
			passed++
		}
	}

	if passed != 10 {
		t.Errorf("expected 10 requests to pass with default RPM, got %d", passed)
	}
}

func TestTokenRateLimit_NoTokenID(t *testing.T) {
	_, tokenSvc := setupRateLimitDB(t)

	middleware := TokenRateLimit(tokenSvc)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	// Don't set token_id — should pass through

	middleware(c)

	if c.IsAborted() {
		t.Error("request without token_id should pass through (not rate limited)")
	}
}

// T-P2-02b: TokenRateLimit — 多 token 隔离（token A 耗尽不影响 token B）
func TestTokenRateLimit_MultipleTokens(t *testing.T) {
	db, tokenSvc := setupRateLimitDB(t)

	createTestToken(t, db, 401, 2, 0) // token A: RPM=2
	createTestToken(t, db, 402, 2, 0) // token B: RPM=2

	middleware := TokenRateLimit(tokenSvc)

	// Exhaust token A's budget (2 requests)
	for i := 0; i < 2; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("token_id", uint(401))
		middleware(c)
		if c.IsAborted() {
			t.Errorf("token A request %d should pass", i+1)
		}
	}

	// Token A's 3rd request should be blocked
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("token_id", uint(401))
	middleware(c)
	if !c.IsAborted() {
		t.Error("token A request 3 should be rate limited")
	}

	// Token B should still pass (independent limiter)
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	c.Request = httptest.NewRequest("GET", "/", nil)
	c.Set("token_id", uint(402))
	middleware(c)
	if c.IsAborted() {
		t.Error("token B should not be affected by token A's rate limit")
	}
}

// T-P2-02c: TokenRateLimit — nil tokenService 使用默认 RPM=10
func TestTokenRateLimit_NilTokenService(t *testing.T) {
	middleware := TokenRateLimit(nil)

	passed := 0
	for i := 0; i < 11; i++ {
		w := httptest.NewRecorder()
		c, _ := gin.CreateTestContext(w)
		c.Request = httptest.NewRequest("GET", "/", nil)
		c.Set("token_id", uint(501))

		middleware(c)
		if !c.IsAborted() {
			passed++
		}
	}

	if passed != 10 {
		t.Errorf("expected 10 requests to pass with nil tokenService (default RPM=10), got %d", passed)
	}
}

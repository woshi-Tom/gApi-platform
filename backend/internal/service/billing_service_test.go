package service

import (
	"testing"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupBillingTest(t *testing.T) (*gorm.DB, *BillingService, *repository.UserRepository, *repository.TokenRepository) {
	t.Helper()
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}
	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.QuotaTransaction{},
		&model.UserRechargeRecord{},
		&model.UsageLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	userRepo := repository.NewUserRepository(db)
	tokenRepo := repository.NewTokenRepository(db)
	usageRepo := repository.NewUsageLogRepository(db)
	txRepo := repository.NewQuotaTransactionRepository(db)
	rechargeRepo := repository.NewUserRechargeRecordRepository(db)

	billing := NewBillingService(userRepo, tokenRepo, usageRepo, txRepo, rechargeRepo)
	return db, billing, userRepo, tokenRepo
}

func createBillingUser(t *testing.T, db *gorm.DB, id uint, freeQuota, vipQuota int64) {
	t.Helper()
	future := time.Now().Add(24 * time.Hour)
	user := &model.User{
		ID:           id,
		Username:     "testuser",
		Email:        "test@test.com",
		PasswordHash: "hash",
		FreeQuota:    freeQuota,
		FreeExpiredAt: &future,
		VIPQuota:     vipQuota,
		VIPExpiredAt: &future,
	}
	if err := db.Create(user).Error; err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
}

func createBillingToken(t *testing.T, db *gorm.DB, id, userID uint, remainQuota int64) {
	t.Helper()
	token := &model.Token{
		ID:          id,
		UserID:      userID,
		Name:        "test-token",
		TokenKey:    "sk-ap-test",
		Status:      "active",
		RemainQuota: remainQuota,
	}
	if err := db.Create(token).Error; err != nil {
		t.Fatalf("failed to create token: %v", err)
	}
}

func createRechargeRecord(t *testing.T, db *gorm.DB, userID, orderID uint, quota, remaining int64, expiredAt time.Time) {
	t.Helper()
	rec := &model.UserRechargeRecord{
		UserID:    userID,
		PackageID: 1,
		OrderID:   orderID,
		Quota:     quota,
		Remaining: remaining,
		ExpiredAt: expiredAt,
		Status:    "active",
	}
	if err := db.Create(rec).Error; err != nil {
		t.Fatalf("failed to create recharge record: %v", err)
	}
}

func TestPostConsumeQuota_FreeQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 1, 10000, 0)  // user with 10000 free quota
	createBillingToken(t, db, 1, 1, 50000) // token with 50000 remain

	err := billing.PostConsumeQuota(1, 1, "gpt-3.5-turbo", 100, 100)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify user free_quota decreased
	var user model.User
	db.First(&user, 1)
	if user.FreeQuota >= 10000 {
		t.Errorf("expected free_quota to decrease from 10000, got %d", user.FreeQuota)
	}

	// Verify token used_quota increased
	var token model.Token
	db.First(&token, 1)
	if token.UsedQuota == 0 {
		t.Error("expected token used_quota > 0")
	}

	// Verify quota_transaction was created
	var txCount int64
	db.Model(&model.QuotaTransaction{}).Where("user_id = ?", 1).Count(&txCount)
	if txCount == 0 {
		t.Error("expected at least 1 quota transaction")
	}
}

func TestPostConsumeQuota_Insufficient(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 2, 10, 0)  // user with very low quota
	createBillingToken(t, db, 2, 2, 10) // token with low remain

	// Try to consume more than available (gpt-4 has 10x multiplier)
	err := billing.PostConsumeQuota(2, 2, "gpt-4", 100, 100)
	// The function doesn't check sufficiency — it just deducts what's available
	// Verify that user free_quota is NOT negative
	var user model.User
	db.First(&user, 2)
	if user.FreeQuota < 0 {
		t.Errorf("free_quota should not be negative, got %d", user.FreeQuota)
	}
	_ = err
}

func TestPostConsumeQuota_TransactionAtomic(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 3, 5000, 0)
	createBillingToken(t, db, 3, 3, 50000)

	// Consume quota
	err := billing.PostConsumeQuota(3, 3, "gpt-3.5-turbo", 50, 50)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify: user free_quota + token used_quota + quota_transactions are all consistent
	var user model.User
	db.First(&user, 3)
	var token model.Token
	db.First(&token, 3)

	var totalTxChange int64
	db.Model(&model.QuotaTransaction{}).
		Where("user_id = ? AND type = 'usage'", 3).
		Select("COALESCE(SUM(ABS(change_amount)), 0)").
		Scan(&totalTxChange)

	// The token's used_quota should match the total transaction changes
	if token.UsedQuota != totalTxChange {
		t.Errorf("token used_quota (%d) != total transaction changes (%d) — atomicity broken",
			token.UsedQuota, totalTxChange)
	}
}

func TestConsumeRechargeQuota_FIFO(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	future := time.Now().Add(48 * time.Hour)
	farFuture := time.Now().Add(96 * time.Hour)

	createBillingUser(t, db, 4, 0, 0) // no free quota
	createBillingToken(t, db, 4, 4, 100000)

	// Create 2 recharge records: one expires sooner (FIFO first)
	createRechargeRecord(t, db, 4, 101, 500, 500, future)     // expires first
	createRechargeRecord(t, db, 4, 102, 1000, 1000, farFuture) // expires later

	// Consume 600 tokens — should take 500 from first record, 100 from second
	err := billing.PostConsumeQuota(4, 4, "gpt-3.5-turbo", 300, 300)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify: first record should be depleted, second partially consumed
	var records []model.UserRechargeRecord
	db.Where("user_id = ?", 4).Order("expired_at ASC").Find(&records)

	if len(records) != 2 {
		t.Fatalf("expected 2 recharge records, got %d", len(records))
	}

	if records[0].Remaining != 0 {
		t.Errorf("first record should be fully consumed (remaining=0), got %d", records[0].Remaining)
	}
	if records[0].Status != "used" {
		t.Errorf("first record status should be 'used', got '%s'", records[0].Status)
	}
	if records[1].Remaining >= 1000 {
		t.Errorf("second record should be partially consumed, remaining=%d", records[1].Remaining)
	}
}

func TestConsumeRechargeQuota_Exhaust(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	future := time.Now().Add(48 * time.Hour)

	createBillingUser(t, db, 5, 0, 0)
	createBillingToken(t, db, 5, 5, 100000)

	// Single recharge record with 100 quota
	createRechargeRecord(t, db, 5, 201, 100, 100, future)

	// Consume exactly 100 tokens — should exhaust the record
	err := billing.PostConsumeQuota(5, 5, "gpt-3.5-turbo", 50, 50)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	var rec model.UserRechargeRecord
	db.First(&rec, "order_id = ?", 201)

	if rec.Remaining != 0 {
		t.Errorf("expected remaining=0, got %d", rec.Remaining)
	}
	if rec.Status != "used" {
		t.Errorf("expected status='used', got '%s'", rec.Status)
	}
}

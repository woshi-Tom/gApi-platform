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
		ID:            id,
		Username:      "testuser",
		Email:         "test@test.com",
		PasswordHash:  "hash",
		FreeQuota:     freeQuota,
		FreeExpiredAt: &future,
		VIPQuota:      vipQuota,
		VIPExpiredAt:  &future,
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

func createRechargeRecord(t *testing.T, db *gorm.DB, userID, orderID uint, quota, remaining int64, expiredAt time.Time) uint {
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
	return rec.ID
}

// T-P2-03: PostConsumeQuota — free_quota 事务扣减
func TestPostConsumeQuota_FreeQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 1, 10000, 0)
	createBillingToken(t, db, 1, 1, 50000)

	err := billing.PostConsumeQuota(1, 1, "gpt-3.5-turbo", 100, 100)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify: quota transaction created
	var txCount int64
	db.Model(&model.QuotaTransaction{}).Where("user_id = ?", 1).Count(&txCount)
	if txCount == 0 {
		t.Error("expected at least 1 quota transaction")
	}

	// Verify: token used_quota updated
	var token model.Token
	db.First(&token, 1)
	if token.UsedQuota == 0 {
		t.Error("expected token used_quota > 0")
	}
}

// T-P2-04: PostConsumeQuota — 配额不为负
func TestPostConsumeQuota_QuotaNotNegative(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 2, 10, 0)
	createBillingToken(t, db, 2, 2, 10)

	// Consume more than available
	billing.PostConsumeQuota(2, 2, "gpt-4", 100, 100)

	var user model.User
	db.First(&user, 2)
	if user.FreeQuota < 0 {
		t.Errorf("free_quota should not be negative, got %d", user.FreeQuota)
	}
}

// T-P2-03b: PostConsumeQuota — token used_quota 原子更新
func TestPostConsumeQuota_TokenUsedQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 6, 100000, 0)
	createBillingToken(t, db, 6, 6, 100000)

	err := billing.PostConsumeQuota(6, 6, "gpt-3.5-turbo", 50, 50)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	var token model.Token
	db.First(&token, 6)
	if token.UsedQuota == 0 {
		t.Error("expected token.UsedQuota > 0")
	}
}

// --- 仓储层 FIFO 测试（直接验证 DB 行为） ---

// T-P2-05: 仓储层 — GetActiveByUser 按 expired_at ASC 排序
func TestRechargeRepo_FIFOOrder(t *testing.T) {
	db, _, _, _ := setupBillingTest(t)
	repo := repository.NewUserRechargeRecordRepository(db)

	future := time.Now().Add(48 * time.Hour)
	farFuture := time.Now().Add(96 * time.Hour)

	// Insert in reverse order (far future first, then sooner)
	createRechargeRecord(t, db, 10, 301, 1000, 1000, farFuture)
	createRechargeRecord(t, db, 10, 302, 500, 500, future)

	records, err := repo.GetActiveByUser(10)
	if err != nil {
		t.Fatalf("GetActiveByUser failed: %v", err)
	}
	if len(records) != 2 {
		t.Fatalf("expected 2 records, got %d", len(records))
	}

	// First record should expire sooner (FIFO)
	if records[0].OrderID != 302 {
		t.Errorf("expected first record (order_id=302, expires sooner), got order_id=%d", records[0].OrderID)
	}
	if records[1].OrderID != 301 {
		t.Errorf("expected second record (order_id=301, expires later), got order_id=%d", records[1].OrderID)
	}
}

// T-P2-06: 仓储层 — UpdateRemaining 耗尽标记 used
func TestRechargeRepo_UpdateToUsed(t *testing.T) {
	db, _, _, _ := setupBillingTest(t)
	repo := repository.NewUserRechargeRecordRepository(db)

	future := time.Now().Add(48 * time.Hour)
	id := createRechargeRecord(t, db, 11, 401, 100, 100, future)

	// Exhaust the record
	err := repo.UpdateRemaining(id, 0, "used")
	if err != nil {
		t.Fatalf("UpdateRemaining failed: %v", err)
	}

	// Verify: record is now "used" and excluded from GetActiveByUser
	records, _ := repo.GetActiveByUser(11)
	for _, r := range records {
		if r.ID == id {
			t.Error("exhausted record should not appear in GetActiveByUser")
		}
	}

	// Verify: record directly has status=used
	var rec model.UserRechargeRecord
	db.First(&rec, id)
	if rec.Status != "used" {
		t.Errorf("expected status='used', got '%s'", rec.Status)
	}
	if rec.Remaining != 0 {
		t.Errorf("expected remaining=0, got %d", rec.Remaining)
	}
}

// T-P2-05b: 仓储层 — 多条记录扣减后 GetActiveByUser 返回正确剩余
func TestRechargeRepo_PartialConsumption(t *testing.T) {
	db, _, _, _ := setupBillingTest(t)
	repo := repository.NewUserRechargeRecordRepository(db)

	future := time.Now().Add(48 * time.Hour)

	id1 := createRechargeRecord(t, db, 12, 501, 500, 500, future)
	createRechargeRecord(t, db, 12, 502, 1000, 1000, future)

	// Exhaust first record
	repo.UpdateRemaining(id1, 0, "used")

	// GetActiveByUser should return only the second record
	records, _ := repo.GetActiveByUser(12)
	if len(records) != 1 {
		t.Fatalf("expected 1 active record, got %d", len(records))
	}
	if records[0].OrderID != 502 {
		t.Errorf("expected order_id=502, got %d", records[0].OrderID)
	}
}

// T-P2-05c: 仓储层 — GetTotalActiveQuota 汇总正确
func TestRechargeRepo_TotalActiveQuota(t *testing.T) {
	db, _, _, _ := setupBillingTest(t)
	repo := repository.NewUserRechargeRecordRepository(db)

	future := time.Now().Add(48 * time.Hour)

	createRechargeRecord(t, db, 13, 601, 500, 500, future)
	createRechargeRecord(t, db, 13, 602, 1000, 1000, future)

	total := repo.GetTotalActiveQuota(13)
	if total != 1500 {
		t.Errorf("expected total=1500, got %d", total)
	}

	// Partially consume first record
	repo.UpdateRemaining(
		func() uint {
			var r model.UserRechargeRecord
			db.Where("order_id = ?", 601).First(&r)
			return r.ID
		}(),
		200, "active",
	)

	total = repo.GetTotalActiveQuota(13)
	if total != 1200 {
		t.Errorf("expected total=1200 after partial consumption, got %d", total)
	}
}

// T-P2-07: PostConsumeQuota — 充值配额扣减（recharge FIFO）
func TestPostConsumeQuota_RechargeQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	// User with no free quota, only recharge
	createBillingUser(t, db, 20, 0, 0)
	createBillingToken(t, db, 20, 20, 100000)

	future := time.Now().Add(48 * time.Hour)
	createRechargeRecord(t, db, 20, 701, 5000, 5000, future)

	err := billing.PostConsumeQuota(20, 20, "gpt-3.5-turbo", 100, 100)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify: recharge record remaining decreased
	var rec model.UserRechargeRecord
	db.Where("order_id = ?", 701).First(&rec)
	if rec.Remaining >= 5000 {
		t.Errorf("expected recharge remaining < 5000, got %d", rec.Remaining)
	}

	// Verify: quota transaction created with quota_type=recharge
	var txCount int64
	db.Model(&model.QuotaTransaction{}).Where("user_id = ? AND quota_type = ?", 20, "recharge").Count(&txCount)
	if txCount == 0 {
		t.Error("expected quota transaction with quota_type='recharge'")
	}
}

// T-P2-08: PostConsumeQuota — VIP 配额扣减（free 和 recharge 都不足时）
func TestPostConsumeQuota_VIPQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	// User with no free quota, no recharge, only VIP
	createBillingUser(t, db, 21, 0, 100000)
	createBillingToken(t, db, 21, 21, 100000)

	err := billing.PostConsumeQuota(21, 21, "gpt-3.5-turbo", 100, 100)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	// Verify: vip_quota decreased
	var user model.User
	db.First(&user, 21)
	if user.VIPQuota >= 100000 {
		t.Errorf("expected vip_quota < 100000, got %d", user.VIPQuota)
	}

	// Verify: quota transaction created with quota_type=vip
	var txCount int64
	db.Model(&model.QuotaTransaction{}).Where("user_id = ? AND quota_type = ?", 21, "vip").Count(&txCount)
	if txCount == 0 {
		t.Error("expected quota transaction with quota_type='vip'")
	}
}

// T-P2-09: PostConsumeQuota — free→recharge→vip 三级级联扣减
func TestPostConsumeQuota_Cascade(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	// free=100, recharge=200 (1 record), vip=10000
	// consume ~400 quota → should drain free + recharge partially + vip
	createBillingUser(t, db, 22, 100, 10000)
	createBillingToken(t, db, 22, 22, 100000)

	future := time.Now().Add(48 * time.Hour)
	createRechargeRecord(t, db, 22, 801, 200, 200, future)

	// 400 tokens at rate 1.0 → 400 quota
	err := billing.PostConsumeQuota(22, 22, "gpt-3.5-turbo", 200, 200)
	if err != nil {
		t.Fatalf("PostConsumeQuota cascade failed: %v", err)
	}

	var user model.User
	db.First(&user, 22)

	// free_quota should be 0 (all 100 consumed)
	if user.FreeQuota != 0 {
		t.Errorf("expected free_quota=0, got %d", user.FreeQuota)
	}

	// vip_quota should have been partially consumed
	if user.VIPQuota >= 10000 {
		t.Errorf("expected vip_quota < 10000 after cascade, got %d", user.VIPQuota)
	}

	// Verify: transactions exist for all three quota types
	var freeCount, rechargeCount, vipCount int64
	db.Model(&model.QuotaTransaction{}).Where("user_id = ? AND quota_type = ?", 22, "free").Count(&freeCount)
	db.Model(&model.QuotaTransaction{}).Where("user_id = ? AND quota_type = ?", 22, "recharge").Count(&rechargeCount)
	db.Model(&model.QuotaTransaction{}).Where("user_id = ? AND quota_type = ?", 22, "vip").Count(&vipCount)

	if freeCount == 0 {
		t.Error("expected free quota transaction")
	}
	if rechargeCount == 0 {
		t.Error("expected recharge quota transaction")
	}
	if vipCount == 0 {
		t.Error("expected vip quota transaction")
	}
}

// T-P2-10: PostConsumeQuota — token.remain_quota 递减
func TestPostConsumeQuota_TokenRemainQuota(t *testing.T) {
	db, billing, _, _ := setupBillingTest(t)

	createBillingUser(t, db, 23, 100000, 0)
	createBillingToken(t, db, 23, 23, 50000)

	before := int64(50000)

	err := billing.PostConsumeQuota(23, 23, "gpt-3.5-turbo", 50, 50)
	if err != nil {
		t.Fatalf("PostConsumeQuota failed: %v", err)
	}

	var token model.Token
	db.First(&token, 23)

	if token.RemainQuota >= before {
		t.Errorf("expected remain_quota < %d, got %d", before, token.RemainQuota)
	}
	if token.UsedQuota == 0 {
		t.Error("expected used_quota > 0")
	}
}

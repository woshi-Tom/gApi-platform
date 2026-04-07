package repository

import (
	"testing"

	"gapi-platform/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	if err != nil {
		t.Fatalf("failed to open test db: %v", err)
	}

	err = db.AutoMigrate(
		&model.User{},
		&model.Token{},
		&model.Channel{},
		&model.Order{},
		&model.Payment{},
		&model.AuditLog{},
		&model.LoginLog{},
		&model.VIPPackage{},
		&model.RechargePackage{},
		&model.APIAccessLog{},
	)
	if err != nil {
		t.Fatalf("failed to migrate: %v", err)
	}

	return db
}

func TestUserRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
		Level:    "free",
		Status:   "active",
	}

	err := repo.Create(user)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}

	if user.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestUserRepository_GetByID(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Username: "testuser",
		Email:    "test@example.com",
	}
	repo.Create(user)

	found, err := repo.GetByID(user.ID)
	if err != nil {
		t.Errorf("GetByID() error = %v", err)
	}
	if found.Username != user.Username {
		t.Errorf("GetByID() = %v, want %v", found.Username, user.Username)
	}
}

func TestUserRepository_GetByEmail(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	email := "unique@example.com"
	user := &model.User{
		Username: "testuser",
		Email:    email,
	}
	repo.Create(user)

	found, err := repo.GetByEmail(email)
	if err != nil {
		t.Errorf("GetByEmail() error = %v", err)
	}
	if found.Username != user.Username {
		t.Errorf("GetByEmail() = %v, want %v", found.Username, user.Username)
	}
}

func TestUserRepository_GetByEmail_NotFound(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	_, err := repo.GetByEmail("notfound@example.com")
	if err == nil {
		t.Error("GetByEmail() expected error for non-existent email")
	}
}

func TestUserRepository_Update(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Username: "original",
		Email:    "test@example.com",
	}
	repo.Create(user)

	user.Username = "updated"
	err := repo.Update(user)
	if err != nil {
		t.Errorf("Update() error = %v", err)
	}

	found, _ := repo.GetByID(user.ID)
	if found.Username != "updated" {
		t.Errorf("Update() did not persist change")
	}
}

func TestUserRepository_Delete(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	user := &model.User{
		Username: "todelete",
		Email:    "delete@example.com",
	}
	repo.Create(user)

	err := repo.Delete(user.ID)
	if err != nil {
		t.Errorf("Delete() error = %v", err)
	}

	_, err = repo.GetByID(user.ID)
	if err == nil {
		t.Error("Delete() did not remove user")
	}
}

func TestUserRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	for i := 0; i < 5; i++ {
		repo.Create(&model.User{
			Username: "user",
			Email:    "test@example.com",
		})
	}

	users, total, err := repo.List(1, 3, "", "", "")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(users) != 3 {
		t.Errorf("List() returned %d users, want 3", len(users))
	}
	if total != 5 {
		t.Errorf("List() total = %d, want 5", total)
	}
}

func TestUserRepository_List_WithFilters(t *testing.T) {
	db := setupTestDB(t)
	repo := NewUserRepository(db)

	repo.Create(&model.User{Username: "vip1", Email: "vip1@example.com", Level: "vip"})
	repo.Create(&model.User{Username: "vip2", Email: "vip2@example.com", Level: "vip"})
	repo.Create(&model.User{Username: "free1", Email: "free1@example.com", Level: "free"})

	users, _, err := repo.List(1, 10, "vip", "", "")
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(users) != 2 {
		t.Errorf("List() filtered returned %d users, want 2", len(users))
	}
}

func TestTokenRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)

	token := &model.Token{
		UserID:      1,
		TokenKey:    "test-key-123",
		Name:        "Test Token",
		RemainQuota: 1000,
	}
	err := repo.Create(token)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if token.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestTokenRepository_GetByKey(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)

	token := &model.Token{
		UserID:   1,
		TokenKey: "unique-key-456",
		Name:     "Test Token",
	}
	repo.Create(token)

	found, err := repo.GetByKey("unique-key-456")
	if err != nil {
		t.Errorf("GetByKey() error = %v", err)
	}
	if found.Name != token.Name {
		t.Errorf("GetByKey() = %v, want %v", found.Name, token.Name)
	}
}

func TestTokenRepository_ListByUser(t *testing.T) {
	db := setupTestDB(t)
	repo := NewTokenRepository(db)

	userID := uint(1)
	for i := 0; i < 3; i++ {
		repo.Create(&model.Token{UserID: userID, TokenKey: "key-" + string(rune('a'+i)), Name: "Token"})
	}
	repo.Create(&model.Token{UserID: userID + 1, TokenKey: "other", Name: "Other"})

	tokens, err := repo.ListByUser(userID)
	if err != nil {
		t.Errorf("ListByUser() error = %v", err)
	}
	if len(tokens) != 3 {
		t.Errorf("ListByUser() returned %d tokens, want 3", len(tokens))
	}
}

func TestChannelRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewChannelRepository(db)

	channel := &model.Channel{
		Name:    "Test Channel",
		Type:    "openai",
		BaseURL: "https://api.openai.com",
		Status:  1,
		Weight:  100,
	}
	err := repo.Create(channel)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if channel.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestChannelRepository_GetActiveChannels(t *testing.T) {
	db := setupTestDB(t)
	repo := NewChannelRepository(db)

	repo.Create(&model.Channel{Name: "Test Active", Status: 1, IsHealthy: true})

	channels, err := repo.GetActiveChannels()
	if err != nil {
		t.Errorf("GetActiveChannels() error = %v", err)
	}
	if len(channels) < 1 {
		t.Errorf("GetActiveChannels() returned %d channels, want >= 1", len(channels))
	}
}

func TestChannelRepository_GetByModel(t *testing.T) {
	t.Skip("Skipping: ILIKE not supported by SQLite")
}

func TestChannelRepository_IncrementFailureCount(t *testing.T) {
	db := setupTestDB(t)
	repo := NewChannelRepository(db)

	channel := &model.Channel{Name: "Test", Status: 1, IsHealthy: true, FailureCount: 0}
	repo.Create(channel)

	err := repo.IncrementFailureCount(channel.ID)
	if err != nil {
		t.Errorf("IncrementFailureCount() error = %v", err)
	}

	updated, _ := repo.GetByID(channel.ID)
	if updated.FailureCount != 1 {
		t.Errorf("IncrementFailureCount() = %d, want 1", updated.FailureCount)
	}
}

func TestChannelRepository_ResetFailureCount(t *testing.T) {
	t.Skip("Skipping: NOW() function not supported by SQLite")
}

func TestOrderRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)

	order := &model.Order{
		UserID:    1,
		OrderNo:   "ORD-001",
		OrderType: "vip",
		Status:    "pending",
	}
	err := repo.Create(order)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if order.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestOrderRepository_GetByOrderNo(t *testing.T) {
	db := setupTestDB(t)
	repo := NewOrderRepository(db)

	order := &model.Order{
		UserID:    1,
		OrderNo:   "ORD-UNIQUE-002",
		OrderType: "recharge",
		Status:    "pending",
	}
	repo.Create(order)

	found, err := repo.GetByOrderNo("ORD-UNIQUE-002")
	if err != nil {
		t.Errorf("GetByOrderNo() error = %v", err)
	}
	if found.OrderType != "recharge" {
		t.Errorf("GetByOrderNo() = %v, want recharge", found.OrderType)
	}
}

func TestVIPPackageRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewVIPPackageRepository(db)

	repo.Create(&model.VIPPackage{Name: "VIP Basic", Status: "active", IsVisible: true})

	packages, err := repo.List()
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(packages) < 1 {
		t.Errorf("List() returned %d packages, want >= 1", len(packages))
	}
}

func TestRechargePackageRepository_List(t *testing.T) {
	db := setupTestDB(t)
	repo := NewRechargePackageRepository(db)

	repo.Create(&model.RechargePackage{Name: "100 Tokens", Status: "active", IsVisible: true})

	packages, err := repo.List()
	if err != nil {
		t.Errorf("List() error = %v", err)
	}
	if len(packages) < 1 {
		t.Errorf("List() returned %d packages, want >= 1", len(packages))
	}
}

func TestAuditRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuditRepository(db)

	userID := uint(1)
	log := &model.AuditLog{
		UserID:        &userID,
		Username:      "testuser",
		Action:        "login",
		ActionGroup:   "auth",
		LogType:       "user",
		Success:       true,
		RequestPath:   "/api/v1/user/login",
		RequestMethod: "POST",
	}
	err := repo.Create(log)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if log.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestAuditRepository_ListBrief(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAuditRepository(db)

	for i := 0; i < 5; i++ {
		uid := uint(i)
		repo.Create(&model.AuditLog{
			UserID:      &uid,
			Username:    "user",
			Action:      "login",
			ActionGroup: "auth",
			LogType:     "user",
			Success:     true,
		})
	}

	logs, total, err := repo.ListBrief(1, 10, 0, "", "user", "", "", nil)
	if err != nil {
		t.Errorf("ListBrief() error = %v", err)
	}
	if total != 5 {
		t.Errorf("ListBrief() total = %d, want 5", total)
	}
	if len(logs) != 5 {
		t.Errorf("ListBrief() returned %d logs, want 5", len(logs))
	}
}

func TestLoginLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewLoginLogRepository(db)

	log := &model.LoginLog{
		Username: "testuser",
		IP:       "192.168.1.1",
		Success:  true,
	}
	err := repo.Create(log)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if log.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestAPIAccessLogRepository_Create(t *testing.T) {
	db := setupTestDB(t)
	repo := NewAPIAccessLogRepository(db)

	tokenID := uint(1)
	log := &model.APIAccessLog{
		UserID:   1,
		TokenID:  &tokenID,
		Endpoint: "/api/v1/chat/completions",
		Method:   "POST",
		Model:    "gpt-4",
	}
	err := repo.Create(log)
	if err != nil {
		t.Errorf("Create() error = %v", err)
	}
	if log.ID == 0 {
		t.Error("Create() did not set ID")
	}
}

func TestChannelRepository_CountByStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewChannelRepository(db)

	repo.Create(&model.Channel{Name: "Enabled 1", Status: 1})
	repo.Create(&model.Channel{Name: "Enabled 2", Status: 1})

	counts, err := repo.CountByStatus()
	if err != nil {
		t.Errorf("CountByStatus() error = %v", err)
	}
	if counts["enabled"] != 2 {
		t.Errorf("CountByStatus() enabled = %d, want 2", counts["enabled"])
	}
}

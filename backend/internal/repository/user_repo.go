package repository

import (
	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

// UserRepository handles user database operations
type UserRepository struct {
	db *gorm.DB
}

// NewUserRepository creates a new user repository
func NewUserRepository(db *gorm.DB) *UserRepository {
	return &UserRepository{db: db}
}

func (r *UserRepository) GetDB() *gorm.DB {
	return r.db
}

// Create creates a new user
func (r *UserRepository) Create(user *model.User) error {
	return r.db.Create(user).Error
}

// GetByID gets a user by ID
func (r *UserRepository) GetByID(id uint) (*model.User, error) {
	var user model.User
	err := r.db.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByEmail gets a user by email
func (r *UserRepository) GetByEmail(email string) (*model.User, error) {
	var user model.User
	err := r.db.Where("email = ?", email).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetByUsername gets a user by username
func (r *UserRepository) GetByUsername(username string) (*model.User, error) {
	var user model.User
	err := r.db.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// Update updates a user
func (r *UserRepository) Update(user *model.User) error {
	return r.db.Save(user).Error
}

// Delete soft deletes a user
func (r *UserRepository) Delete(id uint) error {
	return r.db.Delete(&model.User{}, id).Error
}

// List lists users with pagination
func (r *UserRepository) List(page, pageSize int, level, status, keyword string) ([]model.User, int64, error) {
	var users []model.User
	var total int64

	query := r.db.Model(&model.User{})

	if level != "" {
		query = query.Where("level = ?", level)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if keyword != "" {
		query = query.Where("username ILIKE ? OR email ILIKE ?", "%"+keyword+"%", "%"+keyword+"%")
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("created_at DESC").Find(&users).Error
	if err != nil {
		return nil, 0, err
	}

	return users, total, nil
}

// UpdateQuota updates user quota
func (r *UserRepository) UpdateQuota(userID uint, quotaType string, amount int64) error {
	user, err := r.GetByID(userID)
	if err != nil {
		return err
	}

	if quotaType == "permanent" {
		user.RemainQuota += amount
	} else {
		user.VIPQuota += amount
	}

	return r.Update(user)
}

// TokenRepository handles token database operations
type TokenRepository struct {
	db *gorm.DB
}

// NewTokenRepository creates a new token repository
func NewTokenRepository(db *gorm.DB) *TokenRepository {
	return &TokenRepository{db: db}
}

// Create creates a new token
func (r *TokenRepository) Create(token *model.Token) error {
	return r.db.Create(token).Error
}

// GetByID gets a token by ID
func (r *TokenRepository) GetByID(id uint) (*model.Token, error) {
	var token model.Token
	err := r.db.First(&token, id).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// GetByKey gets a token by token key
func (r *TokenRepository) GetByKey(tokenKey string) (*model.Token, error) {
	var token model.Token
	err := r.db.Where("token_key = ?", tokenKey).First(&token).Error
	if err != nil {
		return nil, err
	}
	return &token, nil
}

// Update updates a token
func (r *TokenRepository) Update(token *model.Token) error {
	return r.db.Save(token).Error
}

// Delete soft deletes a token
func (r *TokenRepository) Delete(id uint) error {
	return r.db.Delete(&model.Token{}, id).Error
}

// List lists tokens for a user
func (r *TokenRepository) ListByUser(userID uint) ([]model.Token, error) {
	var tokens []model.Token
	err := r.db.Where("user_id = ?", userID).Order("created_at DESC").Find(&tokens).Error
	return tokens, err
}

// UpdateQuota updates token quota
func (r *TokenRepository) UpdateQuota(tokenID uint, amount int64) error {
	return r.db.Model(&model.Token{}).Where("id = ?", tokenID).
		UpdateColumn("remain_quota", gorm.Expr("remain_quota + ?", amount)).Error
}

// OrderRepository handles order database operations
type OrderRepository struct {
	db *gorm.DB
}

// NewOrderRepository creates a new order repository
func NewOrderRepository(db *gorm.DB) *OrderRepository {
	return &OrderRepository{db: db}
}

// Create creates a new order
func (r *OrderRepository) Create(order *model.Order) error {
	return r.db.Create(order).Error
}

// GetByID gets an order by ID
func (r *OrderRepository) GetByID(id uint) (*model.Order, error) {
	var order model.Order
	err := r.db.First(&order, id).Error
	return &order, err
}

// List lists orders with pagination and filters
func (r *OrderRepository) List(page, pageSize int, userID uint, orderType, status, startDate, endDate string) ([]model.Order, int64, error) {
	var orders []model.Order
	var total int64
	query := r.db.Model(&model.Order{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if orderType != "" {
		query = query.Where("order_type = ?", orderType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&orders).Error
	return orders, total, err
}

// Update updates an order
func (r *OrderRepository) Update(order *model.Order) error {
	return r.db.Save(order).Error
}

// PaymentRepository handles payment database operations
type PaymentRepository struct {
	db *gorm.DB
}

// NewPaymentRepository creates a new payment repository
func NewPaymentRepository(db *gorm.DB) *PaymentRepository {
	return &PaymentRepository{db: db}
}

// Create creates a new payment
func (r *PaymentRepository) Create(payment *model.Payment) error {
	return r.db.Create(payment).Error
}

// GetByID gets a payment by ID
func (r *PaymentRepository) GetByID(id uint) (*model.Payment, error) {
	var payment model.Payment
	err := r.db.First(&payment, id).Error
	return &payment, err
}

// Update updates a payment
func (r *PaymentRepository) Update(payment *model.Payment) error {
	return r.db.Save(payment).Error
}

// AuditRepository handles audit log database operations
type AuditRepository struct {
	db *gorm.DB
}

// NewAuditRepository creates a new audit repository
func NewAuditRepository(db *gorm.DB) *AuditRepository {
	return &AuditRepository{db: db}
}

// Create creates a new audit log
func (r *AuditRepository) Create(log *model.AuditLog) error {
	return r.db.Create(log).Error
}

// List lists audit logs with pagination and filters
func (r *AuditRepository) List(page, pageSize int, userID uint, actionGroup, action, resourceType, startTime, endTime string, success *bool) ([]model.AuditLog, int64, error) {
	var logs []model.AuditLog
	var total int64
	query := r.db.Model(&model.AuditLog{})
	if userID > 0 {
		query = query.Where("user_id = ?", userID)
	}
	if actionGroup != "" {
		query = query.Where("action_group = ?", actionGroup)
	}
	if success != nil {
		query = query.Where("success = ?", *success)
	}
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

// LoginLogRepository handles login log database operations
type LoginLogRepository struct {
	db *gorm.DB
}

// NewLoginLogRepository creates a new login log repository
func NewLoginLogRepository(db *gorm.DB) *LoginLogRepository {
	return &LoginLogRepository{db: db}
}

// Create creates a new login log
func (r *LoginLogRepository) Create(log *model.LoginLog) error {
	return r.db.Create(log).Error
}

// List lists login logs with filtering
func (r *LoginLogRepository) List(page, pageSize int, username string, ip string, success *bool, startTime string, endTime string) ([]model.LoginLog, int64, error) {
	var logs []model.LoginLog
	var total int64

	query := r.db.Model(&model.LoginLog{})

	if username != "" {
		query = query.Where("username = ?", username)
	}
	if ip != "" {
		query = query.Where("ip = ?", ip)
	}
	if success != nil {
		query = query.Where("success = ?", *success)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	query.Count(&total)
	query = query.Order("created_at DESC")
	query = query.Offset((page - 1) * pageSize).Limit(pageSize)
	err := query.Find(&logs).Error

	return logs, total, err
}

// VIPPackageRepository handles VIP package database operations
type VIPPackageRepository struct {
	db *gorm.DB
}

// NewVIPPackageRepository creates a new VIP package repository
func NewVIPPackageRepository(db *gorm.DB) *VIPPackageRepository {
	return &VIPPackageRepository{db: db}
}

// List lists all active VIP packages
func (r *VIPPackageRepository) List() ([]model.VIPPackage, error) {
	var packages []model.VIPPackage
	err := r.db.Where("status = 'active' AND is_visible = true").Order("sort_order ASC, price ASC").Find(&packages).Error
	return packages, err
}

// GetByID gets a VIP package by ID
func (r *VIPPackageRepository) GetByID(id uint) (*model.VIPPackage, error) {
	var pkg model.VIPPackage
	err := r.db.First(&pkg, id).Error
	return &pkg, err
}

func (r *VIPPackageRepository) Create(pkg *model.VIPPackage) error {
	return r.db.Create(pkg).Error
}

func (r *VIPPackageRepository) Update(pkg *model.VIPPackage) error {
	return r.db.Save(pkg).Error
}

func (r *VIPPackageRepository) Delete(id uint) error {
	return r.db.Model(&model.VIPPackage{}).Where("id = ?", id).Update("status", "deleted").Error
}

// RechargePackageRepository handles recharge package database operations
type RechargePackageRepository struct {
	db *gorm.DB
}

// NewRechargePackageRepository creates a new recharge package repository
func NewRechargePackageRepository(db *gorm.DB) *RechargePackageRepository {
	return &RechargePackageRepository{db: db}
}

// List lists all active recharge packages
func (r *RechargePackageRepository) List() ([]model.RechargePackage, error) {
	var packages []model.RechargePackage
	err := r.db.Where("status = 'active' AND is_visible = true").Order("sort_order ASC, price ASC").Find(&packages).Error
	return packages, err
}

// GetByID gets a recharge package by ID
func (r *RechargePackageRepository) GetByID(id uint) (*model.RechargePackage, error) {
	var pkg model.RechargePackage
	err := r.db.First(&pkg, id).Error
	return &pkg, err
}

func (r *RechargePackageRepository) Create(pkg *model.RechargePackage) error {
	return r.db.Create(pkg).Error
}

func (r *RechargePackageRepository) Update(pkg *model.RechargePackage) error {
	return r.db.Save(pkg).Error
}

func (r *RechargePackageRepository) Delete(id uint) error {
	return r.db.Model(&model.RechargePackage{}).Where("id = ?", id).Update("status", "deleted").Error
}

// APIAccessLogRepository handles API access log database operations
type APIAccessLogRepository struct {
	db *gorm.DB
}

func NewAPIAccessLogRepository(db *gorm.DB) *APIAccessLogRepository {
	return &APIAccessLogRepository{db: db}
}

func (r *APIAccessLogRepository) Create(log *model.APIAccessLog) error {
	return r.db.Create(log).Error
}

func (r *APIAccessLogRepository) ListByUser(userID uint, page, pageSize int) ([]model.APIAccessLog, int64, error) {
	var logs []model.APIAccessLog
	var total int64

	query := r.db.Model(&model.APIAccessLog{}).Where("user_id = ?", userID)
	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

func (r *APIAccessLogRepository) List(page, pageSize int, userID *uint, startTime, endTime string) ([]model.APIAccessLog, int64, error) {
	var logs []model.APIAccessLog
	var total int64

	query := r.db.Model(&model.APIAccessLog{})

	if userID != nil {
		query = query.Where("user_id = ?", *userID)
	}
	if startTime != "" {
		query = query.Where("created_at >= ?", startTime)
	}
	if endTime != "" {
		query = query.Where("created_at <= ?", endTime)
	}

	query.Count(&total)
	err := query.Offset((page - 1) * pageSize).Limit(pageSize).Order("created_at DESC").Find(&logs).Error
	return logs, total, err
}

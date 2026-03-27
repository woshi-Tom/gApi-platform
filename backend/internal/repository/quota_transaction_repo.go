package repository

import (
	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type QuotaTransactionRepository struct {
	db *gorm.DB
}

func NewQuotaTransactionRepository(db *gorm.DB) *QuotaTransactionRepository {
	return &QuotaTransactionRepository{db: db}
}

func (r *QuotaTransactionRepository) Create(tx *model.QuotaTransaction) error {
	return r.db.Create(tx).Error
}

func (r *QuotaTransactionRepository) GetByID(id uint) (*model.QuotaTransaction, error) {
	var tx model.QuotaTransaction
	err := r.db.First(&tx, id).Error
	if err != nil {
		return nil, err
	}
	return &tx, nil
}

func (r *QuotaTransactionRepository) List(userID uint, page, pageSize int) ([]model.QuotaTransaction, int64, error) {
	var txs []model.QuotaTransaction
	var total int64

	query := r.db.Model(&model.QuotaTransaction{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&txs).Error
	if err != nil {
		return nil, 0, err
	}

	return txs, total, nil
}

func (r *QuotaTransactionRepository) GetByType(userID uint, txType string) ([]model.QuotaTransaction, error) {
	var txs []model.QuotaTransaction
	err := r.db.Where("user_id = ? AND type = ?", userID, txType).
		Order("created_at DESC").
		Find(&txs).Error
	return txs, err
}

func (r *QuotaTransactionRepository) GetBalance(userID uint) (int64, int64, error) {
	var user model.User
	err := r.db.First(&user, userID).Error
	if err != nil {
		return 0, 0, err
	}
	return user.RemainQuota, user.VIPQuota, nil
}

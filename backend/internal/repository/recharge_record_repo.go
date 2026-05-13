package repository

import (
	"time"

	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type UserRechargeRecordRepository struct {
	db *gorm.DB
}

func NewUserRechargeRecordRepository(db *gorm.DB) *UserRechargeRecordRepository {
	return &UserRechargeRecordRepository{db: db}
}

// GetActiveByUser returns active (non-expired, non-used) recharge records for a user,
// ordered by expiration time ascending (FIFO).
func (r *UserRechargeRecordRepository) GetActiveByUser(userID uint) ([]model.UserRechargeRecord, error) {
	var records []model.UserRechargeRecord
	err := r.db.Where("user_id = ? AND status = 'active' AND remaining > 0 AND expired_at > ?",
		userID, time.Now()).
		Order("expired_at ASC").
		Find(&records).Error
	return records, err
}

// GetTotalActiveQuota returns the sum of remaining quota from active recharge records.
func (r *UserRechargeRecordRepository) GetTotalActiveQuota(userID uint) int64 {
	var total int64
	r.db.Model(&model.UserRechargeRecord{}).
		Where("user_id = ? AND status = 'active' AND remaining > 0 AND expired_at > ?",
			userID, time.Now()).
		Select("COALESCE(SUM(remaining), 0)").
		Scan(&total)
	return total
}

// UpdateRemaining updates a recharge record's remaining quota and status.
func (r *UserRechargeRecordRepository) UpdateRemaining(recordID uint, remaining int64, status string) error {
	return r.db.Model(&model.UserRechargeRecord{}).
		Where("id = ?", recordID).
		Updates(map[string]interface{}{
			"remaining":  remaining,
			"status":     status,
			"updated_at": time.Now(),
		}).Error
}

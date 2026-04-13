package repository

import (
	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type ChannelTestHistoryRepository struct {
	db *gorm.DB
}

func NewChannelTestHistoryRepository(db *gorm.DB) *ChannelTestHistoryRepository {
	return &ChannelTestHistoryRepository{db: db}
}

func (r *ChannelTestHistoryRepository) Create(history *model.ChannelTestHistory) error {
	return r.db.Create(history).Error
}

func (r *ChannelTestHistoryRepository) ListByChannelID(channelID uint, limit int) ([]model.ChannelTestHistory, error) {
	var history []model.ChannelTestHistory
	err := r.db.Where("channel_id = ?", channelID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

func (r *ChannelTestHistoryRepository) ListByUserID(userID uint, limit int) ([]model.ChannelTestHistory, error) {
	var history []model.ChannelTestHistory
	err := r.db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(limit).
		Find(&history).Error
	return history, err
}

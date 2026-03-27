package repository

import (
	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type ChannelRepository struct {
	db *gorm.DB
}

func NewChannelRepository(db *gorm.DB) *ChannelRepository {
	return &ChannelRepository{db: db}
}

func (r *ChannelRepository) Create(channel *model.Channel) error {
	return r.db.Create(channel).Error
}

func (r *ChannelRepository) GetByID(id uint) (*model.Channel, error) {
	var channel model.Channel
	err := r.db.First(&channel, id).Error
	if err != nil {
		return nil, err
	}
	return &channel, nil
}

func (r *ChannelRepository) Update(channel *model.Channel) error {
	return r.db.Save(channel).Error
}

func (r *ChannelRepository) Delete(id uint) error {
	return r.db.Delete(&model.Channel{}, id).Error
}

func (r *ChannelRepository) List(page, pageSize int, channelType, status, group, keyword string) ([]model.Channel, int64, error) {
	var channels []model.Channel
	var total int64

	query := r.db.Model(&model.Channel{})

	if channelType != "" {
		query = query.Where("type = ?", channelType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if group != "" {
		query = query.Where("group_name = ?", group)
	}
	if keyword != "" {
		query = query.Where("name ILIKE ?", "%"+keyword+"%")
	}

	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("id DESC").Offset(offset).Limit(pageSize).Find(&channels).Error
	if err != nil {
		return nil, 0, err
	}

	return channels, total, nil
}

func (r *ChannelRepository) GetActiveChannels() ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ? AND is_healthy = ?", 1, true).
		Order("priority DESC, weight DESC").
		Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) GetByModel(modelName string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ? AND is_healthy = ?", 1, true).
		Where("models ILIKE ?", "%"+modelName+"%").
		Order("priority DESC, weight DESC").
		Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) UpdateHealthStatus(id uint, isHealthy bool, failureCount int, lastError string) error {
	updates := map[string]interface{}{
		"is_healthy": isHealthy,
	}
	if isHealthy {
		updates["failure_count"] = 0
	} else {
		updates["failure_count"] = failureCount
		updates["last_error"] = lastError
	}
	return r.db.Model(&model.Channel{}).Where("id = ?", id).Updates(updates).Error
}

func (r *ChannelRepository) IncrementFailureCount(id uint) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		UpdateColumn("failure_count", gorm.Expr("failure_count + 1")).Error
}

func (r *ChannelRepository) ResetFailureCount(id uint) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"failure_count":  0,
			"is_healthy":     true,
			"last_success_at": gorm.Expr("NOW()"),
		}).Error
}

func (r *ChannelRepository) UpdateResponseTime(id uint, responseTimeMs int) error {
	return r.db.Model(&model.Channel{}).Where("id = ?", id).
		Updates(map[string]interface{}{
			"response_time_avg": responseTimeMs,
			"last_check_at":     gorm.Expr("NOW()"),
		}).Error
}

func (r *ChannelRepository) GetByGroup(groupName string) ([]model.Channel, error) {
	var channels []model.Channel
	err := r.db.Where("status = ? AND is_healthy = ? AND group_name = ?", 1, true, groupName).
		Order("priority DESC, weight DESC").
		Find(&channels).Error
	return channels, err
}

func (r *ChannelRepository) CountByStatus() (map[string]int64, error) {
	type Result struct {
		Status int64
		Count  int64
	}
	var results []Result
	err := r.db.Model(&model.Channel{}).
		Select("status, COUNT(*) as count").
		Group("status").
		Find(&results).Error
	if err != nil {
		return nil, err
	}

	counts := make(map[string]int64)
	for _, r := range results {
		key := "unknown"
		switch r.Status {
		case 0:
			key = "disabled"
		case 1:
			key = "enabled"
		case 2:
			key = "maintenance"
		}
		counts[key] = r.Count
	}
	return counts, nil
}

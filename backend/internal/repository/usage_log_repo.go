package repository

import (
	"time"

	"gapi-platform/internal/model"
	"gorm.io/gorm"
)

type UsageLogRepository struct {
	db *gorm.DB
}

func NewUsageLogRepository(db *gorm.DB) *UsageLogRepository {
	return &UsageLogRepository{db: db}
}

func (r *UsageLogRepository) Create(log *model.UsageLog) error {
	return r.db.Create(log).Error
}

func (r *UsageLogRepository) GetByID(id uint) (*model.UsageLog, error) {
	var log model.UsageLog
	err := r.db.First(&log, id).Error
	if err != nil {
		return nil, err
	}
	return &log, nil
}

func (r *UsageLogRepository) List(userID uint, page, pageSize int) ([]model.UsageLog, int64, error) {
	var logs []model.UsageLog
	var total int64

	query := r.db.Model(&model.UsageLog{}).Where("user_id = ?", userID)
	query.Count(&total)

	offset := (page - 1) * pageSize
	err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&logs).Error
	if err != nil {
		return nil, 0, err
	}

	return logs, total, nil
}

func (r *UsageLogRepository) GetUserStats(userID uint, days int) (*UsageStats, error) {
	startDate := time.Now().AddDate(0, 0, -days)

	var stats UsageStats

	r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Select("COUNT(*) as total_requests, COALESCE(SUM(total_tokens), 0) as total_tokens, COALESCE(SUM(cost), 0) as total_cost").
		Scan(&stats)

	r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Select("model, SUM(total_tokens) as tokens").
		Group("model").
		Scan(&stats.ModelBreakdown)

	r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, startDate).
		Select("DATE(created_at) as date, SUM(total_tokens) as tokens, COUNT(*) as requests").
		Group("DATE(created_at)").
		Order("date DESC").
		Scan(&stats.DailyTrend)

	return &stats, nil
}

func (r *UsageLogRepository) GetDailyStats(userID uint) (int64, int64, error) {
	today := time.Now().Truncate(24 * time.Hour)

	var totalTokens int64
	var totalRequests int64

	r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Select("COALESCE(SUM(total_tokens), 0)").
		Scan(&totalTokens)

	r.db.Model(&model.UsageLog{}).
		Where("user_id = ? AND created_at >= ?", userID, today).
		Select("COUNT(*)").
		Scan(&totalRequests)

	return totalTokens, totalRequests, nil
}

type UsageStats struct {
	TotalRequests   int64
	TotalTokens     int64
	TotalCost       float64
	ModelBreakdown  map[string]int64
	ChannelBreakdown map[string]int64
	DailyTrend      []DailyUsage
}

type DailyUsage struct {
	Date      string
	Tokens    int64
	Requests  int64
	Cost      float64
}

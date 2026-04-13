package worker

import (
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/logger"
	"gorm.io/gorm"
)

type VIPExpiryWorker struct {
	db     *gorm.DB
	period time.Duration
	stopCh chan struct{}
}

func NewVIPExpiryWorker(db *gorm.DB, period time.Duration) *VIPExpiryWorker {
	return &VIPExpiryWorker{
		db:     db,
		period: period,
		stopCh: make(chan struct{}),
	}
}

func (w *VIPExpiryWorker) Start() {
	ticker := time.NewTicker(w.period)
	defer ticker.Stop()

	logger.Info("VIP worker started", "period", w.period.String())

	for {
		select {
		case <-ticker.C:
			w.processExpiredVIPs()
		case <-w.stopCh:
			logger.Info("VIP worker stopped")
			return
		}
	}
}

func (w *VIPExpiryWorker) Stop() {
	close(w.stopCh)
}

func (w *VIPExpiryWorker) processExpiredVIPs() {
	now := time.Now()

	var users []model.User
	err := w.db.Where("level IN ? AND v_ip_expired_at IS NOT NULL AND v_ip_expired_at < ?",
		[]string{"vip_bronze", "vip_silver", "vip_gold"}, now).Find(&users).Error
	if err != nil {
		logger.Errorf("VIP worker error querying expired VIPs: %v", err)
		return
	}

	if len(users) == 0 {
		return
	}

	for _, user := range users {
		result := w.db.Model(&model.User{}).
			Where("id = ? AND level IN ? AND v_ip_expired_at < ?",
				user.ID, []string{"vip_bronze", "vip_silver", "vip_gold"}, now).
			Updates(map[string]interface{}{
				"level":           "free",
				"v_ip_expired_at": nil,
				"v_ip_package_id": 0,
			})

		if result.Error != nil {
			logger.Errorf("VIP worker error updating user %d: %v", user.ID, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			logger.Infof("VIP expired for user %d (%s), downgraded to free tier", user.ID, user.Email)
		}
	}

	logger.Infof("VIP worker processed %d expired VIP users", len(users))
}

func (w *VIPExpiryWorker) ProcessOnce() {
	w.processExpiredVIPs()
}

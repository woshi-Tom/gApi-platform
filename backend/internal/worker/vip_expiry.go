package worker

import (
	"log"
	"time"

	"gapi-platform/internal/model"
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

	log.Println("[VIPWorker] Started, running every", w.period)

	for {
		select {
		case <-ticker.C:
			w.processExpiredVIPs()
		case <-w.stopCh:
			log.Println("[VIPWorker] Stopped")
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
	err := w.db.Where("level = ? AND vip_expired_at IS NOT NULL AND vip_expired_at < ?", "vip", now).Find(&users).Error
	if err != nil {
		log.Printf("[VIPWorker] Error querying expired VIPs: %v\n", err)
		return
	}

	if len(users) == 0 {
		return
	}

	for _, user := range users {
		result := w.db.Model(&model.User{}).
			Where("id = ? AND level = 'vip' AND vip_expired_at < ?", user.ID, now).
			Updates(map[string]interface{}{
				"level":          "free",
				"vip_expired_at": nil,
				"vip_package_id": 0,
			})

		if result.Error != nil {
			log.Printf("[VIPWorker] Error updating user %d: %v\n", user.ID, result.Error)
			continue
		}

		if result.RowsAffected > 0 {
			log.Printf("[VIPWorker] VIP expired for user %d (%s), downgraded to free tier\n", user.ID, user.Email)
		}
	}

	log.Printf("[VIPWorker] Processed %d expired VIP users\n", len(users))
}

func (w *VIPExpiryWorker) ProcessOnce() {
	w.processExpiredVIPs()
}

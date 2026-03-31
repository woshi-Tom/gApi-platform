package worker

import (
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"github.com/rs/zerolog/log"
)

type ExpiryWorker struct {
	orderRepo *repository.OrderRepository
	auditRepo *repository.AuditRepository
	interval  time.Duration
	stopChan  chan struct{}
}

func NewExpiryWorker(
	orderRepo *repository.OrderRepository,
	auditRepo *repository.AuditRepository,
	interval time.Duration,
) *ExpiryWorker {
	return &ExpiryWorker{
		orderRepo: orderRepo,
		auditRepo: auditRepo,
		interval:  interval,
		stopChan:  make(chan struct{}),
	}
}

func (w *ExpiryWorker) Start() {
	log.Info().Dur("interval", w.interval).Msg("expiry worker starting")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	w.run()

	for {
		select {
		case <-ticker.C:
			w.run()
		case <-w.stopChan:
			log.Info().Msg("expiry worker stopped")
			return
		}
	}
}

func (w *ExpiryWorker) Stop() {
	close(w.stopChan)
}

func (w *ExpiryWorker) run() {
	now := time.Now()
	var orders []model.Order

	if err := w.orderRepo.GetDB().
		Where("status = ? AND expires_at IS NOT NULL AND expires_at <= ?", model.OrderStatusPending, now).
		Find(&orders).Error; err != nil {
		log.Error().Err(err).Msg("failed to query expired orders")
		return
	}

	if len(orders) == 0 {
		return
	}

	log.Info().Int("count", len(orders)).Msg("processing expired orders")

	for _, order := range orders {
		if err := w.expireOrder(&order); err != nil {
			log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to expire order")
			continue
		}
		log.Info().Str("order_no", order.OrderNo).Msg("order expired")
	}
}

func (w *ExpiryWorker) expireOrder(order *model.Order) error {
	order.Status = model.OrderStatusExpired
	order.UpdatedAt = time.Now()

	if err := w.orderRepo.GetDB().Save(order).Error; err != nil {
		return err
	}

	if w.auditRepo != nil {
		w.auditRepo.Create(&model.AuditLog{
			UserID:       &order.UserID,
			Action:       "order.expired",
			ActionGroup:  "payment",
			ResourceType: "order",
			ResourceID:   &order.ID,
			RequestPath:  "/internal/worker/expiry",
			Success:      true,
			CreatedAt:    time.Now(),
		})
	}

	return nil
}

type ReconcileWorker struct {
	orderRepo     *repository.OrderRepository
	paymentRepo   *repository.PaymentRepository
	userRepo      *repository.UserRepository
	vipRepo       *repository.VIPPackageRepository
	alipayService interface {
		IsEnabled() bool
	}
	queryFunc func(orderNo string) (tradeNo, tradeStatus, totalAmount string, err error)
	auditRepo *repository.AuditRepository
	interval  time.Duration
	stopChan  chan struct{}
}

func NewReconcileWorker(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	vipRepo *repository.VIPPackageRepository,
	alipayService interface {
		IsEnabled() bool
	},
	queryFunc func(orderNo string) (tradeNo, tradeStatus, totalAmount string, err error),
	auditRepo *repository.AuditRepository,
	interval time.Duration,
) *ReconcileWorker {
	return &ReconcileWorker{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		userRepo:      userRepo,
		vipRepo:       vipRepo,
		alipayService: alipayService,
		queryFunc:     queryFunc,
		auditRepo:     auditRepo,
		interval:      interval,
		stopChan:      make(chan struct{}),
	}
}

func (w *ReconcileWorker) Start() {
	log.Info().Dur("interval", w.interval).Msg("reconcile worker starting")
	ticker := time.NewTicker(w.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			w.run()
		case <-w.stopChan:
			log.Info().Msg("reconcile worker stopped")
			return
		}
	}
}

func (w *ReconcileWorker) Stop() {
	close(w.stopChan)
}

func (w *ReconcileWorker) run() {
	if w.alipayService == nil || !w.alipayService.IsEnabled() {
		return
	}

	var orders []model.Order

	if err := w.orderRepo.GetDB().
		Where("status IN ?", []string{model.OrderStatusPaid, model.OrderStatusPending}).
		Where("alipay_trade_no IS NOT NULL AND alipay_trade_no != ''").
		Where("completed_at IS NULL").
		Find(&orders).Error; err != nil {
		log.Error().Err(err).Msg("failed to query orders for reconciliation")
		return
	}

	if len(orders) == 0 {
		return
	}

	log.Info().Int("count", len(orders)).Msg("reconciling orders with Alipay")

	for _, order := range orders {
		w.reconcileOrder(&order)
	}
}

func (w *ReconcileWorker) reconcileOrder(order *model.Order) {
	if w.queryFunc == nil {
		log.Warn().Str("order_no", order.OrderNo).Msg("query function not configured")
		return
	}

	_, tradeStatus, _, err := w.queryFunc(order.OrderNo)
	if err != nil {
		log.Warn().Err(err).Str("order_no", order.OrderNo).Msg("failed to query Alipay")
		return
	}

	switch tradeStatus {
	case "TRADE_SUCCESS", "TRADE_FINISHED":
		if order.Status != model.OrderStatusCompleted {
			log.Info().Str("order_no", order.OrderNo).Msg("reconciling paid order to completed")
		}
	case "TRADE_CLOSED":
		if order.Status == model.OrderStatusPending {
			order.Status = model.OrderStatusExpired
			order.CancelReason = "Alipay trade closed"
			order.UpdatedAt = time.Now()
			w.orderRepo.GetDB().Save(order)
			log.Info().Str("order_no", order.OrderNo).Msg("order cancelled due to Alipay trade closure")
		}
	}
}

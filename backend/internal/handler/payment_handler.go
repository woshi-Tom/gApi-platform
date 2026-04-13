package handler

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gorm.io/gorm"
)

type PaymentHandler struct {
	orderRepo     *repository.OrderRepository
	paymentRepo   *repository.PaymentRepository
	userRepo      *repository.UserRepository
	vipRepo       *repository.VIPPackageRepository
	auditRepo     *repository.AuditRepository
	alipayService *service.AlipayService
	redisClient   *repository.RedisClient
}

func NewPaymentHandler(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	vipRepo *repository.VIPPackageRepository,
	auditRepo *repository.AuditRepository,
	alipayService *service.AlipayService,
	redisClient *repository.RedisClient,
) *PaymentHandler {
	return &PaymentHandler{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		userRepo:      userRepo,
		vipRepo:       vipRepo,
		auditRepo:     auditRepo,
		alipayService: alipayService,
		redisClient:   redisClient,
	}
}

type CreatePaymentRequest struct {
	OrderID uint   `json:"order_id"`
	OrderNo string `json:"order_no"`
}

type CreatePaymentResponse struct {
	OrderNo     string `json:"order_no"`
	QRCode      string `json:"qr_code"`
	QRExpireAt  string `json:"qr_expire_at"`
	Amount      string `json:"amount"`
	PackageName string `json:"package_name"`
}

func (h *PaymentHandler) CreateAlipay(c *gin.Context) {
	if h.alipayService == nil || !h.alipayService.IsEnabled() {
		response.Fail(c, "PAYMENT_DISABLED", "Alipay payment is not enabled")
		return
	}

	var req CreatePaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	userIDVal, exists := c.Get("user_id")
	if !exists {
		response.Fail(c, "UNAUTHORIZED", "user not authenticated")
		return
	}
	userID, ok := userIDVal.(uint)
	if !ok {
		response.Fail(c, "UNAUTHORIZED", "invalid user id")
		return
	}

	var order *model.Order
	var err error

	if req.OrderID > 0 {
		order, err = h.orderRepo.GetByID(req.OrderID)
	} else if req.OrderNo != "" {
		order, err = h.orderRepo.GetByOrderNo(req.OrderNo)
	} else {
		response.Fail(c, "INVALID_PARAMETER", "order_id or order_no is required")
		return
	}

	if err != nil || order == nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	if order.Status == model.OrderStatusExpired {
		response.Fail(c, "ORDER_EXPIRED", "order has expired")
		return
	}

	if order.Status != model.OrderStatusPending {
		response.Fail(c, "ORDER_NOT_PENDING", "order is not in pending status")
		return
	}

	if order.ExpiresAt != nil && time.Now().After(*order.ExpiresAt) {
		order.Status = model.OrderStatusExpired
		h.orderRepo.Save(order)
		response.Fail(c, "ORDER_EXPIRED", "order has expired")
		return
	}

	var existingPayment model.Payment
	if err := h.paymentRepo.GetDB().Where("order_id = ? AND status = ?", order.ID, model.PaymentStatusPending).First(&existingPayment).Error; err == nil && existingPayment.ID != 0 {
		log.Info().Str("order_no", order.OrderNo).Str("payment_no", existingPayment.PaymentNo).Msg("found existing pending Alipay payment for order")

		qrCode := existingPayment.PaymentURL
		if qrCode == "" {
			log.Info().Str("order_no", order.OrderNo).Msg("existing payment has empty QR code, requesting new one from Alipay")
			subject := fmt.Sprintf("%s - %s", order.PackageName, order.OrderNo)
			qrCode, _, err = h.alipayService.CreatePayment(order.OrderNo, order.PayAmount, subject)
			if err != nil {
				log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to create alipay payment")
				response.Fail(c, "CREATE_PAYMENT_FAILED", err.Error())
				return
			}
			existingPayment.PaymentURL = qrCode
			existingPayment.UpdatedAt = time.Now()
			if err := h.paymentRepo.GetDB().Save(&existingPayment).Error; err != nil {
				log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to update payment with new QR code")
			}
			order.AlipayQRURL = qrCode
			qrExpireTime := time.Now().Add(15 * time.Minute)
			order.QRExpireAt = &qrExpireTime
			if err := h.orderRepo.Save(order); err != nil {
				log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to update order QR expiry")
			}
		}

		expireAtStr := ""
		if order.QRExpireAt != nil {
			expireAtStr = order.QRExpireAt.Format("2006-01-02 15:04:05")
		} else {
			t := time.Now().Add(15 * time.Minute)
			order.QRExpireAt = &t
			if err := h.orderRepo.Save(order); err != nil {
				log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to update order QR expiry")
			}
			expireAtStr = t.Format("2006-01-02 15:04:05")
		}
		response.Success(c, AlipayPaymentResponse{
			OrderNo:     order.OrderNo,
			QRCode:      qrCode,
			QRExpireAt:  expireAtStr,
			Amount:      fmt.Sprintf("%.2f", order.PayAmount),
			PackageName: order.PackageName,
		})
		return
	}
	// 2) No existing pending payment, proceed to create a new one
	subject := fmt.Sprintf("%s - %s", order.PackageName, order.OrderNo)
	qrCode, expireAt, err := h.alipayService.CreatePayment(order.OrderNo, order.PayAmount, subject)
	if err != nil {
		log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to create alipay payment")
		response.Fail(c, "CREATE_PAYMENT_FAILED", err.Error())
		return
	}

	order.AlipayQRURL = qrCode
	qrExpireTime := time.Now().Add(15 * time.Minute)
	order.QRExpireAt = &qrExpireTime

	var payment *model.Payment
	err = h.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("failed to save order QR URL: %w", err)
		}

		payment = &model.Payment{
			OrderID:       order.ID,
			UserID:        userID,
			PaymentNo:     fmt.Sprintf("PAY%s%s", time.Now().Format("20060102"), uuid.New().String()[:8]),
			PaymentMethod: "alipay",
			Amount:        order.PayAmount,
			Status:        "pending",
			PaymentURL:    qrCode,
			CreatedAt:     time.Now(),
			UpdatedAt:     time.Now(),
		}
		if err := tx.Create(payment).Error; err != nil {
			return fmt.Errorf("failed to create payment record: %w", err)
		}

		if h.auditRepo != nil {
			auditLog := &model.AuditLog{
				UserID:        &userID,
				Action:        "payment.init",
				ActionGroup:   "payment",
				ResourceType:  "payment",
				ResourceID:    &payment.ID,
				RequestMethod: "POST",
				RequestPath:   "/api/v1/payment/alipay",
				RequestBody:   fmt.Sprintf(`{"order_no":"%s","payment_no":"%s","amount":"%.2f"}`, order.OrderNo, payment.PaymentNo, order.PayAmount),
				Success:       true,
				CreatedAt:     time.Now(),
			}
			if err := tx.Create(auditLog).Error; err != nil {
				log.Warn().Err(err).Str("order_no", order.OrderNo).Msg("failed to create audit log")
			}
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to process payment in transaction")
		response.Fail(c, "PROCESS_PAYMENT_FAILED", err.Error())
		return
	}

	log.Info().
		Str("order_no", order.OrderNo).
		Str("payment_no", payment.PaymentNo).
		Str("amount", fmt.Sprintf("%.2f", order.PayAmount)).
		Uint("user_id", userID).
		Msg("alipay payment created")

	response.Success(c, AlipayPaymentResponse{
		OrderNo:     order.OrderNo,
		QRCode:      qrCode,
		QRExpireAt:  expireAt,
		Amount:      fmt.Sprintf("%.2f", order.PayAmount),
		PackageName: order.PackageName,
	})
}

type PaymentStatusResponse struct {
	OrderNo     string  `json:"order_no"`
	Status      string  `json:"status"`
	Amount      float64 `json:"amount"`
	PackageName string  `json:"package_name"`
	PaidAt      *string `json:"paid_at,omitempty"`
	QRCode      string  `json:"qr_code,omitempty"`
	QRExpireAt  *string `json:"qr_expire_at,omitempty"`
}

type AlipayPaymentResponse struct {
	OrderNo     string `json:"order_no"`
	QRCode      string `json:"qr_code"`
	QRExpireAt  string `json:"qr_expire_at"`
	Amount      string `json:"amount"`
	PackageName string `json:"package_name"`
}

func (h *PaymentHandler) QueryAlipayOrder(c *gin.Context) {
	orderNo := c.Param("order_no")

	userID := c.MustGet("user_id").(uint)

	order, err := h.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	resp := PaymentStatusResponse{
		OrderNo:     order.OrderNo,
		Status:      order.Status,
		Amount:      order.PayAmount,
		PackageName: order.PackageName,
	}

	if order.PaidAt != nil {
		paidAt := order.PaidAt.Format("2006-01-02 15:04:05")
		resp.PaidAt = &paidAt
	}

	if order.Status == "pending" && order.AlipayQRURL != "" {
		resp.QRCode = order.AlipayQRURL
		if order.QRExpireAt != nil {
			expireAt := order.QRExpireAt.Format("2006-01-02T15:04:05Z")
			resp.QRExpireAt = &expireAt
		}
	}

	if order.Status == "pending" && h.alipayService != nil && h.alipayService.IsEnabled() {
		result, err := h.alipayService.QueryOrder(orderNo)
		if err == nil && result != nil {
			if result.TradeStatus == "TRADE_SUCCESS" || result.TradeStatus == "TRADE_FINISHED" {
				processErr := h.processPaymentSuccess(order, result.TradeNo, result.TotalAmount)
				if processErr != nil {
					log.Error().Err(processErr).Str("order_no", orderNo).Msg("processPaymentSuccess failed in QueryAlipayOrder")
				} else {
					paidAt := time.Now().Format("2006-01-02 15:04:05")
					resp.PaidAt = &paidAt
				}
				updatedOrder, _ := h.orderRepo.GetByOrderNo(orderNo)
				if updatedOrder != nil {
					resp.Status = updatedOrder.Status
				}
			}
		}
	}

	response.Success(c, resp)
}

func (h *PaymentHandler) CancelAlipayOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	userID := c.MustGet("user_id").(uint)

	var lockToken string
	var lockKey string
	if h.redisClient != nil {
		lockKey = fmt.Sprintf("order:lock:%s", orderNo)
		ctx := context.Background()
		token, err := h.redisClient.AcquireLock(ctx, lockKey, 10*time.Second)
		if err != nil {
			log.Warn().Err(err).Str("order_no", orderNo).Msg("failed to acquire cancel lock")
		}
		if token == "" {
			response.Fail(c, "ORDER_BEING_PROCESSED", "order is being processed, please try again later")
			return
		}
		lockToken = token
		defer h.redisClient.ReleaseLock(context.Background(), lockKey, lockToken)
	}

	order, err := h.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	if order.Status == model.OrderStatusExpired {
		response.Fail(c, "ORDER_EXPIRED", "order has already expired")
		return
	}

	if order.Status != model.OrderStatusPending {
		response.Fail(c, "ORDER_NOT_PENDING", "can only cancel pending orders")
		return
	}

	if order.ExpiresAt != nil && time.Now().After(*order.ExpiresAt) {
		order.Status = model.OrderStatusExpired
		h.orderRepo.Save(order)
		response.Fail(c, "ORDER_EXPIRED", "order has expired")
		return
	}

	if h.alipayService != nil && h.alipayService.IsEnabled() {
		if err := h.alipayService.CancelOrder(orderNo); err != nil {
			log.Warn().Err(err).Str("order_no", orderNo).Msg("failed to cancel alipay order")
		}
	}

	order.Status = model.OrderStatusCancelled
	order.CancelReason = "user cancelled"
	if err := h.orderRepo.Save(order); err != nil {
		log.Error().Err(err).Str("order_no", orderNo).Msg("failed to cancel order")
		response.Fail(c, "CANCEL_ORDER_FAILED", err.Error())
		return
	}

	log.Info().Str("order_no", orderNo).Uint("user_id", userID).Msg("order cancelled")
	response.Success(c, nil)
}

func (h *PaymentHandler) RefundOrder(c *gin.Context) {
	orderNo := c.Param("order_no")
	userID := c.MustGet("user_id").(uint)

	var lockToken string
	var lockKey string
	if h.redisClient != nil {
		lockKey = fmt.Sprintf("order:lock:%s", orderNo)
		ctx := context.Background()
		token, err := h.redisClient.AcquireLock(ctx, lockKey, 30*time.Second)
		if err != nil {
			log.Warn().Err(err).Str("order_no", orderNo).Msg("failed to acquire refund lock")
		}
		if token == "" {
			response.Fail(c, "ORDER_BEING_PROCESSED", "order is being processed, please try again later")
			return
		}
		lockToken = token
		defer h.redisClient.ReleaseLock(context.Background(), lockKey, lockToken)
	}

	order, err := h.orderRepo.GetByOrderNo(orderNo)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	if order.Status != model.OrderStatusCompleted {
		response.Fail(c, "ORDER_NOT_REFUNDABLE", "only completed orders can be refunded")
		return
	}

	err = h.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		var user model.User
		if err := tx.First(&user, order.UserID).Error; err != nil {
			return fmt.Errorf("failed to get user: %w", err)
		}

		if order.OrderType == "vip" {
			originalQuota := user.VIPQuota
			user.Level = "free"
			user.VIPQuota = 0
			user.VIPExpiredAt = nil
			user.VIPPackageID = 0
			if err := tx.Save(&user).Error; err != nil {
				return fmt.Errorf("failed to revoke VIP: %w", err)
			}
			log.Info().
				Str("order_no", orderNo).
				Uint("user_id", userID).
				Int64("revoked_quota", originalQuota).
				Msg("VIP revoked due to refund")
		} else if order.OrderType == "recharge" {
			if user.VIPQuota > 0 {
				user.VIPQuota = 0
				if err := tx.Save(&user).Error; err != nil {
					return fmt.Errorf("failed to revoke quota: %w", err)
				}
			}
		}

		order.Status = model.OrderStatusRefunded
		now := time.Now()
		order.RefundReason = "user requested refund"
		refundAmt := order.PayAmount
		order.RefundAmount = &refundAmt
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("failed to update order: %w", err)
		}

		var payment model.Payment
		if err := tx.Where("order_id = ?", order.ID).First(&payment).Error; err == nil {
			payment.Status = model.PaymentStatusRefunded
			if err := tx.Save(&payment).Error; err != nil {
				return fmt.Errorf("failed to update payment: %w", err)
			}
		}

		log.Info().
			Str("order_no", orderNo).
			Uint("user_id", userID).
			Float64("refund_amount", refundAmt).
			Time("refund_time", now).
			Msg("refund processed successfully")
		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("order_no", orderNo).Msg("refund failed")
		response.Fail(c, "REFUND_FAILED", err.Error())
		return
	}

	response.Success(c, gin.H{
		"order_no":      orderNo,
		"refund_amount": order.PayAmount,
		"message":       "refund processed, VIP has been revoked",
	})
}

func (h *PaymentHandler) processPaymentSuccess(order *model.Order, tradeNo string, amount string) error {
	if order == nil {
		return fmt.Errorf("order is nil")
	}
	if order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusCompleted {
		return nil
	}

	if h.redisClient != nil {
		ctx := context.Background()
		lockKey := fmt.Sprintf("order:lock:%s", order.OrderNo)
		token, err := h.redisClient.AcquireLock(ctx, lockKey, 30*time.Second)
		if err != nil {
			log.Warn().Err(err).Str("order_no", order.OrderNo).Msg("failed to acquire lock")
		}
		if token == "" {
			return fmt.Errorf("order is being processed by another request")
		}
		defer h.redisClient.ReleaseLock(ctx, lockKey, token)
	}

	err := h.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
		var currentOrder model.Order
		if err := tx.First(&currentOrder, order.ID).Error; err != nil {
			return fmt.Errorf("failed to fetch order: %w", err)
		}
		if currentOrder.Status == model.OrderStatusPaid || currentOrder.Status == model.OrderStatusCompleted {
			return fmt.Errorf("order already processed: %s", currentOrder.Status)
		}
		order.Status = model.OrderStatusPaid
		now := time.Now()
		order.PaidAt = &now
		order.AlipayTradeNo = tradeNo
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("update order: %w", err)
		}

		var payment model.Payment
		if err := tx.Where("order_id = ?", order.ID).First(&payment).Error; err != nil {
			log.Warn().Str("order_no", order.OrderNo).Msg("payment record not found")
		} else {
			payment.Status = model.PaymentStatusSuccess
			payment.ChannelOrderNo = tradeNo
			payment.PaidAt = &now
			if err := tx.Save(&payment).Error; err != nil {
				return fmt.Errorf("update payment: %w", err)
			}
		}

		if order.OrderType == "vip" && order.PackageID != nil {
			pkg, err := h.vipRepo.GetByID(*order.PackageID)
			if err != nil {
				return fmt.Errorf("failed to get VIP package: %w", err)
			}
			if pkg == nil {
				return fmt.Errorf("VIP package not found for id %d", *order.PackageID)
			}

			var user model.User
			if err := tx.First(&user, order.UserID).Error; err != nil {
				return fmt.Errorf("failed to get user: %w", err)
			}

			isVIP := user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"

			if isVIP {
				newQuota := user.VIPQuota + pkg.Quota
				if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).
					Update("v_ip_quota", newQuota).Error; err != nil {
					return fmt.Errorf("update VIP quota: %w", err)
				}

				var newExpireAt time.Time
				if user.VIPExpiredAt != nil && user.VIPExpiredAt.After(now) {
					newExpireAt = user.VIPExpiredAt.AddDate(0, 0, pkg.DurationDays)
				} else {
					newExpireAt = now.AddDate(0, 0, pkg.DurationDays)
				}

				if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
					"v_ip_expired_at": newExpireAt,
					"v_ip_package_id": pkg.ID,
				}).Error; err != nil {
					return fmt.Errorf("failed to update VIP status: %w", err)
				}

				log.Info().
					Str("order_no", order.OrderNo).
					Uint("user_id", order.UserID).
					Int64("vip_quota_new", newQuota).
					Str("vip_expire_at", newExpireAt.Format("2006-01-02 15:04:05")).
					Msg("VIP renewed with quota accumulation")
			} else {
				if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).
					Update("v_ip_quota", pkg.Quota).Error; err != nil {
					return fmt.Errorf("update VIP quota: %w", err)
				}

				vipExpireAt := now.AddDate(0, 0, pkg.DurationDays)
				if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
					"level":           h.getVIPLevelName(pkg),
					"v_ip_expired_at": vipExpireAt,
					"v_ip_package_id": pkg.ID,
				}).Error; err != nil {
					return fmt.Errorf("failed to update VIP status: %w", err)
				}

				log.Info().
					Str("order_no", order.OrderNo).
					Uint("user_id", order.UserID).
					Str("vip_level", h.getVIPLevelName(pkg)).
					Str("vip_expire_at", vipExpireAt.Format("2006-01-02 15:04:05")).
					Msg("VIP activated")
			}
		} else if order.OrderType == "recharge" {
			log.Info().
				Str("order_no", order.OrderNo).
				Uint("user_id", order.UserID).
				Msg("Recharge order completed - create user_recharge_record")
		}

		order.Status = model.OrderStatusCompleted
		order.CompletedAt = &now
		if err := tx.Save(order).Error; err != nil {
			return fmt.Errorf("update order to completed: %w", err)
		}

		return nil
	})

	if err != nil {
		log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to process payment success")
		if h.auditRepo != nil {
			h.auditRepo.Create(&model.AuditLog{
				UserID:       &order.UserID,
				Action:       "payment.success",
				ActionGroup:  "payment",
				ResourceType: "payment",
				Success:      false,
				ErrorMessage: err.Error(),
				CreatedAt:    time.Now(),
			})
		}
		return err
	}

	log.Info().
		Str("order_no", order.OrderNo).
		Str("trade_no", tradeNo).
		Str("amount", amount).
		Str("user_id", fmt.Sprintf("%d", order.UserID)).
		Msg("payment success processed")

	if h.auditRepo != nil {
		h.auditRepo.Create(&model.AuditLog{
			UserID:        &order.UserID,
			Action:        "payment.success",
			ActionGroup:   "payment",
			ResourceType:  "order",
			ResourceID:    &order.ID,
			RequestMethod: "POST",
			RequestPath:   "/api/v1/payment/callback/alipay",
			RequestBody:   fmt.Sprintf(`{"order_no":"%s","trade_no":"%s","amount":"%s"}`, order.OrderNo, tradeNo, amount),
			Success:       true,
			CreatedAt:     time.Now(),
		})
	}
	return nil
}

func (h *PaymentHandler) getVIPLevelName(pkg *model.VIPPackage) string {
	if pkg == nil {
		return "vip_bronze"
	}
	if pkg.Level != "" {
		return pkg.Level
	}
	return "vip_bronze"
}

func (h *PaymentHandler) AlipayNotify(c *gin.Context) {
	if h.alipayService == nil || !h.alipayService.IsEnabled() {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if err := c.Request.ParseForm(); err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	params := make(map[string]string)
	for k, v := range c.Request.PostForm {
		if len(v) > 0 {
			params[k] = v[0]
		}
	}

	result, err := h.alipayService.HandleNotify(params)
	if err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	if result.Success {
		order, err := h.orderRepo.GetByOrderNo(result.OutTradeNo)
		if err == nil && order != nil {
			amount := result.TotalAmount
			if amount == "" {
				amount = fmt.Sprintf("%.2f", order.PayAmount)
			}
			if err := h.processPaymentSuccess(order, result.TradeNo, amount); err != nil {
				log.Error().Err(err).Str("order_no", result.OutTradeNo).Msg("AlipayNotify: processPaymentSuccess failed")
			}
		}
	}

	h.alipayService.ACKNotification(c.Writer)
}

func (h *PaymentHandler) GetPaymentConfig(c *gin.Context) {
	enabled := false
	if h.alipayService != nil {
		enabled = h.alipayService.IsEnabled()
	}

	response.Success(c, gin.H{
		"alipay_enabled": enabled,
		"wechat_enabled": false,
	})
}

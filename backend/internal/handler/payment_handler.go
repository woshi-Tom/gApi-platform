package handler

import (
	"fmt"
	"net/http"
	"strconv"
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
}

func NewPaymentHandler(
	orderRepo *repository.OrderRepository,
	paymentRepo *repository.PaymentRepository,
	userRepo *repository.UserRepository,
	vipRepo *repository.VIPPackageRepository,
	auditRepo *repository.AuditRepository,
	alipayService *service.AlipayService,
) *PaymentHandler {
	return &PaymentHandler{
		orderRepo:     orderRepo,
		paymentRepo:   paymentRepo,
		userRepo:      userRepo,
		vipRepo:       vipRepo,
		auditRepo:     auditRepo,
		alipayService: alipayService,
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

	userID := c.MustGet("user_id").(uint)

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
	if err := h.orderRepo.Save(order); err != nil {
		log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to save order QR URL")
		response.Fail(c, "UPDATE_ORDER_FAILED", err.Error())
		return
	}

	payment := &model.Payment{
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
	if err := h.paymentRepo.Create(payment); err != nil {
		log.Error().Err(err).Str("order_no", order.OrderNo).Msg("failed to create payment record")
		response.Fail(c, "CREATE_PAYMENT_RECORD_FAILED", err.Error())
		return
	}

	if h.auditRepo != nil {
		success := true
		h.auditRepo.Create(&model.AuditLog{
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
		})
		_ = success
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
				h.processPaymentSuccess(order, result.TradeNo, result.TotalAmount)
				resp.Status = "paid"
				paidAt := time.Now().Format("2006-01-02 15:04:05")
				resp.PaidAt = &paidAt
			}
		}
	}

	response.Success(c, resp)
}

func (h *PaymentHandler) CancelAlipayOrder(c *gin.Context) {
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

func (h *PaymentHandler) processPaymentSuccess(order *model.Order, tradeNo string, amount string) {
	if order == nil {
		return
	}
	if order.Status == model.OrderStatusPaid || order.Status == model.OrderStatusCompleted {
		return
	}

	err := h.orderRepo.GetDB().Transaction(func(tx *gorm.DB) error {
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

			if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).
				UpdateColumn("vip_quota", gorm.Expr("vip_quota + ?", pkg.Quota)).Error; err != nil {
				return fmt.Errorf("update VIP quota: %w", err)
			}
			log.Info().
				Str("order_no", order.OrderNo).
				Uint("user_id", order.UserID).
				Int64("vip_quota_added", pkg.Quota).
				Msg("VIP quota added to user account")

			vipExpireAt := now.AddDate(0, 0, pkg.DurationDays)
			if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).Updates(map[string]interface{}{
				"level":          "vip",
				"vip_expired_at": vipExpireAt,
				"vip_package_id": pkg.ID,
			}).Error; err != nil {
				return fmt.Errorf("failed to update VIP status: %w", err)
			}
			log.Info().
				Str("order_no", order.OrderNo).
				Uint("user_id", order.UserID).
				Str("vip_expire_at", vipExpireAt.Format("2006-01-02 15:04:05")).
				Msg("VIP status updated")
		} else if order.OrderType == "recharge" || order.OrderType == "package" {
			quota, _ := strconv.ParseFloat(amount, 64)
			tokenAmount := int64(quota * 100000)
			if err := tx.Model(&model.User{}).Where("id = ?", order.UserID).
				UpdateColumn("remain_quota", gorm.Expr("remain_quota + ?", tokenAmount)).Error; err != nil {
				return fmt.Errorf("update user quota: %w", err)
			}
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
	} else {
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
	}
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
			h.processPaymentSuccess(order, result.TradeNo, amount)
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

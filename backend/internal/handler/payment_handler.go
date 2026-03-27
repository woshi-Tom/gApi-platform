package handler

import (
	"io"
	"net/http"

	"gapi-platform/internal/config"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

// PaymentHandler handles payment-related endpoints
type PaymentHandler struct {
	cfg         *config.Config
	orderRepo   *repository.OrderRepository
	paymentRepo *repository.PaymentRepository
	userRepo    *repository.UserRepository
}

// NewPaymentHandler creates a new payment handler
func NewPaymentHandler(cfg *config.Config, orderRepo *repository.OrderRepository, paymentRepo *repository.PaymentRepository, userRepo *repository.UserRepository) *PaymentHandler {
	return &PaymentHandler{
		cfg:         cfg,
		orderRepo:   orderRepo,
		paymentRepo: paymentRepo,
		userRepo:    userRepo,
	}
}

// CreateAlipay creates an Alipay payment
func (h *PaymentHandler) CreateAlipay(c *gin.Context) {
	if !h.cfg.Payment.Alipay.Enabled {
		response.Fail(c, "PAYMENT_DISABLED", "Alipay payment is not enabled")
		return
	}

	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	userID := c.MustGet("user_id").(uint)

	// Get order
	order, err := h.orderRepo.GetByID(req.OrderID)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	if order.Status != "pending" {
		response.Fail(c, "ORDER_NOT_PENDING", "order is not in pending status")
		return
	}

	// TODO: Integrate with Alipay SDK
	// For now, return a placeholder
	response.Success(c, map[string]interface{}{
		"payment_url": "https://openapi.alipay.com/gateway.do?...",
		"order_no":    order.OrderNo,
		"amount":      order.PayAmount,
	})
}

// CreateWechat creates a WeChat payment
func (h *PaymentHandler) CreateWechat(c *gin.Context) {
	if !h.cfg.Payment.Wechat.Enabled {
		response.Fail(c, "PAYMENT_DISABLED", "WeChat payment is not enabled")
		return
	}

	var req struct {
		OrderID uint `json:"order_id" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	userID := c.MustGet("user_id").(uint)

	// Get order
	order, err := h.orderRepo.GetByID(req.OrderID)
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	// TODO: Integrate with WeChat Pay SDK
	// For now, return a placeholder
	response.Success(c, map[string]interface{}{
		"qr_code":  "data:image/png;base64,...",
		"order_no": order.OrderNo,
		"amount":   order.PayAmount,
	})
}

// AlipayCallback handles Alipay async callback
func (h *PaymentHandler) AlipayCallback(c *gin.Context) {
	// Read callback data
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.String(http.StatusBadRequest, "fail")
		return
	}

	// TODO: Verify Alipay signature
	// TODO: Parse callback data and update order status
	// TODO: Grant quota/VIP to user

	_ = body

	// Return success to Alipay
	c.String(http.StatusOK, "success")
}

// WechatCallback handles WeChat Pay async callback
func (h *PaymentHandler) WechatCallback(c *gin.Context) {
	// Read callback data
	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		c.XML(http.StatusBadRequest, map[string]string{"return_code": "FAIL", "return_msg": "invalid body"})
		return
	}

	// TODO: Verify WeChat signature
	// TODO: Parse callback data and update order status
	// TODO: Grant quota/VIP to user

	_ = body

	// Return success to WeChat
	c.XML(http.StatusOK, map[string]string{"return_code": "SUCCESS", "return_msg": "OK"})
}

package handler

import (
	"fmt"
	"strconv"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

// OrderHandler handles order-related endpoints
type OrderHandler struct {
	orderRepo   *repository.OrderRepository
	userRepo    *repository.UserRepository
	paymentRepo *repository.PaymentRepository
}

// NewOrderHandler creates a new order handler
func NewOrderHandler(orderRepo *repository.OrderRepository, userRepo *repository.UserRepository, paymentRepo *repository.PaymentRepository) *OrderHandler {
	return &OrderHandler{
		orderRepo:   orderRepo,
		userRepo:    userRepo,
		paymentRepo: paymentRepo,
	}
}

// List returns the user's orders
func (h *OrderHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orderType := c.Query("type")
	status := c.Query("status")

	orders, total, err := h.orderRepo.List(page, pageSize, userID, orderType, status, "", "")
	if err != nil {
		response.InternalError(c, "failed to list orders")
		return
	}

	response.Paginated(c, orders, page, pageSize, total)
}

// Create creates a new order
func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		PackageID     uint   `json:"package_id" binding:"required"`
		PackageType   string `json:"package_type" binding:"required"` // recharge|vip
		PaymentMethod string `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	// Generate order number
	orderNo := fmt.Sprintf("ORD%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])

	order := &model.Order{
		UserID:      userID,
		OrderNo:     orderNo,
		OrderType:   req.PackageType,
		PackageID:   &req.PackageID,
		Status:      "pending",
		TotalAmount: 0, // Will be set based on package
		PayAmount:   0,
		ExpireAt:    timePtr(time.Now().Add(30 * time.Minute)),
	}

	if err := h.orderRepo.Create(order); err != nil {
		response.Fail(c, "ORDER_CREATE_FAILED", err.Error())
		return
	}

	// Create payment record
	payment := &model.Payment{
		UserID:        userID,
		OrderID:       order.ID,
		PaymentNo:     fmt.Sprintf("PAY%s%s", time.Now().Format("20060102"), uuid.New().String()[:8]),
		PaymentMethod: req.PaymentMethod,
		Amount:        order.PayAmount,
		Status:        "pending",
	}

	h.paymentRepo.Create(payment)

	response.Created(c, map[string]interface{}{
		"order":   order,
		"payment": payment,
	})
}

// GetByID returns an order by ID
func (h *OrderHandler) GetByID(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	idStr := c.Param("id")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid order id")
		return
	}

	order, err := h.orderRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "order not found")
		return
	}

	if order.UserID != userID {
		response.Forbidden(c, "not your order")
		return
	}

	response.Success(c, order)
}

func timePtr(t time.Time) *time.Time {
	return &t
}

package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"strconv"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const (
	OrderDefaultExpireHours     = 4
	IdempotencyKeyExpireMinutes = 10
)

type OrderHandler struct {
	orderRepo    *repository.OrderRepository
	userRepo     *repository.UserRepository
	paymentRepo  *repository.PaymentRepository
	vipRepo      *repository.VIPPackageRepository
	rechargeRepo *repository.RechargePackageRepository
	idempRepo    *repository.IdempotencyRepository
}

func NewOrderHandler(
	orderRepo *repository.OrderRepository,
	userRepo *repository.UserRepository,
	paymentRepo *repository.PaymentRepository,
	vipRepo *repository.VIPPackageRepository,
	rechargeRepo *repository.RechargePackageRepository,
	idempRepo *repository.IdempotencyRepository,
) *OrderHandler {
	return &OrderHandler{
		orderRepo:    orderRepo,
		userRepo:     userRepo,
		paymentRepo:  paymentRepo,
		vipRepo:      vipRepo,
		rechargeRepo: rechargeRepo,
		idempRepo:    idempRepo,
	}
}

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

	response.Success(c, gin.H{
		"list": orders,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

func (h *OrderHandler) Create(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	idempotencyKey := c.GetHeader("X-Idempotency-Key")

	var req struct {
		PackageID     uint   `json:"package_id" binding:"required"`
		PackageType   string `json:"package_type" binding:"required"`
		PaymentMethod string `json:"payment_method" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if idempotencyKey != "" {
		existing, err := h.idempRepo.GetByKey(idempotencyKey)
		if err == nil && existing != nil && existing.OrderID != nil && *existing.OrderID != 0 {
			order, _ := h.orderRepo.GetByID(*existing.OrderID)
			if order != nil {
				response.Created(c, map[string]interface{}{
					"order_no":     order.OrderNo,
					"package_name": order.PackageName,
					"amount":       order.PayAmount,
					"order_id":     order.ID,
					"status":       order.Status,
					"expires_at":   order.ExpiresAt,
					"idempotent":   true,
				})
				return
			}
		}
	}

	var pkgName string
	var price float64

	switch req.PackageType {
	case "vip":
		pkg, err := h.vipRepo.GetByID(req.PackageID)
		if err != nil {
			response.Fail(c, "PACKAGE_NOT_FOUND", "VIP套餐不存在")
			return
		}
		pkgName = pkg.Name
		price = pkg.Price
	case "recharge":
		pkg, err := h.rechargeRepo.GetByID(req.PackageID)
		if err != nil {
			response.Fail(c, "PACKAGE_NOT_FOUND", "充值套餐不存在")
			return
		}
		pkgName = pkg.Name
		price = pkg.Price
	default:
		response.Fail(c, "INVALID_PACKAGE_TYPE", "无效的套餐类型")
		return
	}

	orderNo := fmt.Sprintf("ORD%s%s", time.Now().Format("20060102"), uuid.New().String()[:8])
	now := time.Now()
	expiresAt := now.Add(time.Duration(OrderDefaultExpireHours) * time.Hour)

	order := &model.Order{
		UserID:      userID,
		OrderNo:     orderNo,
		OrderType:   req.PackageType,
		PackageID:   &req.PackageID,
		PackageName: pkgName,
		Status:      model.OrderStatusPending,
		TotalAmount: price,
		PayAmount:   price,
		ExpiresAt:   &expiresAt,
	}

	if err := h.orderRepo.Create(order); err != nil {
		response.Fail(c, "ORDER_CREATE_FAILED", err.Error())
		return
	}

	payment := &model.Payment{
		UserID:        userID,
		OrderID:       order.ID,
		PaymentNo:     fmt.Sprintf("PAY%s%s", time.Now().Format("20060102"), uuid.New().String()[:8]),
		PaymentMethod: req.PaymentMethod,
		Amount:        price,
		Status:        model.PaymentStatusPending,
	}

	if err := h.paymentRepo.Create(payment); err != nil {
		response.Fail(c, "PAYMENT_CREATE_FAILED", err.Error())
		return
	}

	if idempotencyKey != "" {
		h.idempRepo.Create(&model.IdempotencyKey{
			Key:       idempotencyKey,
			UserID:    userID,
			Action:    "create_order",
			OrderID:   &order.ID,
			OrderNo:   orderNo,
			CreatedAt: now,
			ExpiresAt: now.Add(time.Duration(IdempotencyKeyExpireMinutes) * time.Minute),
		})
	}

	response.Created(c, map[string]interface{}{
		"order_no":     orderNo,
		"package_name": pkgName,
		"amount":       price,
		"order_id":     order.ID,
		"expires_at":   expiresAt.Format(time.RFC3339),
	})
}

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

func (h *OrderHandler) GetByOrderNo(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	orderNo := c.Param("order_no")

	order, err := h.orderRepo.GetByOrderNo(orderNo)
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

func generateIdempotencyKey(userID uint, action string, data string) string {
	raw := fmt.Sprintf("%d:%s:%s:%d", userID, action, data, time.Now().Unix()/60)
	hash := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(hash[:16])
}

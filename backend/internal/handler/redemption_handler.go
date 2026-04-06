package handler

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"gorm.io/gorm"

	"github.com/gin-gonic/gin"
)

var (
	ErrAlreadyRedeemed = errors.New("already redeemed")
	ErrCodeExhausted   = errors.New("code exhausted")
)

type RedemptionHandler struct {
	db        *gorm.DB
	userRepo  *repository.UserRepository
	auditRepo *repository.AuditRepository
	cache     *service.RedemptionCacheService
}

func NewRedemptionHandler(db *gorm.DB, userRepo *repository.UserRepository, auditRepo *repository.AuditRepository, cache *service.RedemptionCacheService) *RedemptionHandler {
	return &RedemptionHandler{
		db:        db,
		userRepo:  userRepo,
		auditRepo: auditRepo,
		cache:     cache,
	}
}

type CreateCodeRequest struct {
	Prefix     string  `json:"prefix" binding:"required"`
	Count      int     `json:"count" binding:"required,min=1,max=1000"`
	CodeType   string  `json:"code_type" binding:"required"`
	Quota      int64   `json:"quota"`
	QuotaType  string  `json:"quota_type"`
	VIPDays    int     `json:"vip_days"`
	MaxUses    int     `json:"max_uses"`
	ValidFrom  *string `json:"valid_from"`
	ValidUntil *string `json:"valid_until"`
}

func (h *RedemptionHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	codeType := c.Query("code_type")
	status := c.Query("status")
	batchID := c.Query("batch_id")

	query := h.db.Model(&model.RedemptionCode{}).Where("deleted_at IS NULL")

	if codeType != "" {
		query = query.Where("code_type = ?", codeType)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}
	if batchID != "" {
		query = query.Where("batch_id = ?", batchID)
	}

	var total int64
	query.Count(&total)

	var codes []model.RedemptionCode
	offset := (page - 1) * pageSize
	if err := query.Order("created_at DESC").Offset(offset).Limit(pageSize).Find(&codes).Error; err != nil {
		response.InternalError(c, "failed to list codes")
		return
	}

	response.Paginated(c, codes, page, pageSize, total)
}

func (h *RedemptionHandler) Create(c *gin.Context) {
	var req CreateCodeRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.MaxUses <= 0 {
		req.MaxUses = 1
	}
	if req.QuotaType == "" {
		req.QuotaType = model.QuotaTypePermanent
	}

	batchID := model.GenerateCode("BATCH")
	adminID := getAdminIDFromContext(c)

	var codes []model.RedemptionCode

	for i := 0; i < req.Count; i++ {
		code := model.GenerateCode(req.Prefix)

		redemptionCode := model.RedemptionCode{
			TenantID:  1,
			Code:      code,
			CodeType:  req.CodeType,
			Quota:     req.Quota,
			QuotaType: req.QuotaType,
			VIPDays:   req.VIPDays,
			MaxUses:   req.MaxUses,
			BatchID:   batchID,
			Status:    model.RedemptionStatusActive,
			CreatedBy: adminID,
		}

		if req.ValidFrom != nil {
			t, _ := time.Parse("2006-01-02", *req.ValidFrom)
			redemptionCode.ValidFrom = &t
		}
		if req.ValidUntil != nil {
			t, _ := time.Parse("2006-01-02", *req.ValidUntil)
			redemptionCode.ValidUntil = &t
		}

		codes = append(codes, redemptionCode)
	}

	if err := h.db.Create(&codes).Error; err != nil {
		response.InternalError(c, "failed to create codes")
		return
	}

	response.Created(c, gin.H{
		"batch_id": batchID,
		"codes":    codes,
		"count":    len(codes),
	})
}

func (h *RedemptionHandler) Disable(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid id")
		return
	}

	result := h.db.Model(&model.RedemptionCode{}).Where("id = ?", id).Update("status", model.RedemptionStatusDisabled)
	if result.Error != nil {
		response.InternalError(c, "failed to disable code")
		return
	}

	response.Success(c, "code disabled")
}

func (h *RedemptionHandler) GetUsage(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid id")
		return
	}

	var usage []model.RedemptionUsage
	if err := h.db.Where("code_id = ?", id).Find(&usage).Error; err != nil {
		response.InternalError(c, "failed to get usage")
		return
	}

	response.Success(c, usage)
}

type RedeemRequest struct {
	Code string `json:"code" binding:"required"`
}

func (h *RedemptionHandler) Redeem(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "please login first")
		return
	}

	var req RedeemRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	var redemptionCode model.RedemptionCode
	if err := h.db.Where("code = ? AND deleted_at IS NULL", req.Code).First(&redemptionCode).Error; err != nil {
		response.Fail(c, "CODE_NOT_FOUND", "redemption code not found")
		return
	}

	if !redemptionCode.IsValid() {
		response.Fail(c, "CODE_INVALID", "redemption code is invalid or expired")
		return
	}

	ctx := context.Background()

	if h.cache != nil {
		redeemed, _ := h.cache.IsRedeemed(ctx, redemptionCode.ID, userID)
		if redeemed {
			response.Fail(c, "ALREADY_REDEEMED", "you have already used this code")
			return
		}
	}

	codeID := redemptionCode.ID
	lockToken := ""
	lockAcquired := false

	if h.cache != nil {
		token, err := h.cache.AcquireLock(ctx, codeID, userID, 10*time.Second)
		if err == nil && token != "" {
			lockToken = token
			lockAcquired = true
		}
	}

	if lockAcquired {
		defer h.cache.ReleaseLock(ctx, codeID, userID, lockToken)
	}

	redeemed, _ := h.cache.IsRedeemed(ctx, codeID, userID)
	if redeemed {
		response.Fail(c, "ALREADY_REDEEMED", "you have already used this code")
		return
	}

	var quotaGranted int64
	var vipGranted bool
	var vipDays int
	var user model.User

	err := h.db.Transaction(func(tx *gorm.DB) error {
		if err := tx.First(&user, userID).Error; err != nil {
			return err
		}

		switch redemptionCode.CodeType {
		case model.RedemptionCodeTypeRecharge, model.RedemptionCodeTypeQuota:
			quotaGranted = redemptionCode.Quota
			if redemptionCode.QuotaType == model.QuotaTypeVIP {
				user.VIPQuota += quotaGranted
			} else {
				user.FreeQuota += quotaGranted
			}
			vipGranted = false

		case model.RedemptionCodeTypeVIP:
			vipGranted = true
			vipDays = redemptionCode.VIPDays
			if redemptionCode.IsPermanent {
				user.VIPExpiredAt = nil
			} else {
				now := time.Now()
				if user.VIPExpiredAt != nil && user.VIPExpiredAt.After(now) {
					newTime := user.VIPExpiredAt.Add(time.Duration(vipDays) * 24 * time.Hour)
					user.VIPExpiredAt = &newTime
				} else {
					newTime := now.Add(time.Duration(vipDays) * 24 * time.Hour)
					user.VIPExpiredAt = &newTime
				}
			}
			if user.Level == "free" {
				user.Level = "vip"
			}
			if redemptionCode.Quota > 0 {
				user.VIPQuota += redemptionCode.Quota
			}
		}

		if err := tx.Save(&user).Error; err != nil {
			return err
		}

		redemptionCode.UsedCount++
		if redemptionCode.UsedCount >= redemptionCode.MaxUses {
			redemptionCode.Status = model.RedemptionStatusUsed
		}
		if err := tx.Save(&redemptionCode).Error; err != nil {
			return err
		}

		usage := model.RedemptionUsage{
			CodeID:       redemptionCode.ID,
			UserID:       userID,
			QuotaGranted: quotaGranted,
			VIPGranted:   vipGranted,
			VIPDays:      vipDays,
			IPAddress:    c.ClientIP(),
			UserAgent:    c.GetHeader("User-Agent"),
		}
		if err := tx.Create(&usage).Error; err != nil {
			if isUniqueConstraintError(err) {
				return ErrAlreadyRedeemed
			}
			return err
		}

		return nil
	})

	if err != nil {
		if errors.Is(err, ErrAlreadyRedeemed) {
			if h.cache != nil {
				h.cache.MarkRedeemed(ctx, codeID, userID)
			}
			response.Fail(c, "ALREADY_REDEEMED", "you have already used this code")
			return
		}
		response.Fail(c, "REDEMPTION_FAILED", "failed to apply redemption")
		return
	}

	if h.cache != nil {
		h.cache.MarkRedeemed(ctx, codeID, userID)
	}

	response.Success(c, gin.H{
		"message":       "redemption successful",
		"quota_granted": quotaGranted,
		"vip_granted":   vipGranted,
		"vip_days":      vipDays,
	})
}

func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return contains(errStr, "unique constraint") ||
		contains(errStr, "duplicate key") ||
		contains(errStr, "23505")
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > 0 && containsHelper(s, substr))
}

func containsHelper(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func (h *RedemptionHandler) GetUserHistory(c *gin.Context) {
	userID := getUserIDFromContext(c)
	if userID == 0 {
		response.Unauthorized(c, "please login first")
		return
	}

	var usage []model.RedemptionUsage
	if err := h.db.Where("user_id = ?", userID).Order("redeemed_at DESC").Find(&usage).Error; err != nil {
		response.InternalError(c, "failed to get history")
		return
	}

	response.Success(c, usage)
}

func getAdminIDFromContext(c *gin.Context) *uint {
	if id, exists := c.Get("admin_id"); exists {
		if v, ok := id.(uint); ok {
			return &v
		}
	}
	return nil
}

func getUserIDFromContext(c *gin.Context) uint {
	if id, exists := c.Get("user_id"); exists {
		if v, ok := id.(uint); ok {
			return v
		}
	}
	return 0
}

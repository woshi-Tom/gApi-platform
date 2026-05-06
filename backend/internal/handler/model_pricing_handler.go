package handler

import (
	"strconv"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type ModelPricingHandler struct {
	svc *service.ModelGroupService
}

func NewModelPricingHandler(svc *service.ModelGroupService) *ModelPricingHandler {
	return &ModelPricingHandler{svc: svc}
}

func (h *ModelPricingHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	provider := c.Query("provider")

	pricings, total, err := h.svc.ListPricing(page, pageSize, provider, nil)
	if err != nil {
		response.InternalError(c, "获取定价列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": pricings,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *ModelPricingHandler) Create(c *gin.Context) {
	var req struct {
		Model         string   `json:"model" binding:"required"`
		Provider      string   `json:"provider"`
		DisplayName   string   `json:"display_name"`
		PriceInput    float64  `json:"price_input"`
		PriceOutput   float64  `json:"price_output"`
		AbilityTypes  []string `json:"ability_types"`
		ContextLength int      `json:"context_length"`
		MaxOutput     int      `json:"max_output"`
		GroupID       *uint    `json:"group_id"`
		IsEnabled     bool     `json:"is_enabled"`
		IsFeatured    bool     `json:"is_featured"`
		SortOrder     int      `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	p := &model.ModelPricing{
		Model:         req.Model,
		Provider:      req.Provider,
		DisplayName:   req.DisplayName,
		PriceInput:    req.PriceInput,
		PriceOutput:   req.PriceOutput,
		ContextLength: req.ContextLength,
		MaxOutput:     req.MaxOutput,
		GroupID:       req.GroupID,
		IsEnabled:     req.IsEnabled,
		IsFeatured:    req.IsFeatured,
		SortOrder:     req.SortOrder,
	}
	if len(req.AbilityTypes) > 0 {
		p.SetAbilityTypes(req.AbilityTypes)
	} else {
		p.SetAbilityTypes([]string{"chat"})
	}

	if err := h.svc.CreatePricing(p); err != nil {
		response.InternalError(c, "创建定价失败")
		return
	}

	response.Success(c, p)
}

func (h *ModelPricingHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的定价ID")
		return
	}

	existing, err := h.svc.GetPricing(uint(id))
	if err != nil {
		response.NotFound(c, "定价记录不存在")
		return
	}

	var req struct {
		Provider      string   `json:"provider"`
		DisplayName   string   `json:"display_name"`
		PriceInput    float64  `json:"price_input"`
		PriceOutput   float64  `json:"price_output"`
		AbilityTypes  []string `json:"ability_types"`
		ContextLength int      `json:"context_length"`
		MaxOutput     int      `json:"max_output"`
		GroupID       *uint    `json:"group_id"`
		IsEnabled     *bool    `json:"is_enabled"`
		IsFeatured    *bool    `json:"is_featured"`
		SortOrder     int      `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	if req.Provider != "" {
		existing.Provider = req.Provider
	}
	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	if req.PriceInput > 0 {
		existing.PriceInput = req.PriceInput
	}
	if req.PriceOutput > 0 {
		existing.PriceOutput = req.PriceOutput
	}
	if len(req.AbilityTypes) > 0 {
		existing.SetAbilityTypes(req.AbilityTypes)
	}
	if req.ContextLength > 0 {
		existing.ContextLength = req.ContextLength
	}
	if req.MaxOutput > 0 {
		existing.MaxOutput = req.MaxOutput
	}
	if req.GroupID != nil {
		existing.GroupID = req.GroupID
	}
	if req.IsEnabled != nil {
		existing.IsEnabled = *req.IsEnabled
	}
	if req.IsFeatured != nil {
		existing.IsFeatured = *req.IsFeatured
	}
	existing.SortOrder = req.SortOrder

	if err := h.svc.UpdatePricing(existing); err != nil {
		response.InternalError(c, "更新定价失败")
		return
	}

	response.Success(c, existing)
}

func (h *ModelPricingHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的定价ID")
		return
	}

	if err := h.svc.DeletePricing(uint(id)); err != nil {
		response.InternalError(c, "删除定价失败")
		return
	}

	response.Success(c, nil)
}

func (h *ModelPricingHandler) ListAll(c *gin.Context) {
	pricings, err := h.svc.ListAllPricing()
	if err != nil {
		response.InternalError(c, "获取定价列表失败")
		return
	}
	response.Success(c, pricings)
}

func (h *ModelPricingHandler) GetByModel(c *gin.Context) {
	modelName := c.Param("model")
	pricing, err := h.svc.GetPricingByModel(modelName)
	if err != nil {
		response.NotFound(c, "模型定价不存在")
		return
	}
	response.Success(c, pricing)
}

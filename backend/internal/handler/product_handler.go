package handler

import (
	"strconv"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type ProductHandler struct {
	vipRepo      *repository.VIPPackageRepository
	rechargeRepo *repository.RechargePackageRepository
}

func NewProductHandler(vipRepo *repository.VIPPackageRepository, rechargeRepo *repository.RechargePackageRepository) *ProductHandler {
	return &ProductHandler{
		vipRepo:      vipRepo,
		rechargeRepo: rechargeRepo,
	}
}

func (h *ProductHandler) List(c *gin.Context) {
	productType := c.Query("type")

	var products []model.Product

	if productType == "" || productType == "vip" {
		vipPackages, err := h.vipRepo.List()
		if err == nil {
			for _, pkg := range vipPackages {
				products = append(products, model.Product{
					ID:            pkg.ID,
					Name:          pkg.Name,
					Description:   pkg.Description,
					ProductType:   "vip",
					Price:         pkg.Price,
					OriginalPrice: nil,
					VIPQuota:      pkg.Quota,
					VIPDays:       pkg.DurationDays,
					RPMLimit:      pkg.RPMLimit,
					TPMLimit:      pkg.TPMLimit,
					SortOrder:     pkg.SortOrder,
					IsRecommended: pkg.IsRecommended,
					IsHot:         pkg.IsPopular,
					Status:        pkg.Status,
					CreatedAt:     pkg.CreatedAt,
				})
			}
		}
	}

	if productType == "" || productType == "recharge" {
		rechargePackages, err := h.rechargeRepo.List()
		if err == nil {
			for _, pkg := range rechargePackages {
				products = append(products, model.Product{
					ID:            pkg.ID,
					Name:          pkg.Name,
					Description:   pkg.Description,
					ProductType:   "recharge",
					Price:         pkg.Price,
					OriginalPrice: nil,
					Quota:         pkg.Quota,
					BonusQuota:    pkg.BonusQuota,
					SortOrder:     pkg.SortOrder,
					IsRecommended: pkg.IsRecommended,
					IsHot:         pkg.IsPopular,
					Status:        pkg.Status,
					CreatedAt:     pkg.CreatedAt,
				})
			}
		}
	}

	response.Success(c, products)
}

func (h *ProductHandler) GetByID(c *gin.Context) {
	idStr := c.Param("id")
	productType := c.Query("type")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid product id")
		return
	}

	if productType == "vip" {
		pkg, err := h.vipRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		response.Success(c, pkg)
		return
	}

	if productType == "recharge" {
		pkg, err := h.rechargeRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		response.Success(c, pkg)
		return
	}

	pkg, err := h.vipRepo.GetByID(uint(id))
	if err == nil {
		response.Success(c, pkg)
		return
	}

	pkg2, err := h.rechargeRepo.GetByID(uint(id))
	if err == nil {
		response.Success(c, pkg2)
		return
	}

	response.NotFound(c, "product not found")
}

func (h *ProductHandler) Enable(c *gin.Context) {
	idStr := c.Param("id")
	productType := c.Query("type")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid product id")
		return
	}

	if productType == "vip" {
		pkg, err := h.vipRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		pkg.Status = "active"
		if err := h.vipRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, map[string]interface{}{"id": id, "status": "active"})
		return
	}

	if productType == "recharge" {
		pkg, err := h.rechargeRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		pkg.Status = "active"
		if err := h.rechargeRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, map[string]interface{}{"id": id, "status": "active"})
		return
	}

	response.Fail(c, "INVALID_PRODUCT_TYPE", "product type must be vip or recharge")
}

func (h *ProductHandler) Disable(c *gin.Context) {
	idStr := c.Param("id")
	productType := c.Query("type")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid product id")
		return
	}

	if productType == "vip" {
		pkg, err := h.vipRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		pkg.Status = "inactive"
		if err := h.vipRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, map[string]interface{}{"id": id, "status": "inactive"})
		return
	}

	if productType == "recharge" {
		pkg, err := h.rechargeRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		pkg.Status = "inactive"
		if err := h.rechargeRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, map[string]interface{}{"id": id, "status": "inactive"})
		return
	}

	response.Fail(c, "INVALID_PRODUCT_TYPE", "product type must be vip or recharge")
}

type CreateProductRequest struct {
	ProductType     string   `json:"product_type" binding:"required"`
	Name            string   `json:"name" binding:"required"`
	Description     string   `json:"description"`
	Price           float64  `json:"price" binding:"required,gt=0"`
	OriginalPrice   *float64 `json:"original_price"`
	Quota           int64    `json:"quota"`
	VIPDays         int      `json:"vip_days"`
	BonusQuota      int64    `json:"bonus_quota"`
	RPMLimit        int      `json:"rpm_limit"`
	TPMLimit        int      `json:"tpm_limit"`
	ConcurrentLimit int      `json:"concurrent_limit"`
	SortOrder       int      `json:"sort_order"`
	IsRecommended   bool     `json:"is_recommended"`
	IsPopular       bool     `json:"is_popular"`
	Status          string   `json:"status"`
}

type UpdateProductRequest struct {
	Name            string   `json:"name"`
	Description     string   `json:"description"`
	Price           *float64 `json:"price"`
	OriginalPrice   *float64 `json:"original_price"`
	Quota           *int64   `json:"quota"`
	VIPDays         *int     `json:"vip_days"`
	BonusQuota      *int64   `json:"bonus_quota"`
	RPMLimit        *int     `json:"rpm_limit"`
	TPMLimit        *int     `json:"tpm_limit"`
	ConcurrentLimit *int     `json:"concurrent_limit"`
	SortOrder       *int     `json:"sort_order"`
	IsRecommended   *bool    `json:"is_recommended"`
	IsPopular       *bool    `json:"is_popular"`
	Status          string   `json:"status"`
}

func (h *ProductHandler) Create(c *gin.Context) {
	var req CreateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.ProductType == "vip" {
		pkg := &model.VIPPackage{
			Name:            req.Name,
			Description:     req.Description,
			Price:           req.Price,
			OriginalPrice:   req.OriginalPrice,
			DurationDays:    req.VIPDays,
			Quota:           req.Quota,
			RPMLimit:        req.RPMLimit,
			TPMLimit:        req.TPMLimit,
			ConcurrentLimit: req.ConcurrentLimit,
			SortOrder:       req.SortOrder,
			IsRecommended:   req.IsRecommended,
			IsPopular:       req.IsPopular,
			Status:          req.Status,
			IsVisible:       true,
		}
		if pkg.DurationDays == 0 {
			pkg.DurationDays = 30
		}
		if pkg.Status == "" {
			pkg.Status = "active"
		}
		if err := h.vipRepo.Create(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Created(c, pkg)
		return
	}

	if req.ProductType == "recharge" {
		pkg := &model.RechargePackage{
			Name:          req.Name,
			Description:   req.Description,
			Price:         req.Price,
			OriginalPrice: req.OriginalPrice,
			Quota:         req.Quota,
			BonusQuota:    req.BonusQuota,
			SortOrder:     req.SortOrder,
			IsRecommended: req.IsRecommended,
			IsPopular:     req.IsPopular,
			Status:        req.Status,
			IsVisible:     true,
		}
		if pkg.Status == "" {
			pkg.Status = "active"
		}
		if err := h.rechargeRepo.Create(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Created(c, pkg)
		return
	}

	response.Fail(c, "INVALID_PRODUCT_TYPE", "product type must be vip or recharge")
}

func (h *ProductHandler) Update(c *gin.Context) {
	idStr := c.Param("id")
	productType := c.Query("type")

	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid product id")
		return
	}

	var req UpdateProductRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if productType == "vip" {
		pkg, err := h.vipRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		if req.Name != "" {
			pkg.Name = req.Name
		}
		if req.Description != "" {
			pkg.Description = req.Description
		}
		if req.Price != nil {
			pkg.Price = *req.Price
		}
		if req.OriginalPrice != nil {
			pkg.OriginalPrice = req.OriginalPrice
		}
		if req.VIPDays != nil {
			pkg.DurationDays = *req.VIPDays
		}
		if req.Quota != nil {
			pkg.Quota = *req.Quota
		}
		if req.RPMLimit != nil {
			pkg.RPMLimit = *req.RPMLimit
		}
		if req.TPMLimit != nil {
			pkg.TPMLimit = *req.TPMLimit
		}
		if req.ConcurrentLimit != nil {
			pkg.ConcurrentLimit = *req.ConcurrentLimit
		}
		if req.SortOrder != nil {
			pkg.SortOrder = *req.SortOrder
		}
		if req.IsRecommended != nil {
			pkg.IsRecommended = *req.IsRecommended
		}
		if req.IsPopular != nil {
			pkg.IsPopular = *req.IsPopular
		}
		if req.Status != "" {
			pkg.Status = req.Status
		}
		if err := h.vipRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, pkg)
		return
	}

	if productType == "recharge" {
		pkg, err := h.rechargeRepo.GetByID(uint(id))
		if err != nil {
			response.NotFound(c, "product not found")
			return
		}
		if req.Name != "" {
			pkg.Name = req.Name
		}
		if req.Description != "" {
			pkg.Description = req.Description
		}
		if req.Price != nil {
			pkg.Price = *req.Price
		}
		if req.OriginalPrice != nil {
			pkg.OriginalPrice = req.OriginalPrice
		}
		if req.Quota != nil {
			pkg.Quota = *req.Quota
		}
		if req.BonusQuota != nil {
			pkg.BonusQuota = *req.BonusQuota
		}
		if req.SortOrder != nil {
			pkg.SortOrder = *req.SortOrder
		}
		if req.IsRecommended != nil {
			pkg.IsRecommended = *req.IsRecommended
		}
		if req.IsPopular != nil {
			pkg.IsPopular = *req.IsPopular
		}
		if req.Status != "" {
			pkg.Status = req.Status
		}
		if err := h.rechargeRepo.Update(pkg); err != nil {
			response.Fail(c, "DB_ERROR", err.Error())
			return
		}
		response.Success(c, pkg)
		return
	}

	response.Fail(c, "INVALID_PRODUCT_TYPE", "product type must be vip or recharge")
}

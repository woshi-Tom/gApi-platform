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
		response.Success(c, map[string]interface{}{"id": id, "status": "inactive"})
		return
	}

	response.Fail(c, "INVALID_PRODUCT_TYPE", "product type must be vip or recharge")
}
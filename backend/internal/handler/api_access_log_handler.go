package handler

import (
	"strconv"

	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

type APIAccessLogHandler struct {
	apiAccessLogRepo *repository.APIAccessLogRepository
}

func NewAPIAccessLogHandler(apiAccessLogRepo *repository.APIAccessLogRepository) *APIAccessLogHandler {
	return &APIAccessLogHandler{apiAccessLogRepo: apiAccessLogRepo}
}

func (h *APIAccessLogHandler) List(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	logs, total, err := h.apiAccessLogRepo.ListByUser(userID, page, pageSize)
	if err != nil {
		c.JSON(200, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "DB_ERROR",
				"message": "Failed to fetch API logs",
			},
		})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"data":    logs,
		"pagination": gin.H{
			"page":        page,
			"page_size":   pageSize,
			"total":       total,
			"total_pages": (total + int64(pageSize) - 1) / int64(pageSize),
		},
	})
}

package handler

import (
	"strconv"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type ModelGroupHandler struct {
	svc *service.ModelGroupService
}

func NewModelGroupHandler(svc *service.ModelGroupService) *ModelGroupHandler {
	return &ModelGroupHandler{svc: svc}
}

func (h *ModelGroupHandler) List(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	status := c.Query("status")

	groups, total, err := h.svc.ListGroups(page, pageSize, status)
	if err != nil {
		response.InternalError(c, "获取分组列表失败")
		return
	}

	response.Success(c, gin.H{
		"list": groups,
		"pagination": gin.H{
			"total":     total,
			"page":      page,
			"page_size": pageSize,
		},
	})
}

func (h *ModelGroupHandler) Create(c *gin.Context) {
	var req struct {
		Name        string `json:"name" binding:"required"`
		DisplayName string `json:"display_name" binding:"required"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	g := &model.ModelGroup{
		Name:        req.Name,
		DisplayName: req.DisplayName,
		Description: req.Description,
		SortOrder:   req.SortOrder,
		Status:      "active",
	}
	if err := h.svc.CreateGroup(g); err != nil {
		response.InternalError(c, "创建分组失败")
		return
	}

	response.Success(c, g)
}

func (h *ModelGroupHandler) Update(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的分组ID")
		return
	}

	existing, err := h.svc.GetGroup(uint(id))
	if err != nil {
		response.NotFound(c, "分组不存在")
		return
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Description string `json:"description"`
		SortOrder   int    `json:"sort_order"`
		Status      string `json:"status"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	if req.DisplayName != "" {
		existing.DisplayName = req.DisplayName
	}
	existing.Description = req.Description
	existing.SortOrder = req.SortOrder
	if req.Status != "" {
		existing.Status = req.Status
	}

	if err := h.svc.UpdateGroup(existing); err != nil {
		response.InternalError(c, "更新分组失败")
		return
	}

	response.Success(c, existing)
}

func (h *ModelGroupHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的分组ID")
		return
	}

	if err := h.svc.DeleteGroup(uint(id)); err != nil {
		response.InternalError(c, "删除分组失败")
		return
	}

	response.Success(c, nil)
}

func (h *ModelGroupHandler) AddChannel(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的分组ID")
		return
	}

	var req struct {
		ChannelID uint `json:"channel_id" binding:"required"`
		Priority  int  `json:"priority"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.svc.AddChannelToGroup(req.ChannelID, uint(groupID), req.Priority); err != nil {
		response.InternalError(c, "关联渠道失败")
		return
	}

	response.Success(c, nil)
}

func (h *ModelGroupHandler) RemoveChannel(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的分组ID")
		return
	}
	channelID, err := strconv.ParseUint(c.Param("cid"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的渠道ID")
		return
	}

	if err := h.svc.RemoveChannelFromGroup(uint(channelID), uint(groupID)); err != nil {
		response.InternalError(c, "取消关联失败")
		return
	}

	response.Success(c, nil)
}

func (h *ModelGroupHandler) GetChannels(c *gin.Context) {
	groupID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的分组ID")
		return
	}

	rels, err := h.svc.GetChannelsByGroup(uint(groupID))
	if err != nil {
		response.InternalError(c, "获取渠道列表失败")
		return
	}

	response.Success(c, rels)
}

func (h *ModelGroupHandler) ListAll(c *gin.Context) {
	groups, err := h.svc.ListAllGroups()
	if err != nil {
		response.InternalError(c, "获取分组列表失败")
		return
	}
	response.Success(c, groups)
}

package handler

import (
	"strconv"

	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type UserGroupHandler struct {
	svc *service.ModelGroupService
}

func NewUserGroupHandler(svc *service.ModelGroupService) *UserGroupHandler {
	return &UserGroupHandler{svc: svc}
}

func (h *UserGroupHandler) GetUserGroups(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的用户ID")
		return
	}

	rels, err := h.svc.GetUserGroups(uint(userID))
	if err != nil {
		response.InternalError(c, "获取用户分组失败")
		return
	}

	response.Success(c, rels)
}

func (h *UserGroupHandler) SetUserGroups(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_ID", "无效的用户ID")
		return
	}

	var req struct {
		GroupIDs []uint `json:"group_ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_REQUEST", err.Error())
		return
	}

	if err := h.svc.SetUserGroups(uint(userID), req.GroupIDs); err != nil {
		response.InternalError(c, "设置用户分组失败")
		return
	}

	response.Success(c, gin.H{"group_ids": req.GroupIDs})
}

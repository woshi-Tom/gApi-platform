package handler

import (
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// UserHandler handles user-related endpoints
type UserHandler struct {
	authService *service.AuthService
	userService *service.UserService
	loginLogRepo *repository.LoginLogRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(authService *service.AuthService, userService *service.UserService, loginLogRepo *repository.LoginLogRepository) *UserHandler {
	return &UserHandler{
		authService: authService,
		userService: userService,
		loginLogRepo: loginLogRepo,
	}
}

// Register handles user registration
func (h *UserHandler) Register(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required,min=3,max=50"`
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	result, err := h.authService.Register(req.Username, req.Email, req.Password)
	if err != nil {
		response.Fail(c, "REGISTER_FAILED", err.Error())
		return
	}

	response.Created(c, result)
}

// Login handles user login
func (h *UserHandler) Login(c *gin.Context) {
	var req struct {
		Email    string `json:"email" binding:"required,email"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	result, err := h.authService.Login(req.Email, req.Password)
	if err != nil {
		if h.loginLogRepo != nil {
			h.loginLogRepo.Create(&model.LoginLog{
				Username:  req.Email,
				LoginType: "user",
				IP:        c.ClientIP(),
				UserAgent: c.Request.UserAgent(),
				Success:   false,
				FailReason: err.Error(),
				CreatedAt: time.Now(),
			})
		}
		response.Fail(c, "LOGIN_FAILED", err.Error())
		return
	}

	if h.loginLogRepo != nil {
		userID := result.User.ID
		h.loginLogRepo.Create(&model.LoginLog{
			UserID:    &userID,
			Username:  result.User.Username,
			LoginType: "user",
			IP:        c.ClientIP(),
			UserAgent: c.Request.UserAgent(),
			Success:   true,
			Token:     result.Token,
			CreatedAt: time.Now(),
		})
	}

	response.Success(c, result)
}

// GetProfile returns the current user's profile
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	// Don't return sensitive fields
	user.PasswordHash = ""
	user.VerifyToken = ""

	response.Success(c, user)
}

// UpdateProfile updates the current user's profile
func (h *UserHandler) UpdateProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		Username string `json:"username"`
		Phone    string `json:"phone"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	if req.Username != "" {
		user.Username = req.Username
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}

	if err := h.userService.Update(user); err != nil {
		response.InternalError(c, "failed to update profile")
		return
	}

	response.SuccessWithMessage(c, nil, "profile updated")
}

// ChangePassword changes the user's password
func (h *UserHandler) ChangePassword(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	err := h.authService.ChangePassword(userID, req.OldPassword, req.NewPassword)
	if err != nil {
		response.Fail(c, "PASSWORD_CHANGE_FAILED", err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "password changed successfully")
}

// GetQuota returns the user's quota information
func (h *UserHandler) GetQuota(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	quota, err := h.userService.GetQuota(userID)
	if err != nil {
		response.InternalError(c, "failed to get quota")
		return
	}

	response.Success(c, quota)
}

// GetVIPStatus returns the user's VIP status
func (h *UserHandler) GetVIPStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	status := map[string]interface{}{
		"level":          user.Level,
		"vip_expired_at": user.VIPExpiredAt,
		"vip_quota":      user.VIPQuota,
		"is_vip":         user.Level == "vip" || user.Level == "enterprise",
	}

	response.Success(c, status)
}

// Ensure model import is used
var _ = model.User{}

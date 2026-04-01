package handler

import (
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

// UserProfile represents user profile response
type UserProfile struct {
	ID            uint       `json:"id"`
	Username      string     `json:"username"`
	Email         string     `json:"email"`
	Phone         string     `json:"phone"`
	Level         string     `json:"level"`
	IsVIP         bool       `json:"is_vip"`
	VIPExpiredAt  *time.Time `json:"v_ip_expired_at"`
	FreeQuota     int64      `json:"free_quota"`
	VIPQuota      int64      `json:"v_ip_quota"`
	FreeExpiredAt *time.Time `json:"free_expired_at"`
	Status        string     `json:"status"`
	LastLoginAt   *time.Time `json:"last_login_at"`
	CreatedAt     time.Time  `json:"created_at"`
}

// UserHandler handles user-related endpoints
type UserHandler struct {
	authService  *service.AuthService
	userService  *service.UserService
	loginLogRepo *repository.LoginLogRepository
}

// NewUserHandler creates a new user handler
func NewUserHandler(authService *service.AuthService, userService *service.UserService, loginLogRepo *repository.LoginLogRepository) *UserHandler {
	return &UserHandler{
		authService:  authService,
		userService:  userService,
		loginLogRepo: loginLogRepo,
	}
}

// Register godoc
// @Summary Register a new user
// @Description Create a new user account
// @Tags user
// @Accept json
// @Produce json
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Router /api/v1/user/register [post]
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

// Login godoc
// @Summary User login
// @Description Authenticate user and return JWT token
// @Tags user
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/login [post]
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
				Username:   req.Email,
				LoginType:  "user",
				IP:         c.ClientIP(),
				UserAgent:  c.Request.UserAgent(),
				Success:    false,
				FailReason: err.Error(),
				CreatedAt:  time.Now(),
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

// GetProfile godoc
// @Summary Get user profile
// @Description Get current user's profile
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/profile [get]
func (h *UserHandler) GetProfile(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	profile := UserProfile{
		ID:            user.ID,
		Username:      user.Username,
		Email:         user.Email,
		Phone:         user.Phone,
		Level:         user.Level,
		IsVIP:         h.isVIPUser(user),
		VIPExpiredAt:  user.VIPExpiredAt,
		FreeQuota:     user.FreeQuota,
		VIPQuota:      user.VIPQuota,
		FreeExpiredAt: user.FreeExpiredAt,
		Status:        user.Status,
		LastLoginAt:   user.LastLoginAt,
		CreatedAt:     user.CreatedAt,
	}

	response.Success(c, profile)
}

func (h *UserHandler) isVIPUser(user *model.User) bool {
	if user == nil {
		return false
	}
	hasLevel := user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"
	if !hasLevel {
		return false
	}
	if user.VIPExpiredAt == nil {
		return true
	}
	return user.VIPExpiredAt.After(time.Now())
}

// UpdateProfile godoc
// @Summary Update user profile
// @Description Update current user's profile
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/profile [put]
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

// ChangePassword godoc
// @Summary Change user password
// @Description Change the current user's password
// @Tags user
// @Accept json
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/change-password [post]
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

// GetQuota godoc
// @Summary Get user quota
// @Description Get current user's quota information
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/quota [get]
func (h *UserHandler) GetQuota(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	quota, err := h.userService.GetQuota(userID)
	if err != nil {
		response.InternalError(c, "failed to get quota")
		return
	}

	response.Success(c, quota)
}

// isValidVIP checks if user has VIP level AND VIP hasn't expired
func isValidVIP(user *model.User) bool {
	if user == nil {
		return false
	}
	// Must have VIP level
	hasLevel := user.Level == "enterprise" ||
		user.Level == "vip_bronze" || user.Level == "vip_silver" || user.Level == "vip_gold"
	if !hasLevel {
		return false
	}
	// Must have valid expiry date in the future
	if user.VIPExpiredAt == nil {
		return false
	}
	return user.VIPExpiredAt.After(time.Now())
}

// GetVIPStatus godoc
// @Summary Get VIP status
// @Description Get current user's VIP membership status
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/vip/status [get]
func (h *UserHandler) GetVIPStatus(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)

	user, err := h.userService.GetByID(userID)
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	isVIP := isValidVIP(user)
	var daysRemaining int
	if isVIP && user.VIPExpiredAt != nil {
		days := time.Until(*user.VIPExpiredAt).Hours() / 24
		if days < 0 {
			daysRemaining = 0
		} else {
			daysRemaining = int(days)
		}
	}

	status := map[string]interface{}{
		"level":          user.Level,
		"vip_expired_at": user.VIPExpiredAt,
		"vip_quota":      user.VIPQuota,
		"is_vip":         isVIP,
		"days_remaining": daysRemaining,
	}

	response.Success(c, status)
}

// GetUsageStats godoc
// @Summary Get usage statistics
// @Description Get user's API usage statistics for dashboard charts (last 7 days)
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/stats/usage [get]
func (h *UserHandler) GetUsageStats(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	db := h.userService.GetDB()

	type DailyUsage struct {
		Date          string `json:"date"`
		TotalCalls    int64  `json:"total_calls"`
		TotalTokens   int64  `json:"total_tokens"`
		AvgResponseMs int64  `json:"avg_response_ms"`
	}

	var usageStats []DailyUsage
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		var calls, tokens, avgMs int64
		db.Model(&model.APIAccessLog{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, dayStart, dayEnd).
			Count(&calls)

		db.Model(&model.APIAccessLog{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, dayStart, dayEnd).
			Select("COALESCE(SUM(total_tokens), 0)").Scan(&tokens)

		db.Model(&model.APIAccessLog{}).
			Where("user_id = ? AND created_at >= ? AND created_at < ?", userID, dayStart, dayEnd).
			Select("COALESCE(AVG(response_time), 0)").Scan(&avgMs)

		usageStats = append(usageStats, DailyUsage{
			Date:          dayStart.Format("01-02"),
			TotalCalls:    calls,
			TotalTokens:   tokens,
			AvgResponseMs: avgMs,
		})
	}

	var totalTokensAll, totalCallsAll int64
	db.Model(&model.APIAccessLog{}).
		Where("user_id = ?", userID).
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokensAll)

	db.Model(&model.APIAccessLog{}).
		Where("user_id = ?", userID).
		Count(&totalCallsAll)

	response.Success(c, gin.H{
		"daily_usage":      usageStats,
		"total_tokens_all": totalTokensAll,
		"total_calls_all":  totalCallsAll,
	})
}

// GetRecentActivities godoc
// @Summary Get recent user activities
// @Description Get recent user activities including orders and API access (last 20 items)
// @Tags user
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/user/activities [get]
func (h *UserHandler) GetRecentActivities(c *gin.Context) {
	userID := c.MustGet("user_id").(uint)
	db := h.userService.GetDB()

	type Activity struct {
		ID          uint      `json:"id"`
		Type        string    `json:"type"` // order|vip|token
		Title       string    `json:"title"`
		Description string    `json:"description"`
		Time        time.Time `json:"time"`
	}

	var activities []Activity

	// Get recent orders (last 10)
	var orders []model.Order
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&orders)

	for _, order := range orders {
		var title, desc string
		switch order.OrderType {
		case "vip":
			title = "开通VIP会员"
			desc = order.PackageName
		case "recharge":
			title = "购买配额"
			desc = order.PackageName
		default:
			title = "订单" + order.OrderNo
			desc = order.PackageName
		}

		if order.Status == model.OrderStatusCompleted {
			title = "✗ " + title
		}

		activities = append(activities, Activity{
			ID:          order.ID,
			Type:        "order",
			Title:       title,
			Description: desc,
			Time:        order.CreatedAt,
		})
	}

	// Get recent API access logs (last 10)
	var apiLogs []model.APIAccessLog
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&apiLogs)

	for _, log := range apiLogs {
		var title string
		if log.StatusCode >= 200 && log.StatusCode < 300 {
			title = "API调用成功"
		} else {
			title = "API调用失败"
		}

		activities = append(activities, Activity{
			ID:          log.ID,
			Type:        "token",
			Title:       title,
			Description: log.Model + " - " + log.Endpoint,
			Time:        log.CreatedAt,
		})
	}

	// Get recent login logs (last 10)
	var loginLogs []model.LoginLog
	db.Where("user_id = ?", userID).
		Order("created_at DESC").
		Limit(10).
		Find(&loginLogs)

	for _, log := range loginLogs {
		var title, desc string
		if log.Success {
			title = "用户登录"
			desc = log.IP
		} else {
			title = "登录失败"
			desc = log.FailReason
		}

		activities = append(activities, Activity{
			ID:          log.ID,
			Type:        "login",
			Title:       title,
			Description: desc,
			Time:        log.CreatedAt,
		})
	}

	// Sort by time descending
	for i := 0; i < len(activities)-1; i++ {
		for j := i + 1; j < len(activities); j++ {
			if activities[j].Time.After(activities[i].Time) {
				activities[i], activities[j] = activities[j], activities[i]
			}
		}
	}

	// Limit to 20 total
	if len(activities) > 20 {
		activities = activities[:20]
	}

	response.Success(c, activities)
}

var _ = model.User{}

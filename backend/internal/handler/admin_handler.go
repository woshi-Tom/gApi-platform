package handler

import (
	"encoding/json"
	"sort"
	"strconv"
	"time"

	"gapi-platform/internal/config"
	"gapi-platform/internal/model"
	"gapi-platform/internal/pkg/crypto"
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/repository"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// AdminHandler handles admin endpoints
type AdminHandler struct {
	authService      *service.AuthService
	userRepo         *repository.UserRepository
	channelSvc       *service.ChannelService
	orderRepo        *repository.OrderRepository
	auditRepo        *repository.AuditRepository
	loginLogRepo     *repository.LoginLogRepository
	apiAccessLogRepo *repository.APIAccessLogRepository
	adminUsers       []config.AdminAccount
}

// NewAdminHandler creates a new admin handler
func NewAdminHandler(
	authService *service.AuthService,
	userRepo *repository.UserRepository,
	channelSvc *service.ChannelService,
	orderRepo *repository.OrderRepository,
	auditRepo *repository.AuditRepository,
	loginLogRepo *repository.LoginLogRepository,
	apiAccessLogRepo *repository.APIAccessLogRepository,
	adminUsers []config.AdminAccount,
) *AdminHandler {
	return &AdminHandler{
		authService:      authService,
		userRepo:         userRepo,
		channelSvc:       channelSvc,
		orderRepo:        orderRepo,
		auditRepo:        auditRepo,
		loginLogRepo:     loginLogRepo,
		apiAccessLogRepo: apiAccessLogRepo,
		adminUsers:       adminUsers,
	}
}

// Login godoc
// @Summary Admin login
// @Description Authenticate admin user
// @Tags admin
// @Accept json
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/login [post]
func (h *AdminHandler) Login(c *gin.Context) {
	var req struct {
		Username string `json:"username" binding:"required"`
		Password string `json:"password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	for _, admin := range h.adminUsers {
		if admin.Username == req.Username && admin.Password == req.Password {
			token, _, err := h.authService.GenerateAdminToken(admin.Username, admin.Role)
			if err != nil {
				response.InternalError(c, "failed to generate token")
				return
			}
			response.Success(c, map[string]interface{}{
				"token":    token,
				"username": admin.Username,
				"role":     admin.Role,
			})
			return
		}
	}

	response.Fail(c, "INVALID_CREDENTIALS", "用户名或密码错误")
}

// ListUsers godoc
// @Summary List users
// @Description Get all users with pagination
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param keyword query string false "Search keyword"
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/users [get]
func (h *AdminHandler) ListUsers(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	level := c.Query("level")
	status := c.Query("status")
	keyword := c.Query("keyword")

	users, total, err := h.userRepo.List(page, pageSize, level, status, keyword)
	if err != nil {
		response.InternalError(c, "failed to list users")
		return
	}

	// Mask sensitive data
	for i := range users {
		users[i].PasswordHash = ""
		users[i].VerifyToken = ""
	}

	response.Paginated(c, users, page, pageSize, total)
}

// UpdateUser updates a user
func (h *AdminHandler) UpdateUser(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid user id")
		return
	}

	user, err := h.userRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "user not found")
		return
	}

	var req struct {
		Status         string `json:"status"`
		Level          string `json:"level"`
		QuotaAdjust    int64  `json:"quota_adjust"`
		VIPQuotaAdjust int64  `json:"vip_quota_adjust"`
		VIPExpiredAt   string `json:"vip_expired_at"`
		DisabledReason string `json:"disabled_reason"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Status != "" {
		user.Status = req.Status
	}
	if req.Level != "" {
		user.Level = req.Level
	}
	if req.QuotaAdjust != 0 {
		user.FreeQuota += req.QuotaAdjust
		if user.FreeQuota < 0 {
			user.FreeQuota = 0
		}
	}
	if req.VIPQuotaAdjust != 0 {
		user.VIPQuota += req.VIPQuotaAdjust
		if user.VIPQuota < 0 {
			user.VIPQuota = 0
		}
	}
	if req.VIPExpiredAt != "" {
		t, err := time.Parse(time.RFC3339, req.VIPExpiredAt)
		if err == nil {
			user.VIPExpiredAt = &t
		}
	}
	if req.DisabledReason != "" {
		user.DisabledReason = req.DisabledReason
	}

	if err := h.userRepo.Update(user); err != nil {
		response.InternalError(c, "failed to update user")
		return
	}

	response.SuccessWithMessage(c, nil, "user updated")
}

// ListChannels returns all channels
func (h *AdminHandler) ListChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	channelType := c.Query("type")
	status := c.Query("status")
	group := c.Query("group")
	keyword := c.Query("keyword")

	channels, total, err := h.channelSvc.List(page, pageSize, channelType, status, group, keyword)
	if err != nil {
		response.InternalError(c, "failed to list channels")
		return
	}

	response.Paginated(c, channels, page, pageSize, total)
}

// CreateChannel creates a new channel
func (h *AdminHandler) CreateChannel(c *gin.Context) {
	var req struct {
		Name     string   `json:"name" binding:"required"`
		Type     string   `json:"type" binding:"required"`
		BaseURL  string   `json:"base_url" binding:"required"`
		APIKey   string   `json:"api_key" binding:"required"`
		Models   []string `json:"models"`
		Weight   int      `json:"weight"`
		Priority int      `json:"priority"`
		Group    string   `json:"group_name"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	encryptedKey, _ := crypto.Encrypt(req.APIKey)
	channel := &model.Channel{
		Name:            req.Name,
		Type:            req.Type,
		BaseURL:         req.BaseURL,
		APIKeyEncrypted: encryptedKey,
		Weight:          req.Weight,
		Priority:        req.Priority,
		GroupName:       req.Group,
		Status:          1,
		IsHealthy:       true,
	}

	if len(req.Models) > 0 {
		b, _ := json.Marshal(req.Models)
		channel.Models = string(b)
	}

	if err := h.channelSvc.Create(channel); err != nil {
		response.Fail(c, "CHANNEL_CREATE_FAILED", err.Error())
		return
	}

	response.Created(c, channel)
}

// UpdateChannel updates a channel
func (h *AdminHandler) UpdateChannel(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid channel id")
		return
	}

	channel, err := h.channelSvc.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "channel not found")
		return
	}

	var req struct {
		Name     string `json:"name"`
		BaseURL  string `json:"base_url"`
		Status   int    `json:"status"`
		Priority int    `json:"priority"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Name != "" {
		channel.Name = req.Name
	}
	if req.BaseURL != "" {
		channel.BaseURL = req.BaseURL
	}
	channel.Status = req.Status
	channel.Priority = req.Priority

	if err := h.channelSvc.Update(channel); err != nil {
		response.InternalError(c, "failed to update channel")
		return
	}

	response.SuccessWithMessage(c, channel, "channel updated")
}

// TestChannel tests a channel
func (h *AdminHandler) TestChannel(c *gin.Context) {
	// Delegate to channel handler
	(&ChannelHandler{channelService: h.channelSvc}).Test(c)
}

// ListOrders returns all orders
func (h *AdminHandler) ListOrders(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	orderType := c.Query("type")
	status := c.Query("status")

	orders, total, err := h.orderRepo.List(page, pageSize, 0, orderType, status, "", "")
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

// GetAuditLogs returns brief audit logs for list view
func (h *AdminHandler) GetAuditLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	actionGroup := c.Query("action_group")
	logType := c.Query("log_type")
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var success *bool
	if s := c.Query("success"); s != "" {
		b := s == "true"
		success = &b
	}

	logs, total, err := h.auditRepo.ListBrief(page, pageSize, 0, actionGroup, logType, startTime, endTime, success)
	if err != nil {
		response.InternalError(c, "failed to list audit logs")
		return
	}

	response.Success(c, gin.H{
		"list": logs,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// GetAuditLogDetail returns full audit log by ID
func (h *AdminHandler) GetAuditLogDetail(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 64)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid log id")
		return
	}

	log, err := h.auditRepo.GetByID(uint(id))
	if err != nil {
		response.NotFound(c, "audit log not found")
		return
	}

	response.Success(c, log)
}

// GetLoginLogs returns login logs
func (h *AdminHandler) GetLoginLogs(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	username := c.Query("username")
	ip := c.Query("ip")
	status := c.Query("status") // success or failed
	startTime := c.Query("start_time")
	endTime := c.Query("end_time")

	var success *bool
	if status != "" {
		b := status == "success"
		success = &b
	}

	logs, total, err := h.loginLogRepo.List(page, pageSize, username, ip, success, startTime, endTime)
	if err != nil {
		response.InternalError(c, "failed to list login logs")
		return
	}

	response.Success(c, gin.H{
		"list": logs,
		"pagination": gin.H{
			"page":      page,
			"page_size": pageSize,
			"total":     total,
		},
	})
}

// GetDashboardStats godoc
// @Summary Get dashboard stats
// @Description Get overview statistics for admin dashboard
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/stats/overview [get]
func (h *AdminHandler) GetDashboardStats(c *gin.Context) {
	db := h.userRepo.GetDB()
	now := time.Now()
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())

	var stats model.DashboardStats

	db.Model(&model.User{}).Count((*int64)(&stats.TotalUsers))

	db.Model(&model.User{}).Where("last_login_at >= ?", today).Count((*int64)(&stats.ActiveUsersToday))

	db.Model(&model.User{}).Where("level IN ?", []string{"enterprise", "vip_bronze", "vip_silver", "vip_gold"}).Count((*int64)(&stats.VIPUsersCount))

	db.Model(&model.Channel{}).Count((*int64)(&stats.TotalChannels))

	db.Model(&model.Channel{}).Where("is_healthy = ?", true).Count((*int64)(&stats.HealthyChannels))

	db.Model(&model.Order{}).Where("created_at >= ?", today).Count((*int64)(&stats.TotalOrdersToday))

	db.Model(&model.Order{}).Where("status = ? AND created_at >= ?", "completed", today).
		Select("COALESCE(SUM(pay_amount), 0)").Scan(&stats.TotalRevenueToday)

	db.Model(&model.UsageLog{}).Where("created_at >= ?", today).
		Select("COALESCE(SUM(total_tokens), 0)").Scan(&stats.TotalQuotaUsedToday)

	response.Success(c, stats)
}

// ChangePassword handles admin password change
func (h *AdminHandler) ChangePassword(c *gin.Context) {
	var req struct {
		OldPassword string `json:"old_password" binding:"required"`
		NewPassword string `json:"new_password" binding:"required,min=8"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	adminID, exists := c.Get("admin_id")
	if !exists {
		response.Unauthorized(c, "unauthorized")
		return
	}

	var admin model.AdminUser
	if err := h.userRepo.GetDB().First(&admin, adminID).Error; err != nil {
		response.NotFound(c, "admin user not found")
		return
	}

	if err := bcrypt.CompareHashAndPassword([]byte(admin.PasswordHash), []byte(req.OldPassword)); err != nil {
		response.Fail(c, "INVALID_PASSWORD", "old password is incorrect")
		return
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		response.InternalError(c, "failed to hash password")
		return
	}

	admin.PasswordHash = string(hashedPassword)
	admin.PasswordChangedAt = time.Now()
	admin.FailedLoginAttempts = 0
	admin.LockedUntil = nil

	if err := h.userRepo.GetDB().Save(&admin).Error; err != nil {
		response.InternalError(c, "failed to update password")
		return
	}

	response.SuccessWithMessage(c, nil, "password changed successfully")
}

// GetStatsTrends godoc
// @Summary Get stats trends
// @Description Get API request trends for the last 7 days
// @Tags admin
// @Produce json
// @Security BearerAuth
// @Success 200 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Router /api/v1/admin/stats/trends [get]
func (h *AdminHandler) GetStatsTrends(c *gin.Context) {
	db := h.userRepo.GetDB()

	// Get last 7 days data
	type DailyStat struct {
		Date         string `json:"date"`
		TotalCalls   int64  `json:"total_calls"`
		SuccessCalls int64  `json:"success_calls"`
		FailedCalls  int64  `json:"failed_calls"`
		TotalTokens  int64  `json:"total_tokens"`
	}

	var stats []DailyStat
	now := time.Now()

	for i := 6; i >= 0; i-- {
		date := now.AddDate(0, 0, -i)
		dayStart := time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
		dayEnd := dayStart.Add(24 * time.Hour)

		var totalCalls, successCalls, failedCalls, totalTokens int64

		db.Model(&model.APIAccessLog{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Count(&totalCalls)

		db.Model(&model.APIAccessLog{}).
			Where("created_at >= ? AND created_at < ? AND status_code < 400", dayStart, dayEnd).
			Count(&successCalls)

		db.Model(&model.APIAccessLog{}).
			Where("created_at >= ? AND created_at < ? AND status_code >= 400", dayStart, dayEnd).
			Count(&failedCalls)

		db.Model(&model.APIAccessLog{}).
			Where("created_at >= ? AND created_at < ?", dayStart, dayEnd).
			Select("COALESCE(SUM(total_tokens), 0)").Scan(&totalTokens)

		stats = append(stats, DailyStat{
			Date:         dayStart.Format("01-02"),
			TotalCalls:   totalCalls,
			SuccessCalls: successCalls,
			FailedCalls:  failedCalls,
			TotalTokens:  totalTokens,
		})
	}

	// Get today's success/failure breakdown
	today := time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	tomorrow := today.Add(24 * time.Hour)

	var successCount, failedCount int64
	db.Model(&model.APIAccessLog{}).
		Where("created_at >= ? AND created_at < ? AND status_code < 400", today, tomorrow).
		Count(&successCount)
	db.Model(&model.APIAccessLog{}).
		Where("created_at >= ? AND created_at < ? AND status_code >= 400", today, tomorrow).
		Count(&failedCount)

	response.Success(c, gin.H{
		"daily_trends": stats,
		"today_breakdown": gin.H{
			"success": successCount,
			"failed":  failedCount,
		},
	})
}

func (h *AdminHandler) getDB() *gorm.DB {
	return h.apiAccessLogRepo.GetDB()
}

func (h *AdminHandler) getStartTime(timeRange string) time.Time {
	now := time.Now()
	switch timeRange {
	case "today":
		return time.Date(now.Year(), now.Month(), now.Day(), 0, 0, 0, 0, now.Location())
	case "week":
		return now.AddDate(0, 0, -7)
	case "month":
		return now.AddDate(0, 0, -30)
	default:
		return now.AddDate(0, 0, -7)
	}
}

func (h *AdminHandler) StatsUserOverview(c *gin.Context) {
	timeRange := c.DefaultQuery("time_range", "week")
	startTime := h.getStartTime(timeRange)
	db := h.getDB()

	var result struct {
		TotalRequests int64
		TotalTokens   int64
		TotalFailed   int64
		FailureRate   float64
		AbnormalUsers int64
		ActiveUsers   int64
	}

	db.Table("api_access_logs").Where("created_at >= ?", startTime).Count(&result.TotalRequests)
	db.Table("api_access_logs").Where("created_at >= ?", startTime).Select("COALESCE(SUM(total_tokens), 0)").Scan(&result.TotalTokens)
	db.Table("api_access_logs").Where("created_at >= ? AND status_code NOT IN (200, 201)", startTime).Count(&result.TotalFailed)

	if result.TotalRequests > 0 {
		result.FailureRate = float64(result.TotalFailed) / float64(result.TotalRequests) * 100
	}

	db.Table("api_access_logs").Where("created_at >= ?", startTime).Distinct("user_id").Count(&result.ActiveUsers)

	var userStats []struct {
		UserID uint
		Total  int64
		Failed int64
	}
	db.Table("api_access_logs").
		Select("user_id, COUNT(*) as total, SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed").
		Where("created_at >= ?", startTime).
		Group("user_id").
		Scan(&userStats)

	for _, u := range userStats {
		if u.Total > 0 && float64(u.Failed)/float64(u.Total) > 0.3 {
			result.AbnormalUsers++
		}
	}

	response.Success(c, result)
}

func (h *AdminHandler) StatsUserRanking(c *gin.Context) {
	metric := c.DefaultQuery("type", "requests")
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	timeRange := c.DefaultQuery("time_range", "week")
	startTime := h.getStartTime(timeRange)
	db := h.getDB()

	var results []map[string]interface{}
	query := `
		SELECT 
			u.id as user_id,
			u.username,
			u.email,
			u.level,
			COALESCE(stats.requests, 0) as requests,
			COALESCE(stats.tokens, 0) as tokens,
			COALESCE(stats.failed, 0) as failed,
			CASE WHEN stats.requests > 0 THEN (stats.failed::float / stats.requests * 100) ELSE 0 END as failure_rate
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				COUNT(*) as requests,
				COALESCE(SUM(total_tokens), 0) as tokens,
				SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed
			FROM api_access_logs
			WHERE created_at >= ?
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE stats.requests > 0 OR stats.requests IS NULL
	`
	db.Raw(query, startTime).Scan(&results)

	getFloat64 := func(val interface{}) float64 {
		switch v := val.(type) {
		case float64:
			return v
		case float32:
			return float64(v)
		case int:
			return float64(v)
		case int64:
			return float64(v)
		case int32:
			return float64(v)
		case uint:
			return float64(v)
		case uint64:
			return float64(v)
		default:
			return 0
		}
	}

	sort.Slice(results, func(i, j int) bool {
		var valI, valJ float64
		switch metric {
		case "tokens":
			valI = getFloat64(results[i]["tokens"])
			valJ = getFloat64(results[j]["tokens"])
		case "failed_rate":
			valI = getFloat64(results[i]["failure_rate"])
			valJ = getFloat64(results[j]["failure_rate"])
		default:
			valI = getFloat64(results[i]["requests"])
			valJ = getFloat64(results[j]["requests"])
		}
		return valI > valJ
	})

	if limit > 0 && len(results) > limit {
		results = results[:limit]
	}

	response.Success(c, results)
}

func (h *AdminHandler) StatsUserList(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))
	sortBy := c.DefaultQuery("sort_by", "failure_rate")
	order := c.DefaultQuery("order", "desc")
	timeRange := c.DefaultQuery("time_range", "week")
	level := c.DefaultQuery("level", "all")
	status := c.DefaultQuery("status", "all")
	startTime := h.getStartTime(timeRange)
	db := h.getDB()

	baseQuery := `
		SELECT 
			u.id as user_id,
			u.username,
			u.email,
			u.level,
			u.status,
			COALESCE(stats.requests, 0) as requests,
			COALESCE(stats.tokens, 0) as tokens,
			COALESCE(stats.failed, 0) as failed,
			CASE WHEN stats.requests > 0 THEN (stats.failed::float / stats.requests * 100) ELSE 0 END as failure_rate
		FROM users u
		LEFT JOIN (
			SELECT 
				user_id,
				COUNT(*) as requests,
				COALESCE(SUM(total_tokens), 0) as tokens,
				SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed
			FROM api_access_logs
			WHERE created_at >= ?
			GROUP BY user_id
		) stats ON u.id = stats.user_id
		WHERE 1=1
	`

	args := []interface{}{startTime}
	if level != "all" {
		baseQuery += " AND u.level = ?"
		args = append(args, level)
	}
	if status == "normal" {
		baseQuery += " AND (stats.requests IS NULL OR stats.failed::float / NULLIF(stats.requests, 0) <= 0.3)"
	} else if status == "abnormal" {
		baseQuery += " AND stats.requests > 0 AND stats.failed::float / stats.requests > 0.3"
	}

	var total int64
	countQuery := "SELECT COUNT(*) FROM (" + baseQuery + ") as t"
	db.Raw(countQuery, args...).Scan(&total)

	if order == "asc" {
		baseQuery += " ORDER BY " + sortBy + " ASC"
	} else {
		baseQuery += " ORDER BY " + sortBy + " DESC"
	}
	baseQuery += " LIMIT ? OFFSET ?"
	args = append(args, pageSize, (page-1)*pageSize)

	var results []map[string]interface{}
	db.Raw(baseQuery, args...).Scan(&results)

	response.Success(c, gin.H{
		"data":       results,
		"pagination": gin.H{"total": total, "page": page, "page_size": pageSize},
	})
}

func (h *AdminHandler) StatsAbnormalUsers(c *gin.Context) {
	threshold := 30.0
	if t, err := strconv.ParseFloat(c.DefaultQuery("threshold", "30"), 64); err == nil {
		threshold = t
	}
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	timeRange := c.DefaultQuery("time_range", "week")
	startTime := h.getStartTime(timeRange)
	db := h.getDB()

	var results []map[string]interface{}
	query := `
		SELECT 
			u.id as user_id,
			u.username,
			u.email,
			u.level,
			stats.requests,
			stats.failed,
			(stats.failed::float / stats.requests * 100) as failure_rate
		FROM users u
		INNER JOIN (
			SELECT 
				user_id,
				COUNT(*) as requests,
				SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed
			FROM api_access_logs
			WHERE created_at >= ?
			GROUP BY user_id
			HAVING SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END)::float / COUNT(*) > ?
		) stats ON u.id = stats.user_id
		ORDER BY failure_rate DESC
	`

	args := []interface{}{startTime, threshold / 100}
	if limit > 0 {
		query += " LIMIT ?"
		args = append(args, limit)
	}

	db.Raw(query, args...).Scan(&results)
	response.Success(c, results)
}

func (h *AdminHandler) StatsUserDetail(c *gin.Context) {
	userID, err := strconv.ParseUint(c.Param("id"), 10, 32)
	if err != nil {
		response.Fail(c, "INVALID_PARAMETER", "invalid user id")
		return
	}
	timeRange := c.DefaultQuery("time_range", "week")
	startTime := h.getStartTime(timeRange)
	db := h.getDB()

	var user struct {
		ID       uint
		Username string
		Email    string
		Level    string
		Status   string
	}
	db.Table("users").Where("id = ?", userID).Scan(&user)

	var stats struct {
		Requests    int64
		Tokens      int64
		Failed      int64
		AvgResponse float64
	}
	statsQuery := `
		SELECT 
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as tokens,
			SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed,
			COALESCE(AVG(response_time), 0) as avg_response
		FROM api_access_logs
		WHERE user_id = ? AND created_at >= ?
	`
	db.Raw(statsQuery, userID, startTime).Scan(&stats)

	failureRate := float64(0)
	if stats.Requests > 0 {
		failureRate = float64(stats.Failed) / float64(stats.Requests) * 100
	}

	var dailyTrends []map[string]interface{}
	dailyQuery := `
		SELECT 
			DATE(created_at) as date,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as tokens,
			SUM(CASE WHEN status_code NOT IN (200, 201) THEN 1 ELSE 0 END) as failed
		FROM api_access_logs
		WHERE user_id = ? AND created_at >= NOW() - INTERVAL '30 days'
		GROUP BY DATE(created_at)
		ORDER BY date
	`
	db.Raw(dailyQuery, userID).Scan(&dailyTrends)

	var modelDist []map[string]interface{}
	modelQuery := `
		SELECT 
			COALESCE(model, 'unknown') as model,
			COUNT(*) as requests,
			COALESCE(SUM(total_tokens), 0) as tokens
		FROM api_access_logs
		WHERE user_id = ? AND created_at >= ?
		GROUP BY model
		ORDER BY tokens DESC
	`
	db.Raw(modelQuery, userID, startTime).Scan(&modelDist)

	var resultDist []map[string]interface{}
	resultQuery := `
		SELECT 
			CASE 
				WHEN status_code IN (200, 201) THEN 'success'
				WHEN status_code = 0 THEN 'timeout'
				ELSE 'failed'
			END as status,
			COUNT(*) as count
		FROM api_access_logs
		WHERE user_id = ? AND created_at >= ?
		GROUP BY status
	`
	db.Raw(resultQuery, userID, startTime).Scan(&resultDist)

	response.Success(c, gin.H{
		"user": gin.H{
			"id":       user.ID,
			"username": user.Username,
			"email":    user.Email,
			"level":    user.Level,
			"status":   user.Status,
		},
		"stats": gin.H{
			"requests":     stats.Requests,
			"tokens":       stats.Tokens,
			"failed":       stats.Failed,
			"failure_rate": failureRate,
			"avg_response": stats.AvgResponse,
		},
		"daily_trends":        dailyTrends,
		"model_distribution":  modelDist,
		"result_distribution": resultDist,
	})
}

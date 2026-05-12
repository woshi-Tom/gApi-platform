package handler

import (
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsSvc   *service.SettingsService
	alipayService *service.AlipayService
}

func NewSettingsHandler(settingsSvc *service.SettingsService, alipayService *service.AlipayService) *SettingsHandler {
	return &SettingsHandler{
		settingsSvc:   settingsSvc,
		alipayService: alipayService,
	}
}

func (h *SettingsHandler) GetSMTPConfig(c *gin.Context) {
	cfg, err := h.settingsSvc.GetSMTPConfig()
	if err != nil {
		response.InternalError(c, "failed to get SMTP config: "+err.Error())
		return
	}

	response.Success(c, cfg)
}

func (h *SettingsHandler) UpdateSMTPConfig(c *gin.Context) {
	var req struct {
		Enabled   bool   `json:"enabled"`
		Host      string `json:"host"`
		Port      int    `json:"port"`
		UseTLS    bool   `json:"use_tls"`
		Username  string `json:"username"`
		Password  string `json:"password"`
		FromName  string `json:"from_name"`
		FromEmail string `json:"from_email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Port == 0 {
		req.Port = 587
	}

	cfg := &service.SMTPConfig{
		Enabled:   req.Enabled,
		Host:      req.Host,
		Port:      req.Port,
		UseTLS:    req.UseTLS,
		Username:  req.Username,
		Password:  req.Password,
		FromName:  req.FromName,
		FromEmail: req.FromEmail,
	}

	if err := h.settingsSvc.UpdateSMTPConfig(cfg); err != nil {
		response.InternalError(c, "failed to update SMTP config: "+err.Error())
		return
	}

	h.settingsSvc.InvalidateCache()
	response.Success(c, nil)
}

func (h *SettingsHandler) TestSMTPConnection(c *gin.Context) {
	var req struct {
		TestEmail string `json:"test_email" binding:"required,email"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if err := h.settingsSvc.TestSMTPConnection(req.TestEmail); err != nil {
		response.Fail(c, "SMTP_TEST_FAILED", err.Error())
		return
	}

	response.Success(c, map[string]string{
		"message": "测试邮件发送成功",
	})
}

func (h *SettingsHandler) GetRegisterSettings(c *gin.Context) {
	settings, err := h.settingsSvc.GetRegisterSettings()
	if err != nil {
		response.InternalError(c, "failed to get register settings: "+err.Error())
		return
	}
	response.Success(c, settings)
}

func (h *SettingsHandler) UpdateRegisterSettings(c *gin.Context) {
	var req struct {
		AllowRegister       bool    `json:"allow_register"`
		RequireEmailVerify  *bool   `json:"require_email_verify"`
		EnableCaptcha       bool    `json:"enable_captcha"`
		NewUserQuota        int     `json:"new_user_quota"`
		TrialVIPDays        int     `json:"trial_vip_days"`
		AllowedDomains      *string `json:"allowed_domains"`
		MaxAccountsPerIP    *int    `json:"max_accounts_per_ip"`
		MinPasswordLength   *int    `json:"min_password_length"`
		SignupRewardType    *string `json:"signup_reward_type"`
		SignupRewardAmount  *int64  `json:"signup_reward_amount"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.RequireEmailVerify != nil && *req.RequireEmailVerify {
		if !h.settingsSvc.IsSMTPEnabled() {
			response.Fail(c, "SMTP_NOT_CONFIGURED", "请先在邮箱设置中配置并启用邮箱服务")
			return
		}
	}

	current, err := h.settingsSvc.GetRegisterSettings()
	if err != nil {
		response.InternalError(c, "failed to get current register settings: "+err.Error())
		return
	}
	settings := &service.RegisterSettings{
		AllowRegister:      req.AllowRegister,
		RequireEmailVerify: req.RequireEmailVerify,
		EnableCaptcha:      req.EnableCaptcha,
		NewUserQuota:       req.NewUserQuota,
		TrialVIPDays:       req.TrialVIPDays,
	}
	if current != nil {
		settings.AllowedDomains = current.AllowedDomains
		settings.MaxAccountsPerIP = current.MaxAccountsPerIP
		settings.MinPasswordLength = current.MinPasswordLength
		settings.SignupRewardType = current.SignupRewardType
		settings.SignupRewardAmount = current.SignupRewardAmount
	}
	if req.AllowedDomains != nil {
		settings.AllowedDomains = *req.AllowedDomains
	}
	if req.MaxAccountsPerIP != nil {
		settings.MaxAccountsPerIP = *req.MaxAccountsPerIP
	}
	if req.MinPasswordLength != nil {
		settings.MinPasswordLength = *req.MinPasswordLength
	}
	if req.SignupRewardType != nil {
		settings.SignupRewardType = *req.SignupRewardType
	}
	if req.SignupRewardAmount != nil {
		settings.SignupRewardAmount = *req.SignupRewardAmount
	}

	if err := h.settingsSvc.UpdateRegisterSettings(settings); err != nil {
		response.InternalError(c, "failed to update register settings: "+err.Error())
		return
	}

	response.Success(c, nil)
}

func (h *SettingsHandler) GetPaymentConfig(c *gin.Context) {
	cfg, err := h.settingsSvc.GetAlipayConfig()
	if err != nil {
		response.InternalError(c, "failed to get payment config: "+err.Error())
		return
	}

	response.Success(c, cfg)
}

func (h *SettingsHandler) UpdatePaymentConfig(c *gin.Context) {
	var req struct {
		Enabled    bool   `json:"enabled"`
		AppID      string `json:"app_id"`
		PrivateKey string `json:"private_key"`
		PublicKey  string `json:"public_key"`
		EncryptKey string `json:"encrypt_key"`
		Sandbox    bool   `json:"sandbox"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	cfg := &service.AlipayConfig{
		Enabled:    req.Enabled,
		AppID:      req.AppID,
		PrivateKey: req.PrivateKey,
		PublicKey:  req.PublicKey,
		EncryptKey: req.EncryptKey,
		Sandbox:    req.Sandbox,
	}

	if err := h.settingsSvc.UpdateAlipayConfig(cfg); err != nil {
		response.InternalError(c, "failed to update payment config: "+err.Error())
		return
	}

	h.settingsSvc.InvalidateCache()
	if h.alipayService != nil {
		h.alipayService.ReloadClient()
	}
	response.Success(c, nil)
}

func (h *SettingsHandler) GetGeneralSettings(c *gin.Context) {
	cfg, err := h.settingsSvc.GetGeneralSettings()
	if err != nil {
		response.InternalError(c, "failed to get general settings: "+err.Error())
		return
	}
	response.Success(c, cfg)
}

func (h *SettingsHandler) UpdateGeneralSettings(c *gin.Context) {
	var req struct {
		SiteName        string `json:"site_name"`
		SiteLogo        string `json:"site_logo"`
		SiteDescription string `json:"site_description"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	cfg := &service.GeneralSettings{
		SiteName:        req.SiteName,
		SiteLogo:        req.SiteLogo,
		SiteDescription: req.SiteDescription,
	}

	if err := h.settingsSvc.UpdateGeneralSettings(cfg); err != nil {
		response.InternalError(c, "failed to update general settings: "+err.Error())
		return
	}

	h.settingsSvc.InvalidateCache()
	response.Success(c, nil)
}

func (h *SettingsHandler) GetRateLimitSettings(c *gin.Context) {
	cfg, err := h.settingsSvc.GetRateLimitSettings()
	if err != nil {
		response.InternalError(c, "failed to get rate limit settings: "+err.Error())
		return
	}
	response.Success(c, cfg)
}

func (h *SettingsHandler) UpdateRateLimitSettings(c *gin.Context) {
	var req struct {
		FreeRPM int `json:"free_rpm"`
		FreeTPM int `json:"free_tpm"`
		VIPRPM  int `json:"vip_rpm"`
		VIPTPM  int `json:"vip_tpm"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	cfg := &service.RateLimitSettings{
		FreeRPM: req.FreeRPM,
		FreeTPM: req.FreeTPM,
		VIPRPM:  req.VIPRPM,
		VIPTPM:  req.VIPTPM,
	}

	if err := h.settingsSvc.UpdateRateLimitSettings(cfg); err != nil {
		response.InternalError(c, "failed to update rate limit settings: "+err.Error())
		return
	}

	h.settingsSvc.InvalidateCache()
	response.Success(c, nil)
}

func (h *SettingsHandler) GetSecuritySettings(c *gin.Context) {
	cfg, err := h.settingsSvc.GetSecuritySettings()
	if err != nil {
		response.InternalError(c, "failed to get security settings: "+err.Error())
		return
	}
	response.Success(c, cfg)
}

func (h *SettingsHandler) UpdateSecuritySettings(c *gin.Context) {
	var req struct {
		JWTSecret          string `json:"jwt_secret"`
		JWTExpireHours     int    `json:"jwt_expire_hours"`
		PasswordMinLength  int    `json:"password_min_length"`
		PasswordExpireDays int    `json:"password_expire_days"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	cfg := &service.SecuritySettings{
		JWTSecret:          req.JWTSecret,
		JWTExpireHours:     req.JWTExpireHours,
		PasswordMinLength:  req.PasswordMinLength,
		PasswordExpireDays: req.PasswordExpireDays,
	}

	if err := h.settingsSvc.UpdateSecuritySettings(cfg); err != nil {
		response.InternalError(c, "failed to update security settings: "+err.Error())
		return
	}

	h.settingsSvc.InvalidateCache()
	response.Success(c, nil)
}

package handler

import (
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type SettingsHandler struct {
	settingsSvc *service.SettingsService
}

func NewSettingsHandler(settingsSvc *service.SettingsService) *SettingsHandler {
	return &SettingsHandler{settingsSvc: settingsSvc}
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
		AllowRegister      bool  `json:"allow_register"`
		RequireEmailVerify *bool `json:"require_email_verify"`
		EnableCaptcha      bool  `json:"enable_captcha"`
		NewUserQuota       int   `json:"new_user_quota"`
		TrialVIPDays       int   `json:"trial_vip_days"`
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

	settings := &service.RegisterSettings{
		AllowRegister:      req.AllowRegister,
		RequireEmailVerify: req.RequireEmailVerify,
		EnableCaptcha:      req.EnableCaptcha,
		NewUserQuota:       req.NewUserQuota,
		TrialVIPDays:       req.TrialVIPDays,
	}

	if err := h.settingsSvc.UpdateRegisterSettings(settings); err != nil {
		response.InternalError(c, "failed to update register settings: "+err.Error())
		return
	}

	response.Success(c, nil)
}

package handler

import (
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type EmailVerificationHandler struct {
	emailService *service.EmailVerificationService
}

func NewEmailVerificationHandler(emailService *service.EmailVerificationService) *EmailVerificationHandler {
	return &EmailVerificationHandler{emailService: emailService}
}

func (h *EmailVerificationHandler) SendCode(c *gin.Context) {
	var req struct {
		Email        string `json:"email" binding:"required,email"`
		CaptchaToken string `json:"captcha_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	deviceHash := c.GetHeader("X-Device-Hash")

	if err := h.emailService.CheckSendLimit(req.Email, ip, deviceHash); err != nil {
		response.Fail(c, "RATE_LIMITED", err.Error())
		return
	}

	if err := h.emailService.SendVerificationEmail(req.Email, ip, userAgent, deviceHash, req.CaptchaToken); err != nil {
		response.InternalError(c, "发送验证码失败")
		return
	}

	response.SuccessWithMessage(c, nil, "验证码已发送")
}

func (h *EmailVerificationHandler) VerifyCode(c *gin.Context) {
	var req struct {
		Email string `json:"email" binding:"required,email"`
		Code  string `json:"code" binding:"required,len=6"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	valid, err := h.emailService.VerifyCode(req.Email, req.Code)
	if err != nil {
		response.InternalError(c, "验证失败")
		return
	}

	if !valid {
		response.Fail(c, "INVALID_CODE", "验证码无效或已过期")
		return
	}

	response.SuccessWithMessage(c, nil, "验证成功")
}

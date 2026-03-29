package handler

import (
	"errors"
	"time"

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
		Purpose      string `json:"purpose"`
		CaptchaToken string `json:"captcha_token" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Purpose == "" {
		req.Purpose = "register"
	}
	if req.Purpose != "register" && req.Purpose != "reset" {
		response.Fail(c, "INVALID_PARAMETER", "purpose must be 'register' or 'reset'")
		return
	}

	ip := c.ClientIP()
	userAgent := c.GetHeader("User-Agent")
	deviceHash := c.GetHeader("X-Device-Hash")

	if err := h.emailService.CheckSendLimit(req.Email, ip, deviceHash); err != nil {
		response.Fail(c, "RATE_LIMITED", "验证码发送过于频繁，请稍后再试")
		return
	}

	if req.Purpose == "reset" {
		if err := h.emailService.SendPasswordResetEmail(req.Email, ip, userAgent, deviceHash, req.CaptchaToken); err != nil {
			response.InternalError(c, "发送重置邮件失败")
			return
		}
	} else {
		if err := h.emailService.SendVerificationCode(req.Email, ip, userAgent, deviceHash, req.CaptchaToken, req.Purpose); err != nil {
			response.InternalError(c, "发送验证码失败")
			return
		}
	}

	response.SuccessWithMessage(c, gin.H{
		"expires_in": 600,
	}, "验证码已发送")
}

func (h *EmailVerificationHandler) VerifyCode(c *gin.Context) {
	var req struct {
		Email   string `json:"email" binding:"required,email"`
		Code    string `json:"code" binding:"required,len=6"`
		Purpose string `json:"purpose"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Purpose == "" {
		req.Purpose = "register"
	}

	valid, token, err := h.emailService.VerifyCode(req.Email, req.Code, req.Purpose)
	if err != nil {
		if errors.Is(err, service.ErrCodeInvalid) {
			response.Fail(c, "INVALID_CODE", "验证码无效或已过期")
			return
		}
		if errors.Is(err, service.ErrCodeMaxAttempts) {
			response.Fail(c, "CODE_EXHAUSTED", "验证码尝试次数过多，请重新获取")
			return
		}
		response.InternalError(c, "验证失败")
		return
	}

	if !valid {
		response.Fail(c, "INVALID_CODE", "验证码无效或已过期")
		return
	}

	response.Success(c, gin.H{
		"verification_token": token,
		"expires_in":         300,
	})
}

func (h *EmailVerificationHandler) ForgotPassword(c *gin.Context) {
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
		response.Fail(c, "RATE_LIMITED", "请求过于频繁，请稍后再试")
		return
	}

	if err := h.emailService.SendPasswordResetEmail(req.Email, ip, userAgent, deviceHash, req.CaptchaToken); err != nil {
		response.InternalError(c, "发送重置邮件失败")
		return
	}

	response.SuccessWithMessage(c, nil, "如果该邮箱已注册，重置链接已发送")
}

func (h *EmailVerificationHandler) VerifyResetToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Fail(c, "MISSING_TOKEN", "token is required")
		return
	}

	reset, err := h.emailService.VerifyResetToken(token)
	if err != nil {
		response.Fail(c, "INVALID_TOKEN", "重置链接无效或已过期")
		return
	}

	response.Success(c, gin.H{
		"email":      reset.Email,
		"expires_at": reset.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *EmailVerificationHandler) ResetPassword(c *gin.Context) {
	var req struct {
		Token           string `json:"token" binding:"required"`
		Password        string `json:"password" binding:"required,min=8"`
		ConfirmPassword string `json:"confirm_password" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	if req.Password != req.ConfirmPassword {
		response.Fail(c, "PASSWORD_MISMATCH", "两次输入的密码不一致")
		return
	}

	if err := h.emailService.ResetPassword(req.Token, req.Password); err != nil {
		response.Fail(c, "RESET_FAILED", err.Error())
		return
	}

	response.SuccessWithMessage(c, nil, "密码重置成功")
}

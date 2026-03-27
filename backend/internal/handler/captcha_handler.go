package handler

import (
	"gapi-platform/internal/pkg/response"
	"gapi-platform/internal/service"
	"github.com/gin-gonic/gin"
)

type CaptchaHandler struct {
	captchaService *service.SliderCaptchaService
}

func NewCaptchaHandler(captchaService *service.SliderCaptchaService) *CaptchaHandler {
	return &CaptchaHandler{captchaService: captchaService}
}

func (h *CaptchaHandler) Generate(c *gin.Context) {
	data, err := h.captchaService.Generate()
	if err != nil {
		response.InternalError(c, "生成验证码失败")
		return
	}

	response.Success(c, data)
}

func (h *CaptchaHandler) Verify(c *gin.Context) {
	var req struct {
		Token    string `json:"token" binding:"required"`
		Track    []int  `json:"track" binding:"required"`
		Duration int64  `json:"duration" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		response.Fail(c, "INVALID_PARAMETER", err.Error())
		return
	}

	passed, score, err := h.captchaService.Verify(req.Token, req.Track, req.Duration)
	if err != nil {
		response.Fail(c, "CAPTCHA_INVALID", err.Error())
		return
	}

	if !passed {
		response.Fail(c, "CAPTCHA_FAILED", "验证失败，请重试")
		return
	}

	response.Success(c, gin.H{
		"passed": passed,
		"score":  score,
		"token":  req.Token,
	})
}

func (h *CaptchaHandler) ValidateToken(c *gin.Context) {
	token := c.Query("token")
	if token == "" {
		response.Fail(c, "INVALID_PARAMETER", "token is required")
		return
	}

	valid := h.captchaService.ValidateToken(token)

	response.Success(c, gin.H{
		"valid": valid,
	})
}

package middleware

import (
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/repository"
	"github.com/gin-gonic/gin"
)

var loggedEndpoints = map[string]bool{
	"/api/v1/completions":      true,
	"/api/v1/chat/completions": true,
	"/api/v1/embeddings":       true,
	"/api/v1/models":           true,
}

func APIAccessLog(apiAccessLogRepo *repository.APIAccessLogRepository) gin.HandlerFunc {
	return func(c *gin.Context) {
		if !loggedEndpoints[c.Request.URL.Path] {
			c.Next()
			return
		}

		startTime := time.Now()

		c.Next()

		go func() {
			var userID uint
			if uid, exists := c.Get("user_id"); exists {
				if id, ok := uid.(uint); ok {
					userID = id
				}
			}

			if userID == 0 {
				return
			}

			responseTime := int(time.Since(startTime).Milliseconds())
			statusCode := c.Writer.Status()

			modelName := "unknown"
			if m, exists := c.Get("request_model"); exists {
				if mStr, ok := m.(string); ok && mStr != "" {
					modelName = mStr
				}
			}

			tokenID, _ := c.Get("token_id")

			log := &model.APIAccessLog{
				UserID:       userID,
				Endpoint:     c.Request.URL.Path,
				Method:       c.Request.Method,
				Model:        modelName,
				StatusCode:   statusCode,
				ResponseTime: responseTime,
				RequestIP:    c.ClientIP(),
				UserAgent:    c.Request.UserAgent(),
				CreatedAt:    startTime,
			}

			// T-13: Per-stage latency breakdown
			if v, exists := c.Get("pre_check_duration_ms"); exists {
				if d, ok := v.(int); ok {
					log.PreCheckDurationMs = d
				}
			}
			if v, exists := c.Get("channel_select_duration_ms"); exists {
				if d, ok := v.(int); ok {
					log.ChannelSelectDurationMs = d
				}
			}
			if v, exists := c.Get("upstream_duration_ms"); exists {
				if d, ok := v.(int); ok {
					log.UpstreamDurationMs = d
				}
			}

			if tid, ok := tokenID.(uint); ok {
				log.TokenID = &tid
			}

			if statusCode >= 400 {
				if errMsg, exists := c.Get("error_message"); exists {
					if msg, ok := errMsg.(string); ok {
						log.ErrorMessage = msg
					}
				}
			}

			apiAccessLogRepo.Create(log)
		}()
	}
}

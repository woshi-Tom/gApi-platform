package middleware

import (
	"net/http"

	"gapi-platform/internal/pkg/response"
	"github.com/gin-gonic/gin"
)

func SecurityHeaders() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("X-Content-Type-Options", "nosniff")
		c.Header("X-Frame-Options", "SAMEORIGIN")
		c.Header("X-XSS-Protection", "1; mode=block")
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Header("Referrer-Policy", "strict-origin-when-cross-origin")

		c.Next()
	}
}

const maxBodySize = 10 << 20

func BodySizeLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		if c.Request.Body == nil {
			c.Next()
			return
		}

		contentLength := c.Request.ContentLength
		if contentLength > 0 && contentLength > maxBodySize {
			response.Fail(c, "PAYLOAD_TOO_LARGE", "request body too large")
			c.Abort()
			return
		}

		c.Next()
	}
}

func MaxBodyBytes(bytes int64) gin.HandlerFunc {
	maxSize := bytes
	return func(c *gin.Context) {
		c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxSize)
		c.Next()
	}
}

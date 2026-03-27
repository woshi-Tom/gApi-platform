package middleware

import (
	"net/http"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements a simple in-memory rate limiter
type RateLimiter struct {
	mu       sync.RWMutex
	visitors map[string]*visitor
	limit    rate.Limit
	burst    int
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(rps float64, burst int) *RateLimiter {
	rl := &RateLimiter{
		visitors: make(map[string]*visitor),
		limit:    rate.Limit(rps),
		burst:    burst,
	}

	// Clean up old visitors every 3 minutes
	go rl.cleanupVisitors()

	return rl
}

func (rl *RateLimiter) getVisitor(ip string) *rate.Limiter {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	v, exists := rl.visitors[ip]
	if !exists {
		limiter := rate.NewLimiter(rl.limit, rl.burst)
		rl.visitors[ip] = &visitor{limiter: limiter, lastSeen: time.Now()}
		return limiter
	}

	v.lastSeen = time.Now()
	return v.limiter
}

func (rl *RateLimiter) cleanupVisitors() {
	for {
		time.Sleep(3 * time.Minute)

		rl.mu.Lock()
		for ip, v := range rl.visitors {
			if time.Since(v.lastSeen) > 3*time.Minute {
				delete(rl.visitors, ip)
			}
		}
		rl.mu.Unlock()
	}
}

// RateLimit creates a rate limiting middleware
func RateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.getVisitor(ip).Allow() {
			c.JSON(http.StatusTooManyRequests, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "RATE_LIMITED",
					Message: "Too many requests, please try again later",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// UserRateLimit creates a per-user rate limiting middleware
func UserRateLimit(rps float64, burst int) gin.HandlerFunc {
	limiter := NewRateLimiter(rps, burst)

	return func(c *gin.Context) {
		// Get user ID from context (set by JWT middleware)
		var key string
		if userID, exists := c.Get("user_id"); exists {
			key = string(rune(userID.(uint)))
		} else {
			key = c.ClientIP()
		}

		if !limiter.getVisitor(key).Allow() {
			c.JSON(http.StatusTooManyRequests, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "RATE_LIMITED",
					Message: "Too many requests, please try again later",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// TokenRateLimit creates a per-token rate limiting middleware
func TokenRateLimit() gin.HandlerFunc {
	// Use a map of token ID to rate limiter
	limiters := make(map[uint]*rate.Limiter)
	mu := sync.RWMutex{}

	return func(c *gin.Context) {
		tokenID, exists := c.Get("token_id")
		if !exists {
			c.Next()
			return
		}

		tid := tokenID.(uint)

		mu.RLock()
		limiter, ok := limiters[tid]
		mu.RUnlock()

		if !ok {
			mu.Lock()
			limiter = rate.NewLimiter(60, 60) // 60 requests per minute, burst 60
			limiters[tid] = limiter
			mu.Unlock()
		}

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, model.APIResponse{
				Success: false,
				Error: &model.APIError{
					Code:    "RATE_LIMITED",
					Message: "Token rate limit exceeded",
				},
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

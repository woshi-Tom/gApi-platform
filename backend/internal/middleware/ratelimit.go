package middleware

import (
	"net/http"
	"strconv"
	"sync"
	"time"

	"gapi-platform/internal/model"
	"gapi-platform/internal/service"
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
			key = strconv.FormatUint(uint64(userID.(uint)), 10)
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
// T-05: Supports per-token RPM/TPM configuration from the database
func TokenRateLimit(tokenService *service.TokenService) gin.HandlerFunc {
	limiters := make(map[uint]*tokenLimiterEntry)
	mu := sync.RWMutex{}

	// Cleanup goroutine: remove stale limiters every 10 minutes
	go func() {
		for {
			time.Sleep(10 * time.Minute)
			mu.Lock()
			now := time.Now()
			for tid, entry := range limiters {
				if now.Sub(entry.lastUsed) > 30*time.Minute {
					delete(limiters, tid)
				}
			}
			mu.Unlock()
		}
	}()

	return func(c *gin.Context) {
		tokenID, exists := c.Get("token_id")
		if !exists {
			c.Next()
			return
		}

		tid := tokenID.(uint)

		mu.RLock()
		entry, ok := limiters[tid]
		mu.RUnlock()

		if !ok {
			// Load token config from database
			rpm := 10 // default RPM
			tpm := 0  // default: no TPM limit
			if tokenService != nil {
				if token, err := tokenService.GetByID(tid); err == nil && token != nil {
					if token.RPMLimit != nil && *token.RPMLimit > 0 {
						rpm = *token.RPMLimit
					}
					if token.TPMLimit != nil && *token.TPMLimit > 0 {
						tpm = *token.TPMLimit
					}
				}
			}

			mu.Lock()
			// Double-check after acquiring write lock
			if entry, ok = limiters[tid]; !ok {
				burst := rpm
				if burst < 1 {
					burst = 1
				}
				entry = &tokenLimiterEntry{
					limiter:   rate.NewLimiter(rate.Limit(rpm)/60.0, burst), // RPM → per-second
					tpmLimit:  tpm,
					tpmWindow: time.Now(),
					tpmCount:  0,
					lastUsed:  time.Now(),
				}
				limiters[tid] = entry
			}
			mu.Unlock()
		}

		mu.Lock()
		entry.lastUsed = time.Now()
		mu.Unlock()

		// RPM check
		if !entry.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, model.APIErrorResponse{
				Error: &model.APIError{
					Type:    "rate_limit_error",
					Code:    "rate_limit_rpm",
					Message: "Token RPM rate limit exceeded",
				},
			})
			c.Abort()
			return
		}

		// TPM check (simple per-minute window)
		if entry.tpmLimit > 0 {
			mu.Lock()
			now := time.Now()
			if now.Sub(entry.tpmWindow) > time.Minute {
				entry.tpmWindow = now
				entry.tpmCount = 0
			}
			entry.tpmCount++
			exceeded := entry.tpmCount > entry.tpmLimit
			mu.Unlock()

			if exceeded {
				c.JSON(http.StatusTooManyRequests, model.APIErrorResponse{
					Error: &model.APIError{
						Type:    "rate_limit_error",
						Code:    "rate_limit_tpm",
						Message: "Token TPM rate limit exceeded",
					},
				})
				c.Abort()
				return
			}
		}

		c.Next()
	}
}

type tokenLimiterEntry struct {
	limiter   *rate.Limiter
	lastUsed  time.Time
	tpmLimit  int
	tpmWindow time.Time
	tpmCount  int
}

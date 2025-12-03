package middleware

import (
    "net/http"
    "time"

    "ai-api-gateway/internal/metrics"
    "ai-api-gateway/internal/ratelimiter"

    "github.com/gin-gonic/gin"
)

// RateLimitMiddleware creates rate limiting middleware
type RateLimitMiddleware struct {
	limiter   ratelimiter.RateLimiter
	enabled   bool
	keyFunc   func(*gin.Context) string
	algorithm string
}

// NewRateLimitMiddleware creates a new rate limiting middleware
func NewRateLimitMiddleware(limiter ratelimiter.RateLimiter, enabled bool, algorithm string) *RateLimitMiddleware {
	return &RateLimitMiddleware{
		limiter:   limiter,
		enabled:   enabled,
		keyFunc:   defaultKeyFunc,
		algorithm: algorithm,
	}
}

// SetKeyFunc sets a custom function to extract rate limit key
func (m *RateLimitMiddleware) SetKeyFunc(fn func(*gin.Context) string) {
	m.keyFunc = fn
}

// Middleware returns the rate limiting middleware handler
func (m *RateLimitMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		if !m.enabled {
			c.Next()
			return
		}

		// Extract rate limit key
		key := m.keyFunc(c)

		// Check rate limit
		allowed, limitInfo, err := m.limiter.Allow(c.Request.Context(), key)
		if err != nil {
			// On error, allow the request but log it
			c.Next()
			return
		}

		// Set rate limit headers
		headers := limitInfo.GetHeaders()
		for k, v := range headers {
			c.Header(k, v)
		}

		if !allowed {
			metrics.RateLimitHits.WithLabelValues(key, m.algorithm).Inc()
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":     "Too Many Requests",
				"message":   "Rate limit exceeded",
				"code":      http.StatusTooManyRequests,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// defaultKeyFunc extracts rate limit key from request
func defaultKeyFunc(c *gin.Context) string {
	// Try to get user ID from context (set by auth middleware)
	if userID, exists := c.Get("user_id"); exists {
		if id, ok := userID.(string); ok && id != "" {
			return "user:" + id
		}
	}

	// Fallback to IP address
	return "ip:" + c.ClientIP()
}


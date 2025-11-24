package middleware

import (
	"time"

	"ai-api-gateway/internal/config"

	"github.com/gin-gonic/gin"
)

// RequestLogger logs HTTP requests in structured format
func RequestLogger(logger *config.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(start)

		// Build log fields
		fields := map[string]interface{}{
			"method":     c.Request.Method,
			"path":       path,
			"query":      raw,
			"status":     c.Writer.Status(),
			"latency_ms": latency.Milliseconds(),
			"ip":         c.ClientIP(),
			"user_agent": c.Request.UserAgent(),
			"bytes_in":   c.Request.ContentLength,
			"bytes_out":  c.Writer.Size(),
		}

		// Log based on status code
		if c.Writer.Status() >= 500 {
			logger.Error("HTTP request", fields)
		} else if c.Writer.Status() >= 400 {
			logger.Warn("HTTP request", fields)
		} else {
			logger.Info("HTTP request", fields)
		}
	}
}

package auth

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"ai-api-gateway/internal/config"
	"ai-api-gateway/internal/metrics"

	"github.com/gin-gonic/gin"
)

// AuthMiddleware creates authentication middleware
type AuthMiddleware struct {
	jwtVerifier  *JWTVerifier
	oidcVerifier *OIDCVerifier
	config       *config.AuthConfig
}

// NewAuthMiddleware creates a new authentication middleware
func NewAuthMiddleware(cfg *config.AuthConfig, jwtVerifier *JWTVerifier, oidcVerifier *OIDCVerifier) *AuthMiddleware {
	return &AuthMiddleware{
		jwtVerifier:  jwtVerifier,
		oidcVerifier: oidcVerifier,
		config:       cfg,
	}
}

// Middleware returns the authentication middleware handler
func (m *AuthMiddleware) Middleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Check if path should skip authentication
		if m.shouldSkipAuth(c.Request.URL.Path) {
			c.Next()
			return
		}

		// Extract token from Authorization header
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			metrics.AuthFailures.WithLabelValues("missing_token", m.config.Type).Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Unauthorized",
				"message":   "Missing Authorization header",
				"code":      http.StatusUnauthorized,
				"timestamp": c.GetString("timestamp"),
			})
			c.Abort()
			return
		}

		// Verify token based on auth type
		var claims *Claims
		var err error

		switch m.config.Type {
		case "jwt":
			claims, err = m.verifyJWT(authHeader)
		case "oidc":
			claims, err = m.verifyOIDC(c.Request.Context(), authHeader)
		case "both":
			// Try JWT first, then OIDC
			claims, err = m.verifyJWT(authHeader)
			if err != nil {
				claims, err = m.verifyOIDC(c.Request.Context(), authHeader)
			}
		case "mock":
			// Mock authentication for testing
			claims = &Claims{
				Subject: "mock-user",
				UserID:  "mock-user-id",
				Email:   "mock@example.com",
				Roles:   []string{"user"},
			}
			err = nil
		default:
			c.JSON(http.StatusInternalServerError, gin.H{
				"error":     "Internal Server Error",
				"message":   "Invalid authentication configuration",
				"code":      http.StatusInternalServerError,
				"timestamp": c.GetString("timestamp"),
			})
			c.Abort()
			return
		}

		if err != nil {
			metrics.AuthFailures.WithLabelValues("invalid_token", m.config.Type).Inc()
			c.JSON(http.StatusUnauthorized, gin.H{
				"error":     "Unauthorized",
				"message":   "Invalid or expired token",
				"code":      http.StatusUnauthorized,
				"timestamp": c.GetString("timestamp"),
			})
			c.Abort()
			return
		}

		// Store claims in context
		ctx := context.WithValue(c.Request.Context(), ClaimsContextKey, claims)
		c.Request = c.Request.WithContext(ctx)

		// Store user info in gin context for easy access
		c.Set("user_id", claims.GetUserID())
		c.Set("user_email", claims.Email)
		c.Set("user_roles", claims.Roles)

		c.Next()
	}
}

// verifyJWT verifies a JWT token
func (m *AuthMiddleware) verifyJWT(token string) (*Claims, error) {
	if m.jwtVerifier == nil {
		return nil, fmt.Errorf("JWT verifier not configured")
	}
	return m.jwtVerifier.Verify(token)
}

// verifyOIDC verifies an OIDC token
func (m *AuthMiddleware) verifyOIDC(ctx context.Context, token string) (*Claims, error) {
	if m.oidcVerifier == nil {
		return nil, fmt.Errorf("OIDC verifier not configured")
	}
	return m.oidcVerifier.Verify(ctx, token)
}

// shouldSkipAuth checks if authentication should be skipped for a path
func (m *AuthMiddleware) shouldSkipAuth(path string) bool {
	for _, skipPath := range m.config.SkipAuthPaths {
		if strings.HasPrefix(path, skipPath) {
			return true
		}
	}
	return false
}


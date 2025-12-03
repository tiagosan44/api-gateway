package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

// AdminAuth checks if the user has admin role
func AdminAuth() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Get user roles from context
		roles, exists := c.Get("user_roles")
		if !exists {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Forbidden",
				"message":   "User roles not found",
				"code":      http.StatusForbidden,
			})
			c.Abort()
			return
		}

		rolesSlice, ok := roles.([]string)
		if !ok {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Forbidden",
				"message":   "Invalid user roles",
				"code":      http.StatusForbidden,
			})
			c.Abort()
			return
		}

		// Check if user has admin role
		hasAdmin := false
		for _, role := range rolesSlice {
			if strings.ToLower(role) == "admin" {
				hasAdmin = true
				break
			}
		}

		if !hasAdmin {
			c.JSON(http.StatusForbidden, gin.H{
				"error":     "Forbidden",
				"message":   "Admin role required",
				"code":      http.StatusForbidden,
			})
			c.Abort()
			return
		}

		c.Next()
	}
}


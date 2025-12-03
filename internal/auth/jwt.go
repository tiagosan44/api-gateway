package auth

import (
	"context"
	"fmt"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// JWTVerifier handles JWT token verification
type JWTVerifier struct {
	secret []byte
}

// NewJWTVerifier creates a new JWT verifier
func NewJWTVerifier(secret string) (*JWTVerifier, error) {
	if secret == "" {
		return nil, fmt.Errorf("JWT secret cannot be empty")
	}
	return &JWTVerifier{
		secret: []byte(secret),
	}, nil
}

// Verify verifies a JWT token and returns the claims
func (v *JWTVerifier) Verify(tokenString string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	// Parse and verify token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Validate signing method
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return v.secret, nil
	})

	if err != nil {
		return nil, fmt.Errorf("failed to parse token: %w", err)
	}

	if !token.Valid {
		return nil, fmt.Errorf("invalid token")
	}

	// Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, fmt.Errorf("invalid token claims")
	}

	// Convert to our Claims struct
	result := &Claims{
		Subject: extractString(claims, "sub"),
		UserID:  extractString(claims, "user_id"),
		Email:   extractString(claims, "email"),
		Roles:   extractStringSlice(claims, "roles"),
		Raw:     claims,
	}

	return result, nil
}

// extractString extracts a string value from claims
func extractString(claims jwt.MapClaims, key string) string {
	if val, ok := claims[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// extractStringSlice extracts a string slice from claims
func extractStringSlice(claims jwt.MapClaims, key string) []string {
	if val, ok := claims[key]; ok {
		if arr, ok := val.([]interface{}); ok {
			result := make([]string, 0, len(arr))
			for _, v := range arr {
				if str, ok := v.(string); ok {
					result = append(result, str)
				}
			}
			return result
		}
	}
	return []string{}
}

// Claims represents JWT claims
type Claims struct {
	Subject string
	UserID  string
	Email   string
	Roles   []string
	Raw     jwt.MapClaims
}

// GetUserID returns the user ID from claims
func (c *Claims) GetUserID() string {
	if c.UserID != "" {
		return c.UserID
	}
	return c.Subject
}

// HasRole checks if the user has a specific role
func (c *Claims) HasRole(role string) bool {
	for _, r := range c.Roles {
		if r == role {
			return true
		}
	}
	return false
}

// ContextKey is a type for context keys
type ContextKey string

const (
	// ClaimsContextKey is the key for storing claims in context
	ClaimsContextKey ContextKey = "claims"
)

// GetClaimsFromContext extracts claims from context
func GetClaimsFromContext(ctx context.Context) (*Claims, bool) {
	claims, ok := ctx.Value(ClaimsContextKey).(*Claims)
	return claims, ok
}


package auth

import (
	"context"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"math/big"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/oauth2"
)

// OIDCVerifier handles OIDC token verification
type OIDCVerifier struct {
	provider   *oidc.Provider
	verifier   *oidc.IDTokenVerifier
	httpClient *http.Client
	mu         sync.RWMutex
	jwks       map[string]*rsa.PublicKey
}

// OIDCConfig holds OIDC configuration
type OIDCConfig struct {
	Issuer       string
	ClientID     string
	ClientSecret string
}

// NewOIDCVerifier creates a new OIDC verifier
func NewOIDCVerifier(ctx context.Context, config OIDCConfig) (*OIDCVerifier, error) {
	if config.Issuer == "" {
		return nil, fmt.Errorf("OIDC issuer cannot be empty")
	}

	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: 10 * time.Second,
	}

	// Create context with HTTP client
	ctx = oidc.ClientContext(ctx, httpClient)

	// Create provider
	provider, err := oidc.NewProvider(ctx, config.Issuer)
	if err != nil {
		return nil, fmt.Errorf("failed to create OIDC provider: %w", err)
	}

	// Create verifier
	verifier := provider.Verifier(&oidc.Config{
		ClientID: config.ClientID,
	})

	return &OIDCVerifier{
		provider:   provider,
		verifier:   verifier,
		httpClient: httpClient,
		jwks:       make(map[string]*rsa.PublicKey),
	}, nil
}

// Verify verifies an OIDC token and returns the claims
func (v *OIDCVerifier) Verify(ctx context.Context, tokenString string) (*Claims, error) {
	// Remove "Bearer " prefix if present
	tokenString = strings.TrimPrefix(tokenString, "Bearer ")
	tokenString = strings.TrimSpace(tokenString)

	// Verify token with OIDC provider
	idToken, err := v.verifier.Verify(ctx, tokenString)
	if err != nil {
		return nil, fmt.Errorf("failed to verify token: %w", err)
	}

	// Extract claims
	var claims map[string]interface{}
	if err := idToken.Claims(&claims); err != nil {
		return nil, fmt.Errorf("failed to extract claims: %w", err)
	}

	// Convert to our Claims struct
	result := &Claims{
		Subject: extractStringFromMap(claims, "sub"),
		UserID:  extractStringFromMap(claims, "user_id"),
		Email:   extractStringFromMap(claims, "email"),
		Roles:   extractStringSliceFromMap(claims, "roles"),
		Raw:     jwt.MapClaims(claims),
	}

	return result, nil
}

// extractStringFromMap extracts a string value from a map
func extractStringFromMap(m map[string]interface{}, key string) string {
	if val, ok := m[key]; ok {
		if str, ok := val.(string); ok {
			return str
		}
	}
	return ""
}

// extractStringSliceFromMap extracts a string slice from a map
func extractStringSliceFromMap(m map[string]interface{}, key string) []string {
	if val, ok := m[key]; ok {
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

// GetOAuth2Config returns OAuth2 configuration for token exchange
func (v *OIDCVerifier) GetOAuth2Config(clientID, clientSecret, redirectURL string, scopes []string) *oauth2.Config {
	return &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Endpoint:     v.provider.Endpoint(),
		RedirectURL:  redirectURL,
		Scopes:       scopes,
	}
}

// JWKS represents JSON Web Key Set
type JWKS struct {
	Keys []JWK `json:"keys"`
}

// JWK represents a JSON Web Key
type JWK struct {
	Kty string `json:"kty"`
	Kid string `json:"kid"`
	Use string `json:"use"`
	N   string `json:"n"`
	E   string `json:"e"`
}

// LoadJWKS loads JSON Web Key Set from provider
func (v *OIDCVerifier) LoadJWKS(ctx context.Context) error {
	jwksURL := v.provider.Endpoint().AuthURL + "/.well-known/jwks.json"

	resp, err := v.httpClient.Get(jwksURL)
	if err != nil {
		return fmt.Errorf("failed to fetch JWKS: %w", err)
	}
	defer resp.Body.Close()

	var jwks JWKS
	if err := json.NewDecoder(resp.Body).Decode(&jwks); err != nil {
		return fmt.Errorf("failed to decode JWKS: %w", err)
	}

	v.mu.Lock()
	defer v.mu.Unlock()

	for _, key := range jwks.Keys {
		if key.Kty == "RSA" {
			publicKey, err := keyToRSAPublicKey(key)
			if err != nil {
				continue
			}
			v.jwks[key.Kid] = publicKey
		}
	}

	return nil
}

// keyToRSAPublicKey converts JWK to RSA public key
func keyToRSAPublicKey(jwk JWK) (*rsa.PublicKey, error) {
	// Decode modulus
	nBytes, err := base64.RawURLEncoding.DecodeString(jwk.N)
	if err != nil {
		return nil, fmt.Errorf("failed to decode modulus: %w", err)
	}

	// Decode exponent
	eBytes, err := base64.RawURLEncoding.DecodeString(jwk.E)
	if err != nil {
		return nil, fmt.Errorf("failed to decode exponent: %w", err)
	}

	// Convert exponent bytes to int
	var eInt int
	for _, b := range eBytes {
		eInt = eInt<<8 | int(b)
	}

	// Create RSA public key
	return &rsa.PublicKey{
		N: new(big.Int).SetBytes(nBytes),
		E: eInt,
	}, nil
}


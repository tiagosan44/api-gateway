package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

// Config holds all configuration for the gateway
type Config struct {
	Server        ServerConfig
	Redis         RedisConfig
	Auth          AuthConfig
	RateLimit     RateLimitConfig
	Proxy         ProxyConfig
	Observability ObservabilityConfig
}

// ServerConfig holds server configuration
type ServerConfig struct {
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

// RedisConfig holds Redis configuration
type RedisConfig struct {
	URL           string
	MaxRetries    int
	PoolSize      int
	MinIdleConns  int
	DialTimeout   time.Duration
	ReadTimeout   time.Duration
	WriteTimeout  time.Duration
	PoolTimeout   time.Duration
	IdleTimeout   time.Duration
	IdleCheckFreq time.Duration
}

// AuthConfig holds authentication configuration
type AuthConfig struct {
	Type             string // "jwt", "oidc", "both", "mock"
	JWTSecret        string
	OIDCIssuer       string
	OIDCClientID     string
	OIDCClientSecret string
	SkipAuthPaths    []string
}

// RateLimitConfig holds rate limiting configuration
type RateLimitConfig struct {
	Enabled    bool
	Algorithm  string // "token_bucket", "leaky_bucket", "sliding_window"
	BucketSize int
	RefillRate int // tokens per second
	WindowSize time.Duration
	KeyPrefix  string
}

// ProxyConfig holds proxy configuration
type ProxyConfig struct {
	Upstreams       map[string]UpstreamConfig
	LoadBalancer    string // "round_robin", "least_connections", "weighted"
	Timeout         time.Duration
	MaxIdleConns    int
	IdleConnTimeout time.Duration
}

// UpstreamConfig holds configuration for an upstream service
type UpstreamConfig struct {
	URLs        []string
	Weight      int
	HealthCheck HealthCheckConfig
}

// HealthCheckConfig holds health check configuration
type HealthCheckConfig struct {
	Path     string
	Interval time.Duration
	Timeout  time.Duration
}

// ObservabilityConfig holds observability configuration
type ObservabilityConfig struct {
	LogLevel       string
	TracingEnabled bool
	JaegerEndpoint string
	MetricsEnabled bool
	MetricsPath    string
}

// Load loads configuration from environment variables
func Load() (*Config, error) {
	cfg := &Config{}

	// Server config
	port := getEnvInt("SERVER_PORT", 8080)
	cfg.Server.Port = port
	cfg.Server.ReadTimeout = getEnvDuration("SERVER_READ_TIMEOUT", 30*time.Second)
	cfg.Server.WriteTimeout = getEnvDuration("SERVER_WRITE_TIMEOUT", 30*time.Second)
	cfg.Server.IdleTimeout = getEnvDuration("SERVER_IDLE_TIMEOUT", 120*time.Second)

	// Redis config
	cfg.Redis.URL = getEnvString("REDIS_URL", "redis://localhost:6379")
	cfg.Redis.MaxRetries = getEnvInt("REDIS_MAX_RETRIES", 3)
	cfg.Redis.PoolSize = getEnvInt("REDIS_POOL_SIZE", 10)
	cfg.Redis.MinIdleConns = getEnvInt("REDIS_MIN_IDLE_CONNS", 5)
	cfg.Redis.DialTimeout = getEnvDuration("REDIS_DIAL_TIMEOUT", 5*time.Second)
	cfg.Redis.ReadTimeout = getEnvDuration("REDIS_READ_TIMEOUT", 3*time.Second)
	cfg.Redis.WriteTimeout = getEnvDuration("REDIS_WRITE_TIMEOUT", 3*time.Second)
	cfg.Redis.PoolTimeout = getEnvDuration("REDIS_POOL_TIMEOUT", 4*time.Second)
	cfg.Redis.IdleTimeout = getEnvDuration("REDIS_IDLE_TIMEOUT", 5*time.Minute)
	cfg.Redis.IdleCheckFreq = getEnvDuration("REDIS_IDLE_CHECK_FREQ", 1*time.Minute)

	// Auth config
	cfg.Auth.Type = getEnvString("AUTH_TYPE", "both")
	cfg.Auth.JWTSecret = getEnvString("JWT_SECRET", "")
	cfg.Auth.OIDCIssuer = getEnvString("OIDC_ISSUER", "")
	cfg.Auth.OIDCClientID = getEnvString("OIDC_CLIENT_ID", "")
	cfg.Auth.OIDCClientSecret = getEnvString("OIDC_CLIENT_SECRET", "")
	cfg.Auth.SkipAuthPaths = []string{"/health", "/ready", "/metrics"}

	// Rate limit config
	cfg.RateLimit.Enabled = getEnvBool("RATELIMIT_ENABLED", true)
	cfg.RateLimit.Algorithm = getEnvString("RATELIMIT_ALGORITHM", "token_bucket")
	cfg.RateLimit.BucketSize = getEnvInt("RATELIMIT_BUCKET_SIZE", 100)
	cfg.RateLimit.RefillRate = getEnvInt("RATELIMIT_REFILL_RATE", 10)
	cfg.RateLimit.WindowSize = getEnvDuration("RATELIMIT_WINDOW_SIZE", 60*time.Second)
	cfg.RateLimit.KeyPrefix = getEnvString("RATELIMIT_KEY_PREFIX", "ratelimit:")

	// Proxy config
	cfg.Proxy.LoadBalancer = getEnvString("PROXY_LOAD_BALANCER", "round_robin")
	cfg.Proxy.Timeout = getEnvDuration("PROXY_TIMEOUT", 30*time.Second)
	cfg.Proxy.MaxIdleConns = getEnvInt("PROXY_MAX_IDLE_CONNS", 100)
	cfg.Proxy.IdleConnTimeout = getEnvDuration("PROXY_IDLE_CONN_TIMEOUT", 90*time.Second)
	cfg.Proxy.Upstreams = make(map[string]UpstreamConfig)

	// Observability config
	cfg.Observability.LogLevel = getEnvString("LOG_LEVEL", "info")
	cfg.Observability.TracingEnabled = getEnvBool("TRACING_ENABLED", false)
	cfg.Observability.JaegerEndpoint = getEnvString("JAEGER_ENDPOINT", "")
	cfg.Observability.MetricsEnabled = getEnvBool("METRICS_ENABLED", true)
	cfg.Observability.MetricsPath = getEnvString("METRICS_PATH", "/metrics")

	// Validate configuration
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("configuration validation failed: %w", err)
	}

	return cfg, nil
}

// Validate validates the configuration
func (c *Config) Validate() error {
	if c.Server.Port <= 0 || c.Server.Port > 65535 {
		return fmt.Errorf("invalid server port: %d", c.Server.Port)
	}

	if c.Auth.Type != "jwt" && c.Auth.Type != "oidc" && c.Auth.Type != "both" && c.Auth.Type != "mock" {
		return fmt.Errorf("invalid auth type: %s (must be jwt, oidc, both, or mock)", c.Auth.Type)
	}

	if (c.Auth.Type == "jwt" || c.Auth.Type == "both") && c.Auth.JWTSecret == "" {
		return fmt.Errorf("JWT_SECRET is required when AUTH_TYPE is jwt or both")
	}

	if (c.Auth.Type == "oidc" || c.Auth.Type == "both") && c.Auth.OIDCIssuer == "" {
		return fmt.Errorf("OIDC_ISSUER is required when AUTH_TYPE is oidc or both")
	}

	if c.RateLimit.Algorithm != "token_bucket" && c.RateLimit.Algorithm != "leaky_bucket" && c.RateLimit.Algorithm != "sliding_window" {
		return fmt.Errorf("invalid rate limit algorithm: %s (must be token_bucket, leaky_bucket, or sliding_window)", c.RateLimit.Algorithm)
	}

	if c.RateLimit.BucketSize <= 0 {
		return fmt.Errorf("rate limit bucket size must be greater than 0")
	}

	if c.RateLimit.RefillRate <= 0 {
		return fmt.Errorf("rate limit refill rate must be greater than 0")
	}

	return nil
}

// Helper functions for environment variables

func getEnvString(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		intValue, err := strconv.Atoi(value)
		if err != nil {
			return defaultValue
		}
		return intValue
	}
	return defaultValue
}

func getEnvBool(key string, defaultValue bool) bool {
	if value := os.Getenv(key); value != "" {
		boolValue, err := strconv.ParseBool(value)
		if err != nil {
			return defaultValue
		}
		return boolValue
	}
	return defaultValue
}

func getEnvDuration(key string, defaultValue time.Duration) time.Duration {
	if value := os.Getenv(key); value != "" {
		duration, err := time.ParseDuration(value)
		if err != nil {
			return defaultValue
		}
		return duration
	}
	return defaultValue
}

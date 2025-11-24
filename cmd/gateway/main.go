package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ai-api-gateway/internal/auth"
	"ai-api-gateway/internal/config"
	"ai-api-gateway/internal/metrics"
	"ai-api-gateway/internal/middleware"
	"ai-api-gateway/internal/proxy"
	"ai-api-gateway/internal/ratelimiter"
	"ai-api-gateway/internal/tracing"

	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
)

var (
	startTime           = time.Now()
	logger              *config.Logger
	cfg                 *config.Config
	redisClient         *redis.Client
	authMiddleware      *auth.AuthMiddleware
	rateLimitMiddleware *middleware.RateLimitMiddleware
	router              *proxy.Router
)

func main() {
	// Load configuration
	var err error
	cfg, err = config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to load configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	logger = config.NewLogger(cfg.Observability.LogLevel)
	logger.Info("Starting AI API Gateway", map[string]interface{}{
		"version": "1.0.0",
		"port":    cfg.Server.Port,
	})

	// Initialize metrics
	if cfg.Observability.MetricsEnabled {
		metrics.Initialize()
	}

	// Initialize tracing
	if cfg.Observability.TracingEnabled {
		if err := tracing.Initialize("ai-api-gateway", cfg.Observability.JaegerEndpoint); err != nil {
			logger.Warn("Failed to initialize tracing", map[string]interface{}{
				"error": err.Error(),
			})
		} else {
			logger.Info("Tracing initialized", map[string]interface{}{
				"endpoint": cfg.Observability.JaegerEndpoint,
			})
		}
	}

	// Initialize Redis
	if err := initRedis(); err != nil {
		logger.Fatal("Failed to initialize Redis", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize authentication
	if err := initAuth(); err != nil {
		logger.Fatal("Failed to initialize authentication", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize rate limiting
	if err := initRateLimit(); err != nil {
		logger.Fatal("Failed to initialize rate limiting", map[string]interface{}{
			"error": err.Error(),
		})
	}

	// Initialize proxy router
	router = proxy.NewRouter(&cfg.Proxy)

	// Setup HTTP router
	httpRouter := setupRouter()

	// Create HTTP server
	srv := &http.Server{
		Addr:         fmt.Sprintf(":%d", cfg.Server.Port),
		Handler:      httpRouter,
		ReadTimeout:  cfg.Server.ReadTimeout,
		WriteTimeout: cfg.Server.WriteTimeout,
		IdleTimeout:  cfg.Server.IdleTimeout,
	}

	// Start server in goroutine
	go func() {
		logger.Info("Server starting", map[string]interface{}{
			"address": srv.Addr,
		})
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			logger.Fatal("Failed to start server", map[string]interface{}{
				"error": err.Error(),
			})
		}
	}()

	// Wait for interrupt signal for graceful shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server", nil)

	// Graceful shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown", map[string]interface{}{
			"error": err.Error(),
		})
	}

	logger.Info("Server exited", nil)
}

func setupRouter() *gin.Engine {
	// Set Gin mode based on log level
	if cfg.Observability.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Recovery middleware
	router.Use(gin.Recovery())

	// Security headers
	router.Use(middleware.SecurityHeaders())

	// Tracing middleware (if enabled)
	if cfg.Observability.TracingEnabled {
		router.Use(middleware.TracingMiddleware())
	}

	// Request logging middleware
	router.Use(middleware.RequestLogger(logger))

	// Health endpoints (no auth required)
	router.GET("/health", healthHandler)
	router.GET("/ready", readyHandler)

	// Metrics endpoint (no auth required)
	if cfg.Observability.MetricsEnabled {
		router.GET(cfg.Observability.MetricsPath, metricsHandler)
	}

	// API v1 routes with auth and rate limiting
	v1 := router.Group("/v1")

	// Apply rate limiting middleware
	if rateLimitMiddleware != nil {
		v1.Use(rateLimitMiddleware.Middleware())
	}

	// Apply authentication middleware
	if authMiddleware != nil {
		v1.Use(authMiddleware.Middleware())
	}

	{
		v1.Any("/*path", proxyHandler)
	}

	return router
}

func initRedis() error {
	opt, err := redis.ParseURL(cfg.Redis.URL)
	if err != nil {
		return fmt.Errorf("failed to parse Redis URL: %w", err)
	}

	redisClient = redis.NewClient(&redis.Options{
		Addr:               opt.Addr,
		Password:           opt.Password,
		DB:                 opt.DB,
		MaxRetries:         cfg.Redis.MaxRetries,
		PoolSize:           cfg.Redis.PoolSize,
		MinIdleConns:       cfg.Redis.MinIdleConns,
		DialTimeout:        cfg.Redis.DialTimeout,
		ReadTimeout:        cfg.Redis.ReadTimeout,
		WriteTimeout:       cfg.Redis.WriteTimeout,
		PoolTimeout:        cfg.Redis.PoolTimeout,
		IdleTimeout:        cfg.Redis.IdleTimeout,
		IdleCheckFrequency: cfg.Redis.IdleCheckFreq,
	})

	// Test connection
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := redisClient.Ping(ctx).Err(); err != nil {
		return fmt.Errorf("failed to connect to Redis: %w", err)
	}

	logger.Info("Redis connected", map[string]interface{}{
		"url": cfg.Redis.URL,
	})

	return nil
}

func initAuth() error {
	var jwtVerifier *auth.JWTVerifier
	var oidcVerifier *auth.OIDCVerifier
	var err error

	// Initialize JWT verifier if needed
	if cfg.Auth.Type == "jwt" || cfg.Auth.Type == "both" {
		jwtVerifier, err = auth.NewJWTVerifier(cfg.Auth.JWTSecret)
		if err != nil {
			return fmt.Errorf("failed to create JWT verifier: %w", err)
		}
	}

	// Initialize OIDC verifier if needed
	if cfg.Auth.Type == "oidc" || cfg.Auth.Type == "both" {
		oidcConfig := auth.OIDCConfig{
			Issuer:       cfg.Auth.OIDCIssuer,
			ClientID:     cfg.Auth.OIDCClientID,
			ClientSecret: cfg.Auth.OIDCClientSecret,
		}
		oidcVerifier, err = auth.NewOIDCVerifier(context.Background(), oidcConfig)
		if err != nil {
			return fmt.Errorf("failed to create OIDC verifier: %w", err)
		}
	}

	authMiddleware = auth.NewAuthMiddleware(&cfg.Auth, jwtVerifier, oidcVerifier)
	logger.Info("Authentication initialized", map[string]interface{}{
		"type": cfg.Auth.Type,
	})

	return nil
}

func initRateLimit() error {
	if !cfg.RateLimit.Enabled {
		logger.Info("Rate limiting disabled", nil)
		return nil
	}

	if redisClient == nil {
		return fmt.Errorf("Redis client not initialized")
	}

	// Create rate limiter using factory
	factory := ratelimiter.NewFactory(redisClient, &cfg.RateLimit)
	limiter, err := factory.Create()
	if err != nil {
		return fmt.Errorf("failed to create rate limiter: %w", err)
	}

	rateLimitMiddleware = middleware.NewRateLimitMiddleware(limiter, cfg.RateLimit.Enabled, cfg.RateLimit.Algorithm)
	logger.Info("Rate limiting initialized", map[string]interface{}{
		"algorithm":   cfg.RateLimit.Algorithm,
		"bucket_size": cfg.RateLimit.BucketSize,
		"refill_rate": cfg.RateLimit.RefillRate,
	})

	return nil
}

func healthHandler(c *gin.Context) {
	uptime := int(time.Since(startTime).Seconds())
	c.JSON(http.StatusOK, gin.H{
		"status":    "healthy",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
		"version":   "1.0.0",
		"uptime":    uptime,
	})
}

func readyHandler(c *gin.Context) {
	// Check Redis connection
	if redisClient != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			c.JSON(http.StatusServiceUnavailable, gin.H{
				"status":    "not_ready",
				"message":   "Redis connection failed",
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"status":    "ready",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func metricsHandler(c *gin.Context) {
	metrics.Handler(c.Writer, c.Request)
}

func proxyHandler(c *gin.Context) {
	// Parse service and path from request
	path := c.Param("path")
	service, remainingPath, err := proxy.ParseServicePath(path)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "Bad Request",
			"message":   "Invalid path format",
			"code":      http.StatusBadRequest,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Record metrics
	metrics.HTTPRequestTotal.WithLabelValues(c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%d", c.Writer.Status())).Inc()
	start := time.Now()
	defer func() {
		duration := time.Since(start).Seconds()
		metrics.HTTPRequestDuration.WithLabelValues(c.Request.Method, c.Request.URL.Path, fmt.Sprintf("%d", c.Writer.Status())).Observe(duration)
	}()

	// Proxy request to upstream
	router.Proxy(c, service, remainingPath)
}

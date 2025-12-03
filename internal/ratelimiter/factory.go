package ratelimiter

import (
    "fmt"
    "time"

    "ai-api-gateway/internal/config"

    "github.com/go-redis/redis/v8"
)

// Factory creates rate limiter instances based on algorithm type
type Factory struct {
	client    *redis.Client
	keyPrefix string
	config    *config.RateLimitConfig
}

// NewFactory creates a new rate limiter factory
func NewFactory(client *redis.Client, cfg *config.RateLimitConfig) *Factory {
	return &Factory{
		client:    client,
		keyPrefix: cfg.KeyPrefix,
		config:    cfg,
	}
}

// Create creates a rate limiter based on the configured algorithm
func (f *Factory) Create() (RateLimiter, error) {
	switch f.config.Algorithm {
	case "token_bucket":
		return NewTokenBucket(
			f.client,
			f.keyPrefix,
			f.config.BucketSize,
			f.config.RefillRate,
		), nil

	case "leaky_bucket":
		return NewLeakyBucket(
			f.client,
			f.keyPrefix,
			f.config.BucketSize,
			f.config.RefillRate,
		), nil

	case "sliding_window":
		windowSize := f.config.WindowSize
		if windowSize == 0 {
			windowSize = 60 * time.Second // default to 60 seconds
		}
		return NewSlidingWindow(
			f.client,
			f.keyPrefix,
			f.config.BucketSize,
			windowSize,
		), nil

	default:
		return nil, fmt.Errorf("unsupported rate limit algorithm: %s", f.config.Algorithm)
	}
}

// CreateWithAlgorithm creates a rate limiter with a specific algorithm (for testing or dynamic configuration)
func (f *Factory) CreateWithAlgorithm(algorithm string, bucketSize, refillRate int, windowSize time.Duration) (RateLimiter, error) {
	switch algorithm {
	case "token_bucket":
		return NewTokenBucket(
			f.client,
			f.keyPrefix,
			bucketSize,
			refillRate,
		), nil

	case "leaky_bucket":
		return NewLeakyBucket(
			f.client,
			f.keyPrefix,
			bucketSize,
			refillRate,
		), nil

	case "sliding_window":
		if windowSize == 0 {
			windowSize = 60 * time.Second
		}
		return NewSlidingWindow(
			f.client,
			f.keyPrefix,
			bucketSize,
			windowSize,
		), nil

	default:
		return nil, fmt.Errorf("unsupported rate limit algorithm: %s", algorithm)
	}
}


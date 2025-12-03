package ratelimiter

import (
	"context"
)

// RateLimiter defines the interface for rate limiting algorithms
type RateLimiter interface {
	Allow(ctx context.Context, key string) (bool, *LimitInfo, error)
}


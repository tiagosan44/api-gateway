package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// TokenBucket implements token bucket rate limiting algorithm
type TokenBucket struct {
	client     *redis.Client
	keyPrefix  string
	bucketSize int
	refillRate int // tokens per second
}

// NewTokenBucket creates a new token bucket rate limiter
func NewTokenBucket(client *redis.Client, keyPrefix string, bucketSize, refillRate int) *TokenBucket {
	return &TokenBucket{
		client:     client,
		keyPrefix:  keyPrefix,
		bucketSize: bucketSize,
		refillRate: refillRate,
	}
}

// Allow checks if a request is allowed and consumes a token
func (tb *TokenBucket) Allow(ctx context.Context, key string) (bool, *LimitInfo, error) {
	redisKey := fmt.Sprintf("%s:%s", tb.keyPrefix, key)

	// Calculate current time
	now := time.Now().Unix()

	// Lua script for atomic token bucket operation
	script := `
		local key = KEYS[1]
		local bucket_size = tonumber(ARGV[1])
		local refill_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local window = tonumber(ARGV[4])
		
		-- Get current bucket state
		local bucket = redis.call("HMGET", key, "tokens", "last_refill")
		local tokens = tonumber(bucket[1]) or bucket_size
		local last_refill = tonumber(bucket[2]) or now
		
		-- Calculate time elapsed since last refill
		local elapsed = now - last_refill
		
		-- Refill tokens based on elapsed time
		if elapsed > 0 then
			local tokens_to_add = math.floor(elapsed * refill_rate)
			tokens = math.min(bucket_size, tokens + tokens_to_add)
			last_refill = now
		end
		
		-- Check if we can consume a token
		if tokens >= 1 then
			tokens = tokens - 1
			redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
			redis.call("EXPIRE", key, window)
			return {1, tokens, math.ceil((bucket_size - tokens) / refill_rate)}
		else
			-- Calculate time until next token is available
			local wait_time = math.ceil((1 - tokens) / refill_rate)
			redis.call("HMSET", key, "tokens", tokens, "last_refill", last_refill)
			redis.call("EXPIRE", key, window)
			return {0, tokens, wait_time}
		end
	`

	window := int(time.Duration(tb.bucketSize) * time.Second / time.Duration(tb.refillRate))
	result, err := tb.client.Eval(ctx, script, []string{redisKey},
		tb.bucketSize,
		tb.refillRate,
		now,
		window,
	).Result()

	if err != nil {
		return false, nil, fmt.Errorf("redis error: %w", err)
	}

	// Parse result
	results, ok := result.([]interface{})
	if !ok || len(results) != 3 {
		return false, nil, fmt.Errorf("unexpected result format")
	}

	allowed := results[0].(int64) == 1
	remaining := int(results[1].(int64))
	resetIn := int(results[2].(int64))

	limitInfo := &LimitInfo{
		Allowed:   allowed,
		Remaining: remaining,
		Limit:     tb.bucketSize,
		ResetIn:   time.Duration(resetIn) * time.Second,
	}

	return allowed, limitInfo, nil
}

// LimitInfo contains rate limit information
type LimitInfo struct {
	Allowed   bool
	Remaining int
	Limit     int
	ResetIn   time.Duration
}

// GetHeaders returns rate limit headers
func (li *LimitInfo) GetHeaders() map[string]string {
	headers := make(map[string]string)
	headers["X-RateLimit-Limit"] = fmt.Sprintf("%d", li.Limit)
	headers["X-RateLimit-Remaining"] = fmt.Sprintf("%d", li.Remaining)
	headers["X-RateLimit-Reset"] = fmt.Sprintf("%d", time.Now().Add(li.ResetIn).Unix())
	if !li.Allowed {
		headers["Retry-After"] = fmt.Sprintf("%d", int(li.ResetIn.Seconds()))
	}
	return headers
}


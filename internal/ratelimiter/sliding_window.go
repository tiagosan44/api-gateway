package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// SlidingWindow implements sliding window rate limiting algorithm
type SlidingWindow struct {
	client     *redis.Client
	keyPrefix  string
	limit      int
	windowSize time.Duration
}

// NewSlidingWindow creates a new sliding window rate limiter
func NewSlidingWindow(client *redis.Client, keyPrefix string, limit int, windowSize time.Duration) *SlidingWindow {
	return &SlidingWindow{
		client:     client,
		keyPrefix:  keyPrefix,
		limit:      limit,
		windowSize: windowSize,
	}
}

// Allow checks if a request is allowed within the sliding window
func (sw *SlidingWindow) Allow(ctx context.Context, key string) (bool, *LimitInfo, error) {
	redisKey := fmt.Sprintf("%s:%s", sw.keyPrefix, key)

	// Calculate current time and window boundaries
	now := time.Now()
	windowStart := now.Add(-sw.windowSize).Unix()
	windowEnd := now.Unix()

	// Lua script for atomic sliding window operation
	script := `
		local key = KEYS[1]
		local limit = tonumber(ARGV[1])
		local window_start = tonumber(ARGV[2])
		local window_end = tonumber(ARGV[3])
		local window_size = tonumber(ARGV[4])
		
		-- Remove old entries outside the window
		redis.call("ZREMRANGEBYSCORE", key, 0, window_start)
		
		-- Count current requests in window
		local current = redis.call("ZCARD", key)
		
		-- Check if we can accept a new request
		if current < limit then
			-- Add current request with timestamp as score
			redis.call("ZADD", key, window_end, window_end)
			redis.call("EXPIRE", key, window_size)
			return {1, limit - current - 1, window_size}
		else
			-- Get oldest request timestamp to calculate wait time
			local oldest = redis.call("ZRANGE", key, 0, 0, "WITHSCORES")
			local wait_time = 0
			if oldest and #oldest > 0 then
				local oldest_time = tonumber(oldest[2])
				wait_time = math.ceil(window_start + window_size - oldest_time)
			end
			return {0, 0, wait_time}
		end
	`

	windowSizeSeconds := int(sw.windowSize.Seconds())
	result, err := sw.client.Eval(ctx, script, []string{redisKey},
		sw.limit,
		windowStart,
		windowEnd,
		windowSizeSeconds,
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
		Limit:     sw.limit,
		ResetIn:   time.Duration(resetIn) * time.Second,
	}

	return allowed, limitInfo, nil
}


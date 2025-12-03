package ratelimiter

import (
	"context"
	"fmt"
	"time"

	"github.com/go-redis/redis/v8"
)

// LeakyBucket implements leaky bucket rate limiting algorithm
type LeakyBucket struct {
	client     *redis.Client
	keyPrefix  string
	capacity   int
	leakRate   int // requests per second
}

// NewLeakyBucket creates a new leaky bucket rate limiter
func NewLeakyBucket(client *redis.Client, keyPrefix string, capacity, leakRate int) *LeakyBucket {
	return &LeakyBucket{
		client:    client,
		keyPrefix: keyPrefix,
		capacity:  capacity,
		leakRate:  leakRate,
	}
}

// Allow checks if a request is allowed and processes it
func (lb *LeakyBucket) Allow(ctx context.Context, key string) (bool, *LimitInfo, error) {
	redisKey := fmt.Sprintf("%s:%s", lb.keyPrefix, key)

	// Calculate current time
	now := time.Now().Unix()

	// Lua script for atomic leaky bucket operation
 script := `
		local key = KEYS[1]
		local capacity = tonumber(ARGV[1])
		local leak_rate = tonumber(ARGV[2])
		local now = tonumber(ARGV[3])
		local window = tonumber(ARGV[4])
		
		-- Get current bucket state
		local bucket = redis.call("HMGET", key, "level", "last_leak")
		local level = tonumber(bucket[1]) or 0
		local last_leak = tonumber(bucket[2]) or now
		
		-- Calculate time elapsed since last leak
		local elapsed = now - last_leak
		
		-- Leak requests based on elapsed time
		if elapsed > 0 then
			local leaked = math.floor(elapsed * leak_rate)
			level = math.max(0, level - leaked)
			last_leak = now
		end
		
		-- Check if we can accept a new request
		if level < capacity then
			level = level + 1
			redis.call("HMSET", key, "level", level, "last_leak", last_leak)
			redis.call("EXPIRE", key, window)
			return {1, capacity - level, math.ceil((capacity - level) / leak_rate)}
		else
			-- Calculate time until bucket has space
			local wait_time = math.ceil((level - capacity + 1) / leak_rate)
			redis.call("HMSET", key, "level", level, "last_leak", last_leak)
			redis.call("EXPIRE", key, window)
			return {0, capacity - level, wait_time}
		end
	`

 // Compute a TTL window based on bucket capacity and leak rate.
 // Prevent division by zero if leakRate is misconfigured.
 window := 1
 if lb.leakRate > 0 {
     window = int(time.Duration(lb.capacity) * time.Second / time.Duration(lb.leakRate))
     if window <= 0 {
         window = 1
     }
 }
 result, err := lb.client.Eval(ctx, script, []string{redisKey},
     lb.capacity,
     lb.leakRate,
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
		Limit:     lb.capacity,
		ResetIn:   time.Duration(resetIn) * time.Second,
	}

	return allowed, limitInfo, nil
}


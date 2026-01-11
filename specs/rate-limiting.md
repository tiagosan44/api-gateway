# Rate Limiting Specification

## Goal
Protect the API Gateway and upstream services from abuse, overload, and denial of service attacks by limiting the number of requests per client within a time window. Ensure fair usage across all clients while maintaining system stability.

## Algorithm
- Token Bucket
- Backed by Redis + Lua scripts for atomic operations

## Key Strategy
`{client_id}:{endpoint}`

Examples:
- Anonymous users: `ip:192.168.1.1:/v1/products`
- Authenticated users: `user:abc-123:/v1/products/{id}`
- Admin operations: `user:admin-456:/v1/products/{id}`

## Limits per Endpoint

### Public Endpoints (No Auth Required)
- `GET /v1/products` - 100 requests per minute (burst: 20)
- `GET /v1/products/{id}` - 200 requests per minute (burst: 50)
- `GET /v1/categories` - 50 requests per minute (burst: 10)
- `GET /v1/search` - 30 requests per minute (burst: 10)
- `GET /v1/products/{id}/inventory` - 100 requests per minute (burst: 20)

### Authenticated Endpoints
- `PUT /v1/products/{id}` - 20 requests per minute (burst: 5)
- `DELETE /v1/products/{id}` - 10 requests per minute (burst: 2)
- `PATCH /v1/products/{id}/inventory` - 30 requests per minute (burst: 10)

## Identifier Strategy
1. If JWT present and valid → extract `sub` claim as client_id
2. If no JWT or invalid → use IP address
3. Admin users → higher limits (configurable)

## Failure Mode
- Redis unavailable → FAIL-OPEN (allow traffic, log warning + metric)
- Degraded mode: Local rate limiter (approximate, not distributed)

## Configuration
### Redis Configuration
- **Host**: Configurable via environment variable `REDIS_HOST`
- **Port**: Default 6379
- **Connection pool**: Min 5, Max 20 connections
- **Timeout**: 200ms for rate limit checks

### Endpoint Limits
- Configured per endpoint via application.yml
- Default fallback: 60 requests per minute
- Burst capacity: 20% of limit

## Headers
- `X-RateLimit-Limit` - Maximum requests allowed in window
- `X-RateLimit-Remaining` - Requests remaining in current window
- `X-RateLimit-Reset` - Unix timestamp when the rate limit resets

## Response on Limit Exceeded
- HTTP 429 Too Many Requests
- Body: `{"error": "Rate limit exceeded", "retryAfter": <seconds>}`
- Header: `Retry-After: <seconds>`
# Caching Specification

## Goal
Reduce latency and load on upstream services by caching frequently accessed data. Improve API response times while ensuring data consistency through intelligent cache invalidation strategies.

## Algorithm
Cache-Aside pattern with TTL-based expiration and event-driven invalidation. On cache miss, fetch from upstream, store in cache, and return to client.

## Cache Strategy
- **Implementation**: Redis for distributed caching
- **Serialization**: JSON
- **Eviction Policy**: LRU (Least Recently Used)

## Configuration

### Redis Configuration
- **Host**: Configurable via environment variable `REDIS_CACHE_HOST`
- **Port**: Default 6379
- **Database**: 1 (separate from rate limiter)
- **Max memory**: 512 MB
- **Connection pool**: Min 10, Max 50 connections

## Cache Entries

### Products Cache
- **Key Pattern**: `cache:product:{productId}`
- **TTL**: 5 minutes
- **Use Cases**:
  - `GET /v1/products/{id}` - Individual product details
- **Cache on**: First successful GET request
- **Invalidate on**:
  - `PUT /v1/products/{id}` - Product update
  - `DELETE /v1/products/{id}` - Product deletion

### Product List Cache
- **Key Pattern**: `cache:products:list:{page}:{size}:{category}:{minPrice}:{maxPrice}:{search}`
- **TTL**: 2 minutes (shorter due to frequent changes)
- **Use Cases**:
  - `GET /v1/products` with query parameters
- **Invalidate on**:
  - Any product update/delete operation (clear all list caches)
  - Scheduled invalidation every 2 minutes

### Inventory Cache
- **Key Pattern**: `cache:inventory:{productId}`
- **TTL**: 30 seconds (must be fresh due to stock changes)
- **Use Cases**:
  - `GET /v1/products/{id}/inventory`
- **Cache on**: First successful GET request
- **Invalidate on**:
  - `PATCH /v1/products/{id}/inventory` - Inventory update

### Categories Cache
- **Key Pattern**: `cache:categories:all`
- **TTL**: 15 minutes (categories change infrequently)
- **Use Cases**:
  - `GET /v1/categories`
- **Cache on**: First successful GET request
- **Invalidate on**:
  - Manual invalidation or category changes in upstream

### Search Results Cache
- **Key Pattern**: `cache:search:{query}:{page}:{size}`
- **TTL**: 1 minute
- **Use Cases**:
  - `GET /v1/search?q={query}`
- **Cache on**: First successful search
- **Invalidate on**:
  - Any product update/delete (clear all search caches)

## Invalidation Strategy

### Immediate Invalidation
Triggered synchronously on write operations:
```
PUT /v1/products/{id}:
  1. Execute upstream request
  2. On success (2xx), invalidate:
     - cache:product:{id}
     - cache:products:list:* (all list variations)
     - cache:search:* (all search queries)

DELETE /v1/products/{id}:
  1. Execute upstream request
  2. On success (2xx), invalidate:
     - cache:product:{id}
     - cache:products:list:*
     - cache:search:*

PATCH /v1/products/{id}/inventory:
  1. Execute upstream request
  2. On success (2xx), invalidate:
     - cache:inventory:{id}
```

### Pattern-Based Invalidation
Use Redis `KEYS` command (or `SCAN` in production) for wildcard invalidation:
- `cache:products:list:*` - All product list variations
- `cache:search:*` - All search results

## Cache Headers
Response headers to indicate cache status:
- `X-Cache-Status: HIT` - Served from cache
- `X-Cache-Status: MISS` - Fetched from upstream and cached
- `Cache-Control: private, max-age=300` - For client-side caching hints

## Failure Mode
- **Redis unavailable**: FAIL-THROUGH
  - Skip caching
  - Forward all requests directly to upstream
  - Log warning + emit metric
  - Continue normal operation

## Cache Warming (Future Enhancement)
Pre-populate cache for frequently accessed data:
- Top 100 products by popularity
- All categories
- Popular search queries


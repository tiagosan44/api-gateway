# Load Balancing Specification

## Goal
Distribute incoming requests across multiple upstream service instances to achieve high availability, fault tolerance, and optimal resource utilization. Ensure that unhealthy instances are automatically excluded from the rotation and traffic is only routed to healthy backends.

## Algorithm
Round Robin with health-aware routing. Each request is sent to the next available healthy instance in a circular order.

## Upstream Services
The gateway proxies requests to backend services:
- Product Service (manages products and categories)
- Inventory Service (manages stock levels)

## Configuration

### Upstream Configuration
```yaml
product-service:
  instances:
    - http://product-service-1:8081
    - http://product-service-2:8081
    - http://product-service-3:8081
  endpoints:
    - GET /v1/products
    - GET /v1/products/{id}
    - PUT /v1/products/{id}
    - DELETE /v1/products/{id}
    - GET /v1/categories
    - GET /v1/search

inventory-service:
  instances:
    - http://inventory-service-1:8082
    - http://inventory-service-2:8082
  endpoints:
    - GET /v1/products/{id}/inventory
    - PATCH /v1/products/{id}/inventory
```

## Health Checks
- **Interval**: 10 seconds
- **Timeout**: 2 seconds
- **Failure threshold**: 3 consecutive failures → mark unhealthy
- **Success threshold**: 2 consecutive successes → mark healthy
- **Health endpoint**: `GET /actuator/health` on each upstream

## Circuit Breaker
- **Failure rate threshold**: 50% in 10-second window
- **Open duration**: 30 seconds
- **Half-open requests**: 3 test requests before fully closing

## Retry Policy
- **Max retries**: 2 attempts (3 total including original)
- **Retry conditions**:
  - Connection timeout
  - Connection refused
  - HTTP 503 Service Unavailable
- **Non-retryable**:
  - HTTP 4xx (except 429)
  - Successful responses
  - Write operations (PUT, PATCH, DELETE) - avoid duplicate writes
- **Backoff**: Exponential (100ms, 200ms)

## Failure Mode
- If all instances unhealthy → HTTP 503 Service Unavailable
- Connection timeout → Retry next instance (up to max retries)
- Circuit breaker OPEN → Skip instance, try next available
- Metric: `upstream_health_status{service, instance}`
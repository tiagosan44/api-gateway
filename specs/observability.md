# Observability Specification

## Goal
Provide complete visibility into system behavior, performance, and health through metrics, logs, and traces. Enable proactive monitoring, quick troubleshooting, and data-driven optimization decisions.

## Metrics (Prometheus)

### HTTP Metrics
```
# Request counters
http_requests_total{method, endpoint, status_code}
http_requests_duration_seconds{method, endpoint, status_code}

# Example labels:
# method="GET", endpoint="/v1/products", status_code="200"
```

### Rate Limiting Metrics
```
rate_limit_requests_total{endpoint, client_type, result}
# client_type: anonymous, authenticated, admin
# result: allowed, rejected

rate_limit_remaining_tokens{endpoint, client_id}
```

### Cache Metrics
```
cache_requests_total{cache_name, result}
# cache_name: products, inventory, categories
# result: hit, miss

cache_invalidations_total{cache_name, reason}
# reason: product_updated, product_deleted, inventory_updated

cache_size_bytes{cache_name}
cache_evictions_total{cache_name}
```

### Upstream/Load Balancer Metrics
```
upstream_requests_total{service, instance, status_code}
upstream_requests_duration_seconds{service, instance}
upstream_health_status{service, instance}
# 1 = healthy, 0 = unhealthy

upstream_circuit_breaker_state{service}
# 0 = closed, 1 = open, 2 = half_open

upstream_retries_total{service, reason}
```

### JVM Metrics (Spring Boot Actuator)
```
jvm_memory_used_bytes
jvm_gc_pause_seconds
process_cpu_usage
```

## Tracing (Optional for v1)
- **Standard**: W3C Trace Context
- **Headers**:
  - `traceparent`: version-trace_id-span_id-flags
  - `X-Correlation-ID`: Generated if not provided by client
- **Propagation**: Forward to all upstream services
- **Integration**: OpenTelemetry or Spring Cloud Sleuth

## Logging

### Format
- **Type**: JSON structured logs
- **Level**: INFO for normal operations, WARN for degraded, ERROR for failures

### Log Fields
```json
{
  "timestamp": "2024-01-15T10:30:45.123Z",
  "level": "INFO",
  "correlation_id": "abc-123-def",
  "service": "api-gateway",
  "method": "GET",
  "endpoint": "/v1/products",
  "status_code": 200,
  "duration_ms": 45,
  "client_id": "user:xyz-789",
  "upstream_service": "product-service",
  "upstream_instance": "product-service-2:8081",
  "cache_status": "HIT"
}
```

### Important Events to Log
- Rate limit exceeded (WARN)
- Upstream service failures (ERROR)
- Circuit breaker state changes (WARN)
- Cache invalidations (INFO)
- Redis connection failures (ERROR)
- Health check failures (WARN)

## Dashboards (Grafana)

### Dashboard 1: Request Overview
- Request rate (requests/sec)
- P50, P95, P99 latency
- Error rate by endpoint
- Status code distribution

### Dashboard 2: Rate Limiting
- Requests allowed vs rejected
- Top rate-limited clients
- Rate limit remaining by endpoint

### Dashboard 3: Caching
- Cache hit ratio (%)
- Cache operations/sec
- Invalidations by reason

### Dashboard 4: Upstream Health
- Upstream response times
- Circuit breaker states
- Retry counts
- Healthy vs unhealthy instances

## Configuration
### Prometheus
- **Endpoint**: `/actuator/prometheus`
- **Scrape interval**: 15 seconds
- **Retention**: 15 days

### Logging
- **Format**: JSON
- **Output**: stdout (captured by container runtime)
- **Log level**: INFO (configurable via `LOGGING_LEVEL`)

### Grafana Dashboards
- Import pre-configured dashboards from `/monitoring/grafana-dashboards/`
- Auto-provisioned on startup

## Alerts
- High error rate (>5% 5xx in 5 minutes)
- High latency (P99 > 1s for 5 minutes)
- All upstreams down
- Redis connection lost
- Circuit breaker open for >2 minutes

## Failure Mode
- Prometheus unavailable → Metrics still collected in memory (limited buffer)
- Logging errors → Fallback to console logging (non-JSON)
- Never block requests due to observability failures
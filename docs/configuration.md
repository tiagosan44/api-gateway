# Configuration Guide

## Environment Variables

### Server Configuration

- `SERVER_PORT` (default: 8080) - Port the gateway listens on
- `SERVER_READ_TIMEOUT` (default: 30s) - Read timeout
- `SERVER_WRITE_TIMEOUT` (default: 30s) - Write timeout
- `SERVER_IDLE_TIMEOUT` (default: 120s) - Idle connection timeout

### Redis Configuration

- `REDIS_URL` (required) - Redis connection URL
- `REDIS_MAX_RETRIES` (default: 3) - Maximum retry attempts
- `REDIS_POOL_SIZE` (default: 10) - Connection pool size
- `REDIS_MIN_IDLE_CONNS` (default: 5) - Minimum idle connections
- `REDIS_DIAL_TIMEOUT` (default: 5s) - Dial timeout
- `REDIS_READ_TIMEOUT` (default: 3s) - Read timeout
- `REDIS_WRITE_TIMEOUT` (default: 3s) - Write timeout
- `REDIS_POOL_TIMEOUT` (default: 4s) - Pool timeout
- `REDIS_IDLE_TIMEOUT` (default: 5m) - Idle connection timeout

### Authentication Configuration

- `AUTH_TYPE` (required: jwt|oidc|both|mock) - Authentication type
- `JWT_SECRET` (required if AUTH_TYPE is jwt or both) - JWT signing secret
- `OIDC_ISSUER` (required if AUTH_TYPE is oidc or both) - OIDC issuer URL
- `OIDC_CLIENT_ID` (optional) - OIDC client ID
- `OIDC_CLIENT_SECRET` (optional) - OIDC client secret

### Rate Limiting Configuration

- `RATELIMIT_ENABLED` (default: true) - Enable rate limiting
- `RATELIMIT_ALGORITHM` (default: token_bucket) - Algorithm: token_bucket, leaky_bucket, sliding_window
- `RATELIMIT_BUCKET_SIZE` (default: 100) - Bucket size / limit
- `RATELIMIT_REFILL_RATE` (default: 10) - Refill rate (tokens/requests per second)
- `RATELIMIT_WINDOW_SIZE` (default: 60s) - Window size for sliding window
- `RATELIMIT_KEY_PREFIX` (default: ratelimit:) - Redis key prefix

### Proxy Configuration

- `PROXY_LOAD_BALANCER` (default: round_robin) - Strategy: round_robin, least_connections, weighted
- `PROXY_TIMEOUT` (default: 30s) - Upstream request timeout
- `PROXY_MAX_IDLE_CONNS` (default: 100) - Maximum idle connections
- `PROXY_IDLE_CONN_TIMEOUT` (default: 90s) - Idle connection timeout

### Observability Configuration

- `LOG_LEVEL` (default: info) - Log level: debug, info, warn, error
- `TRACING_ENABLED` (default: false) - Enable OpenTelemetry tracing
- `JAEGER_ENDPOINT` (required if TRACING_ENABLED=true) - Jaeger endpoint URL
- `METRICS_ENABLED` (default: true) - Enable Prometheus metrics
- `METRICS_PATH` (default: /metrics) - Metrics endpoint path

## Example Configuration

```bash
# Server
SERVER_PORT=8080
SERVER_READ_TIMEOUT=30s
SERVER_WRITE_TIMEOUT=30s

# Redis
REDIS_URL=redis://localhost:6379
REDIS_POOL_SIZE=20

# Authentication
AUTH_TYPE=both
JWT_SECRET=your-secret-key-here
OIDC_ISSUER=https://auth.example.com

# Rate Limiting
RATELIMIT_ENABLED=true
RATELIMIT_ALGORITHM=token_bucket
RATELIMIT_BUCKET_SIZE=100
RATELIMIT_REFILL_RATE=10

# Observability
LOG_LEVEL=info
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
METRICS_ENABLED=true
```

## Kubernetes Configuration

Configuration in Kubernetes is done via ConfigMap and Secrets:

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: ai-api-gateway-config
data:
  SERVER_PORT: "8080"
  REDIS_URL: "redis://redis-service:6379"
  AUTH_TYPE: "both"
  RATELIMIT_ALGORITHM: "token_bucket"
  LOG_LEVEL: "info"

---
apiVersion: v1
kind: Secret
metadata:
  name: ai-api-gateway-secrets
type: Opaque
stringData:
  JWT_SECRET: "your-secret-key"
  OIDC_CLIENT_SECRET: "client-secret"
```


# AI API Gateway

High-performance edge gateway in Go providing secure ingress, distributed rate limiting, adaptive load balancing, and full observability for AI microservices.

## Features

- **Authentication**: Dual authentication support (JWT and OIDC)
- **Rate Limiting**: Multiple algorithms (Token Bucket, Leaky Bucket, Sliding Window) with Redis
- **Load Balancing**: Round-robin, least connections, and weighted strategies
- **Observability**: Prometheus metrics, OpenTelemetry tracing, structured logging
- **Kubernetes Ready**: Helm charts and Kubernetes manifests included

## Quick Start

### Local Development

1. Start dependencies with Docker Compose:

```bash
docker-compose up -d
```

2. Run the gateway:

```bash
go run ./cmd/gateway
```

3. Test the health endpoint:

```bash
curl http://localhost:8080/health
```

### Configuration

The gateway is configured via environment variables:

```bash
SERVER_PORT=8080
REDIS_URL=redis://localhost:6379
AUTH_TYPE=both
JWT_SECRET=your-secret-key
OIDC_ISSUER=https://example.com
RATELIMIT_ALGORITHM=token_bucket
RATELIMIT_BUCKET_SIZE=100
RATELIMIT_REFILL_RATE=10
LOG_LEVEL=info
```

See [Configuration Guide](docs/configuration.md) for all available options.

## Architecture

```
Clients → Global LB → Ingress (Envoy) → AI API Gateway → Upstream Services
                                              ↓
                                    Redis (Rate Limiting)
                                              ↓
                                    Prometheus + Jaeger
```

## Deployment

### Kubernetes with Helm

```bash
helm install ai-api-gateway ./deploy/helm/ai-api-gateway \
  --set config.auth.jwtSecret=your-secret \
  --set config.redis.url=redis://redis-service:6379
```

### Docker

```bash
docker build -t ai-api-gateway:latest .
docker run -p 8080:8080 \
  -e REDIS_URL=redis://redis:6379 \
  -e AUTH_TYPE=both \
  ai-api-gateway:latest
```

## API Endpoints

- `GET /health` - Health check
- `GET /ready` - Readiness check
- `GET /metrics` - Prometheus metrics
- `GET /v1/{service}/{path}` - Proxy to upstream service

## Rate Limiting

The gateway supports three rate limiting algorithms:

1. **Token Bucket**: Classic token bucket with configurable refill rate
2. **Leaky Bucket**: Leaky bucket algorithm for smooth rate limiting
3. **Sliding Window**: Sliding window counter for precise rate limiting

## Load Balancing

Three load balancing strategies are available:

1. **Round Robin**: Distributes requests evenly across upstreams
2. **Least Connections**: Routes to upstream with fewest active connections
3. **Weighted**: Weighted round-robin based on configured weights

## Observability

### Metrics

Prometheus metrics are exposed at `/metrics`:

- `http_requests_total` - Total HTTP requests
- `http_request_duration_seconds` - Request latency
- `rate_limit_hits_total` - Rate limit hits
- `auth_failures_total` - Authentication failures
- `upstream_requests_total` - Upstream service requests

### Tracing

OpenTelemetry tracing is supported with Jaeger export:

```bash
TRACING_ENABLED=true
JAEGER_ENDPOINT=http://jaeger:14268/api/traces
```

## Development

### Prerequisites

- Go 1.21+
- Docker and Docker Compose
- Redis (for rate limiting)

### Building

```bash
go build -o gateway ./cmd/gateway
```

### Testing

```bash
go test ./...
```

### Running Tests

```bash
docker-compose -f docker-compose.test.yml up
go test ./tests/integration/...
```

## Documentation

- [Configuration Guide](docs/configuration.md)
- [Runbook](docs/runbook.md)
- [Troubleshooting](docs/troubleshooting.md)
- [API Specification](api/openapi.yaml)

## License

MIT


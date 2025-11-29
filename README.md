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

### Load testing with k6

You can run the load test located at `tests/load/k6_test.js` in a few different ways.

Prerequisites:
- The gateway must be running and expose `/health` and `/metrics` (default http://localhost:8080 when using `docker-compose up -d`).
- Optionally, set `BASE_URL` to point the test at a different host/port.

Option A — Use local k6 CLI (recommended):
1. Install k6: https://k6.io/docs/get-started/installation/
2. Start the stack (if not already running):
   ```bash
   docker-compose up -d
   ```
3. Run the test against the default gateway on port 8080:
   ```bash
   k6 run tests/load/k6_test.js
   ```
   Or explicitly set the target URL:
   ```bash
   BASE_URL=http://localhost:8080 k6 run tests/load/k6_test.js
   ```

Option B — Use the test compose profile (AUTH_TYPE=mock):
1. Start the test stack:
   ```bash
   docker-compose -f docker-compose.test.yml up -d
   ```
   This exposes the gateway on port 8082.
2. Run the test pointing to port 8082:
   ```bash
   BASE_URL=http://localhost:8082 k6 run tests/load/k6_test.js
   ```

Option C — Run k6 via Docker (no local install):
```bash
docker run --rm -i \
  -e BASE_URL=http://host.docker.internal:8080 \
  -v "$PWD":/work -w /work \
  grafana/k6:latest run tests/load/k6_test.js
```
On Linux, replace `host.docker.internal` with your Docker host IP (often `172.17.0.1`) or the host’s LAN IP.

Notes:
- The test stages ramp to 50 virtual users and assert p(95) < 500ms with error rate < 10%.
- You can tweak stages and thresholds inside `tests/load/k6_test.js`.
- If you see 401/403 responses, ensure your `AUTH_TYPE` and related env vars are compatible with your environment. The default `docker-compose.yml` uses `AUTH_TYPE=both` with a dev JWT secret and mock OIDC issuer; the test compose uses `AUTH_TYPE=mock`.

## Documentation

- [Configuration Guide](docs/configuration.md)
- [Runbook](docs/runbook.md)
- [Troubleshooting](docs/troubleshooting.md)
- [API Specification](api/openapi.yaml)
- [Helm: configuración de autenticación](docs/helm_auth.md)

## License

MIT


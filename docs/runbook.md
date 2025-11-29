# AI API Gateway - Runbook

## Overview

This runbook provides operational procedures for the AI API Gateway.

## Health Checks

### Check Gateway Health

```bash
curl http://localhost:8080/health
```

Expected response:
```json
{
  "status": "healthy",
  "timestamp": "2024-01-01T00:00:00Z",
  "version": "1.0.0",
  "uptime": 3600
}
```

### Check Readiness

```bash
curl http://localhost:8080/ready
```

## Monitoring

### Metrics

Prometheus metrics are available at `/metrics`:

```bash
curl http://localhost:8080/metrics
```

Key metrics to monitor:
- `http_requests_total` - Total request count
- `http_request_duration_seconds` - Request latency
- `rate_limit_hits_total` - Rate limit violations
- `auth_failures_total` - Authentication failures
- `upstream_requests_total` - Upstream service requests

### Alerts

Configure alerts for:
- High error rate (> 5%)
- High latency (p95 > 1s)
- Rate limit hits > 100/min
- Upstream service failures

## Troubleshooting

### Gateway Not Starting

1. Check logs:
```bash
kubectl logs -f deployment/ai-api-gateway
```

2. Verify configuration:
```bash
kubectl get configmap ai-api-gateway-config -o yaml
```

3. Check Redis connectivity:
```bash
kubectl exec -it deployment/ai-api-gateway -- redis-cli -h redis-service ping
```

### High Error Rate

1. Check upstream services:
```bash
curl http://upstream-service/health
```

2. Review rate limiting:
- Check Redis for rate limit keys
- Verify rate limit configuration

3. Check authentication:
- Verify JWT/OIDC configuration
- Check token expiration

### Performance Issues

1. Check resource usage:
```bash
kubectl top pods -l app=ai-api-gateway
```

2. Review metrics for bottlenecks:
- High latency endpoints
- Rate limiting impact
- Upstream service delays

## Scaling

### Horizontal Scaling

Update HPA:
```bash
kubectl autoscale deployment ai-api-gateway --min=2 --max=10 --cpu-percent=80
```

### Vertical Scaling

Update resources in values.yaml:
```yaml
resources:
  limits:
    cpu: 1000m
    memory: 1Gi
  requests:
    cpu: 500m
    memory: 512Mi
```

## Backup and Recovery

### Configuration Backup

```bash
kubectl get configmap ai-api-gateway-config -o yaml > config-backup.yaml
kubectl get secret ai-api-gateway-secrets -o yaml > secrets-backup.yaml
```

### Recovery

```bash
kubectl apply -f config-backup.yaml
kubectl apply -f secrets-backup.yaml
kubectl rollout restart deployment/ai-api-gateway
```

## Maintenance

### Rolling Update

```bash
kubectl set image deployment/ai-api-gateway \
  ai-api-gateway=ai-api-gateway:new-version
```

### Rollback

```bash
kubectl rollout undo deployment/ai-api-gateway
```


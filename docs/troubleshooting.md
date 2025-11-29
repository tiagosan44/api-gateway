# Troubleshooting Guide

## Common Issues

### 1. Gateway Fails to Start

**Symptoms**: Gateway pod in CrashLoopBackOff

**Causes**:
- Invalid configuration
- Redis connection failure
- Missing environment variables

**Solutions**:
1. Check logs: `kubectl logs deployment/ai-api-gateway`
2. Verify Redis is accessible
3. Check all required environment variables are set

### 2. Authentication Failures

**Symptoms**: 401 Unauthorized responses

**Causes**:
- Invalid JWT secret
- OIDC issuer unreachable
- Token expiration

**Solutions**:
1. Verify JWT_SECRET matches token signing key
2. Check OIDC issuer is accessible
3. Verify token hasn't expired

### 3. Rate Limiting Not Working

**Symptoms**: No rate limit enforcement

**Causes**:
- Redis connection issues
- Rate limiting disabled
- Invalid algorithm configuration

**Solutions**:
1. Check Redis connectivity
2. Verify RATELIMIT_ENABLED=true
3. Check algorithm is valid (token_bucket, leaky_bucket, sliding_window)

### 4. High Latency

**Symptoms**: Slow response times

**Causes**:
- Upstream service delays
- Redis latency
- Resource constraints

**Solutions**:
1. Check upstream service health
2. Monitor Redis performance
3. Scale gateway pods
4. Review resource limits

### 5. Upstream Service Errors

**Symptoms**: 502 Bad Gateway

**Causes**:
- Upstream service down
- Network issues
- Timeout configuration

**Solutions**:
1. Verify upstream service is running
2. Check network connectivity
3. Increase timeout if needed
4. Review load balancing configuration

## Debugging Commands

### Check Gateway Status
```bash
kubectl get pods -l app=ai-api-gateway
kubectl describe pod <pod-name>
```

### View Logs
```bash
kubectl logs -f deployment/ai-api-gateway
```

### Test Redis Connection
```bash
kubectl exec -it deployment/ai-api-gateway -- redis-cli -h redis-service ping
```

### Check Metrics
```bash
curl http://localhost:8080/metrics | grep rate_limit
```

### Test Health Endpoints
```bash
curl http://localhost:8080/health
curl http://localhost:8080/ready
```

## Performance Tuning

### Redis Connection Pool
Increase pool size if seeing connection errors:
```yaml
REDIS_POOL_SIZE=20
REDIS_MIN_IDLE_CONNS=10
```

### Rate Limiting
Adjust bucket size and refill rate based on traffic:
```yaml
RATELIMIT_BUCKET_SIZE=200
RATELIMIT_REFILL_RATE=20
```

### Timeouts
Adjust timeouts for slow upstreams:
```yaml
PROXY_TIMEOUT=60s
SERVER_READ_TIMEOUT=60s
SERVER_WRITE_TIMEOUT=60s
```


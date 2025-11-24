package proxy

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"ai-api-gateway/internal/config"
	"ai-api-gateway/internal/metrics"

	"github.com/gin-gonic/gin"
)

// Router handles request routing to upstream services
type Router struct {
	upstreams         map[string]*Upstream
	config            *config.ProxyConfig
	client            *http.Client
	connTracker       *ConnectionTracker
	weightedBalancers map[string]*WeightedRoundRobin
}

// Upstream represents an upstream service
type Upstream struct {
	Name    string
	URLs    []string
	Weights []int
	Current int
	Health  *HealthChecker
}

// NewRouter creates a new router
func NewRouter(cfg *config.ProxyConfig) *Router {
	transport := &http.Transport{
		MaxIdleConns:    cfg.MaxIdleConns,
		IdleConnTimeout: cfg.IdleConnTimeout,
	}
	client := &http.Client{
		Timeout:   cfg.Timeout,
		Transport: transport,
	}

	router := &Router{
		upstreams:         make(map[string]*Upstream),
		config:            cfg,
		client:            client,
		connTracker:       NewConnectionTracker(),
		weightedBalancers: make(map[string]*WeightedRoundRobin),
	}

	// Initialize upstreams from config
	for name, upstreamCfg := range cfg.Upstreams {
		upstream := &Upstream{
			Name:    name,
			URLs:    upstreamCfg.URLs,
			Weights: make([]int, len(upstreamCfg.URLs)),
			Current: 0,
		}

		// Initialize weights (default to 1 if not specified)
		for range upstreamCfg.URLs {
			weight := upstreamCfg.Weight
			if weight == 0 {
				weight = 1
			}
			upstream.Weights = append(upstream.Weights, weight)
		}

		// Initialize weighted balancer if using weighted strategy
		if cfg.LoadBalancer == "weighted" {
			router.weightedBalancers[name] = NewWeightedRoundRobin(upstream.URLs, upstream.Weights)
		}

		// Initialize health checker if configured
		if upstreamCfg.HealthCheck.Path != "" {
			upstream.Health = NewHealthChecker(upstream, upstreamCfg.HealthCheck)
			go upstream.Health.Start()
		}

		router.upstreams[name] = upstream
	}

	return router
}

// Proxy proxies a request to an upstream service
func (r *Router) Proxy(c *gin.Context, serviceName, path string) {
	upstream, ok := r.upstreams[serviceName]
	if !ok {
		c.JSON(http.StatusBadGateway, gin.H{
			"error":     "Bad Gateway",
			"message":   fmt.Sprintf("Upstream service '%s' not found", serviceName),
			"code":      http.StatusBadGateway,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Select upstream URL based on load balancing strategy
	upstreamURL := r.selectUpstream(upstream)
	if upstreamURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{
			"error":     "Service Unavailable",
			"message":   "No healthy upstream available",
			"code":      http.StatusServiceUnavailable,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Build target URL
	targetURL := strings.TrimSuffix(upstreamURL, "/") + "/" + strings.TrimPrefix(path, "/")

	// Create request
	req, err := http.NewRequestWithContext(c.Request.Context(), c.Request.Method, targetURL, c.Request.Body)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error":     "Internal Server Error",
			"message":   "Failed to create upstream request",
			"code":      http.StatusInternalServerError,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}

	// Copy headers
	for key, values := range c.Request.Header {
		for _, value := range values {
			req.Header.Add(key, value)
		}
	}

	// Remove hop-by-hop headers
	req.Header.Del("Connection")
	req.Header.Del("Keep-Alive")
	req.Header.Del("Proxy-Authenticate")
	req.Header.Del("Proxy-Authorization")
	req.Header.Del("Te")
	req.Header.Del("Trailers")
	req.Header.Del("Transfer-Encoding")
	req.Header.Del("Upgrade")

	// Track connection for least connections strategy
	if r.config.LoadBalancer == "least_connections" {
		r.connTracker.Increment(upstreamURL)
		defer r.connTracker.Decrement(upstreamURL)
	}

	// Record start time for metrics
	start := time.Now()

	// Make request
	resp, err := r.client.Do(req)
	if err != nil {
		metrics.UpstreamRequests.WithLabelValues(serviceName, "error").Inc()
		c.JSON(http.StatusBadGateway, gin.H{
			"error":     "Bad Gateway",
			"message":   fmt.Sprintf("Failed to connect to upstream: %v", err),
			"code":      http.StatusBadGateway,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
		return
	}
	defer resp.Body.Close()

	// Record metrics
	duration := time.Since(start).Seconds()
	statusCode := fmt.Sprintf("%d", resp.StatusCode)
	metrics.UpstreamRequests.WithLabelValues(serviceName, statusCode).Inc()
	metrics.UpstreamRequestDuration.WithLabelValues(serviceName, statusCode).Observe(duration)

	// Copy response headers
	for key, values := range resp.Header {
		for _, value := range values {
			c.Header(key, value)
		}
	}

	// Set status code
	c.Status(resp.StatusCode)

	// Copy response body
	io.Copy(c.Writer, resp.Body)
}

// selectUpstream selects an upstream URL based on load balancing strategy
func (r *Router) selectUpstream(upstream *Upstream) string {
	if len(upstream.URLs) == 0 {
		return ""
	}

	switch r.config.LoadBalancer {
	case "round_robin":
		return r.roundRobin(upstream)
	case "least_connections":
		return r.leastConnections(upstream)
	case "weighted":
		return r.weighted(upstream)
	default:
		return r.roundRobin(upstream)
	}
}

// roundRobin selects next upstream in round-robin fashion
func (r *Router) roundRobin(upstream *Upstream) string {
	if len(upstream.URLs) == 0 {
		return ""
	}

	// Filter healthy upstreams if health checker is available
	healthyURLs := upstream.URLs
	if upstream.Health != nil {
		healthyURLs = upstream.Health.GetHealthyURLs()
		if len(healthyURLs) == 0 {
			return ""
		}
	}

	upstream.Current = (upstream.Current + 1) % len(healthyURLs)
	return healthyURLs[upstream.Current]
}

// leastConnections selects upstream with least connections
func (r *Router) leastConnections(upstream *Upstream) string {
	// Filter healthy upstreams if health checker is available
	healthyURLs := upstream.URLs
	if upstream.Health != nil {
		healthyURLs = upstream.Health.GetHealthyURLs()
		if len(healthyURLs) == 0 {
			return ""
		}
	}

	return r.connTracker.GetLeastConnections(healthyURLs)
}

// weighted selects upstream based on weights
func (r *Router) weighted(upstream *Upstream) string {
	// Filter healthy upstreams if health checker is available
	healthyURLs := upstream.URLs
	if upstream.Health != nil {
		healthyURLs = upstream.Health.GetHealthyURLs()
		if len(healthyURLs) == 0 {
			return ""
		}
	}

	// Get or create weighted balancer
	balancer, exists := r.weightedBalancers[upstream.Name]
	if !exists {
		balancer = NewWeightedRoundRobin(healthyURLs, upstream.Weights)
		r.weightedBalancers[upstream.Name] = balancer
	}

	return balancer.Next()
}

// ParseServicePath parses service name and path from request path
func ParseServicePath(path string) (service, remainingPath string, err error) {
	// Remove leading slash
	path = strings.TrimPrefix(path, "/")

	// Split by first slash
	parts := strings.SplitN(path, "/", 2)
	if len(parts) == 0 {
		return "", "", fmt.Errorf("invalid path format")
	}

	service = parts[0]
	if len(parts) > 1 {
		remainingPath = "/" + parts[1]
	} else {
		remainingPath = "/"
	}

	return service, remainingPath, nil
}

// ValidateURL validates an upstream URL
func ValidateURL(urlStr string) error {
	_, err := url.Parse(urlStr)
	return err
}

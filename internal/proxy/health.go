package proxy

import (
	"context"
	"net/http"
	"sync"
	"time"
)

// HealthChecker checks health of upstream services
type HealthChecker struct {
	upstream   *Upstream
	path       string
	interval   time.Duration
	timeout    time.Duration
	healthy    map[string]bool
	mu         sync.RWMutex
	httpClient *http.Client
	stop       chan struct{}
}

// NewHealthChecker creates a new health checker
func NewHealthChecker(upstream *Upstream, config struct {
	Path     string
	Interval time.Duration
	Timeout  time.Duration
}) *HealthChecker {
	return &HealthChecker{
		upstream: upstream,
		path:     config.Path,
		interval: config.Interval,
		timeout:  config.Timeout,
		healthy:  make(map[string]bool),
		httpClient: &http.Client{
			Timeout: config.Timeout,
		},
		stop: make(chan struct{}),
	}
}

// Start starts the health checker
func (hc *HealthChecker) Start() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	// Initial health check
	hc.checkAll()

	for {
		select {
		case <-ticker.C:
			hc.checkAll()
		case <-hc.stop:
			return
		}
	}
}

// Stop stops the health checker
func (hc *HealthChecker) Stop() {
	close(hc.stop)
}

// checkAll checks health of all upstream URLs
func (hc *HealthChecker) checkAll() {
	for _, url := range hc.upstream.URLs {
		healthy := hc.checkHealth(url)
		hc.mu.Lock()
		hc.healthy[url] = healthy
		hc.mu.Unlock()
	}
}

// checkHealth checks health of a single upstream URL
func (hc *HealthChecker) checkHealth(baseURL string) bool {
	healthURL := baseURL + hc.path

	ctx, cancel := context.WithTimeout(context.Background(), hc.timeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "GET", healthURL, nil)
	if err != nil {
		return false
	}

	resp, err := hc.httpClient.Do(req)
	if err != nil {
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == http.StatusOK
}

// GetHealthyURLs returns list of healthy upstream URLs
func (hc *HealthChecker) GetHealthyURLs() []string {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	var healthyURLs []string
	for url, isHealthy := range hc.healthy {
		if isHealthy {
			healthyURLs = append(healthyURLs, url)
		}
	}

	// If no healthy URLs found, return all URLs as fallback
	if len(healthyURLs) == 0 {
		return hc.upstream.URLs
	}

	return healthyURLs
}

// IsHealthy checks if a specific URL is healthy
func (hc *HealthChecker) IsHealthy(url string) bool {
	hc.mu.RLock()
	defer hc.mu.RUnlock()
	return hc.healthy[url]
}

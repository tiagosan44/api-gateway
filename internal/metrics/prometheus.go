package metrics

import (
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var (
	// HTTPRequestTotal counts total HTTP requests
	HTTPRequestTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	// HTTPRequestDuration tracks HTTP request latency
	HTTPRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path", "status"},
	)

	// RateLimitHits counts rate limit hits
	RateLimitHits = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "rate_limit_hits_total",
			Help: "Total number of rate limit hits",
		},
		[]string{"key", "algorithm"},
	)

	// AuthFailures counts authentication failures
	AuthFailures = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "auth_failures_total",
			Help: "Total number of authentication failures",
		},
		[]string{"reason", "auth_type"},
	)

	// UpstreamRequests tracks upstream service requests
	UpstreamRequests = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "upstream_requests_total",
			Help: "Total number of upstream service requests",
		},
		[]string{"upstream", "status"},
	)

	// UpstreamRequestDuration tracks upstream request latency
	UpstreamRequestDuration = prometheus.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "upstream_request_duration_seconds",
			Help:    "Upstream request latency in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"upstream", "status"},
	)
)

// Initialize registers all metrics
func Initialize() {
	prometheus.MustRegister(HTTPRequestTotal)
	prometheus.MustRegister(HTTPRequestDuration)
	prometheus.MustRegister(RateLimitHits)
	prometheus.MustRegister(AuthFailures)
	prometheus.MustRegister(UpstreamRequests)
	prometheus.MustRegister(UpstreamRequestDuration)
}

// Handler returns the Prometheus metrics handler
func Handler(w http.ResponseWriter, r *http.Request) {
	promhttp.Handler().ServeHTTP(w, r)
}

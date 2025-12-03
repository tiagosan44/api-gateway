package proxy

import (
	"sync"
)

// ConnectionTracker tracks active connections per upstream
type ConnectionTracker struct {
	connections map[string]int
	mu          sync.RWMutex
}

// NewConnectionTracker creates a new connection tracker
func NewConnectionTracker() *ConnectionTracker {
	return &ConnectionTracker{
		connections: make(map[string]int),
	}
}

// Increment increments connection count for an upstream
func (ct *ConnectionTracker) Increment(url string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	ct.connections[url]++
}

// Decrement decrements connection count for an upstream
func (ct *ConnectionTracker) Decrement(url string) {
	ct.mu.Lock()
	defer ct.mu.Unlock()
	if ct.connections[url] > 0 {
		ct.connections[url]--
	}
}

// GetCount returns connection count for an upstream
func (ct *ConnectionTracker) GetCount(url string) int {
	ct.mu.RLock()
	defer ct.mu.RUnlock()
	return ct.connections[url]
}

// GetLeastConnections returns the URL with least connections
func (ct *ConnectionTracker) GetLeastConnections(urls []string) string {
	if len(urls) == 0 {
		return ""
	}

	leastURL := urls[0]
	leastCount := ct.GetCount(leastURL)

	for _, url := range urls[1:] {
		count := ct.GetCount(url)
		if count < leastCount {
			leastCount = count
			leastURL = url
		}
	}

	return leastURL
}

// WeightedRoundRobin implements weighted round-robin selection
type WeightedRoundRobin struct {
	urls    []string
	weights []int
	current []int // current weight for each URL
	index   int
	mu      sync.Mutex
}

// NewWeightedRoundRobin creates a new weighted round-robin balancer
func NewWeightedRoundRobin(urls []string, weights []int) *WeightedRoundRobin {
	current := make([]int, len(urls))
	copy(current, weights)

	return &WeightedRoundRobin{
		urls:    urls,
		weights: weights,
		current: current,
		index:   0,
	}
}

// Next returns the next URL using weighted round-robin
func (wrr *WeightedRoundRobin) Next() string {
	wrr.mu.Lock()
	defer wrr.mu.Unlock()

	if len(wrr.urls) == 0 {
		return ""
	}

	// Find URL with maximum current weight
	maxWeight := -1
	maxIndex := 0

	for i := range wrr.urls {
		if wrr.current[i] > maxWeight {
			maxWeight = wrr.current[i]
			maxIndex = i
		}
	}

	// Decrease current weight by total weight sum
	totalWeight := 0
	for _, w := range wrr.weights {
		totalWeight += w
	}

	wrr.current[maxIndex] -= totalWeight

	// Increase all current weights by their base weights
	for i := range wrr.current {
		wrr.current[i] += wrr.weights[i]
	}

	return wrr.urls[maxIndex]
}


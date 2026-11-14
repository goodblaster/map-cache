package v1

import (
	"context"
	"sync"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// Cache metrics
	cacheSizeBytes = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_size_bytes",
			Help: "Current size of cache in bytes",
		},
		[]string{"cache"},
	)

	cacheKeyCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_keys_total",
			Help: "Total number of keys in cache",
		},
		[]string{"cache"},
	)

	cacheActivityCount = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_activity_total",
			Help: "Total number of operations performed on cache",
		},
		[]string{"cache"},
	)

	cacheLongOperations = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_long_operations_total",
			Help: "Total number of long-running operations on cache",
		},
		[]string{"cache"},
	)

	cacheTimeoutOperations = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_timeout_operations_total",
			Help: "Total number of timed-out operations on cache",
		},
		[]string{"cache"},
	)

	cacheCount = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "caches_total",
			Help: "Total number of active caches",
		},
	)
)

// reportedCaches is the set of cache names that currently have metric series.
var (
	reportedCachesMu sync.Mutex
	reportedCaches   = map[string]struct{}{}
)

// UpdateCacheMetrics updates Prometheus metrics with current cache statistics.
// This should be called periodically (e.g., every 10 seconds) to keep metrics fresh.
func UpdateCacheMetrics() {
	ctx := context.Background()
	cacheNames := caches.List()

	// Update total cache count
	cacheCount.Set(float64(len(cacheNames)))

	// Drop the series of caches that no longer exist. Caches are often created
	// with unique names and a TTL, so stale labels would otherwise accumulate
	// (and keep reporting their last values) forever.
	current := make(map[string]struct{}, len(cacheNames))
	for _, name := range cacheNames {
		current[name] = struct{}{}
	}
	reportedCachesMu.Lock()
	for name := range reportedCaches {
		if _, ok := current[name]; !ok {
			cacheSizeBytes.DeleteLabelValues(name)
			cacheKeyCount.DeleteLabelValues(name)
			cacheActivityCount.DeleteLabelValues(name)
			cacheLongOperations.DeleteLabelValues(name)
			cacheTimeoutOperations.DeleteLabelValues(name)
		}
	}
	reportedCaches = current
	reportedCachesMu.Unlock()

	for _, name := range cacheNames {
		cache, err := caches.FetchCache(name)
		if err != nil {
			continue
		}

		// Acquire lock briefly to read stats
		cache.Acquire("metrics")

		// Update basic cache metrics
		cacheSizeBytes.WithLabelValues(name).Set(float64(cache.SizeBytes(ctx)))
		cacheActivityCount.WithLabelValues(name).Set(float64(cache.ActivityCount()))

		// Count keys (assuming we can get this from the cache)
		// Note: This might be expensive for large caches
		keys := cache.WildKeys(ctx, "*")
		cacheKeyCount.WithLabelValues(name).Set(float64(len(keys)))

		cache.Release("metrics")

		// Update operation stats (thread-safe, no lock needed)
		opStats := cache.OperationStatsSnapshot()
		longCount := 0
		timeoutCount := 0
		for _, op := range opStats.RecentHistory {
			longCount++
			if op.TimedOut {
				timeoutCount++
			}
		}
		cacheLongOperations.WithLabelValues(name).Set(float64(longCount))
		cacheTimeoutOperations.WithLabelValues(name).Set(float64(timeoutCount))
	}
}

package caches

import (
	"net/http"
	"time"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// CacheStats represents statistics for a single cache
type CacheStats struct {
	Name          string     `json:"name"`
	SizeBytes     int        `json:"size_bytes"`
	TTLMillis     *int64     `json:"ttl_ms,omitempty"` // nil if no TTL set
	LastAccessed  *time.Time `json:"last_accessed,omitempty"`
	ActivityCount int64      `json:"activity_count"`
}

// StatsResponse represents the overall statistics response
type StatsResponse struct {
	Caches []CacheStats `json:"caches"`
}

// handleGetStats returns statistics for all caches
func handleGetStats() echo.HandlerFunc {
	return func(c echo.Context) error {
		cacheNames := caches.List()
		allStats := make([]CacheStats, 0, len(cacheNames))

		for _, name := range cacheNames {
			cache, err := caches.FetchCache(name)
			if err != nil {
				// Skip caches that can't be fetched
				continue
			}

			// Thread-safe: acquire lock for the duration of stat collection
			cache.Acquire("stats")

			stats := CacheStats{
				Name:          name,
				SizeBytes:     cache.SizeBytes(c.Request().Context()),
				TTLMillis:     cache.TTLMillis(),
				LastAccessed:  cache.LastAccessed(),
				ActivityCount: cache.ActivityCount(),
			}

			cache.Release("stats")

			allStats = append(allStats, stats)
		}

		response := StatsResponse{
			Caches: allStats,
		}

		return c.JSON(http.StatusOK, response)
	}
}

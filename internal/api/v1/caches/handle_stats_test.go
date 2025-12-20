package caches

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHandleGetStats_Empty(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/caches/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	handler := handleGetStats()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result StatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result.Caches), 0)
}

func TestHandleGetStats_WithCaches(t *testing.T) {
	// Setup
	e := echo.New()

	// Create test cache with data
	err := caches.AddCache("test-stats-cache-1")
	require.NoError(t, err)
	defer caches.DeleteCache("test-stats-cache-1")

	// Set cache-level TTL
	err = caches.SetCacheTTL("test-stats-cache-1", 60000)
	require.NoError(t, err)

	// Add keys to cache
	cache1, err := caches.FetchCache("test-stats-cache-1")
	require.NoError(t, err)
	cache1.Acquire("test")
	err = cache1.Create(context.Background(), map[string]any{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	})
	cache1.Release("test")
	require.NoError(t, err)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/caches/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	handler := handleGetStats()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result StatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)

	// Find our test cache in the results
	var testCache1Stats *CacheStats
	for i := range result.Caches {
		if result.Caches[i].Name == "test-stats-cache-1" {
			testCache1Stats = &result.Caches[i]
			break
		}
	}

	require.NotNil(t, testCache1Stats, "test-stats-cache-1 should be in results")
	assert.Greater(t, testCache1Stats.SizeBytes, 0)
	assert.NotNil(t, testCache1Stats.TTLMillis)
	assert.Equal(t, int64(60000), *testCache1Stats.TTLMillis)
	assert.NotNil(t, testCache1Stats.LastAccessed)
	assert.Greater(t, testCache1Stats.ActivityCount, int64(0))
}

func TestHandleGetStats_NoCacheTTL(t *testing.T) {
	// Setup
	e := echo.New()

	// Create test cache without TTL
	err := caches.AddCache("test-stats-no-ttl")
	require.NoError(t, err)
	defer caches.DeleteCache("test-stats-no-ttl")

	// Add some data
	cache, err := caches.FetchCache("test-stats-no-ttl")
	require.NoError(t, err)
	cache.Acquire("test")
	err = cache.Create(context.Background(), map[string]any{
		"test": "data",
	})
	cache.Release("test")
	require.NoError(t, err)

	// Make request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/caches/stats", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	handler := handleGetStats()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result StatsResponse
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)

	// Find our test cache
	var testCacheStats *CacheStats
	for i := range result.Caches {
		if result.Caches[i].Name == "test-stats-no-ttl" {
			testCacheStats = &result.Caches[i]
			break
		}
	}

	require.NotNil(t, testCacheStats, "test-stats-no-ttl should be in results")
	assert.Nil(t, testCacheStats.TTLMillis)
	assert.Greater(t, testCacheStats.SizeBytes, 0)
}

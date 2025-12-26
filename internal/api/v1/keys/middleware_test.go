package keys

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCacheMW_DefaultCache(t *testing.T) {
	// Setup - ensure default cache exists
	_ = caches.AddCache(caches.DefaultName)
	defer caches.DeleteCache(caches.DefaultName)

	e := echo.New()
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true

		// Verify cache is set in context
		cache := c.Get("cache")
		assert.NotNil(t, cache)

		// Verify it's the default cache
		_, ok := cache.(*caches.Cache)
		assert.True(t, ok)

		// Verify request_id is set
		requestID := c.Get("request_id")
		assert.NotNil(t, requestID)
		assert.NotEmpty(t, requestID)

		return nil
	}

	// Create request without X-Cache-Name header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Execute
	mw := cacheMW(handler)
	err := mw(ctx)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestCacheMW_CustomCache(t *testing.T) {
	// Setup - create a custom cache
	cacheName := "test-cache-mw"
	err := caches.AddCache(cacheName)
	require.NoError(t, err)
	defer caches.DeleteCache(cacheName)

	e := echo.New()
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true

		// Verify cache is set in context
		cache := c.Get("cache")
		assert.NotNil(t, cache)

		// Verify request_id is set
		requestID := c.Get("request_id")
		assert.NotNil(t, requestID)
		assert.NotEmpty(t, requestID)

		return nil
	}

	// Create request with X-Cache-Name header
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Cache-Name", cacheName)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Execute
	mw := cacheMW(handler)
	err = mw(ctx)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestCacheMW_CacheNotFound(t *testing.T) {
	// Setup - ensure the cache name we use definitely doesn't exist
	cacheName := "definitely-nonexistent-cache-12345"
	_ = caches.DeleteCache(cacheName) // Delete if it somehow exists

	e := echo.New()
	handlerCalled := false
	handler := func(c echo.Context) error {
		handlerCalled = true
		return nil
	}

	// Create request with non-existent cache name
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Cache-Name", cacheName)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Execute
	mw := cacheMW(handler)
	err := mw(ctx)

	// Assert - handler should not be called
	assert.False(t, handlerCalled)

	// Check that error was returned with correct status
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusFailedDependency, he.Code)
}

func TestCacheMW_AcquireAndRelease(t *testing.T) {
	// Setup
	cacheName := "test-cache-acquire"
	err := caches.AddCache(cacheName)
	require.NoError(t, err)
	defer caches.DeleteCache(cacheName)

	cache, err := caches.FetchCache(cacheName)
	require.NoError(t, err)

	e := echo.New()
	var capturedRequestID string
	handler := func(c echo.Context) error {
		// Capture request ID
		capturedRequestID = c.Get("request_id").(string)

		// At this point, cache should be acquired
		// We can't directly test the lock state, but we can verify the request ID is set
		assert.NotEmpty(t, capturedRequestID)

		return nil
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("X-Cache-Name", cacheName)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Execute
	mw := cacheMW(handler)
	err = mw(ctx)

	// Assert
	assert.NoError(t, err)
	assert.NotEmpty(t, capturedRequestID)

	// After middleware completes, cache should be released
	// We can verify by acquiring it ourselves (should succeed immediately)
	done := make(chan bool)
	go func() {
		cache.Acquire("test")
		cache.Release("test")
		done <- true
	}()

	// This should complete quickly if the lock was released
	select {
	case <-done:
		// Success - cache was released
	}
}

func TestCacheMW_HandlerError(t *testing.T) {
	// Setup - ensure default cache exists
	_ = caches.AddCache(caches.DefaultName)
	defer caches.DeleteCache(caches.DefaultName)

	e := echo.New()
	expectedErr := echo.NewHTTPError(http.StatusBadRequest, "handler error")
	handler := func(c echo.Context) error {
		return expectedErr
	}

	// Create request
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rec := httptest.NewRecorder()
	ctx := e.NewContext(req, rec)

	// Execute
	mw := cacheMW(handler)
	err := mw(ctx)

	// Assert - error should be propagated
	assert.Equal(t, expectedErr, err)
}

func TestCacheMW_UniqueRequestIDs(t *testing.T) {
	// Setup - ensure default cache exists
	_ = caches.AddCache(caches.DefaultName)
	defer caches.DeleteCache(caches.DefaultName)

	e := echo.New()
	var requestIDs []string
	handler := func(c echo.Context) error {
		requestID := c.Get("request_id").(string)
		requestIDs = append(requestIDs, requestID)
		return nil
	}

	mw := cacheMW(handler)

	// Execute multiple requests
	for i := 0; i < 10; i++ {
		req := httptest.NewRequest(http.MethodGet, "/", nil)
		rec := httptest.NewRecorder()
		ctx := e.NewContext(req, rec)
		_ = mw(ctx)
	}

	// Assert - all request IDs should be unique
	seen := make(map[string]bool)
	for _, id := range requestIDs {
		assert.False(t, seen[id], "request ID %s was used more than once", id)
		seen[id] = true
	}
	assert.Len(t, requestIDs, 10)
}

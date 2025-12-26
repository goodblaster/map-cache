package caches

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleGetCacheList_Empty(t *testing.T) {
	// Setup
	e := echo.New()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/caches", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	handler := handleGetCacheList()
	err := handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Should return an array (possibly empty or with just default cache)
	var result []string
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)
}

func TestHandleGetCacheList_WithCaches(t *testing.T) {
	// Setup
	e := echo.New()

	// Create some test caches
	err := caches.AddCache("test-cache-1")
	assert.NoError(t, err)
	defer caches.DeleteCache("test-cache-1")

	err = caches.AddCache("test-cache-2")
	assert.NoError(t, err)
	defer caches.DeleteCache("test-cache-2")

	req := httptest.NewRequest(http.MethodGet, "/api/v1/caches", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Execute
	handler := handleGetCacheList()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	var result []string
	err = json.Unmarshal(rec.Body.Bytes(), &result)
	assert.NoError(t, err)

	// Should contain our test caches
	assert.Contains(t, result, "test-cache-1")
	assert.Contains(t, result, "test-cache-2")
}

package caches

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleUpdateCache_SetTTL(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a test cache
	err := caches.AddCache("test-update-cache")
	assert.NoError(t, err)
	defer caches.DeleteCache("test-update-cache")

	// Set TTL to 5000ms
	reqBody := `{"ttl": 5000}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/caches/test-update-cache", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues("test-update-cache")

	// Execute
	handler := handleUpdateCache()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandleUpdateCache_RemoveTTL(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a test cache with TTL
	err := caches.AddCache("test-update-cache-2")
	assert.NoError(t, err)
	defer caches.DeleteCache("test-update-cache-2")

	// Set TTL first
	err = caches.SetCacheTTL("test-update-cache-2", 5000)
	assert.NoError(t, err)

	// Now remove TTL by sending null
	reqBody := `{"ttl": null}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/caches/test-update-cache-2", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues("test-update-cache-2")

	// Execute
	handler := handleUpdateCache()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusNoContent, rec.Code)
}

func TestHandleUpdateCache_InvalidJSON(t *testing.T) {
	// Setup
	e := echo.New()

	reqBody := `{invalid json}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/caches/test-cache", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues("test-cache")

	// Execute
	handler := handleUpdateCache()
	err := handler(c)

	// Assert - should return bad request
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandleUpdateCache_MissingCacheName(t *testing.T) {
	// Setup
	e := echo.New()

	reqBody := `{"ttl": 1000}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/caches/", strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Don't set param

	// Execute
	handler := handleUpdateCache()
	err := handler(c)

	// Assert
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

func TestHandleUpdateCache_DefaultCache(t *testing.T) {
	// Setup
	e := echo.New()

	// Try to update the default cache (should be forbidden)
	reqBody := `{"ttl": 1000}`
	req := httptest.NewRequest(http.MethodPut, "/api/v1/caches/"+caches.DefaultName, strings.NewReader(reqBody))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues(caches.DefaultName)

	// Execute
	handler := handleUpdateCache()
	err := handler(c)

	// Assert - should return bad request (cannot modify default cache)
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

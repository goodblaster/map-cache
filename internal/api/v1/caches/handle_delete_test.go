package caches

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleDeleteCache_Success(t *testing.T) {
	// Setup
	e := echo.New()

	// Create a test cache
	err := caches.AddCache("test-delete-cache")
	assert.NoError(t, err)

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/caches/test-delete-cache", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues("test-delete-cache")

	// Execute
	handler := handleDeleteCache()
	err = handler(c)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, http.StatusOK, rec.Code)

	// Verify cache was actually deleted
	_, err = caches.FetchCache("test-delete-cache")
	assert.Error(t, err) // Should error because cache doesn't exist
}

func TestHandleDeleteCache_NotFound(t *testing.T) {
	// Setup
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/caches/nonexistent", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("name")
	c.SetParamValues("nonexistent")

	// Execute
	handler := handleDeleteCache()
	err := handler(c)

	// Assert - should return 404 for non-existent cache
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusNotFound, he.Code)
}

func TestHandleDeleteCache_MissingName(t *testing.T) {
	// Setup
	e := echo.New()

	req := httptest.NewRequest(http.MethodDelete, "/api/v1/caches/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	// Don't set param - simulating missing name

	// Execute
	handler := handleDeleteCache()
	err := handler(c)

	// Assert - should return bad request
	assert.Error(t, err)
	he, ok := err.(*echo.HTTPError)
	assert.True(t, ok)
	assert.Equal(t, http.StatusBadRequest, he.Code)
}

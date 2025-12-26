package keys

import (
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestCache_Success(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	expectedCache := caches.New()
	c.Set("cache", expectedCache)

	result := Cache(c)
	assert.Equal(t, expectedCache, result)
}

func TestCache_PanicWhenNotSet(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Don't set cache in context
	assert.Panics(t, func() {
		Cache(c)
	}, "cache value is not set")
}

func TestCache_PanicWhenWrongType(t *testing.T) {
	e := echo.New()
	req := httptest.NewRequest("GET", "/", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)

	// Set wrong type in context
	c.Set("cache", "not a cache")

	assert.Panics(t, func() {
		Cache(c)
	}, "cache value is not of type *caches.Cache")
}

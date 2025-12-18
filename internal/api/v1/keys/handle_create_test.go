package keys

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
)

func TestHandleCreate(t *testing.T) {
	e := echo.New()

	t.Run("valid request", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"entries": map[string]any{
				"key1": "value1",
				"key2": 42,
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreate()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}

		// Verify keys were created
		ctx := context.Background()
		cache.Acquire("test")
		val, err := cache.Get(ctx, "key1")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, "value1", val)
	})

	t.Run("empty entries", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"entries": map[string]any{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreate()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("with TTL", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"entries": map[string]any{
				"key1": "value1",
			},
			"ttl": map[string]int64{
				"key1": 5000, // 5 seconds
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreate()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusCreated, rec.Code)
		}
	})

	t.Run("TTL for non-existent key", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"entries": map[string]any{
				"key1": "value1",
			},
			"ttl": map[string]int64{
				"key2": 5000, // TTL for key that doesn't exist
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCreate()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

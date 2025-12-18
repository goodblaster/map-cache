package v1keys

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
	"github.com/stretchr/testify/require"
)

func TestHandlePatch(t *testing.T) {
	e := echo.New()

	t.Run("create operation", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"operations": []map[string]any{
				{
					"type":  "CREATE",
					"key":   "newkey",
					"value": "newvalue",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/newkey", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("newkey")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify key was created
		ctx := context.Background()
		cache.Acquire("test")
		val, err := cache.Get(ctx, "newkey")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, "newvalue", val)
	})

	t.Run("replace operation", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial key
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key1": "oldvalue"})
		cache.Release("test")
		require.NoError(t, err)

		reqBody := map[string]any{
			"operations": []map[string]any{
				{
					"type":  "REPLACE",
					"key":   "key1",
					"value": "newvalue",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify value was replaced
		cache.Acquire("test")
		val, err := cache.Get(ctx, "key1")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, "newvalue", val)
	})

	t.Run("increment operation", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial counter
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"counter": float64(10)})
		cache.Release("test")
		require.NoError(t, err)

		reqBody := map[string]any{
			"operations": []map[string]any{
				{
					"type":  "INC",
					"key":   "counter",
					"value": 5,
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/counter", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("counter")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify value was incremented
		cache.Acquire("test")
		val, err := cache.Get(ctx, "counter")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, float64(15), val)
	})

	t.Run("delete operation", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial key
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key1": "value"})
		cache.Release("test")
		require.NoError(t, err)

		reqBody := map[string]any{
			"operations": []map[string]any{
				{
					"type": "DELETE",
					"key":  "key1",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify key was deleted
		cache.Acquire("test")
		_, err = cache.Get(ctx, "key1")
		cache.Release("test")
		assert.Error(t, err)
	})

	t.Run("empty operations", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"operations": []map[string]any{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("invalid operation type", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"operations": []map[string]any{
				{
					"type":  "INVALID",
					"key":   "key1",
					"value": "value",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPatch, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

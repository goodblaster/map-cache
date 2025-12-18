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

func TestHandlePut(t *testing.T) {
	e := echo.New()

	t.Run("replace existing key", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial key
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key1": "oldvalue"})
		cache.Release("test")
		require.NoError(t, err)

		// Replace the value
		reqBody := map[string]any{
			"value": "newvalue",
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePut()
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

	t.Run("replace with nil", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial key
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"key1": "value"})
		cache.Release("test")
		require.NoError(t, err)

		// Replace with null
		reqBody := map[string]any{
			"value": nil,
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/keys/key1", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("key")
		c.SetParamValues("key1")
		c.Set("cache", cache)

		h := handlePut()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}
	})
}

func TestHandleReplaceBatch(t *testing.T) {
	e := echo.New()

	t.Run("replace multiple keys", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial keys
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"key1": "old1",
			"key2": "old2",
		})
		cache.Release("test")
		require.NoError(t, err)

		// Replace both values
		reqBody := map[string]any{
			"entries": map[string]any{
				"key1": "new1",
				"key2": "new2",
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleReplaceBatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify values were replaced
		cache.Acquire("test")
		val1, _ := cache.Get(ctx, "key1")
		val2, _ := cache.Get(ctx, "key2")
		cache.Release("test")
		assert.Equal(t, "new1", val1)
		assert.Equal(t, "new2", val2)
	})

	t.Run("empty key in batch", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"entries": map[string]any{
				"": "value",
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPut, "/keys", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleReplaceBatch()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
		}
	})
}

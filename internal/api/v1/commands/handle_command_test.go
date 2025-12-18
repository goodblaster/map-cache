package commands

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

func TestHandleCommand(t *testing.T) {
	e := echo.New()

	t.Run("execute single command", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial data
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{"counter": float64(0)})
		cache.Release("test")
		require.NoError(t, err)

		// Execute INC command
		reqBody := map[string]any{
			"commands": []map[string]any{
				{
					"type":  "INC",
					"key":   "counter",
					"value": 5,
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/commands/execute", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCommand()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify counter was incremented
		cache.Acquire("test")
		val, err := cache.Get(ctx, "counter")
		cache.Release("test")
		assert.NoError(t, err)
		assert.Equal(t, float64(5), val)
	})

	t.Run("execute multiple commands", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial data
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"counter": float64(0),
			"name":    "old",
		})
		cache.Release("test")
		require.NoError(t, err)

		// Execute multiple commands
		reqBody := map[string]any{
			"commands": []map[string]any{
				{
					"type":  "INC",
					"key":   "counter",
					"value": 10,
				},
				{
					"type":  "REPLACE",
					"key":   "name",
					"value": "new",
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/commands/execute", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCommand()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify both commands executed
		cache.Acquire("test")
		counter, _ := cache.Get(ctx, "counter")
		name, _ := cache.Get(ctx, "name")
		cache.Release("test")
		assert.Equal(t, float64(10), counter)
		assert.Equal(t, "new", name)
	})

	t.Run("empty commands array", func(t *testing.T) {
		cache := caches.New()
		reqBody := map[string]any{
			"commands": []map[string]any{},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/commands/execute", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCommand()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusBadRequest, rec.Code)
			assert.Contains(t, rec.Body.String(), "at least one command is required")
		}
	})

	t.Run("execute command sequence", func(t *testing.T) {
		cache := caches.New()
		ctx := context.Background()

		// Create initial data
		cache.Acquire("test")
		err := cache.Create(ctx, map[string]any{
			"x": float64(5),
			"y": float64(3),
		})
		cache.Release("test")
		require.NoError(t, err)

		// Execute sequence: increment x, increment y
		reqBody := map[string]any{
			"commands": []map[string]any{
				{
					"type":  "INC",
					"key":   "x",
					"value": 2,
				},
				{
					"type":  "INC",
					"key":   "y",
					"value": 1,
				},
			},
		}
		body, _ := json.Marshal(reqBody)
		req := httptest.NewRequest(http.MethodPost, "/commands/execute", bytes.NewReader(body))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.Set("cache", cache)

		h := handleCommand()
		if assert.NoError(t, h(c)) {
			assert.Equal(t, http.StatusOK, rec.Code)
		}

		// Verify both increments executed
		cache.Acquire("test")
		x, _ := cache.Get(ctx, "x")
		y, _ := cache.Get(ctx, "y")
		cache.Release("test")
		assert.Equal(t, float64(7), x)
		assert.Equal(t, float64(4), y)
	})
}

package v1

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/goodblaster/map-cache/internal/config"
	"github.com/goodblaster/map-cache/internal/log"
	"github.com/goodblaster/map-cache/pkg/caches"
	"github.com/labstack/echo/v4"
)

// TimingMiddleware times API operations and records long-running ones
func TimingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Create context with timeout for all operations
		ctx, cancel := context.WithTimeout(c.Request().Context(), time.Duration(config.CommandTimeoutMs)*time.Millisecond)
		defer cancel()

		// Replace request context with timeout context
		req := c.Request().WithContext(ctx)
		c.SetRequest(req)

		// Start timing
		start := time.Now()

		// Call next handler
		err := next(c)

		// Measure duration
		duration := time.Since(start)

		// Check if operation timed out
		timedOut := errors.Is(ctx.Err(), context.DeadlineExceeded)
		success := err == nil

		// Get operation description
		operation := fmt.Sprintf("%s %s", c.Request().Method, c.Path())
		cacheName := c.Request().Header.Get("X-Cache-Name")

		// Only record and log if operation exceeded threshold
		if duration.Milliseconds() > config.CommandLongThresholdMs {
			// Log long-running operation
			logger := log.
				With("cache", cacheName).
				With("operation", operation).
				With("duration_ms", duration.Milliseconds()).
				With("threshold_ms", config.CommandLongThresholdMs)

			if timedOut {
				logger.With("timeout_ms", config.CommandTimeoutMs).
					Error("API operation timed out")
			} else {
				logger.Warn("Long-running API operation detected")
			}

			// Record in cache stats if cache is available
			if cacheName != "" {
				if cache, cacheErr := caches.FetchCache(cacheName); cacheErr == nil {
					cache.RecordLongOperation(duration, operation, success, timedOut)
				}
			}
		}

		return err
	}
}

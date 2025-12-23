package api

import (
	"time"

	"github.com/goodblaster/map-cache/internal/log"
	"github.com/labstack/echo/v4"
)

// LoggingMiddleware provides structured logging for all HTTP requests with request ID correlation.
// It should be placed after RequestIDMiddleware to ensure request IDs are available.
//
// This middleware:
// - Extracts request ID from context
// - Creates a logger with request ID field
// - Stores logger in context for handler use
// - Logs request completion with status, duration, and request ID
func LoggingMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		start := time.Now()
		req := c.Request()

		// Get request ID from context (set by RequestIDMiddleware)
		requestID := GetRequestID(c)

		// Create a logger with request ID for this request
		requestLogger := log.Default().With("request_id", requestID)

		// Store logger in context for handlers to use
		c.SetRequest(req.WithContext(log.WithLogger(req.Context(), requestLogger)))

		// Execute the handler
		err := next(c)

		// Calculate request duration
		duration := time.Since(start)

		// Get response status (might be set by error handler)
		status := c.Response().Status
		if status == 0 {
			status = 200 // Default status if not set
		}

		// Log the completed request with structured fields
		requestLogger.
			With("method", req.Method).
			With("path", req.URL.Path).
			With("status", status).
			With("duration_ms", duration.Milliseconds()).
			With("remote_ip", c.RealIP()).
			Infof("%d %s %s", status, req.Method, req.URL.Path)

		return err
	}
}

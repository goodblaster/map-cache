package api

import (
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
)

const (
	// RequestIDHeader is the HTTP header name for request correlation ID
	RequestIDHeader = "X-Request-ID"

	// RequestIDContextKey is the key used to store request ID in Echo context
	RequestIDContextKey = "request_id"
)

// RequestIDMiddleware generates or extracts a request ID for correlation and tracing.
// It checks for an existing X-Request-ID header, uses it if present, or generates a new UUID.
// The request ID is:
// - Stored in Echo context as "request_id"
// - Added to response headers as X-Request-ID
// - Added to OpenTelemetry span attributes (if tracing is enabled)
func RequestIDMiddleware(next echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		// Check for existing request ID in header (for distributed tracing)
		requestID := c.Request().Header.Get(RequestIDHeader)

		// Generate new UUID if not provided
		if requestID == "" {
			requestID = uuid.New().String()
		}

		// Store in Echo context for use by handlers and other middleware
		c.Set(RequestIDContextKey, requestID)

		// Add to response headers for client correlation
		c.Response().Header().Set(RequestIDHeader, requestID)

		// Add to OpenTelemetry span if tracing is active
		span := trace.SpanFromContext(c.Request().Context())
		if span.SpanContext().IsValid() {
			span.SetAttributes(attribute.String("request.id", requestID))
		}

		return next(c)
	}
}

// GetRequestID extracts the request ID from Echo context.
// Returns empty string if not found.
func GetRequestID(c echo.Context) string {
	if id := c.Get(RequestIDContextKey); id != nil {
		if requestID, ok := id.(string); ok {
			return requestID
		}
	}
	return ""
}

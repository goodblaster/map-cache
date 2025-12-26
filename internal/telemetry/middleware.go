package telemetry

import (
	"github.com/labstack/echo/v4"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns Echo middleware that traces HTTP requests.
// Uses W3C Trace Context propagation standard (traceparent/tracestate headers).
func Middleware() echo.MiddlewareFunc {
	tracer := otel.Tracer("github.com/goodblaster/map-cache/http")
	propagator := propagation.TraceContext{}

	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			req := c.Request()

			// Extract trace context from headers (W3C Trace Context)
			ctx := propagator.Extract(req.Context(), propagation.HeaderCarrier(req.Header))

			// Start span
			spanName := req.Method + " " + c.Path()
			ctx, span := tracer.Start(ctx, spanName,
				trace.WithSpanKind(trace.SpanKindServer),
				trace.WithAttributes(
					attribute.String("http.method", req.Method),
					attribute.String("http.url", req.URL.String()),
					attribute.String("http.route", c.Path()),
					attribute.String("http.user_agent", req.UserAgent()),
				),
			)
			defer span.End()

			// Add cache name if present
			if cacheName := req.Header.Get("X-Cache-Name"); cacheName != "" {
				span.SetAttributes(attribute.String("cache.name", cacheName))
			}

			// Store updated context in Echo request
			c.SetRequest(req.WithContext(ctx))

			// Execute handler
			err := next(c)

			// Record response attributes
			span.SetAttributes(
				attribute.Int("http.status_code", c.Response().Status),
			)

			if err != nil {
				span.RecordError(err)
				span.SetAttributes(attribute.Bool("error", true))
			}

			return err
		}
	}
}

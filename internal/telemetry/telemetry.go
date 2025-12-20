package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"go.opentelemetry.io/otel/trace"
)

// Config holds telemetry configuration
type Config struct {
	Enabled        bool
	Exporter       string // "none", "stdout", "otlp"
	OTLPEndpoint   string
	ServiceName    string
	ServiceVersion string
}

// Init initializes OpenTelemetry tracing.
// Returns a shutdown function that must be called on application exit to flush traces.
// If disabled, returns a no-op shutdown function with zero overhead.
func Init(cfg Config) (shutdown func(context.Context) error, err error) {
	// If disabled or exporter is none, return no-op shutdown immediately
	if !cfg.Enabled || cfg.Exporter == "none" {
		return func(context.Context) error { return nil }, nil
	}

	// Create resource with service information
	res, err := resource.New(
		context.Background(),
		resource.WithAttributes(
			semconv.ServiceName(cfg.ServiceName),
			semconv.ServiceVersion(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create exporter based on config
	var exporter sdktrace.SpanExporter
	switch cfg.Exporter {
	case "stdout":
		exporter, err = stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
	case "otlp":
		exporter, err = otlptracegrpc.New(
			context.Background(),
			otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
			otlptracegrpc.WithInsecure(), // Use WithTLSCredentials() for production TLS
		)
	default:
		return nil, fmt.Errorf("unknown exporter type: %s", cfg.Exporter)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to create exporter: %w", err)
	}

	// Create tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)

	// Set global tracer provider
	otel.SetTracerProvider(tp)

	// Return shutdown function that flushes traces
	return func(ctx context.Context) error {
		return tp.Shutdown(ctx)
	}, nil
}

// Tracer returns the global tracer for map-cache.
// Safe to call even when tracing is disabled (returns no-op tracer).
func Tracer() trace.Tracer {
	return otel.Tracer("github.com/goodblaster/map-cache")
}

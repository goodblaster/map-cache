package config

import (
	"os"

	"github.com/goodblaster/map-cache/internal/log"
)

var (
	KeyDelimiter = "/"
	WebAddress   = ":8080"

	// Telemetry configuration
	TelemetryEnabled  = false
	TelemetryExporter = "none"
	OTLPEndpoint      = "localhost:4317"
	ServiceName       = "map-cache"
)

func Init(l log.Logger) {
	if val := os.Getenv("KEY_DELIMITER"); val != "" {
		KeyDelimiter = val
	}

	if val := os.Getenv("LISTEN_ADDRESS"); val != "" {
		WebAddress = val
	}

	// Telemetry configuration
	if val := os.Getenv("TELEMETRY_ENABLED"); val == "true" || val == "1" {
		TelemetryEnabled = true
	}

	if val := os.Getenv("TELEMETRY_EXPORTER"); val != "" {
		TelemetryExporter = val
	}

	if val := os.Getenv("OTLP_ENDPOINT"); val != "" {
		OTLPEndpoint = val
	}

	if val := os.Getenv("SERVICE_NAME"); val != "" {
		ServiceName = val
	}

	log.
		With("KEY_DELIMITER", KeyDelimiter).
		With("LISTEN_ADDRESS", WebAddress).
		With("TELEMETRY_ENABLED", TelemetryEnabled).
		With("TELEMETRY_EXPORTER", TelemetryExporter).
		Info("Configuration initialized")
}

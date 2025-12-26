package config

import (
	"os"
	"strconv"

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

	// Command execution configuration
	CommandLongThresholdMs = int64(500)   // 0.5 seconds default
	CommandTimeoutMs       = int64(10000) // 10 seconds default
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

	// Command execution configuration
	if val := os.Getenv("COMMAND_LONG_THRESHOLD_MS"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			CommandLongThresholdMs = parsed
		}
	}

	if val := os.Getenv("COMMAND_TIMEOUT_MS"); val != "" {
		if parsed, err := strconv.ParseInt(val, 10, 64); err == nil {
			CommandTimeoutMs = parsed
		}
	}

	log.
		With("KEY_DELIMITER", KeyDelimiter).
		With("LISTEN_ADDRESS", WebAddress).
		With("TELEMETRY_ENABLED", TelemetryEnabled).
		With("TELEMETRY_EXPORTER", TelemetryExporter).
		With("COMMAND_LONG_THRESHOLD_MS", CommandLongThresholdMs).
		With("COMMAND_TIMEOUT_MS", CommandTimeoutMs).
		Info("Configuration initialized")
}

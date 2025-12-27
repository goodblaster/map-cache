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

	// RESP (Redis Protocol) configuration
	RESPEnabled        = false
	RESPAddress        = ":6379"
	RESPKeyMode        = "translate"   // "translate" (: → /) or "preserve"
	RESPDefaultCache   = "default"
	RESPMaxConnections = 1000
	RESPBackupDir      = "./backups"
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

	// RESP configuration
	if val := os.Getenv("RESP_ENABLED"); val == "true" || val == "1" {
		RESPEnabled = true
	}

	if val := os.Getenv("RESP_ADDRESS"); val != "" {
		RESPAddress = val
	}

	if val := os.Getenv("RESP_KEY_MODE"); val != "" {
		RESPKeyMode = val
	}

	if val := os.Getenv("RESP_DEFAULT_CACHE"); val != "" {
		RESPDefaultCache = val
	}

	if val := os.Getenv("RESP_MAX_CONNECTIONS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			RESPMaxConnections = parsed
		}
	}

	if val := os.Getenv("RESP_BACKUP_DIR"); val != "" {
		RESPBackupDir = val
	}

	log.
		With("KEY_DELIMITER", KeyDelimiter).
		With("LISTEN_ADDRESS", WebAddress).
		With("TELEMETRY_ENABLED", TelemetryEnabled).
		With("TELEMETRY_EXPORTER", TelemetryExporter).
		With("COMMAND_LONG_THRESHOLD_MS", CommandLongThresholdMs).
		With("COMMAND_TIMEOUT_MS", CommandTimeoutMs).
		With("RESP_ENABLED", RESPEnabled).
		With("RESP_ADDRESS", RESPAddress).
		With("RESP_KEY_MODE", RESPKeyMode).
		Info("Configuration initialized")
}

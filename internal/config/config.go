package config

import (
	"os"

	"github.com/goodblaster/map-cache/internal/log"
)

var (
	KeyDelimiter = "/"
	WebAddress   = ":8080"
)

func Init(l log.Logger) {
	if val := os.Getenv("KEY_DELIMITER"); val != "" {
		KeyDelimiter = val
	}

	if val := os.Getenv("LISTEN_ADDRESS"); val != "" {
		WebAddress = val
	}

	log.
		With("KEY_DELIMITER", KeyDelimiter).
		With("LISTEN_ADDRESS", WebAddress).
		Info("Configuration initialized")
}

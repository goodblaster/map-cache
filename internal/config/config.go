package config

import (
	"os"

	"github.com/goodblaster/logos"
)

var (
	KeyDelimiter = "/"
	WebAddress   = ":8080"
)

func Init() {
	if val := os.Getenv("KEY_DELIMITER"); val != "" {
		KeyDelimiter = val
	}

	if val := os.Getenv("LISTEN_ADDRESS"); val != "" {
		WebAddress = val
	}

	logos.
		With("KEY_DELIMITER", KeyDelimiter).
		With("LISTEN_ADDRESS", WebAddress).
		Debug("Configuration initialized")
}

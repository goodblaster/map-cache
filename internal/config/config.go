package config

import "os"

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
}

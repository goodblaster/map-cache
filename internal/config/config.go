package config

import "os"

var (
	KeyDelimiter = "/"
)

func Init() {
	if os.Getenv("KEY_DELIMITER") != "" {
		KeyDelimiter = os.Getenv("KEY_DELIMITER")
	}
}

package mapkeys

import (
	"strings"

	"github.com/goodblaster/map-cache/internal/config"
)

func Split(key string) []string {
	return strings.Split(key, config.KeyDelimiter)
}

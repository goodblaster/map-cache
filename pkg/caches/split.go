package caches

import (
	"strings"

	"github.com/goodblaster/map-cache/internal/config"
)

func SplitKey(key string) []string {
	return strings.Split(key, config.KeyDelimiter)
}

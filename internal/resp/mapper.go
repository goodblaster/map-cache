package resp

import (
	"strings"

	"github.com/goodblaster/map-cache/internal/config"
)

// TranslateKey converts Redis-style keys to map-cache paths.
// If RESP_KEY_MODE is "translate", converts ":" to "/".
// Otherwise preserves the key as-is.
func TranslateKey(redisKey string) string {
	if config.RESPKeyMode == "translate" {
		return strings.ReplaceAll(redisKey, ":", "/")
	}
	return redisKey
}

// TranslateKeys translates multiple Redis keys to map-cache paths.
func TranslateKeys(redisKeys []string) []string {
	if config.RESPKeyMode != "translate" {
		return redisKeys
	}

	result := make([]string, len(redisKeys))
	for i, key := range redisKeys {
		result[i] = TranslateKey(key)
	}
	return result
}

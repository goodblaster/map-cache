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
		return strings.ReplaceAll(redisKey, ":", config.KeyDelimiter)
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

// UntranslateKey converts a map-cache path back to a Redis-style key.
// It is the inverse of TranslateKey.
func UntranslateKey(key string) string {
	if config.RESPKeyMode == "translate" {
		return strings.ReplaceAll(key, config.KeyDelimiter, ":")
	}
	return key
}

// fieldPath returns the map-cache path of a field within a hash key.
func fieldPath(key, field string) string {
	return key + config.KeyDelimiter + field
}

// deepCopy returns a copy of a cache value that shares no objects or arrays
// with the original.
func deepCopy(value any) any {
	switch v := value.(type) {
	case map[string]any:
		out := make(map[string]any, len(v))
		for k, child := range v {
			out[k] = deepCopy(child)
		}
		return out
	case []any:
		out := make([]any, len(v))
		for i, child := range v {
			out[i] = deepCopy(child)
		}
		return out
	default:
		return value
	}
}

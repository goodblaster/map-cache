package caches

import (
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/goodblaster/map-cache/internal/config"
)

func SplitKey(key string) []string {
	return strings.Split(key, config.KeyDelimiter)
}

// trimDelimiters removes leading and trailing key delimiters.
func trimDelimiters(key string) string {
	delim := config.KeyDelimiter
	if delim == "" {
		return key
	}
	for strings.HasPrefix(key, delim) {
		key = key[len(delim):]
	}
	for strings.HasSuffix(key, delim) {
		key = key[:len(key)-len(delim)]
	}
	return key
}

// matchPattern reports whether key matches pattern segment by segment, where a
// "*" segment in pattern matches any single non-empty segment. On a match it
// returns the segments matched by wildcards, in order. A mismatch allocates
// nothing, which matters because every write is checked against every trigger.
func matchPattern(key, pattern string) ([]string, bool) {
	key, pattern = trimDelimiters(key), trimDelimiters(pattern)
	if !walkPattern(key, pattern, nil) {
		return nil, false
	}
	if !strings.Contains(pattern, "*") {
		return nil, true
	}
	var captures []string
	walkPattern(key, pattern, &captures)
	return captures, true
}

func walkPattern(key, pattern string, captures *[]string) bool {
	delim := config.KeyDelimiter
	for {
		keySeg, keyRest, keyMore := strings.Cut(key, delim)
		patSeg, patRest, patMore := strings.Cut(pattern, delim)

		if patSeg == "*" {
			if keySeg == "" {
				return false
			}
			if captures != nil {
				*captures = append(*captures, keySeg)
			}
		} else if patSeg != keySeg {
			return false
		}

		if keyMore != patMore {
			return false
		}
		if !keyMore {
			return true
		}
		key, pattern = keyRest, patRest
	}
}

// FormatValue renders a cache value as text for templates, PRINT and the Redis
// protocol. Unlike fmt's %v it never uses exponent notation for ordinary
// numbers: JSON numbers are stored as float64, and 1000000 must not print as
// "1e+06".
func FormatValue(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case float64:
		if math.Abs(t) < 1e21 {
			return strconv.FormatFloat(t, 'f', -1, 64)
		}
		return strconv.FormatFloat(t, 'g', -1, 64)
	case int:
		return strconv.Itoa(t)
	case int64:
		return strconv.FormatInt(t, 10)
	default:
		return fmt.Sprint(v)
	}
}

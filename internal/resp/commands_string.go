package resp

import (
	"context"
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"

	"github.com/goodblaster/map-cache/pkg/caches"
	respProto "github.com/tidwall/resp"
)

func init() {
	// Register string commands
	RegisterCommand("GET", HandleGet)
	RegisterCommand("SET", HandleSet)
	RegisterCommand("DEL", HandleDel)
	RegisterCommand("EXISTS", HandleExists)
	RegisterCommand("INCR", HandleIncr)
	RegisterCommand("DECR", HandleDecr)
	RegisterCommand("INCRBY", HandleIncrBy)
	RegisterCommand("DECRBY", HandleDecrBy)
	RegisterCommand("MGET", HandleMGet)
	RegisterCommand("MSET", HandleMSet)
	RegisterCommand("GETSET", HandleGetSet)
	RegisterCommand("SETNX", HandleSetNX)
	RegisterCommand("SETEX", HandleSetEX)
	RegisterCommand("PSETEX", HandlePSetEX)
	RegisterCommand("STRLEN", HandleStrLen)
	RegisterCommand("APPEND", HandleAppend)
	RegisterCommand("GETRANGE", HandleGetRange)
	RegisterCommand("SETRANGE", HandleSetRange)
	RegisterCommand("GETEX", HandleGetEX)
	RegisterCommand("GETDEL", HandleGetDel)
	RegisterCommand("INCRBYFLOAT", HandleIncrByFloat)
}

// HandleGet implements the GET command
func HandleGet(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'get' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("GET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	value, err := cache.Get(ctx, key)
	if err != nil {
		// Key not found - return null
		return s.WriteValue(NullBulkString())
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleSet implements the SET command
func HandleSet(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'set' command")
	}

	key := TranslateKey(args[0].String())
	value := args[1].String()

	// Parse optional arguments (EX, PX, NX, XX, etc.)
	var ttlMillis *int64
	nx := false      // Only set if key doesn't exist
	xx := false      // Only set if key exists
	keepTTL := false // Retain the existing TTL

	for i := 2; i < len(args); i++ {
		opt := strings.ToUpper(args[i].String())
		switch opt {
		case "EX": // Expire in seconds
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			seconds, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			ttl := seconds * 1000
			ttlMillis = &ttl
		case "PX": // Expire in milliseconds
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			ms, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			ttlMillis = &ms
		case "NX":
			nx = true
		case "XX":
			xx = true
		case "KEEPTTL":
			keepTTL = true
		default:
			return s.WriteError(fmt.Sprintf("ERR unknown option '%s'", opt))
		}
	}

	if (nx && xx) || (keepTTL && ttlMillis != nil) {
		return s.WriteError("ERR syntax error")
	}
	if ttlMillis != nil && *ttlMillis <= 0 {
		return s.WriteError("ERR invalid expire time in 'set' command")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("SET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check NX/XX conditions
	_, err = cache.Get(ctx, key)
	exists := err == nil
	if nx && exists {
		// NX: only set if doesn't exist, but it exists
		return s.WriteValue(NullBulkString())
	}
	if xx && !exists {
		// XX: only set if exists, but it doesn't
		return s.WriteValue(NullBulkString())
	}

	// Set the value (create if doesn't exist, replace if does)
	if exists {
		if err := cache.Replace(ctx, key, value); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else {
		if err := cache.Create(ctx, map[string]any{key: value}); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	// Like Redis, a plain SET discards any existing TTL unless KEEPTTL is given.
	if ttlMillis != nil {
		if err := cache.SetKeyTTL(ctx, key, *ttlMillis); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else if !keepTTL {
		if err := cache.CancelKeyTTL(ctx, key); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	return s.WriteOK()
}

// HandleDel implements the DEL command
func HandleDel(s *Session, args []respProto.Value) error {
	if len(args) == 0 {
		return s.WriteError("ERR wrong number of arguments for 'del' command")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = TranslateKey(arg.String())
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("DEL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	deletedCount := 0
	for _, key := range keys {
		// Delete succeeds for missing keys too, so check existence to report
		// the number of keys actually removed.
		if _, err := cache.Get(ctx, key); err != nil {
			continue
		}
		if err := cache.Delete(ctx, key); err == nil {
			deletedCount++
		}
	}

	return s.WriteValue(Integer(deletedCount))
}

// HandleExists implements the EXISTS command
func HandleExists(s *Session, args []respProto.Value) error {
	if len(args) == 0 {
		return s.WriteError("ERR wrong number of arguments for 'exists' command")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = TranslateKey(arg.String())
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("EXISTS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	existsCount := 0
	for _, key := range keys {
		if _, err := cache.Get(ctx, key); err == nil {
			existsCount++
		}
	}

	return s.WriteValue(Integer(existsCount))
}

// HandleIncr implements the INCR command
func HandleIncr(s *Session, args []respProto.Value) error {
	return handleIncrBy(s, args, 1)
}

// HandleDecr implements the DECR command
func HandleDecr(s *Session, args []respProto.Value) error {
	return handleIncrBy(s, args, -1)
}

// HandleIncrBy implements the INCRBY command
func HandleIncrBy(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'incrby' command")
	}

	increment, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	return handleIncrBy(s, []respProto.Value{args[0]}, increment)
}

// HandleDecrBy implements the DECRBY command
func HandleDecrBy(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'decrby' command")
	}

	decrement, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	return handleIncrBy(s, []respProto.Value{args[0]}, -decrement)
}

// maxExactFloatInt is the largest magnitude at which every integer is exactly
// representable as a float64 (the type JSON numbers are stored as).
const maxExactFloatInt = 1 << 53

// toRedisInt interprets a stored value as a Redis integer. Values written with
// SET are stored as strings, so "10" must be accepted as well as numbers.
func toRedisInt(value any) (int64, bool) {
	if str, ok := value.(string); ok {
		n, err := strconv.ParseInt(str, 10, 64)
		return n, err == nil
	}
	return caches.ToInt64(value)
}

// handleIncrBy is a helper for INCR, DECR, INCRBY, DECRBY
func handleIncrBy(s *Session, args []respProto.Value, increment int64) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("INCR")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// A missing key counts as 0
	var current int64
	value, getErr := cache.Get(ctx, key)
	exists := getErr == nil
	if exists {
		n, ok := toRedisInt(value)
		if !ok {
			return s.WriteError("ERR value is not an integer or out of range")
		}
		current = n
	}

	if (increment > 0 && current > math.MaxInt64-increment) ||
		(increment < 0 && current < math.MinInt64-increment) {
		return s.WriteError("ERR increment or decrement would overflow")
	}
	newValue := current + increment

	// Store as a JSON number while that is exact; beyond 2^53 a float64 would
	// silently lose precision, so fall back to a decimal string (which is how
	// Redis itself stores integers).
	var stored any = float64(newValue)
	if newValue > maxExactFloatInt || newValue < -maxExactFloatInt {
		stored = strconv.FormatInt(newValue, 10)
	}

	if exists {
		err = cache.Replace(ctx, key, stored)
	} else {
		err = cache.Create(ctx, map[string]any{key: stored})
	}
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	return s.WriteValue(Integer(int(newValue)))
}

// HandleMGet implements the MGET command
func HandleMGet(s *Session, args []respProto.Value) error {
	if len(args) == 0 {
		return s.WriteError("ERR wrong number of arguments for 'mget' command")
	}

	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = TranslateKey(arg.String())
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("MGET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	result := make([]respProto.Value, len(keys))
	for i, key := range keys {
		value, err := cache.Get(ctx, key)
		if err != nil {
			// Key not found - return null for this position
			result[i] = NullBulkString()
		} else {
			result[i] = ConvertToRESP(value)
		}
	}

	return s.WriteValue(Array(result))
}

// HandleMSet implements the MSET command
func HandleMSet(s *Session, args []respProto.Value) error {
	if len(args) == 0 || len(args)%2 != 0 {
		return s.WriteError("ERR wrong number of arguments for 'mset' command")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("MSET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Set all key-value pairs
	for i := 0; i < len(args); i += 2 {
		key := TranslateKey(args[i].String())
		value := args[i+1].String()

		// Create or replace
		_, getErr := cache.Get(ctx, key)
		if getErr == nil {
			if err := cache.Replace(ctx, key, value); err != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
			}
		} else {
			if err := cache.Create(ctx, map[string]any{key: value}); err != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
			}
		}
	}

	return s.WriteOK()
}

// HandleGetSet implements the GETSET command
func HandleGetSet(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'getset' command")
	}

	key := TranslateKey(args[0].String())
	newValue := args[1].String()

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("GETSET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get old value
	oldValue, err := cache.Get(ctx, key)
	var respValue respProto.Value
	if err != nil {
		respValue = NullBulkString()
	} else {
		respValue = ConvertToRESP(oldValue)
	}

	// Set new value (create or replace)
	if err == nil {
		if err := cache.Replace(ctx, key, newValue); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else {
		if err := cache.Create(ctx, map[string]any{key: newValue}); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	return s.WriteValue(respValue)
}

// HandleSetNX implements the SETNX command
func HandleSetNX(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'setnx' command")
	}

	key := TranslateKey(args[0].String())
	value := args[1].String()

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("SETNX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	if _, err := cache.Get(ctx, key); err == nil {
		// Key exists, don't set
		return s.WriteValue(Integer(0))
	}

	// Key doesn't exist, create it
	if err := cache.Create(ctx, map[string]any{key: value}); err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	return s.WriteValue(Integer(1))
}

// HandleSetEX implements the SETEX command
func HandleSetEX(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'setex' command")
	}

	key := TranslateKey(args[0].String())
	seconds, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	value := args[2].String()
	if seconds <= 0 {
		return s.WriteError("ERR invalid expire time in 'setex' command")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("SETEX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Set value (create or replace)
	_, getErr := cache.Get(ctx, key)
	if getErr == nil {
		if err := cache.Replace(ctx, key, value); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else {
		if err := cache.Create(ctx, map[string]any{key: value}); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	// Set TTL
	if err := cache.SetKeyTTL(ctx, key, seconds*1000); err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	return s.WriteOK()
}

// HandlePSetEX implements the PSETEX command (set with millisecond expiration)
func HandlePSetEX(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'psetex' command")
	}

	key := TranslateKey(args[0].String())
	milliseconds, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	value := args[2].String()
	if milliseconds <= 0 {
		return s.WriteError("ERR invalid expire time in 'psetex' command")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("PSETEX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Set value (create or replace)
	_, getErr := cache.Get(ctx, key)
	if getErr == nil {
		if err := cache.Replace(ctx, key, value); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else {
		if err := cache.Create(ctx, map[string]any{key: value}); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	// Set TTL in milliseconds
	if err := cache.SetKeyTTL(ctx, key, milliseconds); err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	return s.WriteOK()
}

// HandleStrLen implements the STRLEN command
func HandleStrLen(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'strlen' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("STRLEN")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	value, err := cache.Get(ctx, key)
	if err != nil {
		// Key not found - return 0
		return s.WriteValue(Integer(0))
	}

	// Convert to string and get length
	str := caches.FormatValue(value)
	return s.WriteValue(Integer(len(str)))
}

// HandleAppend implements the APPEND command
func HandleAppend(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'append' command")
	}

	key := TranslateKey(args[0].String())
	appendValue := args[1].String()

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("APPEND")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing value
	existingValue, err := cache.Get(ctx, key)
	var newValue string
	if err != nil {
		// Key doesn't exist, use append value as new value
		newValue = appendValue
		// Create the key
		if err := cache.Create(ctx, map[string]any{key: newValue}); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	} else {
		// Concatenate
		newValue = caches.FormatValue(existingValue) + appendValue
		// Replace the key
		if err := cache.Replace(ctx, key, newValue); err != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
		}
	}

	return s.WriteValue(Integer(len(newValue)))
}

// HandleGetRange implements the GETRANGE command (get substring)
func HandleGetRange(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'getrange' command")
	}

	key := TranslateKey(args[0].String())
	start, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	end, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("GETRANGE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return empty string
		return s.WriteValue(BulkString(""))
	}

	str := caches.FormatValue(value)
	length := int64(len(str))

	// Handle negative indices
	if start < 0 {
		start = length + start
	}
	if end < 0 {
		end = length + end
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if end >= length {
		end = length - 1
	}

	if start > end || start >= length {
		return s.WriteValue(BulkString(""))
	}

	substring := str[start : end+1]
	return s.WriteValue(BulkString(substring))
}

// maxStringSize is the largest string value SETRANGE will build (512MB).
const maxStringSize = 512 * 1024 * 1024

// HandleSetRange implements the SETRANGE command (set substring)
func HandleSetRange(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'setrange' command")
	}

	key := TranslateKey(args[0].String())
	offset, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	newValue := args[2].String()

	if offset < 0 {
		return s.WriteError("ERR offset is out of range")
	}
	// Same limit as Redis (512MB). Without it a single command can ask for an
	// arbitrarily large allocation.
	if offset > maxStringSize || offset+int64(len(newValue)) > maxStringSize {
		return s.WriteError("ERR string exceeds maximum allowed size (proto-max-bulk-len)")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("SETRANGE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing value or create empty string
	value, getErr := cache.Get(ctx, key)
	var str string
	if getErr != nil {
		str = ""
	} else {
		str = caches.FormatValue(value)
	}

	// Pad with zeros if needed
	if int64(len(str)) < offset {
		str = str + strings.Repeat("\x00", int(offset)-len(str))
	}

	// Replace substring
	result := str[:offset] + newValue
	if int64(len(str)) > offset+int64(len(newValue)) {
		result = result + str[offset+int64(len(newValue)):]
	}

	// Store result
	if getErr != nil {
		// Key didn't exist, create it
		if createErr := cache.Create(ctx, map[string]any{key: result}); createErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
		}
	} else {
		// Key exists, replace it
		if replaceErr := cache.Replace(ctx, key, result); replaceErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
		}
	}

	return s.WriteValue(Integer(len(result)))
}

// HandleGetEX implements the GETEX command (GET with expiration options)
func HandleGetEX(s *Session, args []respProto.Value) error {
	if len(args) < 1 {
		return s.WriteError("ERR wrong number of arguments for 'getex' command")
	}

	key := TranslateKey(args[0].String())

	// Parse expiration options
	var expirationMs int64
	hasExpiration := false
	persist := false
	relative := false // EX/PX: a non-positive value is an error, not "already expired"

	for i := 1; i < len(args); i++ {
		option := strings.ToUpper(args[i].String())
		switch option {
		case "EX":
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			seconds, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			expirationMs = seconds * 1000
			hasExpiration = true
			relative = true
		case "PX":
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			ms, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			expirationMs = ms
			hasExpiration = true
			relative = true
		case "EXAT":
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			timestamp, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			now := time.Now().Unix()
			expirationMs = (timestamp - now) * 1000
			hasExpiration = true
		case "PXAT":
			if i+1 >= len(args) {
				return s.WriteError("ERR syntax error")
			}
			i++
			timestampMs, err := strconv.ParseInt(args[i].String(), 10, 64)
			if err != nil {
				return s.WriteError("ERR value is not an integer or out of range")
			}
			nowMs := time.Now().UnixMilli()
			expirationMs = timestampMs - nowMs
			hasExpiration = true
		case "PERSIST":
			persist = true
		default:
			return s.WriteError("ERR syntax error")
		}
	}

	if persist && hasExpiration {
		return s.WriteError("ERR syntax error")
	}
	if relative && expirationMs <= 0 {
		return s.WriteError("ERR invalid expire time in 'getex' command")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("GETEX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get value
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		return s.WriteValue(NullBulkString())
	}

	switch {
	case persist:
		// Remove any existing TTL
		cache.CancelKeyTTL(ctx, key)
	case hasExpiration && expirationMs > 0:
		cache.SetKeyTTL(ctx, key, expirationMs)
	case hasExpiration:
		// EXAT/PXAT in the past: the key expires immediately
		cache.Delete(ctx, key)
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleGetDel implements the GETDEL command (atomic get and delete)
func HandleGetDel(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'getdel' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("GETDEL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get value
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		return s.WriteValue(NullBulkString())
	}

	// Delete key
	if delErr := cache.Delete(ctx, key); delErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", delErr.Error()))
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleIncrByFloat implements the INCRBYFLOAT command
func HandleIncrByFloat(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'incrbyfloat' command")
	}

	key := TranslateKey(args[0].String())
	increment, err := strconv.ParseFloat(args[1].String(), 64)
	if err != nil {
		return s.WriteError("ERR value is not a valid float")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("INCRBYFLOAT")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get current value or initialize to 0
	value, getErr := cache.Get(ctx, key)
	var currentVal float64
	if getErr != nil {
		// Key doesn't exist, create it with 0
		currentVal = 0.0
		if createErr := cache.Create(ctx, map[string]any{key: currentVal}); createErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
		}
	} else {
		// Parse current value
		switch v := value.(type) {
		case float64:
			currentVal = v
		case int:
			currentVal = float64(v)
		case string:
			parsed, parseErr := strconv.ParseFloat(v, 64)
			if parseErr != nil {
				return s.WriteError("ERR value is not a valid float")
			}
			currentVal = parsed
		default:
			return s.WriteError("ERR value is not a valid float")
		}
	}

	// Increment
	newValue := currentVal + increment

	// Store new value
	if replaceErr := cache.Replace(ctx, key, newValue); replaceErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
	}

	// Return new value as string (Redis INCRBYFLOAT returns string representation)
	return s.WriteValue(BulkString(fmt.Sprintf("%.17g", newValue)))
}

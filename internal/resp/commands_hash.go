package resp

import (
	"context"
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/goodblaster/map-cache/pkg/caches"
	respProto "github.com/tidwall/resp"
)

func init() {
	// Register hash commands
	RegisterCommand("HGET", HandleHGet)
	RegisterCommand("HSET", HandleHSet)
	RegisterCommand("HGETALL", HandleHGetAll)
	RegisterCommand("HDEL", HandleHDel)
	RegisterCommand("HEXISTS", HandleHExists)
	RegisterCommand("HLEN", HandleHLen)
	RegisterCommand("HKEYS", HandleHKeys)
	RegisterCommand("HVALS", HandleHVals)
	RegisterCommand("HMGET", HandleHMGet)
	RegisterCommand("HMSET", HandleHMSet)
	RegisterCommand("HINCRBY", HandleHIncrBy)
	RegisterCommand("HINCRBYFLOAT", HandleHIncrByFloat)
	RegisterCommand("HSETNX", HandleHSetNX)
	RegisterCommand("HRANDFIELD", HandleHRandField)
}

// HandleHGet implements the HGET command
func HandleHGet(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'hget' command")
	}

	key := TranslateKey(args[0].String())
	field := args[1].String()
	fullPath := key + "/" + field

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HGET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	value, getErr := cache.Get(ctx, fullPath)
	if getErr != nil {
		// Field doesn't exist, return nil
		return s.WriteValue(NullBulkString())
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleHSet implements the HSET command
func HandleHSet(s *Session, args []respProto.Value) error {
	if len(args) < 3 || len(args)%2 != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hset' command")
	}

	key := TranslateKey(args[0].String())
	numFieldsSet := 0

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HSET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Process field-value pairs
	for i := 1; i < len(args); i += 2 {
		field := args[i].String()
		value := args[i+1].String()
		fullPath := key + "/" + field

		// Check if field exists
		_, getErr := cache.Get(ctx, fullPath)
		fieldExists := (getErr == nil)

		if fieldExists {
			// Update existing field
			if replaceErr := cache.Replace(ctx, fullPath, value); replaceErr != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
			}
		} else {
			// Create new field
			if createErr := cache.Create(ctx, map[string]any{fullPath: value}); createErr != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
			}
			numFieldsSet++
		}
	}

	// Return number of fields that were added (not updated)
	return s.WriteValue(Integer(numFieldsSet))
}

// HandleHGetAll implements the HGETALL command
func HandleHGetAll(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hgetall' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HGETALL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get the hash object
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Hash doesn't exist, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Convert to map if it's an object
	valueMap, ok := value.(map[string]any)
	if !ok {
		// Not a hash, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Build field-value pairs
	result := make([]respProto.Value, 0, len(valueMap)*2)
	for field, val := range valueMap {
		result = append(result, BulkString(field))
		result = append(result, ConvertToRESP(val))
	}

	return s.WriteValue(Array(result))
}

// HandleHDel implements the HDEL command
func HandleHDel(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'hdel' command")
	}

	key := TranslateKey(args[0].String())
	numDeleted := 0

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HDEL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Delete each field
	for i := 1; i < len(args); i++ {
		field := args[i].String()
		fullPath := key + "/" + field

		// Check if field exists
		_, getErr := cache.Get(ctx, fullPath)
		if getErr == nil {
			// Field exists, delete it
			if delErr := cache.Delete(ctx, fullPath); delErr == nil {
				numDeleted++
			}
		}
	}

	return s.WriteValue(Integer(numDeleted))
}

// HandleHExists implements the HEXISTS command
func HandleHExists(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'hexists' command")
	}

	key := TranslateKey(args[0].String())
	field := args[1].String()
	fullPath := key + "/" + field

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HEXISTS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	_, getErr := cache.Get(ctx, fullPath)
	if getErr != nil {
		// Field doesn't exist
		return s.WriteValue(Integer(0))
	}

	return s.WriteValue(Integer(1))
}

// HandleHLen implements the HLEN command
func HandleHLen(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hlen' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HLEN")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get the hash object
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Hash doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Convert to map if it's an object
	valueMap, ok := value.(map[string]any)
	if !ok {
		// Not a hash, return 0
		return s.WriteValue(Integer(0))
	}

	return s.WriteValue(Integer(len(valueMap)))
}

// HandleHKeys implements the HKEYS command
func HandleHKeys(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hkeys' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HKEYS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get the hash object
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Hash doesn't exist, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Convert to map if it's an object
	valueMap, ok := value.(map[string]any)
	if !ok {
		// Not a hash, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Build array of field names
	result := make([]respProto.Value, 0, len(valueMap))
	for field := range valueMap {
		result = append(result, BulkString(field))
	}

	return s.WriteValue(Array(result))
}

// HandleHVals implements the HVALS command
func HandleHVals(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hvals' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HVALS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get the hash object
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Hash doesn't exist, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Convert to map if it's an object
	valueMap, ok := value.(map[string]any)
	if !ok {
		// Not a hash, return empty array
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Build array of values
	result := make([]respProto.Value, 0, len(valueMap))
	for _, val := range valueMap {
		result = append(result, ConvertToRESP(val))
	}

	return s.WriteValue(Array(result))
}

// HandleHMGet implements the HMGET command
func HandleHMGet(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'hmget' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HMGET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get each field
	result := make([]respProto.Value, 0, len(args)-1)
	for i := 1; i < len(args); i++ {
		field := args[i].String()
		fullPath := key + "/" + field

		value, getErr := cache.Get(ctx, fullPath)
		if getErr != nil {
			// Field doesn't exist, add nil
			result = append(result, NullBulkString())
		} else {
			result = append(result, ConvertToRESP(value))
		}
	}

	return s.WriteValue(Array(result))
}

// HandleHMSet implements the HMSET command (deprecated in Redis, but still supported)
func HandleHMSet(s *Session, args []respProto.Value) error {
	if len(args) < 3 || len(args)%2 != 1 {
		return s.WriteError("ERR wrong number of arguments for 'hmset' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HMSET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Set each field-value pair
	for i := 1; i < len(args); i += 2 {
		field := args[i].String()
		value := args[i+1].String()
		fullPath := key + "/" + field

		// Check if field exists
		_, getErr := cache.Get(ctx, fullPath)
		if getErr == nil {
			// Update existing field
			if replaceErr := cache.Replace(ctx, fullPath, value); replaceErr != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
			}
		} else {
			// Create new field
			if createErr := cache.Create(ctx, map[string]any{fullPath: value}); createErr != nil {
				return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
			}
		}
	}

	return s.WriteOK()
}

// HandleHIncrBy implements the HINCRBY command (increment hash field by integer)
func HandleHIncrBy(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'hincrby' command")
	}

	key := TranslateKey(args[0].String())
	field := args[1].String()
	increment, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	fullPath := key + "/" + field

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("HINCRBY")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get current value or initialize to 0
	value, getErr := cache.Get(ctx, fullPath)
	var currentVal int64
	if getErr != nil {
		// Field doesn't exist, initialize to 0
		currentVal = 0
		if createErr := cache.Create(ctx, map[string]any{fullPath: currentVal}); createErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
		}
	} else {
		// Parse current value
		switch v := value.(type) {
		case int64:
			currentVal = v
		case int:
			currentVal = int64(v)
		case float64:
			currentVal = int64(v)
		case string:
			parsed, parseErr := strconv.ParseInt(v, 10, 64)
			if parseErr != nil {
				return s.WriteError("ERR hash value is not an integer")
			}
			currentVal = parsed
		default:
			return s.WriteError("ERR hash value is not an integer")
		}
	}

	// Increment
	newValue := currentVal + increment

	// Store new value
	if replaceErr := cache.Replace(ctx, fullPath, newValue); replaceErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
	}

	return s.WriteValue(Integer(int(newValue)))
}

// HandleHIncrByFloat implements the HINCRBYFLOAT command (increment hash field by float)
func HandleHIncrByFloat(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'hincrbyfloat' command")
	}

	key := TranslateKey(args[0].String())
	field := args[1].String()
	increment, err := strconv.ParseFloat(args[2].String(), 64)
	if err != nil {
		return s.WriteError("ERR value is not a valid float")
	}

	fullPath := key + "/" + field

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("HINCRBYFLOAT")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get current value or initialize to 0
	value, getErr := cache.Get(ctx, fullPath)
	var currentVal float64
	if getErr != nil {
		// Field doesn't exist, initialize to 0
		currentVal = 0.0
		if createErr := cache.Create(ctx, map[string]any{fullPath: currentVal}); createErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
		}
	} else {
		// Parse current value
		switch v := value.(type) {
		case float64:
			currentVal = v
		case int:
			currentVal = float64(v)
		case int64:
			currentVal = float64(v)
		case string:
			parsed, parseErr := strconv.ParseFloat(v, 64)
			if parseErr != nil {
				return s.WriteError("ERR hash value is not a valid float")
			}
			currentVal = parsed
		default:
			return s.WriteError("ERR hash value is not a valid float")
		}
	}

	// Increment
	newValue := currentVal + increment

	// Store new value
	if replaceErr := cache.Replace(ctx, fullPath, newValue); replaceErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", replaceErr.Error()))
	}

	// Return new value as string (like Redis)
	return s.WriteValue(BulkString(fmt.Sprintf("%.17g", newValue)))
}

// HandleHSetNX implements the HSETNX command (set field if not exists)
func HandleHSetNX(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'hsetnx' command")
	}

	key := TranslateKey(args[0].String())
	field := args[1].String()
	value := args[2].String()
	fullPath := key + "/" + field

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HSETNX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if field exists
	_, getErr := cache.Get(ctx, fullPath)
	if getErr == nil {
		// Field exists, return 0
		return s.WriteValue(Integer(0))
	}

	// Field doesn't exist, create it
	if createErr := cache.Create(ctx, map[string]any{fullPath: value}); createErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
	}

	return s.WriteValue(Integer(1))
}

// HandleHRandField implements the HRANDFIELD command (get random field(s) from hash)
func HandleHRandField(s *Session, args []respProto.Value) error {
	if len(args) < 1 {
		return s.WriteError("ERR wrong number of arguments for 'hrandfield' command")
	}

	key := TranslateKey(args[0].String())

	// Parse optional count argument
	count := 1
	withValues := false
	if len(args) >= 2 {
		var err error
		count64, err := strconv.ParseInt(args[1].String(), 10, 64)
		if err != nil {
			return s.WriteError("ERR value is not an integer or out of range")
		}
		count = int(count64)

		// Check for WITHVALUES flag
		if len(args) >= 3 && args[2].String() == "WITHVALUES" {
			withValues = true
		}
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("HRANDFIELD")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get hash value
	hashValue, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Hash doesn't exist, return null
		if count == 1 && !withValues {
			return s.WriteValue(NullBulkString())
		}
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Ensure it's a hash
	hash, ok := hashValue.(map[string]any)
	if !ok {
		return s.WriteError("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	// Get all field names
	fields := make([]string, 0, len(hash))
	for field := range hash {
		fields = append(fields, field)
	}

	if len(fields) == 0 {
		// Empty hash
		if count == 1 && !withValues {
			return s.WriteValue(NullBulkString())
		}
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Shuffle fields for randomness
	rand.Shuffle(len(fields), func(i, j int) {
		fields[i], fields[j] = fields[j], fields[i]
	})

	// Select random fields
	absCount := count
	if absCount < 0 {
		absCount = -absCount
	}
	if absCount > len(fields) {
		absCount = len(fields)
	}

	selectedFields := fields[:absCount]

	// Return single field if count was not specified or is 1
	if count == 1 && len(args) < 2 {
		return s.WriteValue(BulkString(selectedFields[0]))
	}

	// Return array of fields or field-value pairs
	if withValues {
		result := make([]respProto.Value, 0, len(selectedFields)*2)
		for _, field := range selectedFields {
			result = append(result, BulkString(field))
			value := hash[field]
			result = append(result, ConvertToRESP(value))
		}
		return s.WriteValue(Array(result))
	}

	// Return array of field names
	result := make([]respProto.Value, len(selectedFields))
	for i, field := range selectedFields {
		result[i] = BulkString(field)
	}
	return s.WriteValue(Array(result))
}

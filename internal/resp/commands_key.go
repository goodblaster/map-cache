package resp

import (
	"context"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/goodblaster/map-cache/pkg/caches"
	respProto "github.com/tidwall/resp"
)

func init() {
	// Register key management commands
	RegisterCommand("EXPIRE", HandleExpire)
	RegisterCommand("PEXPIRE", HandlePExpire)
	RegisterCommand("PERSIST", HandlePersist)
	RegisterCommand("TTL", HandleTTL)
	RegisterCommand("PTTL", HandlePTTL)
	RegisterCommand("EXPIRETIME", HandleExpireTime)
	RegisterCommand("PEXPIRETIME", HandlePExpireTime)
	RegisterCommand("KEYS", HandleKeys)
	RegisterCommand("EXPIREAT", HandleExpireAt)
	RegisterCommand("PEXPIREAT", HandlePExpireAt)
	RegisterCommand("RENAME", HandleRename)
	RegisterCommand("RENAMENX", HandleRenameNX)
	RegisterCommand("TYPE", HandleType)
	RegisterCommand("COPY", HandleCopy)
}

// HandleExpire implements the EXPIRE command (set TTL in seconds)
func HandleExpire(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'expire' command")
	}

	key := TranslateKey(args[0].String())
	seconds, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("EXPIRE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Set TTL in milliseconds (Redis EXPIRE uses seconds)
	milliseconds := seconds * 1000
	if setErr := cache.SetKeyTTL(ctx, key, milliseconds); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	// Return 1 for success
	return s.WriteValue(Integer(1))
}

// HandlePExpire implements the PEXPIRE command (set TTL in milliseconds)
func HandlePExpire(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'pexpire' command")
	}

	key := TranslateKey(args[0].String())
	milliseconds, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("PEXPIRE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Set TTL in milliseconds
	if setErr := cache.SetKeyTTL(ctx, key, milliseconds); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	// Return 1 for success
	return s.WriteValue(Integer(1))
}

// HandlePersist implements the PERSIST command (remove TTL)
func HandlePersist(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'persist' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("PERSIST")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists and has a TTL
	keyExps := cache.KeyExpirations()
	_, hasTTL := keyExps[key]

	if !hasTTL {
		// Key doesn't have TTL (or doesn't exist), return 0
		return s.WriteValue(Integer(0))
	}

	// Cancel TTL
	if cancelErr := cache.CancelKeyTTL(ctx, key); cancelErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cancelErr.Error()))
	}

	// Return 1 for success
	return s.WriteValue(Integer(1))
}

// HandleTTL implements the TTL command (get remaining TTL in seconds)
func HandleTTL(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'ttl' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("TTL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return -2
		return s.WriteValue(Integer(-2))
	}

	// Check if key has TTL
	keyExps := cache.KeyExpirations()
	timer, hasTTL := keyExps[key]

	if !hasTTL {
		// Key exists but has no TTL, return -1
		return s.WriteValue(Integer(-1))
	}

	// Calculate remaining TTL in seconds
	remainingSeconds := timer.Expiration - time.Now().Unix()
	if remainingSeconds < 0 {
		remainingSeconds = 0
	}

	return s.WriteValue(Integer(int(remainingSeconds)))
}

// HandlePTTL implements the PTTL command (get remaining TTL in milliseconds)
func HandlePTTL(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'pttl' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("PTTL")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return -2
		return s.WriteValue(Integer(-2))
	}

	// Check if key has TTL
	keyExps := cache.KeyExpirations()
	timer, hasTTL := keyExps[key]

	if !hasTTL {
		// Key exists but has no TTL, return -1
		return s.WriteValue(Integer(-1))
	}

	// Calculate remaining TTL in milliseconds
	remainingMs := (timer.Expiration - time.Now().Unix()) * 1000
	if remainingMs < 0 {
		remainingMs = 0
	}

	return s.WriteValue(Integer(int(remainingMs)))
}

// HandleExpireTime implements the EXPIRETIME command (get expiration timestamp in seconds)
func HandleExpireTime(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'expiretime' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("EXPIRETIME")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return -2
		return s.WriteValue(Integer(-2))
	}

	// Check if key has TTL
	keyExps := cache.KeyExpirations()
	timer, hasTTL := keyExps[key]

	if !hasTTL {
		// Key exists but has no TTL, return -1
		return s.WriteValue(Integer(-1))
	}

	// Return absolute expiration timestamp in seconds
	return s.WriteValue(Integer(int(timer.Expiration)))
}

// HandlePExpireTime implements the PEXPIRETIME command (get expiration timestamp in milliseconds)
func HandlePExpireTime(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'pexpiretime' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("PEXPIRETIME")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return -2
		return s.WriteValue(Integer(-2))
	}

	// Check if key has TTL
	keyExps := cache.KeyExpirations()
	timer, hasTTL := keyExps[key]

	if !hasTTL {
		// Key exists but has no TTL, return -1
		return s.WriteValue(Integer(-1))
	}

	// Return absolute expiration timestamp in milliseconds
	return s.WriteValue(Integer(int(timer.Expiration * 1000)))
}

// HandleKeys implements the KEYS command (pattern matching)
func HandleKeys(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'keys' command")
	}

	pattern := args[0].String()

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("KEYS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Translate Redis glob pattern to map-cache wildcard pattern
	// Redis uses * for any characters, map-cache uses * for path segments
	// For now, translate : to / and use as-is
	translatedPattern := strings.ReplaceAll(pattern, ":", "/")

	// Get matching keys
	keys := cache.WildKeys(ctx, translatedPattern)

	// Convert back to Redis format (/ → :)
	redisKeys := make([]respProto.Value, len(keys))
	for i, key := range keys {
		redisKey := strings.ReplaceAll(key, "/", ":")
		redisKeys[i] = BulkString(redisKey)
	}

	return s.WriteValue(Array(redisKeys))
}

// HandleExpireAt implements the EXPIREAT command (set expiration at timestamp in seconds)
func HandleExpireAt(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'expireat' command")
	}

	key := TranslateKey(args[0].String())
	timestamp, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("EXPIREAT")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Calculate TTL in milliseconds from timestamp
	now := time.Now().Unix()
	milliseconds := (timestamp - now) * 1000

	if milliseconds <= 0 {
		// Already expired, delete the key
		cache.Delete(ctx, key)
		return s.WriteValue(Integer(1))
	}

	// Set TTL
	if setErr := cache.SetKeyTTL(ctx, key, milliseconds); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(1))
}

// HandlePExpireAt implements the PEXPIREAT command (set expiration at timestamp in milliseconds)
func HandlePExpireAt(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'pexpireat' command")
	}

	key := TranslateKey(args[0].String())
	timestampMs, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("PEXPIREAT")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// Key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Calculate TTL in milliseconds from timestamp
	nowMs := time.Now().UnixMilli()
	milliseconds := timestampMs - nowMs

	if milliseconds <= 0 {
		// Already expired, delete the key
		cache.Delete(ctx, key)
		return s.WriteValue(Integer(1))
	}

	// Set TTL
	if setErr := cache.SetKeyTTL(ctx, key, milliseconds); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(1))
}

// HandleRename implements the RENAME command
func HandleRename(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'rename' command")
	}

	oldKey := TranslateKey(args[0].String())
	newKey := TranslateKey(args[1].String())

	if oldKey == newKey {
		// Same key, no-op but still return OK
		return s.WriteOK()
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("RENAME")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get old key value
	value, getErr := cache.Get(ctx, oldKey)
	if getErr != nil {
		return s.WriteError("ERR no such key")
	}

	// Delete new key if it exists
	cache.Delete(ctx, newKey)

	// Create new key
	if createErr := cache.Create(ctx, map[string]any{newKey: value}); createErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
	}

	// Copy TTL if it exists
	keyExps := cache.KeyExpirations()
	if timer, hasTTL := keyExps[oldKey]; hasTTL {
		remainingMs := (timer.Expiration - time.Now().Unix()) * 1000
		if remainingMs > 0 {
			cache.SetKeyTTL(ctx, newKey, remainingMs)
		}
	}

	// Delete old key
	if delErr := cache.Delete(ctx, oldKey); delErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", delErr.Error()))
	}

	return s.WriteOK()
}

// HandleRenameNX implements the RENAMENX command (rename if new key doesn't exist)
func HandleRenameNX(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'renamenx' command")
	}

	oldKey := TranslateKey(args[0].String())
	newKey := TranslateKey(args[1].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("RENAMENX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get old key value
	value, getErr := cache.Get(ctx, oldKey)
	if getErr != nil {
		return s.WriteError("ERR no such key")
	}

	// Check if new key exists
	_, newKeyErr := cache.Get(ctx, newKey)
	if newKeyErr == nil {
		// New key exists, return 0
		return s.WriteValue(Integer(0))
	}

	// Create new key
	if createErr := cache.Create(ctx, map[string]any{newKey: value}); createErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
	}

	// Copy TTL if it exists
	keyExps := cache.KeyExpirations()
	if timer, hasTTL := keyExps[oldKey]; hasTTL {
		remainingMs := (timer.Expiration - time.Now().Unix()) * 1000
		if remainingMs > 0 {
			cache.SetKeyTTL(ctx, newKey, remainingMs)
		}
	}

	// Delete old key
	if delErr := cache.Delete(ctx, oldKey); delErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", delErr.Error()))
	}

	return s.WriteValue(Integer(1))
}

// HandleType implements the TYPE command (get key type)
func HandleType(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'type' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("TYPE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get value
	value, getErr := cache.Get(ctx, key)
	if getErr != nil {
		return s.WriteValue(SimpleString("none"))
	}

	// Determine type
	var keyType string
	switch value.(type) {
	case map[string]any:
		keyType = "hash"
	case []any:
		keyType = "list"
	case string:
		keyType = "string"
	case int, int64, float64:
		keyType = "string" // Redis treats numbers as strings
	default:
		keyType = "string"
	}

	return s.WriteValue(SimpleString(keyType))
}

// HandleCopy implements the COPY command (copy key to destination)
func HandleCopy(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'copy' command")
	}

	sourceKey := TranslateKey(args[0].String())
	destKey := TranslateKey(args[1].String())

	// Parse optional REPLACE flag
	replace := false
	for i := 2; i < len(args); i++ {
		if strings.ToUpper(args[i].String()) == "REPLACE" {
			replace = true
		}
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("COPY")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get source key value
	sourceValue, getErr := cache.Get(ctx, sourceKey)
	if getErr != nil {
		// Source key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Check if destination key exists
	_, destErr := cache.Get(ctx, destKey)
	if destErr == nil && !replace {
		// Destination exists and REPLACE not specified, return 0
		return s.WriteValue(Integer(0))
	}

	// Delete destination if it exists (when REPLACE is specified)
	if destErr == nil {
		cache.Delete(ctx, destKey)
	}

	// Copy the value to destination
	if createErr := cache.Create(ctx, map[string]any{destKey: sourceValue}); createErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", createErr.Error()))
	}

	// Copy TTL if source has one
	keyExps := cache.KeyExpirations()
	if timer, hasTTL := keyExps[sourceKey]; hasTTL {
		remainingMs := (timer.Expiration - time.Now().Unix()) * 1000
		if remainingMs > 0 {
			cache.SetKeyTTL(ctx, destKey, remainingMs)
		}
	}

	return s.WriteValue(Integer(1))
}

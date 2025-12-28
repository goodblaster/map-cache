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
	// Register list commands
	RegisterCommand("LPUSH", HandleLPush)
	RegisterCommand("RPUSH", HandleRPush)
	RegisterCommand("LPUSHX", HandleLPushX)
	RegisterCommand("RPUSHX", HandleRPushX)
	RegisterCommand("LPOP", HandleLPop)
	RegisterCommand("RPOP", HandleRPop)
	RegisterCommand("LLEN", HandleLLen)
	RegisterCommand("LRANGE", HandleLRange)
	RegisterCommand("LINDEX", HandleLIndex)
	RegisterCommand("LSET", HandleLSet)
	RegisterCommand("LTRIM", HandleLTrim)
	RegisterCommand("LINSERT", HandleLInsert)
	RegisterCommand("LREM", HandleLRem)
	RegisterCommand("LPOS", HandleLPos)
}

// getList retrieves a list from cache, returning an empty list if key doesn't exist
func getList(ctx context.Context, cache *caches.Cache, key string) ([]any, error) {
	value, err := cache.Get(ctx, key)
	if err != nil {
		// Key doesn't exist, return empty list
		return []any{}, nil
	}

	// Check if value is already a list
	list, ok := value.([]any)
	if !ok {
		return nil, fmt.Errorf("WRONGTYPE Operation against a key holding the wrong kind of value")
	}

	return list, nil
}

// setList stores a list in cache
func setList(ctx context.Context, cache *caches.Cache, key string, list []any) error {
	// Check if key exists
	_, err := cache.Get(ctx, key)
	if err == nil {
		// Key exists, replace it
		return cache.Replace(ctx, key, list)
	}
	// Key doesn't exist, create it
	return cache.Create(ctx, map[string]any{key: list})
}

// HandleLPush implements the LPUSH command (prepend to list)
func HandleLPush(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'lpush' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("LPUSH")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	// Prepend values (in reverse order to maintain order)
	for i := len(args) - 1; i >= 1; i-- {
		value := args[i].String()
		list = append([]any{value}, list...)
	}

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleRPush implements the RPUSH command (append to list)
func HandleRPush(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'rpush' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("RPUSH")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	// Append values
	for i := 1; i < len(args); i++ {
		value := args[i].String()
		list = append(list, value)
	}

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleLPushX implements the LPUSHX command (prepend only if list exists)
func HandleLPushX(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'lpushx' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("LPUSHX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// List doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	// Prepend values in reverse order (so first arg becomes first element)
	for i := len(args) - 1; i >= 1; i-- {
		value := args[i].String()
		list = append([]any{value}, list...)
	}

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleRPushX implements the RPUSHX command (append only if list exists)
func HandleRPushX(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'rpushx' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("RPUSHX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Check if key exists
	_, getErr := cache.Get(ctx, key)
	if getErr != nil {
		// List doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	// Append values
	for i := 1; i < len(args); i++ {
		value := args[i].String()
		list = append(list, value)
	}

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleLPop implements the LPOP command (remove and return first element)
func HandleLPop(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'lpop' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("LPOP")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		// Empty list or key doesn't exist
		return s.WriteValue(NullBulkString())
	}

	// Pop first element
	value := list[0]
	list = list[1:]

	if len(list) == 0 {
		// List is now empty, delete the key
		cache.Delete(ctx, key)
	} else {
		// Store updated list
		if setErr := setList(ctx, cache, key, list); setErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
		}
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleRPop implements the RPOP command (remove and return last element)
func HandleRPop(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'rpop' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("RPOP")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		// Empty list or key doesn't exist
		return s.WriteValue(NullBulkString())
	}

	// Pop last element
	value := list[len(list)-1]
	list = list[:len(list)-1]

	if len(list) == 0 {
		// List is now empty, delete the key
		cache.Delete(ctx, key)
	} else {
		// Store updated list
		if setErr := setList(ctx, cache, key, list); setErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
		}
	}

	return s.WriteValue(ConvertToRESP(value))
}

// HandleLLen implements the LLEN command (get list length)
func HandleLLen(s *Session, args []respProto.Value) error {
	if len(args) != 1 {
		return s.WriteError("ERR wrong number of arguments for 'llen' command")
	}

	key := TranslateKey(args[0].String())

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("LLEN")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleLRange implements the LRANGE command (get range of elements)
func HandleLRange(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'lrange' command")
	}

	key := TranslateKey(args[0].String())
	start, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LRANGE")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Handle negative indices
	length := int64(len(list))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if start >= length {
		return s.WriteValue(Array([]respProto.Value{}))
	}
	if stop >= length {
		stop = length - 1
	}
	if stop < start {
		return s.WriteValue(Array([]respProto.Value{}))
	}

	// Get range
	result := make([]respProto.Value, 0, stop-start+1)
	for i := start; i <= stop; i++ {
		result = append(result, ConvertToRESP(list[i]))
	}

	return s.WriteValue(Array(result))
}

// HandleLIndex implements the LINDEX command (get element at index)
func HandleLIndex(s *Session, args []respProto.Value) error {
	if len(args) != 2 {
		return s.WriteError("ERR wrong number of arguments for 'lindex' command")
	}

	key := TranslateKey(args[0].String())
	index, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LINDEX")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteValue(NullBulkString())
	}

	// Handle negative index
	if index < 0 {
		index = int64(len(list)) + index
	}

	// Check bounds
	if index < 0 || index >= int64(len(list)) {
		return s.WriteValue(NullBulkString())
	}

	return s.WriteValue(ConvertToRESP(list[index]))
}

// HandleLSet implements the LSET command (set element at index)
func HandleLSet(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'lset' command")
	}

	key := TranslateKey(args[0].String())
	index, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	value := args[2].String()

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LSET")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteError("ERR no such key")
	}

	// Handle negative index
	if index < 0 {
		index = int64(len(list)) + index
	}

	// Check bounds
	if index < 0 || index >= int64(len(list)) {
		return s.WriteError("ERR index out of range")
	}

	// Set value
	list[index] = value

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteOK()
}

// HandleLTrim implements the LTRIM command (trim list to range)
func HandleLTrim(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'ltrim' command")
	}

	key := TranslateKey(args[0].String())
	start, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(args[2].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LTRIM")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteOK()
	}

	// Handle negative indices
	length := int64(len(list))
	if start < 0 {
		start = length + start
	}
	if stop < 0 {
		stop = length + stop
	}

	// Clamp to valid range
	if start < 0 {
		start = 0
	}
	if start >= length {
		// Trim to empty list
		cache.Delete(ctx, key)
		return s.WriteOK()
	}
	if stop >= length {
		stop = length - 1
	}
	if stop < start {
		// Trim to empty list
		cache.Delete(ctx, key)
		return s.WriteOK()
	}

	// Trim list
	list = list[start : stop+1]

	// Store trimmed list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteOK()
}

// HandleLInsert implements the LINSERT command (insert before/after element)
func HandleLInsert(s *Session, args []respProto.Value) error {
	if len(args) != 4 {
		return s.WriteError("ERR wrong number of arguments for 'linsert' command")
	}

	key := TranslateKey(args[0].String())
	where := strings.ToUpper(args[1].String())
	pivot := args[2].String()
	value := args[3].String()

	if where != "BEFORE" && where != "AFTER" {
		return s.WriteError("ERR syntax error")
	}

	cache, err := caches.FetchCache(s.SelectedCache())
	if err != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", err.Error()))
	}

	tag := s.Tag("LINSERT")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		// Key doesn't exist, return 0
		return s.WriteValue(Integer(0))
	}

	// Find pivot
	pivotIndex := -1
	for i, item := range list {
		if fmt.Sprintf("%v", item) == pivot {
			pivotIndex = i
			break
		}
	}

	if pivotIndex == -1 {
		// Pivot not found, return -1
		return s.WriteValue(Integer(-1))
	}

	// Insert value
	if where == "BEFORE" {
		list = append(list[:pivotIndex], append([]any{value}, list[pivotIndex:]...)...)
	} else { // AFTER
		list = append(list[:pivotIndex+1], append([]any{value}, list[pivotIndex+1:]...)...)
	}

	// Store updated list
	if setErr := setList(ctx, cache, key, list); setErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
	}

	return s.WriteValue(Integer(len(list)))
}

// HandleLRem implements the LREM command (remove elements by value)
func HandleLRem(s *Session, args []respProto.Value) error {
	if len(args) != 3 {
		return s.WriteError("ERR wrong number of arguments for 'lrem' command")
	}

	key := TranslateKey(args[0].String())
	count, err := strconv.ParseInt(args[1].String(), 10, 64)
	if err != nil {
		return s.WriteError("ERR value is not an integer or out of range")
	}
	element := args[2].String()

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LREM")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteValue(Integer(0))
	}

	// Remove elements
	removed := 0
	newList := make([]any, 0, len(list))

	if count == 0 {
		// Remove all occurrences
		for _, item := range list {
			if fmt.Sprintf("%v", item) != element {
				newList = append(newList, item)
			} else {
				removed++
			}
		}
	} else if count > 0 {
		// Remove first N occurrences
		for _, item := range list {
			if fmt.Sprintf("%v", item) == element && removed < int(count) {
				removed++
			} else {
				newList = append(newList, item)
			}
		}
	} else {
		// Remove last N occurrences (count is negative)
		count = -count
		// Process in reverse
		for i := len(list) - 1; i >= 0; i-- {
			item := list[i]
			if fmt.Sprintf("%v", item) == element && removed < int(count) {
				removed++
			} else {
				newList = append([]any{item}, newList...)
			}
		}
	}

	if len(newList) == 0 {
		// List is now empty, delete the key
		cache.Delete(ctx, key)
	} else {
		// Store updated list
		if setErr := setList(ctx, cache, key, newList); setErr != nil {
			return s.WriteError(fmt.Sprintf("ERR %s", setErr.Error()))
		}
	}

	return s.WriteValue(Integer(removed))
}

// HandleLPos implements the LPOS command (find position of element)
func HandleLPos(s *Session, args []respProto.Value) error {
	if len(args) < 2 {
		return s.WriteError("ERR wrong number of arguments for 'lpos' command")
	}

	key := TranslateKey(args[0].String())
	element := args[1].String()

	// Parse optional arguments (RANK, COUNT, MAXLEN)
	rank := int64(1)
	count := int64(0)
	maxlen := int64(0)

	for i := 2; i < len(args); i += 2 {
		if i+1 >= len(args) {
			return s.WriteError("ERR syntax error")
		}
		option := strings.ToUpper(args[i].String())
		value, err := strconv.ParseInt(args[i+1].String(), 10, 64)
		if err != nil {
			return s.WriteError("ERR value is not an integer or out of range")
		}

		switch option {
		case "RANK":
			rank = value
		case "COUNT":
			count = value
		case "MAXLEN":
			maxlen = value
		default:
			return s.WriteError("ERR syntax error")
		}
	}

	cache, cacheErr := caches.FetchCache(s.SelectedCache())
	if cacheErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", cacheErr.Error()))
	}

	tag := s.Tag("LPOS")
	cache.Acquire(tag)
	defer cache.Release(tag)

	ctx, cancel := context.WithTimeout(s.Context(), 5*time.Second)
	defer cancel()

	// Get existing list
	list, listErr := getList(ctx, cache, key)
	if listErr != nil {
		return s.WriteError(fmt.Sprintf("ERR %s", listErr.Error()))
	}

	if len(list) == 0 {
		return s.WriteValue(NullBulkString())
	}

	// Determine search range
	searchLen := len(list)
	if maxlen > 0 && int(maxlen) < searchLen {
		searchLen = int(maxlen)
	}

	// Find matching positions
	positions := make([]int64, 0)
	matchCount := int64(0)
	returnArray := len(args) > 2 && strings.ToUpper(args[len(args)-2].String()) == "COUNT"

	if rank >= 0 {
		// Forward search
		for i := 0; i < searchLen; i++ {
			if fmt.Sprintf("%v", list[i]) == element {
				matchCount++
				if matchCount >= rank {
					positions = append(positions, int64(i))
					if count > 0 && int64(len(positions)) >= count {
						break
					}
				}
			}
		}
	} else {
		// Reverse search
		absRank := -rank
		for i := searchLen - 1; i >= 0; i-- {
			if fmt.Sprintf("%v", list[i]) == element {
				matchCount++
				if matchCount >= absRank {
					positions = append(positions, int64(i))
					if count > 0 && int64(len(positions)) >= count {
						break
					}
				}
			}
		}
	}

	// Return result based on whether COUNT was specified
	if !returnArray {
		// Return single position
		if len(positions) == 0 {
			return s.WriteValue(NullBulkString())
		}
		return s.WriteValue(Integer(int(positions[0])))
	}

	// Return array of positions
	result := make([]respProto.Value, len(positions))
	for i, pos := range positions {
		result[i] = Integer(int(pos))
	}
	return s.WriteValue(Array(result))
}

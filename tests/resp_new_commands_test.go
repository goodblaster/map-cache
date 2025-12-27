package tests

import (
	"context"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
)

func setupRESPClient(t *testing.T) *redis.Client {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})

	ctx := context.Background()

	// Verify connection
	_, err := client.Ping(ctx).Result()
	if err != nil {
		t.Fatalf("Failed to connect to RESP server: %v", err)
	}

	return client
}

func cleanupKeys(client *redis.Client, keys ...string) {
	// Use FLUSHDB to cleanly remove all keys without warnings
	ctx := context.Background()
	client.FlushDB(ctx)
}

// Test FLUSHDB and FLUSHALL commands
func TestRESP_GenericCommands_FLUSHDB(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	// Set multiple keys
	client.Set(ctx, "key1", "value1", 0)
	client.Set(ctx, "key2", "value2", 0)
	client.HSet(ctx, "hash1", "field1", "value1")
	client.RPush(ctx, "list1", "item1", "item2")

	// Verify keys exist
	exists, _ := client.Exists(ctx, "key1", "key2", "hash1", "list1").Result()
	assert.Equal(t, int64(4), exists)

	// Flush the database
	err := client.FlushDB(ctx).Err()
	assert.NoError(t, err)

	// Verify all keys are gone
	exists, _ = client.Exists(ctx, "key1", "key2", "hash1", "list1").Result()
	assert.Equal(t, int64(0), exists)
}

func TestRESP_GenericCommands_FLUSHALL(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	// Set multiple keys
	client.Set(ctx, "key1", "value1", 0)
	client.Set(ctx, "key2", "value2", 0)

	// Verify keys exist
	exists, _ := client.Exists(ctx, "key1", "key2").Result()
	assert.Equal(t, int64(2), exists)

	// Flush all databases (in map-cache, same as FLUSHDB)
	err := client.FlushAll(ctx).Err()
	assert.NoError(t, err)

	// Verify all keys are gone
	exists, _ = client.Exists(ctx, "key1", "key2").Result()
	assert.Equal(t, int64(0), exists)
}

// Test additional list commands
func TestRESP_ListCommands_LTRIM(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Create a list
	client.RPush(ctx, "mylist", "one", "two", "three", "four", "five")

	// Trim to keep only elements 1-3
	err := client.LTrim(ctx, "mylist", 1, 3).Err()
	assert.NoError(t, err)

	// Verify result
	list, err := client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"two", "three", "four"}, list)
}

func TestRESP_ListCommands_LINSERT(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Create a list
	client.RPush(ctx, "mylist", "one", "three")

	// Insert before "three"
	length, err := client.LInsertBefore(ctx, "mylist", "three", "two").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(3), length)

	// Verify result
	list, err := client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three"}, list)

	// Insert after "three"
	length, err = client.LInsertAfter(ctx, "mylist", "three", "four").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(4), length)

	list, err = client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"one", "two", "three", "four"}, list)
}

func TestRESP_ListCommands_LREM(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Create a list with duplicates
	client.RPush(ctx, "mylist", "one", "two", "one", "three", "one")

	// Remove first 2 occurrences of "one"
	removed, err := client.LRem(ctx, "mylist", 2, "one").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(2), removed)

	// Verify result
	list, err := client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"two", "three", "one"}, list)
}

func TestRESP_ListCommands_LPOS(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Create a list with duplicates
	client.RPush(ctx, "mylist", "a", "b", "c", "b", "d", "b")

	// Find first occurrence of "b"
	pos, err := client.LPos(ctx, "mylist", "b", redis.LPosArgs{}).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), pos)

	// Find all occurrences of "b"
	positions, err := client.LPosCount(ctx, "mylist", "b", 0, redis.LPosArgs{}).Result()
	assert.NoError(t, err)
	assert.Equal(t, []int64{1, 3, 5}, positions)
}

// Test additional string commands
func TestRESP_StringCommands_GETRANGE(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set a string value
	client.Set(ctx, "mykey", "Hello, World!", 0)

	// Get substring
	substr, err := client.GetRange(ctx, "mykey", 0, 4).Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello", substr)

	// Negative indices
	substr, err = client.GetRange(ctx, "mykey", -6, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, "World!", substr)
}

func TestRESP_StringCommands_SETRANGE(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set initial value
	client.Set(ctx, "mykey", "Hello, World!", 0)

	// Replace substring
	newLen, err := client.SetRange(ctx, "mykey", 7, "Redis").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(13), newLen)

	// Verify result
	value, err := client.Get(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "Hello, Redis!", value)
}

func TestRESP_StringCommands_GETEX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set a value
	client.Set(ctx, "mykey", "value", 0)

	// Get and set expiration
	value, err := client.GetEx(ctx, "mykey", 1*time.Second).Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Verify TTL is set
	ttl, err := client.TTL(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), 0.0)
	assert.LessOrEqual(t, ttl.Seconds(), 1.0)
}

func TestRESP_StringCommands_GETDEL(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set a value
	client.Set(ctx, "mykey", "value", 0)

	// Get and delete
	value, err := client.GetDel(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Verify key is deleted
	exists, err := client.Exists(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)
}

func TestRESP_StringCommands_INCRBYFLOAT(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set initial value
	client.Set(ctx, "mykey", "10.5", 0)

	// Increment by float
	newValue, err := client.IncrByFloat(ctx, "mykey", 2.5).Result()
	assert.NoError(t, err)
	assert.Equal(t, 13.0, newValue)

	// Decrement by float
	newValue, err = client.IncrByFloat(ctx, "mykey", -3.0).Result()
	assert.NoError(t, err)
	assert.Equal(t, 10.0, newValue)
}

func TestRESP_StringCommands_PSETEX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set value with millisecond expiration
	err := client.Do(ctx, "PSETEX", "mykey", "1000", "value").Err()
	assert.NoError(t, err)

	// Verify value is set
	value, err := client.Get(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Verify TTL is set (should be around 1000ms)
	pttl, err := client.PTTL(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Greater(t, pttl.Milliseconds(), int64(0))
	assert.LessOrEqual(t, pttl.Milliseconds(), int64(1100))
}

// Test additional hash commands
func TestRESP_HashCommands_HINCRBY(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "myhash")

	// Set initial hash field
	client.HSet(ctx, "myhash", "counter", "10")

	// Increment by integer
	newValue, err := client.HIncrBy(ctx, "myhash", "counter", 5).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(15), newValue)

	// Decrement
	newValue, err = client.HIncrBy(ctx, "myhash", "counter", -3).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(12), newValue)

	// Increment non-existent field (initialize to 0)
	newValue, err = client.HIncrBy(ctx, "myhash", "new_counter", 7).Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(7), newValue)
}

func TestRESP_HashCommands_HINCRBYFLOAT(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "myhash")

	// Set initial hash field
	client.HSet(ctx, "myhash", "price", "10.5")

	// Increment by float
	newValue, err := client.HIncrByFloat(ctx, "myhash", "price", 2.25).Result()
	assert.NoError(t, err)
	assert.InDelta(t, 12.75, newValue, 0.001)

	// Decrement by float
	newValue, err = client.HIncrByFloat(ctx, "myhash", "price", -1.5).Result()
	assert.NoError(t, err)
	assert.InDelta(t, 11.25, newValue, 0.001)
}

func TestRESP_HashCommands_HSETNX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "myhash")

	// Set field if not exists (should succeed)
	created, err := client.HSetNX(ctx, "myhash", "field1", "value1").Result()
	assert.NoError(t, err)
	assert.True(t, created)

	// Try to set same field again (should fail)
	created, err = client.HSetNX(ctx, "myhash", "field1", "new_value").Result()
	assert.NoError(t, err)
	assert.False(t, created)

	// Verify original value unchanged
	value, err := client.HGet(ctx, "myhash", "field1").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value1", value)
}

// Test additional key commands
func TestRESP_KeyCommands_EXPIREAT(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set a value
	client.Set(ctx, "mykey", "value", 0)

	// Set expiration at timestamp (2 seconds from now)
	expireAt := time.Now().Add(2 * time.Second).Unix()
	success, err := client.ExpireAt(ctx, "mykey", time.Unix(expireAt, 0)).Result()
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify TTL is set
	ttl, err := client.TTL(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), 0.0)
	assert.LessOrEqual(t, ttl.Seconds(), 2.0)
}

func TestRESP_KeyCommands_PEXPIREAT(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Set a value
	client.Set(ctx, "mykey", "value", 0)

	// Set expiration at timestamp in milliseconds (2000ms from now)
	expireAtMs := time.Now().Add(2000 * time.Millisecond).UnixMilli()
	success, err := client.PExpireAt(ctx, "mykey", time.UnixMilli(expireAtMs)).Result()
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify PTTL is set (allow some slack for rounding and timing)
	pttl, err := client.PTTL(ctx, "mykey").Result()
	assert.NoError(t, err)
	assert.Greater(t, pttl.Milliseconds(), int64(0))
	assert.LessOrEqual(t, pttl.Milliseconds(), int64(2100))
}

func TestRESP_KeyCommands_RENAME(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "oldkey", "newkey")

	// Set a value
	client.Set(ctx, "oldkey", "value", 0)

	// Set TTL
	client.Expire(ctx, "oldkey", 10*time.Second)

	// Rename the key
	err := client.Rename(ctx, "oldkey", "newkey").Err()
	assert.NoError(t, err)

	// Verify old key doesn't exist
	exists, err := client.Exists(ctx, "oldkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// Verify new key exists with same value
	value, err := client.Get(ctx, "newkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Verify TTL was preserved
	ttl, err := client.TTL(ctx, "newkey").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), 0.0)
}

func TestRESP_KeyCommands_RENAMENX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "oldkey", "newkey", "anotherkey")

	// Set a value
	client.Set(ctx, "oldkey", "value", 0)

	// Rename if new key doesn't exist (should succeed)
	success, err := client.RenameNX(ctx, "oldkey", "newkey").Result()
	assert.NoError(t, err)
	assert.True(t, success)

	// Verify rename worked
	value, err := client.Get(ctx, "newkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	// Try to rename to existing key (should fail)
	client.Set(ctx, "anotherkey", "another_value", 0)
	success, err = client.RenameNX(ctx, "newkey", "anotherkey").Result()
	assert.NoError(t, err)
	assert.False(t, success)

	// Verify original keys unchanged
	value, err = client.Get(ctx, "newkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value", value)

	value, err = client.Get(ctx, "anotherkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "another_value", value)
}

func TestRESP_KeyCommands_TYPE(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "stringkey", "listkey", "hashkey", "nonexistent")

	// Test string type
	client.Set(ctx, "stringkey", "value", 0)
	keyType, err := client.Type(ctx, "stringkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "string", keyType)

	// Test list type
	client.RPush(ctx, "listkey", "item1", "item2")
	keyType, err = client.Type(ctx, "listkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "list", keyType)

	// Test hash type
	client.HSet(ctx, "hashkey", "field", "value")
	keyType, err = client.Type(ctx, "hashkey").Result()
	assert.NoError(t, err)
	assert.Equal(t, "hash", keyType)

	// Test non-existent key
	keyType, err = client.Type(ctx, "nonexistent").Result()
	assert.NoError(t, err)
	assert.Equal(t, "none", keyType)
}

func TestRESP_KeyCommands_EXPIRETIME(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Test key that doesn't exist
	expireTime, err := client.Do(ctx, "EXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), expireTime)

	// Set a value without expiration
	client.Set(ctx, "mykey", "value", 0)
	expireTime, err = client.Do(ctx, "EXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), expireTime)

	// Set expiration (10 seconds from now)
	client.Expire(ctx, "mykey", 10*time.Second)
	expireTime, err = client.Do(ctx, "EXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)

	// Should be close to 10 seconds from now
	expectedExpireTime := time.Now().Add(10 * time.Second).Unix()
	assert.InDelta(t, expectedExpireTime, expireTime, 2.0)
}

func TestRESP_KeyCommands_PEXPIRETIME(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mykey")

	// Test key that doesn't exist
	pexpireTime, err := client.Do(ctx, "PEXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(-2), pexpireTime)

	// Set a value without expiration
	client.Set(ctx, "mykey", "value", 0)
	pexpireTime, err = client.Do(ctx, "PEXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(-1), pexpireTime)

	// Set expiration (2000ms from now)
	client.PExpire(ctx, "mykey", 2000*time.Millisecond)
	pexpireTime, err = client.Do(ctx, "PEXPIRETIME", "mykey").Int64()
	assert.NoError(t, err)

	// Should be close to 2000ms from now
	expectedPExpireTime := time.Now().Add(2000 * time.Millisecond).UnixMilli()
	assert.InDelta(t, expectedPExpireTime, pexpireTime, 1000.0)
}

func TestRESP_KeyCommands_COPY(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "source", "dest")

	// Test copying non-existent key
	copied, err := client.Do(ctx, "COPY", "source", "dest").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), copied)

	// Set source value with TTL
	client.Set(ctx, "source", "value123", 0)
	client.Expire(ctx, "source", 10*time.Second)

	// Copy to non-existent destination
	copied, err = client.Do(ctx, "COPY", "source", "dest").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), copied)

	// Verify destination has same value
	destValue, err := client.Get(ctx, "dest").Result()
	assert.NoError(t, err)
	assert.Equal(t, "value123", destValue)

	// Verify TTL was preserved
	ttl, err := client.TTL(ctx, "dest").Result()
	assert.NoError(t, err)
	assert.Greater(t, ttl.Seconds(), 0.0)
	assert.LessOrEqual(t, ttl.Seconds(), 10.0)

	// Try to copy to existing destination without REPLACE (should fail)
	copied, err = client.Do(ctx, "COPY", "source", "dest").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), copied)

	// Copy with REPLACE flag (should succeed)
	client.Set(ctx, "source", "newvalue", 0)
	copied, err = client.Do(ctx, "COPY", "source", "dest", "REPLACE").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(1), copied)

	// Verify destination has new value
	destValue, err = client.Get(ctx, "dest").Result()
	assert.NoError(t, err)
	assert.Equal(t, "newvalue", destValue)
}

func TestRESP_ListCommands_LPUSHX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Try to push to non-existent list (should return 0)
	length, err := client.Do(ctx, "LPUSHX", "mylist", "value1").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), length)

	// Verify list wasn't created
	exists, err := client.Exists(ctx, "mylist").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// Create list first
	client.RPush(ctx, "mylist", "a", "b")

	// Now LPUSHX should work
	length, err = client.Do(ctx, "LPUSHX", "mylist", "x", "y").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(4), length)

	// Verify elements were added to the left
	list, err := client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"x", "y", "a", "b"}, list)
}

func TestRESP_ListCommands_RPUSHX(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "mylist")

	// Try to push to non-existent list (should return 0)
	length, err := client.Do(ctx, "RPUSHX", "mylist", "value1").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), length)

	// Verify list wasn't created
	exists, err := client.Exists(ctx, "mylist").Result()
	assert.NoError(t, err)
	assert.Equal(t, int64(0), exists)

	// Create list first
	client.RPush(ctx, "mylist", "a", "b")

	// Now RPUSHX should work
	length, err = client.Do(ctx, "RPUSHX", "mylist", "x", "y").Int64()
	assert.NoError(t, err)
	assert.Equal(t, int64(4), length)

	// Verify elements were added to the right
	list, err := client.LRange(ctx, "mylist", 0, -1).Result()
	assert.NoError(t, err)
	assert.Equal(t, []string{"a", "b", "x", "y"}, list)
}

func TestRESP_HashCommands_HRANDFIELD(t *testing.T) {
	client := setupRESPClient(t)
	defer client.Close()
	ctx := context.Background()

	cleanupKeys(client, "myhash")

	// Test on non-existent hash (should return null/redis.Nil)
	_, err := client.Do(ctx, "HRANDFIELD", "myhash").Result()
	assert.Error(t, err)
	assert.Equal(t, "redis: nil", err.Error())

	// Create hash
	client.HSet(ctx, "myhash", "field1", "value1", "field2", "value2", "field3", "value3")

	// Get single random field
	field, err := client.Do(ctx, "HRANDFIELD", "myhash").Result()
	assert.NoError(t, err)
	assert.Contains(t, []string{"field1", "field2", "field3"}, field)

	// Get multiple random fields
	fields, err := client.Do(ctx, "HRANDFIELD", "myhash", "2").Slice()
	assert.NoError(t, err)
	assert.Len(t, fields, 2)

	// Get fields with values
	fieldsWithValues, err := client.Do(ctx, "HRANDFIELD", "myhash", "2", "WITHVALUES").Slice()
	assert.NoError(t, err)
	assert.Len(t, fieldsWithValues, 4) // 2 fields * 2 (field + value each)

	// Test with count larger than hash size
	allFields, err := client.Do(ctx, "HRANDFIELD", "myhash", "10").Slice()
	assert.NoError(t, err)
	assert.Len(t, allFields, 3) // Only 3 fields exist

	// Test on empty hash
	client.Del(ctx, "myhash")
	client.HSet(ctx, "emptyhash", "nothing", "")
	client.HDel(ctx, "emptyhash", "nothing")
	_, err = client.Do(ctx, "HRANDFIELD", "emptyhash").Result()
	assert.Error(t, err)
	assert.Equal(t, "redis: nil", err.Error())
}

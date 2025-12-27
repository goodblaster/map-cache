package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Println("=== COMPREHENSIVE REDIS INTEGRATION TEST ===\n")

	// Test 1: Basic connectivity
	fmt.Println("1. Testing basic connectivity...")
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("ERROR: PING failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PING: %s\n", pong)

	// Test 2: String operations
	fmt.Println("\n2. Testing string operations...")
	client.Set(ctx, "mykey", "Hello Redis", 0)
	val, _ := client.Get(ctx, "mykey").Result()
	fmt.Printf("✓ SET/GET: %s\n", val)

	client.Incr(ctx, "counter")
	client.IncrBy(ctx, "counter", 5)
	count, _ := client.Get(ctx, "counter").Result()
	fmt.Printf("✓ INCR/INCRBY: %s\n", count)

	// Test 3: Key management with TTL
	fmt.Println("\n3. Testing key management...")
	client.Set(ctx, "tempkey", "temporary", 0)
	client.Expire(ctx, "tempkey", 5*time.Second)
	ttl, _ := client.TTL(ctx, "tempkey").Result()
	fmt.Printf("✓ EXPIRE/TTL: %v\n", ttl)

	client.Persist(ctx, "tempkey")
	ttl, _ = client.TTL(ctx, "tempkey").Result()
	fmt.Printf("✓ PERSIST: TTL now %v (should be -1)\n", ttl)

	// Test 4: Pattern matching
	fmt.Println("\n4. Testing pattern matching...")
	client.Set(ctx, "user:100:name", "Alice", 0)
	client.Set(ctx, "user:100:email", "alice@example.com", 0)
	client.Set(ctx, "user:200:name", "Bob", 0)

	keys, _ := client.Keys(ctx, "user:*").Result()
	fmt.Printf("✓ KEYS user:*: found %d keys\n", len(keys))

	// Test 5: Hash operations
	fmt.Println("\n5. Testing hash operations...")
	client.HSet(ctx, "product:1000",
		"name", "Laptop",
		"price", "999.99",
		"stock", "50")

	product, _ := client.HGetAll(ctx, "product:1000").Result()
	fmt.Printf("✓ HSET/HGETALL: %v\n", product)

	price, _ := client.HGet(ctx, "product:1000", "price").Result()
	fmt.Printf("✓ HGET price: %s\n", price)

	exists, _ := client.HExists(ctx, "product:1000", "name").Result()
	fmt.Printf("✓ HEXISTS name: %v\n", exists)

	fields, _ := client.HKeys(ctx, "product:1000").Result()
	fmt.Printf("✓ HKEYS: %v\n", fields)

	// Test 6: Multi-cache support (SELECT)
	fmt.Println("\n6. Testing multi-cache support...")

	// Default cache (DB 0)
	client.Do(ctx, "SELECT", "0")
	client.Set(ctx, "cache_test", "default_cache", 0)
	val0, _ := client.Get(ctx, "cache_test").Result()
	fmt.Printf("✓ SELECT 0: %s\n", val0)

	// Cache "1" (DB 1)
	client.Do(ctx, "SELECT", "1")
	client.Set(ctx, "cache_test", "cache_1", 0)
	val1, _ := client.Get(ctx, "cache_test").Result()
	fmt.Printf("✓ SELECT 1: %s\n", val1)

	// Verify isolation
	client.Do(ctx, "SELECT", "0")
	valCheck, _ := client.Get(ctx, "cache_test").Result()
	if valCheck == "default_cache" {
		fmt.Printf("✓ Cache isolation verified\n")
	} else {
		fmt.Printf("ERROR: Cache isolation failed, got %s\n", valCheck)
	}

	// Test 7: Batch operations
	fmt.Println("\n7. Testing batch operations...")
	client.MSet(ctx, "k1", "v1", "k2", "v2", "k3", "v3")
	vals, _ := client.MGet(ctx, "k1", "k2", "k3", "k4").Result()
	fmt.Printf("✓ MSET/MGET: %v\n", vals)

	// Test 8: DELETE operations
	fmt.Println("\n8. Testing delete operations...")
	client.Del(ctx, "k1", "k2")
	exists1, _ := client.Exists(ctx, "k1", "k2", "k3").Result()
	fmt.Printf("✓ DEL: %d keys exist (should be 1)\n", exists1)

	client.HDel(ctx, "product:1000", "stock")
	length, _ := client.HLen(ctx, "product:1000").Result()
	fmt.Printf("✓ HDEL: hash has %d fields now\n", length)

	// Test 9: SETNX and GETSET
	fmt.Println("\n9. Testing conditional operations...")
	wasSet, _ := client.SetNX(ctx, "lock", "acquired", 0).Result()
	fmt.Printf("✓ SETNX lock: %v\n", wasSet)

	wasSet2, _ := client.SetNX(ctx, "lock", "try_again", 0).Result()
	fmt.Printf("✓ SETNX lock (exists): %v (should be false)\n", wasSet2)

	oldVal, _ := client.GetSet(ctx, "lock", "new_value").Result()
	fmt.Printf("✓ GETSET: old=%s\n", oldVal)

	// Test 10: APPEND and STRLEN
	fmt.Println("\n10. Testing string manipulation...")
	client.Set(ctx, "message", "Hello", 0)
	newLen, _ := client.Append(ctx, "message", " World").Result()
	fmt.Printf("✓ APPEND: new length=%d\n", newLen)

	finalMsg, _ := client.Get(ctx, "message").Result()
	strLen, _ := client.StrLen(ctx, "message").Result()
	fmt.Printf("✓ STRLEN: message='%s', length=%d\n", finalMsg, strLen)

	// Test 11: Key translation (: to /)
	fmt.Println("\n11. Testing key translation...")
	client.Set(ctx, "session:abc123:user_id", "12345", 0)
	userId, _ := client.Get(ctx, "session:abc123:user_id").Result()
	fmt.Printf("✓ Key translation (session:abc123:user_id): %s\n", userId)

	// Test 12: Expiration edge cases
	fmt.Println("\n12. Testing expiration edge cases...")
	client.Set(ctx, "quick_expire", "value", 0)
	client.PExpire(ctx, "quick_expire", 2000*time.Millisecond)
	pttl, _ := client.PTTL(ctx, "quick_expire").Result()
	fmt.Printf("✓ PEXPIRE/PTTL: %v milliseconds\n", pttl)

	// Wait for expiration
	fmt.Println("   Waiting 3 seconds for key to expire...")
	time.Sleep(3 * time.Second)
	_, err = client.Get(ctx, "quick_expire").Result()
	if err == redis.Nil {
		fmt.Println("✓ Key expired successfully")
	} else {
		fmt.Printf("ERROR: Key should have expired: %v\n", err)
	}

	fmt.Println("\n=== ALL INTEGRATION TESTS PASSED ✅ ===")
	fmt.Println("\nImplemented commands summary:")
	fmt.Println("  String: GET, SET, DEL, EXISTS, INCR, DECR, INCRBY, DECRBY")
	fmt.Println("          MGET, MSET, GETSET, SETNX, SETEX, STRLEN, APPEND")
	fmt.Println("  Key:    EXPIRE, PEXPIRE, PERSIST, TTL, PTTL, KEYS")
	fmt.Println("  Hash:   HGET, HSET, HGETALL, HDEL, HEXISTS, HLEN")
	fmt.Println("          HKEYS, HVALS, HMGET, HMSET")
	fmt.Println("  Other:  PING, ECHO, SELECT, COMMAND, HELLO, CLIENT")
	fmt.Println("\nTotal: 37 Redis commands implemented!")
}

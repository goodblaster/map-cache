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

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Create some test keys
	fmt.Println("Setting up test keys...")
	client.Set(ctx, "key1", "value1", 0)
	client.Set(ctx, "key2", "value2", 0)
	client.Set(ctx, "user:123:name", "Alice", 0)
	client.Set(ctx, "user:456:name", "Bob", 0)
	fmt.Println("✓ Test keys created")

	// Test EXPIRE
	fmt.Println("\nTesting EXPIRE...")
	set, err := client.Expire(ctx, "key1", 10*time.Second).Result()
	if err != nil {
		fmt.Printf("ERROR: EXPIRE failed: %v\n", err)
		return
	}
	fmt.Printf("✓ EXPIRE key1 10s: %v\n", set)

	// Try to expire non-existent key
	set, err = client.Expire(ctx, "nonexistent", 10*time.Second).Result()
	if err != nil {
		fmt.Printf("ERROR: EXPIRE nonexistent failed: %v\n", err)
		return
	}
	fmt.Printf("✓ EXPIRE nonexistent: %v (should be false)\n", set)

	// Test TTL
	fmt.Println("\nTesting TTL...")
	ttl, err := client.TTL(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: TTL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ TTL key1: %v (should be ~10s)\n", ttl)

	// TTL on key without expiration
	ttl, err = client.TTL(ctx, "key2").Result()
	if err != nil {
		fmt.Printf("ERROR: TTL key2 failed: %v\n", err)
		return
	}
	fmt.Printf("✓ TTL key2 (no expiration): %v (should be -1)\n", ttl)

	// TTL on non-existent key
	ttl, err = client.TTL(ctx, "nonexistent").Result()
	if err != nil {
		fmt.Printf("ERROR: TTL nonexistent failed: %v\n", err)
		return
	}
	fmt.Printf("✓ TTL nonexistent: %v (should be -2)\n", ttl)

	// Test PEXPIRE and PTTL
	fmt.Println("\nTesting PEXPIRE and PTTL...")
	set, err = client.PExpire(ctx, "key2", 5000*time.Millisecond).Result()
	if err != nil {
		fmt.Printf("ERROR: PEXPIRE failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PEXPIRE key2 5000ms: %v\n", set)

	pttl, err := client.PTTL(ctx, "key2").Result()
	if err != nil {
		fmt.Printf("ERROR: PTTL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PTTL key2: %v (should be ~5000ms)\n", pttl)

	// Test PERSIST
	fmt.Println("\nTesting PERSIST...")
	persisted, err := client.Persist(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: PERSIST failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PERSIST key1: %v (should be true)\n", persisted)

	// Check TTL after PERSIST
	ttl, err = client.TTL(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: TTL after PERSIST failed: %v\n", err)
		return
	}
	fmt.Printf("✓ TTL key1 after PERSIST: %v (should be -1)\n", ttl)

	// PERSIST on key without TTL
	persisted, err = client.Persist(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: PERSIST (no TTL) failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PERSIST key1 (no TTL): %v (should be false)\n", persisted)

	// Test KEYS with wildcard pattern
	fmt.Println("\nTesting KEYS...")
	keys, err := client.Keys(ctx, "user:*:name").Result()
	if err != nil {
		fmt.Printf("ERROR: KEYS failed: %v\n", err)
		return
	}
	fmt.Printf("✓ KEYS user:*:name: %v (should match user:123:name and user:456:name)\n", keys)

	// KEYS with exact match
	keys, err = client.Keys(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: KEYS exact failed: %v\n", err)
		return
	}
	fmt.Printf("✓ KEYS key1: %v\n", keys)

	// KEYS with * (all keys)
	keys, err = client.Keys(ctx, "*").Result()
	if err != nil {
		fmt.Printf("ERROR: KEYS * failed: %v\n", err)
		return
	}
	fmt.Printf("✓ KEYS * (all keys): %v\n", keys)

	fmt.Println("\n✅ All key management command tests passed!")
}

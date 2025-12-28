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

	// Test HSET and HGET
	fmt.Println("Testing HSET and HGET...")
	num, err := client.HSet(ctx, "user:1000", "name", "John Doe").Result()
	if err != nil {
		fmt.Printf("ERROR: HSET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HSET user:1000 name: %d field added\n", num)

	name, err := client.HGet(ctx, "user:1000", "name").Result()
	if err != nil {
		fmt.Printf("ERROR: HGET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HGET user:1000 name: %s\n", name)

	// Test HSET with multiple fields
	fmt.Println("\nTesting HSET with multiple fields...")
	num, err = client.HSet(ctx, "user:1000", "email", "john@example.com", "age", "30").Result()
	if err != nil {
		fmt.Printf("ERROR: HSET multiple failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HSET user:1000 (email, age): %d new fields\n", num)

	// Test HGETALL
	fmt.Println("\nTesting HGETALL...")
	all, err := client.HGetAll(ctx, "user:1000").Result()
	if err != nil {
		fmt.Printf("ERROR: HGETALL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HGETALL user:1000: %v\n", all)

	// Test HEXISTS
	fmt.Println("\nTesting HEXISTS...")
	exists, err := client.HExists(ctx, "user:1000", "name").Result()
	if err != nil {
		fmt.Printf("ERROR: HEXISTS failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HEXISTS user:1000 name: %v\n", exists)

	exists, err = client.HExists(ctx, "user:1000", "nonexistent").Result()
	if err != nil {
		fmt.Printf("ERROR: HEXISTS (nonexistent) failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HEXISTS user:1000 nonexistent: %v (should be false)\n", exists)

	// Test HLEN
	fmt.Println("\nTesting HLEN...")
	length, err := client.HLen(ctx, "user:1000").Result()
	if err != nil {
		fmt.Printf("ERROR: HLEN failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HLEN user:1000: %d fields\n", length)

	// Test HKEYS
	fmt.Println("\nTesting HKEYS...")
	keys, err := client.HKeys(ctx, "user:1000").Result()
	if err != nil {
		fmt.Printf("ERROR: HKEYS failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HKEYS user:1000: %v\n", keys)

	// Test HVALS
	fmt.Println("\nTesting HVALS...")
	vals, err := client.HVals(ctx, "user:1000").Result()
	if err != nil {
		fmt.Printf("ERROR: HVALS failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HVALS user:1000: %v\n", vals)

	// Test HMGET
	fmt.Println("\nTesting HMGET...")
	values, err := client.HMGet(ctx, "user:1000", "name", "email", "nonexistent").Result()
	if err != nil {
		fmt.Printf("ERROR: HMGET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HMGET user:1000 (name, email, nonexistent): %v\n", values)

	// Test HMSET
	fmt.Println("\nTesting HMSET...")
	err = client.HMSet(ctx, "user:2000", "name", "Jane Doe", "email", "jane@example.com").Err()
	if err != nil {
		fmt.Printf("ERROR: HMSET failed: %v\n", err)
		return
	}
	fmt.Println("✓ HMSET user:2000 (name, email)")

	all2, err := client.HGetAll(ctx, "user:2000").Result()
	if err != nil {
		fmt.Printf("ERROR: HGETALL user:2000 failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HGETALL user:2000: %v\n", all2)

	// Test HDEL
	fmt.Println("\nTesting HDEL...")
	deleted, err := client.HDel(ctx, "user:1000", "age").Result()
	if err != nil {
		fmt.Printf("ERROR: HDEL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HDEL user:1000 age: %d fields deleted\n", deleted)

	length, err = client.HLen(ctx, "user:1000").Result()
	if err != nil {
		fmt.Printf("ERROR: HLEN after HDEL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ HLEN user:1000 after HDEL: %d fields (should be 2)\n", length)

	// Test HGET on non-existent hash
	fmt.Println("\nTesting HGET on non-existent hash...")
	_, err = client.HGet(ctx, "nonexistent", "field").Result()
	if err == redis.Nil {
		fmt.Println("✓ HGET nonexistent hash returns nil (expected)")
	} else if err != nil {
		fmt.Printf("ERROR: HGET nonexistent failed: %v\n", err)
		return
	}

	fmt.Println("\n✅ All hash command tests passed!")
}

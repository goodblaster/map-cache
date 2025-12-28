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

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Test SET and GET
	fmt.Println("Testing SET and GET...")
	err := client.Set(ctx, "key1", "value1", 0).Err()
	if err != nil {
		fmt.Printf("ERROR: SET failed: %v\n", err)
		return
	}
	val, err := client.Get(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: GET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ SET/GET: %s\n", val)

	// Test SET with translation (: → /)
	fmt.Println("\nTesting key translation (user:123:name)...")
	err = client.Set(ctx, "user:123:name", "Alice", 0).Err()
	if err != nil {
		fmt.Printf("ERROR: SET user:123:name failed: %v\n", err)
		return
	}
	val, err = client.Get(ctx, "user:123:name").Result()
	if err != nil {
		fmt.Printf("ERROR: GET user:123:name failed: %v\n", err)
		return
	}
	fmt.Printf("✓ Key translation: %s\n", val)

	// Test INCR
	fmt.Println("\nTesting INCR...")
	newVal, err := client.Incr(ctx, "counter").Result()
	if err != nil {
		fmt.Printf("ERROR: INCR failed: %v\n", err)
		return
	}
	fmt.Printf("✓ INCR counter: %d\n", newVal)

	// Test INCRBY
	fmt.Println("\nTesting INCRBY...")
	newVal, err = client.IncrBy(ctx, "counter", 5).Result()
	if err != nil {
		fmt.Printf("ERROR: INCRBY failed: %v\n", err)
		return
	}
	fmt.Printf("✓ INCRBY counter 5: %d\n", newVal)

	// Test DECR
	fmt.Println("\nTesting DECR...")
	newVal, err = client.Decr(ctx, "counter").Result()
	if err != nil {
		fmt.Printf("ERROR: DECR failed: %v\n", err)
		return
	}
	fmt.Printf("✓ DECR counter: %d\n", newVal)

	// Test EXISTS
	fmt.Println("\nTesting EXISTS...")
	exists, err := client.Exists(ctx, "key1", "key2", "counter").Result()
	if err != nil {
		fmt.Printf("ERROR: EXISTS failed: %v\n", err)
		return
	}
	fmt.Printf("✓ EXISTS (key1, key2, counter): %d keys exist\n", exists)

	// Test DEL
	fmt.Println("\nTesting DEL...")
	deleted, err := client.Del(ctx, "key1").Result()
	if err != nil {
		fmt.Printf("ERROR: DEL failed: %v\n", err)
		return
	}
	fmt.Printf("✓ DEL key1: %d keys deleted\n", deleted)

	// Test MSET and MGET
	fmt.Println("\nTesting MSET and MGET...")
	err = client.MSet(ctx, "key2", "value2", "key3", "value3").Err()
	if err != nil {
		fmt.Printf("ERROR: MSET failed: %v\n", err)
		return
	}
	vals, err := client.MGet(ctx, "key2", "key3", "nonexistent").Result()
	if err != nil {
		fmt.Printf("ERROR: MGET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ MGET (key2, key3, nonexistent): %v\n", vals)

	// Test GETSET
	fmt.Println("\nTesting GETSET...")
	oldVal, err := client.GetSet(ctx, "key2", "newvalue2").Result()
	if err != nil {
		fmt.Printf("ERROR: GETSET failed: %v\n", err)
		return
	}
	fmt.Printf("✓ GETSET key2: old=%s\n", oldVal)

	// Test SETNX
	fmt.Println("\nTesting SETNX...")
	set, err := client.SetNX(ctx, "key4", "value4", 0).Result()
	if err != nil {
		fmt.Printf("ERROR: SETNX failed: %v\n", err)
		return
	}
	fmt.Printf("✓ SETNX key4: %v\n", set)

	set, err = client.SetNX(ctx, "key4", "value4b", 0).Result()
	if err != nil {
		fmt.Printf("ERROR: SETNX (existing) failed: %v\n", err)
		return
	}
	fmt.Printf("✓ SETNX key4 (exists): %v (should be false)\n", set)

	// Test SETEX
	fmt.Println("\nTesting SETEX...")
	err = client.SetEx(ctx, "tempkey", "tempvalue", 60*time.Second).Err()
	if err != nil {
		fmt.Printf("ERROR: SETEX failed: %v\n", err)
		return
	}
	fmt.Printf("✓ SETEX tempkey (60s TTL)\n")

	// Test STRLEN
	fmt.Println("\nTesting STRLEN...")
	length, err := client.StrLen(ctx, "key3").Result()
	if err != nil {
		fmt.Printf("ERROR: STRLEN failed: %v\n", err)
		return
	}
	fmt.Printf("✓ STRLEN key3: %d\n", length)

	// Test APPEND
	fmt.Println("\nTesting APPEND...")
	newLen, err := client.Append(ctx, "key3", "_appended").Result()
	if err != nil {
		fmt.Printf("ERROR: APPEND failed: %v\n", err)
		return
	}
	fmt.Printf("✓ APPEND key3: new length=%d\n", newLen)

	val, _ = client.Get(ctx, "key3").Result()
	fmt.Printf("  Final value: %s\n", val)

	fmt.Println("\n✅ All string command tests passed!")
}

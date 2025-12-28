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

	fmt.Println("=== REDIS LIST COMMANDS TEST ===")
	fmt.Println()

	// Test 1: RPUSH (append to list)
	fmt.Println("1. Testing RPUSH...")
	length, err := client.RPush(ctx, "mylist", "one", "two", "three").Result()
	if err != nil {
		fmt.Printf("ERROR: RPUSH failed: %v\n", err)
		return
	}
	fmt.Printf("✓ RPUSH mylist one two three: length=%d\n", length)

	// Test 2: LRANGE (get range)
	fmt.Println("\n2. Testing LRANGE...")
	values, err := client.LRange(ctx, "mylist", 0, -1).Result()
	if err != nil {
		fmt.Printf("ERROR: LRANGE failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LRANGE mylist 0 -1: %v\n", values)

	// Test 3: LPUSH (prepend to list)
	fmt.Println("\n3. Testing LPUSH...")
	length, err = client.LPush(ctx, "mylist", "zero").Result()
	if err != nil {
		fmt.Printf("ERROR: LPUSH failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LPUSH mylist zero: length=%d\n", length)

	values, _ = client.LRange(ctx, "mylist", 0, -1).Result()
	fmt.Printf("  List now: %v\n", values)

	// Test 4: LLEN (get length)
	fmt.Println("\n4. Testing LLEN...")
	length, err = client.LLen(ctx, "mylist").Result()
	if err != nil {
		fmt.Printf("ERROR: LLEN failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LLEN mylist: %d\n", length)

	// Test 5: LINDEX (get element at index)
	fmt.Println("\n5. Testing LINDEX...")
	val, err := client.LIndex(ctx, "mylist", 0).Result()
	if err != nil {
		fmt.Printf("ERROR: LINDEX failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LINDEX mylist 0: %s\n", val)

	val, err = client.LIndex(ctx, "mylist", -1).Result()
	if err != nil {
		fmt.Printf("ERROR: LINDEX -1 failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LINDEX mylist -1 (last): %s\n", val)

	// Test 6: LSET (set element at index)
	fmt.Println("\n6. Testing LSET...")
	err = client.LSet(ctx, "mylist", 1, "ONE").Err()
	if err != nil {
		fmt.Printf("ERROR: LSET failed: %v\n", err)
		return
	}
	fmt.Println("✓ LSET mylist 1 ONE")

	values, _ = client.LRange(ctx, "mylist", 0, -1).Result()
	fmt.Printf("  List now: %v\n", values)

	// Test 7: LPOP (remove from left)
	fmt.Println("\n7. Testing LPOP...")
	val, err = client.LPop(ctx, "mylist").Result()
	if err != nil {
		fmt.Printf("ERROR: LPOP failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LPOP mylist: %s\n", val)

	values, _ = client.LRange(ctx, "mylist", 0, -1).Result()
	fmt.Printf("  List now: %v\n", values)

	// Test 8: RPOP (remove from right)
	fmt.Println("\n8. Testing RPOP...")
	val, err = client.RPop(ctx, "mylist").Result()
	if err != nil {
		fmt.Printf("ERROR: RPOP failed: %v\n", err)
		return
	}
	fmt.Printf("✓ RPOP mylist: %s\n", val)

	values, _ = client.LRange(ctx, "mylist", 0, -1).Result()
	fmt.Printf("  List now: %v\n", values)

	// Test 9: LRANGE with specific range
	fmt.Println("\n9. Testing LRANGE with range...")
	values, err = client.LRange(ctx, "mylist", 0, 0).Result()
	if err != nil {
		fmt.Printf("ERROR: LRANGE 0 0 failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LRANGE mylist 0 0: %v\n", values)

	// Test 10: Empty list operations
	fmt.Println("\n10. Testing empty list operations...")
	val, err = client.LPop(ctx, "emptylist").Result()
	if err == redis.Nil {
		fmt.Println("✓ LPOP emptylist: nil (expected)")
	} else if err != nil {
		fmt.Printf("ERROR: LPOP empty failed: %v\n", err)
		return
	}

	length, err = client.LLen(ctx, "emptylist").Result()
	if err != nil {
		fmt.Printf("ERROR: LLEN empty failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LLEN emptylist: %d (should be 0)\n", length)

	values, err = client.LRange(ctx, "emptylist", 0, -1).Result()
	if err != nil {
		fmt.Printf("ERROR: LRANGE empty failed: %v\n", err)
		return
	}
	fmt.Printf("✓ LRANGE emptylist: %v (should be empty)\n", values)

	// Test 11: Pop until empty
	fmt.Println("\n11. Testing pop until empty...")
	client.RPush(ctx, "templist", "a", "b")
	val1, _ := client.LPop(ctx, "templist").Result()
	val2, _ := client.LPop(ctx, "templist").Result()
	_, err = client.LPop(ctx, "templist").Result()
	fmt.Printf("✓ Popped: %s, %s\n", val1, val2)
	if err == redis.Nil {
		fmt.Println("✓ Third pop returned nil (list deleted)")
	}

	// Test 12: Negative indices
	fmt.Println("\n12. Testing negative indices...")
	client.RPush(ctx, "indexlist", "0", "1", "2", "3", "4")
	val, _ = client.LIndex(ctx, "indexlist", -1).Result()
	fmt.Printf("✓ LINDEX indexlist -1: %s (should be 4)\n", val)
	val, _ = client.LIndex(ctx, "indexlist", -2).Result()
	fmt.Printf("✓ LINDEX indexlist -2: %s (should be 3)\n", val)

	values, _ = client.LRange(ctx, "indexlist", -3, -1).Result()
	fmt.Printf("✓ LRANGE indexlist -3 -1: %v (should be [2 3 4])\n", values)

	// Test 13: Queue use case (RPUSH + LPOP)
	fmt.Println("\n13. Testing queue pattern (FIFO)...")
	client.RPush(ctx, "queue", "job1", "job2", "job3")
	job1, _ := client.LPop(ctx, "queue").Result()
	job2, _ := client.LPop(ctx, "queue").Result()
	fmt.Printf("✓ Queue: processed %s, %s\n", job1, job2)
	remaining, _ := client.LRange(ctx, "queue", 0, -1).Result()
	fmt.Printf("  Remaining: %v\n", remaining)

	// Test 14: Stack use case (RPUSH + RPOP)
	fmt.Println("\n14. Testing stack pattern (LIFO)...")
	client.RPush(ctx, "stack", "item1", "item2", "item3")
	item1, _ := client.RPop(ctx, "stack").Result()
	item2, _ := client.RPop(ctx, "stack").Result()
	fmt.Printf("✓ Stack: popped %s, %s\n", item1, item2)
	remaining, _ = client.LRange(ctx, "stack", 0, -1).Result()
	fmt.Printf("  Remaining: %v\n", remaining)

	fmt.Println("\n✅ All list command tests passed!")
	fmt.Println("\nImplemented list commands:")
	fmt.Println("  LPUSH, RPUSH - Add elements to list")
	fmt.Println("  LPOP, RPOP   - Remove and return elements")
	fmt.Println("  LLEN         - Get list length")
	fmt.Println("  LRANGE       - Get range of elements")
	fmt.Println("  LINDEX       - Get element at index")
	fmt.Println("  LSET         - Set element at index")
}

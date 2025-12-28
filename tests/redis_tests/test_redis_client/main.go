package main

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// Test PING
	fmt.Println("Testing PING...")
	pong, err := client.Ping(ctx).Result()
	if err != nil {
		fmt.Printf("ERROR: PING failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PING: %s\n", pong)

	// Test PING with message
	fmt.Println("\nTesting PING with message...")
	pong, err = client.Do(ctx, "PING", "hello").Text()
	if err != nil {
		fmt.Printf("ERROR: PING hello failed: %v\n", err)
		return
	}
	fmt.Printf("✓ PING hello: %s\n", pong)

	// Test ECHO
	fmt.Println("\nTesting ECHO...")
	echo, err := client.Echo(ctx, "Hello, Redis!").Result()
	if err != nil {
		fmt.Printf("ERROR: ECHO failed: %v\n", err)
		return
	}
	fmt.Printf("✓ ECHO: %s\n", echo)

	// Test SELECT
	fmt.Println("\nTesting SELECT...")
	err = client.Do(ctx, "SELECT", "0").Err()
	if err != nil {
		fmt.Printf("ERROR: SELECT failed: %v\n", err)
		return
	}
	fmt.Printf("✓ SELECT 0: OK\n")

	fmt.Println("\nAll tests passed! ✓")
}

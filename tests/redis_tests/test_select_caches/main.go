package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/redis/go-redis/v9"
)

func main() {
	// Wait for server to be ready
	time.Sleep(1 * time.Second)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// Create numeric caches via HTTP API
	fmt.Println("Creating numeric caches via HTTP API...")

	// Create cache "1"
	req, _ := http.NewRequest("POST", "http://localhost:8080/api/v1/caches", bytes.NewBufferString(`{"name":"1"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR creating cache 1: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✓ Created cache '1' via HTTP (status: %d)\n", resp.StatusCode)

	// Create cache "2"
	req, _ = http.NewRequest("POST", "http://localhost:8080/api/v1/caches", bytes.NewBufferString(`{"name":"2"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		fmt.Printf("ERROR creating cache 2: %v\n", err)
		return
	}
	resp.Body.Close()
	fmt.Printf("✓ Created cache '2' via HTTP (status: %d)\n", resp.StatusCode)

	// Connect via Redis
	fmt.Println("\nTesting Redis SELECT command...")
	client := redis.NewClient(&redis.Options{
		Addr: "localhost:6379",
	})
	defer client.Close()

	// Test SELECT 0 (should use "default" cache)
	fmt.Println("\nTesting SELECT 0 (default cache)...")
	err = client.Do(ctx, "SELECT", "0").Err()
	if err != nil {
		fmt.Printf("ERROR: SELECT 0 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SELECT 0")

	err = client.Set(ctx, "key0", "value_in_default", 0).Err()
	if err != nil {
		fmt.Printf("ERROR: SET in db 0 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SET key0 in default cache")

	// Test SELECT 1 (should use cache "1")
	fmt.Println("\nTesting SELECT 1 (cache '1')...")
	err = client.Do(ctx, "SELECT", "1").Err()
	if err != nil {
		fmt.Printf("ERROR: SELECT 1 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SELECT 1")

	err = client.Set(ctx, "key1", "value_in_cache_1", 0).Err()
	if err != nil {
		fmt.Printf("ERROR: SET in db 1 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SET key1 in cache '1'")

	// Test SELECT 2 (should use cache "2")
	fmt.Println("\nTesting SELECT 2 (cache '2')...")
	err = client.Do(ctx, "SELECT", "2").Err()
	if err != nil {
		fmt.Printf("ERROR: SELECT 2 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SELECT 2")

	err = client.Set(ctx, "key2", "value_in_cache_2", 0).Err()
	if err != nil {
		fmt.Printf("ERROR: SET in db 2 failed: %v\n", err)
		return
	}
	fmt.Println("✓ SET key2 in cache '2'")

	// Verify data via HTTP API
	fmt.Println("\nVerifying data via HTTP API...")

	// Check default cache
	req, _ = http.NewRequest("GET", "http://localhost:8080/api/v1/keys/key0", nil)
	req.Header.Set("X-Cache-Name", "default")
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("ERROR: Failed to get key0 from default cache (status: %d)\n", resp.StatusCode)
		return
	}
	var val string
	json.NewDecoder(resp.Body).Decode(&val)
	resp.Body.Close()
	fmt.Printf("✓ HTTP GET key0 from 'default' cache: %s\n", val)

	// Check cache "1"
	req, _ = http.NewRequest("GET", "http://localhost:8080/api/v1/keys/key1", nil)
	req.Header.Set("X-Cache-Name", "1")
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("ERROR: Failed to get key1 from cache 1 (status: %d)\n", resp.StatusCode)
		return
	}
	json.NewDecoder(resp.Body).Decode(&val)
	resp.Body.Close()
	fmt.Printf("✓ HTTP GET key1 from cache '1': %s\n", val)

	// Check cache "2"
	req, _ = http.NewRequest("GET", "http://localhost:8080/api/v1/keys/key2", nil)
	req.Header.Set("X-Cache-Name", "2")
	resp, err = http.DefaultClient.Do(req)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("ERROR: Failed to get key2 from cache 2 (status: %d)\n", resp.StatusCode)
		return
	}
	json.NewDecoder(resp.Body).Decode(&val)
	resp.Body.Close()
	fmt.Printf("✓ HTTP GET key2 from cache '2': %s\n", val)

	// Switch back to default and verify isolation
	fmt.Println("\nVerifying cache isolation...")
	err = client.Do(ctx, "SELECT", "0").Err()
	if err != nil {
		fmt.Printf("ERROR: SELECT 0 failed: %v\n", err)
		return
	}

	// key1 should not exist in default cache
	_, err = client.Get(ctx, "key1").Result()
	if err == nil {
		fmt.Printf("ERROR: key1 should not exist in default cache!\n")
		return
	}
	fmt.Println("✓ key1 does not exist in default cache (correct isolation)")

	// key0 should exist
	val0, err := client.Get(ctx, "key0").Result()
	if err != nil {
		fmt.Printf("ERROR: key0 should exist in default cache: %v\n", err)
		return
	}
	fmt.Printf("✓ key0 exists in default cache: %s\n", val0)

	fmt.Println("\n✅ All SELECT/cache tests passed!")
	fmt.Println("\nCache mapping:")
	fmt.Println("  SELECT 0  → cache 'default'")
	fmt.Println("  SELECT 1  → cache '1'")
	fmt.Println("  SELECT 2  → cache '2'")
	fmt.Println("  SELECT N  → cache 'N' (numeric names only)")
}

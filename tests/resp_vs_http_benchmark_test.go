package tests

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"testing"

	"github.com/redis/go-redis/v9"
)

// Benchmark helpers
var (
	httpBaseURL  = "http://localhost:8080"
	respAddr     = "localhost:6379"
	testCacheName = "default"
)

func setupHTTPClient() *http.Client {
	return &http.Client{}
}

func setupRESPClientBench() *redis.Client {
	return redis.NewClient(&redis.Options{
		Addr: respAddr,
	})
}

func httpGET(client *http.Client, key string) (string, error) {
	req, _ := http.NewRequest("GET", fmt.Sprintf("%s/api/v1/keys/%s", httpBaseURL, key), nil)
	req.Header.Set("X-Cache-Name", testCacheName)
	resp, err := client.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var value string
	json.Unmarshal(body, &value)
	return value, nil
}

func httpSET(client *http.Client, key, value string) error {
	data := map[string]interface{}{key: value}
	jsonData, _ := json.Marshal(data)

	req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/keys", httpBaseURL), bytes.NewBuffer(jsonData))
	req.Header.Set("X-Cache-Name", testCacheName)
	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

func httpDEL(client *http.Client, key string) error {
	req, _ := http.NewRequest("DELETE", fmt.Sprintf("%s/api/v1/keys/%s", httpBaseURL, key), nil)
	req.Header.Set("X-Cache-Name", testCacheName)
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	return nil
}

// Benchmarks: Simple String Operations

func BenchmarkRESP_GET(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Setup: Create key
	client.Set(ctx, "bench:key", "value", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Get(ctx, "bench:key")
	}
}

func BenchmarkHTTP_GET(b *testing.B) {
	client := setupHTTPClient()

	// Setup: Create key via RESP (faster than HTTP)
	respClient := setupRESPClientBench()
	ctx := context.Background()
	respClient.Set(ctx, "bench:key", "value", 0)
	respClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpGET(client, "bench:key")
	}
}

func BenchmarkRESP_SET(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Set(ctx, "bench:key", "value", 0)
	}
}

func BenchmarkHTTP_SET(b *testing.B) {
	client := setupHTTPClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpSET(client, "bench:key", "value")
	}
}

func BenchmarkRESP_INCR(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Setup
	client.Set(ctx, "bench:counter", "0", 0)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.Incr(ctx, "bench:counter")
	}
}

func BenchmarkHTTP_INCR(b *testing.B) {
	client := setupHTTPClient()

	// Setup via RESP
	respClient := setupRESPClientBench()
	ctx := context.Background()
	respClient.Set(ctx, "bench:counter", "0", 0)
	respClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// HTTP increment via command API
		cmdData := map[string]interface{}{
			"command": "INC",
			"key":     "bench:counter",
			"value":   1,
		}
		jsonData, _ := json.Marshal(cmdData)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/commands", httpBaseURL), bytes.NewBuffer(jsonData))
		req.Header.Set("X-Cache-Name", testCacheName)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()
	}
}

// Benchmarks: Hash Operations

func BenchmarkRESP_HSET(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.HSet(ctx, "bench:hash", "field", "value")
	}
}

func BenchmarkHTTP_HSET(b *testing.B) {
	client := setupHTTPClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// HTTP hash set via nested path
		data := map[string]interface{}{"bench:hash/field": "value"}
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/keys", httpBaseURL), bytes.NewBuffer(jsonData))
		req.Header.Set("X-Cache-Name", testCacheName)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()
	}
}

func BenchmarkRESP_HGET(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Setup
	client.HSet(ctx, "bench:hash", "field", "value")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.HGet(ctx, "bench:hash", "field")
	}
}

func BenchmarkHTTP_HGET(b *testing.B) {
	client := setupHTTPClient()

	// Setup via RESP
	respClient := setupRESPClientBench()
	ctx := context.Background()
	respClient.HSet(ctx, "bench:hash", "field", "value")
	respClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		httpGET(client, "bench:hash/field")
	}
}

// Benchmarks: List Operations

func BenchmarkRESP_LPUSH(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Setup: Create empty list
	client.Del(ctx, "bench:list")
	client.RPush(ctx, "bench:list", "init")

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.LPush(ctx, "bench:list", "value")
	}
}

func BenchmarkHTTP_LPUSH(b *testing.B) {
	client := setupHTTPClient()

	// Setup via RESP
	respClient := setupRESPClientBench()
	ctx := context.Background()
	respClient.Del(ctx, "bench:list")
	respClient.RPush(ctx, "bench:list", "init")
	respClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// HTTP list operations are more complex, use append
		cmdData := map[string]interface{}{
			"command": "REPLACE",
			"key":     "bench:list",
			"value":   []string{"value", "init"}, // Prepend via full replacement
		}
		jsonData, _ := json.Marshal(cmdData)
		req, _ := http.NewRequest("POST", fmt.Sprintf("%s/api/v1/commands", httpBaseURL), bytes.NewBuffer(jsonData))
		req.Header.Set("X-Cache-Name", testCacheName)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()
	}
}

// Benchmarks: Mixed Operations (Realistic Workload)

func BenchmarkRESP_MixedOperations(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate typical Redis usage pattern
		client.Set(ctx, fmt.Sprintf("bench:user:%d", i), "data", 0)
		client.HSet(ctx, fmt.Sprintf("bench:profile:%d", i), "name", "John")
		client.Get(ctx, fmt.Sprintf("bench:user:%d", i))
		client.Incr(ctx, "bench:total:users")
	}
}

func BenchmarkHTTP_MixedOperations(b *testing.B) {
	client := setupHTTPClient()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate typical HTTP API usage pattern
		httpSET(client, fmt.Sprintf("bench:user:%d", i), "data")

		data := map[string]interface{}{fmt.Sprintf("bench:profile:%d/name", i): "John"}
		jsonData, _ := json.Marshal(data)
		req, _ := http.NewRequest("PUT", fmt.Sprintf("%s/api/v1/keys", httpBaseURL), bytes.NewBuffer(jsonData))
		req.Header.Set("X-Cache-Name", testCacheName)
		req.Header.Set("Content-Type", "application/json")
		resp, _ := client.Do(req)
		resp.Body.Close()

		httpGET(client, fmt.Sprintf("bench:user:%d", i))

		cmdData := map[string]interface{}{
			"command": "INC",
			"key":     "bench:total:users",
			"value":   1,
		}
		jsonData, _ = json.Marshal(cmdData)
		req, _ = http.NewRequest("POST", fmt.Sprintf("%s/api/v1/commands", httpBaseURL), bytes.NewBuffer(jsonData))
		req.Header.Set("X-Cache-Name", testCacheName)
		req.Header.Set("Content-Type", "application/json")
		resp, _ = client.Do(req)
		resp.Body.Close()
	}
}

// Benchmarks: Batch Operations

func BenchmarkRESP_MGET(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Setup: Create multiple keys
	for i := 0; i < 10; i++ {
		client.Set(ctx, fmt.Sprintf("bench:multi:%d", i), fmt.Sprintf("value%d", i), 0)
	}

	keys := make([]string, 10)
	for i := 0; i < 10; i++ {
		keys[i] = fmt.Sprintf("bench:multi:%d", i)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.MGet(ctx, keys...)
	}
}

func BenchmarkHTTP_MGET(b *testing.B) {
	client := setupHTTPClient()

	// Setup via RESP
	respClient := setupRESPClientBench()
	ctx := context.Background()
	for i := 0; i < 10; i++ {
		respClient.Set(ctx, fmt.Sprintf("bench:multi:%d", i), fmt.Sprintf("value%d", i), 0)
	}
	respClient.Close()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// HTTP batch get is not as efficient - need to make multiple requests or use wildcard
		for j := 0; j < 10; j++ {
			httpGET(client, fmt.Sprintf("bench:multi:%d", j))
		}
	}
}

// Cleanup function to run after all benchmarks
func BenchmarkCleanup(b *testing.B) {
	client := setupRESPClientBench()
	defer client.Close()
	ctx := context.Background()

	// Clean up all benchmark keys
	client.FlushDB(ctx)

	b.Skip("Cleanup benchmark - skip reporting")
}

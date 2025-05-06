# map-cache

A map-based, in-memory caching service with an HTTP API.
Built with Go, designed for local or distributed caching scenarios where simple map-based storage and atomic operations are needed.

Supports:

Named caches
CRUD operations
Postman collection for quick API testing
OpenAPI/Swagger documentation

## Features

* Create and manage caches dynamically 
* CRUD operations on cached and nested keys/values
* REST API
* OpenAPI/Swagger UI for easy reference
* Postman collection for testing

# Examples
## Create a cache
```go
package main
import (
    "fmt"
    "net/http"
    "time"

    "github.com/goodblaster/map-cache"
)

func main() {
    // Create a new cache
    cache := mapcache.NewCache("myCache")
    
    // Acquire the cache using a unique key
    key := "myUniqueKey"
    cache.Acquire(key)
    defer cache.Release(key) // Release the cache when done
    
    // Set a value in the cache
    cache.Set("key1", "value1")

    // Get the value from the cache
    value, err := cache.Get("key1")
    if err != nil {
        fmt.Println("Error:", err)
        return
    }

    fmt.Println("Value:", value) // Output: Value: value1
}
```

## Send a request to add some keys to the default cache
```bash
curl --location 'http://localhost:8080/api/v1/keys' \
--header 'Content-Type: application/json' \
--data '{
  "entries": {
    "key1": "value1",
    "key2": 42,
    "key3": {
        "nestedKey1": "nestedValue1",
        "nestedKey2": 3.14
    }
  }
}'
```

## Request a nested value
```bash
curl --location 'http://localhost:8080/api/v1/keys/key3/nestedKey2'
```

## Request a value in an alternate cache
```bash
curl --location 'http://localhost:8080/api/v1/keys/key1' \
--header 'X-Cache-Name: exampleCache'
```
# map-cache

### API documentation and demo at ...
Demo removed due to malicious traffic. Download and run the service to view Swagger docs.

A map-based, in-memory caching service with an HTTP API. Built with Go, designed for local or distributed caching scenarios where simple map-based storage and atomic operations are needed.

## Features

* Create and manage named caches dynamically
* CRUD operations on cached and nested keys/values
* Command execution for batch updates and conditional logic
* Triggers to automate reactions to data changes
* Key expiration (TTL)
* REST API with OpenAPI/Swagger UI
* Postman collection for testing

---

## Quickstart

### Create a cache
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

### Add keys to a cache
```bash
curl --location 'http://localhost:8080/api/v1/keys' --header 'Content-Type: application/json' --data '{
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

### Get a nested value
```bash
curl --location 'http://localhost:8080/api/v1/keys/key3/nestedKey2'
```

### Use a different cache
```bash
curl --location 'http://localhost:8080/api/v1/keys/key1' --header 'X-Cache-Name: exampleCache'
```

> If no `X-Cache-Name` header is provided, the system uses the default cache.

---

## Commands

Commands allow atomic operations across the cache, including logic-based updates. Submit commands via:

```http
POST /api/v1/commands/execute
Header: X-Cache-Name: your-cache
```

### Example
```json
{
  "commands": [
    { "type": "INC", "key": "domains/domain-1/countdown", "value": -1 },
    { "type": "RETURN", "key": "${{status}}" }
  ]
}
```

### Supported Command Types

- **INC**: Increment/decrement a numeric key
- **REPLACE**: Overwrite a key with a new value
- **RETURN**: Return a value or expression result
- **IF**: Conditionally execute one of two subcommands
- **FOR**: Iterate over a wildcard pattern and execute commands with captured values

---

## Interpolation & Key Access

You can reference values dynamically using `${{key/path}}`. Supported patterns include:

- `${{key}}` → gets the value at `key`
- `${{parent/child}}` → gets nested values
- `${{some/*/value}}` → wildcard to resolve many keys

Within `FOR` and `IF` blocks, you can reference captured values using `${{1}}`, `${{2}}`, etc., from the wildcard path match.

### Example:
```json
{
  "commands": [
    {
      "type": "FOR",
      "loop_expr": "${{job-1234/domains/*/countdown}}",
      "commands": [
        {
          "type": "IF",
          "condition": "${{job-1234/domains/${{1}}/countdown}} == 0",
          "if_true": {
            "type": "REPLACE",
            "key": "job-1234/domains/${{1}}/status",
            "value": "complete"
          },
          "if_false": {
            "type": "INC",
            "key": "job-1234/domains/${{1}}/countdown",
            "value": -1
          }
        }
      ]
    },
    {
      "type": "RETURN",
      "key": "current job status is ${{job-1234/status}}"
    }
  ]
}
```

---

## Triggers

Triggers are bound to key patterns. When a matching key is updated, the trigger's command executes automatically.

### Create a trigger
```bash
POST /api/v1/triggers
Header: X-Cache-Name: your-cache
```

### Example:
```json
{
  "key": "domains/*/countdown",
  "command": {
    "type": "IF",
    "condition": "${{domains/${{1}}/countdown}} <= 0",
    "if_true": { "type": "REPLACE", "key": "domains/${{1}}/status", "value": "complete" },
    "if_false": { "type": "NOOP" }
  }
}
```

---

## Expirations (TTL)

When creating or setting a key, you may provide a TTL in **milliseconds**. The key will be automatically removed after this duration.

Support for expirations is currently in the implementation phase. Contact maintainers for progress or contribute via pull request.

---

## Full Scenario (Countdown)

See the included Postman collection (`map-cache-scenario-countdown.postman_collection.json`) for a multi-step example.

Steps:

1. **Create a cache**: `job-1234`
2. **Add keys**: two domains each with `status: busy` and `countdown: 2`
3. **Create trigger** on `domains/*/countdown` to mark status complete when countdown reaches zero
4. **Create another trigger** on `domains/*/status` to mark global `status` as complete when all domains are done
5. **Decrement countdowns** with `INC`
6. **Check final status** via `RETURN`

---

## Development & Testing

Run big scale tests with:
```go
go test -v -run Test_Big ./...
```

This simulates countdown and cascading completion logic over 100+ domains.

---

## License

MIT

---

For questions or contributions, contact [dave@goodblaster.com](mailto:dave@goodblaster.com)

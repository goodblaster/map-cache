# Performance Benchmarks: RESP vs HTTP

This document compares the performance of the Redis Protocol (RESP) implementation against the HTTP/JSON API for common operations.

## Test Environment

- **Platform:** darwin arm64
- **CPU:** Apple M4 Pro
- **Go Version:** (as per go.mod)
- **Test Duration:** 1 second per benchmark

## Benchmark Results

All measurements in nanoseconds per operation (ns/op), with memory allocations and allocation count.

### String Operations

| Operation | Protocol | Time (ns/op) | Memory (B/op) | Allocs/op | Speedup |
|-----------|----------|--------------|---------------|-----------|---------|
| GET       | RESP     | 22,617       | 208           | 7         | **2.1x** |
|           | HTTP     | 48,206       | 5,054         | 60        | |
| SET       | RESP     | 22,383       | 251           | 7         | **2.0x** |
|           | HTTP     | 43,721       | 4,939         | 61        | |
| INCR      | RESP     | 23,876       | 440           | 26        | **9.6x** |
|           | HTTP     | 229,651      | 19,293        | 147       | |

### Hash Operations

| Operation | Protocol | Time (ns/op) | Memory (B/op) | Allocs/op | Speedup |
|-----------|----------|--------------|---------------|-----------|---------|
| HSET      | RESP     | 22,919       | 217           | 6         | **2.0x** |
|           | HTTP     | 46,265       | 4,963         | 60        | |
| HGET      | RESP     | 21,085       | 240           | 8         | **2.4x** |
|           | HTTP     | 49,802       | 5,071         | 60        | |

### List Operations

| Operation | Protocol | Time (ns/op) | Memory (B/op) | Allocs/op | Speedup |
|-----------|----------|--------------|---------------|-----------|---------|
| LPUSH     | RESP     | 86,425       | 200           | 6         | **2.6x** |
|           | HTTP     | 225,302      | 19,499        | 149       | |

### Batch Operations

| Operation | Protocol | Time (ns/op) | Memory (B/op) | Allocs/op | Speedup |
|-----------|----------|--------------|---------------|-----------|---------|
| MGET (10) | RESP     | 26,814       | 888           | 36        | **18.4x** |
|           | HTTP     | 492,617      | 50,712        | 610       | |

### Mixed Workload

Simulates realistic usage: SET user data, HSET profile field, GET user data, INCR counter.

| Protocol | Time (ns/op) | Memory (B/op) | Allocs/op | Speedup |
|----------|--------------|---------------|-----------|---------|
| RESP     | 93,411       | 949           | 31        | **6.2x** |
| HTTP     | 574,839      | 35,118        | 341       | |

## Key Findings

### Performance

1. **RESP is consistently faster:** 2-20x improvement across all operations
2. **Simple operations:** GET/SET show 2-2.5x speedup
3. **Complex operations:** INCR and mixed workloads show 6-10x speedup
4. **Batch operations:** Biggest advantage with 18x speedup for MGET
5. **Best for high-throughput workloads:** RESP excels in scenarios requiring many small operations

### Memory Efficiency

1. **RESP uses 5-25x less memory per operation**
2. **Fewer allocations:** 3-10x fewer allocations per operation
3. **Lower GC pressure:** Reduced memory allocations mean less garbage collection overhead

## Performance Breakdown

### Why RESP is Faster

1. **Binary Protocol:** RESP uses a compact binary format vs HTTP's verbose text-based protocol
2. **Less Parsing:** Simple RESP parser vs full HTTP + JSON parsing
3. **Smaller Payloads:** Binary encoding is more compact than JSON
4. **Connection Reuse:** go-redis client efficiently reuses connections
5. **Lower Overhead:** No HTTP headers, JSON marshaling/unmarshaling overhead

### HTTP Still Has Advantages

1. **Human-readable:** Easier to debug with curl/browser tools
2. **RESTful:** Familiar API for web developers
3. **Widespread tooling:** Many HTTP clients and monitoring tools
4. **Firewall-friendly:** Port 80/443 often open in corporate environments
5. **Rich features:** Easy to add custom headers, authentication, etc.

## Recommendations

### Use RESP When:
- **High throughput required:** Processing thousands of operations per second
- **Low latency critical:** Sub-millisecond response times needed
- **Memory constrained:** Limited RAM or high GC pressure
- **Batch operations:** Need to GET/SET many keys efficiently
- **Redis compatibility:** Migrating from Redis or want to use existing Redis tools

### Use HTTP When:
- **Web integration:** Building web applications or microservices
- **Debugging:** Need to inspect/modify data with curl or browser
- **Complex queries:** Using the command DSL (IF, FOR, triggers, etc.)
- **Authentication/Authorization:** Need custom HTTP middleware
- **Cross-platform:** Accessing from languages without Redis clients

## Throughput Estimates

Based on single-operation benchmarks:

### RESP
- **GET/SET:** ~44,000 ops/sec
- **Hash ops:** ~45,000 ops/sec
- **Mixed workload:** ~10,700 ops/sec

### HTTP
- **GET/SET:** ~20,000-22,000 ops/sec
- **Hash ops:** ~20,000 ops/sec
- **Mixed workload:** ~1,740 ops/sec

*Note: These are single-threaded estimates. Actual throughput depends on concurrency, network latency, and workload characteristics.*

## Conclusion

The RESP protocol implementation provides **significant performance improvements** over HTTP/JSON:

- **2-10x faster** for most operations
- **5-25x less memory** per operation
- **18x faster** for batch operations

For performance-critical applications or high-throughput workloads, RESP is the clear choice. HTTP remains valuable for web integration, debugging, and scenarios where ease of use outweighs raw performance.

Both protocols share the same underlying cache implementation, so you can use both simultaneously and choose the right protocol for each use case.

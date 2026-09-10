# ratelimit

A Go package for token bucket rate limiting. Allows bursts of traffic up to a configured capacity while maintaining a steady average rate over time. Thread-safe and suitable for single-instance in-memory deployments.

## Installation

```bash
go get github.com/hekimapro/ratelimit
```

## How It Works

The token bucket algorithm gives each client a bucket that fills with tokens at a fixed rate up to a maximum capacity. Each request consumes one token. When the bucket is empty, requests are rejected until tokens refill.

```
Bucket capacity: 30 tokens  (max burst)
Refill rate:     10/sec      (steady-state throughput)

[▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓▓] ← full bucket, burst of 30 allowed
[▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░] ← 10 tokens left after 20 requests
[░░░░░░░░░░░░░░░░░░░░░░░░░░░░░░] ← empty, requests blocked
[▓▓▓▓▓▓▓▓▓▓░░░░░░░░░░░░░░░░░░░░] ← 10 tokens back after 1 second
```

## Quick Start

```go
import "github.com/hekimapro/ratelimit"

// One limiter manages all clients — each key gets its own bucket
limiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.DefaultConfiguration())

if !limiter.Allow("user:42") {
    http.Error(w, "Too Many Requests", http.StatusTooManyRequests)
    return
}
```

## Configuration Presets

| Preset | Refill Rate | Burst Capacity | Use Case |
|---|---|---|---|
| `DefaultConfiguration()` | 10 req/s | 30 | General-purpose APIs |
| `StrictConfiguration()` | 1 req/s | 3 | High-security endpoints (login, password reset) |
| `PermissiveConfiguration()` | 100 req/s | 200 | Internal service-to-service calls |

```go
// General API
limiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.DefaultConfiguration())

// Login endpoint
loginLimiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.StrictConfiguration())

// Internal service
internalLimiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.PermissiveConfiguration())
```

## Custom Configuration

```go
limiter := ratelimit.NewInMemoryLimiter(
    25.0,  // refill rate: 25 tokens per second
    50.0,  // bucket capacity: bursts up to 50
)

// Or with full control via the config struct
limiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.Configuration{
    RefillRatePerSecond: 25.0,
    BucketCapacity:      50.0,
    InitialTokens:       10.0, // start with a partial bucket
})
```

`InitialTokens` defaults to `BucketCapacity` (full bucket) if zero or omitted.

## HTTP Middleware

A complete rate limiting middleware that sets standard `X-RateLimit-*` headers:

```go
func rateLimitMiddleware(limiter *ratelimit.InMemoryLimiter, next http.Handler) http.Handler {
    return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        // Key by IP address — or use user ID / API key for authenticated routes
        ip := r.RemoteAddr

        remaining := limiter.Remaining(ip)
        w.Header().Set("X-RateLimit-Remaining", strconv.Itoa(remaining))
        w.Header().Set("X-RateLimit-Limit", "30")

        if !limiter.Allow(ip) {
            w.Header().Set("Retry-After", "1")
            http.Error(w, `{"error":"rate limit exceeded"}`, http.StatusTooManyRequests)
            return
        }

        next.ServeHTTP(w, r)
    })
}
```

## Batch Operations with `AllowN`

Use `AllowN` when a single request should consume more than one token — useful for search endpoints, export jobs, or weighted API operations:

```go
// A bulk export request costs 10 tokens
if !limiter.AllowN("user:42", 10) {
    http.Error(w, "Rate limit exceeded", http.StatusTooManyRequests)
    return
}
```

## Per-Client vs Shared Buckets

**`InMemoryLimiter`** (recommended for HTTP middleware) — manages one bucket per key automatically:

```go
limiter := ratelimit.NewInMemoryLimiterWithConfig(ratelimit.DefaultConfiguration())

limiter.Allow("ip:203.0.113.5")   // bucket auto-created on first call
limiter.Allow("user:99")
limiter.Allow("apikey:abc123")
```

**`TokenBucket`** (lower-level) — manage a single bucket directly:

```go
bucket := ratelimit.NewTokenBucket(ratelimit.DefaultConfiguration())

bucket.Allow()               // consume 1 token
bucket.AllowN(5)             // consume 5 tokens
bucket.RemainingTokens()     // check current balance
bucket.Reset()               // refill to full capacity
```

## Administrative Operations

```go
// Reset a specific client after a cooling-off period
limiter.ResetKey("ip:203.0.113.5")

// Inspect limiter state
stats := limiter.GetStats()
fmt.Println(stats.TotalKeys)      // number of tracked clients
fmt.Println(stats.RefillRate)     // tokens per second
fmt.Println(stats.BucketCapacity) // max burst size
```

## Memory Management

`InMemoryLimiter` grows as new client keys arrive. Call `Cleanup` periodically to remove buckets that have been idle longer than a given duration, preventing unbounded memory growth:

```go
// In a background goroutine, evict buckets idle for more than 10 minutes
go func() {
    ticker := time.NewTicker(5 * time.Minute)
    defer ticker.Stop()
    for range ticker.C {
        limiter.Cleanup(10 * time.Minute)
    }
}()
```

> **Note:** `Cleanup` requires `lastAccessTime` tracking on `TokenBucket` to be fully effective — this is noted in the source as a production enhancement. Until added, it is safe to call but will not evict buckets.

## Distributed Deployments

This implementation stores buckets **in memory** and is scoped to a single process. For deployments with multiple server instances, each instance maintains independent limits. To share limits across instances, pair this package with a distributed backend such as Redis and implement a sliding window or token bucket against a shared counter.

## API Reference

### `InMemoryLimiter`

| Method | Description |
|---|---|
| `NewInMemoryLimiter(rate, capacity)` | Create a limiter with explicit values |
| `NewInMemoryLimiterWithConfig(cfg)` | Create a limiter from a `Configuration` |
| `Allow(key)` | Consume 1 token for `key` — returns `true` if allowed |
| `AllowN(key, n)` | Consume `n` tokens for `key` — returns `true` if allowed |
| `Remaining(key)` | Approximate tokens remaining for `key` |
| `ResetKey(key)` | Refill the bucket for `key` to full capacity |
| `Cleanup(maxIdle)` | Remove buckets idle longer than `maxIdle` |
| `GetStats()` | Return `LimiterStats` with key count and configuration |

### `TokenBucket`

| Method | Description |
|---|---|
| `NewTokenBucket(cfg)` | Create a single token bucket |
| `Allow()` | Consume 1 token — returns `true` if allowed |
| `AllowN(n)` | Consume `n` tokens — returns `true` if allowed |
| `RemainingTokens()` | Approximate current token count |
| `Reset()` | Refill to full capacity |

## License

See [LICENSE](LICENSE).

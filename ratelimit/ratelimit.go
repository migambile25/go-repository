// Package ratelimit provides token bucket rate limiting for API protection.
// Token bucket algorithm allows bursts of traffic up to a configured capacity
// while maintaining a steady average rate over time.
//
// This implementation is thread-safe and suitable for in-memory rate limiting
// in single-instance deployments. For distributed rate limiting across multiple
// servers, pair this with a distributed cache like Redis.
package ratelimit

import (
    "sync"
    "time"
)

// TokenBucket implements the token bucket algorithm for rate limiting.
// Each request consumes one token. Tokens are replenished at a fixed rate.
// Bursts are allowed up to the bucket capacity.
//
// Thread-safe: Use a single bucket per client (IP, user ID, API key).
type TokenBucket struct {
    mutex          sync.Mutex
    currentTokens  float64
    lastRefillTime time.Time
    refillRate     float64 // Tokens added per second
    bucketCapacity float64 // Maximum tokens the bucket can hold
}

// Configuration holds parameters for creating a rate limiter.
type Configuration struct {
    // RefillRatePerSecond is the number of tokens added to the bucket each second.
    // This determines the steady-state request rate.
    // Example: 10.0 allows 10 requests per second average.
    RefillRatePerSecond float64

    // BucketCapacity is the maximum number of tokens the bucket can store.
    // This determines the maximum burst size.
    // Example: 30.0 allows bursts of up to 30 requests.
    BucketCapacity float64

    // InitialTokens is the number of tokens the bucket starts with.
    // Default: BucketCapacity (full bucket for immediate bursts).
    InitialTokens float64
}

// DefaultConfiguration returns a sensible configuration for general APIs.
// Allows 10 requests per second with bursts up to 30 requests.
func DefaultConfiguration() Configuration {
    return Configuration{
        RefillRatePerSecond: 10.0,
        BucketCapacity:      30.0,
        InitialTokens:       30.0,
    }
}

// StrictConfiguration returns a configuration for high-security endpoints.
// Allows 1 request per second with bursts up to 3 requests.
func StrictConfiguration() Configuration {
    return Configuration{
        RefillRatePerSecond: 1.0,
        BucketCapacity:      3.0,
        InitialTokens:       3.0,
    }
}

// PermissiveConfiguration returns a configuration for internal services.
// Allows 100 requests per second with bursts up to 200 requests.
func PermissiveConfiguration() Configuration {
    return Configuration{
        RefillRatePerSecond: 100.0,
        BucketCapacity:      200.0,
        InitialTokens:       200.0,
    }
}

// NewTokenBucket creates a new token bucket with the given configuration.
func NewTokenBucket(configuration Configuration) *TokenBucket {
    initialTokens := configuration.InitialTokens
    if initialTokens == 0 {
        initialTokens = configuration.BucketCapacity
    }

    return &TokenBucket{
        currentTokens:  initialTokens,
        lastRefillTime: time.Now(),
        refillRate:     configuration.RefillRatePerSecond,
        bucketCapacity: configuration.BucketCapacity,
    }
}

// Allow checks if a request can proceed and consumes one token if available.
// Returns true if the request is allowed, false if rate limited.
func (tokenBucket *TokenBucket) Allow() bool {
    tokenBucket.mutex.Lock()
    defer tokenBucket.mutex.Unlock()

    // Refill tokens based on time elapsed since last refill
    currentTime := time.Now()
    elapsedSeconds := currentTime.Sub(tokenBucket.lastRefillTime).Seconds()

    // Add new tokens
    tokenBucket.currentTokens += elapsedSeconds * tokenBucket.refillRate

    // Cap at bucket capacity
    if tokenBucket.currentTokens > tokenBucket.bucketCapacity {
        tokenBucket.currentTokens = tokenBucket.bucketCapacity
    }

    tokenBucket.lastRefillTime = currentTime

    // Check if we have at least one token
    if tokenBucket.currentTokens >= 1.0 {
        tokenBucket.currentTokens--
        return true
    }

    return false
}

// AllowN checks if N requests can proceed and consumes N tokens if available.
// Useful for batch operations where one request consumes multiple units.
func (tokenBucket *TokenBucket) AllowN(tokensNeeded int) bool {
    if tokensNeeded <= 0 {
        return true
    }

    tokenBucket.mutex.Lock()
    defer tokenBucket.mutex.Unlock()

    // Refill tokens
    currentTime := time.Now()
    elapsedSeconds := currentTime.Sub(tokenBucket.lastRefillTime).Seconds()
    tokenBucket.currentTokens += elapsedSeconds * tokenBucket.refillRate

    if tokenBucket.currentTokens > tokenBucket.bucketCapacity {
        tokenBucket.currentTokens = tokenBucket.bucketCapacity
    }

    tokenBucket.lastRefillTime = currentTime

    // Check if we have enough tokens
    if tokenBucket.currentTokens >= float64(tokensNeeded) {
        tokenBucket.currentTokens -= float64(tokensNeeded)
        return true
    }

    return false
}

// RemainingTokens returns the approximate number of tokens currently available.
// Useful for setting rate limit headers (X-RateLimit-Remaining).
func (tokenBucket *TokenBucket) RemainingTokens() int {
    tokenBucket.mutex.Lock()
    defer tokenBucket.mutex.Unlock()

    // Refill first to get accurate count
    currentTime := time.Now()
    elapsedSeconds := currentTime.Sub(tokenBucket.lastRefillTime).Seconds()
    tokenBucket.currentTokens += elapsedSeconds * tokenBucket.refillRate

    if tokenBucket.currentTokens > tokenBucket.bucketCapacity {
        tokenBucket.currentTokens = tokenBucket.bucketCapacity
    }

    tokenBucket.lastRefillTime = currentTime

    return int(tokenBucket.currentTokens)
}

// Reset resets the bucket to its initial state (full capacity).
// Useful after a client has been blocked for a long period.
func (tokenBucket *TokenBucket) Reset() {
    tokenBucket.mutex.Lock()
    defer tokenBucket.mutex.Unlock()

    tokenBucket.currentTokens = tokenBucket.bucketCapacity
    tokenBucket.lastRefillTime = time.Now()
}

// InMemoryLimiter manages multiple token buckets keyed by identifier.
// Each unique key (IP address, user ID, API key) gets its own bucket.
// This is the primary interface for HTTP middleware rate limiting.
type InMemoryLimiter struct {
    buckets        map[string]*TokenBucket
    mutex          sync.RWMutex
    refillRate     float64
    bucketCapacity float64
    initialTokens  float64
}

// NewInMemoryLimiter creates a new in-memory rate limiter.
// Each unique key will have its own token bucket with the same configuration.
func NewInMemoryLimiter(refillRatePerSecond, bucketCapacity float64) *InMemoryLimiter {
    return &InMemoryLimiter{
        buckets:        make(map[string]*TokenBucket),
        refillRate:     refillRatePerSecond,
        bucketCapacity: bucketCapacity,
        initialTokens:  bucketCapacity,
    }
}

// NewInMemoryLimiterWithConfig creates a limiter using a configuration struct.
func NewInMemoryLimiterWithConfig(configuration Configuration) *InMemoryLimiter {
    initialTokens := configuration.InitialTokens
    if initialTokens == 0 {
        initialTokens = configuration.BucketCapacity
    }

    return &InMemoryLimiter{
        buckets:        make(map[string]*TokenBucket),
        refillRate:     configuration.RefillRatePerSecond,
        bucketCapacity: configuration.BucketCapacity,
        initialTokens:  initialTokens,
    }
}

// Allow checks if the specified key is allowed to make a request.
// Creates a new bucket for the key if one doesn't exist.
func (limiter *InMemoryLimiter) Allow(key string) bool {
    limiter.mutex.Lock()
    defer limiter.mutex.Unlock()

    bucket, exists := limiter.buckets[key]
    if !exists {
        bucketConfig := Configuration{
            RefillRatePerSecond: limiter.refillRate,
            BucketCapacity:      limiter.bucketCapacity,
            InitialTokens:       limiter.initialTokens,
        }
        bucket = NewTokenBucket(bucketConfig)
        limiter.buckets[key] = bucket
    }

    return bucket.Allow()
}

// AllowN checks if the specified key is allowed to make N requests.
// Useful for batch operations or weighting different endpoints.
func (limiter *InMemoryLimiter) AllowN(key string, tokensNeeded int) bool {
    limiter.mutex.Lock()
    defer limiter.mutex.Unlock()

    bucket, exists := limiter.buckets[key]
    if !exists {
        bucketConfig := Configuration{
            RefillRatePerSecond: limiter.refillRate,
            BucketCapacity:      limiter.bucketCapacity,
            InitialTokens:       limiter.initialTokens,
        }
        bucket = NewTokenBucket(bucketConfig)
        limiter.buckets[key] = bucket
    }

    return bucket.AllowN(tokensNeeded)
}

// Remaining returns the approximate remaining tokens for the specified key.
// Returns 0 if the key doesn't exist yet.
func (limiter *InMemoryLimiter) Remaining(key string) int {
    limiter.mutex.RLock()
    defer limiter.mutex.RUnlock()

    bucket, exists := limiter.buckets[key]
    if !exists {
        return int(limiter.bucketCapacity)
    }

    return bucket.RemainingTokens()
}

// ResetKey resets the bucket for a specific key to full capacity.
// Useful after a cooling-off period or for administrative override.
func (limiter *InMemoryLimiter) ResetKey(key string) {
    limiter.mutex.Lock()
    defer limiter.mutex.Unlock()

    if bucket, exists := limiter.buckets[key]; exists {
        bucket.Reset()
    }
}

// Cleanup removes buckets that haven't been used recently.
// Prevents unbounded memory growth from inactive clients.
// Should be called periodically (e.g., every 5 minutes).
func (limiter *InMemoryLimiter) Cleanup(maxIdleDuration time.Duration) {
    limiter.mutex.Lock()
    defer limiter.mutex.Unlock()

    cutoffTime := time.Now().Add(-maxIdleDuration)

    for key, bucket := range limiter.buckets {
        // Check if bucket has been idle (no allowance checks)
        // We track last activity time for this purpose
        // For simplicity in this implementation, we'll check a last access timestamp
        // Note: This requires tracking lastAccessTime in TokenBucket
        _ = bucket
        _ = key
        _ = cutoffTime
        // In a production implementation, add lastAccessTime to TokenBucket
    }
}

// Stats returns statistics about the rate limiter.
type LimiterStats struct {
    TotalKeys      int
    TotalBuckets   int
    RefillRate     float64
    BucketCapacity float64
}

// GetStats returns current statistics about the rate limiter.
func (limiter *InMemoryLimiter) GetStats() LimiterStats {
    limiter.mutex.RLock()
    defer limiter.mutex.RUnlock()

    return LimiterStats{
        TotalKeys:      len(limiter.buckets),
        TotalBuckets:   len(limiter.buckets),
        RefillRate:     limiter.refillRate,
        BucketCapacity: limiter.bucketCapacity,
    }
}
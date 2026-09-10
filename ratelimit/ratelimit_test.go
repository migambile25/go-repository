package ratelimit

import (
    "sync"
    "testing"
    "time"
)

func TestTokenBucketAllow(t *testing.T) {
    config := Configuration{
        RefillRatePerSecond: 10.0,
        BucketCapacity:      5.0,
        InitialTokens:       5.0,
    }

    bucket := NewTokenBucket(config)

    // Should allow initial burst of 5 requests
    for i := 0; i < 5; i++ {
        if !bucket.Allow() {
            t.Errorf("Request %d should be allowed", i+1)
        }
    }

    // 6th request should be denied (bucket empty)
    if bucket.Allow() {
        t.Errorf("6th request should be denied")
    }

    // Wait for refill
    time.Sleep(100 * time.Millisecond)

    // Should have at least 1 token now
    if !bucket.Allow() {
        t.Errorf("Request after refill should be allowed")
    }
}

func TestTokenBucketAllowN(t *testing.T) {
    config := Configuration{
        RefillRatePerSecond: 10.0,
        BucketCapacity:      10.0,
        InitialTokens:       10.0,
    }

    bucket := NewTokenBucket(config)

    // Should allow 3 tokens
    if !bucket.AllowN(3) {
        t.Errorf("Should allow 3 tokens")
    }

    // Remaining should be 7
    remaining := bucket.RemainingTokens()
    if remaining != 7 {
        t.Errorf("Expected 7 remaining, got %d", remaining)
    }

    // Should not allow 8 tokens (only 7 left)
    if bucket.AllowN(8) {
        t.Errorf("Should not allow 8 tokens")
    }

    // Should allow 7 tokens
    if !bucket.AllowN(7) {
        t.Errorf("Should allow 7 tokens")
    }

    // Should be empty now
    if bucket.Allow() {
        t.Errorf("Should be empty")
    }
}

func TestTokenBucketRefill(t *testing.T) {
    config := Configuration{
        RefillRatePerSecond: 100.0, // 100 tokens per second = 1 per 10ms
        BucketCapacity:      10.0,
        InitialTokens:       0.0,
    }

    bucket := NewTokenBucket(config)

    // Initially empty
    if bucket.Allow() {
        t.Errorf("Initially should be empty")
    }

    // Wait 50ms for 5 tokens
    time.Sleep(50 * time.Millisecond)

    // Should have 5 tokens
    for i := 0; i < 5; i++ {
        if !bucket.Allow() {
            t.Errorf("Request %d after refill should be allowed", i+1)
        }
    }

    // Should be empty again
    if bucket.Allow() {
        t.Errorf("Should be empty after consuming all tokens")
    }
}

func TestTokenBucketCapacity(t *testing.T) {
    config := Configuration{
        RefillRatePerSecond: 10.0,
        BucketCapacity:      5.0,
        InitialTokens:       0.0,
    }

    bucket := NewTokenBucket(config)

    // Wait long enough to refill many tokens
    time.Sleep(2 * time.Second)

    // Should still only have capacity worth of tokens (5)
    for i := 0; i < 5; i++ {
        if !bucket.Allow() {
            t.Errorf("Request %d should be allowed", i+1)
        }
    }

    // 6th should fail (capacity limit)
    if bucket.Allow() {
        t.Errorf("6th request should fail due to capacity limit")
    }
}

func TestTokenBucketReset(t *testing.T) {
    config := Configuration{
        RefillRatePerSecond: 10.0,
        BucketCapacity:      5.0,
        InitialTokens:       5.0,
    }

    bucket := NewTokenBucket(config)

    // Consume all tokens
    for i := 0; i < 5; i++ {
        bucket.Allow()
    }

    // Should be empty
    if bucket.Allow() {
        t.Errorf("Should be empty after consumption")
    }

    // Reset bucket
    bucket.Reset()

    // Should have full capacity again
    for i := 0; i < 5; i++ {
        if !bucket.Allow() {
            t.Errorf("Request %d after reset should be allowed", i+1)
        }
    }
}

func TestInMemoryLimiter(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 5.0)

    // Test single key
    key := "client-123"

    // Should allow 5 requests
    for i := 0; i < 5; i++ {
        if !limiter.Allow(key) {
            t.Errorf("Request %d should be allowed", i+1)
        }
    }

    // 6th should fail
    if limiter.Allow(key) {
        t.Errorf("6th request should fail")
    }

    // Different key should have its own bucket
    otherKey := "client-456"
    if !limiter.Allow(otherKey) {
        t.Errorf("Different client should be allowed")
    }
}

func TestInMemoryLimiterMultipleKeys(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 3.0)

    keys := []string{"key1", "key2", "key3"}

    // Each key should get its own 3-token bucket
    for _, key := range keys {
        for i := 0; i < 3; i++ {
            if !limiter.Allow(key) {
                t.Errorf("Key %s request %d should be allowed", key, i+1)
            }
        }

        // 4th should fail
        if limiter.Allow(key) {
            t.Errorf("Key %s 4th request should fail", key)
        }
    }
}

func TestInMemoryLimiterAllowN(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 10.0)
    key := "batch-client"

    // Should allow 5 tokens at once
    if !limiter.AllowN(key, 5) {
        t.Errorf("Should allow 5 tokens")
    }

    remaining := limiter.Remaining(key)
    if remaining != 5 {
        t.Errorf("Expected 5 remaining, got %d", remaining)
    }

    // Should not allow 6 tokens (only 5 left)
    if limiter.AllowN(key, 6) {
        t.Errorf("Should not allow 6 tokens")
    }

    // Should allow remaining 5
    if !limiter.AllowN(key, 5) {
        t.Errorf("Should allow remaining 5 tokens")
    }
}

func TestInMemoryLimiterRemaining(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 10.0)
    key := "test-client"

    // Initially should have capacity
    remaining := limiter.Remaining(key)
    if remaining != 10 {
        t.Errorf("Expected 10 remaining, got %d", remaining)
    }

    // Consume 3 tokens
    for i := 0; i < 3; i++ {
        limiter.Allow(key)
    }

    remaining = limiter.Remaining(key)
    if remaining != 7 {
        t.Errorf("Expected 7 remaining, got %d", remaining)
    }
}

func TestInMemoryLimiterReset(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 5.0)
    key := "reset-client"

    // Consume all tokens
    for i := 0; i < 5; i++ {
        limiter.Allow(key)
    }

    // Should be limited
    if limiter.Allow(key) {
        t.Errorf("Should be limited")
    }

    // Reset the key
    limiter.ResetKey(key)

    // Should be allowed again
    for i := 0; i < 5; i++ {
        if !limiter.Allow(key) {
            t.Errorf("Request %d after reset should be allowed", i+1)
        }
    }
}

func TestConcurrentAccess(t *testing.T) {
    limiter := NewInMemoryLimiter(100.0, 1000.0)
    key := "concurrent-test"

    const goroutineCount = 100
    const requestsPerGoroutine = 100

    var waitGroup sync.WaitGroup
    allowedCount := int32(0)

    for i := 0; i < goroutineCount; i++ {
        waitGroup.Add(1)
        go func() {
            defer waitGroup.Done()
            for j := 0; j < requestsPerGoroutine; j++ {
                if limiter.Allow(key) {
                    // This is fine for testing concurrency
                }
            }
        }()
    }

    waitGroup.Wait()

    // No panics means success
    _ = allowedCount
}

func TestTokenBucketConcurrency(t *testing.T) {
    config := DefaultConfiguration()
    bucket := NewTokenBucket(config)

    const goroutineCount = 50
    const requestsPerGoroutine = 100

    var waitGroup sync.WaitGroup

    for i := 0; i < goroutineCount; i++ {
        waitGroup.Add(1)
        go func() {
            defer waitGroup.Done()
            for j := 0; j < requestsPerGoroutine; j++ {
                bucket.Allow()
            }
        }()
    }

    waitGroup.Wait()
    // No panic means thread-safe
}

func TestRateLimitHeaders(t *testing.T) {
    limiter := NewInMemoryLimiter(10.0, 5.0)
    key := "header-test"

    // Check remaining after each request
    expectedRemaining := []int{5, 4, 3, 2, 1, 0}

    for i, expected := range expectedRemaining {
        allowed := limiter.Allow(key)
        remaining := limiter.Remaining(key)

        if i < 5 {
            if !allowed {
                t.Errorf("Request %d should be allowed", i+1)
            }
        } else {
            if allowed {
                t.Errorf("Request %d should be denied", i+1)
            }
        }

        if remaining != expected {
            t.Errorf("Request %d: expected %d remaining, got %d", i+1, expected, remaining)
        }
    }
}

func TestDifferentConfigurations(t *testing.T) {
    tests := []struct {
        name         string
        config       Configuration
        expectedRate float64
        expectedCap  float64
    }{
        {
            name:         "default configuration",
            config:       DefaultConfiguration(),
            expectedRate: 10.0,
            expectedCap:  30.0,
        },
        {
            name:         "strict configuration",
            config:       StrictConfiguration(),
            expectedRate: 1.0,
            expectedCap:  3.0,
        },
        {
            name:         "permissive configuration",
            config:       PermissiveConfiguration(),
            expectedRate: 100.0,
            expectedCap:  200.0,
        },
    }

    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            limiter := NewInMemoryLimiterWithConfig(testCase.config)
            key := "test-key"

            // Should allow burst up to capacity
            for i := 0; i < int(testCase.expectedCap); i++ {
                if !limiter.Allow(key) {
                    t.Errorf("Request %d should be allowed", i+1)
                }
            }

            // Should be limited now
            if limiter.Allow(key) {
                t.Errorf("Should be limited after capacity")
            }
        })
    }
}
package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/migambile25/go-repository/ratelimit"
)

func main() {
	fmt.Println("=== Rate Limiter Package Examples ===")

	exampleBasicTokenBucket()
	exampleInMemoryLimiter()
	exampleDifferentClients()
	exampleHTTPMiddleware()
	exampleBatchOperations()
	exampleRateLimitHeaders()
	exampleCustomConfiguration()
	exampleAPIProtection()
}

func exampleBasicTokenBucket() {
	fmt.Println("1. Basic Token Bucket")
	fmt.Println("   -----------------")

	config := ratelimit.Configuration{
		RefillRatePerSecond: 2.0, // 2 tokens per second
		BucketCapacity:      3.0, // Burst up to 3 requests
		InitialTokens:       3.0, // Start full
	}

	bucket := ratelimit.NewTokenBucket(config)

	fmt.Println("   Making 6 requests to 3-capacity bucket (2 tokens/sec):")

	for i := 1; i <= 6; i++ {
		allowed := bucket.Allow()
		remaining := bucket.RemainingTokens()

		if allowed {
			fmt.Printf("   Request %d: ✅ ALLOWED (remaining: %d tokens)\n", i, remaining)
		} else {
			fmt.Printf("   Request %d: ❌ DENIED  (remaining: %d tokens)\n", i, remaining)
		}

		// Wait a bit to see refill in action
		if i == 3 {
			fmt.Println("   ... waiting 500ms for refill ...")
			time.Sleep(500 * time.Millisecond)
		}
	}

	fmt.Println()
}

func exampleInMemoryLimiter() {
	fmt.Println("2. In-Memory Limiter (Multiple Clients)")
	fmt.Println("   ------------------------------------")

	// Create limiter: 5 requests per second, burst up to 10
	limiter := ratelimit.NewInMemoryLimiter(5.0, 10.0)

	clients := []string{"client-A", "client-B", "client-C"}

	for _, client := range clients {
		fmt.Printf("   Client %s:\n", client)

		// Each client can make 10 requests initially
		for i := 1; i <= 12; i++ {
			allowed := limiter.Allow(client)
			remaining := limiter.Remaining(client)

			if allowed {
				fmt.Printf("      Request %2d: ✅ ALLOWED (remaining: %2d)\n", i, remaining)
			} else {
				fmt.Printf("      Request %2d: ❌ DENIED  (remaining: %2d)\n", i, remaining)
			}
		}
		fmt.Println()
	}
}

func exampleDifferentClients() {
	fmt.Println("3. Different Rate Limits per Client Type")
	fmt.Println("   -------------------------------------")

	// Premium clients get higher limits
	premiumLimiter := ratelimit.NewInMemoryLimiter(20.0, 50.0)
	freeLimiter := ratelimit.NewInMemoryLimiter(2.0, 5.0)

	fmt.Println("   Premium client (20 req/sec, burst 50):")
	for i := 1; i <= 55; i++ {
		if premiumLimiter.Allow("premium-user") {
			if i%10 == 0 {
				fmt.Printf("      Request %d: ✅ ALLOWED\n", i)
			}
		} else {
			fmt.Printf("      Request %d: ❌ DENIED (limit reached)\n", i)
			break
		}
	}

	fmt.Println("\n   Free client (2 req/sec, burst 5):")
	for i := 1; i <= 7; i++ {
		if freeLimiter.Allow("free-user") {
			fmt.Printf("      Request %d: ✅ ALLOWED\n", i)
		} else {
			fmt.Printf("      Request %d: ❌ DENIED (limit reached)\n", i)
			break
		}
	}

	fmt.Println()
}

func exampleHTTPMiddleware() {
	fmt.Println("4. HTTP Middleware Integration")
	fmt.Println("   ---------------------------")

	// Create rate limiter for API endpoints
	apiLimiter := ratelimit.NewInMemoryLimiter(10.0, 30.0)

	// Simulated HTTP handler with rate limiting
	rateLimitHandler := func(w http.ResponseWriter, r *http.Request, clientIP string) {
		if !apiLimiter.Allow(clientIP) {
			w.WriteHeader(http.StatusTooManyRequests)
			fmt.Fprintf(w, `{"error":"rate limit exceeded"}`)
			return
		}

		w.WriteHeader(http.StatusOK)
		fmt.Fprintf(w, `{"message":"request successful","remaining":%d}`,
			apiLimiter.Remaining(clientIP))
	}

	// Simulate requests from different IPs
	clientIPs := []string{"192.168.1.100", "192.168.1.101", "192.168.1.102"}

	fmt.Println("   Simulating 35 requests from 3 different IPs:")
	fmt.Println()

	for i := 1; i <= 35; i++ {
		clientIP := clientIPs[i%len(clientIPs)]

		// Create mock response writer
		mockWriter := &mockResponseWriter{headers: make(http.Header)}
		mockRequest, _ := http.NewRequest("GET", "/api/users", nil)

		rateLimitHandler(mockWriter, mockRequest, clientIP)

		if i%5 == 0 {
			fmt.Printf("   Request %2d from %s: status=%d\n",
				i, clientIP, mockWriter.status)
		}
	}

	fmt.Println()
}

func exampleBatchOperations() {
	fmt.Println("5. Batch Operations (Weighted Requests)")
	fmt.Println("   ------------------------------------")

	config := ratelimit.Configuration{
		RefillRatePerSecond: 10.0,
		BucketCapacity:      100.0,
		InitialTokens:       100.0,
	}

	limiter := ratelimit.NewInMemoryLimiterWithConfig(config)
	client := "batch-client"

	operations := []struct {
		name   string
		weight int
	}{
		{"Simple GET", 1},
		{"Complex Query", 5},
		{"Data Export", 20},
		{"Batch Upload", 50},
		{"Full Sync", 100},
	}

	fmt.Println("   Weighted operations (simple=1, complex=5, export=20, batch=50, sync=100):")
	fmt.Printf("   Starting tokens: %d\n", limiter.Remaining(client))
	fmt.Println()

	for _, op := range operations {
		allowed := limiter.AllowN(client, op.weight)
		remaining := limiter.Remaining(client)

		if allowed {
			fmt.Printf("   ✅ %s (weight=%d) - remaining: %d\n",
				op.name, op.weight, remaining)
		} else {
			fmt.Printf("   ❌ %s (weight=%d) - DENIED (need %d, have %d)\n",
				op.name, op.weight, op.weight, remaining)
			break
		}
	}

	fmt.Println()
}

func exampleRateLimitHeaders() {
	fmt.Println("6. Rate Limit Headers for API Responses")
	fmt.Println("   ------------------------------------")

	limiter := ratelimit.NewInMemoryLimiter(5.0, 10.0)
	client := "api-client"

	fmt.Println("   Standard rate limit headers to return to clients:")
	fmt.Println("   - X-RateLimit-Limit: Maximum requests per window")
	fmt.Println("   - X-RateLimit-Remaining: Remaining requests")
	fmt.Println("   - X-RateLimit-Reset: Time until limit resets")
	fmt.Println()

	for i := 1; i <= 12; i++ {
		allowed := limiter.Allow(client)
		remaining := limiter.Remaining(client)

		limit := 10
		resetSeconds := 2 // Simplified

		if allowed {
			fmt.Printf("   Request %2d: status=200, limit=%d, remaining=%d, reset=%ds\n",
				i, limit, remaining, resetSeconds)
		} else {
			fmt.Printf("   Request %2d: status=429, limit=%d, remaining=%d, reset=%ds\n",
				i, limit, remaining, resetSeconds)

			if i == 11 {
				fmt.Println("\n   ⚠ Rate limit exceeded - client should backoff")
				fmt.Println("   Client should retry after reset period")
			}
		}
	}

	fmt.Println()
}

func exampleCustomConfiguration() {
	fmt.Println("7. Custom Configuration for Different Use Cases")
	fmt.Println("   --------------------------------------------")

	useCases := []struct {
		name        string
		rate        float64
		capacity    float64
		description string
	}{
		{"Login Endpoint", 1.0, 3.0, "Prevent brute force attacks"},
		{"Public API", 10.0, 30.0, "Standard rate limit for public endpoints"},
		{"Internal Service", 100.0, 200.0, "High throughput for service-to-service"},
		{"Admin Dashboard", 2.0, 5.0, "Lower limit for UI interactions"},
	}

	for _, useCase := range useCases {
		config := ratelimit.Configuration{
			RefillRatePerSecond: useCase.rate,
			BucketCapacity:      useCase.capacity,
			InitialTokens:       useCase.capacity,
		}

		limiter := ratelimit.NewInMemoryLimiterWithConfig(config)
		fmt.Printf("   %s:\n", useCase.name)
		fmt.Printf("      Rate: %.1f req/sec, Burst: %.0f\n", useCase.rate, useCase.capacity)
		fmt.Printf("      Purpose: %s\n", useCase.description)

		// Test the limiter
		testKey := "test-" + useCase.name
		for i := 1; i <= int(useCase.capacity)+1; i++ {
			if limiter.Allow(testKey) {
				if i == int(useCase.capacity)+1 {
					fmt.Printf("      Request %d: Denied (as expected)\n", i)
				}
			}
		}
		fmt.Println()
	}
}

func exampleAPIProtection() {
	fmt.Println("8. Real-World API Protection")
	fmt.Println("   -------------------------")

	// Different rate limits for different endpoints
	authLimiter := ratelimit.NewInMemoryLimiter(1.0, 3.0)  // Login: 3 attempts
	apiLimiter := ratelimit.NewInMemoryLimiter(10.0, 30.0) // API: 30 requests
	adminLimiter := ratelimit.NewInMemoryLimiter(2.0, 5.0) // Admin: 5 requests

	clientIP := "203.0.113.42"

	fmt.Println("   Protecting different API endpoints:")
	fmt.Println()

	// Simulate authentication attempts
	fmt.Println("   Authentication endpoint (/auth/login):")
	for i := 1; i <= 5; i++ {
		if authLimiter.Allow(clientIP) {
			fmt.Printf("      Attempt %d: ✅ allowed\n", i)
		} else {
			fmt.Printf("      Attempt %d: ❌ blocked - too many failures\n", i)
			break
		}
	}

	fmt.Println("\n   API endpoint (/api/data):")
	for i := 1; i <= 35; i++ {
		if apiLimiter.Allow(clientIP) {
			if i%10 == 0 {
				fmt.Printf("      Request %d: ✅ allowed\n", i)
			}
		} else {
			fmt.Printf("      Request %d: ❌ rate limited\n", i)
			break
		}
	}

	fmt.Println("\n   Admin endpoint (/admin/config):")
	for i := 1; i <= 7; i++ {
		if adminLimiter.Allow(clientIP) {
			fmt.Printf("      Request %d: ✅ allowed\n", i)
		} else {
			fmt.Printf("      Request %d: ❌ admin rate limit exceeded\n", i)
			break
		}
	}

	fmt.Println()
}

// mockResponseWriter implements http.ResponseWriter for testing
type mockResponseWriter struct {
	headers http.Header
	status  int
	body    []byte
}

func (writer *mockResponseWriter) Header() http.Header {
	if writer.headers == nil {
		writer.headers = make(http.Header)
	}
	return writer.headers
}

func (writer *mockResponseWriter) Write(data []byte) (int, error) {
	writer.body = append(writer.body, data...)
	return len(data), nil
}

func (writer *mockResponseWriter) WriteHeader(statusCode int) {
	writer.status = statusCode
}

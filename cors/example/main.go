package main

import (
	"fmt"
	"net/http"
	"time"

	"github.com/migambile25/go-repository/cors"
)

func main() {
	fmt.Println("=== CORS Middleware Examples ===")

	// Example 1: Production configuration
	exampleProductionConfiguration()

	// Example 2: Development configuration
	exampleDevelopmentConfiguration()

	// Example 3: Custom configuration
	exampleCustomConfiguration()

	// Example 4: Helper function usage
	exampleHelperFunction()

	// Example 5: Complete server setup
	exampleCompleteServer()
}

func exampleProductionConfiguration() {
	fmt.Println("1. Production Configuration (Secure)")
	fmt.Println("   --------------------------------")

	// Create secure CORS configuration for production
	corsConfig := cors.Configuration{
		AllowOrigins: []string{
			"https://example.com",
			"https://app.example.com",
			"https://api.example.com",
		},
		AllowMethods: []string{
			"GET", "POST", "PUT", "DELETE", "OPTIONS",
		},
		AllowHeaders: []string{
			"Accept",
			"Content-Type",
			"Content-Length",
			"Accept-Encoding",
			"Authorization",
			"X-CSRF-Token",
		},
		ExposeHeaders: []string{
			"Content-Length",
			"X-RateLimit-Limit",
			"X-RateLimit-Remaining",
		},
		AllowCredentials: true,
		MaxAge:           86400, // 24 hours
		AllowAllOrigins:  false,
	}

	// corsMiddleware := cors.Middleware(corsConfig)

	// Create a simple API handler
	// apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 	w.Header().Set("Content-Type", "application/json")
	// 	w.WriteHeader(http.StatusOK)
	// 	w.Write([]byte(`{"message":"API response"}`))
	// })

	// Wrap with CORS middleware
	// protectedHandler := corsMiddleware(apiHandler)

	fmt.Println("   Production CORS middleware configured")
	fmt.Printf("   Allowed origins: %v\n", corsConfig.AllowOrigins)
	fmt.Printf("   Credentials allowed: %v\n", corsConfig.AllowCredentials)
	fmt.Println()
}

func exampleDevelopmentConfiguration() {
	fmt.Println("2. Development Configuration (Permissive)")
	fmt.Println("   -------------------------------------")

	// Development configuration - allows any origin
	devConfig := cors.DevelopmentConfiguration()
	corsMiddleware := cors.Middleware(devConfig)

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("Development endpoint"))
	})

	wrappedHandler := corsMiddleware(testHandler)

	// Simulate request from any origin
	req, _ := http.NewRequest("GET", "/debug", nil)
	req.Header.Set("Origin", "http://localhost:3000")

	response := &mockResponseWriter{headers: make(http.Header)}
	wrappedHandler.ServeHTTP(response, req)

	fmt.Println("   Development CORS middleware configured")
	fmt.Printf("   Allow all origins: %v\n", devConfig.AllowAllOrigins)
	fmt.Printf("   Access-Control-Allow-Origin header: %s\n",
		response.headers.Get("Access-Control-Allow-Origin"))
	fmt.Println()
}

func exampleCustomConfiguration() {
	fmt.Println("3. Custom Configuration (Public API)")
	fmt.Println("   --------------------------------")

	// Configuration for a public API that doesn't need credentials
	publicAPIConfig := cors.Configuration{
		AllowOrigins:     []string{"*"}, // Allow any origin (public API)
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE"},
		AllowHeaders:     []string{"Accept", "Content-Type", "Authorization"},
		ExposeHeaders:    []string{"X-RateLimit-Limit", "X-RateLimit-Remaining"},
		AllowCredentials: false, // Don't allow credentials for public API
		MaxAge:           3600,  // 1 hour
		AllowAllOrigins:  true,
	}

	// corsMiddleware := cors.Middleware(publicAPIConfig)

	fmt.Println("   Public API CORS configuration")
	fmt.Printf("   Allow any origin: %v\n", publicAPIConfig.AllowAllOrigins)
	fmt.Printf("   Credentials allowed: %v\n", publicAPIConfig.AllowCredentials)
	fmt.Println()
}

func exampleHelperFunction() {
	fmt.Println("4. Helper Function (Quick Setup)")
	fmt.Println("   -----------------------------")

	// Quick setup with helper function
	allowedOrigins := []string{
		"https://myapp.com",
		"https://dashboard.myapp.com",
	}

	corsMiddleware := cors.New(allowedOrigins, true)

	apiHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`{"status":"ok"}`))
	})

	protectedAPI := corsMiddleware(apiHandler)

	fmt.Println("   Quick CORS setup with helper function")
	fmt.Printf("   Allowed origins: %v\n", allowedOrigins)
	fmt.Printf("   Middleware created: %v\n", protectedAPI != nil)
	fmt.Println()
}

func exampleCompleteServer() {
	fmt.Println("5. Complete Server Example")
	fmt.Println("   -----------------------")

	// Create router
	router := http.NewServeMux()

	// Define endpoints
	router.HandleFunc("GET /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"users":["alice","bob"]}`))
	})

	router.HandleFunc("POST /api/users", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"message":"User created"}`))
	})

	router.HandleFunc("OPTIONS /api/users", func(w http.ResponseWriter, r *http.Request) {
		// OPTIONS requests are handled by CORS middleware
		w.WriteHeader(http.StatusNoContent)
	})

	// Configure CORS for production
	corsConfig := cors.Configuration{
		AllowOrigins: []string{
			"https://example.com",
			"https://staging.example.com",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"Accept", "Content-Type", "Authorization", "X-Request-ID"},
		ExposeHeaders:    []string{"X-Request-ID", "X-Trace-ID"},
		AllowCredentials: true,
		MaxAge:           86400,
		AllowAllOrigins:  false,
	}

	// Apply CORS middleware
	corsMiddleware := cors.Middleware(corsConfig)
	handler := corsMiddleware(router)

	// Create server
	server := &http.Server{
		Addr:         ":8080",
		Handler:      handler,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	fmt.Println("   Complete server configured with CORS")
	fmt.Printf("   Server address: %s\n", server.Addr)
	fmt.Printf("   Allowed origins: %v\n", corsConfig.AllowOrigins)
	fmt.Println("\n   To run the server: go run main.go")
	fmt.Println("   Then test with: curl -H 'Origin: https://example.com' http://localhost:8080/api/users")
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

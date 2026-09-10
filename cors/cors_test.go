package cors

import (
    "net/http"
    "net/http/httptest"
    "testing"
)

func TestMiddlewareNoOrigin(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"https://example.com"}

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    request := httptest.NewRequest("GET", "/test", nil)
    responseRecorder := httptest.NewRecorder()

    handler.ServeHTTP(responseRecorder, request)

    // No CORS headers should be set for non-CORS requests
    if responseRecorder.Header().Get("Access-Control-Allow-Origin") != "" {
        t.Errorf("Expected no Access-Control-Allow-Origin header for non-CORS request")
    }
}

func TestMiddlewareAllowedOrigin(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"https://example.com", "https://app.example.com"}

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    request := httptest.NewRequest("GET", "/test", nil)
    request.Header.Set("Origin", "https://example.com")
    responseRecorder := httptest.NewRecorder()

    handler.ServeHTTP(responseRecorder, request)

    allowedOrigin := responseRecorder.Header().Get("Access-Control-Allow-Origin")
    if allowedOrigin != "https://example.com" {
        t.Errorf("Expected Access-Control-Allow-Origin 'https://example.com', got '%s'", allowedOrigin)
    }
}

func TestMiddlewareDisallowedOrigin(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"https://example.com"}

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    request := httptest.NewRequest("GET", "/test", nil)
    request.Header.Set("Origin", "https://malicious.com")
    responseRecorder := httptest.NewRecorder()

    handler.ServeHTTP(responseRecorder, request)

    // Disallowed origin should not receive the header
    allowedOrigin := responseRecorder.Header().Get("Access-Control-Allow-Origin")
    if allowedOrigin != "" {
        t.Errorf("Expected no Access-Control-Allow-Origin for disallowed origin, got '%s'", allowedOrigin)
    }
}

func TestPreflightRequest(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"https://example.com"}
    configuration.AllowMethods = []string{"GET", "POST", "PUT"}
    configuration.AllowHeaders = []string{"Content-Type", "Authorization"}
    configuration.MaxAge = 3600

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        t.Errorf("Handler should not be called for preflight request")
    }))

    request := httptest.NewRequest("OPTIONS", "/test", nil)
    request.Header.Set("Origin", "https://example.com")
    request.Header.Set("Access-Control-Request-Method", "POST")
    responseRecorder := httptest.NewRecorder()

    handler.ServeHTTP(responseRecorder, request)

    // Check preflight response headers
    if responseRecorder.Header().Get("Access-Control-Allow-Origin") != "https://example.com" {
        t.Errorf("Missing Access-Control-Allow-Origin header")
    }

    if responseRecorder.Header().Get("Access-Control-Allow-Methods") != "GET, POST, PUT" {
        t.Errorf("Incorrect Access-Control-Allow-Methods header")
    }

    if responseRecorder.Header().Get("Access-Control-Allow-Headers") != "Content-Type, Authorization" {
        t.Errorf("Incorrect Access-Control-Allow-Headers header")
    }

    if responseRecorder.Header().Get("Access-Control-Max-Age") != "3600" {
        t.Errorf("Incorrect Access-Control-Max-Age header")
    }

    if responseRecorder.Code != http.StatusNoContent {
        t.Errorf("Expected status 204 No Content, got %d", responseRecorder.Code)
    }
}

func TestAllowAllOriginsDevelopment(t *testing.T) {
    configuration := DevelopmentConfiguration()

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    testOrigins := []string{
        "http://localhost:3000",
        "https://example.com",
        "https://any-domain.com",
    }

    for _, origin := range testOrigins {
        request := httptest.NewRequest("GET", "/test", nil)
        request.Header.Set("Origin", origin)
        responseRecorder := httptest.NewRecorder()

        handler.ServeHTTP(responseRecorder, request)

        allowedOrigin := responseRecorder.Header().Get("Access-Control-Allow-Origin")
        if allowedOrigin != origin {
            t.Errorf("For origin %s, expected '%s', got '%s'", origin, origin, allowedOrigin)
        }
    }
}

func TestAllowCredentials(t *testing.T) {
    tests := []struct {
        name              string
        allowCredentials  bool
        expectCredentials bool
    }{
        {
            name:              "credentials enabled",
            allowCredentials:  true,
            expectCredentials: true,
        },
        {
            name:              "credentials disabled",
            allowCredentials:  false,
            expectCredentials: false,
        },
    }

    for _, testCase := range tests {
        t.Run(testCase.name, func(t *testing.T) {
            configuration := DefaultConfiguration()
            configuration.AllowOrigins = []string{"https://example.com"}
            configuration.AllowCredentials = testCase.allowCredentials

            handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
                w.WriteHeader(http.StatusOK)
            }))

            request := httptest.NewRequest("GET", "/test", nil)
            request.Header.Set("Origin", "https://example.com")
            responseRecorder := httptest.NewRecorder()

            handler.ServeHTTP(responseRecorder, request)

            credentialsHeader := responseRecorder.Header().Get("Access-Control-Allow-Credentials")

            if testCase.expectCredentials && credentialsHeader != "true" {
                t.Errorf("Expected Access-Control-Allow-Credentials: true, got '%s'", credentialsHeader)
            }

            if !testCase.expectCredentials && credentialsHeader != "" {
                t.Errorf("Expected no Access-Control-Allow-Credentials header, got '%s'", credentialsHeader)
            }
        })
    }
}

func TestWildcardSubdomain(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"*.example.com", "https://api.example.com"}

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    testCases := []struct {
        origin        string
        shouldBeAllowed bool
    }{
        {origin: "https://app.example.com", shouldBeAllowed: true},
        {origin: "https://admin.example.com", shouldBeAllowed: true},
        {origin: "https://api.example.com", shouldBeAllowed: true},
        {origin: "https://example.com", shouldBeAllowed: false},
        {origin: "https://malicious.com", shouldBeAllowed: false},
    }

    for _, testCase := range testCases {
        request := httptest.NewRequest("GET", "/test", nil)
        request.Header.Set("Origin", testCase.origin)
        responseRecorder := httptest.NewRecorder()

        handler.ServeHTTP(responseRecorder, request)

        allowedOrigin := responseRecorder.Header().Get("Access-Control-Allow-Origin")

        if testCase.shouldBeAllowed && allowedOrigin != testCase.origin {
            t.Errorf("Origin %s should be allowed, but got header '%s'", testCase.origin, allowedOrigin)
        }

        if !testCase.shouldBeAllowed && allowedOrigin != "" {
            t.Errorf("Origin %s should not be allowed, but got header '%s'", testCase.origin, allowedOrigin)
        }
    }
}

func TestExposeHeaders(t *testing.T) {
    configuration := DefaultConfiguration()
    configuration.AllowOrigins = []string{"https://example.com"}
    configuration.ExposeHeaders = []string{"X-Custom-Header", "X-Request-ID"}

    handler := Middleware(configuration)(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    request := httptest.NewRequest("GET", "/test", nil)
    request.Header.Set("Origin", "https://example.com")
    responseRecorder := httptest.NewRecorder()

    handler.ServeHTTP(responseRecorder, request)

    exposeHeaders := responseRecorder.Header().Get("Access-Control-Expose-Headers")
    expected := "X-Custom-Header, X-Request-ID"

    if exposeHeaders != expected {
        t.Errorf("Expected Access-Control-Expose-Headers '%s', got '%s'", expected, exposeHeaders)
    }
}

func TestNewHelperFunction(t *testing.T) {
    allowedOrigins := []string{"https://example.com", "https://app.example.com"}
    allowCredentials := true

    corsMiddleware := New(allowedOrigins, allowCredentials)

    if corsMiddleware == nil {
        t.Errorf("Expected non-nil middleware")
    }

    testHandler := corsMiddleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusOK)
    }))

    request := httptest.NewRequest("GET", "/test", nil)
    request.Header.Set("Origin", "https://example.com")
    responseRecorder := httptest.NewRecorder()

    testHandler.ServeHTTP(responseRecorder, request)

    allowedOrigin := responseRecorder.Header().Get("Access-Control-Allow-Origin")
    if allowedOrigin != "https://example.com" {
        t.Errorf("Expected Access-Control-Allow-Origin 'https://example.com', got '%s'", allowedOrigin)
    }

    credentialsHeader := responseRecorder.Header().Get("Access-Control-Allow-Credentials")
    if credentialsHeader != "true" {
        t.Errorf("Expected Access-Control-Allow-Credentials 'true', got '%s'", credentialsHeader)
    }
}
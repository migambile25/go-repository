// Package cors provides configurable Cross-Origin Resource Sharing middleware.
// This implementation follows security best practices and is suitable for
// production deployments requiring browser-based cross-origin requests.
//
// The middleware handles preflight OPTIONS requests automatically and
// applies appropriate CORS headers based on the configuration.
package cors

import (
	"net/http"
	"strconv"
	"strings"
)

// Configuration holds all CORS settings for the middleware.
// Each field has a safe default for production use.
type Configuration struct {
	// AllowOrigins lists the origins that can access the resource.
	// Use specific domains in production, never "*" with credentials.
	AllowOrigins []string

	// AllowMethods lists the HTTP methods allowed for cross-origin requests.
	AllowMethods []string

	// AllowHeaders lists the headers that can be used in actual requests.
	AllowHeaders []string

	// ExposeHeaders lists the headers that browsers are allowed to access.
	ExposeHeaders []string

	// AllowCredentials indicates whether the response can include credentials.
	AllowCredentials bool

	// MaxAge indicates how long (in seconds) the results of a preflight request
	// can be cached. Default: 86400 (24 hours)
	MaxAge int

	// AllowAllOrigins is a convenience flag for development only.
	// Never use in production - set false and use AllowOrigins instead.
	AllowAllOrigins bool
}

// DefaultConfiguration returns a secure production-ready configuration.
// This configuration is restrictive and safe for most APIs.
func DefaultConfiguration() Configuration {
	return Configuration{
		AllowOrigins:     []string{},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"Accept", "Content-Type", "Content-Length", "Accept-Encoding", "Authorization", "X-Request-ID", "X-Trace-ID"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "X-Request-ID", "X-Trace-ID"},
		AllowCredentials: false,
		MaxAge:           86400,
		AllowAllOrigins:  false,
	}
}

// DevelopmentConfiguration returns a permissive configuration for local development.
// Warning: Do not use this in production as it allows any origin.
func DevelopmentConfiguration() Configuration {
	return Configuration{
		AllowOrigins:     []string{},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"*"},
		AllowCredentials: true,
		MaxAge:           86400,
		AllowAllOrigins:  true,
	}
}

// Middleware returns an HTTP handler that applies CORS headers to responses.
// It automatically handles preflight OPTIONS requests.
func Middleware(configuration Configuration) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(responseWriter http.ResponseWriter, httpRequest *http.Request) {
			// Get the origin from the request header
			requestOrigin := httpRequest.Header.Get("Origin")

			// Determine which origins are allowed
			allowedOrigin := determineAllowedOrigin(requestOrigin, configuration)

			// Set CORS headers on the response
			if allowedOrigin != "" {
				responseWriter.Header().Set("Access-Control-Allow-Origin", allowedOrigin)
			}

			// Set credentials header if enabled
			if configuration.AllowCredentials && allowedOrigin != "" {
				responseWriter.Header().Set("Access-Control-Allow-Credentials", "true")
			}

			// Set exposed headers
			if len(configuration.ExposeHeaders) > 0 {
				responseWriter.Header().Set("Access-Control-Expose-Headers", strings.Join(configuration.ExposeHeaders, ", "))
			}

			// Handle preflight OPTIONS request
			if httpRequest.Method == "OPTIONS" {
				// Set allowed methods
				if len(configuration.AllowMethods) > 0 {
					responseWriter.Header().Set("Access-Control-Allow-Methods", strings.Join(configuration.AllowMethods, ", "))
				}

				// Set allowed headers
				if len(configuration.AllowHeaders) > 0 {
					responseWriter.Header().Set("Access-Control-Allow-Headers", strings.Join(configuration.AllowHeaders, ", "))
				}

				// Set max age for preflight cache
				if configuration.MaxAge > 0 {
					responseWriter.Header().Set("Access-Control-Max-Age", strconv.Itoa(configuration.MaxAge))
				}

				// Preflight requests should not have a body
				responseWriter.WriteHeader(http.StatusNoContent)
				return
			}

			// Pass through to the next handler for non-preflight requests
			next.ServeHTTP(responseWriter, httpRequest)
		})
	}
}

// determineAllowedOrigin checks if the request origin is allowed and returns it
func determineAllowedOrigin(requestOrigin string, configuration Configuration) string {
	// Handle empty origin (not a CORS request)
	if requestOrigin == "" {
		return ""
	}

	// Development mode - allow all origins
	if configuration.AllowAllOrigins {
		return requestOrigin
	}

	// Check against allowed origins list
	for _, allowedOrigin := range configuration.AllowOrigins {
		if allowedOrigin == "*" {
			return requestOrigin
		}

		if allowedOrigin == requestOrigin {
			return requestOrigin
		}

		// Check for wildcard subdomain matching (*.example.com)
		if strings.HasPrefix(allowedOrigin, "*.") {
			domainSuffix := allowedOrigin[1:] // Remove the *
			if strings.HasSuffix(requestOrigin, domainSuffix) {
				return requestOrigin
			}
		}
	}

	// Origin not allowed
	return ""
}

// New creates a CORS middleware with the default production configuration.
// This is the recommended way to create the middleware for production use.
func New(allowedOrigins []string, allowCredentials bool) func(http.Handler) http.Handler {
	configuration := DefaultConfiguration()
	configuration.AllowOrigins = allowedOrigins
	configuration.AllowCredentials = allowCredentials
	configuration.AllowAllOrigins = false

	return Middleware(configuration)
}

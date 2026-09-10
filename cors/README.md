# cors

A Go middleware package for configurable Cross-Origin Resource Sharing (CORS). Follows security best practices and is suitable for production deployments requiring browser-based cross-origin requests. Preflight `OPTIONS` requests are handled automatically.

## Installation

```bash
go get github.com/hekimapro/cors
```

## Quick Start

```go
import "github.com/hekimapro/cors"

mux := http.NewServeMux()
mux.HandleFunc("/users", getUsers)

handler := cors.New(
    []string{"https://app.example.com", "https://admin.example.com"},
    false, // allowCredentials
)

http.ListenAndServe(":8080", handler(mux))
```

## Configuration Presets

### Production (recommended)

`DefaultConfiguration` is restrictive and safe. You supply the allowed origins and credentials flag; everything else is pre-configured sensibly.

```go
cfg := cors.DefaultConfiguration()
cfg.AllowOrigins = []string{"https://app.example.com"}
cfg.AllowCredentials = true

handler := cors.Middleware(cfg)
```

Default values:

| Setting | Value |
|---|---|
| `AllowMethods` | `GET, POST, PUT, DELETE, OPTIONS, PATCH` |
| `AllowHeaders` | `Accept, Content-Type, Content-Length, Accept-Encoding, Authorization, X-Request-ID, X-Trace-ID` |
| `ExposeHeaders` | `Content-Length, Content-Type, X-Request-ID, X-Trace-ID` |
| `AllowCredentials` | `false` |
| `MaxAge` | `86400` (24 hours) |
| `AllowAllOrigins` | `false` |

### Development

`DevelopmentConfiguration` allows any origin and all headers. **Never use in production.**

```go
cfg := cors.DevelopmentConfiguration()
handler := cors.Middleware(cfg)
```

## Full Configuration

For complete control, build a `Configuration` struct directly:

```go
handler := cors.Middleware(cors.Configuration{
    AllowOrigins:     []string{"https://app.example.com", "*.internal.example.com"},
    AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "OPTIONS", "PATCH"},
    AllowHeaders:     []string{"Authorization", "Content-Type", "X-Request-ID"},
    ExposeHeaders:    []string{"Content-Length", "X-Request-ID"},
    AllowCredentials: true,
    MaxAge:           3600,
    AllowAllOrigins:  false,
})
```

### Origin Matching

`AllowOrigins` supports three formats:

| Format | Example | Behaviour |
|---|---|---|
| Exact domain | `https://app.example.com` | Matches only that origin |
| Wildcard subdomain | `*.example.com` | Matches any subdomain of `example.com` |
| Open wildcard | `*` | Matches any origin (avoid with credentials) |

> **Security note:** Never combine `AllowCredentials: true` with `*` or `AllowAllOrigins: true`. Browsers will reject such responses, and it exposes your API to CSRF attacks.

## Configuration Reference

### `Configuration`

| Field | Type | Description |
|---|---|---|
| `AllowOrigins` | `[]string` | Origins permitted to access the resource |
| `AllowMethods` | `[]string` | HTTP methods allowed for cross-origin requests |
| `AllowHeaders` | `[]string` | Headers the client may send |
| `ExposeHeaders` | `[]string` | Headers the browser is allowed to read from the response |
| `AllowCredentials` | `bool` | Whether cookies and auth headers are included |
| `MaxAge` | `int` | Seconds a preflight response may be cached |
| `AllowAllOrigins` | `bool` | Development shortcut — allows any origin |

## Framework Examples

### Standard `net/http`

```go
mux := http.NewServeMux()
mux.HandleFunc("GET /users", getUsers)

corsHandler := cors.New([]string{"https://app.example.com"}, false)
http.ListenAndServe(":8080", corsHandler(mux))
```

### With Middleware Chaining

```go
func chain(h http.Handler, middlewares ...func(http.Handler) http.Handler) http.Handler {
    for i := len(middlewares) - 1; i >= 0; i-- {
        h = middlewares[i](h)
    }
    return h
}

handler := chain(
    mux,
    cors.New([]string{"https://app.example.com"}, true),
    logging.Middleware,
    auth.Middleware,
)

http.ListenAndServe(":8080", handler)
```

## How Preflight Works

When a browser sends a cross-origin request with custom headers or methods, it first issues a preflight `OPTIONS` request. The middleware intercepts it, responds with the appropriate `Access-Control-*` headers, and returns `204 No Content` — the actual handler is never called. Subsequent real requests receive the `Allow-Origin` and `Expose-Headers` headers on every response.

## License

See [LICENSE](LICENSE).

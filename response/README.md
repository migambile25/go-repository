# response

A Go package for standardized HTTP JSON responses. Every handler in your API uses the same response shape, status codes, and error structure — no inconsistencies across endpoints.

## How It Works

Every response, success or failure, is wrapped in a `StandardResponse`:

```json
{
  "success": true,
  "status_code": 200,
  "status_text": "OK",
  "message": "User fetched successfully",
  "data": { ... },
  "error": null
}
```

On failure:

```json
{
  "success": false,
  "status_code": 422,
  "status_text": "Unprocessable Entity",
  "message": "Validation failed",
  "error": "invalid email"
}
```

## Installation

```bash
go get github.com/hekimapro/response
```

## Usage

### Success Responses

```go
import "github.com/hekimapro/response"

// 200 OK
response.Success(w, "User fetched successfully", user)

// 201 Created
response.Created(w, "User created successfully", user)

// 202 Accepted — for async operations
response.Accepted(w, "Export job queued", jobID)

// 204 No Content — no body is written
response.NoContent(w)
```

### Paginated Responses

Use `WithMetadata` to attach pagination info to any list response:

```go
response.WithMetadata(w, "Users fetched successfully", users, &response.Metadata{
    Page:       1,
    PerPage:    20,
    Total:      200,
    TotalPages: 10,
})
```

### Error Responses

Every error helper accepts a human-readable `message` and a machine-readable `errorCode`:

```go
// 400 Bad Request
response.BadRequest(w, "Request body is malformed", "INVALID_BODY")

// 401 Unauthorized
response.Unauthorized(w, "Token is missing or expired", "TOKEN_INVALID")

// 403 Forbidden
response.Forbidden(w, "You do not have access to this resource", "ACCESS_DENIED")

// 404 Not Found
response.NotFound(w, "User not found", "USER_NOT_FOUND")

// 409 Conflict
response.Conflict(w, "Email address already in use", "EMAIL_TAKEN")

// 422 Validation Error — includes the offending field
response.ValidationError(w, "email", "Email address is not valid", "INVALID_EMAIL")

// 429 Too Many Requests
response.TooManyRequests(w, "Slow down, too many requests", "RATE_LIMITED")

// 500 Internal Server Error
response.InternalServerError(w, "Something went wrong", "INTERNAL_ERROR")

// 503 Service Unavailable
response.ServiceUnavailable(w, "Service is temporarily unavailable", "SERVICE_DOWN")
```

### Low-Level Helper

If the built-in helpers do not cover your status code, use `WriteJSON` directly:

```go
response.WriteJSON(w, http.StatusTeapot, response.StandardResponse{
    Success: true,
    Message: "I am a teapot",
})
```

`status_code` and `status_text` are always overwritten to match the actual status code passed in, so you never need to set them manually.

## Response Structure Reference

### `StandardResponse`

| Field | Type | Description |
|---|---|---|
| `success` | `bool` | `true` for 2xx responses, `false` for errors |
| `status_code` | `int` | HTTP status code |
| `status_text` | `string` | Human-readable HTTP status (e.g. `"Not Found"`) |
| `message` | `string` | Contextual message about the operation |
| `data` | `any` | Response payload — omitted on failure |
| `error` | `ErrorDetail` | Error details — omitted on success |
| `meta` | `Metadata` | Pagination info — omitted when not applicable |

### `ErrorDetail`

| Field | Type | Description |
|---|---|---|
| `code` | `string` | Machine-readable error code (e.g. `"USER_NOT_FOUND"`) |
| `message` | `string` | Human-readable description of the error |
| `field` | `string` | The field that caused the error — omitted when not applicable |

### `Metadata`

| Field | Type | Description |
|---|---|---|
| `page` | `int` | Current page number |
| `per_page` | `int` | Items per page |
| `total` | `int` | Total number of items |
| `total_pages` | `int` | Total number of pages |

## Full Example

```go
package main

import (
    "net/http"
    "github.com/hekimapro/response"
)

type User struct {
    ID    int    `json:"id"`
    Name  string `json:"name"`
    Email string `json:"email"`
}

func getUser(w http.ResponseWriter, r *http.Request) {
    id := r.PathValue("id")
    if id == "" {
        response.BadRequest(w, "Missing user ID", "MISSING_ID")
        return
    }

    user, err := db.FindUser(id)
    if err != nil {
        response.NotFound(w, "User not found", "USER_NOT_FOUND")
        return
    }

    response.Success(w, "User fetched successfully", user)
}
```

## License

See [LICENSE](LICENSE).

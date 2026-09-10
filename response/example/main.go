package main

import (
    "encoding/json"
    "net/http"
    "net/http/httptest"

    "github.com/migambile25/go-repository/response"
)

func main() {
    // This example shows how to use the response package in an HTTP handler

    // Create a handler that uses the response helpers
    handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
        switch r.URL.Path {
        case "/users":
            // Success response with data
            users := []map[string]interface{}{
                {"id": 1, "name": "Alice"},
                {"id": 2, "name": "Bob"},
            }
            response.Success(w, "Users retrieved successfully", users)

        case "/users/1":
            // Success response with single item
            user := map[string]interface{}{"id": 1, "name": "Alice", "email": "alice@example.com"}
            response.Success(w, "User found", user)

        case "/users/create":
            // Created response
            newUser := map[string]interface{}{"id": 123, "name": "Charlie"}
            response.Created(w, "User created successfully", newUser)

        case "/error/bad-request":
            // Bad request error
            response.BadRequest(w, "Invalid email format", "INVALID_EMAIL")

        case "/error/not-found":
            // Not found error
            response.NotFound(w, "User not found", "USER_NOT_FOUND")

        case "/error/validation":
            // Validation error with field
            response.ValidationError(w, "email", "Must be a valid email address", "VALIDATION_ERROR")
        default:
            response.NotFound(w, "Endpoint not found", "ENDPOINT_NOT_FOUND")
        }
    })

    // Test the responses
    testCases := []struct {
        name       string
        path       string
        expectedStatus int
    }{
        {"Success", "/users", 200},
        {"Created", "/users/create", 201},
        {"Bad Request", "/error/bad-request", 400},
        {"Not Found", "/error/not-found", 404},
        {"Validation Error", "/error/validation", 422},
        {"Paginated", "/users/paginated", 200},
    }

    println("=== Response Package Examples ===\n")

    for _, tc := range testCases {
        req := httptest.NewRequest("GET", tc.path, nil)
        rec := httptest.NewRecorder()

        handler.ServeHTTP(rec, req)

        println("Endpoint:", tc.path)
        println("  Status:", rec.Code, http.StatusText(rec.Code))

        var resp response.StandardResponse
        json.Unmarshal(rec.Body.Bytes(), &resp)

        println("  Success:", resp.Success)
        println("  Message:", resp.Message)

    }

    // Example output:
    // {
    //   "success": true,
    //   "status_code": 200,
    //   "status_text": "OK",
    //   "message": "Users retrieved successfully",
    //   "data": [...]
    // }
    //
    // {
    //   "success": false,
    //   "status_code": 400,
    //   "status_text": "Bad Request",
    //   "message": "Invalid email format",
    //   "error": {
    //     "code": "INVALID_EMAIL",
    //     "message": "Invalid email format"
    //   }
    // }
}
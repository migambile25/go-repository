// Package response provides standardized HTTP response helpers for JSON APIs.
// Every API response should use these helpers to ensure consistent formatting,
// proper status codes, and standardized error structures.
package response

import (
	"encoding/json"
	"net/http"
)

// StandardResponse is the base structure for all API responses.
type StandardResponse struct {
	Success    bool    `json:"success"`         // Operation success status
	StatusCode int     `json:"status_code"`     // HTTP status code
	StatusText string  `json:"status_text"`     // Human-readable status
	Message    string  `json:"message"`         // Operation message
	Error      *string `json:"error,omitempty"` // Error details (on failure)
	Data       any     `json:"data,omitempty"`  // Response payload (on success)
}

// WriteJSON writes a StandardResponse as JSON to the response writer.
func WriteJSON(responseWriter http.ResponseWriter, statusCode int, response StandardResponse) {
	responseWriter.Header().Set("Content-Type", "application/json")
	responseWriter.WriteHeader(statusCode)

	// Ensure status codes match
	response.StatusCode = statusCode
	response.StatusText = getStatusText(statusCode)

	if err := json.NewEncoder(responseWriter).Encode(response); err != nil {
		// Fallback if encoding fails
		http.Error(responseWriter, `{"success":false,"status_code":500,"status_text":"Internal Server Error","message":"Failed to encode response"}`, http.StatusInternalServerError)
	}
}

// Success writes a 200 OK success response.
func Success(responseWriter http.ResponseWriter, message string, data any) {
	WriteJSON(responseWriter, http.StatusOK, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Created writes a 201 Created success response.
func Created(responseWriter http.ResponseWriter, message string, data any) {
	WriteJSON(responseWriter, http.StatusCreated, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// Accepted writes a 202 Accepted response for async operations.
func Accepted(responseWriter http.ResponseWriter, message string, data any) {
	WriteJSON(responseWriter, http.StatusAccepted, StandardResponse{
		Success: true,
		Message: message,
		Data:    data,
	})
}

// NoContent writes a 204 No Content response with no body.
func NoContent(responseWriter http.ResponseWriter) {
	responseWriter.WriteHeader(http.StatusNoContent)
}

// BadRequest writes a 400 Bad Request error.
func BadRequest(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusBadRequest, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// ValidationError writes a 422 Unprocessable Entity with field-specific errors.
func ValidationError(responseWriter http.ResponseWriter, field string, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusUnprocessableEntity, StandardResponse{
		Success: false,
		Message: "Validation failed",
		Error:   &errorString,
	})
}

// Unauthorized writes a 401 Unauthorized error.
func Unauthorized(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusUnauthorized, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// Forbidden writes a 403 Forbidden error.
func Forbidden(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusForbidden, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// NotFound writes a 404 Not Found error.
func NotFound(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusNotFound, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// Conflict writes a 409 Conflict error (e.g., duplicate resource).
func Conflict(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusConflict, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// TooManyRequests writes a 429 Too Many Requests error.
func TooManyRequests(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusTooManyRequests, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// InternalServerError writes a 500 Internal Server Error.
func InternalServerError(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusInternalServerError, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// ServiceUnavailable writes a 503 Service Unavailable error.
func ServiceUnavailable(responseWriter http.ResponseWriter, message string, errorString string) {
	WriteJSON(responseWriter, http.StatusServiceUnavailable, StandardResponse{
		Success: false,
		Message: message,
		Error:   &errorString,
	})
}

// getStatusText returns human-readable text for HTTP status codes.
func getStatusText(statusCode int) string {
	return http.StatusText(statusCode)
}

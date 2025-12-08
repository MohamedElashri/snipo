package handlers

import (
	"encoding/json"
	"io"
	"net/http"
)

// MaxJSONBodySize is the maximum allowed size for JSON request bodies (2MB)
const MaxJSONBodySize = 2 * 1024 * 1024

// DecodeJSON safely decodes JSON from request body with size limit
func DecodeJSON(r *http.Request, v interface{}) error {
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(nil, r.Body, MaxJSONBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown fields for stricter validation

	if err := decoder.Decode(v); err != nil {
		return err
	}

	// Ensure only one JSON object in body
	if decoder.More() {
		return io.ErrUnexpectedEOF
	}

	return nil
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code    string            `json:"code"`
	Message string            `json:"message"`
	Details []ValidationError `json:"details,omitempty"`
}

// ValidationError represents a field validation error
type ValidationError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// JSON sends a JSON response
func JSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if data != nil {
		if err := json.NewEncoder(w).Encode(data); err != nil {
			// Log error but can't do much at this point
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		}
	}
}

// Error sends an error response
func Error(w http.ResponseWriter, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationErrors sends a validation error response
func ValidationErrors(w http.ResponseWriter, errors []ValidationError) {
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:    "VALIDATION_ERROR",
			Message: "Invalid request payload",
			Details: errors,
		},
	})
}

// NotFound sends a 404 response
func NotFound(w http.ResponseWriter, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, http.StatusNotFound, "NOT_FOUND", message)
}

// Unauthorized sends a 401 response
func Unauthorized(w http.ResponseWriter) {
	Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
}

// Forbidden sends a 403 response
func Forbidden(w http.ResponseWriter) {
	Error(w, http.StatusForbidden, "FORBIDDEN", "Access denied")
}

// InternalError sends a 500 response
func InternalError(w http.ResponseWriter) {
	Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

// Created sends a 201 response with the created resource
func Created(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusCreated, data)
}

// NoContent sends a 204 response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// OK sends a 200 response
func OK(w http.ResponseWriter, data interface{}) {
	JSON(w, http.StatusOK, data)
}

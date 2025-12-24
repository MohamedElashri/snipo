package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/MohamedElashri/snipo/internal/api/middleware"
	"github.com/MohamedElashri/snipo/internal/validation"
)

const APIVersion = "1.0"

// MaxJSONBodySize is the maximum allowed size for JSON request bodies (2MB)
const MaxJSONBodySize = 2 * 1024 * 1024

// DecodeJSON safely decodes JSON from request body with size limit
func DecodeJSON(r *http.Request, v interface{}) error {
	// Limit request body size to prevent DoS
	r.Body = http.MaxBytesReader(nil, r.Body, MaxJSONBodySize)

	decoder := json.NewDecoder(r.Body)
	decoder.DisallowUnknownFields() // Reject unknown fields for stricter validation

	if err := decoder.Decode(v); err != nil {
		// Return err with more context for debugging
		return err
	}

	// Ensure only one JSON object in body
	if decoder.More() {
		return io.ErrUnexpectedEOF
	}

	return nil
}

// APIResponse is the standard response envelope
type APIResponse struct {
	Data  interface{} `json:"data"`
	Meta  *Meta       `json:"meta,omitempty"`
	Error *ErrorDetail `json:"error,omitempty"`
}

// Meta contains response metadata
type Meta struct {
	RequestID string    `json:"request_id"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version"`
}

// PaginationLinks contains navigation links for paginated responses
type PaginationLinks struct {
	Self string  `json:"self"`
	Next *string `json:"next"`
	Prev *string `json:"prev"`
}

// Pagination contains pagination metadata
type Pagination struct {
	Page       int              `json:"page"`
	Limit      int              `json:"limit"`
	Total      int              `json:"total"`
	TotalPages int              `json:"total_pages"`
	Links      *PaginationLinks `json:"links,omitempty"`
}

// ListResponse wraps list data with pagination
type ListResponse struct {
	Data       interface{} `json:"data"`
	Pagination *Pagination `json:"pagination,omitempty"`
	Meta       *Meta       `json:"meta,omitempty"`
}

// ErrorResponse represents an API error response
type ErrorResponse struct {
	Error ErrorDetail `json:"error"`
}

// ErrorDetail contains error details
type ErrorDetail struct {
	Code      string                      `json:"code"`
	Message   string                      `json:"message"`
	Details   []validation.ValidationError `json:"details,omitempty"`
	RequestID string                      `json:"request_id,omitempty"`
	Timestamp time.Time                   `json:"timestamp,omitempty"`
}

// getMeta extracts metadata from request context
func getMeta(r *http.Request) *Meta {
	requestID := middleware.GetRequestID(r.Context())
	return &Meta{
		RequestID: requestID,
		Timestamp: time.Now().UTC(),
		Version:   APIVersion,
	}
}

// buildPaginationLinks generates navigation links for pagination
func buildPaginationLinks(r *http.Request, page, limit, total int) *PaginationLinks {
	baseURL := fmt.Sprintf("%s://%s%s", scheme(r), r.Host, r.URL.Path)
	query := r.URL.Query()
	
	// Self link
	query.Set("page", fmt.Sprintf("%d", page))
	query.Set("limit", fmt.Sprintf("%d", limit))
	self := fmt.Sprintf("%s?%s", baseURL, query.Encode())
	
	links := &PaginationLinks{
		Self: self,
	}
	
	// Next link
	totalPages := (total + limit - 1) / limit
	if page < totalPages {
		query.Set("page", fmt.Sprintf("%d", page+1))
		next := fmt.Sprintf("%s?%s", baseURL, query.Encode())
		links.Next = &next
	}
	
	// Previous link
	if page > 1 {
		query.Set("page", fmt.Sprintf("%d", page-1))
		prev := fmt.Sprintf("%s?%s", baseURL, query.Encode())
		links.Prev = &prev
	}
	
	return links
}

// scheme returns http or https based on request
func scheme(r *http.Request) string {
	if r.TLS != nil {
		return "https"
	}
	// Check X-Forwarded-Proto header (for reverse proxies)
	if proto := r.Header.Get("X-Forwarded-Proto"); proto != "" {
		return proto
	}
	return "http"
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

// Success sends a standardized success response with metadata
func Success(w http.ResponseWriter, r *http.Request, status int, data interface{}) {
	response := APIResponse{
		Data: data,
		Meta: getMeta(r),
	}
	JSON(w, status, response)
}

// SuccessList sends a standardized list response with pagination
func SuccessList(w http.ResponseWriter, r *http.Request, data interface{}, page, limit, total int) {
	totalPages := (total + limit - 1) / limit
	if totalPages == 0 {
		totalPages = 1
	}
	
	response := ListResponse{
		Data: data,
		Pagination: &Pagination{
			Page:       page,
			Limit:      limit,
			Total:      total,
			TotalPages: totalPages,
			Links:      buildPaginationLinks(r, page, limit, total),
		},
		Meta: getMeta(r),
	}
	JSON(w, http.StatusOK, response)
}

// Error sends an error response
func Error(w http.ResponseWriter, r *http.Request, status int, code, message string) {
	JSON(w, status, ErrorResponse{
		Error: ErrorDetail{
			Code:    code,
			Message: message,
		},
	})
}

// ValidationErrors sends a validation error response
func ValidationErrors(w http.ResponseWriter, r *http.Request, errors validation.ValidationErrors) {
	meta := getMeta(r)
	JSON(w, http.StatusBadRequest, ErrorResponse{
		Error: ErrorDetail{
			Code:      "VALIDATION_ERROR",
			Message:   "Invalid request payload",
			Details:   errors,
			RequestID: meta.RequestID,
			Timestamp: meta.Timestamp,
		},
	})
}

// NotFound sends a 404 response
func NotFound(w http.ResponseWriter, r *http.Request, message string) {
	if message == "" {
		message = "Resource not found"
	}
	Error(w, r, http.StatusNotFound, "NOT_FOUND", message)
}

// Unauthorized sends a 401 response
func Unauthorized(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusUnauthorized, "UNAUTHORIZED", "Authentication required")
}

// Forbidden sends a 403 response
func Forbidden(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusForbidden, "FORBIDDEN", "Access denied")
}

// InternalError sends a 500 response
func InternalError(w http.ResponseWriter, r *http.Request) {
	Error(w, r, http.StatusInternalServerError, "INTERNAL_ERROR", "An internal error occurred")
}

// Created sends a 201 response with the created resource
func Created(w http.ResponseWriter, r *http.Request, data interface{}) {
	Success(w, r, http.StatusCreated, data)
}

// NoContent sends a 204 response
func NoContent(w http.ResponseWriter) {
	w.WriteHeader(http.StatusNoContent)
}

// OK sends a 200 response
func OK(w http.ResponseWriter, r *http.Request, data interface{}) {
	Success(w, r, http.StatusOK, data)
}

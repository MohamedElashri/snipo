package handlers

import (
	"database/sql"
	"net/http"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db *sql.DB
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB) *HealthHandler {
	return &HealthHandler{
		db: db,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status string `json:"status"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	status := "healthy"

	// Check database
	if err := h.db.Ping(); err != nil {
		status = "unhealthy"
	}

	response := HealthResponse{
		Status: status,
	}

	if status == "healthy" {
		OK(w, r, response)
	} else {
		JSON(w, http.StatusServiceUnavailable, response)
	}
}

// Ping handles GET /ping - simple liveness check
func (h *HealthHandler) Ping(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write([]byte("pong"))
}

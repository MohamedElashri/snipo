package handlers

import (
	"database/sql"
	"net/http"
	"runtime"
	"time"
)

// HealthHandler handles health check requests
type HealthHandler struct {
	db        *sql.DB
	startTime time.Time
	version   string
	commit    string
}

// NewHealthHandler creates a new health handler
func NewHealthHandler(db *sql.DB, version, commit string) *HealthHandler {
	return &HealthHandler{
		db:        db,
		startTime: time.Now(),
		version:   version,
		commit:    commit,
	}
}

// HealthResponse represents the health check response
type HealthResponse struct {
	Status    string            `json:"status"`
	Version   string            `json:"version"`
	Commit    string            `json:"commit,omitempty"`
	Uptime    string            `json:"uptime"`
	Checks    map[string]string `json:"checks"`
	Memory    MemoryStats       `json:"memory"`
	Timestamp string            `json:"timestamp"`
}

// MemoryStats represents memory statistics
type MemoryStats struct {
	Alloc      uint64 `json:"alloc_mb"`
	TotalAlloc uint64 `json:"total_alloc_mb"`
	Sys        uint64 `json:"sys_mb"`
	NumGC      uint32 `json:"num_gc"`
}

// Health handles GET /health
func (h *HealthHandler) Health(w http.ResponseWriter, r *http.Request) {
	checks := make(map[string]string)
	status := "healthy"

	// Check database
	if err := h.db.Ping(); err != nil {
		checks["database"] = "unhealthy: " + err.Error()
		status = "unhealthy"
	} else {
		checks["database"] = "healthy"
	}

	// Get memory stats
	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	response := HealthResponse{
		Status:  status,
		Version: h.version,
		Commit:  h.commit,
		Uptime:  time.Since(h.startTime).Round(time.Second).String(),
		Checks:  checks,
		Memory: MemoryStats{
			Alloc:      m.Alloc / 1024 / 1024,
			TotalAlloc: m.TotalAlloc / 1024 / 1024,
			Sys:        m.Sys / 1024 / 1024,
			NumGC:      m.NumGC,
		},
		Timestamp: time.Now().UTC().Format(time.RFC3339),
	}

	if status == "healthy" {
		OK(w, response)
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

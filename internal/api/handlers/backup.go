package handlers

import (
	"io"
	"log/slog"
	"net/http"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/services"
)

// BackupHandler handles backup-related HTTP requests
type BackupHandler struct {
	backupSvc *services.BackupService
	s3SyncSvc *services.S3SyncService // May be nil if S3 is not configured
}

// NewBackupHandler creates a new backup handler
func NewBackupHandler(backupSvc *services.BackupService, s3SyncSvc *services.S3SyncService) *BackupHandler {
	return &BackupHandler{
		backupSvc: backupSvc,
		s3SyncSvc: s3SyncSvc,
	}
}

// Export handles GET /api/v1/backup/export
// Query params: format (json|zip), password (optional)
func (h *BackupHandler) Export(w http.ResponseWriter, r *http.Request) {
	opts := models.ExportOptions{
		Format:   r.URL.Query().Get("format"),
		Password: r.URL.Query().Get("password"),
	}

	if opts.Format == "" {
		opts.Format = "json"
	}

	content, filename, err := h.backupSvc.Export(r.Context(), opts)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "BACKUP_FAILED", err.Error())
		return
	}

	// Determine content type
	contentType := "application/json"
	if opts.Format == "zip" {
		contentType = "application/zip"
	}
	if opts.Password != "" {
		contentType = "application/octet-stream"
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Content-Disposition", "attachment; filename=\""+filename+"\"")
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(content)
}

// Import handles POST /api/v1/backup/import
// Form data: file (multipart), strategy (replace|merge|skip), password (optional)
func (h *BackupHandler) Import(w http.ResponseWriter, r *http.Request) {
	// Parse multipart form (max 50MB)
	if err := r.ParseMultipartForm(50 << 20); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Failed to parse form data")
		return
	}

	file, _, err := r.FormFile("file")
	if err != nil {
		Error(w, r, http.StatusBadRequest, "MISSING_FILE", "No backup file provided")
		return
	}
	defer func() {
		if err := file.Close(); err != nil {
			slog.Error("failed to close file", "error", err)
		}
	}()

	content, err := io.ReadAll(file)
	if err != nil {
		Error(w, r, http.StatusBadRequest, "READ_ERROR", "Failed to read backup file")
		return
	}

	opts := models.ImportOptions{
		Strategy: r.FormValue("strategy"),
		Password: r.FormValue("password"),
	}

	if opts.Strategy == "" {
		opts.Strategy = "merge"
	}

	result, err := h.backupSvc.Import(r.Context(), content, opts)
	if err != nil {
		if err == services.ErrDecryptionFailed {
			Error(w, r, http.StatusBadRequest, "DECRYPTION_FAILED", "Failed to decrypt backup - wrong password?")
			return
		}
		if err == services.ErrInvalidBackupFormat {
			Error(w, r, http.StatusBadRequest, "INVALID_FORMAT", "Invalid backup file format")
			return
		}
		Error(w, r, http.StatusInternalServerError, "IMPORT_FAILED", err.Error())
		return
	}

	OK(w, r, result)
}

// S3Sync handles POST /api/v1/backup/s3/sync
// Body: { "format": "json|zip", "password": "optional" }
func (h *BackupHandler) S3Sync(w http.ResponseWriter, r *http.Request) {
	if h.s3SyncSvc == nil {
		Error(w, r, http.StatusServiceUnavailable, "S3_NOT_CONFIGURED", "S3 storage is not configured")
		return
	}

	var opts models.ExportOptions
	if err := DecodeJSON(r, &opts); err != nil {
		// Use defaults if no body
		opts.Format = "json"
	}

	result, err := h.s3SyncSvc.SyncToS3(r.Context(), opts)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "SYNC_FAILED", err.Error())
		return
	}

	OK(w, r, result)
}

// S3List handles GET /api/v1/backup/s3/list
func (h *BackupHandler) S3List(w http.ResponseWriter, r *http.Request) {
	if h.s3SyncSvc == nil {
		Error(w, r, http.StatusServiceUnavailable, "S3_NOT_CONFIGURED", "S3 storage is not configured")
		return
	}

	backups, err := h.s3SyncSvc.ListBackups(r.Context())
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "LIST_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]interface{}{
		"backups": backups,
	})
}

// S3Restore handles POST /api/v1/backup/s3/restore
// Body: { "key": "backups/snipo-backup-xxx.json", "strategy": "replace|merge|skip", "password": "optional" }
func (h *BackupHandler) S3Restore(w http.ResponseWriter, r *http.Request) {
	if h.s3SyncSvc == nil {
		Error(w, r, http.StatusServiceUnavailable, "S3_NOT_CONFIGURED", "S3 storage is not configured")
		return
	}

	var req struct {
		Key      string `json:"key"`
		Strategy string `json:"strategy"`
		Password string `json:"password"`
	}

	if err := DecodeJSON(r, &req); err != nil {
		Error(w, r, http.StatusBadRequest, "INVALID_REQUEST", "Invalid request body")
		return
	}

	if req.Key == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_KEY", "Backup key is required")
		return
	}

	opts := models.ImportOptions{
		Strategy: req.Strategy,
		Password: req.Password,
	}

	if opts.Strategy == "" {
		opts.Strategy = "merge"
	}

	result, err := h.s3SyncSvc.RestoreFromS3(r.Context(), req.Key, opts)
	if err != nil {
		Error(w, r, http.StatusInternalServerError, "RESTORE_FAILED", err.Error())
		return
	}

	OK(w, r, result)
}

// S3Delete handles DELETE /api/v1/backup/s3/{key}
func (h *BackupHandler) S3Delete(w http.ResponseWriter, r *http.Request) {
	if h.s3SyncSvc == nil {
		Error(w, r, http.StatusServiceUnavailable, "S3_NOT_CONFIGURED", "S3 storage is not configured")
		return
	}

	key := r.URL.Query().Get("key")
	if key == "" {
		Error(w, r, http.StatusBadRequest, "MISSING_KEY", "Backup key is required")
		return
	}

	if err := h.s3SyncSvc.DeleteBackup(r.Context(), key); err != nil {
		Error(w, r, http.StatusInternalServerError, "DELETE_FAILED", err.Error())
		return
	}

	OK(w, r, map[string]string{
		"message": "Backup deleted successfully",
	})
}

// S3Status handles GET /api/v1/backup/s3/status
func (h *BackupHandler) S3Status(w http.ResponseWriter, r *http.Request) {
	status := map[string]interface{}{
		"enabled": h.s3SyncSvc != nil,
	}

	OK(w, r, status)
}

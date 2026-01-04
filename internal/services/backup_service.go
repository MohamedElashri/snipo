package services

import (
	"archive/zip"
	"bytes"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/repository"
	"golang.org/x/crypto/pbkdf2"
)

const (
	BackupVersion = "1.0"
)

var (
	ErrInvalidBackupFormat = errors.New("invalid backup format")
	ErrDecryptionFailed    = errors.New("decryption failed - wrong password?")
)

// BackupService handles backup and restore operations
type BackupService struct {
	db             *sql.DB
	snippetSvc     *SnippetService
	tagRepo        *repository.TagRepository
	folderRepo     *repository.FolderRepository
	fileRepo       *repository.SnippetFileRepository
	logger         *slog.Logger
	encryptionSalt string
}

// NewBackupService creates a new backup service
func NewBackupService(
	db *sql.DB,
	snippetSvc *SnippetService,
	tagRepo *repository.TagRepository,
	folderRepo *repository.FolderRepository,
	fileRepo *repository.SnippetFileRepository,
	logger *slog.Logger,
	encryptionSalt string,
) *BackupService {
	return &BackupService{
		db:             db,
		snippetSvc:     snippetSvc,
		tagRepo:        tagRepo,
		folderRepo:     folderRepo,
		fileRepo:       fileRepo,
		logger:         logger,
		encryptionSalt: encryptionSalt,
	}
}

// Export creates a complete backup of all data
func (b *BackupService) Export(ctx context.Context, opts models.ExportOptions) ([]byte, string, error) {
	data := models.BackupData{
		Version:   BackupVersion,
		CreatedAt: time.Now().UTC(),
	}

	// Gather all snippets with their files
	snippetList, err := b.snippetSvc.List(ctx, models.SnippetFilter{
		Page:  1,
		Limit: 10000, // Get all snippets
	})
	if err != nil {
		return nil, "", fmt.Errorf("failed to get snippets: %w", err)
	}

	// Fetch full details for each snippet (including files, tags, folders)
	for _, s := range snippetList.Data {
		snippet, err := b.snippetSvc.GetByID(ctx, s.ID)
		if err != nil {
			b.logger.Warn("failed to get snippet details", "id", s.ID, "error", err)
			continue
		}
		data.Snippets = append(data.Snippets, *snippet)
	}

	// Gather all tags
	if b.tagRepo != nil {
		tags, err := b.tagRepo.List(ctx)
		if err != nil {
			b.logger.Warn("failed to get tags", "error", err)
		} else {
			data.Tags = tags
		}
	}

	// Gather all folders
	if b.folderRepo != nil {
		folders, err := b.folderRepo.List(ctx)
		if err != nil {
			b.logger.Warn("failed to get folders", "error", err)
		} else {
			data.Folders = folders
		}
	}

	var content []byte
	var filename string

	if opts.Format == "zip" {
		content, err = b.createZipBackup(data)
		if err != nil {
			return nil, "", fmt.Errorf("failed to create zip backup: %w", err)
		}
		filename = fmt.Sprintf("snipo-backup-%s.zip", time.Now().Format("2006-01-02-150405"))
	} else {
		// Default to JSON
		content, err = json.MarshalIndent(data, "", "  ")
		if err != nil {
			return nil, "", fmt.Errorf("failed to marshal backup: %w", err)
		}
		filename = fmt.Sprintf("snipo-backup-%s.json", time.Now().Format("2006-01-02-150405"))
	}

	// Encrypt if password provided
	if opts.Password != "" {
		content, err = b.encrypt(content, opts.Password)
		if err != nil {
			return nil, "", fmt.Errorf("failed to encrypt backup: %w", err)
		}
		filename = filename + ".enc"
	}

	b.logger.Info("backup exported",
		"snippets", len(data.Snippets),
		"tags", len(data.Tags),
		"folders", len(data.Folders),
		"format", opts.Format,
		"encrypted", opts.Password != "",
	)

	return content, filename, nil
}

// Import restores data from a backup
func (b *BackupService) Import(ctx context.Context, content []byte, opts models.ImportOptions) (*models.ImportResult, error) {
	// Decrypt if password provided
	var err error
	if opts.Password != "" {
		content, err = b.decrypt(content, opts.Password)
		if err != nil {
			return nil, ErrDecryptionFailed
		}
	}

	var data models.BackupData

	// Try JSON first
	if err := json.Unmarshal(content, &data); err != nil {
		// Try ZIP
		zr, zipErr := zip.NewReader(bytes.NewReader(content), int64(len(content)))
		if zipErr != nil {
			return nil, ErrInvalidBackupFormat
		}

		for _, f := range zr.File {
			if f.Name == "metadata.json" {
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open metadata: %w", err)
				}
				if err := json.NewDecoder(rc).Decode(&data); err != nil {
					_ = rc.Close()
					return nil, fmt.Errorf("failed to decode metadata: %w", err)
				}
				_ = rc.Close()
				break
			}
		}

		if data.Version == "" {
			return nil, ErrInvalidBackupFormat
		}
	}

	result := &models.ImportResult{}

	// Handle strategy
	if opts.Strategy == "replace" {
		if err := b.clearAllData(ctx); err != nil {
			return nil, fmt.Errorf("failed to clear existing data: %w", err)
		}
	}

	// Build lookup maps for existing data to avoid duplicates
	existingTags, _ := b.tagRepo.List(ctx)
	existingTagsByName := make(map[string]*models.Tag)
	for i := range existingTags {
		existingTagsByName[existingTags[i].Name] = &existingTags[i]
	}

	existingFolders, _ := b.folderRepo.List(ctx)
	existingFoldersByName := make(map[string]*models.Folder)
	for i := range existingFolders {
		existingFoldersByName[existingFolders[i].Name] = &existingFolders[i]
	}

	existingSnippets, _ := b.snippetSvc.List(ctx, models.SnippetFilter{Limit: 10000})
	existingSnippetsByTitle := make(map[string]*models.Snippet)
	if existingSnippets != nil {
		for i := range existingSnippets.Data {
			existingSnippetsByTitle[existingSnippets.Data[i].Title] = &existingSnippets.Data[i]
		}
	}

	// Import tags first (needed for relationships)
	tagMap := make(map[int64]int64) // old ID -> new ID
	for _, tag := range data.Tags {
		oldID := tag.ID
		// Check if tag already exists by name
		if existingTag, exists := existingTagsByName[tag.Name]; exists {
			tagMap[oldID] = existingTag.ID
			// Don't count as imported since it already existed
		} else {
			// Create new tag
			newTag, err := b.tagRepo.Create(ctx, &models.TagInput{
				Name:  tag.Name,
				Color: tag.Color,
			})
			if err == nil {
				tagMap[oldID] = newTag.ID
				existingTagsByName[tag.Name] = newTag // Add to map to prevent duplicates
				result.TagsImported++
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("tag %s: %v", tag.Name, err))
			}
		}
	}

	// Import folders
	folderMap := make(map[int64]int64) // old ID -> new ID
	// First pass: create folders without parent relationships (only if they don't exist)
	for _, folder := range data.Folders {
		oldID := folder.ID
		// Check if folder already exists by name
		if existingFolder, exists := existingFoldersByName[folder.Name]; exists {
			folderMap[oldID] = existingFolder.ID
			// Don't count as imported since it already existed
		} else {
			input := &models.FolderInput{
				Name:      folder.Name,
				Icon:      folder.Icon,
				SortOrder: folder.SortOrder,
			}
			newFolder, err := b.folderRepo.Create(ctx, input)
			if err == nil {
				folderMap[oldID] = newFolder.ID
				existingFoldersByName[folder.Name] = newFolder // Add to map to prevent duplicates
				result.FoldersImported++
			} else {
				result.Errors = append(result.Errors, fmt.Sprintf("folder %s: %v", folder.Name, err))
			}
		}
	}

	// Second pass: update parent relationships for newly created folders
	for _, folder := range data.Folders {
		if folder.ParentID != nil {
			// Only update if this folder was newly created
			if _, existed := existingFoldersByName[folder.Name]; !existed {
				if newID, ok := folderMap[folder.ID]; ok {
					if newParentID, ok := folderMap[*folder.ParentID]; ok {
						_, _ = b.folderRepo.Move(ctx, newID, &newParentID)
					}
				}
			}
		}
	}

	// Import snippets
	for _, snippet := range data.Snippets {
		// Check if snippet with same title already exists
		if _, exists := existingSnippetsByTitle[snippet.Title]; exists {
			// Skip if strategy is "skip" or "merge" (merge doesn't overwrite existing)
			if opts.Strategy == "skip" || opts.Strategy == "merge" {
				continue
			}
		}

		// Prepare input
		input := &models.SnippetInput{
			Title:       snippet.Title,
			Description: snippet.Description,
			Content:     snippet.Content,
			Language:    snippet.Language,
			IsPublic:    snippet.IsPublic,
			IsArchived:  snippet.IsArchived,
		}

		// Map tags
		for _, tag := range snippet.Tags {
			input.Tags = append(input.Tags, tag.Name)
		}

		// Map folder (use first folder if any)
		if len(snippet.Folders) > 0 {
			if newFolderID, ok := folderMap[snippet.Folders[0].ID]; ok {
				input.FolderID = &newFolderID
			}
		}

		// Map files
		for _, file := range snippet.Files {
			input.Files = append(input.Files, models.SnippetFileInput{
				Filename: file.Filename,
				Content:  file.Content,
				Language: file.Language,
			})
		}

		_, err := b.snippetSvc.Create(ctx, input)
		if err == nil {
			result.SnippetsImported++
			// Add to map to prevent duplicates within same import
			existingSnippetsByTitle[snippet.Title] = &snippet
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("snippet %s: %v", snippet.Title, err))
		}
	}

	b.logger.Info("backup imported",
		"snippets", result.SnippetsImported,
		"tags", result.TagsImported,
		"folders", result.FoldersImported,
		"errors", len(result.Errors),
	)

	return result, nil
}

// createZipBackup creates a ZIP archive with snippets as individual files
func (b *BackupService) createZipBackup(data models.BackupData) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	// Add snippets as individual files
	for _, s := range data.Snippets {
		// Add each file in the snippet
		if len(s.Files) > 0 {
			for _, f := range s.Files {
				filename := fmt.Sprintf("snippets/%s/%s", sanitizeFilename(s.Title), f.Filename)
				w, err := zw.Create(filename)
				if err != nil {
					return nil, err
				}
				if _, err := w.Write([]byte(f.Content)); err != nil {
					return nil, err
				}
			}
		} else {
			// Legacy single-file snippet
			ext := getExtension(s.Language)
			filename := fmt.Sprintf("snippets/%s.%s", sanitizeFilename(s.Title), ext)
			w, err := zw.Create(filename)
			if err != nil {
				return nil, err
			}
			if _, err := w.Write([]byte(s.Content)); err != nil {
				return nil, err
			}
		}
	}

	// Add metadata
	metaW, err := zw.Create("metadata.json")
	if err != nil {
		return nil, err
	}
	if err := json.NewEncoder(metaW).Encode(data); err != nil {
		return nil, err
	}

	if err := zw.Close(); err != nil {
		return nil, err
	}

	return buf.Bytes(), nil
}

// clearAllData removes all snippets, tags, and folders
func (b *BackupService) clearAllData(ctx context.Context) error {
	queries := []string{
		"DELETE FROM snippet_tags",
		"DELETE FROM snippet_folders",
		"DELETE FROM snippet_files",
		"DELETE FROM snippets",
		"DELETE FROM tags",
		"DELETE FROM folders",
	}

	for _, q := range queries {
		if _, err := b.db.ExecContext(ctx, q); err != nil {
			b.logger.Warn("failed to execute clear query", "query", q, "error", err)
		}
	}

	return nil
}

// sanitizeFilename removes invalid characters from filename
func sanitizeFilename(name string) string {
	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	result := name
	for _, c := range invalid {
		result = strings.ReplaceAll(result, c, "_")
	}
	// Limit length
	if len(result) > 50 {
		result = result[:50]
	}
	return result
}

// getExtension returns file extension for a language
func getExtension(lang string) string {
	extensions := map[string]string{
		"javascript": "js",
		"typescript": "ts",
		"python":     "py",
		"go":         "go",
		"rust":       "rs",
		"java":       "java",
		"c":          "c",
		"cpp":        "cpp",
		"csharp":     "cs",
		"php":        "php",
		"ruby":       "rb",
		"swift":      "swift",
		"kotlin":     "kt",
		"scala":      "scala",
		"html":       "html",
		"css":        "css",
		"scss":       "scss",
		"json":       "json",
		"yaml":       "yaml",
		"xml":        "xml",
		"markdown":   "md",
		"sql":        "sql",
		"bash":       "sh",
		"shell":      "sh",
		"powershell": "ps1",
		"dockerfile": "dockerfile",
		"nginx":      "conf",
		"toml":       "toml",
		"ini":        "ini",
		"makefile":   "mk",
	}

	if ext, ok := extensions[lang]; ok {
		return ext
	}
	return "txt"
}

// deriveKey derives a 32-byte key from password using PBKDF2
func (b *BackupService) deriveKey(password string) []byte {
	return pbkdf2.Key([]byte(password), []byte(b.encryptionSalt), 100000, 32, sha256.New)
}

// deriveKeyLegacy derives a 32-byte key using the old SHA256 method (for backward compatibility)
func deriveKeyLegacy(password string) []byte {
	hash := sha256.Sum256([]byte(password))
	return hash[:]
}

// encrypt encrypts data using AES-256-GCM
func (b *BackupService) encrypt(data []byte, password string) ([]byte, error) {
	key := b.deriveKey(password)

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonce := make([]byte, gcm.NonceSize())
	if _, err := rand.Read(nonce); err != nil {
		return nil, err
	}

	return gcm.Seal(nonce, nonce, data, nil), nil
}

// decrypt decrypts data using AES-256-GCM with backward compatibility
func (b *BackupService) decrypt(data []byte, password string) ([]byte, error) {
	// Try new PBKDF2 method first
	result, err := decryptWithKey(data, b.deriveKey(password))
	if err == nil {
		return result, nil
	}

	// Fall back to legacy SHA256 method for old backups
	result, legacyErr := decryptWithKey(data, deriveKeyLegacy(password))
	if legacyErr == nil {
		return result, nil
	}

	return nil, err
}

// decryptWithKey performs the actual AES-256-GCM decryption with a given key
func decryptWithKey(data []byte, key []byte) ([]byte, error) {
	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, err
	}

	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, err
	}

	nonceSize := gcm.NonceSize()
	if len(data) < nonceSize {
		return nil, errors.New("ciphertext too short")
	}

	nonce, ciphertext := data[:nonceSize], data[nonceSize:]
	return gcm.Open(nil, nonce, ciphertext, nil)
}

// GetFilename generates a backup filename
func GetBackupFilename(format string, encrypted bool) string {
	timestamp := time.Now().Format("2006-01-02-150405")
	ext := "json"
	if format == "zip" {
		ext = "zip"
	}
	filename := fmt.Sprintf("snipo-backup-%s.%s", timestamp, ext)
	if encrypted {
		filename += ".enc"
	}
	return filename
}

// ValidateBackupFile checks if the content looks like a valid backup
func ValidateBackupFile(content []byte) (string, error) {
	// Check for JSON
	var data models.BackupData
	if err := json.Unmarshal(content, &data); err == nil {
		if data.Version != "" {
			return "json", nil
		}
	}

	// Check for ZIP
	if len(content) > 4 && content[0] == 'P' && content[1] == 'K' {
		return "zip", nil
	}

	// Check for encrypted (starts with random bytes, so just check it's not empty)
	if len(content) > 32 {
		return "encrypted", nil
	}

	return "", ErrInvalidBackupFormat
}

// Unused but kept for reference
var _ = filepath.Base

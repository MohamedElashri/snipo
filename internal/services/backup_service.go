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
	"io"
	"log/slog"
	"path/filepath"
	"strings"
	"time"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
	"golang.org/x/crypto/pbkdf2"
)

const (
	BackupVersion = "1.0"
	// MaxBackupSize bounds uploaded and decoded backup data, including ZIP metadata.
	MaxBackupSize = 50 << 20
)

var (
	ErrBackupTooLarge      = errors.New("backup exceeds the 50 MiB size limit")
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

	// Paginate explicitly; the ordinary list API caps each page at 100 items.
	snippets, err := b.listBackupSnippets(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("failed to get snippets: %w", err)
	}
	data.Snippets = snippets

	// Gather all tags
	if b.tagRepo != nil {
		tags, err := b.tagRepo.List(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get tags: %w", err)
		} else {
			data.Tags = tags
		}
	}

	// Gather all folders
	if b.folderRepo != nil {
		folders, err := b.folderRepo.List(ctx)
		if err != nil {
			return nil, "", fmt.Errorf("failed to get folders: %w", err)
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
	)

	return content, filename, nil
}

// listBackupSnippets includes archived items and every page, while preserving
// the existing exclusion of Trash. List already loads files and relationships.
func (b *BackupService) listBackupSnippets(ctx context.Context) ([]models.Snippet, error) {
	var snippets []models.Snippet
	for _, archived := range []bool{false, true} {
		for page := 1; ; page++ {
			response, err := b.snippetSvc.List(ctx, models.SnippetFilter{Page: page, Limit: 100,
				SortBy: "id", SortOrder: "asc", IsArchived: &archived})
			if err != nil {
				return nil, err
			}
			snippets = append(snippets, response.Data...)
			if page >= response.Pagination.TotalPages {
				break
			}
		}
	}
	return snippets, nil
}

func backupSnippetInput(snippet models.Snippet) *models.SnippetInput {
	input := &models.SnippetInput{Title: snippet.Title, Description: snippet.Description,
		Content: snippet.Content, Language: snippet.Language, IsPublic: snippet.IsPublic,
		IsArchived: snippet.IsArchived}
	for _, tag := range snippet.Tags {
		input.Tags = append(input.Tags, tag.Name)
	}
	for _, file := range snippet.Files {
		input.Files = append(input.Files, models.SnippetFileInput{Filename: file.Filename, Content: file.Content, Language: file.Language})
	}
	return input
}

// Check a replacement before clearing anything. Merge/skip retain partial-import
// behavior, where invalid individual records are reported without deleting data.
func validateReplacementBackup(data models.BackupData) error {
	for i, tag := range data.Tags {
		if errs := validation.ValidateTagInput(tag.Name); errs.HasErrors() {
			return fmt.Errorf("%w: tag %d: %s", ErrInvalidBackupFormat, i+1, errs)
		}
	}
	for i, folder := range data.Folders {
		if errs := validation.ValidateFolderInput(folder.Name); errs.HasErrors() {
			return fmt.Errorf("%w: folder %d: %s", ErrInvalidBackupFormat, i+1, errs)
		}
	}
	for i, snippet := range data.Snippets {
		if errs := validation.ValidateSnippetInput(backupSnippetInput(snippet)); errs.HasErrors() {
			return fmt.Errorf("%w: snippet %d: %s", ErrInvalidBackupFormat, i+1, errs)
		}
	}
	return nil
}

// A matching version alone is not a backup: require its snippets collection.
// An explicit null/empty collection remains valid for genuine empty exports.
func decodeBackupMetadata(content []byte, data *models.BackupData) error {
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(content, &fields); err != nil || fields["snippets"] == nil {
		return ErrInvalidBackupFormat
	}
	if err := json.Unmarshal(content, data); err != nil || data.Version != BackupVersion {
		return ErrInvalidBackupFormat
	}
	return nil
}

// Import restores data from a backup
func (b *BackupService) Import(ctx context.Context, content []byte, opts models.ImportOptions) (*models.ImportResult, error) {
	if len(content) > MaxBackupSize {
		return nil, ErrBackupTooLarge
	}
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
	if err := decodeBackupMetadata(content, &data); err != nil {
		// Try ZIP
		zr, zipErr := zip.NewReader(bytes.NewReader(content), int64(len(content)))
		if zipErr != nil {
			return nil, ErrInvalidBackupFormat
		}

		metadataFound := false
		for _, f := range zr.File {
			if f.Name == "metadata.json" {
				if f.UncompressedSize64 > MaxBackupSize {
					return nil, ErrBackupTooLarge
				}
				rc, err := f.Open()
				if err != nil {
					return nil, fmt.Errorf("failed to open metadata: %w", err)
				}
				metadata, readErr := io.ReadAll(io.LimitReader(rc, MaxBackupSize+1))
				_ = rc.Close()
				if len(metadata) > MaxBackupSize {
					return nil, ErrBackupTooLarge
				}
				if readErr != nil {
					return nil, fmt.Errorf("failed to read metadata: %w", readErr)
				}
				if err := decodeBackupMetadata(metadata, &data); err != nil {
					return nil, fmt.Errorf("failed to decode metadata: %w", err)
				}
				metadataFound = true
				break
			}
		}
		if !metadataFound {
			return nil, ErrInvalidBackupFormat
		}
	}

	if opts.Strategy == "replace" {
		if err := validateReplacementBackup(data); err != nil {
			return nil, err
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

	existingSnippets, err := b.listBackupSnippets(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list existing snippets: %w", err)
	}
	existingSnippetsByTitle := make(map[string]*models.Snippet)
	for i := range existingSnippets {
		existingSnippetsByTitle[existingSnippets[i].Title] = &existingSnippets[i]
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
				result.Errors = append(result.Errors, fmt.Sprintf("tag %s: %v", validation.TruncateForLog(tag.Name, 50), err))
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
				result.Errors = append(result.Errors, fmt.Sprintf("folder %s: %v", validation.TruncateForLog(folder.Name, 100), err))
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

		input := backupSnippetInput(snippet)

		// Map folder (use first folder if any)
		if len(snippet.Folders) > 0 {
			if newFolderID, ok := folderMap[snippet.Folders[0].ID]; ok {
				input.FolderID = &newFolderID
			}
		}

		_, err := b.snippetSvc.Create(ctx, input)
		if err == nil {
			result.SnippetsImported++
			// Add to map to prevent duplicates within same import
			existingSnippetsByTitle[snippet.Title] = &snippet
		} else {
			result.Errors = append(result.Errors, fmt.Sprintf("snippet %s: %v", validation.TruncateForLog(snippet.Title, 100), err))
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
			b.logger.Warn("failed to execute clear query", "query", strings.ReplaceAll(strings.ReplaceAll(q, "\n", ""), "\r", ""), "error", err)
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

// decrypt decrypts data using AES-256-GCM
func (b *BackupService) decrypt(data []byte, password string) ([]byte, error) {
	return decryptWithKey(data, b.deriveKey(password))
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

// Unused but kept for reference
var _ = filepath.Base

package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MohamedElashri/snipo/internal/models"
)

const (
	metadataFilename = "zzz-snipo-metadata.json"
	maxGistTopics    = 20
)

// SnippetToGistRequest converts a snippet to a gist request
func SnippetToGistRequest(snippet *models.Snippet) (*models.GistRequest, error) {
	// Build metadata
	metadata := models.SnipoMetadata{
		Version:    "1.0",
		SnipoID:    snippet.ID,
		Folders:    snippet.Folders,
		IsFavorite: snippet.IsFavorite,
		IsArchived: snippet.IsArchived,
	}

	if len(snippet.Tags) > maxGistTopics {
		metadata.TagsOverflow = make([]string, 0)
		for i := maxGistTopics; i < len(snippet.Tags); i++ {
			metadata.TagsOverflow = append(metadata.TagsOverflow, snippet.Tags[i].Name)
		}
	}

	metadataJSON, err := json.Marshal(metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal metadata: %w", err)
	}

	// Embed metadata in description as a special marker
	description := snippet.Title
	if description == "" {
		description = "Untitled Snippet"
	}
	description = fmt.Sprintf("%s\n[snipo:%s]", description, string(metadataJSON))

	req := &models.GistRequest{
		Description: description,
		Public:      snippet.IsPublic,
		Files:       make(map[string]models.GistFile),
	}

	// Add snippet files
	if len(snippet.Files) == 0 {
		// Use snippet title as filename for single-file snippets
		filename := snippet.Title
		if filename == "" {
			filename = "snippet"
		}

		// Sanitize filename (remove invalid characters)
		filename = strings.Map(func(r rune) rune {
			if r == '/' || r == '\\' || r == ':' || r == '*' || r == '?' || r == '"' || r == '<' || r == '>' || r == '|' {
				return '-'
			}
			return r
		}, filename)

		req.Files[filename] = models.GistFile{
			Content: snippet.Content,
		}
	} else {
		for _, file := range snippet.Files {
			req.Files[file.Filename] = models.GistFile{
				Content: file.Content,
			}
		}
	}

	return req, nil
}

// GistToSnippet converts a gist to a snippet
func GistToSnippet(gist *models.GistResponse, existingSnippet *models.Snippet) (*models.Snippet, error) {
	// Extract title and metadata from description
	title := gist.Description
	var metadata *models.SnipoMetadata

	// Check if description contains embedded metadata
	if strings.Contains(gist.Description, "[snipo:") {
		parts := strings.SplitN(gist.Description, "\n[snipo:", 2)
		if len(parts) == 2 {
			title = parts[0]
			metadataJSON := strings.TrimSuffix(parts[1], "]")
			var meta models.SnipoMetadata
			if err := json.Unmarshal([]byte(metadataJSON), &meta); err == nil {
				metadata = &meta
			}
		}
	}

	snippet := &models.Snippet{
		Title:      title,
		IsPublic:   gist.Public,
		Files:      make([]models.SnippetFile, 0),
		Tags:       make([]models.Tag, 0),
		Folders:    make([]models.Folder, 0),
		IsFavorite: false,
		IsArchived: false,
	}

	if existingSnippet != nil {
		snippet.ID = existingSnippet.ID
		snippet.CreatedAt = existingSnippet.CreatedAt
	}

	// Process files (skip metadata file if it exists for backward compatibility)
	for filename, file := range gist.Files {
		if filename == metadataFilename {
			continue
		}

		language := getLanguageFromFilename(filename)
		snippetFile := models.SnippetFile{
			Filename: filename,
			Content:  file.Content,
			Language: language,
		}
		snippet.Files = append(snippet.Files, snippetFile)
	}

	if metadata != nil {
		snippet.Folders = metadata.Folders
		snippet.IsFavorite = metadata.IsFavorite
		snippet.IsArchived = metadata.IsArchived
		if existingSnippet == nil && metadata.SnipoID != "" {
			snippet.ID = metadata.SnipoID
		}
	}

	if len(snippet.Files) == 0 {
		snippet.Content = ""
		snippet.Language = "plaintext"
	} else {
		snippet.Content = snippet.Files[0].Content
		snippet.Language = snippet.Files[0].Language
	}

	if snippet.Description == "" {
		snippet.Description = ""
	}

	return snippet, nil
}

// getExtensionForLanguage returns file extension for a language
func getExtensionForLanguage(language string) string {
	extensions := map[string]string{
		"go":         "go",
		"python":     "py",
		"javascript": "js",
		"typescript": "ts",
		"java":       "java",
		"c":          "c",
		"cpp":        "cpp",
		"csharp":     "cs",
		"ruby":       "rb",
		"php":        "php",
		"rust":       "rs",
		"swift":      "swift",
		"kotlin":     "kt",
		"scala":      "scala",
		"shell":      "sh",
		"bash":       "sh",
		"sql":        "sql",
		"html":       "html",
		"css":        "css",
		"json":       "json",
		"yaml":       "yaml",
		"xml":        "xml",
		"markdown":   "md",
	}

	if ext, ok := extensions[strings.ToLower(language)]; ok {
		return ext
	}
	return "txt"
}

// getLanguageFromFilename infers language from filename
func getLanguageFromFilename(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	if ext == "" {
		return "plaintext"
	}

	ext = strings.TrimPrefix(ext, ".")

	languages := map[string]string{
		"go":    "go",
		"py":    "python",
		"js":    "javascript",
		"ts":    "typescript",
		"java":  "java",
		"c":     "c",
		"cpp":   "cpp",
		"cc":    "cpp",
		"cxx":   "cpp",
		"cs":    "csharp",
		"rb":    "ruby",
		"php":   "php",
		"rs":    "rust",
		"swift": "swift",
		"kt":    "kotlin",
		"scala": "scala",
		"sh":    "shell",
		"bash":  "bash",
		"sql":   "sql",
		"html":  "html",
		"css":   "css",
		"json":  "json",
		"yaml":  "yaml",
		"yml":   "yaml",
		"xml":   "xml",
		"md":    "markdown",
		"txt":   "plaintext",
	}

	if lang, ok := languages[ext]; ok {
		return lang
	}
	return "plaintext"
}

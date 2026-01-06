package services

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/MohamedElashri/snipo/internal/models"
)

const (
	metadataFilename = ".snipo-metadata.json"
	maxGistTopics    = 20
)

// SnippetToGistRequest converts a snippet to a gist request
func SnippetToGistRequest(snippet *models.Snippet) (*models.GistRequest, error) {
	req := &models.GistRequest{
		Description: snippet.Title,
		Public:      snippet.IsPublic,
		Files:       make(map[string]models.GistFile),
	}

	if len(snippet.Files) == 0 {
		filename := "snippet.txt"
		if snippet.Language != "" && snippet.Language != "plaintext" {
			filename = fmt.Sprintf("snippet.%s", getExtensionForLanguage(snippet.Language))
		}
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

	req.Files[metadataFilename] = models.GistFile{
		Content: string(metadataJSON),
	}

	return req, nil
}

// GistToSnippet converts a gist to a snippet
func GistToSnippet(gist *models.GistResponse, existingSnippet *models.Snippet) (*models.Snippet, error) {
	snippet := &models.Snippet{
		Title:      gist.Description,
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

	var metadata *models.SnipoMetadata
	for filename, file := range gist.Files {
		if filename == metadataFilename {
			if err := json.Unmarshal([]byte(file.Content), &metadata); err != nil {
				return nil, fmt.Errorf("failed to unmarshal metadata: %w", err)
			}
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

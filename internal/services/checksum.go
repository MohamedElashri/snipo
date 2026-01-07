package services

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/MohamedElashri/snipo/internal/models"
)

// CalculateSnippetChecksum calculates a checksum for a snippet
func CalculateSnippetChecksum(snippet *models.Snippet) (string, error) {
	data := map[string]interface{}{
		"title":       snippet.Title,
		"description": snippet.Description,
		"is_public":   snippet.IsPublic,
		"files":       make([]map[string]string, 0),
	}

	for _, file := range snippet.Files {
		data["files"] = append(data["files"].([]map[string]string), map[string]string{
			"filename": file.Filename,
			"content":  file.Content,
			"language": file.Language,
		})
	}

	sortedFiles := data["files"].([]map[string]string)
	sort.Slice(sortedFiles, func(i, j int) bool {
		return sortedFiles[i]["filename"] < sortedFiles[j]["filename"]
	})
	data["files"] = sortedFiles

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal snippet data: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

// CalculateGistChecksum calculates a checksum for a gist
func CalculateGistChecksum(gist *models.GistResponse) (string, error) {
	data := map[string]interface{}{
		"description": gist.Description,
		"public":      gist.Public,
		"files":       make([]map[string]string, 0),
	}

	filenames := make([]string, 0, len(gist.Files))
	for filename := range gist.Files {
		filenames = append(filenames, filename)
	}
	sort.Strings(filenames)

	for _, filename := range filenames {
		file := gist.Files[filename]
		data["files"] = append(data["files"].([]map[string]string), map[string]string{
			"filename": filename,
			"content":  file.Content,
		})
	}

	jsonData, err := json.Marshal(data)
	if err != nil {
		return "", fmt.Errorf("failed to marshal gist data: %w", err)
	}

	hash := sha256.Sum256(jsonData)
	return hex.EncodeToString(hash[:]), nil
}

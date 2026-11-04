package handlers

import (
	"strings"

	"github.com/MohamedElashri/snipo/pkg/models"
)

// snippetSummary is an opt-in list projection. Full content remains available
// from the detail endpoint and the default list representation.
type snippetSummary struct {
	models.Snippet
	IsSummary       bool   `json:"is_summary"`
	Preview         string `json:"preview"`
	PreviewFilename string `json:"preview_filename,omitempty"`
}

func summarizeSnippets(snippets []models.Snippet, query string) []snippetSummary {
	summaries := make([]snippetSummary, 0, len(snippets))
	for _, snippet := range snippets {
		content, filename := snippet.Content, ""
		if len(snippet.Files) > 0 {
			content, filename = snippet.Files[0].Content, snippet.Files[0].Filename
		}
		needle := strings.ToLower(strings.TrimSpace(query))
		if needle != "" {
			for _, file := range snippet.Files {
				if strings.Contains(strings.ToLower(file.Content), needle) {
					content, filename = file.Content, file.Filename
					break
				}
			}
		}
		preview := snippetPreview(content, needle)
		// Copy the file slice so projection never modifies the full response.
		files := append([]models.SnippetFile(nil), snippet.Files...)
		for i := range files {
			files[i].Content = ""
		}
		snippet.Content = ""
		snippet.Files = files
		summaries = append(summaries, snippetSummary{Snippet: snippet, IsSummary: true, Preview: preview, PreviewFilename: filename})
	}
	return summaries
}

func snippetPreview(content, needle string) string {
	if needle != "" {
		// Locate the original line instead of slicing at a lowercased byte offset;
		// Unicode case conversion can change string length.
		offset := 0
		for line := range strings.SplitSeq(content, "\n") {
			if strings.Contains(strings.ToLower(line), needle) {
				content = content[offset:]
				break
			}
			offset += len(line) + 1
		}
	}
	lines := strings.SplitN(content, "\n", 3)
	if len(lines) > 2 {
		lines = lines[:2]
	}
	runes := []rune(strings.Join(lines, "\n"))
	if len(runes) > 500 {
		return string(runes[:500]) + "…"
	}
	return string(runes)
}

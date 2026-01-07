package services

import (
	"strings"
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
)

func TestCalculateSnippetChecksum(t *testing.T) {
	snippet := &models.Snippet{
		Title:       "Test Snippet",
		Description: "Test Description",
		IsPublic:    true,
		Files: []models.SnippetFile{
			{
				Filename: "test.go",
				Content:  "package main\n\nfunc main() {}",
				Language: "go",
			},
		},
	}

	checksum1, err := CalculateSnippetChecksum(snippet)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 == "" {
		t.Error("checksum should not be empty")
	}

	checksum2, err := CalculateSnippetChecksum(snippet)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("checksums should be identical for same snippet")
	}

	snippet.Title = "Modified Title"
	checksum3, err := CalculateSnippetChecksum(snippet)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 == checksum3 {
		t.Error("checksums should differ after modification")
	}
}

func TestCalculateGistChecksum(t *testing.T) {
	gist := &models.GistResponse{
		Description: "Test Gist",
		Public:      true,
		Files: map[string]models.GistFile{
			"test.go": {
				Content: "package main\n\nfunc main() {}",
			},
		},
	}

	checksum1, err := CalculateGistChecksum(gist)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 == "" {
		t.Error("checksum should not be empty")
	}

	checksum2, err := CalculateGistChecksum(gist)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 != checksum2 {
		t.Error("checksums should be identical for same gist")
	}

	gist.Description = "Modified Description"
	checksum3, err := CalculateGistChecksum(gist)
	if err != nil {
		t.Fatalf("failed to calculate checksum: %v", err)
	}

	if checksum1 == checksum3 {
		t.Error("checksums should differ after modification")
	}
}

func TestSnippetToGistRequest(t *testing.T) {
	t.Run("single file snippet", func(t *testing.T) {
		snippet := &models.Snippet{
			ID:          "snippet-123",
			Title:       "Test Snippet",
			Description: "Test Description",
			IsPublic:    true,
			IsFavorite:  true,
			IsArchived:  false,
			Files: []models.SnippetFile{
				{
					Filename: "test.go",
					Content:  "package main",
					Language: "go",
				},
			},
			Tags:    []models.Tag{{Name: "test"}},
			Folders: []models.Folder{{Name: "work"}},
		}

		req, err := SnippetToGistRequest(snippet)
		if err != nil {
			t.Fatalf("failed to convert snippet: %v", err)
		}

		// Description should contain title and embedded metadata
		if !strings.Contains(req.Description, "Test Snippet") {
			t.Errorf("expected description to contain 'Test Snippet', got '%s'", req.Description)
		}

		if !strings.Contains(req.Description, "[snipo:") {
			t.Errorf("expected description to contain embedded metadata, got '%s'", req.Description)
		}

		if !req.Public {
			t.Error("expected public to be true")
		}

		if len(req.Files) != 1 {
			t.Errorf("expected 1 file (metadata embedded in description), got %d", len(req.Files))
		}

		if _, ok := req.Files["test.go"]; !ok {
			t.Error("expected test.go file")
		}
	})

	t.Run("legacy content snippet", func(t *testing.T) {
		snippet := &models.Snippet{
			ID:       "snippet-456",
			Title:    "Legacy Snippet",
			Content:  "print('hello')",
			Language: "python",
			IsPublic: false,
		}

		req, err := SnippetToGistRequest(snippet)
		if err != nil {
			t.Fatalf("failed to convert snippet: %v", err)
		}

		if len(req.Files) != 1 {
			t.Errorf("expected 1 file (metadata embedded in description), got %d", len(req.Files))
		}

		if !strings.Contains(req.Description, "[snipo:") {
			t.Error("expected metadata embedded in description")
		}

		found := false
		for range req.Files {
			found = true
			break
		}
		if !found {
			t.Error("expected a content file")
		}
	})
}

func TestGistToSnippet(t *testing.T) {
	t.Run("gist with metadata", func(t *testing.T) {
		gist := &models.GistResponse{
			ID:          "gist-123",
			Description: "Test Gist\n[snipo:{\"version\":\"1.0\",\"snipo_id\":\"snippet-123\",\"is_favorite\":true,\"is_archived\":false}]",
			Public:      true,
			Files: map[string]models.GistFile{
				"test.go": {
					Content: "package main",
				},
			},
		}

		snippet, err := GistToSnippet(gist, nil)
		if err != nil {
			t.Fatalf("failed to convert gist: %v", err)
		}

		if snippet.Title != "Test Gist" {
			t.Errorf("expected title 'Test Gist', got '%s'", snippet.Title)
		}

		if !snippet.IsPublic {
			t.Error("expected public to be true")
		}

		if !snippet.IsFavorite {
			t.Error("expected favorite to be true from metadata")
		}

		if len(snippet.Files) != 1 {
			t.Errorf("expected 1 file (metadata excluded), got %d", len(snippet.Files))
		}

		if snippet.Files[0].Filename != "test.go" {
			t.Errorf("expected filename 'test.go', got '%s'", snippet.Files[0].Filename)
		}

		if snippet.Files[0].Language != "go" {
			t.Errorf("expected language 'go', got '%s'", snippet.Files[0].Language)
		}
	})

	t.Run("gist without metadata", func(t *testing.T) {
		gist := &models.GistResponse{
			ID:          "gist-456",
			Description: "Simple Gist",
			Public:      false,
			Files: map[string]models.GistFile{
				"script.py": {
					Content: "print('hello')",
				},
			},
		}

		snippet, err := GistToSnippet(gist, nil)
		if err != nil {
			t.Fatalf("failed to convert gist: %v", err)
		}

		if snippet.IsFavorite {
			t.Error("expected favorite to be false without metadata")
		}

		if snippet.IsArchived {
			t.Error("expected archived to be false without metadata")
		}
	})
}

func TestGetLanguageFromFilename(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"test.go", "go"},
		{"script.py", "python"},
		{"app.js", "javascript"},
		{"component.tsx", "plaintext"},
		{"style.css", "css"},
		{"README.md", "markdown"},
		{"config.yaml", "yaml"},
		{"data.json", "json"},
		{"noextension", "plaintext"},
	}

	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			result := getLanguageFromFilename(tt.filename)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestGetExtensionForLanguage(t *testing.T) {
	tests := []struct {
		language string
		expected string
	}{
		{"go", "go"},
		{"python", "py"},
		{"javascript", "js"},
		{"typescript", "ts"},
		{"java", "java"},
		{"unknown", "txt"},
		{"plaintext", "txt"},
	}

	for _, tt := range tests {
		t.Run(tt.language, func(t *testing.T) {
			result := getExtensionForLanguage(tt.language)
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

func TestDetectChangesLogic(t *testing.T) {
	t.Run("no changes", func(t *testing.T) {
		mapping := &models.SnippetGistMapping{
			SnipoChecksum: "abc123",
			GistChecksum:  "def456",
		}

		currentSnipoChecksum := "abc123"
		currentGistChecksum := "def456"

		snipoChanged := currentSnipoChecksum != mapping.SnipoChecksum
		gistChanged := currentGistChecksum != mapping.GistChecksum

		if snipoChanged || gistChanged {
			t.Error("expected no changes")
		}
	})

	t.Run("snipo changed", func(t *testing.T) {
		mapping := &models.SnippetGistMapping{
			SnipoChecksum: "abc123",
			GistChecksum:  "def456",
		}

		currentSnipoChecksum := "xyz789"
		currentGistChecksum := "def456"

		snipoChanged := currentSnipoChecksum != mapping.SnipoChecksum
		gistChanged := currentGistChecksum != mapping.GistChecksum

		if !snipoChanged || gistChanged {
			t.Error("expected only snipo to change")
		}
	})

	t.Run("both changed - conflict", func(t *testing.T) {
		mapping := &models.SnippetGistMapping{
			SnipoChecksum: "abc123",
			GistChecksum:  "def456",
		}

		currentSnipoChecksum := "xyz789"
		currentGistChecksum := "uvw012"

		snipoChanged := currentSnipoChecksum != mapping.SnipoChecksum
		gistChanged := currentGistChecksum != mapping.GistChecksum

		if !snipoChanged || !gistChanged {
			t.Error("expected both to change (conflict)")
		}
	})
}

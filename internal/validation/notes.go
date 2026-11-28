package validation

import (
	"strings"

	"github.com/MohamedElashri/snipo/pkg/models"
)

func ValidateItemInput(input *models.SnippetInput) ValidationErrors {
	switch input.Kind {
	case "", models.KindSnippet:
		return ValidateSnippetInput(input)
	case models.KindNote:
		if len(input.Files) > 1 {
			return ValidationErrors{{Field: "files", Message: "A note can have only one body"}}
		}
		if len(input.Files) == 1 {
			input.Content = input.Files[0].Content
			input.Language = input.Files[0].Language
		}
		input.Language = strings.ToLower(strings.TrimSpace(input.Language))
		if !models.IsNoteFormat(input.Language) {
			return ValidationErrors{{Field: "language", Message: "Choose Markdown or plain text for a note"}}
		}
		if len(input.Files) == 1 {
			input.Files[0].Language = input.Language
		}
		return validateSnippetInput(input, true)
	default:
		return ValidationErrors{{Field: "kind", Message: "Choose note or snippet"}}
	}
}

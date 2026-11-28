package services

import (
	"context"
	"errors"
	"strings"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func (s *SnippetService) ListNoteTargets(ctx context.Context, query string, page int, kinds ...string) (models.NoteTargets, error) {
	return s.repo.ListNoteTargets(ctx, strings.TrimSpace(query), page, 25, kinds...)
}

func (s *SnippetService) AppendNote(ctx context.Context, id string, input models.NoteAppendInput) error {
	if len(input.RequestID) > 100 || strings.TrimSpace(input.RequestID) != input.RequestID {
		return validation.ValidationErrors{{Field: "request_id", Message: "Invalid capture request ID"}}
	}
	if input.FileID < 0 {
		return validation.ValidationErrors{{Field: "file_id", Message: "Invalid file ID"}}
	}
	if strings.TrimSpace(input.Content) == "" {
		return validation.ValidationErrors{{Field: "content", Message: "Content is required"}}
	}
	if len(input.Content) > models.MaxNoteContentBytes {
		return repository.ErrNoteTooLarge
	}
	before, err := s.repo.AppendNote(ctx, id, input)
	if errors.Is(err, repository.ErrNotFound) {
		return ErrSnippetNotFound
	}
	if err != nil {
		return err
	}

	if before != nil {
		_ = s.saveHistory(ctx, before, "update")
	}
	return nil
}

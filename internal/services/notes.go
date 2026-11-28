package services

import (
	"context"
	"errors"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/validation"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func (s *SnippetService) ConvertKind(ctx context.Context, id, kind string) (*models.Snippet, error) {
	if kind != models.KindNote && kind != models.KindSnippet {
		return nil, validation.ValidationErrors{{Field: "kind", Message: "Choose note or snippet"}}
	}
	if err := s.repo.ConvertKind(ctx, id, kind); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSnippetNotFound
		}
		return nil, err
	}
	return s.GetByID(ctx, id)
}

func (s *SnippetService) updateNote(ctx context.Context, existing *models.Snippet, input *models.SnippetInput, setFolder bool) (*models.Snippet, error) {
	if errs := validation.ValidateItemInput(input); errs.HasErrors() {
		return nil, errs
	}
	if input.Revision == nil {
		input.Revision = &existing.Revision
	}
	if err := s.repo.UpdateNote(ctx, existing.ID, input, setFolder); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return nil, ErrSnippetNotFound
		}
		return nil, err
	}
	if err := s.saveHistory(ctx, existing, "update"); err != nil {
		s.logger.Warn("failed to save pre-update note to history")
	}
	if s.tagRepo != nil && input.Tags != nil {
		if err := s.tagRepo.SetSnippetTags(ctx, existing.ID, input.Tags); err != nil {
			s.logger.Warn("failed to update note tags")
		}
	}
	return s.GetByID(ctx, existing.ID)
}

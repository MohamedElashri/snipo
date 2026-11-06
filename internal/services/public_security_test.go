package services

import (
	"errors"
	"testing"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestTrashRevokesPublicAccess(t *testing.T) {
	db := testutil.TestDB(t)
	repo := repository.NewSnippetRepository(db)
	svc := NewSnippetService(repo, testutil.TestLogger())
	ctx := testutil.TestContext()
	snippet, err := svc.Create(ctx, &models.SnippetInput{Title: "Public revocation", Content: "private after trash", Language: "plaintext", IsPublic: true})
	if err != nil {
		t.Fatal(err)
	}
	if _, err = svc.GetByIDPublic(ctx, snippet.ID); err != nil {
		t.Fatal(err)
	}
	if err = repo.Delete(ctx, snippet.ID, false); err != nil {
		t.Fatal(err)
	}
	stored, err := repo.GetByID(ctx, snippet.ID)
	if err != nil || stored == nil || stored.DeletedAt == nil {
		t.Fatal("fixture not soft-deleted", err)
	}
	if _, err = svc.GetByIDPublic(ctx, snippet.ID); !errors.Is(err, ErrSnippetNotFound) {
		t.Fatalf("trashed snippet still public: %v", err)
	}
	if err = repo.Restore(ctx, snippet.ID); err != nil {
		t.Fatal(err)
	}
	if _, err = svc.GetByIDPublic(ctx, snippet.ID); err != nil {
		t.Fatal("restored public snippet unavailable", err)
	}
}

package repository

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestFolderUpdateRejectsDescendantParent(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := context.Background()
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatal(err)
	}
	child, err := repo.Create(ctx, &models.FolderInput{Name: "Child", ParentID: &parent.ID})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := repo.Update(ctx, parent.ID, &models.FolderInput{Name: "Parent", ParentID: &child.ID}); err == nil {
		t.Fatal("folder update created a parent cycle")
	}
	unchanged, err := repo.GetByID(ctx, parent.ID)
	if err != nil || unchanged.ParentID != nil {
		t.Fatalf("rejected update changed folder: %+v %v", unchanged, err)
	}
}

func TestExistingFolderCycleDoesNotLoop(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := context.Background()
	a, _ := repo.Create(ctx, &models.FolderInput{Name: "A"})
	b, _ := repo.Create(ctx, &models.FolderInput{Name: "B", ParentID: &a.ID})
	other, _ := repo.Create(ctx, &models.FolderInput{Name: "Other"})
	// Model data left behind by the original update route, without leaving a live hang.
	if _, err := db.ExecContext(ctx, "UPDATE folders SET parent_id=? WHERE id=?", b.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	bounded, cancel := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancel()
	_, err := repo.Move(bounded, other.ID, &a.ID)
	if err == nil || errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("cycle was not rejected before the deadline: %v", err)
	}
}

func TestConcurrentFolderUpdatesCannotCreateCycle(t *testing.T) {
	db := testutil.TestDB(t)
	db.SetMaxOpenConns(1)
	repo := NewFolderRepository(db)
	ctx := context.Background()
	a, err := repo.Create(ctx, &models.FolderInput{Name: "A"})
	if err != nil {
		t.Fatal(err)
	}
	b, err := repo.Create(ctx, &models.FolderInput{Name: "B"})
	if err != nil {
		t.Fatal(err)
	}
	start := make(chan struct{})
	results := make(chan error, 2)
	go func() {
		<-start
		_, err := repo.Update(ctx, a.ID, &models.FolderInput{Name: "A", ParentID: &b.ID})
		results <- err
	}()
	go func() { <-start; _, err := repo.Move(ctx, b.ID, &a.ID); results <- err }()
	close(start)
	success, rejected := 0, 0
	for i := 0; i < 2; i++ {
		err := <-results
		if err == nil {
			success++
		} else if errors.Is(err, ErrFolderCycle) {
			rejected++
		} else {
			t.Fatal(err)
		}
	}
	if success != 1 || rejected != 1 {
		t.Fatalf("success=%d rejected=%d", success, rejected)
	}
}

func TestFolderCycleCanBeRepairedByMovingToRoot(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := context.Background()
	a, _ := repo.Create(ctx, &models.FolderInput{Name: "A"})
	b, _ := repo.Create(ctx, &models.FolderInput{Name: "B", ParentID: &a.ID})
	if _, err := db.ExecContext(ctx, "UPDATE folders SET parent_id=? WHERE id=?", b.ID, a.ID); err != nil {
		t.Fatal(err)
	}
	repaired, err := repo.Move(ctx, a.ID, nil)
	if err != nil || repaired.ParentID != nil {
		t.Fatalf("cycle repair: %+v %v", repaired, err)
	}
}

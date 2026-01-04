package repository

import (
	"testing"

	"github.com/MohamedElashri/snipo/internal/models"
	"github.com/MohamedElashri/snipo/internal/testutil"
)

func TestFolderRepository_Create(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	input := &models.FolderInput{
		Name:      "Projects",
		Icon:      "folder",
		SortOrder: 1,
	}

	folder, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	if folder.ID == 0 {
		t.Error("expected folder ID to be set")
	}
	if folder.Name != input.Name {
		t.Errorf("expected name %q, got %q", input.Name, folder.Name)
	}
	if folder.Icon != input.Icon {
		t.Errorf("expected icon %q, got %q", input.Icon, folder.Icon)
	}
	if folder.ParentID != nil {
		t.Error("expected parent_id to be nil")
	}
}

func TestFolderRepository_Create_WithParent(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create parent folder
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("Create parent failed: %v", err)
	}

	// Create child folder
	child, err := repo.Create(ctx, &models.FolderInput{
		Name:     "Child",
		ParentID: &parent.ID,
	})
	if err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	if child.ParentID == nil {
		t.Fatal("expected parent_id to be set")
	}
	if *child.ParentID != parent.ID {
		t.Errorf("expected parent_id %d, got %d", parent.ID, *child.ParentID)
	}
}

func TestFolderRepository_GetByID(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	input := &models.FolderInput{Name: "Test Folder"}
	created, err := repo.Create(ctx, input)
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Get by ID
	folder, err := repo.GetByID(ctx, created.ID)
	if err != nil {
		t.Fatalf("GetByID failed: %v", err)
	}

	if folder.ID != created.ID {
		t.Errorf("expected ID %d, got %d", created.ID, folder.ID)
	}
	if folder.Name != input.Name {
		t.Errorf("expected name %q, got %q", input.Name, folder.Name)
	}
}

func TestFolderRepository_GetByID_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	_, err := repo.GetByID(ctx, 99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFolderRepository_List(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create folders
	names := []string{"Alpha", "Beta", "Gamma"}
	for i, name := range names {
		_, err := repo.Create(ctx, &models.FolderInput{
			Name:      name,
			SortOrder: i,
		})
		if err != nil {
			t.Fatalf("Create failed: %v", err)
		}
	}

	// List all
	folders, err := repo.List(ctx)
	if err != nil {
		t.Fatalf("List failed: %v", err)
	}

	if len(folders) != 3 {
		t.Errorf("expected 3 folders, got %d", len(folders))
	}

	// Should be sorted by sort_order then name
	if folders[0].Name != "Alpha" {
		t.Errorf("expected first folder to be 'Alpha', got %q", folders[0].Name)
	}
}

func TestFolderRepository_ListTree(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create parent
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("Create parent failed: %v", err)
	}

	// Create children
	_, err = repo.Create(ctx, &models.FolderInput{Name: "Child1", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("Create child1 failed: %v", err)
	}
	_, err = repo.Create(ctx, &models.FolderInput{Name: "Child2", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("Create child2 failed: %v", err)
	}

	// Create another root
	_, err = repo.Create(ctx, &models.FolderInput{Name: "Root2"})
	if err != nil {
		t.Fatalf("Create root2 failed: %v", err)
	}

	// Get tree
	tree, err := repo.ListTree(ctx)
	if err != nil {
		t.Fatalf("ListTree failed: %v", err)
	}

	// Should have 2 root folders
	if len(tree) != 2 {
		t.Errorf("expected 2 root folders, got %d", len(tree))
	}

	// Find Parent folder and check children
	var parentFolder *models.Folder
	for i := range tree {
		if tree[i].Name == "Parent" {
			parentFolder = &tree[i]
			break
		}
	}

	if parentFolder == nil {
		t.Fatal("Parent folder not found in tree")
		return
	}

	if len(parentFolder.Children) != 2 {
		t.Errorf("expected 2 children, got %d", len(parentFolder.Children))
	}
}

func TestFolderRepository_Update(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	created, err := repo.Create(ctx, &models.FolderInput{Name: "Original"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Update it
	updateInput := &models.FolderInput{
		Name:      "Updated",
		Icon:      "star",
		SortOrder: 5,
	}
	updated, err := repo.Update(ctx, created.ID, updateInput)
	if err != nil {
		t.Fatalf("Update failed: %v", err)
	}

	if updated.Name != updateInput.Name {
		t.Errorf("expected name %q, got %q", updateInput.Name, updated.Name)
	}
	if updated.Icon != updateInput.Icon {
		t.Errorf("expected icon %q, got %q", updateInput.Icon, updated.Icon)
	}
	if updated.SortOrder != updateInput.SortOrder {
		t.Errorf("expected sort_order %d, got %d", updateInput.SortOrder, updated.SortOrder)
	}
}

func TestFolderRepository_Update_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	_, err := repo.Update(ctx, 99999, &models.FolderInput{Name: "Test"})
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFolderRepository_Delete(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	created, err := repo.Create(ctx, &models.FolderInput{Name: "To Delete"})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}

	// Delete it
	err = repo.Delete(ctx, created.ID)
	if err != nil {
		t.Fatalf("Delete failed: %v", err)
	}

	// Verify it's gone
	_, err = repo.GetByID(ctx, created.ID)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound after delete, got %v", err)
	}
}

func TestFolderRepository_Delete_NotFound(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	err := repo.Delete(ctx, 99999)
	if err != ErrNotFound {
		t.Errorf("expected ErrNotFound, got %v", err)
	}
}

func TestFolderRepository_Move(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create folders
	folder1, err := repo.Create(ctx, &models.FolderInput{Name: "Folder1"})
	if err != nil {
		t.Fatalf("Create folder1 failed: %v", err)
	}
	folder2, err := repo.Create(ctx, &models.FolderInput{Name: "Folder2"})
	if err != nil {
		t.Fatalf("Create folder2 failed: %v", err)
	}

	// Move folder2 under folder1
	moved, err := repo.Move(ctx, folder2.ID, &folder1.ID)
	if err != nil {
		t.Fatalf("Move failed: %v", err)
	}

	if moved.ParentID == nil {
		t.Fatal("expected parent_id to be set after move")
	}
	if *moved.ParentID != folder1.ID {
		t.Errorf("expected parent_id %d, got %d", folder1.ID, *moved.ParentID)
	}
}

func TestFolderRepository_Move_ToRoot(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create parent and child
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("Create parent failed: %v", err)
	}
	child, err := repo.Create(ctx, &models.FolderInput{Name: "Child", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	// Move child to root
	moved, err := repo.Move(ctx, child.ID, nil)
	if err != nil {
		t.Fatalf("Move to root failed: %v", err)
	}

	if moved.ParentID != nil {
		t.Error("expected parent_id to be nil after move to root")
	}
}

func TestFolderRepository_Move_CircularReference(t *testing.T) {
	db := testutil.TestDB(t)
	repo := NewFolderRepository(db)
	ctx := testutil.TestContext()

	// Create parent and child
	parent, err := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	if err != nil {
		t.Fatalf("Create parent failed: %v", err)
	}
	child, err := repo.Create(ctx, &models.FolderInput{Name: "Child", ParentID: &parent.ID})
	if err != nil {
		t.Fatalf("Create child failed: %v", err)
	}

	// Try to move parent under child (circular reference)
	_, err = repo.Move(ctx, parent.ID, &child.ID)
	if err == nil {
		t.Error("expected error for circular reference")
	}
}

func TestFolderRepository_SetSnippetFolder(t *testing.T) {
	db := testutil.TestDB(t)
	folderRepo := NewFolderRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	folder, err := folderRepo.Create(ctx, &models.FolderInput{Name: "Test Folder"})
	if err != nil {
		t.Fatalf("Create folder failed: %v", err)
	}

	// Create a snippet
	snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Set folder
	err = folderRepo.SetSnippetFolder(ctx, snippet.ID, &folder.ID)
	if err != nil {
		t.Fatalf("SetSnippetFolder failed: %v", err)
	}

	// Get snippet folders
	folders, err := folderRepo.GetSnippetFolders(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetSnippetFolders failed: %v", err)
	}

	if len(folders) != 1 {
		t.Errorf("expected 1 folder, got %d", len(folders))
	}
	if folders[0].ID != folder.ID {
		t.Errorf("expected folder ID %d, got %d", folder.ID, folders[0].ID)
	}
}

func TestFolderRepository_SetSnippetFolder_Remove(t *testing.T) {
	db := testutil.TestDB(t)
	folderRepo := NewFolderRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	folder, err := folderRepo.Create(ctx, &models.FolderInput{Name: "Test Folder"})
	if err != nil {
		t.Fatalf("Create folder failed: %v", err)
	}

	// Create a snippet
	snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
		Title:    "Test Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}

	// Set folder
	err = folderRepo.SetSnippetFolder(ctx, snippet.ID, &folder.ID)
	if err != nil {
		t.Fatalf("SetSnippetFolder failed: %v", err)
	}

	// Remove folder (set to nil)
	err = folderRepo.SetSnippetFolder(ctx, snippet.ID, nil)
	if err != nil {
		t.Fatalf("SetSnippetFolder (remove) failed: %v", err)
	}

	// Verify no folders
	folders, err := folderRepo.GetSnippetFolders(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("GetSnippetFolders failed: %v", err)
	}

	if len(folders) != 0 {
		t.Errorf("expected 0 folders after remove, got %d", len(folders))
	}
}

func TestFolderRepository_GetFolderSnippetCount(t *testing.T) {
	db := testutil.TestDB(t)
	folderRepo := NewFolderRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	folder, err := folderRepo.Create(ctx, &models.FolderInput{Name: "Count Test"})
	if err != nil {
		t.Fatalf("Create folder failed: %v", err)
	}

	// Create snippets and assign to folder
	for i := 0; i < 5; i++ {
		snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
			Title:    "Snippet",
			Content:  "content",
			Language: "plaintext",
		})
		if err != nil {
			t.Fatalf("Create snippet failed: %v", err)
		}
		err = folderRepo.SetSnippetFolder(ctx, snippet.ID, &folder.ID)
		if err != nil {
			t.Fatalf("SetSnippetFolder failed: %v", err)
		}
	}

	// Get count
	count, err := folderRepo.GetFolderSnippetCount(ctx, folder.ID)
	if err != nil {
		t.Fatalf("GetFolderSnippetCount failed: %v", err)
	}

	if count != 5 {
		t.Errorf("expected count 5, got %d", count)
	}
}

func TestFolderRepository_GetFolderSnippetCount_Archived(t *testing.T) {
	db := testutil.TestDB(t)
	folderRepo := NewFolderRepository(db)
	snippetRepo := NewSnippetRepository(db)
	ctx := testutil.TestContext()

	// Create a folder
	folder, err := folderRepo.Create(ctx, &models.FolderInput{Name: "Archive Count Test"})
	if err != nil {
		t.Fatalf("Create folder failed: %v", err)
	}

	// Create a snippet and assign to folder
	snippet, err := snippetRepo.Create(ctx, &models.SnippetInput{
		Title:    "Snippet",
		Content:  "content",
		Language: "plaintext",
	})
	if err != nil {
		t.Fatalf("Create snippet failed: %v", err)
	}
	err = folderRepo.SetSnippetFolder(ctx, snippet.ID, &folder.ID)
	if err != nil {
		t.Fatalf("SetSnippetFolder failed: %v", err)
	}

	// Initial count should be 1
	count, err := folderRepo.GetFolderSnippetCount(ctx, folder.ID)
	if err != nil {
		t.Fatalf("GetFolderSnippetCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1, got %d", count)
	}

	// Archive the snippet
	_, err = snippetRepo.ToggleArchive(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}

	// Count should now be 0
	count, err = folderRepo.GetFolderSnippetCount(ctx, folder.ID)
	if err != nil {
		t.Fatalf("GetFolderSnippetCount failed: %v", err)
	}
	if count != 0 {
		t.Errorf("expected count 0 after archiving, got %d", count)
	}

	// Unarchive
	_, err = snippetRepo.ToggleArchive(ctx, snippet.ID)
	if err != nil {
		t.Fatalf("ToggleArchive failed: %v", err)
	}

	// Count should be 1 again
	count, err = folderRepo.GetFolderSnippetCount(ctx, folder.ID)
	if err != nil {
		t.Fatalf("GetFolderSnippetCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("expected count 1 after unarchiving, got %d", count)
	}
}

package handlers

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/MohamedElashri/snipo/internal/repository"
	"github.com/MohamedElashri/snipo/internal/testutil"
	"github.com/MohamedElashri/snipo/pkg/models"
)

func TestFolderUpdateRejectsCyclesOverHTTP(t *testing.T) {
	repo := repository.NewFolderRepository(testutil.TestDB(t))
	ctx := context.Background()
	parent, _ := repo.Create(ctx, &models.FolderInput{Name: "Parent"})
	child, _ := repo.Create(ctx, &models.FolderInput{Name: "Child", ParentID: &parent.ID})
	req := httptest.NewRequest(http.MethodPut, "/api/v1/folders/1", bytes.NewBufferString(fmt.Sprintf(`{"name":"Parent","parent_id":%d}`, child.ID)))
	req.Header.Set("Content-Type", "application/json")
	req = withURLParams(req, map[string]string{"id": fmt.Sprint(parent.ID)})
	response := httptest.NewRecorder()
	NewFolderHandler(repo).Update(response, req)
	if response.Code != http.StatusBadRequest {
		t.Fatalf("cycle accepted: %d %s", response.Code, response.Body.String())
	}
}

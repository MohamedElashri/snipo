package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestListPaginationAndFilters(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/snipo/api/v1/snippets" {
			t.Errorf("path = %s", r.URL.Path)
		}
		for key, want := range map[string]string{"page": "2", "limit": "20", "q": "hello & world", "tag_ids": "1,2", "folder_ids": "3,4"} {
			if got := r.URL.Query().Get(key); got != want {
				t.Errorf("%s = %q, want %q", key, got, want)
			}
		}
		if r.Header.Get("Authorization") != "Bearer token" {
			t.Error("missing token")
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":[{"id":"snippet-21","title":"Second page"}],"pagination":{"page":2,"limit":20,"total":160,"totalPages":8}}`)
	}))
	defer server.Close()
	client := NewClient(" "+server.URL+"/snipo/// ", "token")
	snippets, pagination, err := client.ListSnippets(2, 20, "hello & world", []int{1, 2}, []int{3, 4}, "", nil, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(snippets) != 1 || snippets[0].ID != "snippet-21" {
		t.Fatalf("snippets = %+v", snippets)
	}
	if pagination.TotalPages != 8 || pagination.Page != 2 {
		t.Fatalf("pagination = %+v", pagination)
	}
}

func TestRedirectNeverForwardsCredentials(t *testing.T) {
	hits := 0
	target := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits++
		_, _ = fmt.Fprint(w, `{"data":{"id":"unexpected"}}`)
	}))
	defer target.Close()
	origin := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, target.URL, http.StatusTemporaryRedirect)
	}))
	defer origin.Close()
	_, err := NewClient(origin.URL, "test-only-token").GetSnippet("example")
	if err == nil {
		t.Error("API redirect accepted")
	}
	if hits != 0 {
		t.Fatal("redirect target received an authenticated request")
	}
}

package services

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/MohamedElashri/snipo/pkg/models"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync/atomic"
	"testing"
)

type fixtureTransport func(*http.Request) (*http.Response, error)

func (f fixtureTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func fixtureGitHubClient(t *testing.T, handler http.Handler) *GitHubClient {
	t.Helper()
	server := httptest.NewServer(handler)
	t.Cleanup(server.Close)
	endpoint, _ := url.Parse(server.URL)
	client := NewGitHubClient("fixture-token")
	client.httpClient.Transport = fixtureTransport(func(r *http.Request) (*http.Response, error) {
		clone := r.Clone(r.Context())
		target := *r.URL
		target.Scheme = endpoint.Scheme
		target.Host = endpoint.Host
		clone.URL = &target
		return server.Client().Transport.RoundTrip(clone)
	})
	return client
}

func TestPrivateGistEditsNeverReachPublicGist(t *testing.T) {
	var uploads atomic.Int32
	client := fixtureGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPatch {
			uploads.Add(1)
		}
		// GitHub ignores the unsupported public:false update parameter.
		_ = json.NewEncoder(w).Encode(models.GistResponse{ID: "fixture", Public: true})
	}))
	_, err := client.UpdateGist(context.Background(), "fixture", &models.GistRequest{Public: false, Files: map[string]models.GistFile{"private.txt": {Content: "PRIVATE-EDIT-MARKER"}}})
	if !errors.Is(err, ErrGistVisibilityMismatch) {
		t.Errorf("expected privacy refusal, got: %v", err)
	}
	if n := uploads.Load(); n != 0 {
		t.Errorf("private request bodies uploaded to public gist: %d", n)
	}
}

func TestGistVisibilityChecksPreserveSafeUpdates(t *testing.T) {
	for _, tc := range []struct {
		name                          string
		requestedPublic, remotePublic bool
		readStatus                    int
		malformed                     bool
		wantUpload                    bool
	}{
		{name: "secret to secret", readStatus: 200, wantUpload: true},
		{name: "public to public", requestedPublic: true, remotePublic: true, readStatus: 200, wantUpload: true},
		{name: "visibility unavailable", readStatus: 403},
		{name: "gist deleted", readStatus: 404},
		{name: "invalid visibility response", readStatus: 200, malformed: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var uploads atomic.Int32
			client := fixtureGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method == http.MethodGet {
					w.WriteHeader(tc.readStatus)
					if tc.malformed {
						_, _ = w.Write([]byte("invalid json"))
						return
					}
				} else if r.Method == http.MethodPatch {
					uploads.Add(1)
				}
				_ = json.NewEncoder(w).Encode(models.GistResponse{ID: "fixture", Public: tc.remotePublic})
			}))
			_, err := client.UpdateGist(context.Background(), "fixture", &models.GistRequest{Public: tc.requestedPublic, Files: map[string]models.GistFile{"example.txt": {Content: "fixture"}}})
			if tc.wantUpload {
				if err != nil || uploads.Load() != 1 {
					t.Fatalf("safe update: %v, uploads=%d", err, uploads.Load())
				}
			} else if err == nil || uploads.Load() != 0 {
				t.Fatalf("must fail before upload: %v, uploads=%d", err, uploads.Load())
			}
			if tc.readStatus == 404 && !IsGistNotFound(err) {
				t.Errorf("deleted-gist classification lost: %v", err)
			}
		})
	}
}

func TestUnknownGistVisibilityFailsBeforeUpload(t *testing.T) {
	for _, response := range []string{`{"id":"fixture"}`, `{"id":"fixture","public":null}`, `{"id":"fixture","public":"false"}`} {
		var uploads atomic.Int32
		client := fixtureGitHubClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodPatch {
				uploads.Add(1)
			}
			_, _ = w.Write([]byte(response))
		}))
		_, err := client.UpdateGist(context.Background(), "fixture", &models.GistRequest{Public: false})
		if err == nil || uploads.Load() != 0 {
			t.Errorf("unknown visibility must not permit a private upload: %v, uploads=%d", err, uploads.Load())
		}
	}
}

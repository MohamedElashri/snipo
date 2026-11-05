package web

import (
	"bytes"
	"io/fs"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"testing/fstest"
)

func TestAssetFingerprintTracksContentAndPaths(t *testing.T) {
	tree := fstest.MapFS{"static/css/main.css": {Data: []byte("a")}, "static/js/main.js": {Data: []byte("b")}}
	original := fingerprintAssets(tree)
	if got := fingerprintAssets(tree); got != original {
		t.Fatal("fingerprint is not stable")
	}
	tree["static/css/main.css"].Data = []byte("changed")
	if fingerprintAssets(tree) == original {
		t.Fatal("CSS changes must change the namespace")
	}
	tree["static/css/main.css"].Data = []byte("a")
	tree["static/js/main.js"].Data = []byte("changed")
	if fingerprintAssets(tree) == original {
		t.Fatal("module changes must change the namespace")
	}
	tree["static/js/main.js"].Data = []byte("b")
	tree["static/js/renamed.js"] = tree["static/js/main.js"]
	delete(tree, "static/js/main.js")
	if fingerprintAssets(tree) == original {
		t.Fatal("renamed files must change the namespace")
	}
}

func TestVersionedAssets(t *testing.T) {
	for _, base := range []string{"", "/snipo"} {
		t.Run(base, func(t *testing.T) {
			handler := StaticHandler(base)
			prefix := base + "/static/v/" + assetVersion + "/"
			// These references model nested CSS, JS, font, and dynamic Ace requests.
			for _, asset := range []string{"css/main.css", "css/components/workspace.css", "js/main.js", "js/components/snippets/workspace-mixin.js", "vendor/css/fonts.css", "vendor/js/alpine.min.js", "vendor/js/ace/mode-python.js"} {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, httptest.NewRequest("GET", prefix+asset, nil))
				want, err := fs.ReadFile(staticFS, "static/"+asset)
				if err != nil {
					t.Fatal(err)
				}
				if recorder.Code != http.StatusOK || !bytes.Equal(recorder.Body.Bytes(), want) {
					t.Fatalf("wrong asset: %s (status %d)", asset, recorder.Code)
				}
				if !strings.Contains(recorder.Header().Get("Cache-Control"), "immutable") {
					t.Fatal("fingerprinted assets should be cacheable")
				}
			}
			for _, path := range []string{"v/stale/css/main.css", "v/" + assetVersion + "/missing.css", "v/" + assetVersion + "/../css/main.css"} {
				recorder := httptest.NewRecorder()
				handler.ServeHTTP(recorder, httptest.NewRequest("GET", base+"/static/"+path, nil))
				if recorder.Code != http.StatusNotFound || recorder.Header().Get("Cache-Control") != "no-store" {
					t.Fatalf("invalid asset must not be cached: %s", path)
				}
			}
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, httptest.NewRequest("GET", base+"/static/css/main.css", nil))
			if recorder.Code != http.StatusOK || recorder.Header().Get("Cache-Control") != "no-cache, must-revalidate" {
				t.Fatal("legacy URLs must remain available and revalidate")
			}
			parent, _ := url.Parse("https://example.test" + prefix + "js/components/snippets-app.js")
			child, _ := url.Parse("./snippets/workspace-mixin.js")
			if parent.ResolveReference(child).Path != prefix+"js/components/snippets/workspace-mixin.js" {
				t.Fatal("relative imports lost the namespace")
			}
		})
	}
}

package web

import (
	"crypto/sha256"
	"fmt"
	"io/fs"
	"net/http"
	"strings"
)

// One fingerprint for the embedded asset tree keeps relative CSS imports, module
// imports, fonts, and dynamically loaded editor files in the same namespace.
var assetVersion = fingerprintAssets(staticFS)

func fingerprintAssets(assets fs.FS) string {
	hash := sha256.New()
	err := fs.WalkDir(assets, "static", func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		content, err := fs.ReadFile(assets, name)
		if err != nil {
			return err
		}
		// Length framing makes names and contents unambiguous. WalkDir sorts paths.
		_, _ = fmt.Fprintf(hash, "%d:%s%d:", len(name), name, len(content))
		_, _ = hash.Write(content)
		return nil
	})
	if err != nil {
		panic(fmt.Sprintf("fingerprint embedded assets: %v", err))
	}
	return fmt.Sprintf("%x", hash.Sum(nil))
}

// StaticHandler serves fingerprinted assets alongside legacy unversioned URLs.
func StaticHandler(basePath string) http.Handler {
	staticContent, _ := fs.Sub(staticFS, "static")
	prefix := basePath + "/static/"
	versionedPrefix := prefix + "v/" + assetVersion + "/"
	files := http.FileServer(http.FS(staticContent))
	legacy := http.StripPrefix(prefix, files)
	versioned := http.StripPrefix(versionedPrefix, files)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-cache, must-revalidate")
		if strings.HasPrefix(r.URL.Path, prefix+"v/") {
			name := strings.TrimPrefix(r.URL.Path, versionedPrefix)
			// Never serve this build's files under a different build's fingerprint.
			if !strings.HasPrefix(r.URL.Path, versionedPrefix) || !fs.ValidPath(name) {
				w.Header().Set("Cache-Control", "no-store")
				http.NotFound(w, r)
				return
			}
			info, err := fs.Stat(staticContent, name)
			if err != nil || info.IsDir() {
				w.Header().Set("Cache-Control", "no-store")
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			versioned.ServeHTTP(w, r)
			return
		}
		legacy.ServeHTTP(w, r)
	})
}

package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestEmbeddedPagesRender(t *testing.T) {
	for _, page := range []string{"index.html", "public.html", "login.html"} {
		t.Run(page, func(t *testing.T) {
			recorder := httptest.NewRecorder()
			handler := &Handler{}
			handler.render(recorder, "layout.html", page, PageData{Title: "Test", BasePath: "/snipo"})
			if recorder.Code != 200 || strings.Contains(recorder.Body.String(), "Template execute error") {
				t.Fatalf("render failed: %s", recorder.Body.String())
			}
			if !strings.Contains(recorder.Body.String(), `/snipo/static/v/`+assetVersion+`/css/main.css`) {
				t.Fatal("assets must respect the deployment base path")
			}
			if page == "index.html" && !strings.Contains(recorder.Body.String(), `class="settings-page"`) {
				t.Fatal("workspace settings template was not composed")
			}
		})
	}
}

// Refresh must not expose password dialogs before Alpine initializes. Authentication
// is resolved server-side; the workspace should never contain the login form.
func TestWorkspaceInitialHTMLHidesPasswordPrompts(t *testing.T) {
	recorder := httptest.NewRecorder()
	(&Handler{}).render(recorder, "layout.html", "index.html", PageData{Title: "Snippets"})
	html := recorder.Body.String()
	cloak := strings.Index(html, "[x-cloak] { display: none !important; }")
	if cloak < 0 || cloak > strings.Index(html, "</head>") {
		t.Fatal("startup hiding rule must be in the head before body content")
	}
	if strings.Contains(html, `x-data="loginForm()"`) {
		t.Fatal("workspace must not render the login form")
	}
	for _, condition := range []string{"tokenPasswordAction !== null", "showDisableLoginPassword"} {
		expected := `x-show="` + condition + `" class="snipo-modal-overlay" x-cloak style="display: none; z-index: 1001;"`
		if !strings.Contains(html, expected) {
			t.Fatalf("password prompt %s must start hidden without JavaScript", condition)
		}
	}
}

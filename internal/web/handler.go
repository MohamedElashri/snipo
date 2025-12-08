package web

import (
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/MohamedElashri/snipo/internal/auth"
)

//go:embed templates/*.html
var templatesFS embed.FS

//go:embed static/css/*.css static/js/*.js static/vendor/css/*.css static/vendor/js/*.js static/vendor/fonts/*.woff2 static/*.ico static/*.png
var staticFS embed.FS

// Handler handles web page requests
type Handler struct {
	templates   *template.Template
	authService *auth.Service
}

// NewHandler creates a new web handler
func NewHandler(authService *auth.Service) (*Handler, error) {
	// Parse templates
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		templates:   tmpl,
		authService: authService,
	}, nil
}

// StaticHandler returns a handler for static files
func StaticHandler() http.Handler {
	staticContent, _ := fs.Sub(staticFS, "static")
	return http.StripPrefix("/static/", http.FileServer(http.FS(staticContent)))
}

// PageData holds data passed to templates
type PageData struct {
	Title string
}

// Index serves the main application page
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	// Check authentication
	token := auth.GetSessionFromRequest(r)
	if token == "" || !h.authService.ValidateSession(token) {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	data := PageData{Title: "Snippets"}
	h.render(w, "layout.html", "index.html", data)
}

// Login serves the login page
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// If already authenticated, redirect to home
	token := auth.GetSessionFromRequest(r)
	if token != "" && h.authService.ValidateSession(token) {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}

	data := PageData{Title: "Login"}
	h.render(w, "layout.html", "login.html", data)
}

// PublicSnippet serves the public snippet view page (no auth required)
func (h *Handler) PublicSnippet(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Shared Snippet"}
	h.render(w, "layout.html", "public.html", data)
}

// render renders a template with layout
func (h *Handler) render(w http.ResponseWriter, layout, content string, data interface{}) {
	// Create a new template that combines layout and content
	tmpl, err := template.ParseFS(templatesFS,
		filepath.Join("templates", layout),
		filepath.Join("templates", content),
	)
	if err != nil {
		http.Error(w, "Template parse error: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, layout, data); err != nil {
		http.Error(w, "Template execute error: "+err.Error(), http.StatusInternalServerError)
	}
}

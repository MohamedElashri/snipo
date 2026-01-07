package web

import (
	"context"
	"embed"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"

	"github.com/MohamedElashri/snipo/internal/auth"
	"github.com/MohamedElashri/snipo/internal/repository"
)

//go:embed templates/*.html templates/components/*.html
var templatesFS embed.FS

//go:embed static/css/*.css static/css/components/*.css static/js/*.js static/js/modules/*.js static/js/stores/*.js static/js/components/*.js static/js/components/snippets/*.js static/js/utils/*.js static/vendor/css/*.css static/vendor/js/*.js static/vendor/js/ace/*.js static/vendor/fonts/*.woff2 static/*.ico static/*.png
var staticFS embed.FS

// Handler handles web page requests
type Handler struct {
	templates    *template.Template
	authService  *auth.Service
	settingsRepo *repository.SettingsRepository
	demoMode     bool
	basePath     string
}

// NewHandler creates a new web handler
func NewHandler(authService *auth.Service, settingsRepo *repository.SettingsRepository) (*Handler, error) {
	// Parse templates including components
	tmpl, err := template.ParseFS(templatesFS, "templates/*.html", "templates/components/*.html")
	if err != nil {
		return nil, err
	}

	return &Handler{
		templates:    tmpl,
		authService:  authService,
		settingsRepo: settingsRepo,
		demoMode:     false,
		basePath:     "",
	}, nil
}

// WithDemoMode sets the demo mode flag
func (h *Handler) WithDemoMode(enabled bool) *Handler {
	h.demoMode = enabled
	return h
}

// WithBasePath sets the base path for reverse proxy
func (h *Handler) WithBasePath(basePath string) *Handler {
	h.basePath = basePath
	return h
}

// StaticHandler returns a handler for static files
func StaticHandler(basePath string) http.Handler {
	staticContent, _ := fs.Sub(staticFS, "static")
	prefix := basePath + "/static/"
	return http.StripPrefix(prefix, http.FileServer(http.FS(staticContent)))
}

// PageData holds data passed to templates
type PageData struct {
	Title    string
	DemoMode bool
	BasePath string
}

// Index serves the main application page
func (h *Handler) Index(w http.ResponseWriter, r *http.Request) {
	// Skip authentication check if auth is completely disabled
	if h.authService.IsAuthDisabled() {
		data := PageData{Title: "Snippets", DemoMode: h.demoMode, BasePath: h.basePath}
		h.render(w, "layout.html", "index.html", data)
		return
	}

	// Check if login is disabled in settings (but keep password for admin operations)
	ctx := context.Background()
	settings, err := h.settingsRepo.Get(ctx)
	if err == nil && settings.DisableLogin {
		// Login is disabled via settings - allow access without session
		data := PageData{Title: "Snippets", DemoMode: h.demoMode, BasePath: h.basePath}
		h.render(w, "layout.html", "index.html", data)
		return
	}

	// Normal authentication flow: require session
	token := auth.GetSessionFromRequest(r)
	if token == "" || !h.authService.ValidateSession(token) {
		http.Redirect(w, r, h.basePath+"/login", http.StatusSeeOther)
		return
	}

	data := PageData{Title: "Snippets", DemoMode: h.demoMode, BasePath: h.basePath}
	h.render(w, "layout.html", "index.html", data)
}

// Login serves the login page
func (h *Handler) Login(w http.ResponseWriter, r *http.Request) {
	// If auth is completely disabled, redirect to home
	if h.authService.IsAuthDisabled() {
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	// Check if login is disabled in settings (but keep password for admin operations)
	ctx := context.Background()
	settings, err := h.settingsRepo.Get(ctx)
	if err == nil && settings.DisableLogin {
		// Login is disabled via settings - redirect to home
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	// If already authenticated, redirect to home
	token := auth.GetSessionFromRequest(r)
	if token != "" && h.authService.ValidateSession(token) {
		http.Redirect(w, r, h.basePath+"/", http.StatusSeeOther)
		return
	}

	data := PageData{Title: "Login", DemoMode: h.demoMode, BasePath: h.basePath}
	h.render(w, "layout.html", "login.html", data)
}

// PublicSnippet serves the public snippet view page (no auth required)
func (h *Handler) PublicSnippet(w http.ResponseWriter, r *http.Request) {
	data := PageData{Title: "Shared Snippet", DemoMode: h.demoMode, BasePath: h.basePath}
	h.render(w, "layout.html", "public.html", data)
}

// render renders a template with layout
func (h *Handler) render(w http.ResponseWriter, layout, content string, data interface{}) {
	// Create a new template that combines layout, content, and components
	tmpl, err := template.ParseFS(templatesFS,
		filepath.Join("templates", layout),
		filepath.Join("templates", content),
		"templates/components/*.html",
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

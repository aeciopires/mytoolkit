// Package web serves the server-rendered UI: the homepage tool grid and one
// page per tool, sharing a common layout, theme, and frontend JS/CSS.
package web

import (
	"embed"
	"encoding/json"
	"html/template"
	"io/fs"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/go-chi/chi/v5"

	"github.com/aeciopires/mytoolkit/internal/registry"
)

//go:embed templates
var templatesFS embed.FS

//go:embed static
var staticFS embed.FS

var pages map[string]*template.Template

func init() {
	pages = map[string]*template.Template{}

	pages["index"] = template.Must(template.New("layout").ParseFS(templatesFS,
		"templates/layout.html",
		"templates/partials/tool-panel.html",
		"templates/index.html",
	))

	toolFiles, err := fs.Glob(templatesFS, "templates/tools/*.html")
	if err != nil {
		panic(err)
	}
	for _, f := range toolFiles {
		slug := strings.TrimSuffix(filepath.Base(f), ".html")
		pages[slug] = template.Must(template.New("layout").ParseFS(templatesFS,
			"templates/layout.html",
			"templates/partials/tool-panel.html",
			f,
		))
	}
}

// Register mounts the web UI routes and static asset server onto r.
func Register(r chi.Router) {
	sub, err := fs.Sub(staticFS, "static")
	if err != nil {
		panic(err)
	}
	r.Handle("/static/*", http.StripPrefix("/static/", http.FileServer(http.FS(sub))))
	r.Get("/", indexHandler)
	r.Get("/tools/{slug}", toolPageHandler)
}

// toolsJSON is precomputed once at startup: registry.Tools is static, so
// there's no need to re-marshal it on every request. Embedded verbatim into
// each page as window.MYTOOLKIT_TOOLS for the client-side search bar, which
// indexes each tool's name and description (the same content shown in its
// homepage card / tool-page hero card) without a network round-trip.
var toolsJSON template.JS

func init() {
	b, err := json.Marshal(registry.All())
	if err != nil {
		panic(err)
	}
	toolsJSON = template.JS(b)
}

func indexHandler(w http.ResponseWriter, r *http.Request) {
	data := map[string]any{"Title": "Home", "Tools": registry.All(), "ToolsJSON": toolsJSON, "ActiveSlug": ""}
	renderPage(w, "index", data)
}

func toolPageHandler(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	t, ok := registry.BySlug(slug)
	if !ok {
		http.NotFound(w, r)
		return
	}
	data := map[string]any{"Title": t.Name, "Tools": registry.All(), "Tool": t, "ToolsJSON": toolsJSON, "ActiveSlug": t.Slug}
	renderPage(w, slug, data)
}

func renderPage(w http.ResponseWriter, name string, data any) {
	tmpl, ok := pages[name]
	if !ok {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if err := tmpl.ExecuteTemplate(w, "layout", data); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

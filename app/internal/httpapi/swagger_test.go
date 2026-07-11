package httpapi

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"

	_ "github.com/aeciopires/mytoolkit/docs" // registers the generated spec, same as cmd/mytoolkit/main.go
	"github.com/aeciopires/mytoolkit/internal/registry"
)

// TestSwaggerSpecCoversEveryTool guards against the annotation seam this
// app relies on: every /api/v1/tools/{slug} route swag documents is a
// literal path written by hand (see .skills/swagger/SKILL.md) since swag
// can't introspect chi's single templated {slug} route. If a tool is added
// to the registry without also adding its handler's @Router annotation,
// nothing else in the build catches that — this test does.
func TestSwaggerSpecCoversEveryTool(t *testing.T) {
	r := chi.NewRouter()
	Register(r, ToolHandlers{})

	req := httptest.NewRequest(http.MethodGet, "/swagger/doc.json", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /swagger/doc.json = %d, want 200", w.Code)
	}

	var spec struct {
		Info struct {
			Title   string `json:"title"`
			Version string `json:"version"`
		} `json:"info"`
		Paths map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(w.Body.Bytes(), &spec); err != nil {
		t.Fatalf("doc.json is not valid JSON: %v", err)
	}

	if spec.Info.Title == "" {
		t.Error("spec.info.title is empty — general API annotations missing from cmd/mytoolkit/main.go?")
	}

	for _, sysPath := range []string{"/healthz", "/readyz", "/api/v1/tools", "/api/v1/metrics/ranking"} {
		if _, ok := spec.Paths[sysPath]; !ok {
			t.Errorf("spec is missing system path %q", sysPath)
		}
	}

	for _, tool := range registry.All() {
		path := "/api/v1/tools/" + tool.Slug
		if _, ok := spec.Paths[path]; !ok {
			t.Errorf("spec is missing %q for tool %q — add a @Router annotation on its handler and re-run `make swagger-gen`", path, tool.Slug)
		}
	}
}

// TestSwaggerUIServesIndex is a light smoke test that the UI itself (not
// just the JSON spec) is actually mounted and reachable.
func TestSwaggerUIServesIndex(t *testing.T) {
	r := chi.NewRouter()
	Register(r, ToolHandlers{})

	req := httptest.NewRequest(http.MethodGet, "/swagger/index.html", nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("GET /swagger/index.html = %d, want 200", w.Code)
	}
}

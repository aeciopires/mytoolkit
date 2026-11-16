package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	httpSwagger "github.com/swaggo/http-swagger"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	mw "github.com/aeciopires/mytoolkit/internal/httpapi/middleware"
	"github.com/aeciopires/mytoolkit/internal/metrics"
	"github.com/aeciopires/mytoolkit/internal/registry"
	"github.com/aeciopires/mytoolkit/internal/response"
)

// ToolHandlers maps a tool slug to its REST handler.
type ToolHandlers map[string]http.HandlerFunc

// Register mounts the health, metrics, and /api/v1 routes onto r.
func Register(r chi.Router, handlers ToolHandlers) {
	r.Get("/healthz", healthHandler)
	r.Get("/readyz", readyHandler)
	r.Handle("/metrics", metrics.Handler())
	r.Get("/swagger/*", httpSwagger.WrapHandler)

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/tools", listToolsHandler)
		api.Get("/metrics/ranking", rankingHandler)

		api.Route("/tools/{slug}", func(tr chi.Router) {
			tr.Use(chimw.RequestID)
			tr.Use(mw.ToolSpan)
			tr.Use(mw.Logging)
			tr.Use(mw.Metrics)
			tr.Use(limitBody)
			tr.Post("/", dispatchTool(handlers))
		})
	})
}

// MaxRequestBodyBytes caps a tool request body. Without it any client
// could stream an arbitrarily large body into json.Decoder and exhaust
// memory; 10 MiB is far above any realistic tool input (a pasted
// manifest, JSON document, or key).
const MaxRequestBodyBytes = 10 << 20

func limitBody(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, MaxRequestBodyBytes)
		next.ServeHTTP(w, r)
	})
}

func dispatchTool(handlers ToolHandlers) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		h, ok := handlers[slug]
		if !ok {
			response.WriteError(w, apperr.New(http.StatusNotFound, "UNKNOWN_TOOL", "unknown tool: "+slug))
			return
		}
		h(w, r)
	}
}

// listToolsHandler godoc
// @Summary List all tools
// @Description Returns metadata (slug, name, emoji, description, client_side) for every tool — the same data backing the web UI's navigation drawer, search bar, and homepage grid.
// @Tags system
// @Produce json
// @Success 200 {object} object{tools=[]registry.Tool}
// @Router /api/v1/tools [get]
func listToolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tools": registry.All()})
}

// rankingHandler godoc
// @Summary Tool usage ranking
// @Description Returns every tool that has had at least one successful REST/web invocation since the process started, ranked by usage count descending. In-memory only — resets on restart; does not reflect CLI usage. Backed by the same counter as the mytoolkit_tool_usage_total Prometheus metric.
// @Tags system
// @Produce json
// @Success 200 {object} object{ranking=[]metrics.RankEntry}
// @Router /api/v1/metrics/ranking [get]
func rankingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ranking": metrics.Ranking()})
}

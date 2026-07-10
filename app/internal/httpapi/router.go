package httpapi

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/prometheus/client_golang/prometheus/promhttp"

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
	r.Handle("/metrics", promhttp.Handler())

	r.Route("/api/v1", func(api chi.Router) {
		api.Get("/tools", listToolsHandler)
		api.Get("/metrics/ranking", rankingHandler)

		api.Route("/tools/{slug}", func(tr chi.Router) {
			tr.Use(chimw.RequestID)
			tr.Use(mw.Logging)
			tr.Use(mw.Metrics)
			tr.Post("/", dispatchTool(handlers))
		})
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

func listToolsHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"tools": registry.All()})
}

func rankingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"ranking": metrics.Ranking()})
}

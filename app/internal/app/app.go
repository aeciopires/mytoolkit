// Package app wires the shared router, middleware, and template renderer
// into a single http.Handler.
package app

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/aeciopires/mytoolkit/internal/httpapi"
	mw "github.com/aeciopires/mytoolkit/internal/httpapi/middleware"
	"github.com/aeciopires/mytoolkit/internal/web"
)

// New builds the top-level HTTP handler: health/metrics/API routes plus the
// server-rendered web UI.
func New(toolHandlers httpapi.ToolHandlers) http.Handler {
	r := chi.NewRouter()
	// Tracing first, so a recovered panic's 500 is recorded on the request
	// span and the panic log line carries its trace_id.
	r.Use(mw.Tracing)
	r.Use(mw.Recover)

	httpapi.Register(r, toolHandlers)
	web.Register(r)

	return r
}

package httpapi

import (
	"net/http"

	"github.com/aeciopires/mytoolkit/internal/health"
)

// healthHandler godoc
// @Summary Liveness probe
// @Description Returns 200 while the process is up. Used by Docker (`mytoolkit healthcheck`) and the Helm chart's Kubernetes liveness probe; not traced and excluded from request metrics. The streamable-HTTP MCP server (`mytoolkit mcp --transport http`) serves the same endpoint on its own port.
// @Tags system
// @Produce json
// @Success 200 {object} health.Status "{\"status\": \"ok\"}"
// @Router /healthz [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	health.LiveHandler(w, r)
}

// readyHandler godoc
// @Summary Readiness probe
// @Description Returns 200 {"status":"ready"} while the process accepts traffic, and 503 {"status":"shutting_down"} once graceful shutdown has started (SIGINT/SIGTERM), so Kubernetes removes the pod from Service endpoints while in-flight requests drain. Used by the Helm chart's readiness probe; not traced and excluded from request metrics.
// @Tags system
// @Produce json
// @Success 200 {object} health.Status "{\"status\": \"ready\"}"
// @Failure 503 {object} health.Status "{\"status\": \"shutting_down\"}"
// @Router /readyz [get]
func readyHandler(w http.ResponseWriter, r *http.Request) {
	health.ReadyHandler(w, r)
}

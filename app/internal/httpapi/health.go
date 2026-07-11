package httpapi

import (
	"encoding/json"
	"net/http"
)

// healthHandler godoc
// @Summary Liveness probe
// @Description Always returns 200 while the process is up. Used by Docker/Kubernetes liveness checks; excluded from request metrics.
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string "{\"status\": \"ok\"}"
// @Router /healthz [get]
func healthHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ok"})
}

// readyHandler godoc
// @Summary Readiness probe
// @Description Always returns 200 once the process has started (no external dependencies to wait on). Used by Kubernetes readiness checks; excluded from request metrics.
// @Tags system
// @Produce json
// @Success 200 {object} map[string]string "{\"status\": \"ready\"}"
// @Router /readyz [get]
func readyHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]string{"status": "ready"})
}

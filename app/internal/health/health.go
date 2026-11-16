// Package health serves the liveness (/healthz) and readiness (/readyz)
// endpoints shared by the web server (`serve`) and the streamable-HTTP MCP
// server (`mcp --transport http`), which Docker (`mytoolkit healthcheck`)
// and the Helm chart's Kubernetes probes call.
//
// Liveness is always 200 while the process can answer HTTP at all.
// Readiness turns 503 as soon as graceful shutdown starts (SIGTERM), so
// Kubernetes takes the pod out of its Service endpoints while in-flight
// requests drain, instead of routing new requests to a closing server.
package health

import (
	"encoding/json"
	"net/http"
	"sync/atomic"
)

var shuttingDown atomic.Bool

// MarkShuttingDown flips readiness to 503 for the rest of the process.
func MarkShuttingDown() { shuttingDown.Store(true) }

// Ready reports whether the process accepts new traffic.
func Ready() bool { return !shuttingDown.Load() }

// Status is the JSON body of both endpoints.
type Status struct {
	Status string `json:"status" example:"ok"`
}

// LiveHandler answers the liveness probe: {"status":"ok"}.
func LiveHandler(w http.ResponseWriter, _ *http.Request) {
	write(w, http.StatusOK, "ok")
}

// ReadyHandler answers the readiness probe: {"status":"ready"}, or 503
// {"status":"shutting_down"} once graceful shutdown has begun.
func ReadyHandler(w http.ResponseWriter, _ *http.Request) {
	if !Ready() {
		write(w, http.StatusServiceUnavailable, "shutting_down")
		return
	}
	write(w, http.StatusOK, "ready")
}

// Register mounts /healthz and /readyz (GET and HEAD) on mux.
func Register(mux *http.ServeMux) {
	mux.HandleFunc("GET /healthz", LiveHandler)
	mux.HandleFunc("GET /readyz", ReadyHandler)
}

func write(w http.ResponseWriter, code int, status string) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(code)
	_ = json.NewEncoder(w).Encode(Status{Status: status})
}

package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
)

// Logging logs one structured JSON line per completed request, tagged with
// a request_id, the matched tool slug and, when tracing is enabled, the
// trace_id/span_id of the request's span.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slug := chi.URLParam(r, "slug")

		logCtx := log.With().Str("request_id", chimw.GetReqID(r.Context())).Str("tool", slug)
		// trace_id/span_id let Grafana jump from a Loki log line to its
		// trace in Tempo/Jaeger (and back) — present only when tracing is on.
		if traceID, spanID, ok := telemetry.IDs(r.Context()); ok {
			logCtx = logCtx.Str("trace_id", traceID).Str("span_id", spanID)
		}
		reqLogger := logCtx.Logger()
		ctx := reqLogger.WithContext(r.Context())

		sw := newStatusWriter(w)
		next.ServeHTTP(sw, r.WithContext(ctx))

		duration := time.Since(start)
		ev := reqLogger.Info()
		switch {
		case sw.status >= 500:
			ev = reqLogger.Error()
		case sw.status >= 400:
			ev = reqLogger.Warn()
		}
		ev.Str("method", r.Method).
			Str("path", r.URL.Path).
			Int("status", sw.status).
			Float64("duration_ms", float64(duration.Microseconds())/1000.0).
			Msg("request completed")
	})
}

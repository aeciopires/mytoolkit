package middleware

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/rs/zerolog/log"
)

// Logging logs one structured JSON line per completed request, tagged with
// a request_id and the matched tool slug.
func Logging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		slug := chi.URLParam(r, "slug")

		reqLogger := log.With().Str("request_id", chimw.GetReqID(r.Context())).Str("tool", slug).Logger()
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

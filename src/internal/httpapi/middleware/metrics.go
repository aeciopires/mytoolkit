package middleware

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"

	"github.com/aeciopires/mytoolkit/internal/metrics"
)

// Metrics records the shared request-count and duration Prometheus metrics,
// and the usage-ranking counter for the matched tool slug.
func Metrics(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		sw := newStatusWriter(w)
		next.ServeHTTP(sw, r)

		slug := chi.URLParam(r, "slug")
		if slug == "" {
			slug = "unknown"
		}
		duration := time.Since(start)
		metrics.RequestsTotal.WithLabelValues(slug, r.Method, strconv.Itoa(sw.status)).Inc()
		metrics.RequestDuration.WithLabelValues(slug, r.Method).Observe(duration.Seconds())
		if sw.status < 400 {
			metrics.RecordUsage(slug)
		}
	})
}

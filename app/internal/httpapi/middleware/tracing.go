package middleware

import (
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
)

// Tracing starts the OpenTelemetry server span for every request (an
// incoming W3C traceparent header is honored) and, once chi has routed the
// request, renames it to "METHOD /route/{pattern}" with the http.route
// attribute — low-cardinality span names, the same grouping Tempo's span
// metrics and the Grafana traces dashboard aggregate on.
//
// Prometheus scrapes, health probes and static assets are not traced:
// they would drown real traffic in noise. With tracing disabled (no
// OTEL_EXPORTER_OTLP_ENDPOINT) the global provider is a no-op and this
// costs next to nothing.
func Tracing(next http.Handler) http.Handler {
	withRoute := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		next.ServeHTTP(w, r)
		// otelhttp doesn't record http.route itself; chi only knows the
		// pattern once routing is done.
		if pattern := routePattern(r); pattern != "" {
			trace.SpanFromContext(r.Context()).SetAttributes(attribute.String("http.route", pattern))
		}
	})
	return otelhttp.NewHandler(withRoute, "http.server",
		otelhttp.WithFilter(traced),
		// Called once when the span starts (no route yet: just the method)
		// and again by otelhttp after the handler returns, once chi has
		// set r.Pattern and the matched route pattern.
		otelhttp.WithSpanNameFormatter(func(_ string, r *http.Request) string {
			if pattern := routePattern(r); pattern != "" {
				return r.Method + " " + pattern
			}
			return r.Method
		}),
	)
}

func routePattern(r *http.Request) string {
	if rctx := chi.RouteContext(r.Context()); rctx != nil {
		return rctx.RoutePattern()
	}
	return ""
}

func traced(r *http.Request) bool {
	switch p := r.URL.Path; {
	case p == "/metrics", p == "/healthz", p == "/readyz", p == "/favicon.ico",
		strings.HasPrefix(p, "/static/"):
		return false
	}
	return true
}

// ToolSpan wraps a tool invocation in a child span named "tool <slug>",
// tagged with mytoolkit.tool and the chi request_id, and marked as an
// error for 5xx responses. 4xx responses (invalid input) are client
// errors and stay unset, per the OpenTelemetry HTTP server conventions.
func ToolSpan(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		slug := chi.URLParam(r, "slug")
		ctx, span := telemetry.Tracer().Start(r.Context(), "tool "+slug,
			trace.WithAttributes(
				attribute.String("mytoolkit.tool", slug),
				attribute.String("mytoolkit.request_id", chimw.GetReqID(r.Context())),
			))
		defer span.End()

		sw := newStatusWriter(w)
		next.ServeHTTP(sw, r.WithContext(ctx))

		span.SetAttributes(attribute.Int("http.response.status_code", sw.status))
		if sw.status >= 500 {
			span.SetStatus(codes.Error, http.StatusText(sw.status))
		}
	})
}

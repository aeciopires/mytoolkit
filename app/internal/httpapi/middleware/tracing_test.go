package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
)

func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	prev := otel.GetTracerProvider()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	return rec
}

func newRouter(status int) http.Handler {
	r := chi.NewRouter()
	r.Use(Tracing)
	r.Get("/metrics", func(w http.ResponseWriter, _ *http.Request) {})
	r.With(ToolSpan, Logging).Post("/api/v1/tools/{slug}", func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(status)
	})
	return r
}

func TestTracingNamesSpansByRouteAndNestsToolSpan(t *testing.T) {
	rec := recordSpans(t)
	newRouter(http.StatusInternalServerError).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/tools/base64", nil))

	spans := rec.Ended()
	if len(spans) != 2 {
		t.Fatalf("got %d spans, want 2 (server + tool)", len(spans))
	}
	tool, server := spans[0], spans[1]
	if server.Name() != "POST /api/v1/tools/{slug}" {
		t.Errorf("server span name = %q", server.Name())
	}
	if !slices.Contains(server.Attributes(), attribute.String("http.route", "/api/v1/tools/{slug}")) {
		t.Errorf("server span lacks http.route: %v", server.Attributes())
	}
	if tool.Name() != "tool base64" {
		t.Errorf("tool span name = %q", tool.Name())
	}
	if tool.Parent().SpanID() != server.SpanContext().SpanID() {
		t.Error("tool span is not a child of the server span")
	}
	if tool.Status().Code != codes.Error {
		t.Errorf("5xx tool span status = %v, want Error", tool.Status().Code)
	}
}

func TestToolSpanLeaves4xxUnset(t *testing.T) {
	rec := recordSpans(t)
	newRouter(http.StatusBadRequest).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/tools/base64", nil))
	if got := rec.Ended()[0].Status().Code; got != codes.Unset {
		t.Errorf("4xx tool span status = %v, want Unset", got)
	}
}

func TestTracingSkipsScrapes(t *testing.T) {
	rec := recordSpans(t)
	newRouter(http.StatusOK).ServeHTTP(httptest.NewRecorder(), httptest.NewRequest(http.MethodGet, "/metrics", nil))
	if n := len(rec.Ended()); n != 0 {
		t.Errorf("got %d spans for /metrics, want 0", n)
	}
}

func TestTracingHonorsIncomingTraceparent(t *testing.T) {
	rec := recordSpans(t)
	prevProp := otel.GetTextMapPropagator()
	t.Cleanup(func() { otel.SetTextMapPropagator(prevProp) })
	otel.SetTextMapPropagator(propagatorForTest())

	req := httptest.NewRequest(http.MethodPost, "/api/v1/tools/base64", nil)
	req.Header.Set("traceparent", "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01")
	newRouter(http.StatusOK).ServeHTTP(httptest.NewRecorder(), req)

	server := rec.Ended()[1]
	if got := server.SpanContext().TraceID().String(); got != "4bf92f3577b34da6a3ce929d0e0e4736" {
		t.Errorf("trace id = %s, want the incoming one", got)
	}
}

func TestLoggingIncludesTraceIDs(t *testing.T) {
	rec := recordSpans(t)
	var buf bytes.Buffer
	prev := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = prev })

	newRouter(http.StatusOK).ServeHTTP(httptest.NewRecorder(),
		httptest.NewRequest(http.MethodPost, "/api/v1/tools/base64", nil))

	traceID := rec.Ended()[0].SpanContext().TraceID().String()
	if line := buf.String(); !strings.Contains(line, `"trace_id":"`+traceID+`"`) || !strings.Contains(line, `"span_id":"`) {
		t.Errorf("log line lacks trace_id %s / span_id: %s", traceID, line)
	}
}

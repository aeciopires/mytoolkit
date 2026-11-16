package metrics

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	dto "github.com/prometheus/client_model/go"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
)

func exemplarOf(t *testing.T, h prometheus.Histogram) *dto.Exemplar {
	t.Helper()
	var m dto.Metric
	if err := h.Write(&m); err != nil {
		t.Fatal(err)
	}
	for _, b := range m.GetHistogram().GetBucket() {
		if ex := b.GetExemplar(); ex != nil {
			return ex
		}
	}
	return nil
}

func TestObserveAttachesTraceExemplar(t *testing.T) {
	prev := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(prev) })
	otel.SetTracerProvider(sdktrace.NewTracerProvider())

	h := prometheus.NewHistogram(prometheus.HistogramOpts{Name: "test_seconds", Help: "x"})
	ctx, span := telemetry.Tracer().Start(t.Context(), "op")
	defer span.End()
	Observe(ctx, h, 0.2)

	ex := exemplarOf(t, h)
	if ex == nil {
		t.Fatal("no exemplar recorded")
	}
	want := span.SpanContext().TraceID().String()
	if got := ex.GetLabel()[0]; got.GetName() != "trace_id" || got.GetValue() != want {
		t.Errorf("exemplar label = %s=%s, want trace_id=%s", got.GetName(), got.GetValue(), want)
	}
}

func TestObserveWithoutSpanHasNoExemplar(t *testing.T) {
	h := prometheus.NewHistogram(prometheus.HistogramOpts{Name: "test_seconds", Help: "x"})
	Observe(t.Context(), h, 0.2)
	if ex := exemplarOf(t, h); ex != nil {
		t.Errorf("unexpected exemplar %v", ex)
	}
}

// Exemplars only exist in the OpenMetrics exposition format, so the
// scrape handler must offer it — promhttp.Handler() alone does not.
func TestHandlerNegotiatesOpenMetrics(t *testing.T) {
	req := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	req.Header.Set("Accept", "application/openmetrics-text;version=1.0.0")
	rec := httptest.NewRecorder()
	Handler().ServeHTTP(rec, req)
	if ct := rec.Header().Get("Content-Type"); !strings.HasPrefix(ct, "application/openmetrics-text") {
		t.Errorf("Content-Type = %q, want OpenMetrics", ct)
	}
	if !strings.Contains(rec.Body.String(), "# EOF") {
		t.Error("OpenMetrics body missing # EOF terminator")
	}
}

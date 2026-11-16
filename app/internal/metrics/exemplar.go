package metrics

import (
	"context"
	"net/http"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
)

// Observe records v on obs, attaching the current sampled trace ID as a
// Prometheus exemplar ({trace_id="..."}) when ctx carries one. Exemplars
// are what let a Grafana latency panel jump straight to the trace behind
// an outlier (Prometheus → Tempo). Without a sampled span (tracing
// disabled) it is a plain Observe.
func Observe(ctx context.Context, obs prometheus.Observer, v float64) {
	if traceID, ok := telemetry.SampledTraceID(ctx); ok {
		if eo, ok := obs.(prometheus.ExemplarObserver); ok {
			eo.ObserveWithExemplar(v, prometheus.Labels{"trace_id": traceID})
			return
		}
	}
	obs.Observe(v)
}

// Handler serves the default registry for Prometheus scrapes. Unlike
// promhttp.Handler() it enables the OpenMetrics exposition format when
// the scraper negotiates it — the only format that carries exemplars —
// while keeping the promhttp_metric_handler_* self-instrumentation the
// dashboard's "In-flight /metrics Scrapes" panel uses.
func Handler() http.Handler {
	return promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, promhttp.HandlerOpts{EnableOpenMetrics: true}),
	)
}

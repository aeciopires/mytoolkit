// Package telemetry wires OpenTelemetry distributed tracing for the
// long-running surfaces (`serve` and `mcp`).
//
// Tracing is opt-in and configured exclusively through the standard
// OpenTelemetry environment variables, so the same binary runs unchanged
// with or without a collector:
//
//   - OTEL_EXPORTER_OTLP_ENDPOINT (or OTEL_EXPORTER_OTLP_TRACES_ENDPOINT):
//     enables tracing; spans are exported over OTLP/HTTP (e.g.
//     http://alloy:4318). Unset means tracing is disabled and every span is
//     a zero-cost no-op.
//   - OTEL_SDK_DISABLED=true or OTEL_TRACES_EXPORTER=none: force-disable.
//   - OTEL_SERVICE_NAME / OTEL_RESOURCE_ATTRIBUTES: override the default
//     service.name (the caller's serviceName) and add resource attributes.
//   - OTEL_TRACES_SAMPLER / OTEL_TRACES_SAMPLER_ARG: sampling (default
//     parentbased_always_on).
//
// The W3C traceparent/baggage propagators are always installed, so an
// incoming trace context is honored even when export is disabled.
package telemetry

import (
	"context"
	"os"
	"strings"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

// InstrumentationName is the tracer name used by every MyToolkit span.
const InstrumentationName = "github.com/aeciopires/mytoolkit"

// Tracer returns the MyToolkit tracer from the global provider (a no-op
// provider until Setup installs a real one).
func Tracer() trace.Tracer {
	return otel.Tracer(InstrumentationName)
}

// Enabled reports whether the environment asks for trace export.
func Enabled() bool {
	if strings.EqualFold(os.Getenv("OTEL_SDK_DISABLED"), "true") ||
		strings.EqualFold(os.Getenv("OTEL_TRACES_EXPORTER"), "none") {
		return false
	}
	return os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT") != "" ||
		os.Getenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT") != ""
}

// Setup installs the global propagator and, when Enabled, an OTLP/HTTP
// tracer provider. The returned shutdown flushes buffered spans and must
// be called before the process exits; it is a no-op when tracing is off.
func Setup(ctx context.Context, serviceName, serviceVersion string) (shutdown func(context.Context) error, err error) {
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{}, propagation.Baggage{},
	))
	noop := func(context.Context) error { return nil }
	if !Enabled() {
		return noop, nil
	}

	exporter, err := otlptracehttp.New(ctx)
	if err != nil {
		return noop, err
	}
	// Detectors are merged in order, later ones winning: the caller's
	// defaults first, so OTEL_SERVICE_NAME/OTEL_RESOURCE_ATTRIBUTES
	// (WithFromEnv) can override them.
	res, err := resource.New(ctx,
		resource.WithAttributes(
			attribute.String("service.name", serviceName),
			attribute.String("service.version", serviceVersion),
		),
		resource.WithFromEnv(),
		resource.WithTelemetrySDK(),
		resource.WithHost(),
		resource.WithProcessRuntimeVersion(),
	)
	if err != nil {
		return noop, err
	}

	// No explicit sampler: the SDK then honors OTEL_TRACES_SAMPLER.
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	return tp.Shutdown, nil
}

// IDs returns the trace and span IDs of the span in ctx, or ok=false when
// there is no valid span (tracing disabled, or no span started).
func IDs(ctx context.Context) (traceID, spanID string, ok bool) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() {
		return "", "", false
	}
	return sc.TraceID().String(), sc.SpanID().String(), true
}

// SampledTraceID returns the trace ID of a sampled span in ctx, for use as
// a Prometheus exemplar label — only sampled traces are actually stored
// in Tempo/Jaeger, so an unsampled ID would be a dead link.
func SampledTraceID(ctx context.Context) (string, bool) {
	sc := trace.SpanContextFromContext(ctx)
	if !sc.IsValid() || !sc.IsSampled() {
		return "", false
	}
	return sc.TraceID().String(), true
}

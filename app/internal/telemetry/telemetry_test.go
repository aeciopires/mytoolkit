package telemetry

import (
	"testing"

	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestEnabled(t *testing.T) {
	cases := []struct {
		name string
		env  map[string]string
		want bool
	}{
		{"nothing set", nil, false},
		{"endpoint", map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://alloy:4318"}, true},
		{"traces endpoint", map[string]string{"OTEL_EXPORTER_OTLP_TRACES_ENDPOINT": "http://alloy:4318/v1/traces"}, true},
		{"sdk disabled", map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://x", "OTEL_SDK_DISABLED": "TRUE"}, false},
		{"exporter none", map[string]string{"OTEL_EXPORTER_OTLP_ENDPOINT": "http://x", "OTEL_TRACES_EXPORTER": "none"}, false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			for _, k := range []string{"OTEL_EXPORTER_OTLP_ENDPOINT", "OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "OTEL_SDK_DISABLED", "OTEL_TRACES_EXPORTER"} {
				t.Setenv(k, tc.env[k])
			}
			if got := Enabled(); got != tc.want {
				t.Errorf("Enabled() = %v, want %v", got, tc.want)
			}
		})
	}
}

func TestSetupDisabledIsNoop(t *testing.T) {
	t.Setenv("OTEL_EXPORTER_OTLP_ENDPOINT", "")
	t.Setenv("OTEL_EXPORTER_OTLP_TRACES_ENDPOINT", "")
	shutdown, err := Setup(t.Context(), "svc", "test")
	if err != nil {
		t.Fatalf("Setup: %v", err)
	}
	if err := shutdown(t.Context()); err != nil {
		t.Fatalf("shutdown: %v", err)
	}
	if _, _, ok := IDs(t.Context()); ok {
		t.Error("IDs reported a span on an empty context")
	}
}

func TestIDsAndSampledTraceID(t *testing.T) {
	prev := otel.GetTracerProvider()
	t.Cleanup(func() { otel.SetTracerProvider(prev) })

	for _, tc := range []struct {
		name        string
		sampler     sdktrace.Sampler
		wantSampled bool
	}{
		{"sampled", sdktrace.AlwaysSample(), true},
		{"not sampled", sdktrace.NeverSample(), false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSampler(tc.sampler)))
			ctx, span := Tracer().Start(t.Context(), "op")
			defer span.End()

			traceID, spanID, ok := IDs(ctx)
			if !ok || len(traceID) != 32 || len(spanID) != 16 {
				t.Fatalf("IDs = %q, %q, %v", traceID, spanID, ok)
			}
			sampledID, ok := SampledTraceID(ctx)
			if ok != tc.wantSampled {
				t.Fatalf("SampledTraceID ok = %v, want %v", ok, tc.wantSampled)
			}
			if ok && sampledID != traceID {
				t.Errorf("SampledTraceID = %q, want %q", sampledID, traceID)
			}
		})
	}
}

package mcp

import (
	"bytes"
	"context"
	"net/http"
	"strings"
	"testing"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/sdk/trace/tracetest"
	"go.opentelemetry.io/otel/trace"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

func recordSpans(t *testing.T) *tracetest.SpanRecorder {
	t.Helper()
	rec := tracetest.NewSpanRecorder()
	prevTP, prevProp := otel.GetTracerProvider(), otel.GetTextMapPropagator()
	otel.SetTracerProvider(sdktrace.NewTracerProvider(sdktrace.WithSpanProcessor(rec)))
	otel.SetTextMapPropagator(propagation.TraceContext{})
	t.Cleanup(func() {
		otel.SetTracerProvider(prevTP)
		otel.SetTextMapPropagator(prevProp)
	})
	return rec
}

func spanNamed(rec *tracetest.SpanRecorder, name string) sdktrace.ReadOnlySpan {
	for _, s := range rec.Ended() {
		if s.Name() == name {
			return s
		}
	}
	return nil
}

func TestTracingMiddlewareSpansAndLogsToolCall(t *testing.T) {
	rec := recordSpans(t)
	var buf bytes.Buffer
	prevLog := log.Logger
	log.Logger = zerolog.New(&buf)
	t.Cleanup(func() { log.Logger = prevLog })

	session, closeFn := connect(t)
	defer closeFn()

	const parent = "4bf92f3577b34da6a3ce929d0e0e4736"
	params := &sdkmcp.CallToolParams{Name: "base64", Arguments: map[string]any{"input": "hi"}}
	params.SetMeta(map[string]any{"traceparent": "00-" + parent + "-00f067aa0ba902b7-01"})
	if _, err := session.CallTool(t.Context(), params); err != nil {
		t.Fatalf("CallTool: %v", err)
	}

	span := spanNamed(rec, "mcp tools/call base64")
	if span == nil {
		t.Fatalf("no tools/call span; got %d spans", len(rec.Ended()))
	}
	if got := span.SpanContext().TraceID().String(); got != parent {
		t.Errorf("trace id = %s, want the _meta traceparent's %s", got, parent)
	}
	attrs := span.Attributes()
	for _, want := range []attribute.KeyValue{
		attribute.String("mcp.method.name", "tools/call"),
		attribute.String("gen_ai.tool.name", "base64"),
	} {
		found := false
		for _, a := range attrs {
			found = found || a == want
		}
		if !found {
			t.Errorf("missing attribute %v in %v", want, attrs)
		}
	}
	if line := buf.String(); !strings.Contains(line, `"trace_id":"`+parent+`"`) || !strings.Contains(line, `"tool":"base64"`) {
		t.Errorf("log line lacks trace_id/tool: %s", line)
	}
}

func TestTracingMiddlewareMarksToolErrors(t *testing.T) {
	rec := recordSpans(t)
	session, closeFn := connect(t)
	defer closeFn()

	res, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": "hi", "variant": "bogus"},
	})
	if err != nil || !res.IsError {
		t.Fatalf("want a tool error, got res=%+v err=%v", res, err)
	}
	span := spanNamed(rec, "mcp tools/call base64")
	if span == nil || span.Status().Code != codes.Error {
		t.Fatalf("tool-error span status = %+v, want Error", span)
	}
}

// The SDK runs handlers on a session-scoped context: without WithNewRoot
// every call of a session would land in one trace.
func TestTracingMiddlewareStartsOneTracePerCall(t *testing.T) {
	rec := recordSpans(t)
	session, closeFn := connect(t)
	defer closeFn()

	for range 2 {
		if _, err := session.CallTool(t.Context(), &sdkmcp.CallToolParams{
			Name: "base64", Arguments: map[string]any{"input": "hi"},
		}); err != nil {
			t.Fatalf("CallTool: %v", err)
		}
	}
	traces := map[string]bool{}
	for _, s := range rec.Ended() {
		if s.Name() == "mcp tools/call base64" {
			traces[s.SpanContext().TraceID().String()] = true
		}
	}
	if len(traces) != 2 {
		t.Errorf("2 calls produced %d distinct traces, want 2", len(traces))
	}
}

func TestRemoteParentPrecedence(t *testing.T) {
	recordSpans(t) // installs the TraceContext propagator
	const (
		metaTrace   = "4bf92f3577b34da6a3ce929d0e0e4736"
		headerTrace = "0af7651916cd43dd8448eb211c80319c"
	)
	header := http.Header{}
	header.Set("traceparent", "00-"+headerTrace+"-b7ad6b7169203331-01")

	withMeta := &sdkmcp.CallToolParamsRaw{Name: "base64"}
	withMeta.SetMeta(map[string]any{"traceparent": "00-" + metaTrace + "-00f067aa0ba902b7-01"})

	for _, tc := range []struct {
		name   string
		params *sdkmcp.CallToolParamsRaw
		header http.Header
		want   string
	}{
		{"meta wins over header", withMeta, header, metaTrace},
		{"header when no meta", &sdkmcp.CallToolParamsRaw{Name: "base64"}, header, headerTrace},
		{"nothing", &sdkmcp.CallToolParamsRaw{Name: "base64"}, nil, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			req := &sdkmcp.ServerRequest[*sdkmcp.CallToolParamsRaw]{Params: tc.params, Extra: &sdkmcp.RequestExtra{Header: tc.header}}
			sc := trace.SpanContextFromContext(remoteParent(t.Context(), req))
			got := ""
			if sc.IsValid() {
				got = sc.TraceID().String()
			}
			if got != tc.want {
				t.Errorf("parent trace = %q, want %q", got, tc.want)
			}
		})
	}
}

// Regression: notifications/initialized arrives with a typed-nil
// *InitializedParams; reading its _meta crashed the MCP server.
func TestRemoteParentTypedNilParams(t *testing.T) {
	req := &sdkmcp.ServerRequest[*sdkmcp.InitializedParams]{Params: nil}
	if sc := trace.SpanContextFromContext(remoteParent(t.Context(), req)); sc.IsValid() {
		t.Errorf("unexpected parent %v", sc)
	}
}

func TestTracingMiddlewareRecoversPanics(t *testing.T) {
	recordSpans(t)
	panicky := func(context.Context, string, sdkmcp.Request) (sdkmcp.Result, error) { panic("boom") }
	req := &sdkmcp.ServerRequest[*sdkmcp.CallToolParamsRaw]{Params: &sdkmcp.CallToolParamsRaw{Name: "base64"}}
	res, err := tracingMiddleware(panicky)(t.Context(), "tools/call", req)
	if err == nil || res != nil {
		t.Fatalf("got res=%v err=%v, want a recovered error", res, err)
	}
}

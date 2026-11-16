package mcp

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/rs/zerolog/log"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/telemetry"
)

// tracingMiddleware wraps every received JSON-RPC method in an
// OpenTelemetry span ("mcp tools/call <tool>", "mcp initialize", ...) and
// writes one structured log line per request carrying its trace_id, so
// MCP traffic shows up in Tempo/Jaeger and Loki exactly like REST traffic.
// It is registered before metricsMiddleware (outermost), so the duration
// histograms recorded there get a trace_id exemplar.
//
// Each JSON-RPC request starts its own trace unless the client propagates
// one: a W3C traceparent in params._meta (the MCP convention, and the only
// option over stdio) wins, then the traceparent header of the HTTP request
// that carried it (streamable-HTTP transport). Any span already in ctx is
// deliberately ignored: the SDK runs handlers on a session-scoped context,
// so inheriting it would put every call of a session into one giant trace.
func tracingMiddleware(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
	return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
		ctx = remoteParent(ctx, req)

		name := "mcp " + method
		attrs := []attribute.KeyValue{attribute.String("mcp.method.name", method)}
		var tool string
		if method == "tools/call" {
			tool = toolName(req)
			name += " " + tool
			attrs = append(attrs, attribute.String("gen_ai.tool.name", tool), attribute.String("mytoolkit.tool", tool))
		}
		opts := []trace.SpanStartOption{trace.WithSpanKind(trace.SpanKindServer), trace.WithAttributes(attrs...)}
		if !trace.SpanContextFromContext(ctx).IsRemote() {
			opts = append(opts, trace.WithNewRoot())
		}
		ctx, span := telemetry.Tracer().Start(ctx, name, opts...)
		defer span.End()

		start := time.Now()
		result, err := callRecovering(ctx, next, method, req)
		duration := time.Since(start)

		status := "success"
		switch {
		case err != nil:
			status = "error"
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
		case isToolError(result):
			status = "error"
			span.SetStatus(codes.Error, "tool returned an error result")
		}

		// Notifications and the discovery chatter are high-volume and
		// uninteresting; keep them at debug so info-level logs stay useful.
		ev := log.Debug()
		switch {
		case status == "error":
			ev = log.Warn()
		case method == "tools/call":
			ev = log.Info()
		}
		if traceID, spanID, ok := telemetry.IDs(ctx); ok {
			ev = ev.Str("trace_id", traceID).Str("span_id", spanID)
		}
		if tool != "" {
			ev = ev.Str("tool", tool)
		}
		ev.Str("method", method).
			Str("status", status).
			Float64("duration_ms", float64(duration.Microseconds())/1000.0).
			Msg("mcp request completed")

		return result, err
	}
}

// remoteParent returns ctx carrying the client's trace context from
// params._meta or, failing that, the HTTP request headers; ctx unchanged
// when the client sent none.
func remoteParent(ctx context.Context, req sdkmcp.Request) context.Context {
	prop := otel.GetTextMapPropagator()
	carrier := propagation.MapCarrier{}
	meta := metaOf(req)
	for _, k := range []string{"traceparent", "tracestate", "baggage"} {
		if v, ok := meta[k].(string); ok {
			carrier[k] = v
		}
	}
	if carrier["traceparent"] != "" {
		return prop.Extract(ctx, carrier)
	}
	if extra := req.GetExtra(); extra != nil && extra.Header != nil {
		return prop.Extract(ctx, propagation.HeaderCarrier(extra.Header))
	}
	return ctx
}

// metaOf returns the request's params._meta, or nil. Notifications sent
// without params (e.g. notifications/initialized) carry a typed-nil
// *XxxParams inside the non-nil Params interface, and calling GetMeta on
// it dereferences nil — that crashed the whole MCP server once.
func metaOf(req sdkmcp.Request) map[string]any {
	params := req.GetParams()
	if params == nil {
		return nil
	}
	if v := reflect.ValueOf(params); v.Kind() == reflect.Pointer && v.IsNil() {
		return nil
	}
	return params.GetMeta()
}

// callRecovering runs next, converting a panic into a JSON-RPC error
// (logged with its stack trace context) instead of letting it unwind
// through the SDK's handler goroutine — where nothing recovers it and the
// whole MCP server process dies, taking every other session with it.
func callRecovering(ctx context.Context, next sdkmcp.MethodHandler, method string, req sdkmcp.Request) (result sdkmcp.Result, err error) {
	defer func() {
		if rec := recover(); rec != nil {
			ev := log.Error().Interface("panic", rec).Str("method", method)
			if traceID, spanID, ok := telemetry.IDs(ctx); ok {
				ev = ev.Str("trace_id", traceID).Str("span_id", spanID)
			}
			ev.Msg("panic recovered in MCP handler")
			result, err = nil, fmt.Errorf("internal error handling %s", method)
		}
	}()
	return next(ctx, method, req)
}

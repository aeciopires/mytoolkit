package mcp

import (
	"context"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/metrics"
)

// metricsMiddleware records the shared MCP Prometheus metrics for every
// JSON-RPC method (initialize, tools/list, tools/call, ...), plus a
// per-tool breakdown for tools/call specifically. Registered once via
// server.AddReceivingMiddleware in NewServer, so it applies uniformly to
// both transports and every registered tool without editing each
// handle<Name> function individually.
func metricsMiddleware(next sdkmcp.MethodHandler) sdkmcp.MethodHandler {
	return func(ctx context.Context, method string, req sdkmcp.Request) (sdkmcp.Result, error) {
		start := time.Now()
		result, err := next(ctx, method, req)
		duration := time.Since(start).Seconds()

		status := "success"
		if err != nil || isToolError(result) {
			status = "error"
		}

		metrics.MCPRequestsTotal.WithLabelValues(method, status).Inc()
		metrics.Observe(ctx, metrics.MCPRequestDuration.WithLabelValues(method), duration)

		switch method {
		// A session starts with the legacy "initialize" handshake or, from
		// protocol 2026-07-28 (SEP-2575, go-sdk >= v1.8), the stateless
		// "server/discover" request that replaces it — count both.
		case "initialize", "server/discover":
			if err == nil {
				metrics.MCPSessionsTotal.Inc()
			}
		case "tools/call":
			tool := toolName(req)
			metrics.MCPToolCallsTotal.WithLabelValues(tool, status).Inc()
			metrics.Observe(ctx, metrics.MCPToolCallDuration.WithLabelValues(tool), duration)
		}

		return result, err
	}
}

func isToolError(result sdkmcp.Result) bool {
	res, ok := result.(*sdkmcp.CallToolResult)
	return ok && res.IsError
}

func toolName(req sdkmcp.Request) string {
	if params, ok := req.GetParams().(*sdkmcp.CallToolParamsRaw); ok && params.Name != "" {
		return params.Name
	}
	return "unknown"
}

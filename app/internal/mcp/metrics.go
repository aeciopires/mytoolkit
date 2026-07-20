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
		metrics.MCPRequestDuration.WithLabelValues(method).Observe(duration)

		switch method {
		case "initialize":
			if err == nil {
				metrics.MCPSessionsTotal.Inc()
			}
		case "tools/call":
			tool := toolName(req)
			metrics.MCPToolCallsTotal.WithLabelValues(tool, status).Inc()
			metrics.MCPToolCallDuration.WithLabelValues(tool).Observe(duration)
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

package mcp

import (
	"context"
	"testing"

	"github.com/prometheus/client_golang/prometheus/testutil"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/metrics"
)

// These tests assert deltas, not absolute values: metrics.MCP* are package
// (process) globals shared with every other test in this package via
// connect()'s server.Connect/client.Connect handshake, which itself
// triggers "initialize" — asserting an exact total would make tests order-
// dependent.

func TestMetricsMiddlewareRecordsSuccessfulToolCall(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()
	ctx := context.Background()

	before := testutil.ToFloat64(metrics.MCPToolCallsTotal.WithLabelValues("base64", "success"))

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": "hi"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected tool error: %+v", res.Content)
	}

	after := testutil.ToFloat64(metrics.MCPToolCallsTotal.WithLabelValues("base64", "success"))
	if after != before+1 {
		t.Errorf("mytoolkit_mcp_tool_calls_total{tool=base64,status=success}: got %v, want %v", after, before+1)
	}
}

func TestMetricsMiddlewareRecordsFailedToolCall(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()
	ctx := context.Background()

	before := testutil.ToFloat64(metrics.MCPToolCallsTotal.WithLabelValues("base64", "error"))

	res, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": "hi", "variant": "bogus"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected a tool error, got %+v", res)
	}

	after := testutil.ToFloat64(metrics.MCPToolCallsTotal.WithLabelValues("base64", "error"))
	if after != before+1 {
		t.Errorf("mytoolkit_mcp_tool_calls_total{tool=base64,status=error}: got %v, want %v", after, before+1)
	}
}

func TestMetricsMiddlewareRecordsSession(t *testing.T) {
	before := testutil.ToFloat64(metrics.MCPSessionsTotal)

	_, closeFn := connect(t)
	defer closeFn()

	after := testutil.ToFloat64(metrics.MCPSessionsTotal)
	if after != before+1 {
		t.Errorf("mytoolkit_mcp_sessions_total: got %v, want %v", after, before+1)
	}
}

func TestMetricsMiddlewareRecordsToolsListRequest(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()

	before := testutil.ToFloat64(metrics.MCPRequestsTotal.WithLabelValues("tools/list", "success"))

	if _, err := session.ListTools(context.Background(), nil); err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	after := testutil.ToFloat64(metrics.MCPRequestsTotal.WithLabelValues("tools/list", "success"))
	if after != before+1 {
		t.Errorf("mytoolkit_mcp_requests_total{method=tools/list,status=success}: got %v, want %v", after, before+1)
	}
}

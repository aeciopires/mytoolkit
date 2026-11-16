// Package metrics registers the Prometheus collectors shared by every tool
// and maintains an in-memory usage-ranking aggregator.
package metrics

import (
	"cmp"
	"slices"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/prometheus/client_golang/prometheus"
)

var (
	RequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mytoolkit_http_requests_total",
		Help: "Total number of HTTP requests processed, per tool.",
	}, []string{"tool", "method", "status"})

	RequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mytoolkit_http_request_duration_seconds",
		Help:    "HTTP request duration in seconds, per tool.",
		Buckets: prometheus.DefBuckets,
	}, []string{"tool", "method"})

	ToolUsageTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mytoolkit_tool_usage_total",
		Help: "Total number of successful tool invocations, per tool.",
	}, []string{"tool"})

	// MCP metrics are recorded by internal/mcp's AddReceivingMiddleware hook
	// (one middleware wraps every JSON-RPC method, both transports), kept
	// separate from ToolUsageTotal/RequestsTotal above: the MCP surface is a
	// distinct process/client population from the REST/web surface those
	// track (see PLAN_ARCHITECTURE.md's usage-ranking scope note), so
	// conflating them would misattribute usage between surfaces.
	MCPRequestsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mytoolkit_mcp_requests_total",
		Help: "Total number of MCP JSON-RPC requests processed, per method.",
	}, []string{"method", "status"})

	MCPRequestDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mytoolkit_mcp_request_duration_seconds",
		Help:    "MCP JSON-RPC request duration in seconds, per method.",
		Buckets: prometheus.DefBuckets,
	}, []string{"method"})

	MCPToolCallsTotal = prometheus.NewCounterVec(prometheus.CounterOpts{
		Name: "mytoolkit_mcp_tool_calls_total",
		Help: "Total number of MCP tools/call requests, per tool.",
	}, []string{"tool", "status"})

	MCPToolCallDuration = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name:    "mytoolkit_mcp_tool_call_duration_seconds",
		Help:    "MCP tools/call request duration in seconds, per tool.",
		Buckets: prometheus.DefBuckets,
	}, []string{"tool"})

	MCPSessionsTotal = prometheus.NewCounter(prometheus.CounterOpts{
		Name: "mytoolkit_mcp_sessions_total",
		Help: "Total number of MCP sessions established (successful initialize requests).",
	})
)

func init() {
	prometheus.MustRegister(
		RequestsTotal, RequestDuration, ToolUsageTotal,
		MCPRequestsTotal, MCPRequestDuration, MCPToolCallsTotal, MCPToolCallDuration, MCPSessionsTotal,
	)
}

var (
	usageMu     sync.Mutex
	usageCounts = map[string]*atomic.Int64{}
)

// RecordUsage increments the in-memory usage counter for tool, used to
// derive the /api/v1/metrics/ranking endpoint.
func RecordUsage(tool string) {
	usageMu.Lock()
	c, ok := usageCounts[tool]
	if !ok {
		c = new(atomic.Int64)
		usageCounts[tool] = c
	}
	usageMu.Unlock()
	c.Add(1)
	ToolUsageTotal.WithLabelValues(tool).Inc()
}

type RankEntry struct {
	Tool  string `json:"tool"`
	Count int64  `json:"count"`
	Rank  int    `json:"rank"`
}

// Ranking returns tools sorted by usage count, descending (ties broken by
// tool name). The map is read under usageMu: RecordUsage may insert a new
// tool concurrently, and an unsynchronized map iteration racing an insert
// is a fatal runtime error, not just a stale read.
func Ranking() []RankEntry {
	usageMu.Lock()
	entries := make([]RankEntry, 0, len(usageCounts))
	for tool, c := range usageCounts {
		entries = append(entries, RankEntry{Tool: tool, Count: c.Load()})
	}
	usageMu.Unlock()

	slices.SortFunc(entries, func(a, b RankEntry) int {
		return cmp.Or(cmp.Compare(b.Count, a.Count), strings.Compare(a.Tool, b.Tool))
	})
	for i := range entries {
		entries[i].Rank = i + 1
	}
	return entries
}

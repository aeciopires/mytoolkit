package cli

import (
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/config"
	"github.com/aeciopires/mytoolkit/internal/health"
	"github.com/aeciopires/mytoolkit/internal/logging"
	mcpserver "github.com/aeciopires/mytoolkit/internal/mcp"
	"github.com/aeciopires/mytoolkit/internal/metrics"
	"github.com/aeciopires/mytoolkit/internal/version"
)

var mcpFlags config.MCPFlags

var mcpCmd = &cobra.Command{
	Use:   "mcp",
	Short: "Start the MyToolkit MCP (Model Context Protocol) server",
	Long: "Exposes every MyToolkit tool as an MCP tool, reusing the same internal/tools/<name>\n" +
		"functions the web UI, REST API, and CLI subcommands already call. See mcp/README.md\n" +
		"for installation and client configuration.",
	RunE: runMCP,
}

func init() {
	rootCmd.AddCommand(mcpCmd)
	mcpCmd.Flags().StringVar(&mcpFlags.Transport, "transport", config.DefaultMCPTransport, "transport: stdio or http")
	mcpCmd.Flags().StringVar(&mcpFlags.Host, "host", config.DefaultHost, "interface to bind to (http transport only)")
	mcpCmd.Flags().StringVar(&mcpFlags.Port, "port", config.DefaultMCPPort, "TCP port to listen on (http transport only)")
	mcpCmd.Flags().StringVar(&mcpFlags.LogLevel, "log-level", config.DefaultLogLevel, "log level: debug, info, warn, error")
}

func runMCP(cmd *cobra.Command, args []string) error {
	mcpFlags.TransportChanged = cmd.Flags().Changed("transport")
	mcpFlags.HostChanged = cmd.Flags().Changed("host")
	mcpFlags.PortChanged = cmd.Flags().Changed("port")
	mcpFlags.LogChanged = cmd.Flags().Changed("log-level")
	cfg := config.LoadMCP(mcpFlags)

	// Logs always go to stderr, exactly like `serve` — for the stdio
	// transport this isn't just convention, it's a correctness
	// requirement: stdout carries nothing but JSON-RPC protocol bytes, and
	// any stray write (a log line, a fmt.Println) would corrupt the stream.
	logging.Setup(os.Stderr, cfg.LogLevel, "mytoolkit-mcp", version.Version)

	shutdownTracing, err := setupTracing(cmd.Context(), "mytoolkit-mcp")
	if err != nil {
		return logFatal(err, "tracing setup failed")
	}
	defer shutdownTracing()

	switch cfg.Transport {
	case "stdio":
		log.Info().Str("transport", "stdio").Msg("starting MyToolkit MCP server")
		server := mcpserver.NewServer(version.Version)
		if err := server.Run(cmd.Context(), &sdkmcp.StdioTransport{}); err != nil {
			return logFatal(err, "MCP stdio server failed")
		}
		return nil
	case "http":
		mcpHandler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
			return mcpserver.NewServer(version.Version)
		}, nil)
		mux := http.NewServeMux()
		// No otelhttp wrapper here: internal/mcp's tracingMiddleware opens
		// one span per JSON-RPC request (honoring a traceparent header or
		// params._meta), which is the unit that matters; an extra
		// per-HTTP-request span would only add empty "POST /" traces.
		mux.Handle("/", mcpHandler)
		// Prometheus metrics for this MCP process, same mytoolkit_mcp_*
		// collectors metricsMiddleware records — reachable here because,
		// unlike stdio, the http transport already has a listening port to
		// hang a second path off of. See mcp/README.md's Observability
		// section and observability/prometheus.yml's mytoolkit-mcp job.
		mux.Handle("/metrics", metrics.Handler())
		// Liveness/readiness for Docker (`mytoolkit healthcheck --port`)
		// and the Helm chart's httpGet probes — the MCP endpoint itself
		// only answers JSON-RPC POSTs, so it can't serve as a probe.
		health.Register(mux)
		log.Info().Str("transport", "http").Str("addr", cfg.Addr()).Msg("starting MyToolkit MCP server")
		if err := listenAndServe(cmd.Context(), cfg.Addr(), mux); err != nil {
			return logFatal(err, "MCP HTTP server failed")
		}
		return nil
	default:
		return logFatal(apperr.Newf(400, "INVALID_TRANSPORT", "transport must be one of: stdio, http (got %q)", cfg.Transport), "invalid MCP transport")
	}
}

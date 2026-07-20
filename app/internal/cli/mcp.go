package cli

import (
	"net/http"
	"os"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"github.com/aeciopires/mytoolkit/internal/apperr"
	"github.com/aeciopires/mytoolkit/internal/config"
	mcpserver "github.com/aeciopires/mytoolkit/internal/mcp"
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
	zerolog.SetGlobalLevel(parseLevel(cfg.LogLevel))
	log.Logger = zerolog.New(os.Stderr).With().Timestamp().Logger()

	switch cfg.Transport {
	case "stdio":
		log.Info().Str("transport", "stdio").Msg("starting MyToolkit MCP server")
		server := mcpserver.NewServer(version.Version)
		return server.Run(cmd.Context(), &sdkmcp.StdioTransport{})
	case "http":
		mcpHandler := sdkmcp.NewStreamableHTTPHandler(func(*http.Request) *sdkmcp.Server {
			return mcpserver.NewServer(version.Version)
		}, nil)
		mux := http.NewServeMux()
		mux.Handle("/", mcpHandler)
		// Prometheus metrics for this MCP process, same mytoolkit_mcp_*
		// collectors metricsMiddleware records — reachable here because,
		// unlike stdio, the http transport already has a listening port to
		// hang a second path off of. See mcp/README.md's Observability
		// section and observability/prometheus.yml's mytoolkit-mcp job.
		mux.Handle("/metrics", promhttp.Handler())
		log.Info().Str("transport", "http").Str("addr", cfg.Addr()).Msg("starting MyToolkit MCP server")
		return http.ListenAndServe(cfg.Addr(), mux)
	default:
		return apperr.Newf(400, "INVALID_TRANSPORT", "transport must be one of: stdio, http (got %q)", cfg.Transport)
	}
}

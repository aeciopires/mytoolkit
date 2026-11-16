package cli

import (
	"net/http"
	"os"

	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/app"
	"github.com/aeciopires/mytoolkit/internal/config"
	"github.com/aeciopires/mytoolkit/internal/httpapi"
	"github.com/aeciopires/mytoolkit/internal/logging"
	"github.com/aeciopires/mytoolkit/internal/version"
)

// toolHandlers accumulates REST handlers registered by each feature's init().
var toolHandlers = httpapi.ToolHandlers{}

func registerToolHandler(slug string, h http.HandlerFunc) {
	toolHandlers[slug] = h
}

var serveFlags config.Flags

var serveCmd = &cobra.Command{
	Use:   "serve",
	Short: "Start the MyToolkit web server (REST API + web UI)",
	RunE:  runServe,
}

func init() {
	rootCmd.AddCommand(serveCmd)
	for _, cmd := range []*cobra.Command{serveCmd, rootCmd} {
		cmd.Flags().StringVar(&serveFlags.Host, "host", config.DefaultHost, "interface to bind to")
		cmd.Flags().StringVar(&serveFlags.Port, "port", config.DefaultPort, "TCP port to listen on")
		cmd.Flags().StringVar(&serveFlags.LogLevel, "log-level", config.DefaultLogLevel, "log level: debug, info, warn, error")
	}
}

func runServe(cmd *cobra.Command, args []string) error {
	serveFlags.HostChanged = cmd.Flags().Changed("host")
	serveFlags.PortChanged = cmd.Flags().Changed("port")
	serveFlags.LogChanged = cmd.Flags().Changed("log-level")
	cfg := config.Load(serveFlags)

	logging.Setup(os.Stderr, cfg.LogLevel, "mytoolkit", version.Version)

	shutdownTracing, err := setupTracing(cmd.Context(), "mytoolkit")
	if err != nil {
		return logFatal(err, "tracing setup failed")
	}
	defer shutdownTracing()

	handler := app.New(toolHandlers)
	log.Info().Str("addr", cfg.Addr()).Msg("starting MyToolkit web server")
	if err := listenAndServe(cmd.Context(), cfg.Addr(), handler); err != nil {
		return logFatal(err, "web server failed")
	}
	return nil
}

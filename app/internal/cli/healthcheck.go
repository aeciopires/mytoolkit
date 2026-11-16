package cli

import (
	"cmp"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/config"
)

var healthcheckFlags struct {
	port string
	path string
	mcp  bool
}

// healthcheckCmd performs a local HTTP GET against /healthz (or --path) and
// exits non-zero on failure. It exists because the distroless production
// image has no shell/curl for a Docker HEALTHCHECK, so Docker Compose
// (unlike Kubernetes, which probes HTTP directly) needs a self-contained
// check. --mcp targets the streamable-HTTP MCP server's port instead.
var healthcheckCmd = &cobra.Command{
	Use:    "healthcheck",
	Short:  "Check the local server's /healthz endpoint (used by Docker HEALTHCHECK)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		port := healthcheckFlags.port
		if port == "" {
			if healthcheckFlags.mcp {
				port = cmp.Or(os.Getenv(config.EnvMCPPort), config.DefaultMCPPort)
			} else {
				port = cmp.Or(os.Getenv(config.EnvPort), config.DefaultPort)
			}
		}
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://127.0.0.1:" + port + healthcheckFlags.path)
		if err != nil {
			return err
		}
		defer func() { _ = resp.Body.Close() }()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("%s returned status %d", healthcheckFlags.path, resp.StatusCode)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
	healthcheckCmd.Flags().StringVar(&healthcheckFlags.port, "port", "", "port to check (default: $MYTOOLKIT_PORT or 8080; with --mcp, $MYTOOLKIT_MCP_PORT or 8081)")
	healthcheckCmd.Flags().StringVar(&healthcheckFlags.path, "path", "/healthz", "endpoint to check: /healthz (liveness) or /readyz (readiness)")
	healthcheckCmd.Flags().BoolVar(&healthcheckFlags.mcp, "mcp", false, "check the streamable-HTTP MCP server (mcp --transport http) instead of the web server")
}

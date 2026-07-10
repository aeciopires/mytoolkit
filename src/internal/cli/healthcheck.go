package cli

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/config"
)

// healthcheckCmd performs a local HTTP GET against /healthz and exits
// non-zero on failure. It exists because the distroless production image has
// no shell/curl for a Docker HEALTHCHECK, so Docker Compose (unlike
// Kubernetes, which probes HTTP directly) needs a self-contained check.
var healthcheckCmd = &cobra.Command{
	Use:    "healthcheck",
	Short:  "Check the local server's /healthz endpoint (used by Docker HEALTHCHECK)",
	Hidden: true,
	RunE: func(cmd *cobra.Command, args []string) error {
		port := os.Getenv(config.EnvPort)
		if port == "" {
			port = config.DefaultPort
		}
		client := http.Client{Timeout: 2 * time.Second}
		resp, err := client.Get("http://127.0.0.1:" + port + "/healthz")
		if err != nil {
			return err
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("healthz returned status %d", resp.StatusCode)
		}
		return nil
	},
}

func init() {
	rootCmd.AddCommand(healthcheckCmd)
}

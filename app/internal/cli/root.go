// Package cli implements the MyToolkit CLI: a root command that defaults to
// starting the web server, plus one subcommand per feature.
package cli

import (
	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/version"
)

var rootCmd = &cobra.Command{
	Use:     "mytoolkit",
	Short:   "MyToolkit — developer utilities as a web app or CLI",
	Version: version.Version,
	Long: "MyToolkit is a suite of utilities for developers, analysts, and IT professionals.\n" +
		"Run with no arguments to start the web server, or use a subcommand to run a single tool from the CLI.",
	RunE: func(cmd *cobra.Command, args []string) error {
		return runServe(cmd, args)
	},
}

func init() {
	rootCmd.Flags().BoolP("version", "v", false, "print the mytoolkit version and exit")
}

// Execute runs the root command.
func Execute() error {
	return rootCmd.Execute()
}

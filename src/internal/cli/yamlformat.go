package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/yamlformat"
)

func init() {
	rootCmd.AddCommand(newYAMLFormatCommand())
	registerToolHandler("yaml-format", handlers.Wrap("yaml-format", yamlformat.Format))
}

func newYAMLFormatCommand() *cobra.Command {
	var opts yamlformat.Options
	return newTextToolCommand("yaml-format", "Reformat a YAML document with consistent indentation", func(fs *pflag.FlagSet) {
		fs.IntVar(&opts.Indent, "indent", 2, "spaces per indent level")
	}, func(input []byte) (string, error) {
		return yamlformat.Format(input, opts)
	})
}

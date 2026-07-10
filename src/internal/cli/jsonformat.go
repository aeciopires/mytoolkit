package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsonformat"
)

func init() {
	rootCmd.AddCommand(newJSONFormatCommand())
	registerToolHandler("json-format", handlers.Wrap("json-format", jsonformat.Format))
}

func newJSONFormatCommand() *cobra.Command {
	var opts jsonformat.Options
	var minify bool
	return newTextToolCommand("json-format", "Pretty-print or minify a JSON document", func(fs *pflag.FlagSet) {
		fs.BoolVar(&minify, "minify", false, "minify instead of pretty-print")
		fs.IntVar(&opts.Indent, "indent", 2, "spaces per indent level (pretty mode only)")
	}, func(input []byte) (string, error) {
		if minify {
			opts.Mode = "minify"
		}
		return jsonformat.Format(input, opts)
	})
}

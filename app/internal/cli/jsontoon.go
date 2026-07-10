package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsontoon"
)

func init() {
	rootCmd.AddCommand(newJSONToonCommand())
	registerToolHandler("json-toon", handlers.Wrap("json-toon", jsontoon.Convert))
}

func newJSONToonCommand() *cobra.Command {
	var opts jsontoon.Options
	return newTextToolCommand("json-toon", "Convert JSON into TOON (Token-Oriented Object Notation) to reduce LLM token usage", func(fs *pflag.FlagSet) {
		fs.StringVar(&opts.Delimiter, "delimiter", "comma", "delimiter: comma, tab, pipe")
		fs.IntVar(&opts.IndentSize, "indent", 2, "spaces per indent level")
	}, func(input []byte) (string, error) {
		return jsontoon.Convert(input, opts)
	})
}

package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/yamltojson"
)

func init() {
	rootCmd.AddCommand(newYAMLToJSONCommand())
	registerToolHandler("yaml-to-json", handlers.Wrap("yaml-to-json", yamltojson.Convert))
}

func newYAMLToJSONCommand() *cobra.Command {
	var opts yamltojson.Options
	return newTextToolCommand("yaml-to-json", "Convert a YAML document to pretty-printed JSON", func(fs *pflag.FlagSet) {
		fs.IntVar(&opts.Indent, "indent", 2, "spaces per indent level")
	}, func(input []byte) (string, error) {
		return yamltojson.Convert(input, opts)
	})
}

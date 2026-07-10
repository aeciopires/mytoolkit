package cli

import (
	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsontoyaml"
)

func init() {
	rootCmd.AddCommand(newJSONToYAMLCommand())
	registerToolHandler("json-to-yaml", handlers.Wrap("json-to-yaml", jsontoyaml.Convert))
}

func newJSONToYAMLCommand() *cobra.Command {
	var opts jsontoyaml.Options
	return newTextToolCommand("json-to-yaml", "Convert a JSON document to YAML", nil, func(input []byte) (string, error) {
		return jsontoyaml.Convert(input, opts)
	})
}

package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/yamltojson"
)

func init() {
	rootCmd.AddCommand(newYAMLToJSONCommand())
	registerToolHandler("yaml-to-json", yamlToJSONHandler())
}

// yamlToJSONHandler godoc
// @Summary Convert YAML to JSON
// @Description Converts a YAML document to pretty-printed JSON using sigs.k8s.io/yaml.YAMLToJSONStrict — the same library kubectl/client-go/the API server use, and strict about rejecting duplicate mapping keys. Only the first document of a multi-document stream is converted.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{indent=int}} true "indent: spaces per level, default 2"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_YAML"
// @Router /api/v1/tools/yaml-to-json [post]
func yamlToJSONHandler() http.HandlerFunc {
	return handlers.Wrap("yaml-to-json", yamltojson.Convert)
}

func newYAMLToJSONCommand() *cobra.Command {
	var opts yamltojson.Options
	return newTextToolCommand("yaml-to-json", "Convert a YAML document to pretty-printed JSON", func(fs *pflag.FlagSet) {
		fs.IntVar(&opts.Indent, "indent", 2, "spaces per indent level")
	}, func(input []byte) (string, error) {
		return yamltojson.Convert(input, opts)
	})
}

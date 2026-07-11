package cli

import (
	"net/http"

	"github.com/spf13/cobra"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsontoyaml"
)

func init() {
	rootCmd.AddCommand(newJSONToYAMLCommand())
	registerToolHandler("json-to-yaml", jsonToYAMLHandler())
}

// jsonToYAMLHandler godoc
// @Summary Convert JSON to YAML
// @Description Converts a JSON document to YAML using sigs.k8s.io/yaml (the same library Kubernetes uses to render its typed API objects as YAML). Input is validated as strict JSON before conversion.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string} true "JSON document"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_JSON"
// @Router /api/v1/tools/json-to-yaml [post]
func jsonToYAMLHandler() http.HandlerFunc {
	return handlers.Wrap("json-to-yaml", jsontoyaml.Convert)
}

func newJSONToYAMLCommand() *cobra.Command {
	var opts jsontoyaml.Options
	return newTextToolCommand("json-to-yaml", "Convert a JSON document to YAML", nil, func(input []byte) (string, error) {
		return jsontoyaml.Convert(input, opts)
	})
}

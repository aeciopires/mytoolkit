package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/yamlformat"
)

func init() {
	rootCmd.AddCommand(newYAMLFormatCommand())
	registerToolHandler("yaml-format", yamlFormatHandler())
}

// yamlFormatHandler godoc
// @Summary Reformat YAML with consistent indentation
// @Description Reformats every document in a YAML stream ("---"-separated multi-document streams are fully supported). style=flow collapses collections to compact {}/[] notation (indent option is ignored); style=block (default) uses indented layout.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{indent=int,style=string}} true "indent: spaces per level, block style only, default 2; style: block (default) or flow"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse "e.g. INVALID_YAML, INVALID_OPTION"
// @Router /api/v1/tools/yaml-format [post]
func yamlFormatHandler() http.HandlerFunc {
	return handlers.Wrap("yaml-format", yamlformat.Format)
}

func newYAMLFormatCommand() *cobra.Command {
	var opts yamlformat.Options
	return newTextToolCommand("yaml-format", "Reformat a YAML document with consistent indentation", func(fs *pflag.FlagSet) {
		fs.IntVar(&opts.Indent, "indent", 2, "spaces per indent level (block style only)")
		fs.StringVar(&opts.Style, "style", "block", `collection style: "block" or "flow"`)
	}, func(input []byte) (string, error) {
		return yamlformat.Format(input, opts)
	})
}

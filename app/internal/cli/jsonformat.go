package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsonformat"
)

func init() {
	rootCmd.AddCommand(newJSONFormatCommand())
	registerToolHandler("json-format", jsonFormatHandler())
}

// jsonFormatHandler godoc
// @Summary Pretty-print or minify JSON
// @Description Formats a JSON document: pretty-print with a configurable indent, or minify to a single compact line. Note: the web page's Validate/Beautify/Minify buttons run entirely client-side via the browser's native JSON.parse()/JSON.stringify() and never call this endpoint — it exists for REST/CLI/scripted use.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{mode=string,indent=int}} true "mode: pretty (default) or minify; indent: spaces per level, pretty mode only, default 2"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/json-format [post]
func jsonFormatHandler() http.HandlerFunc {
	return handlers.Wrap("json-format", jsonformat.Format)
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

package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/jsontoon"
)

func init() {
	rootCmd.AddCommand(newJSONToonCommand())
	registerToolHandler("json-toon", jsonToonHandler())
}

// jsonToonHandler godoc
// @Summary Convert JSON to TOON
// @Description Converts JSON into TOON (Token-Oriented Object Notation) to reduce LLM token usage. This REST endpoint and the CLI are full Go implementations; the web page instead converts entirely client-side in JavaScript (no data sent to the server) — see docs/api/json-toon.md.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{delimiter=string,indent_size=int}} true "delimiter: comma (default), tab, pipe; indent_size: default 2"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/json-toon [post]
func jsonToonHandler() http.HandlerFunc {
	return handlers.Wrap("json-toon", jsontoon.Convert)
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

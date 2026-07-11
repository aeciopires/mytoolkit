package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/caseconvert"
)

func init() {
	rootCmd.AddCommand(newCaseConvertCommand())
	registerToolHandler("case-convert", caseConvertHandler())
}

// caseConvertHandler godoc
// @Summary Convert text case
// @Description Converts text between Sentence case, UPPER CASE, lower case, Title Case, MiXeD CaSe, and iNvErSe CaSe.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{mode=string}} true "mode: sentence, upper, lower, title, mixed, inverse"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/case-convert [post]
func caseConvertHandler() http.HandlerFunc {
	return handlers.Wrap("case-convert", caseconvert.Convert)
}

func newCaseConvertCommand() *cobra.Command {
	var opts caseconvert.Options
	return newTextToolCommand("case-convert", "Convert text case: sentence, upper, lower, title, mixed, inverse", func(fs *pflag.FlagSet) {
		fs.StringVar(&opts.Mode, "mode", "sentence", "sentence, upper, lower, title, mixed, inverse")
	}, func(input []byte) (string, error) {
		return caseconvert.Convert(input, opts)
	})
}

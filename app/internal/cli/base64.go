package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/base64enc"
)

func init() {
	rootCmd.AddCommand(newBase64Command())
	registerToolHandler("base64", base64Handler())
}

// base64Handler godoc
// @Summary Encode or decode Base64
// @Description Encodes or decodes data using Base64 (standard or URL-safe alphabet), with optional padding control.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{decode=bool,variant=string,padding=bool}} true "variant: standard, url"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/base64 [post]
func base64Handler() http.HandlerFunc {
	return handlers.Wrap("base64", base64enc.Process)
}

func newBase64Command() *cobra.Command {
	var opts base64enc.Options
	var noPadding bool
	return newTextToolCommand("base64", "Encode or decode data using Base64", func(fs *pflag.FlagSet) {
		fs.BoolVar(&opts.Decode, "decode", false, "decode instead of encode")
		fs.StringVar(&opts.Variant, "variant", "standard", "base64 variant: standard, url")
		fs.BoolVar(&noPadding, "no-padding", false, "omit padding characters")
	}, func(input []byte) (string, error) {
		if noPadding {
			p := false
			opts.Padding = &p
		}
		return base64enc.Process(input, opts)
	})
}

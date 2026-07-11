package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/urlencode"
)

func init() {
	rootCmd.AddCommand(newURLEncodeCommand())
	registerToolHandler("url-encode", urlEncodeHandler())
}

// urlEncodeHandler godoc
// @Summary Encode or decode a URL component
// @Description Encodes or decodes text using URL percent-encoding for the query, path, or a full-URL component.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{decode=bool,component=string}} true "component: query (default), path, full"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/url-encode [post]
func urlEncodeHandler() http.HandlerFunc {
	return handlers.Wrap("url-encode", urlencode.Process)
}

func newURLEncodeCommand() *cobra.Command {
	var opts urlencode.Options
	return newTextToolCommand("url-encode", "Encode or decode text using URL percent-encoding", func(fs *pflag.FlagSet) {
		fs.BoolVar(&opts.Decode, "decode", false, "decode instead of encode")
		fs.StringVar(&opts.Component, "component", "query", "component: query, path, full")
	}, func(input []byte) (string, error) {
		return urlencode.Process(input, opts)
	})
}

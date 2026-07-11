package cli

import (
	"net/http"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/hashgen"
)

func init() {
	rootCmd.AddCommand(newHashGenCommand())
	registerToolHandler("hash-gen", hashGenHandler())
}

// hashGenHandler godoc
// @Summary Generate a hash digest
// @Description Generates a hex-encoded hash digest of the input text using MD5, SHA-1, SHA-256, or SHA-512.
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{algorithm=string}} true "algorithm: md5, sha1, sha256 (default), sha512"
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/hash-gen [post]
func hashGenHandler() http.HandlerFunc {
	return handlers.Wrap("hash-gen", hashgen.Generate)
}

func newHashGenCommand() *cobra.Command {
	var opts hashgen.Options
	return newTextToolCommand("hash-gen", "Generate a hash digest of input text", func(fs *pflag.FlagSet) {
		fs.StringVar(&opts.Algorithm, "algo", "sha256", "hash algorithm: md5, sha1, sha256, sha512")
	}, func(input []byte) (string, error) {
		return hashgen.Generate(input, opts)
	})
}

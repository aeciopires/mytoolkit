package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/hashgen"
)

func init() {
	rootCmd.AddCommand(newHashGenCommand())
	registerToolHandler("hash-gen", handlers.Wrap("hash-gen", hashgen.Generate))
}

func newHashGenCommand() *cobra.Command {
	var opts hashgen.Options
	return newTextToolCommand("hash-gen", "Generate a hash digest of input text", func(fs *pflag.FlagSet) {
		fs.StringVar(&opts.Algorithm, "algo", "sha256", "hash algorithm: md5, sha1, sha256, sha512")
	}, func(input []byte) (string, error) {
		return hashgen.Generate(input, opts)
	})
}

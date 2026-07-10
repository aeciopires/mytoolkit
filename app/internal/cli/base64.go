package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/base64enc"
)

func init() {
	rootCmd.AddCommand(newBase64Command())
	registerToolHandler("base64", handlers.Wrap("base64", base64enc.Process))
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

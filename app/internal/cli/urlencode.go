package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/urlencode"
)

func init() {
	rootCmd.AddCommand(newURLEncodeCommand())
	registerToolHandler("url-encode", handlers.Wrap("url-encode", urlencode.Process))
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

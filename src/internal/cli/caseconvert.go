package cli

import (
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/httpapi/handlers"
	"github.com/aeciopires/mytoolkit/internal/tools/caseconvert"
)

func init() {
	rootCmd.AddCommand(newCaseConvertCommand())
	registerToolHandler("case-convert", handlers.Wrap("case-convert", caseconvert.Convert))
}

func newCaseConvertCommand() *cobra.Command {
	var opts caseconvert.Options
	return newTextToolCommand("case-convert", "Convert text case: sentence, upper, lower, title, mixed, inverse", func(fs *pflag.FlagSet) {
		fs.StringVar(&opts.Mode, "mode", "sentence", "sentence, upper, lower, title, mixed, inverse")
	}, func(input []byte) (string, error) {
		return caseconvert.Convert(input, opts)
	})
}

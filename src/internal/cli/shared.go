package cli

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/aeciopires/mytoolkit/internal/textio"
)

// newTextToolCommand builds a cobra.Command for the common "read --in,
// transform, write --out" shape shared by most tool subcommands.
func newTextToolCommand(use, short string, bindFlags func(*pflag.FlagSet), run func(input []byte) (string, error)) *cobra.Command {
	var inPath, outPath string

	cmd := &cobra.Command{
		Use:   use,
		Short: short,
		RunE: func(cmd *cobra.Command, args []string) error {
			input, err := textio.Read(inPath)
			if err != nil {
				return err
			}
			out, err := run(input)
			if err != nil {
				return err
			}
			if (outPath == "" || outPath == "-") && !strings.HasSuffix(out, "\n") {
				out += "\n"
			}
			return textio.Write(outPath, []byte(out))
		},
	}
	cmd.Flags().StringVar(&inPath, "in", "-", "input file, or - for stdin")
	cmd.Flags().StringVar(&outPath, "out", "-", "output file, or - for stdout")
	if bindFlags != nil {
		bindFlags(cmd.Flags())
	}
	return cmd
}

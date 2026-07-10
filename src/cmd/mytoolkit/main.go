// Command mytoolkit is the MyToolkit application entrypoint: it starts the
// web server by default, or runs a single tool as a CLI subcommand.
package main

import (
	"fmt"
	"os"

	"github.com/aeciopires/mytoolkit/internal/cli"
)

func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

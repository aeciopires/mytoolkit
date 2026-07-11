// Command mytoolkit is the MyToolkit application entrypoint: it starts the
// web server by default, or runs a single tool as a CLI subcommand.
package main

import (
	"fmt"
	"os"

	"github.com/aeciopires/mytoolkit/internal/cli"

	_ "github.com/aeciopires/mytoolkit/docs" // swag-generated spec, registered via its init()
)

// @title MyToolkit REST API
// @version 1.0
// @description Developer utilities (JSON/YAML formatting and conversion, JWT, hashing, encoding, password generation, QR codes, and more) exposed as a REST API. Every tool also has a CLI subcommand and a web UI backed by this same API — see the project README for the full picture.
// @contact.name Aecio Pires
// @contact.url https://github.com/aeciopires
// @license.name MIT
// @license.url https://github.com/aeciopires/mytoolkit/blob/main/LICENSE
// @BasePath /
// @schemes http https
func main() {
	if err := cli.Execute(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

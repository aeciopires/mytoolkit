package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jsonformat"
)

type jsonFormatIn struct {
	Input  string `json:"input" jsonschema:"JSON text to format"`
	Mode   string `json:"mode,omitempty" jsonschema:"pretty or minify (default: pretty)"`
	Indent int    `json:"indent,omitempty" jsonschema:"spaces per indent level, pretty mode only (default: 2)"`
}

func handleJSONFormat(_ context.Context, _ *sdkmcp.CallToolRequest, in jsonFormatIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := jsonformat.Format([]byte(in.Input), jsonformat.Options{
		Mode:   in.Mode,
		Indent: in.Indent,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "json-format",
			Description: "Format (pretty-print) or minify a JSON document.",
		}, handleJSONFormat)
	})
}

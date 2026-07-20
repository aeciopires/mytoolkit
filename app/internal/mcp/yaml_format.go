package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/yamlformat"
)

type yamlFormatIn struct {
	Input  string `json:"input" jsonschema:"YAML text to format"`
	Indent int    `json:"indent,omitempty" jsonschema:"spaces per indent level (default: 2)"`
	Style  string `json:"style,omitempty" jsonschema:"block or flow (default: block)"`
}

func handleYAMLFormat(_ context.Context, _ *sdkmcp.CallToolRequest, in yamlFormatIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := yamlformat.Format([]byte(in.Input), yamlformat.Options{
		Indent: in.Indent,
		Style:  in.Style,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "yaml-format",
			Description: "Format a YAML document (single or multi-document stream) with consistent indentation.",
		}, handleYAMLFormat)
	})
}

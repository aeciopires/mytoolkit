package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jsontoon"
)

type jsonToonIn struct {
	Input      string `json:"input" jsonschema:"JSON text to convert to TOON"`
	Delimiter  string `json:"delimiter,omitempty" jsonschema:"comma, tab, or pipe (default: comma)"`
	IndentSize int    `json:"indent_size,omitempty" jsonschema:"spaces per indent level (default: 2)"`
}

func handleJSONToon(_ context.Context, _ *sdkmcp.CallToolRequest, in jsonToonIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := jsontoon.Convert([]byte(in.Input), jsontoon.Options{
		Delimiter:  in.Delimiter,
		IndentSize: in.IndentSize,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "json-toon",
			Description: "Convert JSON into TOON (Token-Oriented Object Notation) to shrink LLM token usage.",
		}, handleJSONToon)
	})
}

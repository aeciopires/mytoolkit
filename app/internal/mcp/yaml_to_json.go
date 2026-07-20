package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/yamltojson"
)

type yamlToJSONIn struct {
	Input  string `json:"input" jsonschema:"YAML document to convert to JSON"`
	Indent int    `json:"indent,omitempty" jsonschema:"spaces per indent level (default: 2)"`
}

func handleYAMLToJSON(_ context.Context, _ *sdkmcp.CallToolRequest, in yamlToJSONIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := yamltojson.Convert([]byte(in.Input), yamltojson.Options{
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
			Name:        "yaml-to-json",
			Description: "Convert a YAML document to pretty-printed JSON.",
		}, handleYAMLToJSON)
	})
}

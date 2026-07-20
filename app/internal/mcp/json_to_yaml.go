package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jsontoyaml"
)

type jsonToYAMLIn struct {
	Input string `json:"input" jsonschema:"JSON document to convert to YAML"`
}

func handleJSONToYAML(_ context.Context, _ *sdkmcp.CallToolRequest, in jsonToYAMLIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := jsontoyaml.Convert([]byte(in.Input), jsontoyaml.Options{})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "json-to-yaml",
			Description: "Convert a JSON document to YAML.",
		}, handleJSONToYAML)
	})
}

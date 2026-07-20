package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jsontree"
)

type jsonTreeIn struct {
	Input string `json:"input" jsonschema:"JSON text to parse into a tree"`
}

// Out is 'any', not jsontree.Node: Node is a recursive type
// (Children []Node), and the SDK's schema inference panics on cyclic
// struct types when it tries to build an output JSON Schema. Using 'any'
// skips schema generation entirely (documented AddTool behavior) while
// still returning the node as structured JSON content.
func handleJSONTree(_ context.Context, _ *sdkmcp.CallToolRequest, in jsonTreeIn) (*sdkmcp.CallToolResult, any, error) {
	node, err := jsontree.Parse([]byte(in.Input), jsontree.Options{})
	if err != nil {
		return nil, nil, toolErr(err)
	}
	return nil, node, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "json-tree",
			Description: "Parse JSON into a tree structure (key, type, value, children) for structural analysis.",
		}, handleJSONTree)
	})
}

package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/urlencode"
)

type urlEncodeIn struct {
	Input     string `json:"input" jsonschema:"text to encode, or the URL-encoded text to decode"`
	Decode    bool   `json:"decode,omitempty" jsonschema:"decode instead of encode (default: false)"`
	Component string `json:"component,omitempty" jsonschema:"query, path, or full (default: query)"`
}

func handleURLEncode(_ context.Context, _ *sdkmcp.CallToolRequest, in urlEncodeIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := urlencode.Process([]byte(in.Input), urlencode.Options{
		Decode:    in.Decode,
		Component: in.Component,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "url-encode",
			Description: "Encode or decode text according to URL encoding rules (query, path segment, or full URL).",
		}, handleURLEncode)
	})
}

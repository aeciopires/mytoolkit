package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/hashgen"
)

type hashGenIn struct {
	Input     string `json:"input" jsonschema:"text to hash"`
	Algorithm string `json:"algorithm,omitempty" jsonschema:"md5, sha1, sha256, or sha512 (default: sha256)"`
}

func handleHashGen(_ context.Context, _ *sdkmcp.CallToolRequest, in hashGenIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := hashgen.Generate([]byte(in.Input), hashgen.Options{
		Algorithm: in.Algorithm,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "hash-gen",
			Description: "Generate a hex-encoded hash (MD5, SHA-1, SHA-256, or SHA-512) of text.",
		}, handleHashGen)
	})
}

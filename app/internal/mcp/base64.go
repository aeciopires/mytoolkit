package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/base64enc"
)

type base64In struct {
	Input   string `json:"input" jsonschema:"text to encode, or the Base64 text to decode"`
	Decode  bool   `json:"decode,omitempty" jsonschema:"decode instead of encode (default: false)"`
	Variant string `json:"variant,omitempty" jsonschema:"standard or url (default: standard)"`
	Padding *bool  `json:"padding,omitempty" jsonschema:"include '=' padding (default: true; set false to omit padding)"`
}

func handleBase64(_ context.Context, _ *sdkmcp.CallToolRequest, in base64In) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := base64enc.Process([]byte(in.Input), base64enc.Options{
		Decode:  in.Decode,
		Variant: in.Variant,
		Padding: in.Padding,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "base64",
			Description: "Encode or decode text using Base64 (standard or URL-safe variant, with optional padding).",
		}, handleBase64)
	})
}

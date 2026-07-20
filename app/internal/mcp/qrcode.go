package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/qrcode"
)

type qrcodeIn struct {
	Text string `json:"text" jsonschema:"text or URL to encode"`
	Size int    `json:"size,omitempty" jsonschema:"image size in pixels, up to 2048 (default: 256)"`
}

// handleQRCode is the one MCP tool with binary output: unlike every other
// tool it returns *sdkmcp.CallToolResult directly with an ImageContent
// block instead of relying on the SDK's automatic JSON-of-Out content
// (Out is 'any', so no output schema is generated, matching the REST
// endpoint's own image/png exception documented in PLAN_ARCHITECTURE.md).
func handleQRCode(_ context.Context, _ *sdkmcp.CallToolRequest, in qrcodeIn) (*sdkmcp.CallToolResult, any, error) {
	png, err := qrcode.Generate(in.Text, qrcode.Options{Size: in.Size})
	if err != nil {
		return nil, nil, toolErr(err)
	}
	return &sdkmcp.CallToolResult{
		Content: []sdkmcp.Content{
			&sdkmcp.ImageContent{Data: png, MIMEType: "image/png"},
		},
	}, nil, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "qrcode",
			Description: "Generate a QR code PNG image from text or a URL.",
		}, handleQRCode)
	})
}

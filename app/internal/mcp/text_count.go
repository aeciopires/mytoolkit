package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/textcount"
)

type textCountIn struct {
	Input string `json:"input" jsonschema:"text to count"`
}

func handleTextCount(_ context.Context, _ *sdkmcp.CallToolRequest, in textCountIn) (*sdkmcp.CallToolResult, textcount.Counts, error) {
	counts, err := textcount.Count([]byte(in.Input), textcount.Options{})
	if err != nil {
		return nil, textcount.Counts{}, toolErr(err)
	}
	return nil, counts, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "text-count",
			Description: "Count characters (with and without spaces), words, and lines in text.",
		}, handleTextCount)
	})
}

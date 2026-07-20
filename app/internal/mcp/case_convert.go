package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/caseconvert"
)

type caseConvertIn struct {
	Input string `json:"input" jsonschema:"text to convert"`
	Mode  string `json:"mode" jsonschema:"required: sentence, upper, lower, title, mixed, or inverse"`
}

func handleCaseConvert(_ context.Context, _ *sdkmcp.CallToolRequest, in caseConvertIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := caseconvert.Convert([]byte(in.Input), caseconvert.Options{
		Mode: in.Mode,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "case-convert",
			Description: "Convert text between Sentence case, UPPER CASE, lower case, Title Case, mIxEd cAsE, and InVeRsE cAsE.",
		}, handleCaseConvert)
	})
}

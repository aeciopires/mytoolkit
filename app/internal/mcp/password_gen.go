package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/password"
)

// passwordGenIn mirrors password.Options exactly, with no MCP-side
// defaulting — same behavior as the REST endpoint (only the CLI applies
// convenience flag defaults like --lowercase=true). A JSON boolean can't
// distinguish "omitted" from "explicitly false", so at least one of
// Lowercase/Uppercase/Numbers/Symbols must be set true or the call fails
// with NO_CHARSET_SELECTED, same as REST.
type passwordGenIn struct {
	Length           int  `json:"length" jsonschema:"password length, 1-512"`
	Lowercase        bool `json:"lowercase,omitempty" jsonschema:"include lowercase letters"`
	Uppercase        bool `json:"uppercase,omitempty" jsonschema:"include uppercase letters"`
	Numbers          bool `json:"numbers,omitempty" jsonschema:"include numbers"`
	Symbols          bool `json:"symbols,omitempty" jsonschema:"include symbols"`
	ExcludeConfusing bool `json:"exclude_confusing,omitempty" jsonschema:"exclude visually confusing characters (i l L 1 o 0 O)"`
	ExcludeAmbiguous bool `json:"exclude_ambiguous,omitempty" jsonschema:"exclude ambiguous punctuation characters"`
}

func handlePasswordGen(_ context.Context, _ *sdkmcp.CallToolRequest, in passwordGenIn) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := password.Generate(password.Options{
		Length:           in.Length,
		Lowercase:        in.Lowercase,
		Uppercase:        in.Uppercase,
		Numbers:          in.Numbers,
		Symbols:          in.Symbols,
		ExcludeConfusing: in.ExcludeConfusing,
		ExcludeAmbiguous: in.ExcludeAmbiguous,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "password-gen",
			Description: "Generate a cryptographically random password. At least one of lowercase/uppercase/numbers/symbols must be true.",
		}, handlePasswordGen)
	})
}

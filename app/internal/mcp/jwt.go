package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jwttool"
)

// jwt splits into two MCP tools (jwt_encode/jwt_decode) rather than one
// tool with an overloaded field, unlike the REST endpoint's single
// /api/v1/tools/jwt route with a "mode" option — two clean, independently
// documented input shapes are a better fit for how MCP clients pick a
// tool than replicating the REST mode discriminator.

type jwtDecodeIn struct {
	Token  string `json:"token" jsonschema:"the JWT to decode"`
	Secret string `json:"secret,omitempty" jsonschema:"shared secret to verify HMAC-family tokens (HS256/384/512); omit to decode without verification"`
	Key    string `json:"key,omitempty" jsonschema:"PEM-encoded public key to verify RSA/ECDSA/EdDSA tokens; omit to decode without verification"`
}

func handleJWTDecode(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtDecodeIn) (*sdkmcp.CallToolResult, jwttool.DecodeResult, error) {
	result, err := jwttool.Decode(in.Token, in.Secret, in.Key)
	if err != nil {
		return nil, jwttool.DecodeResult{}, toolErr(err)
	}
	return nil, result, nil
}

type jwtEncodeIn struct {
	Claims    string `json:"claims" jsonschema:"claims as JSON object text, e.g. {\"sub\":\"123\"}"`
	Secret    string `json:"secret,omitempty" jsonschema:"shared secret for HMAC-family algorithms (HS256/384/512)"`
	Key       string `json:"key,omitempty" jsonschema:"PEM-encoded private key for RSA/ECDSA/EdDSA algorithms"`
	Algorithm string `json:"algorithm,omitempty" jsonschema:"signing algorithm (default: HS256); see jwttool.SupportedAlgorithms"`
}

func handleJWTEncode(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtEncodeIn) (*sdkmcp.CallToolResult, textOut, error) {
	claims, err := jwttool.ParseClaimsJSON([]byte(in.Claims))
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	token, err := jwttool.Encode(claims, in.Secret, in.Key, in.Algorithm)
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: token}, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_decode",
			Description: "Decode a JWT's header and claims, optionally verifying its signature with a secret (HMAC) or public key (RSA/ECDSA/EdDSA).",
		}, handleJWTDecode)
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_encode",
			Description: "Sign a JSON claims object into a JWT using the given algorithm (default HS256).",
		}, handleJWTEncode)
	})
}

package mcp

import (
	"context"
	"time"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/jwttool"
)

// jwt splits into four MCP tools (jwt_decode/jwt_encode/jwt_keygen/
// jwt_jwks) rather than one tool with an overloaded field, unlike the REST
// endpoint's single /api/v1/tools/jwt route with a "mode" option — clean,
// independently documented input shapes are a better fit for how MCP
// clients pick a tool than replicating the REST mode discriminator.

type jwtDecodeIn struct {
	Token  string `json:"token" jsonschema:"the JWT to decode"`
	Secret string `json:"secret,omitempty" jsonschema:"shared secret to verify HMAC-family tokens (HS256/384/512); omit to decode without verification"`
	Key    string `json:"key,omitempty" jsonschema:"public key (PEM), JWK or JWKS to verify RSA/ECDSA/EdDSA tokens; a JWKS key is selected by the token's kid header; omit to decode without verification"`
}

func handleJWTDecode(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtDecodeIn) (*sdkmcp.CallToolResult, jwttool.DecodeResult, error) {
	result, err := jwttool.Decode(in.Token, in.Secret, in.Key)
	if err != nil {
		return nil, jwttool.DecodeResult{}, toolErr(err)
	}
	return nil, result, nil
}

type jwtEncodeIn struct {
	Claims       string   `json:"claims,omitempty" jsonschema:"base claims as JSON object text, e.g. {\"sub\":\"123\"}; may be omitted when the other claim fields are set"`
	CustomClaims []string `json:"custom_claims,omitempty" jsonschema:"custom claims as name=value pairs; a repeated name, or name[]=value, produces a list of strings"`
	Issuer       string   `json:"issuer,omitempty" jsonschema:"iss claim"`
	Subject      string   `json:"subject,omitempty" jsonschema:"sub claim"`
	Audience     []string `json:"audience,omitempty" jsonschema:"aud claim (one value is emitted as a string, several as a list)"`
	ExpiresIn    string   `json:"expires_in,omitempty" jsonschema:"sets exp relative to now, e.g. 15m, 1h, 7d"`
	NotBefore    string   `json:"not_before,omitempty" jsonschema:"sets nbf relative to now, e.g. 0s or -30s"`
	IssuedAt     bool     `json:"issued_at,omitempty" jsonschema:"set iat to now"`
	JWTID        bool     `json:"jwt_id,omitempty" jsonschema:"set jti to a random value"`
	Secret       string   `json:"secret,omitempty" jsonschema:"shared secret for HMAC-family algorithms (HS256/384/512)"`
	Key          string   `json:"key,omitempty" jsonschema:"private key (PEM or private JWK, e.g. from jwt_keygen) for RSA/ECDSA/EdDSA algorithms"`
	KeyID        string   `json:"kid,omitempty" jsonschema:"kid header; defaults to the private JWK's kid when key is a JWK"`
	Algorithm    string   `json:"algorithm,omitempty" jsonschema:"signing algorithm (default: HS256): HS256/384/512, RS256/384/512, PS256/384/512, ES256/384/512, EdDSA"`
}

func handleJWTEncode(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtEncodeIn) (*sdkmcp.CallToolResult, textOut, error) {
	base, err := jwttool.ParseClaimsJSON([]byte(in.Claims))
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	exp, err := jwttool.ParseDuration("expires_in", in.ExpiresIn)
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	nbf, err := jwttool.ParseDuration("not_before", in.NotBefore)
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	claims, err := jwttool.BuildClaims(base, jwttool.ClaimsSpec{
		Issuer:     in.Issuer,
		Subject:    in.Subject,
		Audience:   in.Audience,
		ExpiresIn:  exp,
		NotBefore:  nbf,
		IssuedAt:   in.IssuedAt,
		GenerateID: in.JWTID,
		Custom:     in.CustomClaims,
	}, time.Now())
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	token, err := jwttool.Encode(claims, jwttool.EncodeOptions{
		Algorithm: in.Algorithm,
		Secret:    in.Secret,
		Key:       in.Key,
		KeyID:     in.KeyID,
	})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: token}, nil
}

type jwtKeygenIn struct {
	Algorithm string `json:"algorithm,omitempty" jsonschema:"algorithm the key is for (default: HS256): HS* yields a random secret, RS*/PS* an RSA key, ES* an EC key on the matching curve, EdDSA an Ed25519 key"`
	Bits      int    `json:"bits,omitempty" jsonschema:"RSA modulus size for RS*/PS*: 2048 (default), 3072 or 4096"`
	KeyID     string `json:"kid,omitempty" jsonschema:"kid to assign; defaults to the RFC 7638 JWK thumbprint"`
}

func handleJWTKeygen(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtKeygenIn) (*sdkmcp.CallToolResult, jwttool.KeyPair, error) {
	kp, err := jwttool.GenerateKey(jwttool.KeyOptions{Algorithm: in.Algorithm, Bits: in.Bits, KeyID: in.KeyID})
	if err != nil {
		return nil, jwttool.KeyPair{}, toolErr(err)
	}
	return nil, kp, nil
}

type jwtJWKSIn struct {
	Key       string `json:"key" jsonschema:"one or more PEM keys/certificates (public or private), a JWK, or a JWKS; private material is reduced to its public half"`
	Algorithm string `json:"algorithm,omitempty" jsonschema:"alg to stamp on every key; inferred from the key type when omitted"`
	KeyID     string `json:"kid,omitempty" jsonschema:"kid to assign (single-key input only); defaults to the RFC 7638 JWK thumbprint"`
}

func handleJWTJWKS(_ context.Context, _ *sdkmcp.CallToolRequest, in jwtJWKSIn) (*sdkmcp.CallToolResult, jwttool.JWKSet, error) {
	set, err := jwttool.BuildJWKS(in.Key, jwttool.JWKSOptions{Algorithm: in.Algorithm, KeyID: in.KeyID})
	if err != nil {
		return nil, jwttool.JWKSet{}, toolErr(err)
	}
	return nil, set, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_decode",
			Description: "Decode a JWT's header and claims, optionally verifying its signature and exp/nbf/iat with a secret (HMAC) or a public key, JWK or JWKS (RSA/ECDSA/EdDSA).",
		}, handleJWTDecode)
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_encode",
			Description: "Sign a JWT offline from a JSON claims object plus optional registered claims (iss/sub/aud/exp/nbf/iat/jti) and custom string or string-list claims, using the given algorithm (default HS256).",
		}, handleJWTEncode)
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_keygen",
			Description: "Generate JWT signing key material offline: an HMAC secret, or an RSA/EC/Ed25519 key pair as PEM plus private/public JWK and a ready-to-publish JWKS.",
		}, handleJWTKeygen)
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "jwt_jwks",
			Description: "Convert PEM keys/certificates or a JWK into a JWKS containing only public keys, ready to publish for token verification.",
		}, handleJWTJWKS)
	})
}

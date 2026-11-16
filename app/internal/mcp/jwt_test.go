package mcp

import (
	"encoding/json"
	"testing"
)

func TestHandleJWTEncodeDecodeRoundTrip(t *testing.T) {
	_, encodeOut, err := handleJWTEncode(t.Context(), nil, jwtEncodeIn{
		Claims: `{"sub":"123"}`,
		Secret: "s3cr3t",
	})
	if err != nil {
		t.Fatalf("encode: unexpected error: %v", err)
	}
	if encodeOut.Output == "" {
		t.Fatal("expected a non-empty token")
	}

	_, decodeOut, err := handleJWTDecode(t.Context(), nil, jwtDecodeIn{
		Token:  encodeOut.Output,
		Secret: "s3cr3t",
	})
	if err != nil {
		t.Fatalf("decode: unexpected error: %v", err)
	}
	if decodeOut.Claims["sub"] != "123" {
		t.Errorf("got claims %+v", decodeOut.Claims)
	}
	if decodeOut.Valid == nil || !*decodeOut.Valid {
		t.Errorf("expected Valid=true, got %+v", decodeOut.Valid)
	}
}

func TestHandleJWTEncodeEmptyClaims(t *testing.T) {
	_, _, err := handleJWTEncode(t.Context(), nil, jwtEncodeIn{})
	if err == nil {
		t.Fatal("expected an error for empty claims")
	}
}

func TestHandleJWTKeygenEncodeVerifyWithJWKS(t *testing.T) {
	_, kp, err := handleJWTKeygen(t.Context(), nil, jwtKeygenIn{Algorithm: "ES256", KeyID: "k1"})
	if err != nil {
		t.Fatalf("keygen: %v", err)
	}
	_, set, err := handleJWTJWKS(t.Context(), nil, jwtJWKSIn{Key: kp.PublicKey, KeyID: "k1"})
	if err != nil {
		t.Fatalf("jwks: %v", err)
	}
	jwks, err := json.Marshal(set)
	if err != nil {
		t.Fatal(err)
	}

	_, enc, err := handleJWTEncode(t.Context(), nil, jwtEncodeIn{
		Algorithm:    "ES256",
		Key:          kp.PrivateKey,
		KeyID:        "k1",
		Subject:      "user-1",
		Audience:     []string{"api"},
		ExpiresIn:    "1h",
		CustomClaims: []string{"role=admin", "groups=dev", "groups=ops"},
	})
	if err != nil {
		t.Fatalf("encode: %v", err)
	}

	_, dec, err := handleJWTDecode(t.Context(), nil, jwtDecodeIn{Token: enc.Output, Key: string(jwks)})
	if err != nil {
		t.Fatalf("decode: %v", err)
	}
	if dec.Valid == nil || !*dec.Valid {
		t.Fatalf("expected Valid=true, got %v (%s)", dec.Valid, dec.VerifyError)
	}
	if dec.Header["kid"] != "k1" || dec.Claims["role"] != "admin" || dec.Claims["sub"] != "user-1" {
		t.Errorf("unexpected header/claims: %+v %+v", dec.Header, dec.Claims)
	}
	if groups, ok := dec.Claims["groups"].([]any); !ok || len(groups) != 2 {
		t.Errorf("groups: got %#v, want a 2-item list", dec.Claims["groups"])
	}
}

func TestHandleJWTEncodeInvalidDuration(t *testing.T) {
	_, _, err := handleJWTEncode(t.Context(), nil, jwtEncodeIn{Subject: "x", Secret: "s", ExpiresIn: "soon"})
	if err == nil {
		t.Fatal("expected an error for an invalid expires_in")
	}
}

func TestHandleJWTDecodeMalformedToken(t *testing.T) {
	_, _, err := handleJWTDecode(t.Context(), nil, jwtDecodeIn{Token: "not-a-jwt"})
	if err == nil {
		t.Fatal("expected an error for a malformed token")
	}
}

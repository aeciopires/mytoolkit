package mcp

import (
	"context"
	"testing"
)

func TestHandleJWTEncodeDecodeRoundTrip(t *testing.T) {
	_, encodeOut, err := handleJWTEncode(context.TODO(), nil, jwtEncodeIn{
		Claims: `{"sub":"123"}`,
		Secret: "s3cr3t",
	})
	if err != nil {
		t.Fatalf("encode: unexpected error: %v", err)
	}
	if encodeOut.Output == "" {
		t.Fatal("expected a non-empty token")
	}

	_, decodeOut, err := handleJWTDecode(context.TODO(), nil, jwtDecodeIn{
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
	_, _, err := handleJWTEncode(context.TODO(), nil, jwtEncodeIn{})
	if err == nil {
		t.Fatal("expected an error for empty claims")
	}
}

func TestHandleJWTDecodeMalformedToken(t *testing.T) {
	_, _, err := handleJWTDecode(context.TODO(), nil, jwtDecodeIn{Token: "not-a-jwt"})
	if err == nil {
		t.Fatal("expected an error for a malformed token")
	}
}

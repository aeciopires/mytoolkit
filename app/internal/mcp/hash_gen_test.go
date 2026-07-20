package mcp

import (
	"context"
	"testing"
)

func TestHandleHashGenDefaultAlgorithm(t *testing.T) {
	_, out, err := handleHashGen(context.TODO(), nil, hashGenIn{Input: "hello"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	const wantSHA256 = "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824"
	if out.Output != wantSHA256 {
		t.Errorf("got %q, want %q", out.Output, wantSHA256)
	}
}

func TestHandleHashGenUnsupportedAlgorithm(t *testing.T) {
	_, _, err := handleHashGen(context.TODO(), nil, hashGenIn{Input: "hello", Algorithm: "bogus"})
	if err == nil {
		t.Fatal("expected an error for an unsupported algorithm")
	}
}

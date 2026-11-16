package mcp

import (
	"testing"
)

func TestHandleURLEncode(t *testing.T) {
	_, out, err := handleURLEncode(t.Context(), nil, urlEncodeIn{Input: "a b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "a+b" {
		t.Errorf("got %q, want %q", out.Output, "a+b")
	}
}

func TestHandleURLEncodeDecode(t *testing.T) {
	_, out, err := handleURLEncode(t.Context(), nil, urlEncodeIn{Input: "a+b", Decode: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "a b" {
		t.Errorf("got %q, want %q", out.Output, "a b")
	}
}

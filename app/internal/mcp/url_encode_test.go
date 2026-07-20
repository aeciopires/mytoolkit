package mcp

import (
	"context"
	"testing"
)

func TestHandleURLEncode(t *testing.T) {
	_, out, err := handleURLEncode(context.TODO(), nil, urlEncodeIn{Input: "a b"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "a+b" {
		t.Errorf("got %q, want %q", out.Output, "a+b")
	}
}

func TestHandleURLEncodeDecode(t *testing.T) {
	_, out, err := handleURLEncode(context.TODO(), nil, urlEncodeIn{Input: "a+b", Decode: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "a b" {
		t.Errorf("got %q, want %q", out.Output, "a b")
	}
}

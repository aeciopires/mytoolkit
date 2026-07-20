package mcp

import (
	"context"
	"testing"
)

func TestHandlePasswordGen(t *testing.T) {
	_, out, err := handlePasswordGen(context.TODO(), nil, passwordGenIn{Length: 20, Lowercase: true, Numbers: true})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(out.Output) != 20 {
		t.Errorf("got password of length %d, want 20", len(out.Output))
	}
}

func TestHandlePasswordGenNoCharsetSelected(t *testing.T) {
	_, _, err := handlePasswordGen(context.TODO(), nil, passwordGenIn{Length: 10})
	if err == nil {
		t.Fatal("expected an error when no character class is enabled")
	}
}

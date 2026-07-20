package mcp

import (
	"context"
	"testing"
)

func TestHandleTextCount(t *testing.T) {
	_, counts, err := handleTextCount(context.TODO(), nil, textCountIn{Input: "hello world\nsecond line"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if counts.Words != 4 {
		t.Errorf("got %d words, want 4", counts.Words)
	}
	if counts.Lines != 2 {
		t.Errorf("got %d lines, want 2", counts.Lines)
	}
}

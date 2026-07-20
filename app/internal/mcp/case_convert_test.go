package mcp

import (
	"context"
	"testing"
)

func TestHandleCaseConvert(t *testing.T) {
	_, out, err := handleCaseConvert(context.TODO(), nil, caseConvertIn{Input: "hello world", Mode: "title"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "Hello World" {
		t.Errorf("got %q, want %q", out.Output, "Hello World")
	}
}

func TestHandleCaseConvertMissingMode(t *testing.T) {
	_, _, err := handleCaseConvert(context.TODO(), nil, caseConvertIn{Input: "hello"})
	if err == nil {
		t.Fatal("expected an error when mode is missing")
	}
}

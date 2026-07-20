package mcp

import (
	"context"
	"testing"
)

func TestHandleYAMLFormat(t *testing.T) {
	_, out, err := handleYAMLFormat(context.TODO(), nil, yamlFormatIn{Input: "a: 1\nb: 2\n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestHandleYAMLFormatInvalidStyle(t *testing.T) {
	_, _, err := handleYAMLFormat(context.TODO(), nil, yamlFormatIn{Input: "a: 1\n", Style: "bogus"})
	if err == nil {
		t.Fatal("expected an error for an invalid style")
	}
}

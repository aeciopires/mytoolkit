package mcp

import (
	"context"
	"testing"
)

func TestHandleJSONFormat(t *testing.T) {
	_, out, err := handleJSONFormat(context.TODO(), nil, jsonFormatIn{Input: `{"a":1}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "{\n  \"a\": 1\n}" {
		t.Errorf("got %q", out.Output)
	}
}

func TestHandleJSONFormatEmptyInput(t *testing.T) {
	_, _, err := handleJSONFormat(context.TODO(), nil, jsonFormatIn{})
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}

package mcp

import (
	"testing"
)

func TestHandleJSONFormat(t *testing.T) {
	_, out, err := handleJSONFormat(t.Context(), nil, jsonFormatIn{Input: `{"a":1}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "{\n  \"a\": 1\n}" {
		t.Errorf("got %q", out.Output)
	}
}

func TestHandleJSONFormatEmptyInput(t *testing.T) {
	_, _, err := handleJSONFormat(t.Context(), nil, jsonFormatIn{})
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}

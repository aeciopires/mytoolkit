package mcp

import (
	"testing"
)

func TestHandleJSONToYAML(t *testing.T) {
	_, out, err := handleJSONToYAML(t.Context(), nil, jsonToYAMLIn{Input: `{"a":1}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "a: 1\n" {
		t.Errorf("got %q", out.Output)
	}
}

func TestHandleJSONToYAMLInvalidJSON(t *testing.T) {
	_, _, err := handleJSONToYAML(t.Context(), nil, jsonToYAMLIn{Input: "{not json"})
	if err == nil {
		t.Fatal("expected an error for invalid JSON")
	}
}

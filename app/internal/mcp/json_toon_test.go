package mcp

import (
	"testing"
)

func TestHandleJSONToon(t *testing.T) {
	_, out, err := handleJSONToon(t.Context(), nil, jsonToonIn{Input: `{"a":1}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output == "" {
		t.Error("expected non-empty output")
	}
}

func TestHandleJSONToonInvalidDelimiter(t *testing.T) {
	_, _, err := handleJSONToon(t.Context(), nil, jsonToonIn{Input: `{"a":1}`, Delimiter: "bogus"})
	if err == nil {
		t.Fatal("expected an error for an invalid delimiter")
	}
}

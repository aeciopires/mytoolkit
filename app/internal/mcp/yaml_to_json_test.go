package mcp

import (
	"context"
	"testing"
)

func TestHandleYAMLToJSON(t *testing.T) {
	_, out, err := handleYAMLToJSON(context.TODO(), nil, yamlToJSONIn{Input: "a: 1\n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.Output != "{\n  \"a\": 1\n}" {
		t.Errorf("got %q", out.Output)
	}
}

func TestHandleYAMLToJSONEmptyInput(t *testing.T) {
	_, _, err := handleYAMLToJSON(context.TODO(), nil, yamlToJSONIn{})
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}

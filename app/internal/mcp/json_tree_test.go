package mcp

import (
	"testing"

	"github.com/aeciopires/mytoolkit/internal/tools/jsontree"
)

func TestHandleJSONTree(t *testing.T) {
	_, out, err := handleJSONTree(t.Context(), nil, jsonTreeIn{Input: `{"a":1}`})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	node, ok := out.(jsontree.Node)
	if !ok {
		t.Fatalf("got %T, want jsontree.Node", out)
	}
	if node.Type != "object" || len(node.Children) != 1 {
		t.Errorf("got %+v", node)
	}
}

func TestHandleJSONTreeEmptyInput(t *testing.T) {
	_, _, err := handleJSONTree(t.Context(), nil, jsonTreeIn{})
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}

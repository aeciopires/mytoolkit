package jsontree

import (
	"strings"
	"testing"
)

func TestParseFlatObject(t *testing.T) {
	node, err := Parse([]byte(`{"a":1,"b":"x"}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if node.Type != "object" || len(node.Children) != 2 {
		t.Fatalf("unexpected node: %+v", node)
	}
	if node.Children[0].Key != "a" || node.Children[1].Key != "b" {
		t.Errorf("key order not preserved: %+v", node.Children)
	}
}

func TestParseNested(t *testing.T) {
	node, err := Parse([]byte(`{"a":1,"b":[true,null]}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	b := node.Children[1]
	if b.Type != "array" || len(b.Children) != 2 {
		t.Fatalf("unexpected array node: %+v", b)
	}
	if b.Children[0].Type != "bool" || b.Children[1].Type != "null" {
		t.Errorf("unexpected children types: %+v", b.Children)
	}
}

func TestParseEmptyInput(t *testing.T) {
	if _, err := Parse([]byte(""), Options{}); err == nil {
		t.Error("expected error for empty input")
	}
}

func TestParseMalformed(t *testing.T) {
	if _, err := Parse([]byte(`{"a":}`), Options{}); err == nil {
		t.Error("expected error for malformed JSON")
	}
}

func TestParseLargeNumberPreserved(t *testing.T) {
	node, err := Parse([]byte(`{"n":123456789012345678}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if node.Children[0].Value != "123456789012345678" {
		t.Errorf("large number not preserved exactly: %v", node.Children[0].Value)
	}
}

func TestParseUnicodeString(t *testing.T) {
	node, err := Parse([]byte(`{"s":"café 🦊"}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if node.Children[0].Value != "café 🦊" {
		t.Errorf("unicode not preserved: %v", node.Children[0].Value)
	}
}

func TestParseErrorIncludesPosition(t *testing.T) {
	_, err := Parse([]byte(`{"a":}`), Options{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "line 1") || !strings.Contains(err.Error(), "column") {
		t.Errorf("error message missing position info: %q", err.Error())
	}
}

func TestParseErrorPositionOnLaterLine(t *testing.T) {
	_, err := Parse([]byte("{\n  \"a\": 1,\n  \"b\":\n}"), Options{})
	if err == nil {
		t.Fatal("expected error")
	}
	if !strings.Contains(err.Error(), "line 4") {
		t.Errorf("expected error to point at line 4, got: %q", err.Error())
	}
}

func TestParseRejectsTrailingData(t *testing.T) {
	_, err := Parse([]byte(`{"a":1}{"b":2}`), Options{})
	if err == nil {
		t.Fatal("expected error for trailing data after a complete JSON value")
	}
	if !strings.Contains(err.Error(), "extra data") {
		t.Errorf("expected an 'extra data' error, got: %q", err.Error())
	}
}

func TestParseRejectsTrailingGarbage(t *testing.T) {
	_, err := Parse([]byte(`{"a":1}garbage`), Options{})
	if err == nil {
		t.Fatal("expected error for trailing garbage after a complete JSON value")
	}
}

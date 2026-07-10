package jsontree

import "testing"

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

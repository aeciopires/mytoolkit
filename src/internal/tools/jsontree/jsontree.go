// Package jsontree implements the JSON Tree Viewer tool's pure logic.
package jsontree

import (
	"bytes"
	"encoding/json"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Node struct {
	Key      string `json:"key,omitempty"`
	Type     string `json:"type"`
	Value    any    `json:"value,omitempty"`
	Children []Node `json:"children,omitempty"`
}

type Options struct{}

func Parse(input []byte, _ Options) (Node, error) {
	if len(bytes.TrimSpace(input)) == 0 {
		return Node{}, apperr.ErrEmptyInput
	}

	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber()

	node, err := parseValue(dec, "")
	if err != nil {
		return Node{}, apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
	}
	return node, nil
}

// parseValue reads one JSON value (of any kind) from dec via streaming
// tokens, preserving object key order (which map[string]any would lose).
func parseValue(dec *json.Decoder, key string) (Node, error) {
	tok, err := dec.Token()
	if err != nil {
		return Node{}, err
	}

	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			children := []Node{}
			for dec.More() {
				keyTok, err := dec.Token()
				if err != nil {
					return Node{}, err
				}
				childKey, _ := keyTok.(string)
				child, err := parseValue(dec, childKey)
				if err != nil {
					return Node{}, err
				}
				children = append(children, child)
			}
			if _, err := dec.Token(); err != nil { // consume '}'
				return Node{}, err
			}
			return Node{Key: key, Type: "object", Children: children}, nil
		case '[':
			children := []Node{}
			for dec.More() {
				child, err := parseValue(dec, "")
				if err != nil {
					return Node{}, err
				}
				children = append(children, child)
			}
			if _, err := dec.Token(); err != nil { // consume ']'
				return Node{}, err
			}
			return Node{Key: key, Type: "array", Children: children}, nil
		}
	case string:
		return Node{Key: key, Type: "string", Value: t}, nil
	case json.Number:
		return Node{Key: key, Type: "number", Value: t.String()}, nil
	case bool:
		return Node{Key: key, Type: "bool", Value: t}, nil
	case nil:
		return Node{Key: key, Type: "null"}, nil
	}
	return Node{}, nil
}

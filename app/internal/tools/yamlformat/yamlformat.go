// Package yamlformat implements the YAML Formatter tool's pure logic.
package yamlformat

import (
	"bytes"
	"io"

	"gopkg.in/yaml.v3"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Indent int    `json:"indent"`
	Style  string `json:"style"` // "block" (default) | "flow"
}

// Format reformats every document in a YAML stream with consistent
// indentation. A stream may contain multiple "---"-separated documents
// (YAML spec §9.1.1-9.1.2); each is decoded and re-encoded independently,
// and the encoder re-inserts "---" between them automatically. Comments,
// anchors/aliases, and explicit tags are preserved because decoding targets
// a yaml.Node tree (not a plain Go value), not because the spec requires
// it — the spec explicitly leaves comment placement as a presentation
// detail with no formal attachment semantics, so comment position may
// still shift in edge cases (e.g. a comment on its own line between two
// mapping keys can attach to either).
func Format(input []byte, opts Options) (string, error) {
	if len(input) == 0 {
		return "", apperr.ErrEmptyInput
	}
	indent := opts.Indent
	if indent <= 0 {
		indent = 2
	}
	if err := apperr.OneOf("style", styleOrDefault(opts.Style), "block", "flow"); err != nil {
		return "", err
	}
	style := blockOrFlowStyle(opts.Style)

	dec := yaml.NewDecoder(bytes.NewReader(input))
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)

	docs := 0
	for {
		var node yaml.Node
		err := dec.Decode(&node)
		if err == io.EOF {
			break
		}
		if err != nil {
			return "", apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
		}
		normalizeStyle(&node, style)
		if err := enc.Encode(&node); err != nil {
			return "", apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
		}
		docs++
	}
	_ = enc.Close()

	// A stream of only comments/whitespace decodes to zero documents
	// without error; treat it the same as empty input rather than
	// returning a silently empty result.
	if docs == 0 {
		return "", apperr.ErrEmptyInput
	}

	return buf.String(), nil
}

func styleOrDefault(style string) string {
	if style == "" {
		return "block"
	}
	return style
}

// normalizeStyle forces every collection (mapping/sequence) node in the
// tree to the requested presentation style. Per the YAML spec, "the node
// style is a presentation detail and is not reflected in the
// serialization tree or representation graph" — so this is always a safe,
// lossless transformation, not a reinterpretation of the data. Scalar
// nodes are left untouched: their original style (plain/quoted/block) can
// carry meaning (e.g. a quoted "yes" is a string, an unquoted yes may be
// resolved as a value by the reader) and must not be altered by a
// formatter.
func normalizeStyle(n *yaml.Node, style yaml.Style) {
	if n == nil {
		return
	}
	if n.Kind == yaml.SequenceNode || n.Kind == yaml.MappingNode {
		n.Style = style
	}
	for _, c := range n.Content {
		normalizeStyle(c, style)
	}
}

func blockOrFlowStyle(style string) yaml.Style {
	if style == "flow" {
		return yaml.FlowStyle
	}
	return 0
}

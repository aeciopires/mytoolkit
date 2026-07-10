// Package yamltojson implements the YAML to JSON Converter tool's pure logic.
package yamltojson

import (
	"bytes"
	"encoding/json"

	"sigs.k8s.io/yaml"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Indent int `json:"indent"`
}

// Convert parses a single YAML document and re-emits it as pretty-printed
// JSON, using sigs.k8s.io/yaml (the same library Kubernetes uses to accept
// YAML manifests as JSON internally).
//
// It uses YAMLToJSONStrict rather than YAMLToJSON: the YAML spec forbids
// duplicate mapping keys, but the plain (non-strict) converter silently
// keeps one of them in an undefined order. Strict mode surfaces that as an
// error instead of silently dropping data.
func Convert(input []byte, opts Options) (string, error) {
	if len(input) == 0 {
		return "", apperr.ErrEmptyInput
	}
	indent := opts.Indent
	if indent <= 0 {
		indent = 2
	}

	j, err := yaml.YAMLToJSONStrict(input)
	if err != nil {
		return "", apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
	}

	var buf bytes.Buffer
	// j is always well-formed JSON here (YAMLToJSONStrict only returns nil
	// error alongside valid output), so json.Indent cannot fail on it.
	_ = json.Indent(&buf, j, "", spaces(indent))
	return buf.String(), nil
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

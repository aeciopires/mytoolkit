// Package yamlformat implements the YAML Formatter tool's pure logic.
package yamlformat

import (
	"bytes"

	"gopkg.in/yaml.v3"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Indent int `json:"indent"`
}

func Format(input []byte, opts Options) (string, error) {
	if len(input) == 0 {
		return "", apperr.ErrEmptyInput
	}
	indent := opts.Indent
	if indent <= 0 {
		indent = 2
	}

	var node yaml.Node
	if err := yaml.Unmarshal(input, &node); err != nil {
		return "", apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
	}

	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(indent)
	if err := enc.Encode(&node); err != nil {
		return "", apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
	}
	_ = enc.Close()

	return buf.String(), nil
}

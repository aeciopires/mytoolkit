// Package jsonformat implements the JSON Formatter tool's pure logic.
package jsonformat

import (
	"bytes"
	"encoding/json"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Mode   string `json:"mode"` // "pretty" | "minify", default "pretty"
	Indent int    `json:"indent"`
}

func Format(input []byte, opts Options) (string, error) {
	if len(input) == 0 {
		return "", apperr.ErrEmptyInput
	}
	mode := opts.Mode
	if mode == "" {
		mode = "pretty"
	}
	if err := apperr.OneOf("mode", mode, "pretty", "minify"); err != nil {
		return "", err
	}

	if mode == "minify" {
		var buf bytes.Buffer
		if err := json.Compact(&buf, input); err != nil {
			return "", apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
		}
		return buf.String(), nil
	}

	indent := opts.Indent
	if indent <= 0 {
		indent = 2
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, input, "", spaces(indent)); err != nil {
		return "", apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
	}
	return buf.String(), nil
}

func spaces(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = ' '
	}
	return string(b)
}

// Package jsontoyaml implements the JSON to YAML Converter tool's pure logic.
package jsontoyaml

import (
	"encoding/json"

	"sigs.k8s.io/yaml"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct{}

// Convert parses a JSON document and re-emits it as YAML, using
// sigs.k8s.io/yaml (the same library Kubernetes uses to render its typed
// API objects as YAML).
//
// sigs.k8s.io/yaml.JSONToYAML parses its input with a YAML decoder (YAML
// is a superset of JSON), which means it happily accepts things that
// aren't valid JSON at all — trailing commas, unquoted keys, single-quoted
// strings, even a line with just a comment — rather than reporting an
// error. Since a "JSON to YAML converter" should reject non-JSON input
// instead of silently reinterpreting it as YAML, input is validated with
// encoding/json first (which does enforce the JSON grammar), and only
// converted via the library afterward.
func Convert(input []byte, opts Options) (string, error) {
	if len(input) == 0 {
		return "", apperr.ErrEmptyInput
	}

	var v any
	if err := json.Unmarshal(input, &v); err != nil {
		return "", apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
	}

	y, err := yaml.JSONToYAML(input)
	if err != nil {
		return "", apperr.Newf(400, "INVALID_JSON", "%s", err.Error())
	}
	return string(y), nil
}

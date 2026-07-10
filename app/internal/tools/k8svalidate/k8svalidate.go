// Package k8svalidate implements the Kubernetes YAML Validator tool's pure
// logic: checking that a YAML document (or multi-document stream) is both
// well-formed YAML and shaped the way the Kubernetes API server requires —
// an object with apiVersion, kind, and (if present) an object-shaped
// metadata field.
package k8svalidate

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	yamlv3 "gopkg.in/yaml.v3"
	k8syaml "sigs.k8s.io/yaml"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct{}

type DocumentResult struct {
	Index      int    `json:"index"`
	APIVersion string `json:"api_version,omitempty"`
	Kind       string `json:"kind,omitempty"`
	Name       string `json:"name,omitempty"`
	Valid      bool   `json:"valid"`
	Error      string `json:"error,omitempty"`
}

type Result struct {
	Valid     bool             `json:"valid"`
	Documents []DocumentResult `json:"documents"`
}

// Validate parses a YAML stream (one or more "---"-separated documents) and
// checks each document against the two requirements every Kubernetes API
// object must satisfy: a non-empty string apiVersion and a non-empty
// string kind (k8s.io/apimachinery's TypeMeta), plus an object-shaped
// metadata field if one is present. It does not validate against any
// specific resource's schema (Deployment, Service, a CRD, ...) — that
// requires the full Kubernetes OpenAPI schema database, which this
// lightweight tool does not embed; see the "Known limitations" note in
// PLANS/PLAN_K8S_YAML_VALIDATOR.md.
//
// Each document is converted to JSON via sigs.k8s.io/yaml.YAMLToJSONStrict
// — the same library used by kubectl/client-go/the API server itself to
// accept YAML manifests as JSON — so a document that fails here would also
// fail against a real cluster for the same reason.
//
// A document-stream-level YAML syntax error (the input can't be parsed as
// YAML at all) aborts the whole call with an error, since there's no safe
// way to resync to the next document boundary. Once a document is
// successfully isolated, a per-document problem (duplicate keys, missing
// apiVersion/kind, wrong field types) is reported in that document's
// DocumentResult instead of aborting the batch, so a stream of 10
// documents with one mistake still reports on the other 9.
func Validate(input []byte, opts Options) (Result, error) {
	if len(input) == 0 {
		return Result{}, apperr.ErrEmptyInput
	}

	dec := yamlv3.NewDecoder(bytes.NewReader(input))
	var docs []DocumentResult
	for {
		var node yamlv3.Node
		err := dec.Decode(&node)
		if err == io.EOF {
			break
		}
		if err != nil {
			return Result{}, apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
		}

		raw, err := yamlv3.Marshal(&node)
		if err != nil {
			return Result{}, apperr.Newf(400, "INVALID_YAML", "%s", err.Error())
		}

		j, err := k8syaml.YAMLToJSONStrict(raw)
		if err != nil {
			docs = append(docs, DocumentResult{Index: len(docs) + 1, Valid: false, Error: err.Error()})
			continue
		}
		if strings.TrimSpace(string(j)) == "null" {
			// A blank document (e.g. a stray leading/trailing "---") isn't
			// a Kubernetes object and isn't an error either — kubectl
			// itself silently skips these.
			continue
		}
		docs = append(docs, validateDocument(len(docs)+1, j))
	}

	if len(docs) == 0 {
		return Result{}, apperr.New(400, "NO_DOCUMENTS", "no YAML documents found in input")
	}

	valid := true
	for _, d := range docs {
		if !d.Valid {
			valid = false
			break
		}
	}
	return Result{Valid: valid, Documents: docs}, nil
}

func validateDocument(index int, docJSON []byte) DocumentResult {
	var obj map[string]any
	if err := json.Unmarshal(docJSON, &obj); err != nil {
		return DocumentResult{
			Index: index,
			Valid: false,
			Error: "document must be a YAML mapping (object) at the root — apiVersion, kind, metadata, spec — not a list or a bare scalar",
		}
	}

	apiVersionVal, present := obj["apiVersion"]
	if !present {
		return DocumentResult{Index: index, Valid: false, Error: `missing required field "apiVersion"`}
	}
	apiVersion, ok := apiVersionVal.(string)
	if !ok {
		return DocumentResult{Index: index, Valid: false, Error: fmt.Sprintf(`field "apiVersion" must be a string, got %s`, jsonTypeName(apiVersionVal))}
	}
	if apiVersion == "" {
		return DocumentResult{Index: index, Valid: false, Error: `field "apiVersion" must not be empty`}
	}

	kindVal, present := obj["kind"]
	if !present {
		return DocumentResult{Index: index, APIVersion: apiVersion, Valid: false, Error: `missing required field "kind"`}
	}
	kind, ok := kindVal.(string)
	if !ok {
		return DocumentResult{Index: index, APIVersion: apiVersion, Valid: false, Error: fmt.Sprintf(`field "kind" must be a string, got %s`, jsonTypeName(kindVal))}
	}
	if kind == "" {
		return DocumentResult{Index: index, APIVersion: apiVersion, Valid: false, Error: `field "kind" must not be empty`}
	}

	name := ""
	if metaVal, present := obj["metadata"]; present {
		metaObj, ok := metaVal.(map[string]any)
		if !ok {
			return DocumentResult{
				Index: index, APIVersion: apiVersion, Kind: kind, Valid: false,
				Error: fmt.Sprintf(`field "metadata" must be a mapping (object), got %s`, jsonTypeName(metaVal)),
			}
		}
		if n, ok := metaObj["name"].(string); ok {
			name = n
		}
	}

	return DocumentResult{Index: index, APIVersion: apiVersion, Kind: kind, Name: name, Valid: true}
}

func jsonTypeName(v any) string {
	switch v.(type) {
	case nil:
		return "null"
	case bool:
		return "boolean"
	case float64:
		return "number"
	case string:
		return "string"
	case []any:
		return "array"
	case map[string]any:
		return "object"
	default:
		return "unknown"
	}
}

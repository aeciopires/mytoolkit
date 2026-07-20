package mcp

import (
	"context"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/aeciopires/mytoolkit/internal/tools/k8svalidate"
)

type k8sValidateIn struct {
	Input string `json:"input" jsonschema:"YAML document or multi-document stream (---separated) to validate"`
}

// handleK8sValidate mirrors the REST handler's behavior: a document that
// is well-formed YAML but semantically invalid for Kubernetes (missing
// apiVersion/kind, ...) is NOT a tool error — it's a successful call whose
// structured result reports Valid: false for that document. Only a hard
// parse failure (bad YAML syntax, no documents at all) becomes a Go error.
func handleK8sValidate(_ context.Context, _ *sdkmcp.CallToolRequest, in k8sValidateIn) (*sdkmcp.CallToolResult, k8svalidate.Result, error) {
	result, err := k8svalidate.Validate([]byte(in.Input), k8svalidate.Options{})
	if err != nil {
		return nil, k8svalidate.Result{}, toolErr(err)
	}
	return nil, result, nil
}

func init() {
	register(func(s *sdkmcp.Server) {
		sdkmcp.AddTool(s, &sdkmcp.Tool{
			Name:        "k8s-validate",
			Description: "Validate that a YAML document (single or multi-document) has the fields the Kubernetes API requires: apiVersion, kind, and a well-formed metadata block.",
		}, handleK8sValidate)
	})
}

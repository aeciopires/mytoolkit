package mcp

import (
	"context"
	"testing"
)

func TestHandleK8sValidateValidDocument(t *testing.T) {
	_, result, err := handleK8sValidate(context.TODO(), nil,k8sValidateIn{Input: "apiVersion: v1\nkind: Pod\nmetadata:\n  name: test\n"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if !result.Valid {
		t.Errorf("expected valid=true, got %+v", result)
	}
}

func TestHandleK8sValidateBadYAML(t *testing.T) {
	_, _, err := handleK8sValidate(context.TODO(), nil,k8sValidateIn{Input: "a: [unterminated"})
	if err == nil {
		t.Fatal("expected an error for malformed YAML")
	}
}

func TestHandleK8sValidateEmptyInput(t *testing.T) {
	_, _, err := handleK8sValidate(context.TODO(), nil,k8sValidateIn{})
	if err == nil {
		t.Fatal("expected an error for empty input")
	}
}

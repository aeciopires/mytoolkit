package k8svalidate

import "testing"

func TestValidateSingleDocument(t *testing.T) {
	tests := []struct {
		name       string
		input      string
		wantValid  bool
		wantErr    bool
		wantDocErr string
	}{
		{
			name:      "valid minimal object",
			input:     "apiVersion: v1\nkind: Namespace\nmetadata:\n  name: my-ns\n",
			wantValid: true,
		},
		{
			name:      "valid without metadata",
			input:     "apiVersion: v1\nkind: Namespace\n",
			wantValid: true,
		},
		{
			name:       "missing apiVersion",
			input:      "kind: Pod\nmetadata:\n  name: p\n",
			wantValid:  false,
			wantDocErr: `missing required field "apiVersion"`,
		},
		{
			name:       "missing kind",
			input:      "apiVersion: v1\nmetadata:\n  name: p\n",
			wantValid:  false,
			wantDocErr: `missing required field "kind"`,
		},
		{
			name:       "empty apiVersion",
			input:      "apiVersion: \"\"\nkind: Pod\n",
			wantValid:  false,
			wantDocErr: `field "apiVersion" must not be empty`,
		},
		{
			name:       "kind wrong type",
			input:      "apiVersion: v1\nkind: 5\n",
			wantValid:  false,
			wantDocErr: `field "kind" must be a string, got number`,
		},
		{
			name:       "metadata wrong type",
			input:      "apiVersion: v1\nkind: Pod\nmetadata: oops\n",
			wantValid:  false,
			wantDocErr: `field "metadata" must be a mapping (object), got string`,
		},
		{
			name:       "root is an array",
			input:      "- 1\n- 2\n",
			wantValid:  false,
			wantDocErr: "document must be a YAML mapping (object) at the root — apiVersion, kind, metadata, spec — not a list or a bare scalar",
		},
		{name: "empty input", input: "", wantErr: true},
		{name: "only separators", input: "---\n---\n", wantErr: true},
		{name: "malformed yaml", input: "apiVersion: [1, 2\n", wantErr: true},
		{name: "tab indentation", input: "apiVersion:\n\tv1\n", wantErr: true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := Validate([]byte(tt.input), Options{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Validate() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr {
				return
			}
			if len(result.Documents) != 1 {
				t.Fatalf("expected exactly 1 document, got %d", len(result.Documents))
			}
			doc := result.Documents[0]
			if doc.Valid != tt.wantValid {
				t.Errorf("Documents[0].Valid = %v, want %v (error=%q)", doc.Valid, tt.wantValid, doc.Error)
			}
			if result.Valid != tt.wantValid {
				t.Errorf("Result.Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if tt.wantDocErr != "" && doc.Error != tt.wantDocErr {
				t.Errorf("Documents[0].Error = %q, want %q", doc.Error, tt.wantDocErr)
			}
		})
	}
}

func TestValidateMultiDocument(t *testing.T) {
	input := `apiVersion: v1
kind: ConfigMap
metadata:
  name: cm1
data:
  a: b
---
apiVersion: apps/v1
metadata:
  name: dep1
---
apiVersion: v1
kind: Secret
metadata:
  name: sec1
`
	result, err := Validate([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Documents) != 3 {
		t.Fatalf("expected 3 documents, got %d", len(result.Documents))
	}
	if result.Valid {
		t.Error("expected overall Result.Valid = false (document 2 is missing kind)")
	}
	if !result.Documents[0].Valid || result.Documents[0].Kind != "ConfigMap" || result.Documents[0].Name != "cm1" {
		t.Errorf("document 1 = %+v, want valid ConfigMap/cm1", result.Documents[0])
	}
	if result.Documents[1].Valid {
		t.Errorf("document 2 should be invalid (missing kind), got %+v", result.Documents[1])
	}
	if !result.Documents[2].Valid || result.Documents[2].Kind != "Secret" {
		t.Errorf("document 3 = %+v, want valid Secret", result.Documents[2])
	}
}

func TestValidateSkipsBlankDocuments(t *testing.T) {
	input := "---\napiVersion: v1\nkind: Namespace\nmetadata:\n  name: n\n---\n---\n"
	result, err := Validate([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Documents) != 1 {
		t.Fatalf("expected blank documents to be skipped, got %d documents: %+v", len(result.Documents), result.Documents)
	}
	if !result.Valid {
		t.Errorf("expected valid result, got %+v", result)
	}
}

func TestValidateDuplicateKeyRejectedPerDocument(t *testing.T) {
	input := "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: ok\n---\na: 1\na: 2\n"
	result, err := Validate([]byte(input), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if len(result.Documents) != 2 {
		t.Fatalf("expected 2 documents, got %d", len(result.Documents))
	}
	if !result.Documents[0].Valid {
		t.Errorf("document 1 should still be valid despite document 2's problem, got %+v", result.Documents[0])
	}
	if result.Documents[1].Valid {
		t.Errorf("document 2 (duplicate key) should be invalid, got %+v", result.Documents[1])
	}
}

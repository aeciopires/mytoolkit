package yamltojson

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{
			"flat mapping",
			"a: 1\nb: hello\n",
			Options{},
			"{\n  \"a\": 1,\n  \"b\": \"hello\"\n}",
			false,
		},
		{
			"nested mapping and sequence",
			"a:\n  b: 1\n  c:\n    - x\n    - z\n",
			Options{},
			"{\n  \"a\": {\n    \"b\": 1,\n    \"c\": [\n      \"x\",\n      \"z\"\n    ]\n  }\n}",
			false,
		},
		{
			"custom indent",
			"a: 1\n",
			Options{Indent: 4},
			"{\n    \"a\": 1\n}",
			false,
		},
		{"empty input", "", Options{}, "", true},
		{"malformed yaml", "a: [1, 2\n", Options{}, "", true},
		{"tab indentation rejected", "a:\n\tb: 1\n", Options{}, "", true},
		{"duplicate keys rejected", "a: 1\na: 2\n", Options{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Convert() error = %v, wantErr %v, got=%q", err, tt.wantErr, got)
			}
			if err == nil && got != tt.want {
				t.Errorf("Convert() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestConvertYAML11BooleanResolution locks in a real, spec-documented
// gotcha: sigs.k8s.io/yaml resolves plain (unquoted) y/n/yes/no/on/off per
// the YAML 1.1 core schema, not YAML 1.2 — so an unquoted "y" or "no" in
// the source becomes a JSON boolean, not a JSON string. This is documented
// library behavior (see yaml.go's doc comment on YAMLToJSON), not a bug in
// this package; the test exists so a future dependency bump that changes
// this behavior is caught, not missed.
func TestConvertYAML11BooleanResolution(t *testing.T) {
	got, err := Convert([]byte("country: NO\nreply: y\nversion: \"1.2\"\n"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `"country": false`) {
		t.Errorf("expected unquoted NO to resolve to JSON false, got %q", got)
	}
	if !strings.Contains(got, `"reply": true`) {
		t.Errorf("expected unquoted y to resolve to JSON true, got %q", got)
	}
	if !strings.Contains(got, `"version": "1.2"`) {
		t.Errorf("expected quoted \"1.2\" to remain a JSON string, got %q", got)
	}
}

func TestConvertLargeIntegerPreserved(t *testing.T) {
	got, err := Convert([]byte("n: 123456789012345678\n"), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "123456789012345678") {
		t.Errorf("expected large integer to be preserved exactly, got %q", got)
	}
}

package jsontoyaml

import (
	"strings"
	"testing"
)

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"flat object", `{"a":1,"b":"hello"}`, "a: 1\nb: hello\n", false},
		{"nested object and array", `{"a":{"b":1,"c":["x","z"]}}`, "a:\n  b: 1\n  c:\n  - x\n  - z\n", false},
		{"empty input", "", "", true},
		{"malformed json: trailing comma", `{"a":1,}`, "", true},
		{"malformed json: unquoted key", `{a:1}`, "", true},
		{"malformed json: single-quoted string", `{"a": 'hello'}`, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input), Options{})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Convert() error = %v, wantErr %v, got=%q", err, tt.wantErr, got)
			}
			if err == nil && got != tt.want {
				t.Errorf("Convert() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestConvertQuotesStringsThatWouldResolveAsOtherTypes(t *testing.T) {
	// "no" and "y" are YAML 1.1 boolean aliases; the JSON string "no" must
	// round-trip as a YAML string, not an unquoted (and therefore
	// boolean-resolving) scalar.
	got, err := Convert([]byte(`{"country":"NO","reply":"y"}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `country: "NO"`) {
		t.Errorf(`expected "NO" to be quoted in YAML output, got %q`, got)
	}
	if !strings.Contains(got, `reply: "y"`) {
		t.Errorf(`expected "y" to be quoted in YAML output, got %q`, got)
	}
}

func TestConvertLargeIntegerPreserved(t *testing.T) {
	got, err := Convert([]byte(`{"n":123456789012345678}`), Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, "123456789012345678") {
		t.Errorf("expected large integer to be preserved exactly, got %q", got)
	}
}

package yamlformat

import (
	"strings"
	"testing"
)

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{"reindent list", "a: 1\nb:\n    - x\n    - y\n", Options{Indent: 2}, "a: 1\nb:\n  - x\n  - y\n", false},
		{"empty input", "", Options{}, "", true},
		{"malformed yaml", "a: [1, 2\n", Options{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format([]byte(tt.input), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Format() error = %v, wantErr %v, got=%q", err, tt.wantErr, got)
			}
			if err == nil && got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestFormatIdempotent(t *testing.T) {
	input := "a: 1\nb:\n  - x\n  - y\n"
	once, err := Format([]byte(input), Options{Indent: 2})
	if err != nil {
		t.Fatal(err)
	}
	twice, err := Format([]byte(once), Options{Indent: 2})
	if err != nil {
		t.Fatal(err)
	}
	if strings.TrimSpace(once) != strings.TrimSpace(twice) {
		t.Errorf("formatting is not idempotent:\nonce=%q\ntwice=%q", once, twice)
	}
}

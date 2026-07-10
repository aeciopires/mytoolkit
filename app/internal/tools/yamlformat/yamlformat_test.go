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
		{"tab indentation rejected", "a:\n\tb: 1\n", Options{}, "", true},
		{"invalid style option", "a: 1\n", Options{Style: "compact"}, "", true},
		{"whitespace-only stream", "\n\n", Options{}, "", true},
		{
			"multi-document stream reformatted",
			"a: 1\n---\nb:\n    - x\n    - y\n---\nc: 3\n",
			Options{Indent: 2},
			"a: 1\n---\nb:\n  - x\n  - y\n---\nc: 3\n",
			false,
		},
		{
			"flow style forces compact single-line collections",
			"a:\n  b: 1\n  c:\n    - 1\n    - 2\n",
			Options{Style: "flow"},
			"{a: {b: 1, c: [1, 2]}}\n",
			false,
		},
		{
			"block style normalizes mixed flow input",
			"a: {b: 1, c: [1, 2]}\n",
			Options{Style: "block", Indent: 2},
			"a:\n  b: 1\n  c:\n    - 1\n    - 2\n",
			false,
		},
		{
			"comments are preserved",
			"# head comment\na: 1 # line comment\nb: 2\n",
			Options{Indent: 2},
			"# head comment\na: 1 # line comment\nb: 2\n",
			false,
		},
		{
			"anchors and aliases are preserved",
			"a: &x\n  foo: 1\nb: *x\n",
			Options{Indent: 2},
			"a: &x\n  foo: 1\nb: *x\n",
			false,
		},
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

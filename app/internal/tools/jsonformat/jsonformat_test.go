package jsonformat

import "testing"

func TestFormat(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{"pretty default indent", `{"a":1,"b":2}`, Options{}, "{\n  \"a\": 1,\n  \"b\": 2\n}", false},
		{"pretty custom indent", `{"a":1}`, Options{Indent: 4}, "{\n    \"a\": 1\n}", false},
		{"minify", "{\n  \"a\": 1,\n  \"b\": 2\n}", Options{Mode: "minify"}, `{"a":1,"b":2}`, false},
		{"minify idempotent", `{"a":1}`, Options{Mode: "minify"}, `{"a":1}`, false},
		{"empty input", "", Options{}, "", true},
		{"malformed json", `{"a":}`, Options{}, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Format([]byte(tt.input), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Format() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("Format() = %q, want %q", got, tt.want)
			}
		})
	}
}

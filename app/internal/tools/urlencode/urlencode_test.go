package urlencode

import "testing"

func TestProcess(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{"encode query spaces", "hello world & friends", Options{}, "hello+world+%26+friends", false},
		{"decode query spaces", "hello+world+%26+friends", Options{Decode: true}, "hello world & friends", false},
		{"encode path spaces", "hello world", Options{Component: "path"}, "hello%20world", false},
		{"decode path spaces", "hello%20world", Options{Decode: true, Component: "path"}, "hello world", false},
		{"empty input", "", Options{}, "", false},
		{"invalid component", "x", Options{Component: "bogus"}, "", true},
		{"invalid encoding on decode", "%ZZ", Options{Decode: true}, "", true},
		{"unicode round trip encode", "café", Options{}, "caf%C3%A9", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Process([]byte(tt.input), tt.opts)
			if (err != nil) != tt.wantErr {
				t.Fatalf("Process() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("Process() = %q, want %q", got, tt.want)
			}
		})
	}
}

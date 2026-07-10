package base64enc

import "testing"

func boolPtr(b bool) *bool { return &b }

func TestProcess(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		opts    Options
		want    string
		wantErr bool
	}{
		{"encode standard", "hello world", Options{}, "aGVsbG8gd29ybGQ=", false},
		{"decode standard", "aGVsbG8gd29ybGQ=", Options{Decode: true}, "hello world", false},
		{"encode no padding", "hello world", Options{Padding: boolPtr(false)}, "aGVsbG8gd29ybGQ", false},
		{"decode no padding", "aGVsbG8gd29ybGQ", Options{Decode: true, Padding: boolPtr(false)}, "hello world", false},
		{"encode url variant", "a~b", Options{Variant: "url"}, "YX5i", false},
		{"empty input encode", "", Options{}, "", false},
		{"empty input decode", "", Options{Decode: true}, "", false},
		{"invalid base64", "not base64!!", Options{Decode: true}, "", true},
		{"invalid variant", "hello", Options{Variant: "bogus"}, "", true},
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

func TestRoundTrip(t *testing.T) {
	original := "The quick brown fox jumps over the lazy dog! 🦊"
	encoded, err := Process([]byte(original), Options{})
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := Process([]byte(encoded), Options{Decode: true})
	if err != nil {
		t.Fatal(err)
	}
	if decoded != original {
		t.Errorf("round trip = %q, want %q", decoded, original)
	}
}

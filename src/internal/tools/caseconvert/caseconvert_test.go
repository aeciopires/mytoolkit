package caseconvert

import "testing"

func TestConvert(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		mode    string
		want    string
		wantErr bool
	}{
		{"sentence basic", "hello world. this IS a test!", "sentence", "Hello world. This is a test!", false},
		{"sentence consecutive terminators", "Really?! Yes.", "sentence", "Really?! Yes.", false},
		{"upper", "Hello World", "upper", "HELLO WORLD", false},
		{"lower", "Hello World", "lower", "hello world", false},
		{"title", "hello WORLD example", "title", "Hello World Example", false},
		{"title collapses whitespace", "hello   world", "title", "Hello World", false},
		{"mixed pattern", "mixed case", "mixed", "MiXeD CaSe", false},
		{"inverse basic", "Hello World", "inverse", "hELLO wORLD", false},
		{"inverse round trip", "MiXeD CaSe", "inverse", "mIxEd cAsE", false},
		{"empty input", "", "upper", "", false},
		{"unsupported mode", "x", "bogus", "", true},
		{"unicode upper", "café", "upper", "CAFÉ", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Convert([]byte(tt.input), Options{Mode: tt.mode})
			if (err != nil) != tt.wantErr {
				t.Fatalf("Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err == nil && got != tt.want {
				t.Errorf("Convert() = %q, want %q", got, tt.want)
			}
		})
	}
}

func TestInverseIsSelfInverting(t *testing.T) {
	original := "Hello World! 123 café"
	once, err := Convert([]byte(original), Options{Mode: "inverse"})
	if err != nil {
		t.Fatal(err)
	}
	twice, err := Convert([]byte(once), Options{Mode: "inverse"})
	if err != nil {
		t.Fatal(err)
	}
	if twice != original {
		t.Errorf("double inverse = %q, want %q", twice, original)
	}
}

package textcount

import "testing"

func TestCount(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  Counts
	}{
		{"simple sentence", "Hello world", Counts{Characters: 11, CharactersNoSpaces: 10, Words: 2, Lines: 1}},
		{"two lines with trailing newline", "Hello world\nSecond line\n", Counts{Characters: 24, CharactersNoSpaces: 20, Words: 4, Lines: 2}},
		{"two lines no trailing newline", "Hello world\nSecond line", Counts{Characters: 23, CharactersNoSpaces: 20, Words: 4, Lines: 2}},
		{"empty input", "", Counts{Characters: 0, CharactersNoSpaces: 0, Words: 0, Lines: 0}},
		{"whitespace only", "   \n  ", Counts{Characters: 6, CharactersNoSpaces: 0, Words: 0, Lines: 2}},
		{"crlf normalized", "a\r\nb\r\n", Counts{Characters: 4, CharactersNoSpaces: 2, Words: 2, Lines: 2}},
		{"unicode characters", "café 🦊", Counts{Characters: 6, CharactersNoSpaces: 5, Words: 2, Lines: 1}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Count([]byte(tt.input), Options{})
			if err != nil {
				t.Fatalf("Count() error = %v", err)
			}
			if got != tt.want {
				t.Errorf("Count() = %+v, want %+v", got, tt.want)
			}
		})
	}
}

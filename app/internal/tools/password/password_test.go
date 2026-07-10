package password

import "testing"

func TestGenerateLength(t *testing.T) {
	for _, length := range []int{1, 8, 16, 64, 128} {
		got, err := Generate(Options{Length: length, Lowercase: true, Uppercase: true, Numbers: true})
		if err != nil {
			t.Fatalf("Generate(length=%d) error = %v", length, err)
		}
		if len(got) != length {
			t.Errorf("Generate(length=%d) len = %d", length, len(got))
		}
	}
}

func TestGenerateOnlyEnabledClasses(t *testing.T) {
	got, err := Generate(Options{Length: 200, Lowercase: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if !containsRune(Lowercase, r) {
			t.Fatalf("unexpected character %q in lowercase-only output", r)
		}
	}
}

func TestGenerateNoCharsetSelected(t *testing.T) {
	if _, err := Generate(Options{Length: 10}); err == nil {
		t.Error("expected error when no charset is enabled")
	}
}

func TestGenerateInvalidLength(t *testing.T) {
	if _, err := Generate(Options{Length: 0, Lowercase: true}); err == nil {
		t.Error("expected error for length 0")
	}
	if _, err := Generate(Options{Length: -1, Lowercase: true}); err == nil {
		t.Error("expected error for negative length")
	}
	if _, err := Generate(Options{Length: 1000, Lowercase: true}); err == nil {
		t.Error("expected error for length over 512")
	}
}

func TestGenerateExcludeConfusing(t *testing.T) {
	got, err := Generate(Options{Length: 300, Lowercase: true, Uppercase: true, Numbers: true, ExcludeConfusing: true})
	if err != nil {
		t.Fatal(err)
	}
	for _, r := range got {
		if containsRune(ConfusingChars, r) {
			t.Fatalf("confusing character %q found in output", r)
		}
	}
}

func TestGenerateExcludeAmbiguous(t *testing.T) {
	got, err := Generate(Options{Length: 300, Symbols: true, ExcludeAmbiguous: true})
	if err != nil {
		t.Fatal(err)
	}
	safeSymbols := "!#$%&*+-=?@^_"
	for _, r := range got {
		if !containsRune(safeSymbols, r) {
			t.Fatalf("unexpected symbol %q after excluding ambiguous characters", r)
		}
	}
}

func TestGenerateProducesDifferentOutputs(t *testing.T) {
	opts := Options{Length: 20, Lowercase: true, Uppercase: true, Numbers: true, Symbols: true}
	a, err := Generate(opts)
	if err != nil {
		t.Fatal(err)
	}
	different := false
	for i := 0; i < 10; i++ {
		b, err := Generate(opts)
		if err != nil {
			t.Fatal(err)
		}
		if a != b {
			different = true
			break
		}
	}
	if !different {
		t.Error("expected at least one different output across repeated generations")
	}
}

func containsRune(s string, r rune) bool {
	for _, c := range s {
		if c == r {
			return true
		}
	}
	return false
}

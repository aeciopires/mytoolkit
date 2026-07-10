package qrcode

import (
	"bytes"
	"strings"
	"testing"
)

var pngMagic = []byte{0x89, 'P', 'N', 'G'}

func TestGenerateValidText(t *testing.T) {
	png, err := Generate("https://example.com", Options{Size: 256})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(png, pngMagic) {
		t.Error("output does not start with PNG magic header")
	}
}

func TestGenerateUnicodeText(t *testing.T) {
	png, err := Generate("héllo wörld 🦊", Options{})
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.HasPrefix(png, pngMagic) {
		t.Error("output does not start with PNG magic header")
	}
}

func TestGenerateEmptyText(t *testing.T) {
	if _, err := Generate("", Options{}); err == nil {
		t.Error("expected error for empty text")
	}
}

func TestGenerateTooLarge(t *testing.T) {
	big := strings.Repeat("a", maxTextBytes+1)
	if _, err := Generate(big, Options{}); err == nil {
		t.Error("expected error for text exceeding capacity")
	}
}

func TestGenerateDifferentSizes(t *testing.T) {
	small, err := Generate("hello", Options{Size: 64})
	if err != nil {
		t.Fatal(err)
	}
	large, err := Generate("hello", Options{Size: 512})
	if err != nil {
		t.Fatal(err)
	}
	if len(small) == 0 || len(large) == 0 {
		t.Fatal("expected non-empty output")
	}
	if len(small) == len(large) {
		t.Error("expected different byte lengths for different sizes")
	}
}

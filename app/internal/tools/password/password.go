// Package password implements the Password Generator tool's pure logic.
package password

import (
	"crypto/rand"
	"math/big"
	"strings"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

const (
	Lowercase      = "abcdefghijklmnopqrstuvwxyz"
	Uppercase      = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
	Numbers        = "0123456789"
	Symbols        = "!#$%&*+-=?@^_{}[]()/'\"`~,;:.<>\\"
	ConfusingChars = "ilL1o0O"
	AmbiguousChars = "{}[]()/\\'\"`~,;:.<>"
)

type Options struct {
	Length           int  `json:"length"`
	Lowercase        bool `json:"lowercase"`
	Uppercase        bool `json:"uppercase"`
	Numbers          bool `json:"numbers"`
	Symbols          bool `json:"symbols"`
	ExcludeConfusing bool `json:"exclude_confusing"`
	ExcludeAmbiguous bool `json:"exclude_ambiguous"`
}

func Generate(opts Options) (string, error) {
	if opts.Length < 1 {
		return "", apperr.New(400, "INVALID_LENGTH", "length must be at least 1")
	}
	if opts.Length > 512 {
		return "", apperr.New(400, "INVALID_LENGTH", "length must not exceed 512")
	}

	classes := buildClasses(opts)
	pool := ""
	seeds := make([]byte, 0, len(classes))
	for _, c := range classes {
		if c == "" {
			continue
		}
		pool += c
		seeds = append(seeds, c[0])
	}
	if pool == "" {
		return "", apperr.New(400, "NO_CHARSET_SELECTED", "at least one character class must be enabled and non-empty after exclusions")
	}

	result := make([]byte, opts.Length)
	for i := range result {
		c, err := randomByte(pool)
		if err != nil {
			return "", err
		}
		result[i] = c
	}

	// Best-effort: ensure at least one char from each enabled, non-empty class.
	for i, seedClass := range classes {
		if seedClass == "" || i >= len(result) {
			continue
		}
		c, err := randomByte(seedClass)
		if err != nil {
			return "", err
		}
		result[i] = c
	}
	if err := shuffle(result); err != nil {
		return "", err
	}

	return string(result), nil
}

func buildClasses(opts Options) []string {
	classes := []string{}
	if opts.Lowercase {
		classes = append(classes, excludeChars(Lowercase, opts))
	}
	if opts.Uppercase {
		classes = append(classes, excludeChars(Uppercase, opts))
	}
	if opts.Numbers {
		classes = append(classes, excludeChars(Numbers, opts))
	}
	if opts.Symbols {
		classes = append(classes, excludeChars(Symbols, opts))
	}
	return classes
}

func excludeChars(class string, opts Options) string {
	var b strings.Builder
	for _, r := range class {
		if opts.ExcludeConfusing && strings.ContainsRune(ConfusingChars, r) {
			continue
		}
		if opts.ExcludeAmbiguous && strings.ContainsRune(AmbiguousChars, r) {
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

func randomByte(pool string) (byte, error) {
	n, err := rand.Int(rand.Reader, big.NewInt(int64(len(pool))))
	if err != nil {
		return 0, err
	}
	return pool[n.Int64()], nil
}

func shuffle(b []byte) error {
	for i := len(b) - 1; i > 0; i-- {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(i+1)))
		if err != nil {
			return err
		}
		j := n.Int64()
		b[i], b[j] = b[j], b[i]
	}
	return nil
}

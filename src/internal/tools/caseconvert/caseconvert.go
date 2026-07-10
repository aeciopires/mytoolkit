// Package caseconvert implements the Case Converter tool's pure logic.
package caseconvert

import (
	"strings"
	"unicode"

	"github.com/aeciopires/mytoolkit/internal/apperr"
)

type Options struct {
	Mode string `json:"mode"` // sentence | upper | lower | title | mixed | inverse
}

func Convert(input []byte, opts Options) (string, error) {
	if err := apperr.OneOf("mode", opts.Mode, "sentence", "upper", "lower", "title", "mixed", "inverse"); err != nil {
		return "", err
	}
	text := string(input)
	switch opts.Mode {
	case "sentence":
		return sentenceCase(text), nil
	case "upper":
		return strings.ToUpper(text), nil
	case "lower":
		return strings.ToLower(text), nil
	case "title":
		return titleCase(text), nil
	case "mixed":
		return mixedCase(text), nil
	case "inverse":
		return inverseCase(text), nil
	}
	return "", apperr.New(400, "UNSUPPORTED_MODE", "unsupported mode")
}

func sentenceCase(text string) string {
	lower := strings.ToLower(text)
	runes := []rune(lower)
	capitalizeNext := true
	for i, r := range runes {
		if capitalizeNext && unicode.IsLetter(r) {
			runes[i] = unicode.ToUpper(r)
			capitalizeNext = false
		}
		if r == '.' || r == '!' || r == '?' {
			capitalizeNext = true
		}
	}
	return string(runes)
}

func titleCase(text string) string {
	words := strings.Fields(text)
	for i, w := range words {
		lw := strings.ToLower(w)
		runes := []rune(lw)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
		}
		words[i] = string(runes)
	}
	return strings.Join(words, " ")
}

func mixedCase(text string) string {
	runes := []rune(text)
	for i, r := range runes {
		if i%2 == 0 {
			runes[i] = unicode.ToUpper(r)
		} else {
			runes[i] = unicode.ToLower(r)
		}
	}
	return string(runes)
}

func inverseCase(text string) string {
	runes := []rune(text)
	for i, r := range runes {
		switch {
		case unicode.IsUpper(r):
			runes[i] = unicode.ToLower(r)
		case unicode.IsLower(r):
			runes[i] = unicode.ToUpper(r)
		}
	}
	return string(runes)
}

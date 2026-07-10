// Package textcount implements the Character, Word & Line Counter tool's
// pure logic.
package textcount

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type Counts struct {
	Characters         int `json:"characters"`
	CharactersNoSpaces int `json:"characters_no_spaces"`
	Words              int `json:"words"`
	Lines              int `json:"lines"`
}

type Options struct{}

func Count(input []byte, _ Options) (Counts, error) {
	text := strings.ReplaceAll(string(input), "\r\n", "\n")

	characters := utf8.RuneCountInString(text)
	noSpaces := 0
	for _, r := range text {
		if !unicode.IsSpace(r) {
			noSpaces++
		}
	}
	words := len(strings.Fields(text))

	lines := 0
	if len(text) > 0 {
		lines = strings.Count(text, "\n") + 1
		if strings.HasSuffix(text, "\n") {
			lines--
		}
	}

	return Counts{
		Characters:         characters,
		CharactersNoSpaces: noSpaces,
		Words:              words,
		Lines:              lines,
	}, nil
}

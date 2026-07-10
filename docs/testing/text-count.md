<!-- TOC -->

- [Character, Word \& Line Counter — Testing](#character-word--line-counter--testing)

<!-- TOC -->

# Character, Word & Line Counter — Testing

```
$ cd src && go test ./internal/tools/textcount/... -v
--- PASS: TestCount (0.00s)
    --- PASS: TestCount/simple_sentence
    --- PASS: TestCount/two_lines_with_trailing_newline
    --- PASS: TestCount/two_lines_no_trailing_newline
    --- PASS: TestCount/empty_input
    --- PASS: TestCount/whitespace_only
    --- PASS: TestCount/crlf_normalized
    --- PASS: TestCount/unicode_characters
PASS
```

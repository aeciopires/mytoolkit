<!-- TOC -->

- [Character, Word \& Line Counter — Testing](#character-word--line-counter--testing)

<!-- TOC -->

# Character, Word & Line Counter — Testing

```
$ cd app && go test ./internal/tools/textcount/... -v
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

## Web UI verification (manual/scripted, no Go test coverage)

The page's "Clear" button (`internal/web/templates/tools/text-count.html`) is frontend-only. Verified with a real browser (Playwright driving the actual binary): typing text updates all four stats via the debounced `POST /api/v1/tools/text-count` call; clicking "Clear" empties the input textarea and resets Characters/Characters (no spaces)/Words/Lines back to `0`, with no console errors, in both light and dark themes.

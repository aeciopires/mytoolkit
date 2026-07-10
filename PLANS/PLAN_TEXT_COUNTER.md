<!-- TOC -->

- [PLAN\_TEXT\_COUNTER](#plan_text_counter)
  - [Description](#description)
  - [Business logic](#business-logic)
  - [CLI](#cli)
  - [REST](#rest)
  - [Web UI](#web-ui)
  - [Metrics](#metrics)
  - [Unit tests](#unit-tests)
  - [Documentation](#documentation)
  - [Skill](#skill)
  - [New dependencies](#new-dependencies)

<!-- TOC -->

# PLAN_TEXT_COUNTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the Character, Word & Line Counter feature. Tool slug: `text-count`.

## Description

Counts characters, words, and lines in arbitrary text.

## Business logic

Package: `internal/tools/textcount/textcount.go`.

```go
package textcount

type Counts struct {
    Characters      int `json:"characters"`       // count in runes (unicode-correct)
    CharactersNoSpaces int `json:"characters_no_spaces"`
    Words           int `json:"words"`
    Lines           int `json:"lines"`
}

func Count(input []byte) (Counts, error)
```

Implementation:
- `Characters`: counted with `utf8.RuneCountInString`, not `len()`, so multi-byte unicode characters count as 1 (documented explicitly since this is the most common bug in naive implementations).
- `CharactersNoSpaces`: same rune count, excluding `unicode.IsSpace` runes.
- `Words`: `strings.Fields` (splits on any run of whitespace, matches typical word-counting semantics used by editors).
- `Lines`: count of `\n`-separated segments; an empty input has 0 lines, a non-empty input with no trailing newline still counts as 1 line, consistent with common text-editor status bars — documented precisely since "line counting" has several reasonable conventions.

Edge cases:
- Empty input → all counts 0, not an error (unlike other tools, empty text is a valid, meaningful input here).
- Input with only whitespace → 0 words, but non-zero character/line counts.
- CRLF line endings (`\r\n`) → normalized so they don't double-count lines.

## CLI

```
mytoolkit text-count --in <file|->
```

Example:
```
$ printf 'Hello world\nSecond line\n' | mytoolkit text-count
characters: 23
characters_no_spaces: 21
words: 4
lines: 2
```

## REST

`POST /api/v1/tools/text-count`

Request:
```json
{ "input": "Hello world\nSecond line\n" }
```

Success (200):
```json
{
  "success": true,
  "data": { "characters": 23, "characters_no_spaces": 21, "words": 4, "lines": 2 },
  "meta": { "tool": "text-count", "duration_ms": 0.04 }
}
```

This tool never returns an error for empty/whitespace-only input — errors are reserved for transport-level issues (e.g. malformed request JSON), not for the text content itself.

## Web UI

- Single `<textarea>` input.
- Live-updating stat row below it (characters / characters without spaces / words / lines), recalculated on every keystroke — this can run **entirely client-side in JS** without a round trip, but still POSTs to the REST endpoint so usage metrics are recorded consistently with the other tools (debounced, e.g. every 500ms of inactivity, to avoid excessive requests while still counting the interaction as "usage").
- A **Clear** button empties the input, resets all four stats to `0`, and clears any error banner — added after the fact since the page originally had no way to reset without manually selecting and deleting the textarea contents.

## Metrics

Shared `tool="text-count"` label; no custom metric.

## Unit tests

`internal/tools/textcount/textcount_test.go`:
- Simple ASCII sentence.
- Unicode text (e.g. emoji, accented characters) — character count matches rune count, not byte length.
- Empty input → all zeros, no error.
- Whitespace-only input → 0 words, correct character/line counts.
- CRLF input → line count matches LF-equivalent input.
- Trailing newline vs. no trailing newline → documented line-count behavior verified.

## Documentation

- `docs/api/text-count.md`, `docs/cli/text-count.md`, `docs/testing/text-count.md`.

## Skill

`.skills/text-count/SKILL.md` — triggers on "implement text counter"; documents the rune-vs-byte counting rule and the line-counting convention, links to this plan.

## New dependencies

None — `strings`, `unicode`, `unicode/utf8` from the standard library.

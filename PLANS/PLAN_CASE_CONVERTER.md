<!-- TOC -->

- [PLAN\_CASE\_CONVERTER](#plan_case_converter)
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

# PLAN_CASE_CONVERTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the Case Converter feature. Tool slug: `case-convert`.

This feature was added after the initial planning pass and is not part of the original `README.md` feature list; `README.md` will be updated during implementation to include it as an 11th tool (see `PLAN_ARCHITECTURE.md`'s Assumptions section).

## Description

Converts input text between six casing styles: Sentence case, UPPER CASE, lower case, Title Case, Mixed/Alternating Case, and Inverse/Toggle Case.

## Business logic

Package: `internal/tools/caseconvert/caseconvert.go`.

```go
package caseconvert

type Mode string
// "sentence" | "upper" | "lower" | "title" | "mixed" | "inverse"

type Options struct {
    Mode Mode
}

func Convert(input string, opts Options) (string, error)
```

All six modes operate on `[]rune(input)`, not raw bytes, so multi-byte unicode characters are handled correctly. Exact semantics per mode:

- **`sentence`** (Sentence case): lowercases everything, then uppercases the first letter of the input and the first letter following each sentence-terminating rune (`.`, `!`, `?`) once whitespace has been skipped. Example: `"hello world. this IS a test!"` → `"Hello world. This is a test!"`.
- **`upper`** (UPPER CASE): every letter uppercased, via `unicode.ToUpper` per rune (through `strings.ToUpper`). Example: `"Hello World"` → `"HELLO WORLD"`.
- **`lower`** (lower case): every letter lowercased (`strings.ToLower`). Example: `"Hello World"` → `"hello world"`.
- **`title`** (Title Case): splits on whitespace (`strings.Fields`, whitespace runs collapsed — documented behavior, not whitespace-preserving), uppercases the first letter of each word and lowercases the rest, rejoins with single spaces. Example: `"hello WORLD example"` → `"Hello World Example"`.
- **`mixed`** (MiXeD CaSe / Alternating Case): iterates every rune in the input by absolute position (position 0-indexed, counting *all* characters including spaces and punctuation, not just letters); a rune at an even position is uppercased, a rune at an odd position is lowercased, non-letter runes pass through unchanged but still occupy a position, which is what produces the exact `"MiXeD CaSe"` pattern from the space-separated example given in the requirements: position 0 `M` upper, 1 `i` lower, 2 `X` upper, 3 `e` lower, 4 `D` upper, 5 ` ` (space, no-op), 6 `C` upper, 7 `a` lower, 8 `S` upper, 9 `e` lower.
- **`inverse`** (iNvErSe cAsE / Toggle Case): for each rune, swaps its *own* case independent of position — uppercase becomes lowercase and vice versa via `unicode.IsUpper`/`unicode.IsLower` + `unicode.ToLower`/`unicode.ToUpper`; non-letters are unchanged. Example: `"Hello World"` → `"hELLO wORLD"`.

Both `mixed` and `inverse` are fully deterministic (no randomness), which keeps them easy to unit test and makes REST/CLI output reproducible for the same input.

Edge cases:
- Empty input → empty output, not an error (same convention as Text Counter and URL Encode/Decode: casing empty text is well-defined and harmless).
- Unknown/unsupported `Mode` value → `UNSUPPORTED_MODE`, HTTP 400, listing the six valid values in the error message.
- Input with no letters (e.g. only digits/punctuation) → returned unchanged for all modes except `mixed`, which still applies its positional pass-through logic (a no-op here since there are no letters to case).

## CLI

```
mytoolkit case-convert --mode sentence|upper|lower|title|mixed|inverse --in <file|-> [--out <file|->]
```

`--mode` is required; `--help` lists all six modes with a one-line example each.

Example:
```
$ echo 'hello WORLD example' | mytoolkit case-convert --mode title
Hello World Example

$ echo 'hello world' | mytoolkit case-convert --mode mixed
HeLlO WoRlD
```

## REST

`POST /api/v1/tools/case-convert`

Request:
```json
{ "input": "hello world. this IS a test!", "options": { "mode": "sentence" } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "Hello world. This is a test!" },
  "meta": { "tool": "case-convert", "duration_ms": 0.03 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "UNSUPPORTED_MODE", "message": "mode must be one of: sentence, upper, lower, title, mixed, inverse" } }
```

## Web UI

- Input `<textarea>` + output `<textarea readonly>`.
- Mode selector as six buttons/pills labeled exactly as in the requirements: "Sentence case", "UPPER CASE", "lower case", "Title Case", "MiXeD CaSe", "iNvErSe cAsE" — each button's own label is rendered in its target casing style, doubling as a live preview of what the mode does.
- Live conversion on input change or mode change via debounced `fetch()`.
- "Copy to clipboard" and "Reset" buttons.

## Metrics

Shared `tool="case-convert"` label; no custom metric (mode choice is not added as a metric label to avoid unnecessary cardinality, consistent with `PLAN_HASH_GENERATOR.md`'s algorithm-label decision).

## Unit tests

`internal/tools/caseconvert/caseconvert_test.go`:
- `sentence`: multiple sentences, leading/trailing whitespace, consecutive terminators (`"Really?! Yes."`).
- `upper` / `lower`: simple ASCII and unicode (accented letters) input.
- `title`: multiple words, extra internal whitespace collapsed, single-word input.
- `mixed`: exact reproduction of the `"MiXeD CaSe"` pattern from `"mixed case"` input; verifies non-letter characters occupy a position without changing case.
- `inverse`: round-trip check — applying `inverse` twice returns the original input; explicit case `"MiXeD CaSe"` → `"mIxEd cAsE"`.
- Empty input → empty output, no error, for every mode.
- Unsupported mode string → error.
- Unicode input (e.g. `"café ÀÉ"`) cased correctly for each mode without corrupting multi-byte runes.

## Documentation

- `docs/api/case-convert.md`, `docs/cli/case-convert.md`, `docs/testing/case-convert.md`.

## Skill

`.skills/case-convert/SKILL.md` — triggers on "implement case converter"; documents the exact per-mode algorithms above (especially the position-based `mixed` rule and the self-case-swap `inverse` rule, which are easy to get subtly wrong), links to this plan.

## New dependencies

None — `strings`, `unicode`, `unicode/utf8` from the standard library.

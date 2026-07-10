<!-- TOC -->

- [PLAN\_PASSWORD\_GENERATOR](#plan_password_generator)
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

# PLAN_PASSWORD_GENERATOR

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the Password Generator feature. Tool slug: `password-gen`.

## Description

Generates strong, customizable random passwords (length, character classes).

## Business logic

Package: `internal/tools/password/password.go`.

```go
package password

type Options struct {
    Length           int  // default 16
    Lowercase        bool // default true  - a b c d ...
    Uppercase        bool // default true  - A B C D ...
    Numbers          bool // default true  - 1 2 3 4 ...
    Symbols          bool // default false - ! # $ % & * + - = ? @ ^ _ { } [ ] ( ) / ' " ` ~ , ; : . < > \
    ExcludeConfusing bool // default false - removes i l L 1 o 0 O from the built pool
    ExcludeAmbiguous bool // default false - removes { } [ ] ( ) / \ ' " ` ~ , ; : . < > from the built pool
}

func Generate(opts Options) (string, error)
```

Character sets are defined as package constants so the CLI help, REST docs, and web UI tooltips all reference the exact same characters:

```go
const (
    Lowercase        = "abcdefghijklmnopqrstuvwxyz"
    Uppercase        = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"
    Numbers          = "0123456789"
    Symbols          = "!#$%&*+-=?@^_{}[]()/'\"`~,;:.<>\\"
    ConfusingChars   = "ilL1o0O"
    AmbiguousChars   = "{}[]()/\\'\"`~,;:.<>"
)
```

Implementation uses `crypto/rand` exclusively (never `math/rand`) since generated passwords are a security-sensitive artifact. Character pool is built by concatenating the enabled classes (`Lowercase`/`Uppercase`/`Numbers`/`Symbols`), then, if `ExcludeConfusing`/`ExcludeAmbiguous` are set, every rune present in `ConfusingChars`/`AmbiguousChars` is removed from the pool. Note `AmbiguousChars` is a subset of `Symbols`, so enabling `Symbols` + `ExcludeAmbiguous` together still leaves a "safe" symbol subset (`! # $ % & * + - = ? @ ^ _`) rather than eliminating symbols entirely. Each character is chosen via `rand.Int(rand.Reader, big.NewInt(int64(len(pool))))` (or an equivalent unbiased rejection-sampling helper) to avoid modulo bias.

Edge cases:
- `Length < 1` → error `INVALID_LENGTH` (recommend a sane cap too, e.g. reject `Length > 512`, HTTP 400).
- No character class enabled → error `NO_CHARSET_SELECTED`, HTTP 400 (there is nothing to draw from).
- Exclusions reduce the pool to empty (e.g. an enabled class whose characters are entirely covered by the excluded sets) → same `NO_CHARSET_SELECTED` error, HTTP 400, since there is nothing left to draw from after exclusion.
- Ensure at least one character from each *enabled* class appears when `Length` allows it, seeding from that class's pool *after* exclusions are applied — documented as a best-effort guarantee, not applied when `Length` is smaller than the number of enabled classes, and a class contributes no seed character if exclusions emptied it entirely.

## CLI

```
mytoolkit password-gen [--length N] [--lowercase] [--uppercase] [--numbers] [--symbols] [--exclude-confusing] [--exclude-ambiguous] [--out <file|->]
```

All charset flags default to the same defaults as `Options` above; passing any charset flag explicitly overrides the defaults (documented clearly in `--help` to avoid surprising "why did my flag do nothing" issues). `--exclude-confusing` and `--exclude-ambiguous` default to `false` and can be combined with any charset flags; `--help` lists the exact characters each one removes (`i l L 1 o 0 O` and `{ } [ ] ( ) / \ ' " ` ~ , ; : . < >` respectively).

Example:
```
$ mytoolkit password-gen --length 20 --symbols
Kx8#mQ2!vL9pR4$wN7&z

$ mytoolkit password-gen --length 20 --symbols --exclude-confusing --exclude-ambiguous
Kx8#mQ2!vP9rR4$wN7&z
```

## REST

`POST /api/v1/tools/password-gen`

Request:
```json
{
  "options": {
    "length": 20,
    "lowercase": true,
    "uppercase": true,
    "numbers": true,
    "symbols": true,
    "exclude_confusing": false,
    "exclude_ambiguous": true
  }
}
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "Kx8#mQ2!vL9pR4$wN7&z" },
  "meta": { "tool": "password-gen", "duration_ms": 0.05 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "NO_CHARSET_SELECTED", "message": "at least one character class must be enabled" } }
```

Note: this endpoint has no `input` field in the request (no text to transform), unlike most other tools — documented explicitly as the request-shape exception for this feature.

## Web UI

- Length slider/number input (1–128 in the UI, backend allows up to 512).
- Checkboxes: "Include lowercase letters" (`a b c d ...`), "Include uppercase letters" (`A B C D ...`), "Include numbers" (`1 2 3 4 ...`), "Include symbols" (`! # $ % & * + - = ? @ ^ _ { } [ ] ( ) / ' " ` ~ , ; : . < > \`).
- Two additional checkboxes, off by default: "Exclude confusing characters" (tooltip lists `i l L 1 o 0 O`) and "Exclude ambiguous characters" (tooltip lists `{ } [ ] ( ) / \ ' " ` ~ , ; : . < >`).
- Generated password shown in a readonly field with "Copy" and "Regenerate" buttons.
- Regenerates live whenever a control changes (debounced `fetch()`), no explicit submit needed.
- Optional client-side strength indicator (entropy estimate computed from length × charset size — purely cosmetic, not sent to the server).

## Metrics

Shared `tool="password-gen"` label; no custom metric.

## Unit tests

`internal/tools/password/password_test.go`:
- Generated length matches requested length for various `Length` values.
- Output only contains characters from the enabled classes (property-based check over many runs).
- No charset enabled → error.
- `Length < 1` → error.
- `ExcludeConfusing` true → output never contains any of `i l L 1 o 0 O`, across many runs and all charset combinations.
- `ExcludeAmbiguous` true → output never contains any of `{ } [ ] ( ) / \ ' " ` ~ , ; : . < >`, across many runs.
- `Symbols` + `ExcludeAmbiguous` both true → output symbols are restricted to the safe subset `! # $ % & * + - = ? @ ^ _`.
- Two consecutive calls with the same options produce different outputs (randomness sanity check, not a strict guarantee test — assert non-equal with acceptable flake tolerance documented, or repeat N times and require at least one difference).

## Documentation

- `docs/api/password-gen.md`, `docs/cli/password-gen.md`, `docs/testing/password-gen.md`.

## Skill

`.skills/password-gen/SKILL.md` — triggers on "implement password generator"; documents the `crypto/rand`-only requirement and unbiased sampling approach, links to this plan.

## New dependencies

None — `crypto/rand` and `math/big` from the standard library.

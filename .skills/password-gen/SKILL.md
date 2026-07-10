---
name: password-gen
description: Implement or modify the Password Generator tool (internal/tools/password) — crypto/rand-based password generation with charset and exclusion controls. Trigger on "implement password generator".
---

# Password Generator

`src/internal/tools/password/password.go`, `func Generate(opts Options) (string, error)`. Uses `crypto/rand` exclusively (never `math/rand`) via unbiased `rand.Int` sampling — this is a hard requirement, not a style preference, since output is security-sensitive.

`Options` has no `input`/text field — REST/CLI wiring is bespoke (`src/internal/cli/password.go`), not `handlers.Wrap`/`newTextToolCommand`, because those assume a text-transform shape.

Charset constants (`Lowercase`, `Uppercase`, `Numbers`, `Symbols`, `ConfusingChars`, `AmbiguousChars`) are package-level so CLI help/REST docs/web tooltips stay in sync. `ExcludeConfusing`/`ExcludeAmbiguous` filter the built pool *after* concatenating enabled classes — `AmbiguousChars` is a subset of `Symbols`, so `Symbols=true` + `ExcludeAmbiguous=true` leaves the safe subset `!#$%&*+-=?@^_`.

`Length < 1` (including the JSON zero-value, i.e. omitted) is a hard `INVALID_LENGTH` error — there is deliberately no "default to 16" fallback in the pure function; CLI flag defaults (`--length 16`) provide that UX instead.

Plan: `PLANS/PLAN_PASSWORD_GENERATOR.md`. Docs: `docs/api|cli|testing/password-gen.md`.

---
name: text-count
description: Implement or modify the Character, Word & Line Counter tool (internal/tools/textcount). Trigger on "implement text counter".
---

# Character, Word & Line Counter

`src/internal/tools/textcount/textcount.go`, `func Count(input []byte, opts Options) (Counts, error)`. `Counts` has 4 fields, not a single string — bespoke REST/CLI wiring (`src/internal/cli/textcount.go`), not `handlers.Wrap`.

Rules: `characters` via `utf8.RuneCountInString` (never `len()` — multi-byte runes must count as 1). `Lines` = `strings.Count(text,"\n")+1`, minus 1 if the text ends with `\n` (so both `"a\nb"` and `"a\nb\n"` report 2 lines, matching common editor status-bar behavior). Never errors — empty/whitespace input is valid.

Plan: `PLANS/PLAN_TEXT_COUNTER.md`. Docs: `docs/api|cli|testing/text-count.md`.

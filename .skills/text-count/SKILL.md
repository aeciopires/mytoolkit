---
name: text-count
description: Implement or modify the Character, Word & Line Counter tool (internal/tools/textcount). Trigger on "implement text counter".
---

# Character, Word & Line Counter

`app/internal/tools/textcount/textcount.go`, `func Count(input []byte, opts Options) (Counts, error)`. `Counts` has 4 fields, not a single string — bespoke REST/CLI wiring (`app/internal/cli/textcount.go`), not `handlers.Wrap`.

Rules: `characters` via `utf8.RuneCountInString` (never `len()` — multi-byte runes must count as 1). `Lines` = `strings.Count(text,"\n")+1`, minus 1 if the text ends with `\n` (so both `"a\nb"` and `"a\nb\n"` report 2 lines, matching common editor status-bar behavior). Never errors — empty/whitespace input is valid.

Web page (`internal/web/templates/tools/text-count.html`) is fully custom (no `tool-panel.html` partial, no output textarea — just an input and a live stats row), so it has its own `#clear-btn` wired directly in an inline script, not the shared `tool-common.js` data-action machinery. `Clear` empties the input, resets all four stat spans to `0`, clears the error banner, and refocuses the input.

MCP: `text-count` tool (`app/internal/mcp/text_count.go`), returns `textcount.Counts` as structured output. Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_TEXT_COUNTER.md`. Docs: `docs/api|cli|testing/text-count.md`.

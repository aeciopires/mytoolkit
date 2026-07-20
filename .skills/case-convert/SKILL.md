---
name: case-convert
description: Implement or modify the Case Converter tool (internal/tools/caseconvert) — sentence/upper/lower/title/mixed/inverse. Trigger on "implement case converter".
---

# Case Converter

`app/internal/tools/caseconvert/caseconvert.go`, `func Convert(input []byte, opts Options) (string, error)`.

`mixed` mode alternates case by **absolute character position** (including spaces/punctuation in the position count, though they don't themselves get cased) — this is what produces the exact `"MiXeD CaSe"` pattern from `"mixed case"`. `inverse` mode swaps each letter's own case independently of position, and is self-inverting (`inverse(inverse(x)) == x`). Both are deterministic — no randomness — which is what makes them reproducible in tests/REST/CLI.

Fully generic wiring via `handlers.Wrap` / `newTextToolCommand`.

MCP: `case-convert` tool (`app/internal/mcp/case_convert.go`). Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_CASE_CONVERTER.md`. Docs: `docs/api|cli|testing/case-convert.md`.

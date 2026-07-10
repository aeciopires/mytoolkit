---
name: json-format
description: Implement or modify the JSON Formatter tool (internal/tools/jsonformat) — pretty-print/minify JSON. Trigger on "implement JSON formatter", "add JSON pretty-print/minify".
---

# JSON Formatter

`src/internal/tools/jsonformat/jsonformat.go`, `func Format(input []byte, opts Options) (string, error)`. Uses stdlib `encoding/json.Indent`/`json.Compact` only — both also validate syntax as a side effect.

Fully generic: uses `handlers.Wrap` for REST and `newTextToolCommand` for CLI (see `src/internal/cli/jsonformat.go`), no bespoke wiring needed.

Plan: `PLANS/PLAN_JSON_FORMATTER.md`. Docs: `docs/api|cli|testing/json-format.md`.

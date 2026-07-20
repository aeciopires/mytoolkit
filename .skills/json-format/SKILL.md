---
name: json-format
description: Implement or modify the JSON Formatter tool (internal/tools/jsonformat) — pretty-print/minify JSON. Trigger on "implement JSON formatter", "add JSON pretty-print/minify".
---

# JSON Formatter

`app/internal/tools/jsonformat/jsonformat.go`, `func Format(input []byte, opts Options) (string, error)`. Uses stdlib `encoding/json.Indent`/`json.Compact` only — both also validate syntax as a side effect.

REST and CLI are fully generic: `handlers.Wrap` for REST and `newTextToolCommand` for CLI (see `app/internal/cli/jsonformat.go`), no bespoke wiring needed.

## Web page is client-side, not a `Format()` caller

Unlike REST/CLI, the web page (`app/internal/web/templates/tools/json-format.html`) does **not** call `jsonformat.Format` or `POST /api/v1/tools/json-format` at all. It runs entirely in the browser via native `JSON.parse()`/`JSON.stringify()`, because the feature spec requires validation results and error messages to exactly match real browser/Node.js `JSON.parse()` behavior (e.g. `Unexpected token } in JSON at position 45`) — the Go backend's `encoding/json` errors are similar but not textually identical, so proxying through REST would break that guarantee.

- `.tool-panel` carries `data-client-side` (same convention as `json-toon`/`json-tree` — see `tool-common.js`'s comment on that attribute) to skip the shared fetch-on-input wiring.
- Three actions, each its own button: **Validate JSON** (parse only, green `.success-banner` or red `.error-banner`, doesn't touch the output field), **Beautify** (`JSON.stringify(parsed, null, 2)`), **Minify** (`JSON.stringify(parsed)`). Copy/Clear are custom buttons in this page too (not the shared `tool-panel.html` partial's Copy/Reset), so exact labels/order match the spec.
- If you add a `.success-banner` elsewhere, reuse the CSS added in `app.css` next to `.error-banner` rather than inventing new colors.

Because of this, `mytoolkit_tool_usage_total{tool="json-format"}` only reflects REST/CLI usage, not web page usage — same caveat as `json-toon` (see its `SKILL.md`).

MCP: `json-format` tool (`app/internal/mcp/json_format.go`) — a full Go implementation, same as REST/CLI, unaffected by the web page's client-side-only behavior above. Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_JSON_FORMATTER.md`. Docs: `docs/api|cli|testing/json-format.md`.

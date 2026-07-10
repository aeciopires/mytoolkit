<!-- TOC -->

- [PLAN\_JSON\_FORMATTER](#plan_json_formatter)
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

# PLAN_JSON_FORMATTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the JSON Formatter feature. Tool slug: `json-format`.

## Description

Pretty-prints or minifies a JSON document to improve or reduce readability.

**Revised after initial implementation**, per a detailed feature spec: the web page's validation must exactly match real-world `JSON.parse()` behavior (the same parsing engine used by Node.js and browsers), which the Go backend's `encoding/json` error messages do not — they're similar but not textually identical. The web page therefore converts entirely client-side using native `JSON.parse()`/`JSON.stringify()` (no `fetch()` call from the interactive tool), following the same `data-client-side` pattern established by JSON to TOON Converter and JSON Tree Viewer. REST and CLI are unchanged: both remain full Go implementations backed by `jsonformat.Format`, for scripted/automated use.

## Business logic

Package: `internal/tools/jsonformat/jsonformat.go`.

```go
package jsonformat

type Mode string // "pretty" | "minify"

type Options struct {
    Mode   Mode
    Indent int // spaces, used only when Mode == "pretty", default 2
}

func Format(input []byte, opts Options) (string, error)
```

Implementation: `json.Indent` for pretty mode (validates syntax as a side effect), `json.Compact` for minify mode. Both operate on `[]byte`, both return `encoding/json` syntax errors on invalid input.

Edge cases:
- Empty input → `ErrEmptyInput`, HTTP 400.
- Malformed JSON → HTTP 400, `INVALID_JSON`.
- `Indent <= 0` when `Mode == "pretty"` → default to 2 rather than erroring (defensive default, not a hard failure).
- Already-minified input in minify mode → returned unchanged (idempotent).

## CLI

```
mytoolkit json-format --in <file|-> [--out <file|->] [--minify] [--indent N]
```

- `--in` / `--out`: default `-` (stdin/stdout).
- `--minify`: switches to minify mode (default is pretty).
- `--indent`: spaces for pretty mode, default `2`, ignored with `--minify`.

Example:
```
$ echo '{"a":1,"b":2}' | mytoolkit json-format
{
  "a": 1,
  "b": 2
}

$ echo '{"a": 1, "b": 2}' | mytoolkit json-format --minify
{"a":1,"b":2}
```

## REST

`POST /api/v1/tools/json-format`

Request:
```json
{ "input": "{\"a\":1,\"b\":2}", "options": { "mode": "pretty", "indent": 2 } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "{\n  \"a\": 1,\n  \"b\": 2\n}" },
  "meta": { "tool": "json-format", "duration_ms": 0.18 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "invalid character '}' looking for beginning of object key string" } }
```

## Web UI

- Input `<textarea>` + output `<textarea readonly>` side by side (or stacked on mobile, responsive via CSS grid).
- Actions: **Validate JSON** (parses the input via `JSON.parse()`; shows a green success banner on success, or the exact `JSON.parse()` error message — e.g. `Unexpected token } in JSON at position 45` — in the error banner on failure; does not touch the output field), **Beautify** (`JSON.stringify(parsed, null, 2)` into the output), **Minify** (`JSON.stringify(parsed)`, no whitespace, into the output), **Copy Output**, **Clear** (empties both fields and any banner).
- **No live-as-you-type conversion and no REST call from the interactive page** — `.tool-panel` carries `data-client-side` to opt out of `tool-common.js`'s fetch-on-input wiring; all three actions call `JSON.parse`/`JSON.stringify` directly in the browser. This is a deliberate deviation, driven by the requirement that validation results and error messages exactly match native browser/Node.js `JSON.parse()` behavior — routing through the Go backend's `encoding/json` would produce similar but not identical error text.
- Ctrl/Cmd+Enter in the input textarea also triggers Beautify, for keyboard workflows.
- A one-line note near the actions discloses that the page runs entirely client-side and links to the docs for CLI/REST scripting.
- Reuses the shared layout/theme/nav from `PLAN_ARCHITECTURE.md`; a new `.success-banner` style (green, light/dark variants) was added alongside the existing `.error-banner`.

## Metrics

Shared `tool="json-format"` label; no custom metric.

## Unit tests

`internal/tools/jsonformat/jsonformat_test.go`:
- Pretty with default indent.
- Pretty with custom indent (e.g. 2 or 4).
- Minify collapses whitespace.
- Minify is idempotent on already-minified input.
- Empty input → error.
- Malformed JSON → error.

The web page's Validate/Beautify/Minify/Copy/Clear buttons are frontend-only (native `JSON.parse`/`JSON.stringify`, no Go code path) and have no `go test` coverage — verified manually/with a scripted browser check instead (see `docs/testing/json-format.md`).

## Documentation

- `docs/api/json-format.md`, `docs/cli/json-format.md`, `docs/testing/json-format.md`, following the examples above.

## Skill

`.skills/json-format/SKILL.md` — triggers on "implement JSON formatter", "add JSON pretty-print/minify"; documents the `Mode`/`Options` API and links to this plan.

## New dependencies

None — `encoding/json` from the standard library.

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
- Controls: mode toggle (Pretty/Minify), indent stepper (1–8), "Copy to clipboard" and "Download .json" buttons, "Reset".
- Live formatting on input change via debounced `fetch()`.

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

## Documentation

- `docs/api/json-format.md`, `docs/cli/json-format.md`, `docs/testing/json-format.md`, following the examples above.

## Skill

`.skills/json-format/SKILL.md` — triggers on "implement JSON formatter", "add JSON pretty-print/minify"; documents the `Mode`/`Options` API and links to this plan.

## New dependencies

None — `encoding/json` from the standard library.

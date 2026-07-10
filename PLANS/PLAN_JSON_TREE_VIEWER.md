<!-- TOC -->

- [PLAN\_JSON\_TREE\_VIEWER](#plan_json_tree_viewer)
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

# PLAN_JSON_TREE_VIEWER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the JSON Tree Viewer feature. Tool slug: `json-tree`.

## Description

Parses raw JSON text and returns a navigable tree structure (nested nodes with type and value metadata) so the web UI can render an expandable/collapsible tree view, inspired by `10015.io/tools/json-tree-viewer`.

**Revised after initial implementation**, per a detailed feature spec: this is a comprehension/debugging tool for large, deeply-nested API responses (hundreds of lines), not a live-as-you-type formatter. The web page therefore requires an explicit "Generate Tree View" action (plus "Expand All"/"Collapse All"/"Clear") rather than converting on every keystroke; error messages must carry the exact line/column of the problem so users can locate malformed JSON without guessing; and the tree is color-coded by value type (VS Code–inspired) so type mistakes (e.g. the string `"null"` vs. the literal `null`) are visually obvious, not just readable.

## Business logic

Package: `internal/tools/jsontree/jsontree.go`.

```go
package jsontree

type NodeType string // "object" | "array" | "string" | "number" | "bool" | "null"

type Node struct {
    Key      string `json:"key,omitempty"`
    Type     NodeType `json:"type"`
    Value    any    `json:"value,omitempty"`
    Children []Node `json:"children,omitempty"`
}

type Options struct {
    // reserved for future display hints (e.g. max depth); empty for the MVP
}

func Parse(input []byte, opts Options) (Node, error)
```

Implementation uses `encoding/json` with `json.Decoder` (not `Unmarshal` into `map[string]any]` alone) so key order is preserved via `json.Token` streaming, since Go maps do not preserve insertion order and the tree view should show keys in source order.

Edge cases:
- Empty input → `ErrEmptyInput`, mapped to HTTP 400.
- Malformed JSON → wraps the `encoding/json` syntax error with a `(at line L, column C)` suffix (1-indexed), computed from `json.SyntaxError.Offset` when available or `dec.InputOffset()` otherwise, mapped to HTTP 400 (`INVALID_JSON`). This is a hard requirement, not a nice-to-have — the whole point of the tool is fast comprehension of large responses, and "unexpected end of JSON input" with no location forces the user to hunt for the problem manually.
- Trailing content after a complete top-level value (e.g. `{"a":1}garbage`, or two concatenated JSON values) → rejected explicitly via a `dec.More()` check after the top-level parse, not silently ignored.
- Deeply nested input → no artificial depth limit for MVP; document as a known limitation (potential stack growth) rather than adding complexity upfront.
- Very large numbers → preserved as `json.Number` (via decoder's `UseNumber()`) to avoid float64 precision loss.

## CLI

```
mytoolkit json-tree --in <file|-> [--out <file|->]
```

- `--in`: path to a JSON file, or `-` for stdin (default `-`).
- `--out`: path to write the tree as indented JSON, or `-` for stdout (default `-`).
- `--help` shows the above plus one example.

Example:

```
$ echo '{"a":1,"b":[true,null]}' | mytoolkit json-tree
{
  "type": "object",
  "children": [
    {
      "key": "a",
      "type": "number",
      "value": "1"
    },
    {
      "key": "b",
      "type": "array",
      "children": [
        { "type": "bool", "value": true },
        { "type": "null" }
      ]
    }
  ]
}
```

Errors report the exact position of the problem:

```
$ echo '{"a":}' | mytoolkit json-tree
Error: invalid character '}' looking for beginning of value (at line 1, column 6)
```

## REST

`POST /api/v1/tools/json-tree`

Request:
```json
{ "input": "{\"a\":1,\"b\":[true,null]}" }
```

Success response (200):
```json
{
  "success": true,
  "data": {
    "tree": {
      "type": "object",
      "children": [
        { "key": "a", "type": "number", "value": "1" },
        { "key": "b", "type": "array", "children": [
          { "type": "bool", "value": true },
          { "type": "null" }
        ]}
      ]
    }
  },
  "meta": { "tool": "json-tree", "duration_ms": 0.31 }
}
```

Error response (400) — request `{ "input": "{\"a\":}" }`:
```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "invalid character '}' looking for beginning of value (at line 1, column 6)" } }
```

## Web UI

- Left panel: `<textarea id="tool-input">` labeled "Raw JSON".
- Right panel: `<div id="tree-output">` rendered as nested `<details open>`/`<summary>` elements (native disclosure semantics, custom-styled — no native marker, a CSS-only `▶`/`▼` icon driven by the `[open]` attribute).
- Object/array summaries read exactly `Object {N keys}` / `Array [N items]` (singular "key"/"item" at N=1), not a generic `type(size)` placeholder.
- Actions: **Generate Tree View** (primary — calls `POST /api/v1/tools/json-tree`, the only trigger for a REST call on this page), **Expand All** / **Collapse All** (client-side only, toggle `.open` on every node, no re-fetch), **Clear** (empties input and output).
- **No live-as-you-type conversion** — `.tool-panel` carries `data-client-side` specifically to opt out of the shared `tool-common.js` fetch-on-input wiring (see `PLAN_ARCHITECTURE.md`'s Shared code and configuration reuse table), since re-parsing a large pasted API response on every keystroke is wasteful and janky. This is a deliberate deviation from most other tools' live-update pattern, justified by this tool's use case (large, deliberately-pasted responses, not short incremental text).
- Type-based syntax coloring: object/array keys, string/number/bool/null values each get a distinct color via `--json-key`/`--json-string`/`--json-number`/`--json-bool`/`--json-null` CSS custom properties (VS Code–inspired, separate light/dark values in `theme.css`) — so e.g. the string `"null"` (orange/red, quoted) is visually distinguishable from the literal `null` (gray, unquoted) at a glance, not just on close reading.
- A one-line tip near the input links to the JSON Formatter tool for a first-pass validate/pretty-print — this app has no separate "JSON Validator"/"JSONPath Tester" tool, so the UI doesn't reference tools that don't exist.
- Reuses the shared layout/theme/nav from `PLAN_ARCHITECTURE.md`.

## Metrics

Uses the shared `tool="json-tree"` label on `mytoolkit_http_requests_total`, `mytoolkit_http_request_duration_seconds`, and `mytoolkit_tool_usage_total`. No custom metric needed.

## Unit tests

Table-driven tests in `internal/tools/jsontree/jsontree_test.go`:
- Valid flat object.
- Valid nested object/array mix.
- Empty input → error.
- Malformed JSON (trailing comma, unclosed brace) → error.
- Error message includes `line`/`column` position, including for a problem on a line other than the first.
- Trailing data after a complete value (e.g. `{"a":1}garbage`, two concatenated values) → rejected, not silently ignored.
- Large integer (beyond float64 precision) preserved exactly via `json.Number`.
- Unicode string values preserved.

The web page's Expand All/Collapse All, Generate-on-click-not-on-keystroke, and color coding are frontend-only and have no `go test` coverage — verified manually/with a scripted browser check instead (see `docs/testing/json-tree.md`).

## Documentation

- `docs/api/json-tree.md` — endpoint, request/response examples above, error codes.
- `docs/cli/json-tree.md` — CLI usage/examples above.
- `docs/testing/json-tree.md` — how to run `go test ./internal/tools/jsontree/...` with sample output.

## Skill

`.skills/json-tree/SKILL.md` — triggers on "implement JSON tree viewer", "add tree node parsing"; summarizes the `Node`/`Options` API, the key-order-preservation requirement, and links to this plan.

## New dependencies

None — implemented with the standard library only (`encoding/json`).

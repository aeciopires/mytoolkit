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
- Malformed JSON → wraps the `encoding/json` syntax error with position info, mapped to HTTP 400 (`INVALID_JSON`).
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
    {"key": "a", "type": "number", "value": 1},
    {"key": "b", "type": "array", "children": [
      {"type": "bool", "value": true},
      {"type": "null"}
    ]}
  ]
}
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
        { "key": "a", "type": "number", "value": 1 },
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

Error response (400):
```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "unexpected end of JSON input" } }
```

## Web UI

- Left panel: `<textarea>` labeled "Paste raw JSON", with a "Reset" button.
- Right panel: rendered tree, collapsible per node (client-side JS toggles a `hidden` class on `<ul>` children), with object/array size shown next to the key (e.g. `b: Array(2)`).
- Live update on input change (debounced `fetch()` call to `/api/v1/tools/json-tree`), matching the reference site's "no separate submit button" pattern.
- Reuses the shared layout/theme/nav from `PLAN_ARCHITECTURE.md`.

## Metrics

Uses the shared `tool="json-tree"` label on `mytoolkit_http_requests_total`, `mytoolkit_http_request_duration_seconds`, and `mytoolkit_tool_usage_total`. No custom metric needed.

## Unit tests

Table-driven tests in `internal/tools/jsontree/jsontree_test.go`:
- Valid flat object.
- Valid nested object/array mix.
- Empty input → error.
- Malformed JSON (trailing comma, unclosed brace) → error.
- Large integer (beyond float64 precision) preserved exactly via `json.Number`.
- Unicode string values preserved.

## Documentation

- `docs/api/json-tree.md` — endpoint, request/response examples above, error codes.
- `docs/cli/json-tree.md` — CLI usage/examples above.
- `docs/testing/json-tree.md` — how to run `go test ./internal/tools/jsontree/...` with sample output.

## Skill

`.skills/json-tree/SKILL.md` — triggers on "implement JSON tree viewer", "add tree node parsing"; summarizes the `Node`/`Options` API, the key-order-preservation requirement, and links to this plan.

## New dependencies

None — implemented with the standard library only (`encoding/json`).

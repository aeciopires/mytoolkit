<!-- TOC -->

- [JSON Tree Viewer — REST API](#json-tree-viewer--rest-api)
  - [Request](#request)
  - [Success response (200)](#success-response-200)
  - [Error response (400)](#error-response-400)
  - [Workflow](#workflow)

<!-- TOC -->

# JSON Tree Viewer — REST API

`POST /api/v1/tools/json-tree`

Parses raw JSON into a navigable tree structure, preserving object key order. Called by the web page only when the user clicks "Generate Tree View" — not on every keystroke — since large pasted responses make live-as-you-type re-parsing wasteful.

## Request

```json
{ "input": "{\"a\":1,\"b\":[true,null]}" }
```

## Success response (200)

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
  "meta": { "tool": "json-tree", "duration_ms": 0.08 }
}
```

## Error response (400)

Every error message includes the exact 1-indexed line and column of the problem, not just a bare parser message. Request:

```json
{ "input": "{\"a\":}" }
```

Response:

```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "invalid character '}' looking for beginning of value (at line 1, column 6)" } }
```

A second, multi-line example — request `{ "input": "{\n  \"a\": 1,\n  \"b\":\n}" }` — reports the problem on the line it actually occurs on:

```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "invalid character '}' looking for beginning of value (at line 4, column 1)" } }
```

Content after a complete JSON value (e.g. two top-level values, or trailing garbage like `{"a":1}garbage`) is rejected rather than silently ignored:

```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "unexpected extra data after the JSON value (at line 1, column 8)" } }
```

Error codes: `EMPTY_INPUT`, `INVALID_JSON`.

## Workflow

```mermaid
sequenceDiagram
    participant Client
    participant Router as chi router
    participant Handler as json-tree handler
    participant Tool as internal/tools/jsontree

    Client->>Router: POST /api/v1/tools/json-tree
    Router->>Handler: dispatch (metrics + logging middleware)
    Handler->>Tool: Parse(input)
    alt valid JSON, nothing trailing
        Tool-->>Handler: Node tree (key order preserved)
        Handler-->>Client: 200 {success, data.tree}
    else empty, invalid, or trailing data
        Tool-->>Handler: apperr.Error with line/column position
        Handler-->>Client: 400 {success:false, error}
    end
```

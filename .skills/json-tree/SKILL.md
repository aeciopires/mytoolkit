---
name: json-tree
description: Implement or modify the JSON Tree Viewer tool (internal/tools/jsontree) — parsing raw JSON into a key-order-preserving navigable tree, with a color-coded, expand/collapse-all web UI. Trigger on "implement JSON tree viewer", "add tree node parsing", "fix json-tree".
---

# JSON Tree Viewer

Business logic: `app/internal/tools/jsontree/jsontree.go`, `func Parse(input []byte, opts Options) (Node, error)`.

Key implementation rule: object keys must keep source order. Do **not** decode into `map[string]any` — Go maps are unordered. Instead stream `json.Token`s from a `json.Decoder` (see `parseValue` in `jsontree.go`) and use `dec.UseNumber()` so large integers aren't rounded through `float64`.

`Node{Key, Type, Value, Children}` — `Type` is one of `object|array|string|number|bool|null`.

## Error messages carry position

`Parse` doesn't just wrap the raw `encoding/json` error — `positionedError()` appends a 1-indexed `(at line L, column C)` suffix, computed from `json.SyntaxError.Offset` when the error is a syntax error (most accurate) or from `dec.InputOffset()` otherwise (truncated input, trailing-data checks). `Parse` also explicitly rejects trailing content after a complete top-level value (`dec.More()` check) — without this, `{"a":1}garbage` or two concatenated JSON values would silently parse only the first and ignore the rest. If you change the error-wrapping logic, keep both: the position suffix and the trailing-data check are both load-bearing, tested in `jsontree_test.go` (`TestParseErrorIncludesPosition`, `TestParseErrorPositionOnLaterLine`, `TestParseRejectsTrailingData`).

## Wiring

`app/internal/cli/jsontree.go` registers both the `json-tree` CLI subcommand and the `json-tree` REST handler (bespoke, not `handlers.Wrap`, since the response shape is `{tree: Node}` not `{output: string}`).

Web page: `app/internal/web/templates/tools/json-tree.html` — a custom recursive JS renderer, **not** the generic `tool-common.js` fetch-on-input flow. Its `.tool-panel` carries `data-client-side` — not because this tool avoids the network (it doesn't; `generate()` does call `POST /api/v1/tools/json-tree`), but to opt out of `tool-common.js`'s automatic fetch-on-every-keystroke wiring. Conversion only happens when the user clicks "Generate Tree View" (or Ctrl/Cmd+Enter in the textarea) — large pasted API responses make live-as-you-type re-parsing wasteful and janky. See `tool-common.js`'s comment on `data-client-side` for the two different reasons a tool can use that attribute.

## Web UI details

- Nodes render as `<details class="json-node" open>` with a custom `▶`/`▼` toggle icon (native `<details>` markers are hidden via CSS; the icon glyph is driven purely by the `details[open]` CSS selector, no JS state syncing needed).
- Object/array summaries read exactly `Object {N keys}` / `Array [N items]` (singular "key"/"item" at N=1) — don't drift from this wording, it's specified.
- "Expand All"/"Collapse All" just set `.open = true/false` on every `details.json-node` in the output — no re-fetch.
- "Clear" empties both the input textarea and the tree output.
- Type-based syntax coloring (`.json-key`, `.json-value.json-string/.json-number/.json-bool/.json-null`) uses the `--json-*` CSS custom properties in `theme.css` (VS Code-inspired, separate light/dark values) — reuse those tokens for any similar syntax-highlighting need elsewhere rather than hardcoding new colors.

Plan: `PLANS/PLAN_JSON_TREE_VIEWER.md`. Docs: `docs/api|cli|testing/json-tree.md`.

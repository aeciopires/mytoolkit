---
name: json-tree
description: Implement or modify the JSON Tree Viewer tool (internal/tools/jsontree) — parsing raw JSON into a key-order-preserving navigable tree. Trigger on "implement JSON tree viewer", "add tree node parsing", "fix json-tree".
---

# JSON Tree Viewer

Business logic: `src/internal/tools/jsontree/jsontree.go`, `func Parse(input []byte, opts Options) (Node, error)`.

Key implementation rule: object keys must keep source order. Do **not** decode into `map[string]any` — Go maps are unordered. Instead stream `json.Token`s from a `json.Decoder` (see `parseValue` in `jsontree.go`) and use `dec.UseNumber()` so large integers aren't rounded through `float64`.

`Node{Key, Type, Value, Children}` — `Type` is one of `object|array|string|number|bool|null`.

Wiring: `src/internal/cli/jsontree.go` registers both the `json-tree` CLI subcommand and the `json-tree` REST handler (bespoke, not `handlers.Wrap`, since the response shape is `{tree: Node}` not `{output: string}`). Web page: `src/internal/web/templates/tools/json-tree.html` (custom recursive JS renderer, not the generic `tool-common.js` textarea flow).

Plan: `PLANS/PLAN_JSON_TREE_VIEWER.md`. Docs: `docs/api|cli|testing/json-tree.md`.

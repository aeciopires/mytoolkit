---
name: json-toon
description: Implement or modify the JSON to TOON Converter tool (internal/tools/jsontoon + internal/web/static/js/json-toon.js) — TOON-format encoding, with a client-side-only web page. Trigger on "implement JSON to TOON converter", "add TOON encoder", "fix json-toon".
---

# JSON to TOON Converter

Converts JSON into [TOON](https://github.com/toon-format/spec) (Token-Oriented Object Notation) — tabular arrays, minimal quoting, indentation instead of braces.

## Two implementations that must stay in sync

This is the only tool in the app with **two independent implementations of the same algorithm**:

- `app/internal/tools/jsontoon/jsontoon.go` — Go, backs the REST endpoint (`handlers.Wrap`) and the CLI subcommand (`newTextToolCommand`). Standard for this app.
- `app/internal/web/static/js/json-toon.js` — vanilla JS, backs **only** the web page. The web page's `data-client-side` attribute (see below) tells `tool-common.js` to skip its normal fetch-based wiring; the page's own inline `<script>` (in `json-toon.html`'s `{{define "extra-scripts"}}` block) calls `window.jsonToToon()` directly on every input/option change.

**Any change to the encoding rules (quoting, tabular detection, number formatting, delimiter/indent handling) must be made in both files**, then verified with the parity check in `docs/testing/json-toon.md` (loads `json-toon.js` in headless Chrome and diffs its output against the same fixture table used by `jsontoon_test.go`). There is no automated CI gate for this — it's a manual/scripted step you must run yourself; don't skip it.

## Why client-side for the web page

Product requirement: the web page must send zero network requests for its live conversion (matching the privacy claim of the reference site, `scalevise.com/json-toon-converter`) — verifiable via the browser's network tab or by grepping `json-toon.js` for `fetch`/`XMLHttpRequest`/`WebSocket` (must return nothing). REST and CLI are **not** part of this constraint — they are full, required, first-class Go implementations, identical in status to every other tool's REST/CLI support. Don't ever remove or stub them out under the assumption "it's client-side now."

## The `data-client-side` convention

`registry.Tool.ClientSide bool` → `internal/web/templates/partials/tool-panel.html` renders `data-client-side` on `.tool-panel` when true → `tool-common.js` checks `panel.hasAttribute('data-client-side')` and skips wiring its fetch-based `run()` (copy/reset/download button wiring is unaffected). Reuse this convention for any future tool with the same no-network-call requirement — don't reinvent it.

## Order-preserving decode

Object key order from the source JSON must be preserved (TOON is order-sensitive, unlike a generic JSON pretty-printer). The Go side re-implements the same `json.Decoder.Token()` + `UseNumber()` streaming technique already used by `internal/tools/jsontree` — **duplicated on purpose, not imported**, because `internal/tools/<name>` packages may only depend on `apperr` (see `CLAUDE.md`'s Conventions). The JS side relies on `JSON.parse()` preserving string-key insertion order, which is an engine guarantee in all evergreen browsers, not a language-spec requirement — documented as a relied-upon assumption, not re-verified per change.

## Scope

Only `delimiter` (comma/tab/pipe) and `indent_size` are exposed as options — the TOON spec's `keyFolding`/`flattenDepth` knobs are left at spec defaults, a known MVP limitation. Deeply nested/exotic array shapes fall back to a simplified list form rather than exhaustively implementing every corner of spec §9–§14; this is intentional, not a bug — see `PLAN_JSON_TOON_CONVERTER.md`'s TOON format primer for exactly what's implemented.

MCP: `json-toon` tool (`app/internal/mcp/json_toon.go`) — calls the same full Go `jsontoon.Convert`, independent of the web page's client-side-only mirror. Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_JSON_TOON_CONVERTER.md`. Shared architecture: `PLANS/PLAN_ARCHITECTURE.md`. Docs: `docs/api|cli|testing/json-toon.md`.

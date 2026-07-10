<!-- TOC -->

- [PLAN\_JSON\_TOON\_CONVERTER](#plan_json_toon_converter)
  - [Description](#description)
  - [TOON format primer](#toon-format-primer)
  - [Architecture deviation: client-side-only web conversion](#architecture-deviation-client-side-only-web-conversion)
  - [Business logic (Go)](#business-logic-go)
  - [Client-side logic (JavaScript)](#client-side-logic-javascript)
  - [CLI](#cli)
  - [REST](#rest)
  - [Web UI](#web-ui)
  - [Metrics](#metrics)
  - [Unit tests](#unit-tests)
  - [Documentation](#documentation)
  - [Skill](#skill)
  - [CHANGELOG.md](#changelogmd)
  - [CLAUDE.md](#claudemd)
  - [New dependencies](#new-dependencies)

<!-- TOC -->

# PLAN_JSON_TOON_CONVERTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document covers what is specific to the JSON to TOON Converter feature, **plus a new architectural pattern** — the first tool in this app whose web page performs its live conversion entirely client-side rather than calling the REST API. Tool slug: `json-toon`.

This feature is added after the initial 11-tool implementation, inspired by `https://scalevise.com/json-toon-converter`, at the user's explicit request. `PLAN_ARCHITECTURE.md` is updated alongside this document to register `json-toon` as a 12th tool and to document the new client-side-only pattern as a reusable convention (see `PLAN_ARCHITECTURE.md`'s "Additional instructions received during implementation" section).

## Description

Converts a JSON document into [TOON](https://github.com/toon-format/spec) (Token-Oriented Object Notation) — a compact, indentation-based, schema-aware text format designed to reduce the token count of structured data sent to LLMs, while staying human-readable and losslessly convertible back to JSON. Typical savings are 30–60% fewer tokens than compact JSON for arrays of uniform objects (tabular data), which is TOON's main use case.

**Available on all three surfaces, like every other tool in this app**: a full Go implementation backs the CLI subcommand (`mytoolkit json-toon`) and the REST endpoint (`POST /api/v1/tools/json-toon`) — see [Business logic (Go)](#business-logic-go), [CLI](#cli), and [REST](#rest) below. The *only* thing that's different about this tool is its **web page**, which additionally ships a client-side JavaScript mirror of the same algorithm so the live in-browser conversion never calls the server — see [Architecture deviation](#architecture-deviation-client-side-only-web-conversion). CLI and REST are not optional or secondary here; they are first-class, required deliverables of this feature, identical in status to every other tool's CLI/REST support.

## TOON format primer

(Reference: TOON spec v3.2, `toon-format/spec`, MIT licensed — implementation follows this spec, not the reference site's exact wording, to avoid copying UI copy verbatim.)

- **Objects**: indentation instead of braces. Primitive fields: `key: value` (one space after colon). Nested/empty objects: `key:` alone on its line, children at depth+1. Field (key) order is preserved from the source JSON — this requires an order-preserving JSON decode, not a plain `map[string]any` unmarshal (the same constraint already solved for the JSON Tree Viewer via token streaming; see `PLAN_JSON_TREE_VIEWER.md`).
- **Arrays of primitives**: inline with a declared length and delimiter: `key[N]: v1,v2,v3`.
- **Arrays of uniform objects (tabular form)** — TOON's headline feature: when every element is an object with the same keys and only primitive values, emit one header line declaring length and fields, then one row per element:
  ```
  users[2]{id,name,role}:
    1,Alice,admin
    2,Bob,user
  ```
- **Arrays that aren't uniform objects** (mixed types, nested arrays/objects, differing keys): fall back to a non-tabular list form, one element per line at depth+1 (e.g. `pairs[2]:` followed by `- [2]: 1,2` / `- [2]: 3,4` for nested arrays).
- **String quoting**: a string is quoted only when required — empty, leading/trailing whitespace, equals `true`/`false`/`null`, matches a numeric pattern, contains a colon/quote/backslash/bracket/brace/control character, contains the active delimiter, or starts with a hyphen. Unquoted otherwise (this is most of the token savings).
- **Numbers**: canonical decimal form — no unnecessary exponents in normal ranges, no leading zeros, no trailing fractional zeros, `-0` → `0`.
- **Booleans/null**: lowercase `true`/`false`/`null` only.
- **Root value**: object by default; a root array uses the same header form at depth 0; a bare root scalar (number/string/bool/null) is emitted as a single token.
- **Options this tool exposes** (subset of the spec's configurable knobs — delimiter and indent only; `keyFolding`/`flattenDepth` are left at spec defaults for MVP, documented as a known limitation):
  - `delimiter`: `comma` (default), `tab`, or `pipe` — selects the array/row delimiter and the header symbol per spec §9.1.
  - `indent_size`: spaces per depth level, default 2 (spec default; tabs are forbidden for indentation per spec).

## Architecture deviation: client-side-only web conversion

Every other tool's web page calls `POST /api/v1/tools/<slug>` via `fetch()` (see `PLAN_ARCHITECTURE.md`'s "Avoiding 3x duplicated logic"). This tool's page does **not** — matching the reference site's core promise ("100% client-side conversion, your data never leaves the browser, no uploads, no logs, no servers"), the JSON-to-TOON web page converts entirely in the browser via a small embedded JavaScript encoder, with no network call for the live conversion.

The REST endpoint and CLI subcommand still exist and still run a Go implementation of the same algorithm — required for API/CLI parity with every other tool and for programmatic/scripted use, where "runs in a browser" isn't the point. This means **two independent implementations of the same TOON encoding algorithm exist**: `internal/tools/jsontoon` (Go, used by REST + CLI) and `internal/web/static/js/json-toon.js` (vanilla JS, used only by the web page). This is a deliberate, explicit trade-off:

- **Risk**: the two implementations can drift apart over time (a bug fixed in one and not the other).
- **Mitigation**: both are tested against the same fixture table (see [Unit tests](#unit-tests)) — the Go side via `go test`, the JS side via a documented manual/scripted parity check — and `.skills/json-toon/SKILL.md` calls this out explicitly so future changes touch both.
- **Metrics caveat**: `mytoolkit_tool_usage_total{tool="json-toon"}` (see `PLAN_ARCHITECTURE.md`'s Metrics design) only reflects direct REST API calls — web UI conversions never hit the server, so they're invisible to server-side usage metrics. This mirrors the already-documented "CLI invocations aren't counted either" caveat; both are stated together in `docs/api/json-toon.md`.

**New shared convention** (added to `internal/web/static/js/tool-common.js`, documented in `PLAN_ARCHITECTURE.md`'s "Shared code and configuration reuse" for reuse by any future browser-only tool): a `.tool-panel` element carrying a `data-client-side` attribute causes `tool-common.js` to skip wiring its `fetch()`-based `run()` entirely (copy/reset/download button wiring still applies) — the page's own inline `<script>` owns input → output conversion. This is the same lightweight opt-in-attribute pattern already used for `data-autorun` (Password Generator).

## Business logic (Go)

Package: `internal/tools/jsontoon/jsontoon.go`.

```go
package jsontoon

type Options struct {
    Delimiter  string `json:"delimiter"`   // "comma" (default) | "tab" | "pipe"
    IndentSize int    `json:"indent_size"` // default 2
}

func Convert(input []byte, opts Options) (string, error)
```

Implementation:
1. Decode `input` into an order-preserving generic value using the same `json.Decoder` token-streaming technique as `internal/tools/jsontree` (`json.Decoder.Token()` + `UseNumber()`) — reused, not reimplemented; consider factoring the shared streaming-decode-to-ordered-value step into a small internal helper both packages call, evaluated during implementation against `PLAN_ARCHITECTURE.md`'s reuse principle (only worth extracting if it doesn't complicate either package's pure-function signature).
2. Walk the ordered value tree and emit TOON text per the rules above: detect "uniform array of flat objects" (all elements are objects, all have the same key set in the same order, all values are primitives) to choose tabular vs. list form; apply the quoting rules character-by-character; canonicalize numbers via `strconv`.
3. `apperr.OneOf("delimiter", ...)` validates `Delimiter` against `comma`/`tab`/`pipe`.

Edge cases:
- Empty input → `apperr.ErrEmptyInput`, HTTP 400 (same convention as JSON Formatter/YAML Formatter/JSON Tree Viewer).
- Malformed JSON → `INVALID_JSON`, HTTP 400, reusing the same code as `internal/tools/jsonformat`/`jsontree`.
- Invalid `delimiter` value → `INVALID_OPTION`, HTTP 400 (via `apperr.OneOf`).
- Non-uniform arrays → not an error; falls back to list form (documented, not a limitation).
- Deeply nested input → no artificial limit for MVP, same documented stack-growth caveat as JSON Tree Viewer.
- Root scalar (bare `42`, `"hi"`, `true`, `null` as the entire input) → emitted as a single bare token, not wrapped.

## Client-side logic (JavaScript)

`internal/web/static/js/json-toon.js` — a standalone, dependency-free encoder implementing the same rules as the Go package above (object/array walking, tabular-array detection, quoting rules, number canonicalization), operating on the result of `JSON.parse()` (which preserves string-key insertion order in all evergreen browsers this app targets — documented as a relied-upon JS engine guarantee, not a spec requirement of `JSON.parse` itself). Exposes one function, e.g. `window.jsonToToon(text, { delimiter, indentSize }) -> string`, called directly by `json-toon.html`'s inline script on input/option change — no `fetch()`, no network activity, verifiable by inspecting the browser's network tab (call this out in `docs/api/json-toon.md` as a testable claim, not just marketing copy).

## CLI

```
mytoolkit json-toon --in <file|-> [--out <file|->] [--delimiter comma|tab|pipe] [--indent N]
```

Fully generic — built with `newTextToolCommand` like most tools (see `PLAN_ARCHITECTURE.md`).

Example:
```
$ echo '{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}' | mytoolkit json-toon
users[2]{id,name,role}:
  1,Alice,admin
  2,Bob,user
```

## REST

`POST /api/v1/tools/json-toon`

Request:
```json
{ "input": "{\"id\":123,\"name\":\"Ada\",\"active\":true}", "options": { "delimiter": "comma", "indent_size": 2 } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "id: 123\nname: Ada\nactive: true\n" },
  "meta": { "tool": "json-toon", "duration_ms": 0.05 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "unexpected end of JSON input" } }
```

Fully generic — output is a plain string, so this endpoint uses `handlers.Wrap` like most tools (no bespoke handler needed, unlike JSON Tree Viewer/Text Counter/Password Generator/JWT/QR Code).

## Web UI

- Input `<textarea>` (raw JSON) + output `<textarea readonly>` (TOON), reusing the shared `tool-panel` partial's two-column layout.
- `data-client-side` attribute on `.tool-panel` (see [Architecture deviation](#architecture-deviation-client-side-only-web-conversion)) — the page's own inline script wires `input.addEventListener('input', ...)` directly to `window.jsonToToon()`, bypassing `tool-common.js`'s fetch-based `run()`.
- Options row: delimiter selector (Comma/Tab/Pipe), indent stepper (2/4).
- Copy/Reset buttons reuse the existing shared `[data-action]` wiring in `tool-common.js` (unaffected by `data-client-side`, which only skips the fetch call).
- A short, factual privacy note near the panel — e.g. "Runs entirely in your browser. This page never sends your JSON to the server." — phrased in this project's own words rather than copied from the reference site.
- Optional stretch goal, not required for MVP: a small client-side character-count comparison (JSON size vs. TOON size) as a visual token-savings indicator — purely cosmetic, computed from `input.value.length` vs `output.value.length`, no server round-trip either.

## Metrics

Shared `tool="json-toon"` label on REST-path metrics only; see the [Metrics caveat](#architecture-deviation-client-side-only-web-conversion) above — web-originated conversions are not counted. No custom metric.

## Unit tests

`internal/tools/jsontoon/jsontoon_test.go` (Go), table-driven, covering — and shared as a fixture table with the JS side for manual/scripted parity checking:
- Tabular array of uniform objects (the `users` example above).
- Simple flat object (the `id`/`name`/`active` example above).
- Array of arrays (non-uniform-object fallback to list form).
- Non-uniform array of objects (differing key sets) → list form, not tabular.
- String quoting: empty string, string with leading/trailing whitespace, string literally `"true"`/`"false"`/`"null"`, numeric-looking string (`"123"`), string containing a colon/comma/bracket, string starting with `-`.
- Number canonicalization: integers, floats, negative zero, large/small magnitudes.
- Delimiter option: `tab` and `pipe` produce the documented header/row syntax.
- Indent option: custom `indent_size`.
- Empty input → error. Malformed JSON → error. Invalid delimiter option → error.
- Root scalar input (bare `42`) → bare token output.

`docs/testing/json-toon.md` documents this table and additionally records the **JS parity check procedure**: since this project has no Node/npm/Jest tooling by design (see `PLAN_ARCHITECTURE.md`'s "no Node build step" decision), the JS encoder is verified by loading the real page in headless Chrome (the same tool already used for README screenshots) and asserting `window.jsonToToon()`'s output against the same fixture table used by the Go tests — a scripted manual check run whenever `json-toon.js` changes, not a CI-automated test suite. This limitation is stated plainly, not glossed over.

## Documentation

- `docs/api/json-toon.md` — REST reference, **two** Mermaid diagrams (not the usual one): one for the REST/CLI path through the Go implementation, one for the web page's entirely-client-side path, since they genuinely diverge (see [Architecture deviation](#architecture-deviation-client-side-only-web-conversion)).
- `docs/cli/json-toon.md` — CLI reference with the example above.
- `docs/testing/json-toon.md` — Go test reference + the JS parity check procedure described above.
- `README.md`: add to the Features list (12th tool) and the Documentation table, following the existing row format.

## Skill

`.skills/json-toon/SKILL.md` — triggers on "implement JSON to TOON converter", "add TOON encoder"; documents: the TOON spec subset implemented (tabular arrays, quoting rules, delimiter/indent options), the order-preserving-decode requirement (reuse `jsontree`'s streaming technique, don't reintroduce `map[string]any`), and — most importantly — the dual Go+JS implementation requirement: **any change to the encoding rules must be made in both `internal/tools/jsontoon/jsontoon.go` and `internal/web/static/js/json-toon.js`, verified against the same fixture table**. Links to this plan and to `PLAN_ARCHITECTURE.md`'s client-side-only pattern section.

## CHANGELOG.md

Add an `### Added` entry under `[Unreleased]` when this feature is implemented: "JSON to TOON Converter (`json-toon`) — the first tool with a 100%-client-side web conversion path (no data sent to the server for the web UI); REST/CLI still available via a Go implementation of the same TOON spec subset." Cross-reference the metrics-visibility caveat.

## CLAUDE.md

When implemented, update:
- The tool count (11 → 12) in the Project summary line.
- "Adding a new tool" step 4 (web template) to mention the `data-client-side` opt-out attribute as an option for tools that shouldn't call the REST API from their own web page.
- "Conventions" to add: pure `internal/tools/<name>` packages are the default, but a tool may additionally ship a client-side JS mirror under `internal/web/static/js/<name>.js` when a no-network-call guarantee is a product requirement — such tools must state this explicitly in their `.skills/<name>/SKILL.md` and keep both implementations tested against one shared fixture table.

## New dependencies

None — Go stdlib only (`encoding/json` token streaming, `strconv` for number canonicalization), and dependency-free vanilla JavaScript for the client-side mirror.

<!-- TOC -->

- [PLAN\_YAML\_FORMATTER](#plan_yaml_formatter)
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

# PLAN_YAML_FORMATTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the YAML Formatter feature. Tool slug: `yaml-format`.

## Description

Reformats a YAML document with consistent indentation.

**Revised after reading the YAML spec (yaml.org/spec/1.2.2)**: two gaps were found against the spec and fixed. (1) The spec defines `---`/`...` as document stream markers (§9.1) — a stream may hold multiple documents — but the original implementation silently discarded every document after the first. It now decodes and reformats every document in the stream. (2) The spec states block vs. flow collection style is "a presentation detail and is not reflected in the serialization tree or representation graph" (§10.1/§10.2), meaning normalizing style is always lossless — so a `style` option was added (`block`/`flow`) that forces one consistent style across the whole document, the YAML equivalent of JSON Formatter's pretty/minify. Previously the formatter only re-indented; it left flow-style input (`{a: 1}`) untouched, so "consistent indentation" wasn't actually true for mixed-style input.

## Business logic

Package: `internal/tools/yamlformat/yamlformat.go`.

```go
package yamlformat

type Options struct {
    Indent int    // default 2, block style only
    Style  string // "block" (default) | "flow"
}

func Format(input []byte, opts Options) (string, error)
```

Implementation: `yaml.NewDecoder` loops over `dec.Decode(&node)` until `io.EOF`, decoding each document into a `yaml.Node` (not `any`) so key order, comments, anchors/aliases, and explicit tags survive the round-trip. Each node's collection (mapping/sequence) style is forced recursively to the requested `Style` before being written to a single shared `yaml.Encoder`, which automatically re-inserts `---` between documents when `Encode` is called more than once. Scalar node styles (plain/quoted/block) are left untouched — quoting can carry meaning (e.g. `"yes"` vs. `yes`), so only collection style is normalized.

Edge cases:
- Empty input → `ErrEmptyInput`, HTTP 400.
- A stream containing only comments/blank lines decodes to zero documents without a parse error; treated the same as empty input rather than silently returning an empty string.
- Malformed YAML → HTTP 400, `INVALID_YAML`, message from the underlying parser (includes a line number).
- Tab characters used for indentation → rejected by the parser (YAML forbids tabs for indentation, spec §6.1); surfaces as `INVALID_YAML`.
- `Indent <= 0` → default to 2.
- Invalid `style` value (not `block`/`flow`) → `INVALID_OPTION`, HTTP 400, via the shared `apperr.OneOf` validator.
- Multi-document YAML (`---`/`...` separators) → **now fully supported**; every document is reformatted and re-joined with `---`.
- Comment preservation: preserved via `yaml.Node` head/line/foot comment fields — verified for head comments and same-line trailing comments. The spec explicitly defines no formal comment-attachment semantics ("comments... must not be used to convey content information", and their association with nodes is left to the implementation), so a comment sitting alone between two mapping keys can still shift which key it's considered attached to. This is a `yaml.v3` / spec-inherent limitation, not a bug to "fix."
- Anchors (`&name`), aliases (`*name`), and merge keys (`<<`) round-trip correctly; merge keys are re-emitted with an explicit `!!merge` tag (e.g. `!!merge <<: *x`) even if the source omitted it — semantically identical, and arguably more spec-precise, not a defect.
- Plain scalars that look like other types in YAML 1.1 (`yes`, `no`, `NO`, etc. — the classic "Norway problem") are preserved as their original plain text, not re-resolved or re-quoted, because the formatter re-serializes the already-parsed node's scalar value rather than re-interpreting its string form.

## CLI

```
mytoolkit yaml-format --in <file|-> [--out <file|->] [--indent N] [--style block|flow]
```

Example:
```
$ printf 'a: 1\nb:\n    - x\n    - y\n' | mytoolkit yaml-format --indent 2
a: 1
b:
  - x
  - y

$ printf 'a: 1\n---\nb: 2\n' | mytoolkit yaml-format
a: 1
---
b: 2

$ printf 'a:\n  b: 1\n  c:\n    - 1\n    - 2\n' | mytoolkit yaml-format --style flow
{a: {b: 1, c: [1, 2]}}
```

## REST

`POST /api/v1/tools/yaml-format`

Request:
```json
{ "input": "a: 1\nb:\n    - x\n    - y\n", "options": { "indent": 2, "style": "block" } }
```

`options.style`: `block` (default) or `flow`. `options.indent`: spaces per level, block style only, default 2.

Success (200):
```json
{
  "success": true,
  "data": { "output": "a: 1\nb:\n  - x\n  - y\n" },
  "meta": { "tool": "yaml-format", "duration_ms": 0.27 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_YAML", "message": "yaml: line 2: did not find expected key" } }
```

Error codes: `EMPTY_INPUT`, `INVALID_YAML`, `INVALID_OPTION` (bad `style`).

## Web UI

- Input `<textarea>` + output `<textarea readonly>`, a Style select (Block/Flow) and indent select (2/4), Copy/Reset from the shared `tool-panel` partial.
- Live formatting on input change via debounced `fetch()`.
- A one-line note discloses that multi-document streams are fully supported and what "Flow" does.

## Metrics

Shared `tool="yaml-format"` label; no custom metric.

## Unit tests

`internal/tools/yamlformat/yamlformat_test.go`:
- Basic mapping reformatted with a different indent.
- Nested list under a key.
- Empty input → error.
- Whitespace/comment-only stream (decodes to zero documents) → error.
- Malformed YAML (bad indentation, unclosed flow sequence) → error.
- Tab character used for indentation → error.
- Invalid `style` option value → error.
- Multi-document stream (`---`-separated) → every document reformatted, separators preserved.
- `style: flow` forces compact single-line output; `style: block` normalizes mixed flow input back to indented block style.
- Comments preserved (head comment + same-line trailing comment).
- Anchors and aliases preserved.
- Idempotency: formatting already-formatted output twice yields the same result.

## Documentation

- `docs/api/yaml-format.md`, `docs/cli/yaml-format.md`, `docs/testing/yaml-format.md`.

## Skill

`.skills/yaml-format/SKILL.md` — triggers on "implement YAML formatter"; documents the `yaml.Node` round-trip approach and its limitations, links to this plan.

## New dependencies

`gopkg.in/yaml.v3` (added to `go.mod`).

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

## Business logic

Package: `internal/tools/yamlformat/yamlformat.go`.

```go
package yamlformat

type Options struct {
    Indent int // default 2
}

func Format(input []byte, opts Options) (string, error)
```

Implementation: decode with `yaml.v3`'s `yaml.Node` (not directly into `any`) so document structure and comments are preserved where the library supports it, then re-encode with `yaml.Encoder.SetIndent(opts.Indent)`. Decoding into `yaml.Node` instead of `map[string]any` avoids losing key order (maps are unordered) and keeps scalar style hints.

Edge cases:
- Empty input → `ErrEmptyInput`, HTTP 400.
- Malformed YAML → HTTP 400, `INVALID_YAML`, message from the underlying parser.
- `Indent <= 0` → default to 2.
- Multi-document YAML (`---` separators) → out of scope for MVP; only the first document is processed, documented as a known limitation.
- Comment preservation: best-effort via `yaml.Node`; document that some comment placements may shift — this is a known limitation of `yaml.v3`, not a bug to "fix."

## CLI

```
mytoolkit yaml-format --in <file|-> [--out <file|->] [--indent N]
```

Example:
```
$ printf 'a: 1\nb:\n    - x\n    - y\n' | mytoolkit yaml-format --indent 2
a: 1
b:
  - x
  - y
```

## REST

`POST /api/v1/tools/yaml-format`

Request:
```json
{ "input": "a: 1\nb:\n    - x\n    - y\n", "options": { "indent": 2 } }
```

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

## Web UI

- Input `<textarea>` + output `<textarea readonly>`, indent stepper (2/4), "Copy", "Download .yaml", "Reset".
- Live formatting on input change via debounced `fetch()`.
- Note in the UI (tooltip or small text) that only the first YAML document is processed if multiple are pasted.

## Metrics

Shared `tool="yaml-format"` label; no custom metric.

## Unit tests

`internal/tools/yamlformat/yamlformat_test.go`:
- Basic mapping reformatted with a different indent.
- Nested list under a key.
- Empty input → error.
- Malformed YAML (bad indentation, tab character) → error.
- Idempotency: formatting already-formatted output twice yields the same result.

## Documentation

- `docs/api/yaml-format.md`, `docs/cli/yaml-format.md`, `docs/testing/yaml-format.md`.

## Skill

`.skills/yaml-format/SKILL.md` — triggers on "implement YAML formatter"; documents the `yaml.Node` round-trip approach and its limitations, links to this plan.

## New dependencies

`gopkg.in/yaml.v3` (added to `go.mod`).

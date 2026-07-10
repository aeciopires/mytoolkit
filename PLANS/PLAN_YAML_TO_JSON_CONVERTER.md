<!-- TOC -->

- [PLAN\_YAML\_TO\_JSON\_CONVERTER](#plan_yaml_to_json_converter)
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

# PLAN_YAML_TO_JSON_CONVERTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document covers what is specific to the YAML to JSON Converter feature. Tool slug: `yaml-to-json`.

Added at the user's explicit request, alongside [PLAN_JSON_TO_YAML_CONVERTER.md](PLAN_JSON_TO_YAML_CONVERTER.md), to use [`sigs.k8s.io/yaml`](https://github.com/kubernetes-sigs/yaml) — the library Kubernetes itself uses to accept YAML manifests as JSON internally. This differs deliberately from `internal/tools/yamlformat`'s choice of `gopkg.in/yaml.v3`: the Formatter needs `yaml.v3`'s `yaml.Node` API to preserve comments/anchors/key order for a same-format round trip, while a cross-format converter has no such round-trip to preserve — `sigs.k8s.io/yaml` converts by decoding YAML into a generic value and re-encoding it as JSON (or vice versa), which is exactly the right shape here and is battle-tested at Kubernetes' scale.

## Description

Converts a single YAML document into pretty-printed JSON.

## Business logic

Package: `internal/tools/yamltojson/yamltojson.go`.

```go
package yamltojson

type Options struct {
    Indent int `json:"indent"` // default 2
}

func Convert(input []byte, opts Options) (string, error)
```

Implementation: `sigs.k8s.io/yaml.YAMLToJSONStrict` (not the plain `YAMLToJSON`) converts the input to compact JSON, then `encoding/json.Indent` pretty-prints it. `YAMLToJSONStrict` is chosen over `YAMLToJSON` specifically because the YAML spec forbids duplicate mapping keys, but the plain converter's documented behavior is to silently keep one of them "in an undefined order" — strict mode turns that into a reported error instead of silent data loss.

Edge cases (verified against the real library, not assumed):
- Empty input → `ErrEmptyInput`, HTTP 400. (The library itself does **not** treat empty YAML as an error — `YAMLToJSONStrict("")` returns `"null"` with no error — so this check must happen before calling into the library, not be delegated to it.)
- Malformed YAML (unclosed flow sequence, tab-character indentation) → HTTP 400, `INVALID_YAML`, with the library's line-numbered message.
- Duplicate mapping keys → HTTP 400, `INVALID_YAML` (via strict mode, see above).
- `Indent <= 0` → default to 2.
- Multi-document YAML (`---` separators) → only the first document is converted; documented as a known limitation (unlike `yaml-format`, which was revised to support every document in a stream — deliberately not replicated here, since "convert a multi-document YAML stream into one JSON array" is a different, unrequested product decision, not the same fix).
- **YAML 1.1 boolean/null resolution ("Norway problem")**: `sigs.k8s.io/yaml`'s own doc comment states plainly that "literal 'yes' and 'no' strings without quotation marks will be converted to true/false implicitly." Verified directly: unquoted `NO`/`y` in source YAML become JSON `false`/`true`, not the strings `"NO"`/`"y"`. This is documented library behavior, not a bug — callers who want a literal string must quote it in the source YAML (`"NO"`).
- Large integers (up to 64 bits) are preserved exactly, per the library's own documented guarantee — verified with an 18-digit integer.

## CLI

```
mytoolkit yaml-to-json --in <file|-> [--out <file|->] [--indent N]
```

Example:
```
$ printf 'a: 1\nb:\n  - x\n  - y\n' | mytoolkit yaml-to-json
{
  "a": 1,
  "b": [
    "x",
    true
  ]
}
```

(Note `"y"` converting to `true` — the YAML 1.1 boolean resolution described above, not a bug in this tool.)

Errors report the underlying parser's line number:
```
$ printf 'a: [1, 2\n' | mytoolkit yaml-to-json
Error: yaml: line 1: did not find expected ',' or ']'
```

## REST

`POST /api/v1/tools/yaml-to-json`

Request:
```json
{ "input": "a: 1\nb:\n  - x\n  - y\n", "options": { "indent": 2 } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "{\n  \"a\": 1,\n  \"b\": [\n    \"x\",\n    true\n  ]\n}" },
  "meta": { "tool": "yaml-to-json", "duration_ms": 0.1 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_YAML", "message": "yaml: line 1: did not find expected ',' or ']'" } }
```

Error codes: `EMPTY_INPUT`, `INVALID_YAML`.

## Web UI

- Input `<textarea>` (YAML) + output `<textarea readonly>` (JSON), indent select (2/4), Copy/Reset from the shared `tool-panel` partial. Fully generic — live formatting on input change via the shared debounced `fetch()`, no bespoke wiring needed.
- A one-line note discloses: duplicate keys are rejected, only the first document of a multi-document stream is converted, and the YAML 1.1 boolean-resolution gotcha for unquoted `y`/`n`/`yes`/`no`/`on`/`off`.

## Metrics

Shared `tool="yaml-to-json"` label; no custom metric.

## Unit tests

`internal/tools/yamltojson/yamltojson_test.go`:
- Flat mapping, nested mapping/sequence.
- Custom indent.
- Empty input → error.
- Malformed YAML → error.
- Tab-character indentation → error.
- Duplicate keys → error.
- YAML 1.1 boolean resolution (`NO`→`false`, `y`→`true`) locked in by a dedicated test, so a future dependency bump that changes this behavior is caught rather than silently shipped.
- Large integer (18 digits) preserved exactly.

## Documentation

- `docs/api/yaml-to-json.md`, `docs/cli/yaml-to-json.md`, `docs/testing/yaml-to-json.md`.
- `README.md`: add to the Features list and Documentation table.

## Skill

`.skills/yaml-to-json/SKILL.md` — triggers on "implement YAML to JSON converter"; documents the `YAMLToJSONStrict` choice and the YAML 1.1 boolean-resolution gotcha, links to this plan.

## New dependencies

`sigs.k8s.io/yaml` (added to `go.mod`), shared with [PLAN_JSON_TO_YAML_CONVERTER.md](PLAN_JSON_TO_YAML_CONVERTER.md).

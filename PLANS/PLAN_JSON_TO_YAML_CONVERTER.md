<!-- TOC -->

- [PLAN\_JSON\_TO\_YAML\_CONVERTER](#plan_json_to_yaml_converter)
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

# PLAN_JSON_TO_YAML_CONVERTER

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document covers what is specific to the JSON to YAML Converter feature. Tool slug: `json-to-yaml`.

Added at the user's explicit request, alongside [PLAN_YAML_TO_JSON_CONVERTER.md](PLAN_YAML_TO_JSON_CONVERTER.md), using [`sigs.k8s.io/yaml`](https://github.com/kubernetes-sigs/yaml) — see that plan's intro for why this library was chosen over `gopkg.in/yaml.v3` (already used by `yaml-format`) for a cross-format converter specifically.

## Description

Converts a JSON document into YAML.

## Business logic

Package: `internal/tools/jsontoyaml/jsontoyaml.go`.

```go
package jsontoyaml

type Options struct{}

func Convert(input []byte, opts Options) (string, error)
```

Implementation: `sigs.k8s.io/yaml.JSONToYAML`. **Important, verified-by-testing detail**: `JSONToYAML` parses its input with a YAML decoder internally (YAML is a superset of JSON), so on its own it does *not* reject invalid JSON — trailing commas, unquoted keys, single-quoted strings, and even a bare comment line were all confirmed to "succeed" (the comment case produces nonsense output instead of an error). A "JSON to YAML converter" rejecting non-JSON input is a correctness requirement, not a nice-to-have, so `Convert` first validates the input with `encoding/json.Unmarshal` (which does enforce the JSON grammar) and only calls into `sigs.k8s.io/yaml` once that succeeds.

Edge cases (verified against the real library, not assumed):
- Empty input → `ErrEmptyInput`, HTTP 400.
- Invalid JSON (trailing comma, unquoted key, single-quoted string, unbalanced braces) → HTTP 400, `INVALID_JSON`, with `encoding/json`'s own error message (line/offset where available).
- Duplicate JSON keys → last value wins, silently — same as `encoding/json` itself; JSON (unlike YAML) does not forbid duplicate keys, so this isn't a bug to flag.
- Large integers (up to 64 bits) are preserved exactly, per the library's documented guarantee — verified with an 18-digit integer.
- **Values that would resolve as another type if left unquoted in YAML 1.1** (e.g. the JSON string `"NO"` or `"y"`) are automatically double-quoted in the YAML output by the library — verified directly (`"country":"NO"` → `country: "NO"`, not bare `NO`). This is the safe, correct direction of the YAML 1.1 boolean-resolution gotcha documented in `PLAN_YAML_TO_JSON_CONVERTER.md`: converting *to* YAML, the library actively protects against it; converting *from* YAML, there's nothing to protect against since the ambiguity already exists in the source.

## CLI

```
mytoolkit json-to-yaml --in <file|-> [--out <file|->]
```

Example:
```
$ printf '{"a":1,"b":["x","y"]}' | mytoolkit json-to-yaml
a: 1
b:
- x
- "y"
```

(Note `"y"` stays quoted in the output — see the "values that would resolve as another type" edge case above.)

Errors report the underlying `encoding/json` message:
```
$ printf '{"a":1,}' | mytoolkit json-to-yaml
Error: invalid character '}' looking for beginning of object key string
```

## REST

`POST /api/v1/tools/json-to-yaml`

Request:
```json
{ "input": "{\"a\":1,\"b\":[\"x\",\"y\"]}" }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "a: 1\nb:\n- x\n- \"y\"\n" },
  "meta": { "tool": "json-to-yaml", "duration_ms": 0.07 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_JSON", "message": "invalid character '}' looking for beginning of object key string" } }
```

Error codes: `EMPTY_INPUT`, `INVALID_JSON`.

## Web UI

- Input `<textarea>` (JSON) + output `<textarea readonly>` (YAML), Copy/Reset from the shared `tool-panel` partial. No custom options — fully generic, live formatting on input change via the shared debounced `fetch()`, no bespoke wiring or `tool-options` block needed.

## Metrics

Shared `tool="json-to-yaml"` label; no custom metric.

## Unit tests

`internal/tools/jsontoyaml/jsontoyaml_test.go`:
- Flat object, nested object/array.
- Empty input → error.
- Invalid JSON: trailing comma, unquoted key, single-quoted string → error (each verified to actually fail; the naive implementation without the `encoding/json` pre-validation step let all three through).
- Values that would otherwise resolve as booleans in YAML 1.1 (`"NO"`, `"y"`) come out quoted in the YAML output.
- Large integer (18 digits) preserved exactly.

## Documentation

- `docs/api/json-to-yaml.md`, `docs/cli/json-to-yaml.md`, `docs/testing/json-to-yaml.md`.
- `README.md`: add to the Features list and Documentation table.

## Skill

`.skills/json-to-yaml/SKILL.md` — triggers on "implement JSON to YAML converter"; documents the `encoding/json` pre-validation requirement (and why it's necessary — `JSONToYAML` alone is too lenient), links to this plan.

## New dependencies

`sigs.k8s.io/yaml` (added to `go.mod`), shared with [PLAN_YAML_TO_JSON_CONVERTER.md](PLAN_YAML_TO_JSON_CONVERTER.md).

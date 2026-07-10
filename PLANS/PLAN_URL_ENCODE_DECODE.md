<!-- TOC -->

- [PLAN\_URL\_ENCODE\_DECODE](#plan_url_encode_decode)
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

# PLAN_URL_ENCODE_DECODE

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the URL Encode/Decode feature. Tool slug: `url-encode`.

## Description

Encodes and decodes text according to URL percent-encoding standards.

## Business logic

Package: `internal/tools/urlencode/urlencode.go`.

```go
package urlencode

type Mode string // "encode" | "decode"
type Component string // "query" | "path" | "full" - which escaping rules to apply

type Options struct {
    Mode      Mode
    Component Component // default "query"
}

func Process(input string, opts Options) (string, error)
```

Implementation uses `net/url`:
- `Component == "query"` → `url.QueryEscape` / `url.QueryUnescape` (spaces become `+`, matches `application/x-www-form-urlencoded` semantics, the most common "URL encode" expectation).
- `Component == "path"` → `url.PathEscape` / `url.PathUnescape` (spaces become `%20`).
- `Component == "full"` → parses/re-encodes a full URL via `url.Parse`/`url.String()` for encode, and simply returns the input decoded component-wise for decode — documented as best-effort since "decoding a full URL" is not a single well-defined stdlib operation.

Edge cases:
- Empty input → returns empty output, not an error (encoding/decoding empty text is well-defined and harmless).
- Malformed percent-encoding on decode (e.g. `%ZZ`) → `INVALID_ENCODING`, HTTP 400, with the underlying `net/url` error message.

## CLI

```
mytoolkit url-encode [--decode] [--component query|path|full] --in <file|-> [--out <file|->]
```

Default mode is encode; `--decode` switches to decode. `--component` defaults to `query`.

Example:
```
$ echo 'hello world & friends' | mytoolkit url-encode
hello+world+%26+friends

$ echo 'hello+world+%26+friends' | mytoolkit url-encode --decode
hello world & friends
```

## REST

`POST /api/v1/tools/url-encode`

Request:
```json
{ "input": "hello world & friends", "options": { "mode": "encode", "component": "query" } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "hello+world+%26+friends" },
  "meta": { "tool": "url-encode", "duration_ms": 0.03 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_ENCODING", "message": "invalid URL escape \"%ZZ\"" } }
```

## Web UI

- Input `<textarea>` + output `<textarea readonly>`.
- Encode/Decode toggle, Component selector (Query/Path/Full), "Copy", "Swap input/output", "Reset".
- Live processing on input change via debounced `fetch()`.

## Metrics

Shared `tool="url-encode"` label; no custom metric.

## Unit tests

`internal/tools/urlencode/urlencode_test.go`:
- Encode with spaces and special characters (`&`, `=`, `#`) for each `Component` value.
- Decode round-trips encode output back to the original for each `Component` value.
- Empty input → empty output, no error.
- Malformed percent-encoding on decode → error.
- Unicode input encode/decode round-trip.

## Documentation

- `docs/api/url-encode.md`, `docs/cli/url-encode.md`, `docs/testing/url-encode.md`.

## Skill

`.skills/url-encode/SKILL.md` — triggers on "implement URL encode/decode"; documents the query vs. path vs. full component distinction, links to this plan.

## New dependencies

None — `net/url` from the standard library.

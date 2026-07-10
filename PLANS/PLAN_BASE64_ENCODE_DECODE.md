<!-- TOC -->

- [PLAN\_BASE64\_ENCODE\_DECODE](#plan_base64_encode_decode)
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

# PLAN_BASE64_ENCODE_DECODE

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the Base64 Encode/Decode feature. Tool slug: `base64`.

## Description

Encodes and decodes data using Base64.

## Business logic

Package: `internal/tools/base64enc/base64enc.go`.

```go
package base64enc

type Mode string // "encode" | "decode"
type Variant string // "standard" | "url" - RFC 4648 alphabet variant

type Options struct {
    Mode    Mode
    Variant Variant // default "standard"
    Padding bool    // default true; false uses RawStdEncoding/RawURLEncoding
}

func Process(input []byte, opts Options) (string, error)
```

Implementation uses `encoding/base64`, selecting `StdEncoding`/`URLEncoding`/`RawStdEncoding`/`RawURLEncoding` based on `Variant` and `Padding`.

Edge cases:
- Empty input → returns empty output, not an error.
- Malformed base64 on decode (invalid characters, incorrect padding) → `INVALID_BASE64`, HTTP 400, with the underlying `encoding/base64.CorruptInputError` position included in the message.
- Decoded output containing non-UTF-8 bytes → returned as-is in the JSON response using base64 re-encoding for transport safety is unnecessary here since the "output" of a decode is typically expected to be displayed as text; document that non-UTF-8 decoded output is shown using the Go `string()` conversion (which is lossy for invalid UTF-8) and note this as a known display limitation in the web UI (CLI writes raw bytes to `--out` without this limitation).

## CLI

```
mytoolkit base64 [--decode] [--variant standard|url] [--no-padding] --in <file|-> [--out <file|->]
```

Default mode is encode; `--decode` switches to decode. When `--out` is a file, `--decode` writes raw decoded bytes (safe for binary data); when writing to stdout, decoded bytes are written as-is (caller's responsibility to redirect to a file if binary).

Example:
```
$ echo -n 'hello world' | mytoolkit base64
aGVsbG8gd29ybGQ=

$ echo -n 'aGVsbG8gd29ybGQ=' | mytoolkit base64 --decode
hello world
```

## REST

`POST /api/v1/tools/base64`

Request:
```json
{ "input": "hello world", "options": { "mode": "encode", "variant": "standard", "padding": true } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "aGVsbG8gd29ybGQ=" },
  "meta": { "tool": "base64", "duration_ms": 0.02 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_BASE64", "message": "illegal base64 data at input byte 4" } }
```

## Web UI

- Input `<textarea>` + output `<textarea readonly>`.
- Encode/Decode switch, Variant selector (Standard/URL-safe), "Include padding" switch, "Copy", "Reset". (Both are single standalone on/off settings, styled as M3 switches, not checkboxes — see `PLAN_ARCHITECTURE.md`'s Theming section.)
- Live processing on input change via debounced `fetch()`.
- Note near the decode output that non-UTF-8 decoded content may render as replacement characters in the browser (matches the documented business-logic limitation above).

## Metrics

Shared `tool="base64"` label; no custom metric.

## Unit tests

`internal/tools/base64enc/base64enc_test.go`:
- Encode/decode round-trip for standard variant with and without padding.
- Encode/decode round-trip for URL-safe variant with and without padding.
- Empty input → empty output, no error.
- Malformed base64 on decode → error, with the byte position included.
- Binary (non-UTF-8) input round-trips correctly through encode→decode at the byte level (CLI file-based test).

## Documentation

- `docs/api/base64.md`, `docs/cli/base64.md`, `docs/testing/base64.md`.

## Skill

`.skills/base64/SKILL.md` — triggers on "implement Base64 encode/decode"; documents the standard vs. URL-safe variant and padding options, links to this plan.

## New dependencies

None — `encoding/base64` from the standard library.

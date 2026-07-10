<!-- TOC -->

- [PLAN\_QR\_CODE\_GENERATOR](#plan_qr_code_generator)
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

# PLAN_QR_CODE_GENERATOR

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the QR Code Generator feature. Tool slug: `qrcode`.

## Description

Generates a downloadable QR code image from text, a URL, or unicode text, inspired by `10015.io/tools/qr-code-generator`.

## Business logic

Package: `internal/tools/qrcode/qrcode.go`.

```go
package qrcode

type Options struct {
    Size int // pixels, square, default 256
}

func Generate(text string, opts Options) ([]byte, error) // returns PNG-encoded bytes
```

Implementation via `github.com/skip2/go-qrcode` (verify still maintained at implementation time; fallback `github.com/yeqown/go-qrcode`). Input is UTF-8 text (unicode supported natively by the encoding, satisfying the README's "Generate unicode text too").

Edge cases:
- Empty text → `EMPTY_INPUT`, HTTP 400.
- Text exceeding the QR spec's practical byte capacity for the library's chosen error-correction level (~2900 bytes for alphanumeric/byte mode at the lowest EC level) → `INPUT_TOO_LARGE`, HTTP 400, with the limit stated in the error message.
- `Size <= 0` or absurdly large (e.g. > 2048) → clamp to default/max rather than erroring, documented as a defensive default.

## CLI

```
mytoolkit qrcode --text <string> [--size N] --out <file>
```

`--out` is required in CLI mode (binary output must go to a file, not stdout, to avoid corrupting terminal output) — `--help` documents this explicitly and errors clearly if `--out` is omitted or `-`.

Example:
```
$ mytoolkit qrcode --text "https://example.com" --size 256 --out qr.png
Wrote 256x256 PNG to qr.png
```

## REST

`POST /api/v1/tools/qrcode`

Request:
```json
{ "input": "https://example.com", "options": { "size": 256 } }
```

Success (200): raw `image/png` bytes, `Content-Type: image/png`, `Content-Disposition: inline; filename="qrcode.png"`. This is the one deliberate deviation from the shared JSON envelope (documented in `PLAN_ARCHITECTURE.md`'s REST design section), since a JSON-wrapped base64 payload would prevent direct `<img src="/api/v1/tools/qrcode?...">` embedding and complicate the download flow.

Error (400), still JSON since there's no image to return:
```json
{ "success": false, "error": { "code": "EMPTY_INPUT", "message": "text must not be empty" } }
```

## Web UI

- Single text input labeled "URL or text".
- QR code image regenerates live on input change (debounced `fetch()`, image `src` set to an object URL built from the response blob).
- "Download QR Code" button (triggers a browser download of the current image).
- "Reset" button.

## Metrics

Shared `tool="qrcode"` label; no custom metric.

## Unit tests

`internal/tools/qrcode/qrcode_test.go`:
- Valid short text produces non-empty PNG bytes starting with the PNG magic header (`\x89PNG`).
- Unicode text produces valid output.
- Empty text → error.
- Text over the capacity limit → error.
- Different `Size` values produce different output byte lengths (sanity check, not exact pixel assertion).

## Documentation

- `docs/api/qrcode.md` (note the binary response type prominently), `docs/cli/qrcode.md`, `docs/testing/qrcode.md`.

## Skill

`.skills/qrcode/SKILL.md` — triggers on "implement QR code generator"; documents the binary-response exception to the JSON envelope and the CLI's mandatory `--out`, links to this plan.

## New dependencies

`github.com/skip2/go-qrcode` (or the verified maintained alternative) — added to `go.mod`.

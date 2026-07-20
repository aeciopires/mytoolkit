---
name: url-encode
description: Implement or modify the URL Encode/Decode tool (internal/tools/urlencode). Trigger on "implement URL encode/decode".
---

# URL Encode/Decode

`app/internal/tools/urlencode/urlencode.go`, `func Process(input []byte, opts Options) (string, error)`. Uses stdlib `net/url` only: `query` component → `QueryEscape`/`QueryUnescape` (spaces become `+`), `path` component → `PathEscape`/`PathUnescape` (spaces become `%20`), `full` → best-effort via `url.Parse`.

Fully generic wiring via `handlers.Wrap` / `newTextToolCommand`.

MCP: `url-encode` tool (`app/internal/mcp/url_encode.go`). Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_URL_ENCODE_DECODE.md`. Docs: `docs/api|cli|testing/url-encode.md`.

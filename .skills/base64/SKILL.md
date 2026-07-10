---
name: base64
description: Implement or modify the Base64 Encode/Decode tool (internal/tools/base64enc). Trigger on "implement Base64 encode/decode".
---

# Base64 Encode/Decode

Package is `base64enc` (not `base64`, to avoid colliding with stdlib `encoding/base64`). `app/internal/tools/base64enc/base64enc.go`, `func Process(input []byte, opts Options) (string, error)`.

`Options.Padding` is `*bool` (pointer), not `bool` — needed to distinguish "omitted" (defaults to `true`) from an explicit `false`, matching JSON semantics. `Variant` selects among `base64.StdEncoding`/`URLEncoding`/`RawStdEncoding`/`RawURLEncoding`.

Fully generic wiring via `handlers.Wrap` / `newTextToolCommand`.

Plan: `PLANS/PLAN_BASE64_ENCODE_DECODE.md`. Docs: `docs/api|cli|testing/base64.md`.

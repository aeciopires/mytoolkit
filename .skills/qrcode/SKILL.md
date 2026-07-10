---
name: qrcode
description: Implement or modify the QR Code Generator tool (internal/tools/qrcode) — text/URL to PNG, binary REST response. Trigger on "implement QR code generator".
---

# QR Code Generator

`src/internal/tools/qrcode/qrcode.go`, `func Generate(text string, opts Options) ([]byte, error)` — returns PNG bytes via `github.com/skip2/go-qrcode`.

REST response is raw `image/png`, not the shared JSON envelope — the one deliberate exception, so `<img src="/api/v1/tools/qrcode">` and downloads work directly. CLI requires `--out <file>` (binary can't go to stdout) — see `src/internal/cli/qrcode.go`. Web page uses a hidden `<img class="tool-image-output">`, populated by `tool-common.js`'s image-response branch.

Plan: `PLANS/PLAN_QR_CODE_GENERATOR.md`. Docs: `docs/api|cli|testing/qrcode.md`.

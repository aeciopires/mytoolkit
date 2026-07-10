---
name: jwt
description: Implement or modify the JWT Encode/Decode tool (internal/tools/jwttool) — inspection-mode decode plus HMAC-only encode. Trigger on "implement JWT encode/decode".
---

# JWT Encode/Decode

Package is `jwttool`, not `jwt` — avoids colliding with the `github.com/golang-jwt/jwt/v5` import name. `src/internal/tools/jwttool/jwttool.go`.

`Decode(token, secret)` always parses unverified first (via `jwt.NewParser().ParseUnverified`) so header/claims/signature are inspectable without a secret; if `secret != ""` it additionally attempts verification and sets `Valid`. `Encode` only supports HMAC algorithms (HS256/HS384/HS512) — RSA/ECDSA would need key-pair management out of scope for this tool.

Bespoke REST/CLI wiring (`src/internal/cli/jwt.go`), not `handlers.Wrap`, since decode/encode have different option shapes and decode's response isn't a single `output` string.

Plan: `PLANS/PLAN_JWT_ENCODE_DECODE.md`. Docs: `docs/api|cli|testing/jwt.md`.

<!-- TOC -->

- [PLAN\_JWT\_ENCODE\_DECODE](#plan_jwt_encode_decode)
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

# PLAN_JWT_ENCODE_DECODE

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the JWT Encode/Decode feature. Tool slug: `jwt`.

## Description

Decodes a JWT to inspect its header and claims (without requiring the signing secret), and encodes a new signed JWT from user-supplied claims and a secret, for inspection and testing purposes.

## Business logic

Package: `internal/tools/jwt/jwt.go`.

```go
package jwt

type DecodeResult struct {
    Header    map[string]any `json:"header"`
    Claims    map[string]any `json:"claims"`
    Signature string         `json:"signature"` // base64url, not validated
    Valid     *bool          `json:"valid,omitempty"` // set only if a secret was supplied for verification
}

func Decode(token string, secret string) (DecodeResult, error) // secret optional; empty = inspect only, no verification
func Encode(claims map[string]any, secret string, algorithm string) (string, error) // algorithm e.g. "HS256"
```

Implementation via `golang-jwt/jwt/v5`:
- `Decode` with an empty `secret` parses the token without verifying the signature (`jwt.ParseUnverified` or a `Keyfunc` that skips verification), purely for inspection — this mirrors the tool's purpose ("inspection and testing"), not production token validation.
- `Decode` with a non-empty `secret` additionally attempts verification and sets `Valid`.
- `Encode` only supports HMAC algorithms (`HS256`, `HS384`, `HS512`) for the MVP, since RSA/ECDSA would require key-pair management in the UI/CLI that is out of scope; documented as a known limitation.

Edge cases:
- Malformed token (not 3 dot-separated base64url segments) → `INVALID_TOKEN`, HTTP 400.
- Unsupported algorithm on encode → `UNSUPPORTED_ALGORITHM`, HTTP 400.
- Empty claims on encode → `EMPTY_CLAIMS`, HTTP 400.
- Header/claims segments that aren't valid JSON → `INVALID_TOKEN`, HTTP 400.

## CLI

```
mytoolkit jwt --decode --token <token> [--secret <secret>] [--out <file|->]
mytoolkit jwt --encode --claims <file|-> --secret <secret> [--algorithm HS256] [--out <file|->]
```

`--decode` and `--encode` are mutually exclusive; `--help` documents both modes with one example each.

Examples:
```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.abc123
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"abc123"}

$ echo '{"sub":"123","exp":1735689600}' | mytoolkit jwt --encode --secret mysecret
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJleHAiOjE3MzU2ODk2MDB9.xyz789
```

## REST

`POST /api/v1/tools/jwt`

Decode request:
```json
{ "input": "eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.abc123", "options": { "mode": "decode", "secret": "" } }
```

Decode success (200):
```json
{
  "success": true,
  "data": { "header": {"alg":"HS256","typ":"JWT"}, "claims": {"sub":"123"}, "signature": "abc123" },
  "meta": { "tool": "jwt", "duration_ms": 0.09 }
}
```

Encode request:
```json
{ "input": "{\"sub\":\"123\"}", "options": { "mode": "encode", "secret": "mysecret", "algorithm": "HS256" } }
```

Encode success (200):
```json
{
  "success": true,
  "data": { "output": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.xyz789" },
  "meta": { "tool": "jwt", "duration_ms": 0.12 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_TOKEN", "message": "token contains an invalid number of segments" } }
```

## Web UI

- Two tabs: "Decode" and "Encode".
- Decode tab: token input textarea, optional secret field, output shows header/claims/signature in three labeled read-only panels, with a color-coded validity badge when a secret is supplied.
- Encode tab: claims JSON textarea, secret field, algorithm dropdown (HS256/HS384/HS512), output token in a readonly field with "Copy".
- Explicit note near the secret field that secrets are never persisted server-side and are only used in-memory for the single request (privacy/security note relevant given this is a "testing" tool).

## Metrics

Shared `tool="jwt"` label; no custom metric. Optionally add a `mode` dimension is deliberately **not** added to avoid unbounded label cardinality from user input — keep only the shared `tool` label per `PLAN_ARCHITECTURE.md`.

## Unit tests

`internal/tools/jwt/jwt_test.go`:
- Decode a well-formed unsigned-inspection token → header/claims parsed correctly.
- Decode with correct secret → `Valid == true`.
- Decode with wrong secret → `Valid == false`, no error (still inspectable).
- Decode malformed token → error.
- Encode with valid claims + HS256 → produces a token that this package's own `Decode` with the same secret reports as valid (round-trip test).
- Encode with empty claims → error.
- Encode with unsupported algorithm → error.

## Documentation

- `docs/api/jwt.md`, `docs/cli/jwt.md`, `docs/testing/jwt.md`.

## Skill

`.skills/jwt/SKILL.md` — triggers on "implement JWT encode/decode"; documents the unverified-inspection vs. verified-decode distinction and the HMAC-only encode scope, links to this plan.

## New dependencies

`github.com/golang-jwt/jwt/v5` (added to `go.mod`).

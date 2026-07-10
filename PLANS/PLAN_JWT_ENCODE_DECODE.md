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

Decodes a JWT to inspect its header and claims (without requiring any signing key), and encodes a new signed JWT from user-supplied claims, for inspection and testing purposes.

**Revised after initial implementation**, at the user's explicit request to use `golang-jwt/jwt` (already a dependency, but only exercised for HMAC) "to allow support to many algorithms" plus a web UI combobox to pick one: `Encode`/`Decode` now support the full algorithm set `golang-jwt/jwt/v5` offers — HMAC (`HS256`/`HS384`/`HS512`, unchanged), RSA (`RS256`/`RS384`/`RS512`), RSA-PSS (`PS256`/`PS384`/`PS512`), ECDSA (`ES256`/`ES384`/`ES512`), and EdDSA. The default algorithm stays `HS256`, per the explicit requirement that "the default algorithm must be the [one] currently used."

## Business logic

Package: `internal/tools/jwttool/jwttool.go` (not `jwt`, to avoid colliding with the `golang-jwt/jwt/v5` import name).

```go
package jwttool

var SupportedAlgorithms = []string{
    "HS256", "HS384", "HS512",
    "RS256", "RS384", "RS512",
    "PS256", "PS384", "PS512",
    "ES256", "ES384", "ES512",
    "EdDSA",
}
const DefaultAlgorithm = "HS256"

type DecodeResult struct {
    Header    map[string]any `json:"header"`
    Claims    map[string]any `json:"claims"`
    Signature string         `json:"signature"` // base64url, not validated
    Valid     *bool          `json:"valid,omitempty"` // set only if secret/key was supplied for verification
}

func Decode(token string, secret string, key string) (DecodeResult, error)
func Encode(claims map[string]any, secret string, key string, algorithm string) (string, error)
```

Implementation via `golang-jwt/jwt/v5`:
- `Decode` with both `secret` and `key` empty parses the token without verifying the signature (`jwt.ParseUnverified`), purely for inspection — this mirrors the tool's purpose ("inspection and testing"), not production token validation.
- `Decode` with either non-empty additionally attempts verification and sets `Valid`. Which of `secret`/`key` is actually used is decided by the token's own `alg` header (via `t.Method`'s concrete type in the `Keyfunc`), not by which field the caller filled in — so pasting a value into the "wrong" field for the token's algorithm fails verification (`Valid: false`) instead of erroring or (worse) silently succeeding.
- `Encode` picks the signing method and key type from the `algorithm` parameter: HMAC algorithms sign with `[]byte(secret)`; every other algorithm parses `key` as a PEM private key via the matching `jwt.Parse*PrivateKeyFromPEM` helper (`ParseRSAPrivateKeyFromPEM` for RSA/RSA-PSS, `ParseECPrivateKeyFromPEM` for ECDSA, `ParseEdPrivateKeyFromPEM` for EdDSA).

Edge cases:
- Malformed token (not 3 dot-separated base64url segments) → `INVALID_TOKEN`, HTTP 400.
- Unsupported algorithm string on encode → `UNSUPPORTED_ALGORITHM`, HTTP 400.
- Malformed/wrong-type PEM key on encode → `INVALID_KEY`, HTTP 400 (distinct from `UNSUPPORTED_ALGORITHM` — a good algorithm name with a bad key).
- Empty claims on encode → `EMPTY_CLAIMS`, HTTP 400.
- Header/claims segments that aren't valid JSON → `INVALID_TOKEN`, HTTP 400.
- Verifying with the wrong field for the token's algorithm family (e.g. a `secret` against an RS256 token) → `Valid: false`, not an error — verified directly, not assumed.

## CLI

```
mytoolkit jwt --decode --token <token> [--secret <secret>] [--key <file|->] [--out <file|->]
mytoolkit jwt --encode --claims <file|-> [--secret <secret>] [--key <file|->] [--algorithm HS256] [--out <file|->]
```

`--decode` and `--encode` are mutually exclusive. `--key` reads a **file path** (or `-` for stdin) — unlike `--secret`, a PEM key doesn't fit as a single shell argument value. `--algorithm` defaults to `HS256`, unchanged.

Examples:
```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiJ9.eyJzdWIiOiIxMjMifQ.abc123
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"abc123"}

$ echo '{"sub":"123","exp":1735689600}' | mytoolkit jwt --encode --secret mysecret
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMiLCJleHAiOjE3MzU2ODk2MDB9.xyz789
```

RSA (real key, verified against the running binary — see `docs/cli/jwt.md`):
```
$ echo '{"sub":"1234","name":"Aecio"}' | mytoolkit jwt --encode --algorithm RS256 --key rsa_priv.pem
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiQWVjaW8iLCJzdWIiOiIxMjM0In0.SCq3ovM-...
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

Encode request with an asymmetric algorithm uses `key` (PEM) instead of `secret`:
```json
{ "input": "{\"sub\":\"1234\"}", "options": { "mode": "encode", "algorithm": "RS256", "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" } }
```

Error (400):
```json
{ "success": false, "error": { "code": "INVALID_TOKEN", "message": "token contains an invalid number of segments" } }
```

New error code for this revision: `INVALID_KEY` (malformed or wrong-type PEM key on encode).

## Web UI

- Single form (not tabs, matching the actual implementation): Mode select (Decode/Encode), Algorithm select (all 13 supported algorithms, `HS256` first/default — same order as `SupportedAlgorithms`), a one-line Secret text input (HMAC only), and a full-width "Key (PEM)" `<textarea>` (asymmetric algorithms only — private key to Encode, public key to verify a Decode), with a label stating which field applies to which algorithm family so users don't have to guess.
- Token/claims input textarea + read-only result textarea, reusing the shared `tool-panel` partial's Copy/Reset buttons.
- Explicit note near the panel that secrets *and keys* are never persisted server-side and are only used in-memory for the single request (privacy/security note relevant given this is a "testing" tool) — updated from the original secret-only wording.

## Metrics

Shared `tool="jwt"` label; no custom metric. Optionally add a `mode` dimension is deliberately **not** added to avoid unbounded label cardinality from user input — keep only the shared `tool` label per `PLAN_ARCHITECTURE.md`.

## Unit tests

`internal/tools/jwttool/jwttool_test.go`:
- Decode a well-formed unsigned-inspection token → header/claims parsed correctly.
- Decode with correct secret → `Valid == true`. Decode with wrong secret → `Valid == false`, no error (still inspectable). Decode malformed token → error.
- Encode with valid claims + HS256 → produces a token that this package's own `Decode` with the same secret reports as valid (round-trip test). Default algorithm (empty string) resolves to `HS256`.
- Encode with empty claims → error. Encode with an unsupported algorithm string → error.
- Full encode→decode round trips against freshly-generated (not checked-in) key pairs for all six RSA-family algorithms, `ES256`, and `EdDSA`.
- Decode fails (`Valid: false`, not an error) when verified against a different key pair's public key, or when an HMAC secret is supplied for an RSA-signed token.
- Encode with an invalid/malformed PEM key string → `INVALID_KEY` error.

## Documentation

- `docs/api/jwt.md`, `docs/cli/jwt.md`, `docs/testing/jwt.md`.

## Skill

`.skills/jwt/SKILL.md` — triggers on "implement JWT encode/decode", "add JWT algorithm"; documents the unverified-inspection vs. verified-decode distinction, the secret-vs-key-by-algorithm-family split, and the steps to add a new algorithm, links to this plan.

## New dependencies

`github.com/golang-jwt/jwt/v5` (already a dependency; this revision uses more of its surface — `Parse{RSA,EC,Ed}{Private,Public}KeyFromPEM`, `SigningMethodRS*`/`PS*`/`ES*`/`EdDSA` — not a new module).

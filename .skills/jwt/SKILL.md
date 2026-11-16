---
name: jwt
description: Implement or modify the JWT tool (internal/tools/jwttool) — decode/verify (secret, PEM, JWK or JWKS), offline signing with registered + custom string/list claims, key generation (HMAC/RSA/EC/Ed25519) and JWK/JWKS conversion, via golang-jwt/jwt/v5. Trigger on "implement JWT encode/decode", "add JWT algorithm", "JWT claims", "JWKS", "JWT keygen".
---

# JWT tool

Package is `jwttool`, not `jwt` — avoids colliding with the `github.com/golang-jwt/jwt/v5` import name. Everything runs **offline**: no JWKS URL is ever fetched; verification only uses the key material the caller passes.

| File | Responsibility |
|---|---|
| `jwttool.go` | `Decode`, `Encode(claims, EncodeOptions)`, `lookupMethod`, `signingMethodAndKey`, generic `parsePrivateKey[K]` |
| `claims.go` | `ClaimsSpec`, `BuildClaims`, `ParseClaimPairs`, `ParseClaimsJSON`, `ParseDuration`, `SplitLines`/`SplitList` |
| `keys.go` | `GenerateKey`, `BuildJWKS`, `JWK`/`JWKSet`, `JWK.Key()`, `JWK.Thumbprint()` (RFC 7638), PEM/JWK/JWKS parsing, `checkKeyFitsAlgorithm` |

## Decode: key chosen by the token, never by the caller

`Decode(token, secret, key)` always parses unverified first so header/claims/signature are inspectable without key material. If `secret != "" || key != ""` it verifies with `jwt.WithValidMethods(SupportedAlgorithms)` (so `none` is never valid) and `jwt.WithIssuedAt()`, setting `Valid` and `VerifyError`.

`verificationKey()` switches on **the token's own parsed `t.Method`**: HMAC → `secret`; everything else → `key`. Two invariants, both regression-tested:

- An HMAC token with an empty `secret` returns an **error**, not `[]byte("")`. Returning an empty key verified tokens forged with an empty secret (`TestDecodeHMACTokenWithOnlyKeyIsInvalid`).
- `key` may be PEM (public, private or certificate; multiple blocks), a JWK or a JWKS. Candidates are filtered by the token `kid` (when both sides have one), JWK `alg`, and `checkKeyFitsAlgorithm`; one survivor is returned directly, several as a `jwt.VerificationKeySet` (the parser tries each).

Don't refactor this to take an explicit "which key type" parameter.

## Encode

`Encode(claims, EncodeOptions{Algorithm, Secret, Key, KeyID})`. `lookupMethod` restricts to `SupportedAlgorithms`. HMAC requires a non-empty secret (`MISSING_SECRET`). Asymmetric keys go through `parsePrivateKey[K]` — a JWK when the input starts with `{` (its `kid` becomes the default header `kid`), otherwise the matching `jwt.Parse*PrivateKeyFromPEM` — then `checkKeyFitsAlgorithm` (e.g. ES384 needs P-384), so mismatches are `400 INVALID_KEY`, not `500 SIGNING_FAILED`.

## Claims

Adapters build claims as: `ParseClaimsJSON(raw)` → `BuildClaims(base, ClaimsSpec, time.Now())` → `Encode`. `now` is a parameter so tests are deterministic. Precedence: base JSON < `Custom` pairs < registered fields.

Custom pairs (`ParseClaimPairs`): `name=value` → string; a repeated name → `[]string`; `name[]=value` → forced `[]string`. `exp`/`nbf`/`iat` are rejected as pairs (they must be NumericDates).

`ParseClaimsJSON` uses **`encoding/json/v2`** (new code on Go 1.27, per the `use-modern-go` guidelines): it rejects duplicate member names and invalid UTF-8 (token-confusion vectors), and the `exactNumbers` unmarshaler keeps numbers as `json.Number` so large integers survive exactly. Existing wire output (CLI `--decode` JSON, REST envelope) stays on v1 `encoding/json` on purpose — v2 doesn't sort map keys by default and would change documented output.

## Keys and JWKS

`GenerateKey`: HS* → random secret (hash-size bytes, base64url text) + `oct` JWK, **no** JWKS (a shared secret must never be published); RS*/PS* → RSA (`SupportedRSABits`, default 2048); ES* → the RFC 7518 curve for the algorithm; EdDSA → Ed25519. Default `kid` = RFC 7638 thumbprint (tested against the RFC's §3.1 vector). EC coordinates use Go 1.25+ `ecdsa.PublicKey.Bytes()`/`PrivateKey.Bytes()` and `ParseUncompressedPublicKey`/`ParseRawPrivateKey` — never the deprecated `X`/`Y` big.Int fields.

`BuildJWKS` always emits public JWKs only (private input is reduced via `crypto.Signer.Public()`).

## Adding a new algorithm

1. Add it to `SupportedAlgorithms` **and** to both `<select name="algorithm">` lists in `jwt.html` (they're not generated from the Go slice).
2. If it's a new key family: add cases to `signingMethodAndKey`, `checkKeyFitsAlgorithm`, `inferAlgorithm`, `GenerateKey`, `publicJWK`/`privateJWK`/`JWK.Key()`.
3. `TestGenerateKeySignVerifyAllForms` iterates `SupportedAlgorithms`, so a new asymmetric algorithm is covered automatically once keygen supports it.

## `DefaultAlgorithm` must stay `"HS256"`

The feature request that added the other algorithms was explicit that the default must not change. An empty algorithm string means `HS256` (`cmp.Or(opts.Algorithm, DefaultAlgorithm)`).

## Adapters

- CLI (`app/internal/cli/jwt.go`): `--decode`/`--encode` on `jwt`, plus `jwt keygen` and `jwt jwks` subcommands. Claim flags `--iss/--sub/--aud/--exp/--nbf/--iat/--jti/--claim/--kid`. `--key` reads a file (or `-`); unset means "no key". With the default `--claims -`, an interactive terminal stdin is skipped (`readClaimsInput`) so flag-only tokens don't hang. `keygen --private-out` writes with mode `0600` (`textio.WritePrivate`).
- REST: bespoke `jwtHandler`, `options.mode` ∈ `decode|encode|keygen|jwks` (unknown → `INVALID_OPTION`). `audience`/`claims` are `flexList`: a JSON array, or a string split by `jwttool.SplitList` (commas/newlines) / `SplitLines` (newlines), since the web form always sends strings.
- Web (`jwt.html`): one page, four modes; an inline script shows/hides `[data-modes]` sections and **disables** hidden controls — `tool-common.js`'s `collectOptions()` skips disabled controls, which is how the two same-named `algorithm` selects (sign vs. JWKS "Auto") never collide.
- MCP: four tools, `jwt_decode`/`jwt_encode`/`jwt_keygen`/`jwt_jwks` (`app/internal/mcp/jwt.go`); see `.skills/mcp/SKILL.md`.

Plan: `PLANS/PLAN_JWT_ENCODE_DECODE.md`. Docs: `docs/api|cli|testing/jwt.md`.

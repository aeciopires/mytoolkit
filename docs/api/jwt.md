<!-- TOC -->

- [JWT Encode/Decode — REST API](#jwt-encodedecode--rest-api)
  - [Modes](#modes)
  - [Decode / verify (`mode: decode`, default)](#decode--verify-mode-decode-default)
  - [Encode / sign (`mode: encode`)](#encode--sign-mode-encode)
    - [Custom claims (string or list of strings)](#custom-claims-string-or-list-of-strings)
  - [Generate a key (`mode: keygen`)](#generate-a-key-mode-keygen)
  - [Build a JWKS (`mode: jwks`)](#build-a-jwks-mode-jwks)
  - [End-to-end: keygen → sign → verify with a JWKS](#end-to-end-keygen--sign--verify-with-a-jwks)
  - [Error response (400)](#error-response-400)
  - [Workflow](#workflow)

<!-- TOC -->

# JWT Encode/Decode — REST API

`POST /api/v1/tools/jwt` — built on [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt). **Every mode runs offline**: no token, key or JWKS is ever fetched from (or sent to) the network — JWKS verification uses the JWKS you paste, never a URL.

Supported algorithms: `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `PS256`, `PS384`, `PS512`, `ES256`, `ES384`, `ES512`, `EdDSA` (default `HS256`). `none` is never accepted.

## Modes

| `options.mode` | `input` | Result (`data`) |
|---|---|---|
| `decode` (default) | a token | `header`, `claims`, `signature`, and — when `secret`/`key` is supplied — `valid` (+ `verify_error` on failure) |
| `encode` | optional base claims JSON object | `output`: the signed token |
| `keygen` | ignored | `algorithm`, `kid`, `secret` (HS\*) or `private_key`/`public_key` (PEM), `private_jwk`, `public_jwk`, `jwks` |
| `jwks` | PEM key(s)/certificate(s), a JWK, or a JWKS | `keys`: a JWKS of public keys only |

Any other `mode` is rejected with `INVALID_OPTION` (it previously fell back silently to decode).

## Decode / verify (`mode: decode`, default)

```json
{ "input": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls", "options": { "mode": "decode", "secret": "mysecret" } }
```

```json
{"success":true,"data":{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls","valid":true},"meta":{"tool":"jwt","duration_ms":0.072}}
```

- No `secret`/`key` → inspection only (`valid` absent).
- `secret` verifies HMAC tokens (`HS*`). `key` verifies every other algorithm and accepts a **PEM** public key, private key (its public half is used) or certificate — several concatenated PEM blocks are all tried — a **JWK**, or a **JWKS** (`{"keys":[...]}`), in which case the key whose `kid` matches the token header is selected.
- Which of `secret`/`key` applies is decided by the token's own `alg` header. Supplying the wrong one yields `valid: false`, never a false positive — in particular an `HS*` token is **never** verified against an empty secret.
- Verification also checks the time-based claims `exp`, `nbf` and `iat`; on failure `verify_error` explains why, e.g. `"token has invalid claims: token is expired"` or `"token signature is invalid: signature is invalid"`.

## Encode / sign (`mode: encode`)

Backwards-compatible request (unchanged):

```json
{ "input": "{\"sub\":\"123\"}", "options": { "mode": "encode", "secret": "mysecret", "algorithm": "HS256" } }
```

```json
{"success":true,"data":{"output":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls"},"meta":{"tool":"jwt","duration_ms":0.048}}
```

| Option | Type | Meaning |
|---|---|---|
| `algorithm` | string | signing algorithm, default `HS256` |
| `secret` | string | HMAC secret — **required and non-empty** for `HS*` |
| `key` | string | private key for `RS*`/`PS*`/`ES*`/`EdDSA`: PEM (PKCS#1, PKCS#8, SEC 1) or a private JWK (e.g. `private_jwk` from `keygen`). `ES*` keys must be on the matching curve (P-256/P-384/P-521). |
| `kid` | string | `kid` header; defaults to the private JWK's `kid` when `key` is a JWK |
| `issuer` / `subject` | string | `iss` / `sub` |
| `audience` | string[] or string | `aud`; a string is split on commas/newlines. One value is emitted as a string, several as a list. |
| `expires_in` | duration | sets `exp` = now + duration (`15m`, `1h`, `7d`, …) |
| `not_before` | duration | sets `nbf` = now + duration (`0s`, `-30s`, …) |
| `issued_at` | bool | sets `iat` = now |
| `jwt_id` | bool | sets `jti` to a random value |
| `claims` | string[] or string | custom claims as `name=value` pairs (a string is split on newlines; blank and `#` lines ignored) — see below |

`input` (base claims JSON) is optional when the other options produce at least one claim. Precedence, lowest to highest: `input` JSON → `claims` pairs → `issuer`/`subject`/`audience`/time options. Base claims JSON must be an object; duplicate member names (e.g. two `"sub"`) are **rejected**, since verifiers disagree on which one wins.

### Custom claims (string or list of strings)

| Pair(s) | Resulting claim |
|---|---|
| `role=admin` | `"role": "admin"` |
| `groups=dev`, `groups=ops` (repeated name) | `"groups": ["dev", "ops"]` |
| `scopes[]=read` (forced list) | `"scopes": ["read"]` |

The value is everything after the first `=`, verbatim. `exp`/`nbf`/`iat` cannot be custom claims (they must be numeric dates — use `expires_in`/`not_before`/`issued_at`).

Request with array options:

```json
{"input":"","options":{"mode":"encode","secret":"mysecret","subject":"user-1","issuer":"https://auth.example.com","audience":["api","web"],"claims":["role=admin","groups=dev","groups=ops"]}}
```

```json
{"success":true,"data":{"output":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYXBpIiwid2ViIl0sImdyb3VwcyI6WyJkZXYiLCJvcHMiXSwiaXNzIjoiaHR0cHM6Ly9hdXRoLmV4YW1wbGUuY29tIiwicm9sZSI6ImFkbWluIiwic3ViIjoidXNlci0xIn0.ergDDMYdttxxpswS-N_HaWD7qOp2DbKzXm6cQTMf_20"},"meta":{"tool":"jwt","duration_ms":0.264}}
```

The same request in the string form the web UI sends, plus a base claims object:

```json
{"input":"{\"tenant\":\"acme\"}","options":{"mode":"encode","secret":"mysecret","subject":"user-1","audience":"api, web","claims":"role=admin\ngroups=dev\ngroups=ops"}}
```

```json
{"success":true,"data":{"output":"eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJhdWQiOlsiYXBpIiwid2ViIl0sImdyb3VwcyI6WyJkZXYiLCJvcHMiXSwicm9sZSI6ImFkbWluIiwic3ViIjoidXNlci0xIiwidGVuYW50IjoiYWNtZSJ9.BfDiEZLsQ_s59WzVQQ0ICzCGFiVxQsP9t_lJhk9yCc0"},"meta":{"tool":"jwt","duration_ms":0.269}}
```

## Generate a key (`mode: keygen`)

Options: `algorithm` (default `HS256`), `bits` (RSA only: `2048` default, `3072`, `4096`), `kid` (default: the [RFC 7638](https://www.rfc-editor.org/rfc/rfc7638) JWK thumbprint).

| Algorithm | Generated material |
|---|---|
| `HS256`/`HS384`/`HS512` | random secret (32/48/64 bytes, base64url text) + `oct` private JWK. No public key and no JWKS — a shared secret must never be published. |
| `RS*`/`PS*` | RSA key |
| `ES256`/`ES384`/`ES512` | EC key on P-256/P-384/P-521 |
| `EdDSA` | Ed25519 key |

```json
{"input":"","options":{"mode":"keygen","algorithm":"EdDSA","kid":"ed-1"}}
```

```json
{"success":true,"data":{"algorithm":"EdDSA","kid":"ed-1","private_key":"-----BEGIN PRIVATE KEY-----\nMC4CAQAwBQYDK2VwBCIEICIXyv35JtuAuOBUYiAhU60afeuFHq7+IbGUu5IPK2b/\n-----END PRIVATE KEY-----\n","public_key":"-----BEGIN PUBLIC KEY-----\nMCowBQYDK2VwAyEAKR2n8dISVpujb6q28FC5VXWBtvH79Ocy6M+wR1hP11E=\n-----END PUBLIC KEY-----\n","private_jwk":{"kty":"OKP","use":"sig","alg":"EdDSA","kid":"ed-1","crv":"Ed25519","x":"KR2n8dISVpujb6q28FC5VXWBtvH79Ocy6M-wR1hP11E","d":"IhfK_fkm24C44FRiICFTrRp964Uerv4hsZS7kg8rZv8"},"public_jwk":{"kty":"OKP","use":"sig","alg":"EdDSA","kid":"ed-1","crv":"Ed25519","x":"KR2n8dISVpujb6q28FC5VXWBtvH79Ocy6M-wR1hP11E"},"jwks":{"keys":[{"kty":"OKP","use":"sig","alg":"EdDSA","kid":"ed-1","crv":"Ed25519","x":"KR2n8dISVpujb6q28FC5VXWBtvH79Ocy6M-wR1hP11E"}]}},"meta":{"tool":"jwt","duration_ms":3.639}}
```

(A throwaway key generated for this document. Keys are generated with `crypto/rand` and never stored, but the private key does travel back over this HTTP response — for production signing keys prefer `mytoolkit jwt keygen` locally, see [docs/cli/jwt.md](../cli/jwt.md).)

## Build a JWKS (`mode: jwks`)

Options: `algorithm` (default: inferred — RSA → `RS256`, EC → `ES256`/`ES384`/`ES512` by curve, Ed25519 → `EdDSA`), `kid` (single-key input only; default: RFC 7638 thumbprint). Private input is reduced to its public half, so the result is always safe to publish. Multiple PEM blocks yield a multi-key set (e.g. current + next key during rotation).

```json
{"input":"-----BEGIN PUBLIC KEY-----\nMFkwEwYHKoZIzj0CAQYIKoZIzj0DAQcDQgAEoYOWCdz0ZVuUFizRWrjy0DoFdpRJ\nz5v6Cncv/L9YSqNJOeSktkKLAn4ulkxKrEhoRZCE1Wa8BsGkP/K8cCCaeA==\n-----END PUBLIC KEY-----\n","options":{"mode":"jwks","kid":"my-key-2026"}}
```

```json
{"success":true,"data":{"keys":[{"kty":"EC","use":"sig","alg":"ES256","kid":"my-key-2026","crv":"P-256","x":"oYOWCdz0ZVuUFizRWrjy0DoFdpRJz5v6Cncv_L9YSqM","y":"STnkpLZCiwJ-LpZMSqxIaEWQhNVmvAbBpD_yvHAgmng"}]},"meta":{"tool":"jwt","duration_ms":0.13}}
```

## End-to-end: keygen → sign → verify with a JWKS

A token signed by `mytoolkit jwt --encode` (ES256, `kid` `my-key-2026`, custom string and list claims — see [docs/cli/jwt.md](../cli/jwt.md)) verified against the JWKS above. Output captured when the token was fresh; once its `exp` passes, the same request returns `"valid":false,"verify_error":"token has invalid claims: token is expired"`.

```json
{"input":"eyJhbGciOiJFUzI1NiIsImtpZCI6Im15LWtleS0yMDI2IiwidHlwIjoiSldUIn0.eyJhdWQiOlsiYXBpIiwid2ViIl0sImV4cCI6MTc5NDgxOTI4OSwiZ3JvdXBzIjpbImRldiIsIm9wcyJdLCJpYXQiOjE3OTQ4MTU2ODksImlzcyI6Imh0dHBzOi8vYXV0aC5leGFtcGxlLmNvbSIsImp0aSI6IjdXSUlaUEVHSU5OWjJGS0kzRTVUV09GVlRUIiwicm9sZSI6ImFkbWluIiwic2NvcGVzIjpbInJlYWQiXSwic3ViIjoidXNlci00MiJ9.HJpw2FzWDaD1CGGT2iWxQo_NSUeTmLAcKLOMOt_BpRLGPaCfh0j0kC1BRW600wjPbs5FC-SyfxW6hbONt-9i8A","options":{"mode":"decode","key":"{\n  \"keys\": [\n    {\n      \"kty\": \"EC\",\n      \"use\": \"sig\",\n      \"alg\": \"ES256\",\n      \"kid\": \"my-key-2026\",\n      \"crv\": \"P-256\",\n      \"x\": \"oYOWCdz0ZVuUFizRWrjy0DoFdpRJz5v6Cncv_L9YSqM\",\n      \"y\": \"STnkpLZCiwJ-LpZMSqxIaEWQhNVmvAbBpD_yvHAgmng\"\n    }\n  ]\n}\n"}}
```

```json
{"success":true,"data":{"header":{"alg":"ES256","kid":"my-key-2026","typ":"JWT"},"claims":{"aud":["api","web"],"exp":1794819289,"groups":["dev","ops"],"iat":1794815689,"iss":"https://auth.example.com","jti":"7WIIZPEGINNZ2FKI3E5TWOFVTT","role":"admin","scopes":["read"],"sub":"user-42"},"signature":"HJpw2FzWDaD1CGGT2iWxQo_NSUeTmLAcKLOMOt_BpRLGPaCfh0j0kC1BRW600wjPbs5FC-SyfxW6hbONt-9i8A","valid":true},"meta":{"tool":"jwt","duration_ms":0.2}}
```

## Error response (400)

```json
{"success":false,"error":{"code":"INVALID_TOKEN","message":"token contains an invalid number of segments"}}
```

```json
{"success":false,"error":{"code":"MISSING_SECRET","message":"a non-empty secret is required for HS256"}}
```

```json
{"success":false,"error":{"code":"INVALID_OPTION","message":"mode must be one of [decode encode keygen jwks], got sign"}}
```

| Code | When |
|---|---|
| `INVALID_TOKEN` | malformed token (decode) |
| `INVALID_CLAIMS` | base claims not a JSON object / invalid JSON / duplicate member, malformed `name=value` pair, `exp`/`nbf`/`iat` as a custom claim, negative `expires_in` |
| `EMPTY_CLAIMS` | encode would produce a token with no claims |
| `MISSING_SECRET` | `HS*` encode without a secret (previously signed with an empty key — a token anyone could forge) |
| `INVALID_KEY` | malformed/wrong-type key, EC curve not matching `ES*`, JWKS input without keys |
| `UNSUPPORTED_ALGORITHM` | algorithm not in the list above |
| `INVALID_OPTION` | unknown `mode`, unsupported RSA `bits`, bad duration, `kid` with multi-key JWKS input |
| `INVALID_REQUEST` | body is not valid JSON or exceeds the 10 MiB request limit |

Secrets and keys are used in-memory for the single request only and are never persisted server-side.

## Workflow

```mermaid
sequenceDiagram
    participant C as Client (Web/REST/CLI)
    participant H as jwtHandler (bespoke)
    participant T as internal/tools/jwttool

    C->>H: POST /api/v1/tools/jwt {input, options.mode}
    alt mode == encode
        H->>T: ParseClaimsJSON(input) + BuildClaims(spec, now)
        H->>T: Encode(claims, {algorithm, secret, key, kid})
        T->>T: HMAC secret, or parse PEM / private JWK, check key fits algorithm
        T-->>H: signed token
        H-->>C: 200 {data.output}
    else mode == keygen
        H->>T: GenerateKey({algorithm, bits, kid})
        T-->>H: secret or PEM pair + JWKs + JWKS
        H-->>C: 200 {data: key pair}
    else mode == jwks
        H->>T: BuildJWKS(input, {algorithm, kid})
        T-->>H: public-only JWKS
        H-->>C: 200 {data.keys}
    else mode == decode (default)
        H->>T: Decode(token, secret, key)
        T->>T: secret for HS*, else PEM / JWK / JWKS by the token's alg and kid
        T-->>H: header, claims, signature, valid?, verify_error?
        H-->>C: 200 {data}
    end
    T--xH: apperr.Error (invalid token / claims / key / option, unsupported algorithm)
    H-->>C: 4xx {success:false, error}
```

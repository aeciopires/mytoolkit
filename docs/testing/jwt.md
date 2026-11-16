<!-- TOC -->

- [JWT Encode/Decode — Testing](#jwt-encodedecode--testing)
  - [Unit tests](#unit-tests)
  - [Regression tests for fixed bugs](#regression-tests-for-fixed-bugs)
  - [MCP adapter tests](#mcp-adapter-tests)
  - [End-to-end checks against the real binary](#end-to-end-checks-against-the-real-binary)
  - [Manual cross-check against OpenSSL-generated keys](#manual-cross-check-against-openssl-generated-keys)

<!-- TOC -->

# JWT Encode/Decode — Testing

## Unit tests

```
$ cd app && go test -race ./internal/tools/jwttool/... ./internal/mcp/...
ok  	github.com/aeciopires/mytoolkit/internal/tools/jwttool
ok  	github.com/aeciopires/mytoolkit/internal/mcp
```

`app/internal/tools/jwttool/`:

- `jwttool_test.go` — the original HMAC round trip (encode/decode, wrong secret, no secret, malformed token, empty claims, unsupported algorithm), `HS256` staying the default, round trips for all RSA/RSA-PSS algorithms plus `ES256`/`EdDSA` against in-test generated keys, a different key pair failing verification, invalid PEM producing a clear error, and an HMAC secret against an RSA token failing cleanly.
- `keys_test.go` — `TestGenerateKeySignVerifyAllForms`: for **every** asymmetric algorithm, `GenerateKey` → sign with both the PEM and the private JWK → verify with the PEM public key, the public JWK, the JWKS, and the private PEM (public half derived). Also: HMAC keygen (secret length, `oct` JWK, no JWKS), RSA `bits` and `none` rejection, JWKS key selection by `kid` (and a clear `verify_error` naming an unknown `kid`), multi-block PEM → multi-key JWKS (private input reduced to public), `kid`/algorithm validation in `BuildJWKS`, the **RFC 7638 §3.1 thumbprint test vector**, and an `ES384`/P-256 curve mismatch reported as `INVALID_KEY`.
- `claims_test.go` — `ParseClaimPairs` (string, repeated-name list, `name[]=` forced list, values containing `=`/`,`, malformed and reserved `exp`/`iat` pairs), `BuildClaims` precedence and deterministic `exp`/`nbf`/`iat` from a fixed `now`, single vs multiple audience, generated `jti`, `exp` enforcement on decode, `ParseClaimsJSON` (blank/`null` → empty, non-object/trailing data rejected, **duplicate member names rejected**, integers above 2^53 preserved exactly in the signed payload), `ParseDuration` (`d` suffix), and the split helpers.

## Regression tests for fixed bugs

| Test | Bug |
|---|---|
| `TestEncodeHMACRequiresSecret` | `HS*` encode with an empty secret used to succeed, producing a token anyone can forge. |
| `TestDecodeHMACTokenWithOnlyKeyIsInvalid` | Verifying an `HS*` token while supplying only `key` used an empty HMAC key, so a token forged with an empty secret reported `valid: true`. |
| `TestDecodeRejectsAlgNone` | `alg: none` tokens are never reported valid (`jwt.WithValidMethods`). |
| `TestEncodeECDSACurveMismatchIsInvalidKey` | A curve/algorithm mismatch surfaced as a generic `500 SIGNING_FAILED` instead of `400 INVALID_KEY`. |
| `TestParseClaimsJSONStrict` | Duplicate claim names were silently accepted (last one won) — a token-confusion vector. |

## MCP adapter tests

`app/internal/mcp/jwt_test.go` — `TestHandleJWTKeygenEncodeVerifyWithJWKS` chains the four MCP tools: `jwt_keygen` (ES256) → `jwt_jwks` → `jwt_encode` (with `kid`, `subject`, `audience`, `expires_in`, custom string and list claims) → `jwt_decode` against the JWKS, asserting `valid: true` and the list-valued claim. `server_test.go` asserts all 18 MCP tools are registered (including `jwt_keygen`/`jwt_jwks`).

## End-to-end checks against the real binary

Every example in [docs/cli/jwt.md](../cli/jwt.md) and [docs/api/jwt.md](../api/jwt.md) was produced by running the command/request verbatim against a freshly built binary. The web page was driven with Playwright against the running server through all four modes — keygen (ES256) → encode (subject, comma-separated audience, `expires_in`, `issued_at`, custom claims textarea, PEM private key) → JWKS from the public key → decode with that JWKS — asserting `valid: true`, `groups == ["dev","ops"]`, `aud == ["api","web"]`, and zero console errors. Re-run that flow whenever `jwt.html` or `tool-common.js` changes.

## Manual cross-check against OpenSSL-generated keys

The Go-generated test keys exercise the code path, but real-world users paste keys from `openssl`/`ssh-keygen`/cloud KMS exports, which can differ in PEM header (`PRIVATE KEY` vs `RSA PRIVATE KEY` vs `EC PRIVATE KEY`) or encoding details. Verified manually against `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048` output (see `docs/cli/jwt.md`'s RSA example) through the CLI, REST API, and the running web page. Re-run this manual check if the PEM/JWK parsing in `keys.go` (`parseKeys`, `parsePEMBlock`) or `jwttool.go` (`parsePrivateKey`) changes.

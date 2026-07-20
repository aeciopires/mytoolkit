---
name: jwt
description: Implement or modify the JWT Encode/Decode tool (internal/tools/jwttool) — inspection-mode decode plus multi-algorithm encode (HMAC, RSA, RSA-PSS, ECDSA, EdDSA) via golang-jwt/jwt/v5. Trigger on "implement JWT encode/decode", "add JWT algorithm".
---

# JWT Encode/Decode

Package is `jwttool`, not `jwt` — avoids colliding with the `github.com/golang-jwt/jwt/v5` import name. `app/internal/tools/jwttool/jwttool.go`.

`Decode(token, secret, key)` always parses unverified first (via `jwt.NewParser().ParseUnverified`) so header/claims/signature are inspectable without any key material; if `secret != "" || key != ""` it additionally attempts verification and sets `Valid`.

## Two key parameters, chosen by algorithm family — not by the caller

`secret` (raw string) is used only for HMAC tokens (`HS256`/`HS384`/`HS512`). `key` (PEM text) is used for every other supported algorithm — `RS256`/`RS384`/`RS512`/`PS256`/`PS384`/`PS512` (RSA), `ES256`/`ES384`/`ES512` (ECDSA), `EdDSA` (Ed25519). On `Encode`, `key` must be a **private** key; on `Decode` verification, `key` must be a **public** key.

`verificationKey()` picks which of `secret`/`key` to use by switching on **the token's own parsed `t.Method` type** (`*jwt.SigningMethodHMAC`, `*jwt.SigningMethodRSA`, `*jwt.SigningMethodRSAPSS`, `*jwt.SigningMethodECDSA`, `*jwt.SigningMethodEd25519`), not on a caller-supplied hint. This means pasting an HMAC secret against an RS256 token doesn't panic or silently mis-verify — it correctly reports `Valid: false` (see `TestDecodeSecretAgainstAsymmetricTokenFailsCleanly`). Don't refactor this to take an explicit "which key type" parameter; deriving it from the token is what makes Decode safe to call generically without the caller having pre-inspected the token.

`signingMethodAndKey()` (used by `Encode`) does the mirror-image job explicitly, since encoding requires *choosing* the algorithm rather than reading it off an existing token: it switches on the `algorithm` string and either returns `[]byte(secret)` (HMAC) or parses `key` via the matching `jwt.Parse*PrivateKeyFromPEM` helper, wrapping any parse failure as `apperr` code `INVALID_KEY` (distinct from `UNSUPPORTED_ALGORITHM`, which is for an unrecognized algorithm string, not a bad key).

## Adding a new algorithm

1. Add the algorithm's string identifier to `SupportedAlgorithms` (used for both the error message and — indirectly, since the web page's `<select>` options aren't generated from it — the web template; **update `jwt.html`'s `<option>` list too**, they're not derived from the Go slice).
2. Add a `case` to `signingMethodAndKey` (encode) mapping the string to the right `jwt.SigningMethod*` constant and PEM parser.
3. Add the method's concrete type to `verificationKey`'s switch (decode) if it's a new family (HMAC/RSA/RSA-PSS/ECDSA/Ed25519 are already covered; a genuinely new family needs a new case).
4. Add a round-trip test following the `TestEncodeDecodeRoundTripRSA`/`ECDSA`/`EdDSA` pattern — generate a real key pair in-test with the matching `crypto/*` package (don't check in fixture PEM keys), encode, decode, assert `Valid == true`.

## `DefaultAlgorithm` must stay `"HS256"`

This was already the tool's only supported algorithm before RSA/ECDSA/EdDSA support was added, and the feature request that added the rest was explicit that the default must not change — `SupportedAlgorithms[0]` and the web `<select>`'s first `<option>` are both `HS256`, and `signingMethodAndKey`'s `case "", DefaultAlgorithm:` treats an empty algorithm string as `HS256` too (matters for CLI/REST callers that omit the field entirely).

Bespoke REST/CLI wiring (`app/internal/cli/jwt.go`), not `handlers.Wrap`, since decode/encode have different option shapes and decode's response isn't a single `output` string. The CLI's `--key` flag reads a **file path** (or `-` for stdin), unlike `--secret` which is the literal value — PEM keys are inherently multi-line and don't belong in a single shell argument. An unset `--key` (default `""`) means "no key," distinguished from `-` (read stdin) by `readKeyFlag()`.

The web page's Key field is a `<textarea data-option name="key">` (not the single-line `<input>` used for Secret), since `tool-common.js`'s `collectOptions()` already handles textareas via its generic `else` branch (`el.value`) — no JS changes were needed there.

MCP: two tools, `jwt_encode`/`jwt_decode` (`app/internal/mcp/jwt.go`) — split from REST's single `mode`-discriminated endpoint into two clean input shapes; see `.skills/mcp/SKILL.md`. Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_JWT_ENCODE_DECODE.md`. Docs: `docs/api|cli|testing/jwt.md`.

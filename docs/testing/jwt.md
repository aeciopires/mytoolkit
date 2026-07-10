<!-- TOC -->

- [JWT Encode/Decode — Testing](#jwt-encodedecode--testing)

<!-- TOC -->

# JWT Encode/Decode — Testing

```
$ cd app && go test ./internal/tools/jwttool/... -v
--- PASS: TestEncodeDecodeRoundTrip (0.00s)
--- PASS: TestDecodeWrongSecret (0.00s)
--- PASS: TestDecodeWithoutSecret (0.00s)
--- PASS: TestDecodeMalformedToken (0.00s)
--- PASS: TestEncodeEmptyClaims (0.00s)
--- PASS: TestEncodeUnsupportedAlgorithm (0.00s)
--- PASS: TestEncodeDefaultAlgorithmIsHS256 (0.00s)
=== RUN   TestEncodeDecodeRoundTripRSA
--- PASS: TestEncodeDecodeRoundTripRSA (0.06s)
    --- PASS: TestEncodeDecodeRoundTripRSA/RS256 (0.00s)
    --- PASS: TestEncodeDecodeRoundTripRSA/RS384 (0.00s)
    --- PASS: TestEncodeDecodeRoundTripRSA/RS512 (0.00s)
    --- PASS: TestEncodeDecodeRoundTripRSA/PS256 (0.00s)
    --- PASS: TestEncodeDecodeRoundTripRSA/PS384 (0.00s)
    --- PASS: TestEncodeDecodeRoundTripRSA/PS512 (0.00s)
--- PASS: TestEncodeDecodeRoundTripECDSA (0.00s)
--- PASS: TestEncodeDecodeRoundTripEdDSA (0.00s)
--- PASS: TestDecodeRSAWrongPublicKeyFails (0.08s)
--- PASS: TestEncodeRSAInvalidKeyPEM (0.00s)
--- PASS: TestDecodeSecretAgainstAsymmetricTokenFailsCleanly (0.09s)
PASS
```

Covers: the original HMAC round trip (encode/decode, wrong secret, no secret, malformed token, empty claims, unsupported algorithm), the default algorithm staying `HS256`, full encode→decode round trips for all six RSA-family algorithms (`RS256`/`RS384`/`RS512`/`PS256`/`PS384`/`PS512`) plus `ES256` and `EdDSA` against freshly-generated in-test key pairs (`crypto/rsa`, `crypto/ecdsa`, `crypto/ed25519` + `x509`/`pem`, not checked-in fixture keys), verification correctly failing against a different key pair's public key, an invalid PEM key producing a clear encode-time error, and pasting an HMAC secret against an RSA-signed token failing verification cleanly instead of panicking or silently succeeding.

## Manual cross-check against OpenSSL-generated keys

The Go-generated test keys above exercise the code path, but real-world users paste keys from `openssl`/`ssh-keygen`/cloud KMS exports, which can differ in PEM header (`PRIVATE KEY` vs `RSA PRIVATE KEY`) or encoding details. Verified manually against `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048` output (see `docs/cli/jwt.md`'s RSA example) through the CLI, REST API, and the actual running web page (Playwright) — all three encode with the private key and verify with the public key correctly. Re-run this manual check if `signingMethodAndKey`/`verificationKey`'s PEM-parsing logic changes.

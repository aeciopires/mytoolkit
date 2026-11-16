<!-- TOC -->

- [JWT Encode/Decode — CLI](#jwt-encodedecode--cli)
  - [Usage](#usage)
  - [Claim flags (`--encode`)](#claim-flags---encode)
  - [Examples](#examples)
    - [Generate keys and a JWKS offline](#generate-keys-and-a-jwks-offline)
    - [Sign with registered and custom claims](#sign-with-registered-and-custom-claims)
    - [Decode and verify](#decode-and-verify)
    - [HMAC](#hmac)
    - [RSA keys from OpenSSL](#rsa-keys-from-openssl)
    - [Errors](#errors)

<!-- TOC -->

# JWT Encode/Decode — CLI

Everything runs **offline** — keys are generated locally with `crypto/rand`, tokens are signed locally, and verification uses the key/JWKS file you pass (no JWKS URL is ever fetched). Built on [`golang-jwt/jwt/v5`](https://github.com/golang-jwt/jwt).

## Usage

```
mytoolkit jwt --decode --token <token> [--secret <secret>] [--key <file|->] [--out <file|->]
mytoolkit jwt --encode [--claims <file|->] [claim flags] [--algorithm HS256] [--secret <secret>] [--key <file|->] [--kid <kid>] [--out <file|->]
mytoolkit jwt keygen [--algorithm HS256] [--bits 2048] [--kid <kid>] [--private-out <file>] [--public-out <file>] [--jwks-out <file>] [--out <file|->]
mytoolkit jwt jwks --key <file|-> [--algorithm <alg>] [--kid <kid>] [--out <file|->]
```

- `--algorithm`: `HS256` (default), `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `PS256`, `PS384`, `PS512`, `ES256`, `ES384`, `ES512`, `EdDSA`.
- `--secret`: raw HMAC shared secret, `HS*` only. **Required** for `HS*` encoding — an empty secret is rejected.
- `--key` (a file path, or `-` for stdin):
  - `--encode`: private key, PEM (PKCS#1/PKCS#8/SEC 1) or a private JWK (e.g. `private_jwk` from `keygen`). `ES*` keys must be on the matching curve.
  - `--decode`: public key, private key or certificate (PEM; several blocks are all tried), a JWK, or a JWKS — a JWKS key is selected by the token's `kid`.
- `--claims`: base claims JSON object (default `-` = stdin). With the default, an interactive terminal stdin is **not** read, so you can build a token purely from flags without the command hanging.
- `keygen`: prints the full key pair as JSON (PEM, private/public JWK, JWKS). With `--private-out`/`--public-out`/`--jwks-out`, writes those files instead (the private key/secret file with mode `0600`) and prints nothing unless `--out` is also given. `HS*` produces a secret only — `--public-out`/`--jwks-out` are rejected for it.
- `jwks`: converts PEM key(s)/certificate(s), a JWK or a JWKS into a JWKS of public keys only (private input is reduced to its public half). `--algorithm` defaults to the key type (RSA → `RS256`, EC → by curve, Ed25519 → `EdDSA`); `--kid` defaults to the RFC 7638 thumbprint.

Decode output adds `valid` (and `verify_error` when it is `false`) only when `--secret`/`--key` is given. Verification also checks `exp`/`nbf`/`iat`.

## Claim flags (`--encode`)

| Flag | Claim |
|---|---|
| `--iss <s>` | `iss` |
| `--sub <s>` | `sub` |
| `--aud <s>` (repeatable) | `aud` — one value is a string, several a list |
| `--exp <duration>` | `exp` = now + duration (`15m`, `1h`, `7d`, …) |
| `--nbf <duration>` | `nbf` = now + duration (`0s`, `-30s`, …) |
| `--iat` | `iat` = now |
| `--jti` | random `jti` |
| `--claim name=value` (repeatable) | custom claim: `role=admin` → `"admin"`; repeating a name (`--claim groups=dev --claim groups=ops`) → `["dev","ops"]`; `name[]=value` forces a list (`["read"]`) |
| `--kid <kid>` | `kid` header (defaults to the JWK's `kid` when `--key` is a private JWK) |

Precedence, lowest to highest: `--claims` JSON → `--claim` pairs → `--iss`/`--sub`/`--aud`/time flags.

## Examples

Every output below was produced by running the command against the real binary (keys and tokens are throwaway values generated for this document).

### Generate keys and a JWKS offline

```
$ mytoolkit jwt keygen --algorithm ES256 --kid my-key-2026 --private-out es256_priv.pem --public-out es256_pub.pem --jwks-out jwks.json
$ ls -l es256_priv.pem es256_pub.pem jwks.json | awk '{print $1, $NF}'
-rw------- es256_priv.pem
-rw-r--r-- es256_pub.pem
-rw-r--r-- jwks.json
$ cat jwks.json
{
  "keys": [
    {
      "kty": "EC",
      "use": "sig",
      "alg": "ES256",
      "kid": "my-key-2026",
      "crv": "P-256",
      "x": "oYOWCdz0ZVuUFizRWrjy0DoFdpRJz5v6Cncv_L9YSqM",
      "y": "STnkpLZCiwJ-LpZMSqxIaEWQhNVmvAbBpD_yvHAgmng"
    }
  ]
}
```

A JWKS from an existing public key (default `kid` = RFC 7638 thumbprint):

```
$ mytoolkit jwt jwks --key es256_pub.pem
{
  "keys": [
    {
      "kty": "EC",
      "use": "sig",
      "alg": "ES256",
      "kid": "PIyuyhLLgGXKwgAmfk_vPx1kQ7MtRsqhgbfMhFO1iXo",
      "crv": "P-256",
      "x": "oYOWCdz0ZVuUFizRWrjy0DoFdpRJz5v6Cncv_L9YSqM",
      "y": "STnkpLZCiwJ-LpZMSqxIaEWQhNVmvAbBpD_yvHAgmng"
    }
  ]
}
```

An HMAC secret (no public key, no JWKS):

```
$ mytoolkit jwt keygen --algorithm HS256 --kid hmac-1
{
  "algorithm": "HS256",
  "kid": "hmac-1",
  "secret": "EQOuZm77tljx9PI5cRIWeRwXy5cOLCJ4TQ_bTZpM5lQ",
  "private_jwk": {
    "kty": "oct",
    "use": "sig",
    "alg": "HS256",
    "kid": "hmac-1",
    "k": "RVFPdVptNzd0bGp4OVBJNWNSSVdlUndYeTVjT0xDSjRUUV9iVFpwTTVsUQ"
  }
}
```

### Sign with registered and custom claims

```
$ mytoolkit jwt --encode --algorithm ES256 --key es256_priv.pem --kid my-key-2026 --iss https://auth.example.com --sub user-42 --aud api --aud web --exp 1h --iat --jti --claim role=admin --claim groups=dev --claim groups=ops --claim "scopes[]=read" < /dev/null
eyJhbGciOiJFUzI1NiIsImtpZCI6Im15LWtleS0yMDI2IiwidHlwIjoiSldUIn0.eyJhdWQiOlsiYXBpIiwid2ViIl0sImV4cCI6MTc5NDgxOTI4OSwiZ3JvdXBzIjpbImRldiIsIm9wcyJdLCJpYXQiOjE3OTQ4MTU2ODksImlzcyI6Imh0dHBzOi8vYXV0aC5leGFtcGxlLmNvbSIsImp0aSI6IjdXSUlaUEVHSU5OWjJGS0kzRTVUV09GVlRUIiwicm9sZSI6ImFkbWluIiwic2NvcGVzIjpbInJlYWQiXSwic3ViIjoidXNlci00MiJ9.HJpw2FzWDaD1CGGT2iWxQo_NSUeTmLAcKLOMOt_BpRLGPaCfh0j0kC1BRW600wjPbs5FC-SyfxW6hbONt-9i8A
```

(`< /dev/null` only makes the example reproducible in a script; in an interactive shell it isn't needed.)

### Decode and verify

With the JWKS (the key is picked by the token's `kid`):

```
$ mytoolkit jwt --decode --token eyJhbGciOiJFUzI1NiIsImtpZCI6Im15LWtleS0yMDI2IiwidHlwIjoiSldUIn0.eyJhdWQiOlsiYXBpIiwid2ViIl0sImV4cCI6MTc5NDgxOTI4OSwiZ3JvdXBzIjpbImRldiIsIm9wcyJdLCJpYXQiOjE3OTQ4MTU2ODksImlzcyI6Imh0dHBzOi8vYXV0aC5leGFtcGxlLmNvbSIsImp0aSI6IjdXSUlaUEVHSU5OWjJGS0kzRTVUV09GVlRUIiwicm9sZSI6ImFkbWluIiwic2NvcGVzIjpbInJlYWQiXSwic3ViIjoidXNlci00MiJ9.HJpw2FzWDaD1CGGT2iWxQo_NSUeTmLAcKLOMOt_BpRLGPaCfh0j0kC1BRW600wjPbs5FC-SyfxW6hbONt-9i8A --key jwks.json
{"header":{"alg":"ES256","kid":"my-key-2026","typ":"JWT"},"claims":{"aud":["api","web"],"exp":1794819289,"groups":["dev","ops"],"iat":1794815689,"iss":"https://auth.example.com","jti":"7WIIZPEGINNZ2FKI3E5TWOFVTT","role":"admin","scopes":["read"],"sub":"user-42"},"signature":"HJpw2FzWDaD1CGGT2iWxQo_NSUeTmLAcKLOMOt_BpRLGPaCfh0j0kC1BRW600wjPbs5FC-SyfxW6hbONt-9i8A","valid":true}
```

`--key es256_pub.pem` (PEM) gives the same result. Once `exp` has passed, verification fails with a reason:

```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJleHAiOjE3OTQ4MTU2OTEsInN1YiI6IngifQ.tuBuqJP_wh-kCqJ_sdn-xIWSHeXwlq_zrtW5MuOLwT8 --secret s
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"exp":1794815691,"sub":"x"},"signature":"tuBuqJP_wh-kCqJ_sdn-xIWSHeXwlq_zrtW5MuOLwT8","valid":false,"verify_error":"token has invalid claims: token is expired"}
```

### HMAC

Unchanged from earlier versions:

```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls --secret mysecret
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls","valid":true}

$ echo '{"sub":"123"}' | mytoolkit jwt --encode --secret mysecret
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls
```

Base claims JSON merged with claim flags:

```
$ echo '{"tenant":"acme"}' | mytoolkit jwt --encode --secret mysecret --sub user-1 --claim role=admin
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJyb2xlIjoiYWRtaW4iLCJzdWIiOiJ1c2VyLTEiLCJ0ZW5hbnQiOiJhY21lIn0.vWZcVjLd94DBgogjW1grMDEQujm1vmIl285gwBJ2oOw
```

A wrong secret now explains itself:

```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls --secret wrong
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls","valid":false,"verify_error":"token signature is invalid: signature is invalid"}
```

### RSA keys from OpenSSL

Keys produced by other tools work as well, e.g. `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out rsa_priv.pem && openssl pkey -in rsa_priv.pem -pubout -out rsa_pub.pem`:

```
$ echo '{"sub":"1234","name":"Aecio"}' | mytoolkit jwt --encode --algorithm RS256 --key rsa_priv.pem
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiQWVjaW8iLCJzdWIiOiIxMjM0In0.SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA

$ mytoolkit jwt --decode --token eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiQWVjaW8iLCJzdWIiOiIxMjM0In0.SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA --key rsa_pub.pem
{"header":{"alg":"RS256","typ":"JWT"},"claims":{"name":"Aecio","sub":"1234"},"signature":"SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA","valid":true}
```

### Errors

Errors print a single `Error: …` line on stderr and exit with status 1.

```
$ echo '{"a":1}' | mytoolkit jwt --encode
Error: a non-empty secret is required for HS256

$ echo '{"a":1}' | mytoolkit jwt --encode --algorithm RS256
Error: invalid RSA private key: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key

$ mytoolkit jwt --encode --algorithm ES384 --key es256_priv.pem --sub x < /dev/null
Error: ES384 requires a P-384 key, got P-256

$ mytoolkit jwt --encode --secret s --claim exp=123 < /dev/null
Error: "exp" must be a numeric date: use the expires_in/not_before/issued_at options (CLI: --exp/--nbf/--iat) instead of a custom claim

$ mytoolkit jwt --encode --secret s --claims claims.json   # claims.json: {"sub":"a","sub":"b"}
Error: invalid claims JSON: jsontext: duplicate object member name "sub"
```

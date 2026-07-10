<!-- TOC -->

- [JWT Encode/Decode — REST API](#jwt-encodedecode--rest-api)
  - [Decode request](#decode-request)
  - [Decode success (200)](#decode-success-200)
  - [Encode request](#encode-request)
  - [Encode success (200)](#encode-success-200)
  - [Error response (400)](#error-response-400)
  - [Workflow](#workflow)

<!-- TOC -->

# JWT Encode/Decode — REST API

`POST /api/v1/tools/jwt`

## Decode request

```json
{ "input": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls", "options": { "mode": "decode", "secret": "mysecret" } }
```

## Decode success (200)

```json
{
  "success": true,
  "data": { "header": {"alg":"HS256","typ":"JWT"}, "claims": {"sub":"123"}, "signature": "9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls", "valid": true },
  "meta": { "tool": "jwt", "duration_ms": 0.06 }
}
```

`secret` is optional for decode: omitted → inspection only (`valid` absent); supplied → `valid` reflects signature verification against that secret. `secret` only applies to HMAC-family tokens (`HS256`/`HS384`/`HS512`); for every other algorithm, verification uses `key` instead (see below).

## Encode request

```json
{ "input": "{\"sub\":\"123\"}", "options": { "mode": "encode", "secret": "mysecret", "algorithm": "HS256" } }
```

`options.algorithm` (default `HS256`, unchanged from before): one of `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `PS256`, `PS384`, `PS512`, `ES256`, `ES384`, `ES512`, `EdDSA`.

`options.secret`: the shared secret, used only for `HS256`/`HS384`/`HS512`.

`options.key`: a PEM-encoded key, used for every other algorithm — a **private** key when encoding, a **public** key when decoding with verification. Unused (and not required) for `HS*`.

## Encode success (200)

```json
{
  "success": true,
  "data": { "output": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls" },
  "meta": { "tool": "jwt", "duration_ms": 0.06 }
}
```

Encoding with an asymmetric algorithm — request:
```json
{ "input": "{\"sub\":\"1234\"}", "options": { "mode": "encode", "algorithm": "RS256", "key": "-----BEGIN PRIVATE KEY-----\n...\n-----END PRIVATE KEY-----\n" } }
```

Response (real 2048-bit RSA key, verified against the running binary):
```json
{
  "success": true,
  "data": { "output": "eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0In0.rX0OB-uJ-H9c1Iza0w1MxTg4Pkxd4JEkTv8q5i046-5UiWFD6fjh6FbJDCQfhNfPYLTH8B1beuFvqKyx_Q10Fb6IJ4SVwbeA0RP9t2tWkIKo-FLfmUpceOVoNvNIhhHEZLmZEHi_8KpJLV4uujJEh_fQ8HW-fW_IZW--gF_F705cMV-9u73jhqOz5wrtxDlROxXA1ToS4diEU_iUEMbmVUJLVKD1MwDBkZCLCrUoSuaapdophhs1dj5IBCEMTrGmc3FL4tZIiuqGVe7VfC2E00S0ikxhbWOshnaDeWa-TZNMaN3BOMyH5rc5fkmileRx07DMNjPEreYJ8ZjbXJ_okg" },
  "meta": { "tool": "jwt", "duration_ms": 0.85 }
}
```

## Error response (400)

```json
{ "success": false, "error": { "code": "INVALID_TOKEN", "message": "token contains an invalid number of segments" } }
```

Error codes: `INVALID_TOKEN`, `UNSUPPORTED_ALGORITHM`, `INVALID_KEY` (malformed or wrong-type PEM key on encode), `EMPTY_CLAIMS`.

Secrets and keys are used in-memory for the single request only and are never persisted server-side.

## Workflow

```mermaid
sequenceDiagram
    participant C as Client (Web/REST/CLI)
    participant H as jwtHandler (bespoke)
    participant T as internal/tools/jwttool

    C->>H: POST /api/v1/tools/jwt {input, options.mode}
    alt mode == encode
        H->>T: ParseClaimsJSON(input) + Encode(claims, secret, key, algorithm)
        T->>T: pick HMAC secret or parse PEM private key by algorithm family
        T-->>H: signed token
        H-->>C: 200 {data.output}
    else mode == decode (default)
        H->>T: Decode(token, secret, key)
        T->>T: pick HMAC secret or parse PEM public key by the token's own alg header
        T-->>H: header, claims, signature, valid?
        H-->>C: 200 {data: header/claims/signature/valid}
    end
    T--xH: apperr.Error on malformed token / empty claims / unsupported algorithm / invalid key
    H-->>C: 4xx {success:false, error}
```

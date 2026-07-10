<!-- TOC -->

- [JWT Encode/Decode — REST API](#jwt-encodedecode--rest-api)
  - [Decode request](#decode-request)
  - [Decode success (200)](#decode-success-200)
  - [Encode request](#encode-request)
  - [Encode success (200)](#encode-success-200)
  - [Error response (400)](#error-response-400)

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

`secret` is optional for decode: omitted → inspection only (`valid` absent); supplied → `valid` reflects signature verification against that secret.

## Encode request

```json
{ "input": "{\"sub\":\"123\"}", "options": { "mode": "encode", "secret": "mysecret", "algorithm": "HS256" } }
```

## Encode success (200)

```json
{
  "success": true,
  "data": { "output": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls" },
  "meta": { "tool": "jwt", "duration_ms": 0.06 }
}
```

## Error response (400)

```json
{ "success": false, "error": { "code": "INVALID_TOKEN", "message": "token contains an invalid number of segments" } }
```

Error codes: `INVALID_TOKEN`, `UNSUPPORTED_ALGORITHM` (encode only supports `HS256`/`HS384`/`HS512`), `EMPTY_CLAIMS`.

Secrets are used in-memory for the single request only and are never persisted server-side.

## Workflow

```mermaid
sequenceDiagram
    participant C as Client (Web/REST/CLI)
    participant H as jwtHandler (bespoke)
    participant T as internal/tools/jwttool

    C->>H: POST /api/v1/tools/jwt {input, options.mode}
    alt mode == encode
        H->>T: ParseClaimsJSON(input) + Encode(claims, secret, algorithm)
        T-->>H: signed token
        H-->>C: 200 {data.output}
    else mode == decode (default)
        H->>T: Decode(token, secret)
        T-->>H: header, claims, signature, valid?
        H-->>C: 200 {data: header/claims/signature/valid}
    end
    T--xH: apperr.Error on malformed token / empty claims / unsupported algorithm
    H-->>C: 4xx {success:false, error}
```

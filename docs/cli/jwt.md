<!-- TOC -->

- [JWT Encode/Decode — CLI](#jwt-encodedecode--cli)
  - [Examples](#examples)

<!-- TOC -->

# JWT Encode/Decode — CLI

```
mytoolkit jwt --decode --token <token> [--secret <secret>] [--key <file|->] [--out <file|->]
mytoolkit jwt --encode --claims <file|-> [--secret <secret>] [--key <file|->] [--algorithm HS256] [--out <file|->]
```

`--algorithm` (default `HS256`, unchanged): one of `HS256`, `HS384`, `HS512`, `RS256`, `RS384`, `RS512`, `PS256`, `PS384`, `PS512`, `ES256`, `ES384`, `ES512`, `EdDSA`.

`--secret` is a raw shared secret, used only by `HS256`/`HS384`/`HS512`. `--key` is a path to a PEM-encoded key file (or `-` for stdin), used by every other algorithm — a **private** key for `--encode`, a **public** key for `--decode` verification.

## Examples

HMAC (unchanged from before this tool supported other algorithm families):
```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls --secret mysecret
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls","valid":true}

$ echo '{"sub":"123"}' | mytoolkit jwt --encode --secret mysecret
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls
```

RSA (`RS256`/`RS384`/`RS512`/`PS256`/`PS384`/`PS512`), using a real key pair generated with `openssl genpkey -algorithm RSA -pkeyopt rsa_keygen_bits:2048 -out rsa_priv.pem && openssl pkey -in rsa_priv.pem -pubout -out rsa_pub.pem`:
```
$ echo '{"sub":"1234","name":"Aecio"}' | mytoolkit jwt --encode --algorithm RS256 --key rsa_priv.pem
eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiQWVjaW8iLCJzdWIiOiIxMjM0In0.SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA

$ mytoolkit jwt --decode --token eyJhbGciOiJSUzI1NiIsInR5cCI6IkpXVCJ9.eyJuYW1lIjoiQWVjaW8iLCJzdWIiOiIxMjM0In0.SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA --key rsa_pub.pem
{"header":{"alg":"RS256","typ":"JWT"},"claims":{"name":"Aecio","sub":"1234"},"signature":"SCq3ovM-iWkB3XbXL6Gq2L80VUuMU2AsoZ3Gc1dP1lIduPPJBbLanmUMX2ZmJpJZ6ki9N0cQwtDlBb89ZhhAapSWrs16dghhwsljcSPdIPpXr1sSZ4oct9iFF1DgINNMGK9FEX3qGCwJPoC99bUfof1xdlT718YLC8VK_SOg780M1ry7hvDQ1ctXL4RQbHN6FpbBWigTdwt6L9EynfKBJgIPCPvihjMYGsLZbGCL7_5y7AjnsAmxfk4NOykdtiZskDBKh3CLjxMWGjBvCfebmyKEzzREoLmUMl7AcFoaDQ5wu5RXVCXQLfDuJ_XyY3eFLIBPOZFfPsme1Bka1gZxiA","valid":true}
```

Encoding with `RS*`/`PS*`/`ES*`/`EdDSA` without `--key` fails clearly instead of silently falling back to HMAC:
```
$ echo '{"a":1}' | mytoolkit jwt --encode --algorithm RS256
Error: invalid RSA private key: invalid key: Key must be a PEM encoded PKCS1 or PKCS8 key
```

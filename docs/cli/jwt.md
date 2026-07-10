<!-- TOC -->

- [JWT Encode/Decode — CLI](#jwt-encodedecode--cli)
  - [Examples](#examples)

<!-- TOC -->

# JWT Encode/Decode — CLI

```
mytoolkit jwt --decode --token <token> [--secret <secret>] [--out <file|->]
mytoolkit jwt --encode --claims <file|-> --secret <secret> [--algorithm HS256] [--out <file|->]
```

## Examples

```
$ mytoolkit jwt --decode --token eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls --secret mysecret
{"header":{"alg":"HS256","typ":"JWT"},"claims":{"sub":"123"},"signature":"9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls","valid":true}

$ echo '{"sub":"123"}' | mytoolkit jwt --encode --secret mysecret
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjMifQ.9hTwgEDMPX_PVRr1ke0l2cO2goPzH7j40OL5pSxUzls
```

<!-- TOC -->

- [JWT Encode/Decode — Testing](#jwt-encodedecode--testing)

<!-- TOC -->

# JWT Encode/Decode — Testing

```
$ cd src && go test ./internal/tools/jwttool/... -v
--- PASS: TestEncodeDecodeRoundTrip (0.00s)
--- PASS: TestDecodeWrongSecret (0.00s)
--- PASS: TestDecodeWithoutSecret (0.00s)
--- PASS: TestDecodeMalformedToken (0.00s)
--- PASS: TestEncodeEmptyClaims (0.00s)
--- PASS: TestEncodeUnsupportedAlgorithm (0.00s)
PASS
```

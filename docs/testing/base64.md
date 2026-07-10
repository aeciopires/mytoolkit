<!-- TOC -->

- [Base64 Encode/Decode — Testing](#base64-encodedecode--testing)

<!-- TOC -->

# Base64 Encode/Decode — Testing

```
$ cd src && go test ./internal/tools/base64enc/... -v
--- PASS: TestProcess (0.00s)
    --- PASS: TestProcess/encode_standard
    --- PASS: TestProcess/decode_standard
    --- PASS: TestProcess/encode_no_padding
    --- PASS: TestProcess/decode_no_padding
    --- PASS: TestProcess/encode_url_variant
    --- PASS: TestProcess/empty_input_encode
    --- PASS: TestProcess/empty_input_decode
    --- PASS: TestProcess/invalid_base64
    --- PASS: TestProcess/invalid_variant
--- PASS: TestRoundTrip (0.00s)
PASS
```

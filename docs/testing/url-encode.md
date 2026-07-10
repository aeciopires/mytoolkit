<!-- TOC -->

- [URL Encode/Decode — Testing](#url-encodedecode--testing)

<!-- TOC -->

# URL Encode/Decode — Testing

```
$ cd src && go test ./internal/tools/urlencode/... -v
--- PASS: TestProcess (0.00s)
    --- PASS: TestProcess/encode_query_spaces
    --- PASS: TestProcess/decode_query_spaces
    --- PASS: TestProcess/encode_path_spaces
    --- PASS: TestProcess/decode_path_spaces
    --- PASS: TestProcess/empty_input
    --- PASS: TestProcess/invalid_component
    --- PASS: TestProcess/invalid_encoding_on_decode
    --- PASS: TestProcess/unicode_round_trip_encode
PASS
```

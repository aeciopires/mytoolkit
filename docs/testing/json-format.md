<!-- TOC -->

- [JSON Formatter — Testing](#json-formatter--testing)

<!-- TOC -->

# JSON Formatter — Testing

```
$ cd src && go test ./internal/tools/jsonformat/... -v
=== RUN   TestFormat
--- PASS: TestFormat (0.00s)
    --- PASS: TestFormat/pretty_default_indent
    --- PASS: TestFormat/pretty_custom_indent
    --- PASS: TestFormat/minify
    --- PASS: TestFormat/minify_idempotent
    --- PASS: TestFormat/empty_input
    --- PASS: TestFormat/malformed_json
PASS
```

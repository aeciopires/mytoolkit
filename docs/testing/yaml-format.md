<!-- TOC -->

- [YAML Formatter — Testing](#yaml-formatter--testing)

<!-- TOC -->

# YAML Formatter — Testing

```
$ cd src && go test ./internal/tools/yamlformat/... -v
=== RUN   TestFormat
--- PASS: TestFormat (0.00s)
    --- PASS: TestFormat/reindent_list
    --- PASS: TestFormat/empty_input
    --- PASS: TestFormat/malformed_yaml
=== RUN   TestFormatIdempotent
--- PASS: TestFormatIdempotent (0.00s)
PASS
```

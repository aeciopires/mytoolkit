<!-- TOC -->

- [JSON to YAML Converter — Testing](#json-to-yaml-converter--testing)

<!-- TOC -->

# JSON to YAML Converter — Testing

```
$ cd app && go test ./internal/tools/jsontoyaml/... -v
=== RUN   TestConvert
--- PASS: TestConvert (0.00s)
    --- PASS: TestConvert/flat_object
    --- PASS: TestConvert/nested_object_and_array
    --- PASS: TestConvert/empty_input
    --- PASS: TestConvert/malformed_json:_trailing_comma
    --- PASS: TestConvert/malformed_json:_unquoted_key
    --- PASS: TestConvert/malformed_json:_single-quoted_string
=== RUN   TestConvertQuotesStringsThatWouldResolveAsOtherTypes
--- PASS: TestConvertQuotesStringsThatWouldResolveAsOtherTypes (0.00s)
=== RUN   TestConvertLargeIntegerPreserved
--- PASS: TestConvertLargeIntegerPreserved (0.00s)
PASS
```

Covers: flat/nested conversion, empty input, three flavors of invalid JSON that `sigs.k8s.io/yaml.JSONToYAML` alone would silently accept (trailing comma, unquoted key, single-quoted string — all confirmed to actually fail, verifying the `encoding/json` pre-validation step in `Convert` is load-bearing), automatic quoting of strings that would otherwise resolve as a different YAML type (`"NO"`, `"y"`), and exact preservation of a large (18-digit) integer.

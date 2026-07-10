<!-- TOC -->

- [YAML to JSON Converter — Testing](#yaml-to-json-converter--testing)

<!-- TOC -->

# YAML to JSON Converter — Testing

```
$ cd app && go test ./internal/tools/yamltojson/... -v
=== RUN   TestConvert
--- PASS: TestConvert (0.00s)
    --- PASS: TestConvert/flat_mapping
    --- PASS: TestConvert/nested_mapping_and_sequence
    --- PASS: TestConvert/custom_indent
    --- PASS: TestConvert/empty_input
    --- PASS: TestConvert/malformed_yaml
    --- PASS: TestConvert/tab_indentation_rejected
    --- PASS: TestConvert/duplicate_keys_rejected
=== RUN   TestConvertYAML11BooleanResolution
--- PASS: TestConvertYAML11BooleanResolution (0.00s)
=== RUN   TestConvertLargeIntegerPreserved
--- PASS: TestConvertLargeIntegerPreserved (0.00s)
PASS
```

Covers: flat/nested conversion, custom indent, empty input, malformed YAML, tab-character indentation, duplicate mapping keys (rejected via `YAMLToJSONStrict`), the YAML 1.1 boolean-resolution behavior (`NO`→`false`, `y`→`true` — verified real library behavior, not assumed), and exact preservation of a large (18-digit) integer.

<!-- TOC -->

- [JSON Tree Viewer — Testing](#json-tree-viewer--testing)

<!-- TOC -->

# JSON Tree Viewer — Testing

```
$ cd src && go test ./internal/tools/jsontree/... -v
=== RUN   TestParseFlatObject
--- PASS: TestParseFlatObject (0.00s)
=== RUN   TestParseNested
--- PASS: TestParseNested (0.00s)
=== RUN   TestParseEmptyInput
--- PASS: TestParseEmptyInput (0.00s)
=== RUN   TestParseMalformed
--- PASS: TestParseMalformed (0.00s)
=== RUN   TestParseLargeNumberPreserved
--- PASS: TestParseLargeNumberPreserved (0.00s)
=== RUN   TestParseUnicodeString
--- PASS: TestParseUnicodeString (0.00s)
PASS
```

Covers: flat/nested objects with key-order preservation, empty input, malformed JSON, large integers (via `json.Number`), and unicode strings.

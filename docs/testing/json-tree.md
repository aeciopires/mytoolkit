<!-- TOC -->

- [JSON Tree Viewer — Testing](#json-tree-viewer--testing)

<!-- TOC -->

# JSON Tree Viewer — Testing

```
$ cd app && go test ./internal/tools/jsontree/... -v
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
=== RUN   TestParseErrorIncludesPosition
--- PASS: TestParseErrorIncludesPosition (0.00s)
=== RUN   TestParseErrorPositionOnLaterLine
--- PASS: TestParseErrorPositionOnLaterLine (0.00s)
=== RUN   TestParseRejectsTrailingData
--- PASS: TestParseRejectsTrailingData (0.00s)
=== RUN   TestParseRejectsTrailingGarbage
--- PASS: TestParseRejectsTrailingGarbage (0.00s)
PASS
```

Covers: flat/nested objects with key-order preservation, empty input, malformed JSON, large integers (via `json.Number`), unicode strings, error messages carrying accurate line/column position (including on a line other than the first), and rejection of trailing data/garbage after a complete JSON value.

## Web UI verification (manual/scripted, no Go test coverage)

The web page's tree rendering, Expand All/Collapse All, and Generate-on-click (not on-keystroke) behavior are frontend-only (`internal/web/templates/tools/json-tree.html`) and aren't exercised by `go test`. Verified instead with a real browser (Playwright driving the actual binary):

- Typing in the input textarea triggers **zero** `POST /api/v1/tools/json-tree` calls (confirms the tool only calls the server when "Generate Tree View" is clicked, not on every keystroke).
- Clicking "Generate Tree View" triggers exactly one call and renders the tree.
- "Collapse All" / "Expand All" toggle every `<details class="json-node">` in the output.
- "Clear" empties both the input and the tree output.
- An invalid-JSON input shows the exact backend error message (including line/column) in the error banner.

Re-run this check whenever `json-tree.html` changes.

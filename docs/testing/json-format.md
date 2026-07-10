<!-- TOC -->

- [JSON Formatter — Testing](#json-formatter--testing)

<!-- TOC -->

# JSON Formatter — Testing

```
$ cd app && go test ./internal/tools/jsonformat/... -v
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

Covers REST/CLI usage of `jsonformat.Format` (pretty/minify, custom indent, empty input, malformed JSON).

## Web UI verification (manual/scripted, no Go test coverage)

The web page (`internal/web/templates/tools/json-format.html`) runs entirely client-side via native `JSON.parse()`/`JSON.stringify()` and never calls `Format` or the REST endpoint, so it isn't exercised by `go test`. Verified instead with a real browser (Playwright driving the actual binary):

- Typing in the input textarea triggers **zero** `POST /api/v1/tools/json-format` calls (confirms Validate/Beautify/Minify never hit the network).
- "Validate JSON" on valid input shows the green success banner ("Valid JSON.") and leaves the output field untouched.
- "Validate JSON" on invalid input (e.g. `{"a":}`) shows the browser's own `JSON.parse()` error message in the red error banner.
- "Beautify" on valid input fills the output with 2-space-indented JSON; on invalid input, clears the output and shows the error banner.
- "Minify" on valid input fills the output with whitespace-free JSON.
- "Copy Output" copies the output field to the clipboard.
- "Clear" empties both fields and hides any banner.

Re-run this check whenever `json-format.html` changes.

<!-- TOC -->

- [YAML Formatter — Testing](#yaml-formatter--testing)

<!-- TOC -->

# YAML Formatter — Testing

```
$ cd app && go test ./internal/tools/yamlformat/... -v
=== RUN   TestFormat
--- PASS: TestFormat (0.00s)
    --- PASS: TestFormat/reindent_list
    --- PASS: TestFormat/empty_input
    --- PASS: TestFormat/malformed_yaml
    --- PASS: TestFormat/tab_indentation_rejected
    --- PASS: TestFormat/invalid_style_option
    --- PASS: TestFormat/whitespace-only_stream
    --- PASS: TestFormat/multi-document_stream_reformatted
    --- PASS: TestFormat/flow_style_forces_compact_single-line_collections
    --- PASS: TestFormat/block_style_normalizes_mixed_flow_input
    --- PASS: TestFormat/comments_are_preserved
    --- PASS: TestFormat/anchors_and_aliases_are_preserved
=== RUN   TestFormatIdempotent
--- PASS: TestFormatIdempotent (0.00s)
PASS
```

Covers: reindentation, empty/whitespace-only input, malformed YAML (bad indentation, tab characters), invalid `style` option, multi-document streams (every document reformatted, `---` separators preserved), block/flow style normalization in both directions, comment preservation, anchor/alias preservation, and idempotency.

## Web UI verification (manual/scripted)

The new "Style" select was verified with a real browser (Playwright driving the actual binary), which also caught a real, pre-existing bug: `tool-common.js` sent every `<select>` value as a JSON string. `Options.Style` (a Go `string`) tolerated that, but `Options.Indent` (a Go `int`) didn't — `json.Unmarshal` rejected it with `invalid options object`, so the indent selector had never worked from the web page at all (same bug also affected QR Code's `size` selector). Fixed in `tool-common.js`; verified after the fix:
- Selecting "Flow" reformats the output to compact single-line `{}`/`[]` notation.
- Selecting "Block" reformats mixed-style input back to consistent indentation.
- Pasting a `---`-separated multi-document stream reformats and preserves every document.

Re-run this check whenever `yaml-format.html` or `tool-common.js` changes.

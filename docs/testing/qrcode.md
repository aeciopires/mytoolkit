<!-- TOC -->

- [QR Code Generator — Testing](#qr-code-generator--testing)

<!-- TOC -->

# QR Code Generator — Testing

```
$ cd app && go test ./internal/tools/qrcode/... -v
--- PASS: TestGenerateValidText (0.00s)
--- PASS: TestGenerateUnicodeText (0.00s)
--- PASS: TestGenerateEmptyText (0.00s)
--- PASS: TestGenerateTooLarge (0.00s)
--- PASS: TestGenerateDifferentSizes (0.00s)
PASS
```

## Web UI note

The "Size" `<select>` on this page (`options.size`, a Go `int`) was silently broken until `tool-common.js`'s `collectOptions()` was fixed to send numeric-looking `<select>` values as JSON numbers instead of strings — every size change previously 400'd with `invalid options object`. Found while browser-testing an unrelated YAML Formatter change; re-verified here by selecting each size option and confirming a 200 response. See `docs/testing/yaml-format.md` for the full writeup.

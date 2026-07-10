<!-- TOC -->

- [QR Code Generator — Testing](#qr-code-generator--testing)

<!-- TOC -->

# QR Code Generator — Testing

```
$ cd src && go test ./internal/tools/qrcode/... -v
--- PASS: TestGenerateValidText (0.00s)
--- PASS: TestGenerateUnicodeText (0.00s)
--- PASS: TestGenerateEmptyText (0.00s)
--- PASS: TestGenerateTooLarge (0.00s)
--- PASS: TestGenerateDifferentSizes (0.00s)
PASS
```

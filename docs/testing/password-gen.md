<!-- TOC -->

- [Password Generator — Testing](#password-generator--testing)

<!-- TOC -->

# Password Generator — Testing

```
$ cd src && go test ./internal/tools/password/... -v
--- PASS: TestGenerateLength (0.00s)
--- PASS: TestGenerateOnlyEnabledClasses (0.00s)
--- PASS: TestGenerateNoCharsetSelected (0.00s)
--- PASS: TestGenerateInvalidLength (0.00s)
--- PASS: TestGenerateExcludeConfusing (0.00s)
--- PASS: TestGenerateExcludeAmbiguous (0.00s)
--- PASS: TestGenerateProducesDifferentOutputs (0.00s)
PASS
```

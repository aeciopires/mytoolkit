<!-- TOC -->

- [Case Converter — Testing](#case-converter--testing)

<!-- TOC -->

# Case Converter — Testing

```
$ cd src && go test ./internal/tools/caseconvert/... -v
--- PASS: TestConvert (0.00s)
    --- PASS: TestConvert/sentence_basic
    --- PASS: TestConvert/sentence_consecutive_terminators
    --- PASS: TestConvert/upper
    --- PASS: TestConvert/lower
    --- PASS: TestConvert/title
    --- PASS: TestConvert/title_collapses_whitespace
    --- PASS: TestConvert/mixed_pattern
    --- PASS: TestConvert/inverse_basic
    --- PASS: TestConvert/inverse_round_trip
    --- PASS: TestConvert/empty_input
    --- PASS: TestConvert/unsupported_mode
    --- PASS: TestConvert/unicode_upper
--- PASS: TestInverseIsSelfInverting (0.00s)
PASS
```

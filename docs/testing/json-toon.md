<!-- TOC -->

- [JSON to TOON Converter — Testing](#json-to-toon-converter--testing)
  - [Go implementation](#go-implementation)
  - [JavaScript implementation (client-side mirror)](#javascript-implementation-client-side-mirror)
  - [Zero-network verification (web page)](#zero-network-verification-web-page)

<!-- TOC -->

# JSON to TOON Converter — Testing

## Go implementation

```
$ cd app && go test ./internal/tools/jsontoon/... -v
--- PASS: TestConvertTabularArray (0.00s)
--- PASS: TestConvertSimpleObject (0.00s)
--- PASS: TestConvertArrayOfArrays (0.00s)
--- PASS: TestConvertNonUniformObjectArray (0.00s)
--- PASS: TestStringQuoting (0.00s)
    --- PASS: TestStringQuoting/empty_string
    --- PASS: TestStringQuoting/leading/trailing_whitespace
    --- PASS: TestStringQuoting/literal_true
    --- PASS: TestStringQuoting/literal_false
    --- PASS: TestStringQuoting/literal_null
    --- PASS: TestStringQuoting/numeric-looking
    --- PASS: TestStringQuoting/contains_colon
    --- PASS: TestStringQuoting/contains_bracket
    --- PASS: TestStringQuoting/starts_with_hyphen
    --- PASS: TestStringQuoting/plain_word_not_quoted
--- PASS: TestNumberCanonicalization (0.00s)
--- PASS: TestDelimiterOptions (0.00s)
--- PASS: TestIndentOption (0.00s)
--- PASS: TestConvertErrors (0.00s)
--- PASS: TestConvertRootScalar (0.00s)
PASS
```

## JavaScript implementation (client-side mirror)

This project has no Node/npm/Jest tooling by design (see `PLAN_ARCHITECTURE.md`'s "no Node build step" decision), so `internal/web/static/js/json-toon.js` is **not** covered by an automated CI test suite. Instead it is verified by running the exact same fixture table used by `jsontoon_test.go` (above) through the real file in headless Chrome, and diffing against the same expected values:

```bash
# Minimal harness: loads json-toon.js, calls window.jsonToToon() with each
# Go-test fixture, writes PASS/FAIL per case into the DOM.
google-chrome --headless=new --disable-gpu --no-sandbox --dump-dom \
  file:///path/to/toon-parity.html
```

Last run: **20/20 cases passed**, exact string equality with the Go output for every fixture (tabular array, simple object, array-of-arrays fallback, all 9 string-quoting cases, number canonicalization, `tab`/`pipe` delimiters, custom indent, and all 4 root-scalar cases).

**Run this check whenever `json-toon.js` changes** — there is no CI gate enforcing it, so it's a manual/scripted step, not an automated safety net. If you change the encoding rules in `internal/tools/jsontoon/jsontoon.go`, mirror the change in `internal/web/static/js/json-toon.js` and re-run this parity check before committing (see `.skills/json-toon/SKILL.md`).

## Zero-network verification (web page)

Confirmed two ways:
1. **Source inspection**: `internal/web/static/js/json-toon.js` contains no `fetch`, `XMLHttpRequest`, `WebSocket`, or `sendBeacon` calls — `grep -n "fetch\|XMLHttpRequest\|WebSocket" internal/web/static/js/json-toon.js` (from `app/`) returns nothing.
2. **Runtime inspection**: load `/tools/json-toon` in a real browser, open DevTools → Network, type/paste JSON into the input — no request to `/api/v1/tools/json-toon` (or anywhere else) appears.

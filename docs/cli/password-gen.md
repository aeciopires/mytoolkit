<!-- TOC -->

- [Password Generator — CLI](#password-generator--cli)
  - [Examples](#examples)

<!-- TOC -->

# Password Generator — CLI

```
mytoolkit password-gen [--length N] [--lowercase] [--uppercase] [--numbers] [--symbols] [--exclude-confusing] [--exclude-ambiguous] [--out <file|->]
```

Lowercase/uppercase/numbers default to `true`; symbols defaults to `false`. Passing any of these flags overrides its default.

## Examples

```
$ mytoolkit password-gen --length 20 --symbols
ceCwwxDf^R9vHClG-9JE

$ mytoolkit password-gen --length 20 --symbols --exclude-confusing --exclude-ambiguous
```

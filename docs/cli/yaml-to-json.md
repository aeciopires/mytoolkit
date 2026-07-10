<!-- TOC -->

- [YAML to JSON Converter — CLI](#yaml-to-json-converter--cli)
  - [Examples](#examples)

<!-- TOC -->

# YAML to JSON Converter — CLI

```
mytoolkit yaml-to-json --in <file|-> [--out <file|->] [--indent N]
```

## Examples

```
$ printf 'a: 1\nb:\n  - x\n  - y\n' | mytoolkit yaml-to-json
{
  "a": 1,
  "b": [
    "x",
    true
  ]
}
```

Note `"y"` converting to `true` — unquoted `y`/`n`/`yes`/`no`/`on`/`off` resolve as booleans per YAML 1.1 (the "Norway problem"), not a bug in this tool.

```
$ printf 'a: 1\n' | mytoolkit yaml-to-json --indent 4
{
    "a": 1
}
```

Errors report the underlying parser's line number:

```
$ printf 'a: [1, 2\n' | mytoolkit yaml-to-json
Error: yaml: line 1: did not find expected ',' or ']'
```

Duplicate mapping keys are rejected:

```
$ printf 'a: 1\na: 2\n' | mytoolkit yaml-to-json
Error: yaml: unmarshal errors:
  line 2: key "a" already set in map
```

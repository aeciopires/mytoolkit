<!-- TOC -->

- [YAML Formatter — CLI](#yaml-formatter--cli)
  - [Examples](#examples)

<!-- TOC -->

# YAML Formatter — CLI

```
mytoolkit yaml-format --in <file|-> [--out <file|->] [--indent N] [--style block|flow]
```

## Examples

```
$ printf 'a: 1\nb:\n    - x\n    - y\n' | mytoolkit yaml-format --indent 2
a: 1
b:
  - x
  - y
```

Multi-document streams are fully reformatted, not just the first document:

```
$ printf 'a: 1\n---\nb: 2\n' | mytoolkit yaml-format
a: 1
---
b: 2
```

`--style flow` collapses collections to compact `{}`/`[]` notation on one line (the YAML equivalent of minifying):

```
$ printf 'a:\n  b: 1\n  c:\n    - 1\n    - 2\n' | mytoolkit yaml-format --style flow
{a: {b: 1, c: [1, 2]}}
```

Errors report the underlying parser's line number:

```
$ printf 'a: 1\n  b: 2\n' | mytoolkit yaml-format
Error: yaml: line 2: mapping values are not allowed in this context
```

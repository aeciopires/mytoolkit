<!-- TOC -->

- [JSON Formatter — CLI](#json-formatter--cli)
  - [Examples](#examples)

<!-- TOC -->

# JSON Formatter — CLI

```
mytoolkit json-format --in <file|-> [--out <file|->] [--minify] [--indent N]
```

## Examples

```
$ echo '{"a":1,"b":2}' | mytoolkit json-format
{
  "a": 1,
  "b": 2
}

$ echo '{"a": 1, "b": 2}' | mytoolkit json-format --minify
{"a":1,"b":2}
```

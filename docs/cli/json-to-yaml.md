<!-- TOC -->

- [JSON to YAML Converter — CLI](#json-to-yaml-converter--cli)
  - [Examples](#examples)

<!-- TOC -->

# JSON to YAML Converter — CLI

```
mytoolkit json-to-yaml --in <file|-> [--out <file|->]
```

## Examples

```
$ printf '{"a":1,"b":["x","y"]}' | mytoolkit json-to-yaml
a: 1
b:
- x
- "y"
```

Note `"y"` stays quoted in the output — see `docs/api/json-to-yaml.md` for why.

Errors report the underlying `encoding/json` message:

```
$ printf '{"a":1,}' | mytoolkit json-to-yaml
Error: invalid character '}' looking for beginning of object key string
```

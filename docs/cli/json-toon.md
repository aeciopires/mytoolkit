<!-- TOC -->

- [JSON to TOON Converter — CLI](#json-to-toon-converter--cli)
  - [Example](#example)

<!-- TOC -->

# JSON to TOON Converter — CLI

```
mytoolkit json-toon --in <file|-> [--out <file|->] [--delimiter comma|tab|pipe] [--indent N]
```

## Example

```
$ echo '{"users":[{"id":1,"name":"Alice","role":"admin"},{"id":2,"name":"Bob","role":"user"}]}' | mytoolkit json-toon
users[2]{id,name,role}:
  1,Alice,admin
  2,Bob,user
```

```
$ echo '{"id":123,"name":"Ada","active":true}' | mytoolkit json-toon
id: 123
name: Ada
active: true
```

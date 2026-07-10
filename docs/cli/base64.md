<!-- TOC -->

- [Base64 Encode/Decode — CLI](#base64-encodedecode--cli)
  - [Example](#example)

<!-- TOC -->

# Base64 Encode/Decode — CLI

```
mytoolkit base64 [--decode] [--variant standard|url] [--no-padding] --in <file|-> [--out <file|->]
```

## Example

```
$ echo -n 'hello world' | mytoolkit base64
aGVsbG8gd29ybGQ=

$ echo -n 'aGVsbG8gd29ybGQ=' | mytoolkit base64 --decode
hello world
```

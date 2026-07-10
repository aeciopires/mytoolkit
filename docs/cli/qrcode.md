<!-- TOC -->

- [QR Code Generator — CLI](#qr-code-generator--cli)
  - [Example](#example)

<!-- TOC -->

# QR Code Generator — CLI

```
mytoolkit qrcode --text <string> [--size N] --out <file>
```

`--out` is **required** (binary PNG output cannot go to stdout).

## Example

```
$ mytoolkit qrcode --text "https://example.com" --size 256 --out qr.png
$ mytoolkit qrcode --text "hello"
Error: --out <file> is required (binary PNG output cannot go to stdout)
```

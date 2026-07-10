<!-- TOC -->

- [Hash Generator — CLI](#hash-generator--cli)
  - [Example](#example)

<!-- TOC -->

# Hash Generator — CLI

```
mytoolkit hash-gen --algo md5|sha1|sha256|sha512 --in <file|-> [--out <file|->]
```

## Example

```
$ echo -n 'hello' | mytoolkit hash-gen --algo sha256
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
```

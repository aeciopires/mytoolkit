<!-- TOC -->

- [YAML Formatter — CLI](#yaml-formatter--cli)
  - [Example](#example)

<!-- TOC -->

# YAML Formatter — CLI

```
mytoolkit yaml-format --in <file|-> [--out <file|->] [--indent N]
```

## Example

```
$ printf 'a: 1\nb:\n    - x\n    - y\n' | mytoolkit yaml-format --indent 2
a: 1
b:
  - x
  - y
```

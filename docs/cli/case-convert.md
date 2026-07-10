<!-- TOC -->

- [Case Converter — CLI](#case-converter--cli)
  - [Example](#example)

<!-- TOC -->

# Case Converter — CLI

```
mytoolkit case-convert --mode sentence|upper|lower|title|mixed|inverse --in <file|-> [--out <file|->]
```

## Example

```
$ echo 'hello WORLD example' | mytoolkit case-convert --mode title
Hello World Example

$ echo 'hello world' | mytoolkit case-convert --mode mixed
HeLlO WoRlD
```

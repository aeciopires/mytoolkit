<!-- TOC -->

- [Character, Word \& Line Counter — CLI](#character-word--line-counter--cli)
  - [Example](#example)

<!-- TOC -->

# Character, Word & Line Counter — CLI

```
mytoolkit text-count --in <file|->
```

## Example

```
$ printf 'Hello world\nSecond line\n' | mytoolkit text-count
characters: 24
characters_no_spaces: 20
words: 4
lines: 2
```

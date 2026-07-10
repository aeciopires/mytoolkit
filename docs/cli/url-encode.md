<!-- TOC -->

- [URL Encode/Decode — CLI](#url-encodedecode--cli)
  - [Example](#example)

<!-- TOC -->

# URL Encode/Decode — CLI

```
mytoolkit url-encode [--decode] [--component query|path|full] --in <file|-> [--out <file|->]
```

## Example

```
$ echo -n 'hello world & friends' | mytoolkit url-encode
hello+world+%26+friends

$ echo -n 'hello+world+%26+friends' | mytoolkit url-encode --decode
hello world & friends
```

Note: `url-encode` encodes every byte it receives, including a trailing newline — use `echo -n` (as above) or `printf` to avoid a stray `%0A` in the output when piping from `echo`.

<!-- TOC -->

- [Hash Generator — Testing](#hash-generator--testing)

<!-- TOC -->

# Hash Generator — Testing

```
$ cd app && go test ./internal/tools/hashgen/... -v
--- PASS: TestGenerate (0.00s)
    --- PASS: TestGenerate/md5_hello
    --- PASS: TestGenerate/sha1_hello
    --- PASS: TestGenerate/sha256_hello
    --- PASS: TestGenerate/sha512_hello
    --- PASS: TestGenerate/empty_md5
    --- PASS: TestGenerate/default_algorithm_is_sha256
    --- PASS: TestGenerate/unsupported_algorithm
PASS
```

Test vectors are verified against the actual computed digests (not hand-copied from external sources) to avoid shipping an incorrect "known answer".

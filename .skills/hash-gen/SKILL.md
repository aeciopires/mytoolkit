---
name: hash-gen
description: Implement or modify the Hash Generator tool (internal/tools/hashgen) — MD5/SHA-1/SHA-256/SHA-512 digests. Trigger on "implement hash generator".
---

# Hash Generator

`src/internal/tools/hashgen/hashgen.go`, `func Generate(input []byte, opts Options) (string, error)`. Stdlib only (`crypto/md5`, `crypto/sha1`, `crypto/sha256`, `crypto/sha512`, `encoding/hex`). No `sha1024` — that string in early drafts of `README.md`/`TASK.md` was a typo; the four supported algorithms are md5/sha1/sha256/sha512.

Fully generic wiring via `handlers.Wrap` / `newTextToolCommand`.

Plan: `PLANS/PLAN_HASH_GENERATOR.md`. Docs: `docs/api|cli|testing/hash-gen.md`.

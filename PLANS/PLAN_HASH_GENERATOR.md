<!-- TOC -->

- [PLAN\_HASH\_GENERATOR](#plan_hash_generator)
  - [Description](#description)
  - [Business logic](#business-logic)
  - [CLI](#cli)
  - [REST](#rest)
  - [Web UI](#web-ui)
  - [Metrics](#metrics)
  - [Unit tests](#unit-tests)
  - [Documentation](#documentation)
  - [Skill](#skill)
  - [New dependencies](#new-dependencies)

<!-- TOC -->

# PLAN_HASH_GENERATOR

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document only covers what is specific to the Hash Generator feature. Tool slug: `hash-gen`.

## Description

Generates a cryptographic hash digest of input text using a selected algorithm.

This feature implements **MD5, SHA-1, SHA-256, SHA-512** — all natively supported by the Go standard library and consistent with the algorithms explicitly named elsewhere in the requirements.

## Business logic

Package: `internal/tools/hashgen/hashgen.go`.

```go
package hashgen

type Algorithm string // "md5" | "sha1" | "sha256" | "sha512"

type Options struct {
    Algorithm Algorithm
}

func Generate(input []byte, opts Options) (string, error) // returns lowercase hex digest
```

Implementation dispatches to `crypto/md5`, `crypto/sha1`, `crypto/sha256`, `crypto/sha512`, each writing `input` into the respective `hash.Hash` and returning `hex.EncodeToString(h.Sum(nil))`.

Edge cases:
- Unsupported/unknown algorithm string → `UNSUPPORTED_ALGORITHM`, HTTP 400, listing valid values in the error message.
- Empty input → valid, returns the well-known hash of the empty string for the chosen algorithm (not an error — hashing empty input is well-defined).
- Large input (e.g. multi-MB paste) → streamed via `io.Writer` into the hash rather than loaded twice, to keep memory bounded; document a reasonable request body size limit at the HTTP layer (shared middleware concern, noted here as a dependency on `PLAN_ARCHITECTURE.md`'s middleware).

## CLI

```
mytoolkit hash-gen --algo md5|sha1|sha256|sha512 --in <file|-> [--out <file|->]
```

Example:
```
$ echo -n 'hello' | mytoolkit hash-gen --algo sha256
2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824
```

## REST

`POST /api/v1/tools/hash-gen`

Request:
```json
{ "input": "hello", "options": { "algorithm": "sha256" } }
```

Success (200):
```json
{
  "success": true,
  "data": { "output": "2cf24dba5fb0a30e26e83b2ac5b9e29e1b161e5c1fa7425e73043362938b9824" },
  "meta": { "tool": "hash-gen", "duration_ms": 0.02 }
}
```

Error (400):
```json
{ "success": false, "error": { "code": "UNSUPPORTED_ALGORITHM", "message": "algorithm must be one of: md5, sha1, sha256, sha512" } }
```

## Web UI

- Input `<textarea>`.
- Algorithm selector (MD5 / SHA-1 / SHA-256 / SHA-512), can compute and display all four simultaneously in a small table for convenience (matches common "hash generator" tool UX), each with its own "Copy" button.
- Live computation on input change via debounced `fetch()` (or one request per algorithm, or a single request returning all four — recommend extending `Options.Algorithm` to accept `"all"` server-side to return a map of all four digests in one call, documented as a UI-convenience mode alongside the single-algorithm mode).
- A small warning note that MD5/SHA-1 are not collision-resistant and shown for reference/interoperability only, not recommended for security-sensitive use — informational, matches common hash-tool UX conventions.

## Metrics

Shared `tool="hash-gen"` label; no custom metric (algorithm choice is not added as a metric label to avoid unnecessary cardinality).

## Unit tests

`internal/tools/hashgen/hashgen_test.go`:
- Known test vectors for MD5, SHA-1, SHA-256, SHA-512 (e.g. hash of `"hello"` against well-known published digests).
- Empty input produces the well-known empty-string digest for each algorithm.
- Unsupported algorithm string → error.
- Large input (e.g. 1MB) hashes without error and matches a precomputed reference digest.

## Documentation

- `docs/api/hash-gen.md`, `docs/cli/hash-gen.md`, `docs/testing/hash-gen.md`.

## Skill

`.skills/hash-gen/SKILL.md` — triggers on "implement hash generator"; documents the supported algorithm list (and the SHA-1024 typo correction), links to this plan.

## New dependencies

None — `crypto/md5`, `crypto/sha1`, `crypto/sha256`, `crypto/sha512`, `encoding/hex` from the standard library.

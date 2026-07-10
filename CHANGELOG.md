<!-- TOC -->

- [Changelog](#changelog)
  - [\[1.0.0\] - 2026-07-10](#010---2026-07-10)
    - [Added](#added)
    - [Changed](#changed)
    - [Fixed](#fixed)

<!-- TOC -->

# Changelog

All notable changes to this project are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [1.0.0] - 2026-07-10

### Added

- Initial implementation of MyToolkit: a Go web application and CLI exposing 11 tools — JSON Tree Viewer, JSON Formatter, YAML Formatter, Password Generator, JWT Encode/Decode, QR Code Generator, Character/Word/Line Counter, URL Encode/Decode, Hash Generator, Base64 Encode/Decode, and Case Converter (added beyond the original scope).
- REST API under `/api/v1/tools/<slug>` for every tool, with a shared JSON success/error envelope.
- CLI subcommand per tool, with a shared `--in`/`--out` (file-or-stdio) convention.
- `mytoolkit --version`/`-v`, reading from the repo-root `VERSION` file — the single source of truth also used to tag Docker images.
- Server-rendered web UI with a Material Design 3–inspired dark/light theme; pages accept a `?theme=light|dark` override for deterministic screenshots/demos.
- `/healthz`, `/readyz`, `/metrics` (Prometheus), and `/api/v1/metrics/ranking` (usage ranking) endpoints.
- Structured JSON logging via zerolog, with request correlation IDs.
- Multi-stage, multi-arch (linux/amd64, linux/arm64) Dockerfile, validated with `docker buildx`, and `docker-compose.yml`.
- `make docker-push`: interactive Docker Hub publish (hidden credential prompt, `--password-stdin`, never logged).
- Helm chart under `helm/mytoolkit` — Prometheus scrape annotations, `helm-docs`-generated chart README, deployed and `helm test`-verified against a real `kind-multinodes` cluster.
- Unit tests for every tool's business logic, plus per-feature documentation (`docs/api|cli|testing`, with a Mermaid workflow diagram each) and `.skills/`.
- JSON to TOON Converter (`json-toon`), the 12th tool — converts JSON into [TOON](https://github.com/toon-format/spec) to reduce LLM token usage (tabular arrays, minimal quoting, indentation-based nesting). REST (`POST /api/v1/tools/json-toon`) and CLI (`mytoolkit json-toon`) are full Go implementations like every other tool. The web page is the first exception in this app: it converts entirely client-side via an independent JavaScript implementation (`internal/web/static/js/json-toon.js`), so no JSON is ever sent to the server from the interactive tool — verified by grepping the JS file for network APIs (none) and by an empty browser network tab during use. Introduces a new reusable `data-client-side` convention on `.tool-panel` (`registry.Tool.ClientSide`) for any future browser-only tool. Because two implementations of the same algorithm exist, `mytoolkit_tool_usage_total{tool="json-toon"}` only reflects REST/CLI usage — web conversions aren't visible to server-side metrics (same caveat as CLI usage on every other tool). The Go and JS implementations are verified against one shared fixture table (20/20 cases match exactly).

### Changed

- Go/HTML/CSS/JS source (`cmd/`, `internal/`) moved under `src/` (its own Go module root), separating application source from planning, documentation, and deployment files at the repo root.

### Fixed

- `README.md` previously listed a non-existent "SHA-1024" hash algorithm; the Hash Generator implements MD5, SHA-1, SHA-256, and SHA-512.
- The Helm chart's test hook was initially placed at `tests/test-connection.yaml`; Helm only discovers hooks under `templates/`, so it was silently never run until moved to `templates/tests/test-connection.yaml`.
- Verification pass across all 12 tools: added missing Mermaid workflow diagrams to 7 `docs/api/<tool>.md` files that didn't have one (base64, case-convert, hash-gen, jwt, qrcode, text-count, url-encode).
- `docs/api/url-encode.md` documented a nonexistent `options.mode` field; the real field is `options.decode` (boolean). Fixed to match `urlencode.Options`.
- `docs/cli/url-encode.md`'s example used `echo` (which appends a trailing newline that gets percent-encoded as a stray `%0A`); switched to `echo -n` and documented why.
- `docs/cli/json-tree.md`'s example output used a hand-typed compact JSON format that didn't match the real `json.MarshalIndent` output; corrected to the actual (fully expanded) output.
- `docs/api/json-format.md` and `docs/api/yaml-format.md` error examples used unverified, hypothetical parser error messages; both now include the exact request that reproduces the documented error message.
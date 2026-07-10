<!-- TOC -->

- [PLAN\_K8S\_YAML\_VALIDATOR](#plan_k8s_yaml_validator)
  - [Description](#description)
  - [Business logic](#business-logic)
  - [Known limitations](#known-limitations)
  - [CLI](#cli)
  - [REST](#rest)
  - [Web UI](#web-ui)
  - [Metrics](#metrics)
  - [Unit tests](#unit-tests)
  - [Documentation](#documentation)
  - [Skill](#skill)
  - [New dependencies](#new-dependencies)

<!-- TOC -->

# PLAN_K8S_YAML_VALIDATOR

Shared architecture, REST envelope, CLI conventions, metrics, health check, and theming are defined in [PLAN_ARCHITECTURE.md](PLAN_ARCHITECTURE.md). This document covers what is specific to the Kubernetes YAML Validator feature. Tool slug: `k8s-validate`.

Added at the user's explicit request, using [`sigs.k8s.io/yaml`](https://github.com/kubernetes-sigs/yaml) — see `PLAN_YAML_TO_JSON_CONVERTER.md`'s intro for why this library (not `gopkg.in/yaml.v3` alone) is the right choice whenever "would the real Kubernetes API accept this YAML" is the question being answered.

## Description

Validates that a YAML document — or a `---`-separated multi-document stream, as accepted by `kubectl apply -f` — is well-formed YAML *and* satisfies the two structural requirements every Kubernetes API object must have: a non-empty string `apiVersion`, a non-empty string `kind`, and (if present) an object-shaped `metadata` field. Each document is converted with `sigs.k8s.io/yaml.YAMLToJSONStrict`, the exact library kubectl/client-go/the API server itself use to turn YAML manifests into JSON before deserializing them — so a document this tool rejects would also be rejected by a real cluster, for the same underlying reason.

## Business logic

Package: `internal/tools/k8svalidate/k8svalidate.go`.

```go
package k8svalidate

type Options struct{}

type DocumentResult struct {
    Index      int    `json:"index"`
    APIVersion string `json:"api_version,omitempty"`
    Kind       string `json:"kind,omitempty"`
    Name       string `json:"name,omitempty"`
    Valid      bool   `json:"valid"`
    Error      string `json:"error,omitempty"`
}

type Result struct {
    Valid     bool             `json:"valid"`
    Documents []DocumentResult `json:"documents"`
}

func Validate(input []byte, opts Options) (Result, error)
```

This is the second tool in this app (after JSON Tree Viewer, Text Counter, Password Generator, JWT, QR Code) whose business logic returns a structured Go type instead of a plain string — `handlers.Wrap`/`newTextToolCommand` don't fit, since there's no single "output string," so CLI/REST wiring is bespoke (see below).

Implementation, in two stages, deliberately using two different YAML libraries for two different jobs:
1. **Document splitting**: `gopkg.in/yaml.v3`'s `yaml.NewDecoder(...).Decode(&node)`, looped until `io.EOF` — the same technique already used by `internal/tools/yamlformat` to handle `---`-separated streams. `sigs.k8s.io/yaml` has no public multi-document API (its `Unmarshal`/`YAMLToJSON` operate on one document's worth of bytes), so `yaml.v3` does the stream splitting; each isolated document is re-marshaled back to bytes (`yaml.Marshal(&node)`) for the next stage. **Verified** that this round trip does not silently deduplicate or otherwise "fix" a document with a duplicate key — the duplicate survives the round trip and is still caught by stage 2.
2. **Per-document validation**: `sigs.k8s.io/yaml.YAMLToJSONStrict` converts each isolated document's bytes to JSON (rejecting duplicate mapping keys, exactly like `yaml-to-json`). The resulting JSON is decoded into a generic `map[string]any` (not a `TypeMeta` struct — see below) and manually checked for `apiVersion`/`kind`/`metadata` shape, producing a plain-English error message that doesn't leak Go type/field names.

## Known limitations

Stated explicitly, not glossed over:
- **No resource-schema validation.** This tool checks the two universal TypeMeta fields, nothing resource-specific (it doesn't know that a `Deployment`'s `spec.replicas` must be an integer, or validate a CRD's schema). That requires the full Kubernetes OpenAPI schema database, which is what `kubectl apply --dry-run=server` (against a real API server) or a dedicated tool like `kubeconform`/`kubeval` does — pulling in `k8s.io/apimachinery`/`k8s.io/api` for that would be a large, heavy dependency for a lightweight browser/CLI utility, and was deliberately not done here. The web page and docs state this limitation plainly so users don't mistake a pass here for "this manifest will definitely apply cleanly."
- **A document-stream-level YAML syntax error aborts the whole call** (returns an `apperr`, HTTP 400) rather than being reported per-document, since once the tokenizer fails there's no reliable way to resync to the next `---` boundary. Once a document is successfully isolated by stage 1, though, a per-document problem (duplicate keys, missing/malformed fields) is reported in that document's `DocumentResult` without aborting the rest of the batch — so a 10-document stream with one mistake still reports on the other 9.
- **Blank documents are silently skipped**, not reported as invalid — a stray leading/trailing `---` (e.g. `---\napiVersion: v1\n...\n---\n`) is common and harmless, and `kubectl` itself ignores it; verified this directly (an all-separator input decodes each blank slot to JSON `null`, which is filtered out before validation).
- Inherits the same YAML 1.1 boolean/null-resolution behavior documented in `PLAN_YAML_TO_JSON_CONVERTER.md` (e.g. an unquoted `kind: no` would resolve to the boolean `false`, not the string `"no"`, and then correctly fail as "field kind must be a string, got boolean" — arguably a feature here, since it surfaces a real footgun a user would hit against a real cluster too).

## CLI

```
mytoolkit k8s-validate --in <file|-> [--out <file|->]
```

Exit code `0` if every document is valid, `1` otherwise (including on a YAML syntax error) — usable directly in CI (`mytoolkit k8s-validate --in manifest.yaml || exit 1`).

Example (valid):
```
$ printf 'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\ndata:\n  a: b\n' | mytoolkit k8s-validate
Document 1: VALID (apiVersion=v1, kind=ConfigMap, name=cm1)

1/1 document(s) valid.
```

Example (multi-document, one invalid):
```
$ printf 'apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\n---\napiVersion: apps/v1\nmetadata:\n  name: dep1\n' | mytoolkit k8s-validate
Document 1: VALID (apiVersion=v1, kind=ConfigMap, name=cm1)
Document 2: INVALID: missing required field "kind" (apiVersion=apps/v1)

1/2 document(s) valid.
Error: 1 of 2 document(s) failed Kubernetes validation
```

## REST

`POST /api/v1/tools/k8s-validate`

Request:
```json
{ "input": "apiVersion: v1\nkind: ConfigMap\nmetadata:\n  name: cm1\n" }
```

Success (200) — note `success: true` even when `data.valid` is `false`: an HTTP-level success means "the validator ran," which is a different question from "is your YAML valid" (reported inside `data`), the same way a linter's own process exit code is separate from its findings:
```json
{
  "success": true,
  "data": {
    "valid": true,
    "documents": [
      { "index": 1, "api_version": "v1", "kind": "ConfigMap", "name": "cm1", "valid": true }
    ]
  },
  "meta": { "tool": "k8s-validate", "duration_ms": 0.15 }
}
```

A document-level failure still returns `success: true` with `data.valid: false`:
```json
{
  "success": true,
  "data": {
    "valid": false,
    "documents": [
      { "index": 1, "api_version": "v1", "kind": "ConfigMap", "name": "cm1", "valid": true },
      { "index": 2, "api_version": "apps/v1", "valid": false, "error": "missing required field \"kind\"" }
    ]
  },
  "meta": { "tool": "k8s-validate", "duration_ms": 0.2 }
}
```

Only a hard YAML syntax error returns HTTP 400:
```json
{ "success": false, "error": { "code": "INVALID_YAML", "message": "yaml: line 1: did not find expected ',' or ']'" } }
```

Error codes: `EMPTY_INPUT`, `NO_DOCUMENTS` (input parses but contains no real documents, e.g. only `---` separators), `INVALID_YAML`.

Bespoke handler (`internal/cli/k8svalidate.go`'s `k8sValidateHandler`), not `handlers.Wrap` — the response shape is `Result` (`{valid, documents}`), not `{output: string}`.

## Web UI

No custom template beyond the shared `tool-panel` partial and a `tool-options` note — this tool needed **zero bespoke JavaScript**, because `tool-common.js`'s existing generic `run()` already falls back to `JSON.stringify(json.data, null, 2)` in the output textarea whenever a REST response has no `data.output` field (this fallback already existed, added for robustness, but no prior tool actually exercised it — `k8s-validate` is the first). The pretty-printed JSON result (`{valid, documents: [...]}`) is a perfectly readable report as-is; adding a bespoke renderer (like JSON Tree Viewer's colored tree) would have been unjustified complexity for what's already a small, flat structure. If a future revision wants a friendlier report (colored VALID/INVALID badges, e.g.), that's a deliberate UI upgrade to design then, not a gap today.

## Metrics

Shared `tool="k8s-validate"` label; no custom metric.

## Unit tests

`internal/tools/k8svalidate/k8svalidate_test.go`:
- Valid minimal object, valid object without `metadata`.
- Missing `apiVersion`, missing `kind`, empty `apiVersion`.
- `kind` wrong type (number), `metadata` wrong type (string).
- Root value is an array (not a mapping).
- Empty input → error. Input with only `---` separators → error (`NO_DOCUMENTS`). Malformed YAML → error. Tab-character indentation → error.
- Multi-document stream: one invalid document doesn't affect the others' results; overall `Result.Valid` is `false` if any document is invalid.
- Blank documents (stray `---`) are skipped, not reported.
- A duplicate key in one document of a multi-document stream is caught for that document only, without aborting the rest of the batch.

## Documentation

- `docs/api/k8s-validate.md`, `docs/cli/k8s-validate.md`, `docs/testing/k8s-validate.md`.
- `README.md`: add to the Features list and Documentation table.

## Skill

`.skills/k8s-validate/SKILL.md` — triggers on "implement Kubernetes YAML validator", "validate k8s manifest"; documents the two-stage (yaml.v3 split, then sigs.k8s.io/yaml per-document) design and why, and the explicit "no schema validation" scope boundary.

## New dependencies

None beyond `sigs.k8s.io/yaml` (already added for `yaml-to-json`/`json-to-yaml`) and `gopkg.in/yaml.v3` (already a dependency for `yaml-format`).

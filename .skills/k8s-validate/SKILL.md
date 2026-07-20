---
name: k8s-validate
description: Implement or modify the Kubernetes YAML Validator tool (internal/tools/k8svalidate) — checking that YAML is well-formed and has the apiVersion/kind/metadata shape the Kubernetes API requires, via sigs.k8s.io/yaml. Trigger on "implement Kubernetes YAML validator", "validate k8s manifest".
---

# Kubernetes YAML Validator

`app/internal/tools/k8svalidate/k8svalidate.go`, `func Validate(input []byte, opts Options) (Result, error)`. Returns a structured `Result{Valid, Documents []DocumentResult}`, not a plain string — like `jsontree`/`textcount`/`password`/`jwt`/`qrcode`, this means bespoke REST/CLI wiring (`app/internal/cli/k8svalidate.go`), not `handlers.Wrap`/`newTextToolCommand`.

## Two YAML libraries, two different jobs — don't collapse them into one

1. **`gopkg.in/yaml.v3`** splits a `---`-separated multi-document stream into individual documents: `yaml.NewDecoder(...).Decode(&node)` looped until `io.EOF`, the same technique `internal/tools/yamlformat` uses (see `.skills/yaml-format/SKILL.md`). `sigs.k8s.io/yaml` has no public multi-document API — its `Unmarshal`/`YAMLToJSON` only operate on one document's worth of bytes — so this stage has to be `yaml.v3`.
2. **`sigs.k8s.io/yaml.YAMLToJSONStrict`** validates each already-isolated document exactly the way kubectl/client-go/the API server itself would decode it (same choice as `yaml-to-json`, see `.skills/yaml-to-json/SKILL.md` — strict mode rejects duplicate keys instead of silently keeping one).

**Verified, not assumed**: round-tripping a document through `yaml.v3` (`Decode` then `Marshal(&node)` to get clean bytes for stage 2) does not silently deduplicate a duplicate key — it survives to be caught by `YAMLToJSONStrict`. If you change the splitting approach, re-verify this; it's easy to accidentally "fix" the duplicate away by decoding into a `map[string]any` instead of a `yaml.Node` somewhere in the pipeline.

## Why `map[string]any`, not a `TypeMeta` struct

`validateDocument` decodes each document's JSON into a generic `map[string]any` and manually checks `apiVersion`/`kind`/`metadata`, rather than `json.Unmarshal`ing into a Go struct. A struct-based approach was tried first and produces uglier errors — Go's `encoding/json` embeds the (unexported, internal) struct type name in type-mismatch messages, e.g. `json: cannot unmarshal number into Go struct field k8svalidate.typeMeta.kind of type string`, which leaks implementation detail a user shouldn't have to parse. The manual approach produces `field "kind" must be a string, got number` instead — keep it that way; don't "simplify" back to a struct without fixing the error-message quality it would regress.

## Deliberate scope boundary — do not add resource-schema validation without discussing it first

This tool checks exactly two universal fields (non-empty `apiVersion`, non-empty `kind`) plus `metadata`'s shape (object, if present) — the TypeMeta-level requirements every Kubernetes API object has, regardless of resource type. It does **not** know that a `Deployment`'s `spec.replicas` must be an integer, or validate any CRD's schema. Doing that properly needs the full Kubernetes OpenAPI schema database (`k8s.io/apimachinery`/`k8s.io/api`, or an embedded copy of the OpenAPI spec) — a large dependency this lightweight app deliberately doesn't carry. If a future request asks for "full" validation, that's a scope change requiring a new plan document and probably a different dependency strategy, not an incremental patch to this package.

## Error-handling split: stream-level vs. document-level

A YAML *syntax* error (can't parse at all) aborts the whole `Validate` call via `apperr` (HTTP 400) — there's no reliable way to resync to the next `---` boundary after a tokenizer failure. A problem discovered *after* a document is successfully isolated (duplicate key, missing/wrong-type field) is recorded in that document's `DocumentResult` instead, and the loop continues to the next document. Don't change duplicate-key/field errors to abort the whole batch — that would make a 10-document manifest stream with one typo stop reporting on the other 9, which is strictly worse for a validator's usefulness.

## Blank documents are skipped, not flagged

A document that decodes to JSON `null` (a stray leading/trailing `---`, or an explicit `---\n---\n`) is silently skipped, not counted or reported — this matches real `kubectl` behavior and was verified directly (`TestValidateSkipsBlankDocuments`).

## Web UI needed no custom code

`k8s-validate.html` uses the shared `tool-panel` partial and nothing else. `tool-common.js`'s generic `run()` already falls back to `JSON.stringify(json.data, null, 2)` when a response has no `data.output` — this tool is the first to actually rely on that fallback, and it produces a perfectly readable report on its own. Don't add a bespoke renderer unless there's a real product ask for one (colored badges, etc.); it would be unjustified complexity for what's already a small, flat JSON structure.

## CLI exit code carries validity, not just process success

Unlike every other tool's CLI command, `mytoolkit k8s-validate`'s `RunE` still writes the full report to `--out` on a semantic failure (missing kind, etc.), then returns a *plain* `fmt.Errorf` (not routed through `apperr`) to get a non-zero exit code — this is deliberate, so the command is usable directly in CI (`mytoolkit k8s-validate --in manifest.yaml || exit 1`). A hard YAML syntax error still goes through the normal `apperr`-driven error path with no report written.

MCP: `k8s-validate` tool (`app/internal/mcp/k8s_validate.go`) — mirrors the REST handler: a semantically-invalid-but-parseable document is `isError: false` with `Result.Valid == false`, not a tool error. Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_K8S_YAML_VALIDATOR.md`. Docs: `docs/api|cli|testing/k8s-validate.md`.

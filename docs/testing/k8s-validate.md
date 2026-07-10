<!-- TOC -->

- [Kubernetes YAML Validator — Testing](#kubernetes-yaml-validator--testing)

<!-- TOC -->

# Kubernetes YAML Validator — Testing

```
$ cd app && go test ./internal/tools/k8svalidate/... -v
=== RUN   TestValidateSingleDocument
--- PASS: TestValidateSingleDocument (0.00s)
    --- PASS: TestValidateSingleDocument/valid_minimal_object
    --- PASS: TestValidateSingleDocument/valid_without_metadata
    --- PASS: TestValidateSingleDocument/missing_apiVersion
    --- PASS: TestValidateSingleDocument/missing_kind
    --- PASS: TestValidateSingleDocument/empty_apiVersion
    --- PASS: TestValidateSingleDocument/kind_wrong_type
    --- PASS: TestValidateSingleDocument/metadata_wrong_type
    --- PASS: TestValidateSingleDocument/root_is_an_array
    --- PASS: TestValidateSingleDocument/empty_input
    --- PASS: TestValidateSingleDocument/only_separators
    --- PASS: TestValidateSingleDocument/malformed_yaml
    --- PASS: TestValidateSingleDocument/tab_indentation
=== RUN   TestValidateMultiDocument
--- PASS: TestValidateMultiDocument (0.00s)
=== RUN   TestValidateSkipsBlankDocuments
--- PASS: TestValidateSkipsBlankDocuments (0.00s)
=== RUN   TestValidateDuplicateKeyRejectedPerDocument
--- PASS: TestValidateDuplicateKeyRejectedPerDocument (0.00s)
PASS
```

Covers: valid objects (with and without `metadata`), missing/empty/wrong-type `apiVersion`/`kind`, wrong-type `metadata`, a non-mapping root value, empty input, an all-separator input (`NO_DOCUMENTS`), malformed YAML, tab-character indentation, a 3-document stream where one document is invalid without affecting the others' results, silent skipping of blank documents (stray `---`), and per-document duplicate-key rejection in a multi-document stream.

## Web UI note

This tool's web page uses the shared generic `tool-panel` wiring with **no bespoke JavaScript** — `tool-common.js`'s existing `JSON.stringify(json.data, null, 2)` fallback (used whenever a REST response has no `data.output` field) renders the `{valid, documents}` result directly and readably. Verified with a real browser (Playwright driving the actual binary): typing valid/invalid manifests into the input updates the output textarea with the same JSON shape documented in `docs/api/k8s-validate.md`, in both light and dark themes, with no console errors beyond the expected one from an intentionally-triggered request (there are none here, since this tool never returns an HTTP error status for a semantically-invalid manifest — only for a hard YAML syntax error).

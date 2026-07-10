---
name: yaml-to-json
description: Implement or modify the YAML to JSON Converter tool (internal/tools/yamltojson) — converting a YAML document to pretty-printed JSON via sigs.k8s.io/yaml. Trigger on "implement YAML to JSON converter", "yaml to json".
---

# YAML to JSON Converter

`app/internal/tools/yamltojson/yamltojson.go`, `func Convert(input []byte, opts Options) (string, error)`. Uses `sigs.k8s.io/yaml.YAMLToJSONStrict`, not `gopkg.in/yaml.v3` (used by `yaml-format` — see that skill) and not the library's plain `YAMLToJSON`.

## Why `YAMLToJSONStrict`, not `YAMLToJSON`

The YAML spec forbids duplicate mapping keys. `sigs.k8s.io/yaml.YAMLToJSON`'s own doc comment admits it doesn't enforce that — duplicates are "case-sensitively ignored in an undefined order." `YAMLToJSONStrict` reports them as an error instead. Don't switch back to the plain variant without a strong reason; it trades a loud error for silent data loss.

## Why not `gopkg.in/yaml.v3` (the library `yaml-format` uses)

`yaml-format` needs `yaml.Node` to preserve comments/anchors/key order for a same-format round trip. This tool converts *across* formats — there's no such round trip to preserve, and `sigs.k8s.io/yaml` (decode YAML into a generic value, re-encode as JSON) is a better fit and is what Kubernetes itself uses for the same job.

## The YAML 1.1 boolean-resolution gotcha ("Norway problem")

Verified directly, not assumed: unquoted `NO`/`y` (and `n`/`yes`/`on`/`off`) in source YAML resolve to JSON `false`/`true`, not the strings `"NO"`/`"y"`. This is the library's own documented behavior (see its `YAMLToJSON` doc comment), not a bug in this package — `TestConvertYAML11BooleanResolution` in `yamltojson_test.go` locks this in so a future dependency bump that changes it gets caught. Don't try to "fix" this by adding custom string-preservation logic; it would diverge from what the library (and Kubernetes' own YAML handling) actually does.

## Other verified behaviors

- Empty input is **not** rejected by the library itself (`YAMLToJSONStrict("")` → `"null"`, no error) — the `len(input) == 0` check in `Convert` must stay before the library call.
- Only the first document of a `---`-separated multi-document stream is converted — a documented limitation, deliberately not fixed the way `yaml-format` was (see `.skills/yaml-format/SKILL.md`); turning a multi-doc stream into a JSON array would be a different, unrequested design decision.
- Large integers (up to 64 bits) are preserved exactly.

REST/CLI wiring is fully generic via `handlers.Wrap` / `newTextToolCommand` (see `app/internal/cli/yamltojson.go`).

Plan: `PLANS/PLAN_YAML_TO_JSON_CONVERTER.md`. Docs: `docs/api|cli|testing/yaml-to-json.md`.

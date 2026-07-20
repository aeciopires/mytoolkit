---
name: json-to-yaml
description: Implement or modify the JSON to YAML Converter tool (internal/tools/jsontoyaml) — converting a JSON document to YAML via sigs.k8s.io/yaml. Trigger on "implement JSON to YAML converter", "json to yaml".
---

# JSON to YAML Converter

`app/internal/tools/jsontoyaml/jsontoyaml.go`, `func Convert(input []byte, opts Options) (string, error)`. Uses `sigs.k8s.io/yaml.JSONToYAML` — see `.skills/yaml-to-json/SKILL.md` for why this library (not `gopkg.in/yaml.v3`) was chosen for this converter.

## Why `encoding/json.Unmarshal` runs before `yaml.JSONToYAML`

This is the load-bearing part of `Convert` — don't remove it. `sigs.k8s.io/yaml.JSONToYAML` parses its input with a **YAML** decoder internally (YAML is technically a superset of JSON), so called alone it does not enforce JSON's stricter grammar. Verified directly: trailing commas (`{"a":1,}`), unquoted keys (`{a:1}`), single-quoted strings (`{"a": 'hello'}`), and even a bare `// comment` line were all silently "accepted" — the comment case produced nonsense output instead of an error. `Convert` therefore validates with `encoding/json.Unmarshal` first (which does enforce the JSON grammar) and only calls the library once that succeeds. `TestConvert`'s three "malformed json" subtests in `jsontoyaml_test.go` guard this — if a future refactor removes the pre-validation step, those tests catch it.

## Automatic quoting protects against the YAML 1.1 boolean-resolution gotcha

The JSON string `"NO"` or `"y"` comes out of `JSONToYAML` as quoted YAML (`"NO"`, `"y"`), not bare `NO`/`y` — verified directly. This is the safe direction of the same "Norway problem" documented in `.skills/yaml-to-json/SKILL.md`: converting *to* YAML, the library actively avoids creating an ambiguous unquoted scalar that a YAML 1.1 reader would misresolve as a boolean.

## Other verified behaviors

- Duplicate JSON keys: last value wins, silently — same as `encoding/json` itself elsewhere in this app. JSON (unlike YAML) doesn't forbid duplicates, so this isn't inconsistent with `yaml-to-json`'s stricter duplicate-key rejection; that rule comes from the YAML spec, not this codebase's own policy.
- Large integers (up to 64 bits) are preserved exactly.
- `Options` is intentionally empty (`type Options struct{}`) — the library's `Marshal`/`JSONToYAML` functions take no formatting parameters (no indent knob to expose), unlike `yaml-format` or `yaml-to-json`.

REST/CLI wiring is fully generic via `handlers.Wrap` / `newTextToolCommand` (see `app/internal/cli/jsontoyaml.go`). The web page (`json-to-yaml.html`) has no `tool-options` block since there are no options to expose.

MCP: `json-to-yaml` tool (`app/internal/mcp/json_to_yaml.go`), no options field (mirrors `jsontoyaml.Options{}`). Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_JSON_TO_YAML_CONVERTER.md`. Docs: `docs/api|cli|testing/json-to-yaml.md`.

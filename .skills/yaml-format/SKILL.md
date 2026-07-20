---
name: yaml-format
description: Implement or modify the YAML Formatter tool (internal/tools/yamlformat) — reformatting YAML with consistent indentation. Trigger on "implement YAML formatter".
---

# YAML Formatter

`app/internal/tools/yamlformat/yamlformat.go`, `func Format(input []byte, opts Options) (string, error)`. Decodes into `yaml.Node` (gopkg.in/yaml.v3), not `map[string]any`, to preserve key order, comments, anchors/aliases, and scalar style hints.

## Multi-document streams

`Format` loops `yaml.NewDecoder(...).Decode(&node)` until `io.EOF`, feeding every document to a single shared `yaml.Encoder` — the encoder automatically emits `---` between documents when `Encode` is called more than once. Do **not** go back to a single `yaml.Unmarshal` call; that was the original (buggy) implementation and silently dropped every document after the first, which is a real spec violation (YAML streams can hold multiple `---`-separated documents, spec §9.1). A stream that decodes to zero documents (comments/whitespace only) is treated as `ErrEmptyInput`, not an empty success.

## Style normalization (`block` / `flow`)

`Options.Style` forces every mapping/sequence node's `.Style` field recursively (`normalizeStyle` in `yamlformat.go`) to either block (indented, the default) or `yaml.FlowStyle` (compact `{}`/`[]`, single line) — this is the YAML equivalent of JSON Formatter's pretty/minify. It's always safe because the YAML spec states collection style is a presentation detail not reflected in the representation graph. Only collection nodes are touched; scalar node styles (plain/quoted) are left alone since quoting can be semantically meaningful (e.g. `"yes"` is a string, unquoted `yes` may resolve differently depending on schema) — never call `normalizeStyle` on `yaml.ScalarNode`s.

## Known, spec-grounded limitations (not bugs)

- Comment attachment can shift for a comment sitting alone between two mapping keys — the spec defines no formal comment-to-node attachment rule, so this isn't fixable by this formatter or `yaml.v3`.
- Merge keys (`<<`) round-trip with an explicit `!!merge` tag added even if the source omitted it — semantically identical output, not a defect.
- Plain scalars affected by the "Norway problem" (`yes`/`no`/`NO`/etc. in YAML 1.1) are preserved as their original text rather than re-resolved, because `Format` re-serializes the parsed node's existing scalar value, not a re-interpreted Go type.

REST/CLI wiring is fully generic via `handlers.Wrap` / `newTextToolCommand` (see `app/internal/cli/yamlformat.go`); the CLI also exposes `--style`.

## Not the same library as the YAML⇄JSON converters

`yaml-to-json`/`json-to-yaml` deliberately use `sigs.k8s.io/yaml` instead of `gopkg.in/yaml.v3` — see `.skills/yaml-to-json/SKILL.md` for why (this formatter needs `yaml.Node` for a same-format round trip; those converters don't have one to preserve). Don't "consolidate" onto one YAML library without re-reading both skills first.

MCP: `yaml-format` tool (`app/internal/mcp/yaml_format.go`). Docs: `mcp/README.md`.

Plan: `PLANS/PLAN_YAML_FORMATTER.md`. Docs: `docs/api|cli|testing/yaml-format.md`.

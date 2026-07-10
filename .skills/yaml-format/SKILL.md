---
name: yaml-format
description: Implement or modify the YAML Formatter tool (internal/tools/yamlformat) — reformatting YAML with consistent indentation. Trigger on "implement YAML formatter".
---

# YAML Formatter

`src/internal/tools/yamlformat/yamlformat.go`, `func Format(input []byte, opts Options) (string, error)`. Decodes into `yaml.Node` (gopkg.in/yaml.v3), not `map[string]any`, to preserve key order and scalar style hints; re-encodes via `yaml.NewEncoder(...).SetIndent(n)`. Only the first document in a multi-document stream is processed — a known, documented limitation.

Fully generic wiring via `handlers.Wrap` / `newTextToolCommand`.

Plan: `PLANS/PLAN_YAML_FORMATTER.md`. Docs: `docs/api|cli|testing/yaml-format.md`.

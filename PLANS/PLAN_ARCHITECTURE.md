<!-- TOC -->

- [PLAN\_ARCHITECTURE](#plan_architecture)
  - [Context](#context)
  - [Assumptions and open decisions](#assumptions-and-open-decisions)
  - [Go module layout](#go-module-layout)
  - [Shared code and configuration reuse](#shared-code-and-configuration-reuse)
  - [Library and framework choices](#library-and-framework-choices)
  - [Dual mode: web (default) and CLI](#dual-mode-web-default-and-cli)
  - [REST API design](#rest-api-design)
  - [Metrics design](#metrics-design)
  - [Logging](#logging)
  - [Health check design](#health-check-design)
  - [Configuration](#configuration)
  - [Theming and visual design system](#theming-and-visual-design-system)
  - [Testing strategy](#testing-strategy)
  - [Docker, Helm and Makefile shape](#docker-helm-and-makefile-shape)
  - [Documentation structure](#documentation-structure)
  - [Version control hygiene (.gitignore)](#version-control-hygiene-gitignore)
  - [Full repository tree](#full-repository-tree)
  - [Suggested build sequencing](#suggested-build-sequencing)
  - [Verification](#verification)
  - [Additional instructions received during implementation](#additional-instructions-received-during-implementation)

<!-- TOC -->

# PLAN_ARCHITECTURE

## Context

`PLANS/TASK.md` asks for a Go, web-first application ("MyToolkit") that also runs as a CLI, exposing every feature listed in `README.md` as a REST endpoint, a responsive web page, and a CLI subcommand, plus health-check and metrics endpoints, dark/light theming, tests, and full documentation.

This document is the **shared architectural foundation** referenced by every `PLANS/PLAN_<FEATURE>.md`. It exists so the twelve feature plans don't each repeat the same routing, CLI, metrics, theming, and packaging decisions. A future implementation session should read this document first, then read the specific feature plan for the tool being built.

## Assumptions and open decisions

- `README.md` lists MD5, SHA-256, SHA-512 explicitly. The Hash Generator will implement **MD5, SHA-1, SHA-256, SHA-512** (all available in the Go standard library). `README.md` will be corrected during implementation and the fix noted in `CHANGELOG.md`.
- **`kind-king-multinodes`**: this is a pre-existing kind cluster name given by the user, not something this project creates. Makefile/Helm instructions target it via `--kube-context kind-king-multinodes`; cluster creation is out of scope.
- **"Latest stable versions"**: this document intentionally does not pin exact version numbers. `go.mod` will be initialized with the latest stable Go toolchain and each dependency's latest stable tagged release at implementation time.
- **Scope of this plan**: this document and the per-feature plans are **design documents only**. No Go code, Dockerfile, Helm chart, or Makefile is created while writing the plans; that is the next, separate implementation phase.
- **Case Converter feature**: added after the initial planning pass, at the user's request, and is not part of the original `README.md` feature list. `README.md` will be updated during implementation to list it as an 11th feature, alongside the SHA algorithm fix noted above.
- **`.gitignore`**: the repository currently has a generic Visual Studio/.NET `.gitignore` template, unrelated to this Go project. It will be replaced with the Go-focused version defined in [Version control hygiene](#version-control-hygiene-gitignore) during Phase 0 scaffolding.

## Go module layout

Standard Go project layout. The central rule: **`internal/tools/<name>` packages are pure functions** with zero dependency on `net/http` or `cobra`. Every other layer (REST handler, CLI command, web page) is a thin adapter around them, so business logic is implemented once and reused by all three surfaces.

```
mytoolkit/
├── cmd/mytoolkit/main.go            # entrypoint: builds root cobra command, calls Execute()
├── internal/
│   ├── app/app.go                   # wires router, middleware, template renderer -> http.Server
│   ├── config/config.go             # single Load() resolving flag > env > default, see Configuration
│   ├── apperr/apperr.go             # shared error type + codes + OneOf[T] validator, reused by all tools/handlers/CLI
│   ├── textio/textio.go             # shared --in/--out (file-or-stdio) read/write helpers, reused by all CLI subcommands
│   ├── cli/
│   │   ├── root.go                  # root cobra.Command; RunE defaults to serve when no subcommand given
│   │   ├── serve.go                 # `mytoolkit serve` (web mode)
│   │   ├── shared.go                # newTextToolCommand[Opts] generic builder, reused by most feature subcommands
│   │   └── <tool>.go                # one file per feature subcommand
│   ├── httpapi/
│   │   ├── router.go                # chi router, route table, versioning
│   │   ├── response.go              # shared JSON envelope + error codes
│   │   ├── health.go                # /healthz, /readyz
│   │   ├── handlers/
│   │   │   ├── generic.go           # Wrap[Opts] generic handler adapter, reused by most feature handlers
│   │   │   └── <tool>_handler.go    # thin per-feature handler; QR/password-gen/jwt opt out of generic.go, see Shared code and configuration reuse
│   │   └── middleware/{metrics,logging,recover}.go
│   ├── web/
│   │   ├── handlers.go              # server-rendered page handlers (index + one per tool)
│   │   ├── templates/               # go:embed - layout.html, index.html, tools/<tool>.html
│   │   │   └── partials/tool-panel.html  # shared {{define "tool-panel"}} input/output block, reused by every tools/<tool>.html
│   │   └── static/                  # go:embed - css/theme.css, css/app.css, js/theme-toggle.js, js/tool-common.js (shared fetch/copy/download helpers)
│   ├── metrics/metrics.go           # Prometheus collectors + in-memory ranking aggregator
│   ├── tools/<name>/<name>.go       # pure business logic, one package per feature
│   │                                 # + <name>_test.go colocated in the same package
│   └── registry/registry.go         # tool metadata (slug, name, description) shared by nav, CLI help, metrics labels, and route/command registration
├── docs/
│   ├── api/<tool>.md                # REST docs with example request/response
│   ├── cli/<tool>.md                # CLI usage/output docs
│   ├── testing/<tool>.md            # unit test docs/examples
│   └── environment-variables.md
├── .skills/<feature-slug>/SKILL.md  # one dev skill per feature
├── helm/mytoolkit/
│   ├── Chart.yaml
│   ├── values.yaml
│   ├── templates/{deployment,service,ingress,hpa,serviceaccount,configmap,secret,servicemonitor}.yaml
│   ├── templates/{_helpers.tpl,NOTES.txt}
│   └── tests/test-connection.yaml
├── PLANS/{TASK.md, PLAN_ARCHITECTURE.md, PLAN_<FEATURE>.md x12}
├── Dockerfile
├── docker-compose.yml
├── Makefile
├── go.mod / go.sum
├── .env-example                     # KEY=default pairs for every env var in the Configuration table, copied to .env locally
├── README.md / CHANGELOG.md / CONTRIBUTING.md / CLAUDE.md / ROADMAP.md / LICENSE
```

## Shared code and configuration reuse

Cross-cutting behavior that would otherwise be re-implemented in each of the 11 feature packages is centralized into a small set of shared internal packages, composed by every feature instead of copy-pasted:

| Package | Responsibility | Used by |
|---|---|---|
| `internal/apperr` | One error type of shared shape (`Code`, `Message`, `Status`) plus common pre-built errors (`ErrEmptyInput`) and a generic `OneOf[T]` enum validator. Every `internal/tools/<name>` function that needs a specific failure mode returns an `*apperr.Error` instead of an ad hoc string-coded error. | All 11 `internal/tools/<name>` packages, the generic REST handler, the generic CLI command builder. |
| `internal/textio` | `Read(path string) ([]byte, error)` / `Write(path string, data []byte) error` — resolves `""`/`"-"` to stdin/stdout, a real path otherwise. | Every `internal/cli/<name>.go` subcommand's `--in`/`--out` handling (otherwise 11 copies of the same logic). |
| `internal/config` | Single `Load() (Config, error)` implementing the CLI-flag > env-var > default precedence from [Configuration](#configuration) once, returning a typed struct (`Host`, `Port`, `LogLevel`, …). | `internal/cli/serve.go`, and any future command needing a configurable value — no `os.Getenv` calls scattered across the codebase. |
| `internal/httpapi/handlers` generic wrapper | `Wrap[Opts any](slug string, fn func(input []byte, opts Opts) (string, error)) http.HandlerFunc` — decodes the shared request envelope, calls the tool function, encodes the shared success/error envelope via `apperr`, and feeds the shared metrics/logging middleware. | Registration of the 8 "text in, text out" tool routes in `router.go`. QR Code (binary body), Password Generator (no `input` field), and JWT (two option shapes) register their own thin handler, still reusing `apperr` + logging + metrics directly instead of the generic wrapper — documented as explicit exceptions, not omissions. |
| `internal/cli` shared command builder | `newTextToolCommand[Opts any](use, short string, flags func(*pflag.FlagSet, *Opts), fn func(input []byte, opts Opts) (string, error)) *cobra.Command` — wires `--in`/`--out` via `textio`, calls `fn`, maps `apperr` to a process exit code + stderr message. | The same 8 "text in, text out" subcommands; Password Generator/JWT/QR Code again build their own `RunE`, still reusing `textio`/`apperr`. |
| `internal/registry` | Already-planned single source of truth for tool metadata (slug, name, description) — also drives route registration (iterate `registry.All()` instead of hand-listing 11 `r.Post(...)` calls) and the web nav/homepage grid. | Router setup, CLI root command listing, web nav, `GET /api/v1/tools`, and the `README.md` [Documentation](#documentation-structure) table. |
| `internal/web/templates/partials/tool-panel.html` + `internal/web/static/js/tool-common.js` | Shared `{{define "tool-panel"}}` markup block (input/output textareas, button row) and shared vanilla-JS helpers (debounced `fetch()`, copy-to-clipboard, blob download, error-banner rendering), reused by every `tools/<tool>.html` page. A `data-client-side` attribute on `.tool-panel` opts a page out of the shared fetch-based `run()` wiring (copy/reset/download wiring still applies) for tools whose web page must never call the server — see `PLAN_JSON_TOON_CONVERTER.md`. `collectOptions()` sends `<input type="number">` and `<select>` elements whose value is a plain integer/decimal as a JSON number (not a string) — required whenever a `data-option` maps to a Go `int` field (e.g. YAML Formatter's `indent`, QR Code's `size`); a `<select>`-only code path was missing this coercion for months and every such tool's web page silently 400'd (`invalid options object`) until caught while browser-testing the YAML Formatter `style` option — see `CHANGELOG.md`. | All 15 tool pages; each page template supplies only its tool-specific extra controls (e.g. Password Generator's checkboxes, Case Converter's mode buttons) around the shared partial. |

Illustrative signatures (finalized during implementation):

```go
// internal/apperr/apperr.go
package apperr

type Error struct {
    Code    string // stable machine-readable code, e.g. "INVALID_JSON"
    Message string
    Status  int    // HTTP status this maps to
}

func (e *Error) Error() string { return e.Message }
func New(status int, code, message string) *Error

var ErrEmptyInput = New(400, "EMPTY_INPUT", "input must not be empty")

// OneOf validates value is one of allowed; used by every Mode/Algorithm-style
// enum: json-format, url-encode, base64, case-convert, hash-gen, jwt, json-toon.
func OneOf[T comparable](value T, allowed ...T) error
```

```go
// internal/textio/textio.go
package textio

func Read(path string) ([]byte, error)     // "" or "-" => os.Stdin
func Write(path string, data []byte) error // "" or "-" => os.Stdout
```

**Why this matters**: without these shared packages, each of the 12 features would separately re-derive "read stdin-or-file", "write stdout-or-file", "validate this mode value", "map an error to an HTTP status and JSON error code", and "render an input/output panel with a copy button" — a large amount of near-identical boilerplate for what `PLANS/TASK.md` frames as small, focused tools. Centralizing them means a fix or improvement (a better stdin-detection heuristic, a new button style) is made once and every feature picks it up automatically; each `PLAN_<FEATURE>.md`'s Business logic/CLI/REST/Web UI sections then describe only what's genuinely tool-specific, since the shared plumbing is intentionally not re-explained per feature.

This complements, and does not replace, the purity rule already established above: `internal/tools/<name>` stays free of `net/http`/`cobra` imports; it may depend on `apperr` (a small, dependency-free error-shape package) but never on `textio`, `httpapi`, or `cli`, which sit one layer up and depend on `tools` — never the other way around.

## Library and framework choices

| Concern | Choice | Why |
|---|---|---|
| HTTP router | `github.com/go-chi/chi/v5` | Idiomatic, stays on stdlib `net/http` handler signatures (unlike gin's custom context), strong middleware composition, good fit for REST + server-rendered pages side by side. |
| CLI framework | `github.com/spf13/cobra` | De facto standard; native subcommand-per-feature model gives auto-generated `--help` at both root and subcommand level for free. |
| Metrics | `github.com/prometheus/client_golang` | Industry standard; exposes `/metrics` in Prometheus text format; pairs naturally with the Helm chart's `servicemonitor.yaml`. |
| Templating / frontend | `html/template` + `embed`, plain CSS (flexbox/grid) + vanilla JS | No Node build step; the whole UI ships inside the Go binary. A JS framework would add build complexity disproportionate to simple form-in/result-out tool pages. |
| QR code generation | `github.com/skip2/go-qrcode` (verify still maintained at implementation time; fallback `github.com/yeqown/go-qrcode`) | Pure Go, no cgo, produces PNG bytes directly. |
| JWT | `github.com/golang-jwt/jwt/v5` | Actively maintained successor to the archived `dgrijalva/jwt-go`. |
| YAML | `gopkg.in/yaml.v3` | De facto standard Go YAML library with indentation control. |
| Logging | `github.com/rs/zerolog` | JSON-structured by default, fluent strongly-typed field API avoids the mismatched-key-value footgun of variadic loggers, zero-allocation encoder keeps request-logging middleware cheap. See [Logging](#logging). |

## Dual mode: web (default) and CLI

```
mytoolkit                       # no args -> defaults to web server (serve)
mytoolkit serve [--host 0.0.0.0] [--port 8080] [--log-level info]
mytoolkit json-format --in file.json --out out.json [--minify]
mytoolkit json-tree --in file.json
mytoolkit yaml-format --in file.yaml --out out.yaml
mytoolkit password-gen --length 20 --symbols --numbers --uppercase
mytoolkit jwt --decode --token "..."
mytoolkit jwt --encode --claims claims.json --secret "..."
mytoolkit qrcode --text "https://example.com" --out qr.png
mytoolkit text-count --in file.txt
mytoolkit url-encode --decode --in "..."
mytoolkit hash-gen --algo sha256 --in file.txt
mytoolkit base64 --decode --in "..."
mytoolkit case-convert --mode title --in file.txt
mytoolkit json-toon --in file.json
mytoolkit --help                # general app help (auto-generated by cobra)
mytoolkit <subcommand> --help   # per-feature help (auto-generated by cobra)
```

The root cobra command's own `RunE` invokes the same logic as `serve`, so running the binary with zero arguments starts the web server — satisfying "web is the default mode, CLI is opt-in via a subcommand."

**Avoiding 3x duplicated logic**: each feature exposes a small typed API in `internal/tools/<name>`, e.g.:

```go
package hashgen

type Options struct { Algorithm string }
func Generate(input []byte, opts Options) (string, error)
```

- The REST handler — for most tools, just `handlers.Wrap("hash-gen", hashgen.Generate)` (see [Shared code and configuration reuse](#shared-code-and-configuration-reuse)) — decodes the request body into `Options` + input, calls `hashgen.Generate`, encodes the JSON response via the shared envelope.
- The CLI command — for most tools, built with `cli.newTextToolCommand(...)` over the same function — parses flags into the same `Options` via `textio`, calls `hashgen.Generate`, writes to `--out`/stdout.
- The web UI page's JS calls the REST endpoint via `fetch()` (using the shared `tool-common.js` helpers) — it is a client of the REST API, not a third Go code path. This keeps HTML mostly static and avoids duplicating server-side form handling.

This means a typical feature contributes exactly one new thing per layer — its `internal/tools/<name>` function and a short registry entry — while the REST/CLI/Web adapters are largely generic wiring, not hand-written per feature.

## REST API design

**URL scheme**: `/api/v1/tools/<tool-slug>`, one `POST` endpoint per feature (every operation transforms an input into an output, so a request body is appropriate even for read-feeling operations like formatting).

Tool slugs (kebab-case, shared with CLI subcommand names): `json-tree`, `json-format`, `yaml-format`, `password-gen`, `jwt`, `qrcode`, `text-count`, `url-encode`, `hash-gen`, `base64`, `case-convert`, `json-toon`, `yaml-to-json`, `json-to-yaml`, `k8s-validate`.

Request example (JSON Formatter):

```json
POST /api/v1/tools/json-format
{
  "input": "{\"a\":1}",
  "options": { "mode": "pretty", "indent": 2 }
}
```

Success envelope:

```json
{
  "success": true,
  "data": { "output": "{\n  \"a\": 1\n}" },
  "meta": { "tool": "json-format", "duration_ms": 0.42 }
}
```

Error envelope (standard HTTP status codes: 400 invalid input, 422 semantic validation error, 500 internal):

```json
{
  "success": false,
  "error": { "code": "INVALID_JSON", "message": "unexpected end of JSON input" }
}
```

**Binary exception**: the QR Code Generator returns `image/png` directly (not JSON-wrapped) so it can be used in `<img src>` and downloaded directly. This is the one deliberate deviation from the envelope, documented in `PLAN_QR_CODE_GENERATOR.md`.

**Versioning**: path-based (`/api/v1/...`). `GET /api/v1/tools` lists all tools (slug, name, description) from `internal/registry`, used by the frontend nav and CLI discovery.

## Metrics design

Registered in `internal/metrics/metrics.go` via `prometheus/client_golang`:

- `mytoolkit_http_requests_total{tool,method,status}` — Counter. Total request count per tool.
- `mytoolkit_http_request_duration_seconds{tool,method}` — Histogram. Response time per tool.
- `mytoolkit_tool_usage_total{tool}` — Counter, incremented once per successful tool invocation via the REST/web surface. **Scope note**: usage counting is scoped to the running web server process; CLI invocations are separate process runs and are not aggregated into this counter (documented assumption).
- Default `promhttp` process/Go collectors (`go_*`, `process_*`) included automatically.

**Exposed at**: `GET /metrics` (Prometheus text format via `promhttp.Handler()`), mounted outside `/api/v1` since it's operational, not a tool endpoint. Health-check and metrics requests themselves are excluded from `mytoolkit_http_requests_total` to avoid noise.

**Usage ranking**: Prometheus has no built-in ranking concept, so a derived endpoint is added: `GET /api/v1/metrics/ranking`, backed by a parallel in-memory `sync.Map[string]*atomic.Int64` updated alongside the Prometheus counter, sorted descending:

```json
{
  "ranking": [
    { "tool": "json-format", "count": 152, "rank": 1 },
    { "tool": "base64", "count": 98, "rank": 2 }
  ]
}
```

The same data is also derivable via PromQL (`topk(N, mytoolkit_tool_usage_total)`) for users who prefer querying Prometheus/Grafana directly — both routes are documented.

## Logging

Structured, JSON-formatted logging via `github.com/rs/zerolog`, replacing the earlier `log/slog` choice — same "no extra framework weight" spirit as the rest of the stack, but zerolog's fluent, strongly-typed field API (`.Str()`, `.Int()`, `.Err()`, …) avoids the "silently broken log entry" risk of variadic key-value APIs, and its zero-allocation JSON encoder keeps the request-logging middleware cheap. The framework-agnostic best practices from [dash0.com's structured logging guide](https://www.dash0.com/guides/logging-in-go-with-slog) are applied here using zerolog instead of `slog`:

- **Levels**: four levels used consistently — `debug`, `info`, `warn`, `error` (zerolog's `trace`/`panic`/`fatal` are not used). Controlled by `MYTOOLKIT_LOG_LEVEL` / `--log-level` from [Configuration](#configuration), applied once via `zerolog.SetGlobalLevel(...)` at startup. Expensive-to-compute debug fields are guarded with an enabled check (`if e := logger.Debug(); e.Enabled() { e.Interface("field", expensive()).Send() }`) so the cost is paid only when debug logging is active.
- **Output format & destination**: always JSON — no separate "pretty console" mode, since containerized/Kubernetes log collection (matching the Helm chart's deployment) expects structured JSON, not human-formatted text. Logs always go to `os.Stderr`, in both `serve` and CLI subcommand modes: CLI subcommands write their *tool output* to stdout/`--out` (see [Dual mode](#dual-mode-web-default-and-cli)), so application logs must never touch stdout or they would corrupt piped tool output (e.g. `mytoolkit hash-gen --in file | next-command`). Routing logs to stderr uniformly avoids a mode-conditional branch and matches standard CLI tool convention.
- **Field naming**: snake_case, consistent across the whole codebase — `tool`, `method`, `path`, `status`, `duration_ms`, `request_id`, `error`, `error_code`. These reuse the same vocabulary as the REST JSON envelope's `meta`/`error` objects from [REST API design](#rest-api-design) (`tool`, `duration_ms`, `code`→`error_code`), so a log line and its corresponding API response describe the same event with matching field names.
- **Correlation / request ID**: `internal/httpapi/middleware/logging.go` assigns a `request_id` per incoming request (via chi's `middleware.RequestID` or a UUID) and derives a per-request child logger — `logger := log.With().Str("request_id", id).Str("tool", slug).Logger()` — stored in the request's `context.Context` via `logger.WithContext(ctx)`. Downstream code pulls it back out with `zerolog.Ctx(ctx)`, so every log line tied to one HTTP request carries the same `request_id` (the "logger in context" pattern from the guide).
- **Error logging**: errors are logged with `.Err(err)` (zerolog's dedicated structured error field) plus `.Str("error_code", code)`, reusing the same error codes already defined in [REST API design](#rest-api-design) (`INVALID_JSON`, `UNSUPPORTED_ALGORITHM`, etc.) — never a bare formatted string, so code/message/cause stay structured instead of collapsing into one opaque message field.
- **Sensitive data**: request/response bodies are never logged wholesale. Fields that must never appear in a log line: JWT `secret`, Password Generator's generated `output`, and raw tool `input`/`output` payloads in general (a pasted input could itself be a real secret, e.g. a production JWT someone is decoding). Log lines only carry call metadata (`tool`, `status`, `duration_ms`, `request_id`) — the practical equivalent of the guide's `LogValuer` redaction pattern, enforced by only ever passing explicitly allow-listed scalar fields into a logging call, never an options/payload struct directly.
- **Middleware**: `internal/httpapi/middleware/logging.go` logs one structured line per completed request (`method`, `path`, `tool`, `status`, `duration_ms`, `request_id`) — `info` for 2xx/3xx, `warn`/`error` for 4xx/5xx. This is separate from, but correlated with, the Prometheus metrics recorded by `middleware/metrics.go` for the same request (see [Metrics design](#metrics-design)). `/healthz`, `/readyz`, and `/metrics` requests are logged at `debug` only, mirroring their exclusion from `mytoolkit_http_requests_total`.

## Health check design

The app is stateless (no DB, no external dependency at MVP), so a single check would suffice, but Kubernetes/Helm conventions benefit from the liveness/readiness split, so both are provided:

- `GET /healthz` — liveness. Returns `200 {"status":"ok"}`.
- `GET /readyz` — readiness. Returns `200 {"status":"ready"}` (reserved for future use if an external dependency is added later).

Both live outside `/api/v1` and are excluded from request-count metrics.

## Configuration

Every runtime-configurable value follows one precedence rule: **CLI flag > environment variable > built-in default**. Every environment variable that exists must be documented in three places, kept in sync: `docs/environment-variables.md` (source of truth, full descriptions), a table in `README.md` (user-facing summary), and `.env-example` at the repository root (copy-pasteable starting point for local dev / `docker-compose`). A configurable value is not considered done until all three are updated.

`internal/config.Load()` (see [Shared code and configuration reuse](#shared-code-and-configuration-reuse)) is the single place this precedence is implemented: if a flag was explicitly passed (`cmd.Flags().Changed("<name>")`), use it; else if the corresponding env var is set, use it; else fall back to the default below. `internal/cli/serve.go` calls `config.Load()` once and passes the resulting struct to `internal/app.New(cfg)` — no `os.Getenv` calls elsewhere in the codebase. No extra config library (e.g. viper) is introduced for this — it's a handful of values, resolved with plain `os.Getenv` plus cobra's `Changed()`.

| Variable | CLI flag (`serve`) | Default | Description |
|---|---|---|---|
| `MYTOOLKIT_HOST` | `--host` | `0.0.0.0` | Interface the HTTP server binds to. |
| `MYTOOLKIT_PORT` | `--port` | `8080` | TCP port the HTTP server listens on. |
| `MYTOOLKIT_LOG_LEVEL` | `--log-level` | `info` | zerolog level: `debug`, `info`, `warn`, `error`. See [Logging](#logging). |

`.env-example` (repo root) mirrors this table as `KEY=default` pairs with a one-line comment above each:

```
# Interface the HTTP server binds to.
MYTOOLKIT_HOST=0.0.0.0

# TCP port the HTTP server listens on.
MYTOOLKIT_PORT=8080

# zerolog level: debug, info, warn, error.
MYTOOLKIT_LOG_LEVEL=info
```

`docker-compose.yml` reads from a git-ignored `.env` (the user's local copy of `.env-example`) via `env_file: .env`. The Helm chart's `values.yaml` / `configmap.yaml` expose the same three keys under an `env:` map, so CLI flags, Docker Compose, and Helm all stay consistent with this one table. As future features introduce new configurable values, they are added to this table, to `docs/environment-variables.md`, to the README table, and to `.env-example` in the same change — never just one of the four.

## Theming and visual design system

Adopts a Material Design 3 (M3)–inspired visual language, matching the reference screenshots supplied by the user (`light.png` / `dark.png`, the Material Design 3 website's Color System page): fully rounded "pill" buttons and active-state highlights, a purple/lavender accent palette, and softly tinted card surfaces. This is implemented with hand-written CSS custom properties — no Material Web Components library is added, so the "no Node build step" decision in [Library and framework choices](#library-and-framework-choices) still holds; only the color tokens and shape rules are borrowed from M3.

**Color tokens** (`internal/web/static/css/theme.css`), M3 baseline purple palette:

| Token | Light value | Dark value | Used for |
|---|---|---|---|
| `--color-bg` | `#FEF7FF` | `#141218` | Page background |
| `--color-bg-alt` | `#F7F2FA` | `#211F26` | Nav bar / sidebar background |
| `--color-surface` | `#EFE3F6` | `#2B2930` | Cards, tool input/output panels |
| `--color-primary` | `#6750A4` | `#D0BCFF` | Links, focus rings, icon accents |
| `--color-primary-container` | `#EADDFF` | `#4F378B` | Pill button / active nav-item fill |
| `--color-on-primary-container` | `#21005D` | `#EADDFF` | Text/icon on a pill button or active nav item |
| `--color-text` | `#1D1B20` | `#E6E0E9` | Primary text |
| `--color-text-muted` | `#49454F` | `#CAC4D0` | Secondary text |
| `--color-border` | `#CAC4D0` | `#49454F` | Card/input borders |

**Shape rules**:
- Primary/CTA buttons and the active tab/nav-item indicator: fully rounded pill shape (`border-radius: 999px`), background `--color-primary-container`, text/icon `--color-on-primary-container`, bold label, generous horizontal padding (`0.75rem 1.5rem`+) — matches the "Overview" tab and the highlighted "Color system" nav entry in the reference screenshots.
- Secondary icon buttons (theme toggle, copy, reset, download): circular (`border-radius: 50%`), fixed square size, transparent background that fills with `--color-primary-container` on hover/focus — mirrors the circular theme-toggle button visible in the bottom corner of both reference screenshots.
- Cards (hero header on each tool page, tool input/output panels, the homepage's tool grid tiles): large rounded corners (`border-radius: 20–28px`), `--color-surface` background.
- Current-tool nav/menu item: pill highlight behind the label using `--color-primary-container`, same treatment as the reference's active sidebar entry.

**Scope note**: MyToolkit keeps its own simpler page structure (top nav bar + homepage tool grid + one page per tool) rather than replicating the reference's two-level docs-site sidebar chrome — only the color palette, pill/circle button shapes, and card styling are adopted, layered onto the layout already described below.

**Mechanics** (unchanged from the original plan, now using the tokens above):
- Colors defined as CSS custom properties in `internal/web/static/css/theme.css`, `:root` for light defaults, `[data-theme="dark"]` overrides.
- `<html data-theme="light|dark">` attribute drives the active theme.
- `internal/web/static/js/theme-toggle.js` (vanilla JS, no framework): reads `localStorage.getItem('theme')` on load, falls back to `prefers-color-scheme` if absent; the circular toggle button flips the attribute and persists the choice.
- The theme script runs as an inline `<script>` in `layout.html`'s `<head>`, before first paint, to avoid a flash of the wrong theme.
- No cookie is needed — theme is purely client-side, keeping the server stateless.
- `theme.css` + `theme-toggle.js` are built once as part of the shared layout and reused by all twelve tool pages.

**Navigation shell** (added after the initial theming pass, at the user's request — hamburger menu, back-to-home, footer, search): every page shares one chrome, defined once in `layout.html`, never per tool page.

- **Navigation drawer**: an M3 "modal navigation drawer" — `#nav-drawer-toggle` (hamburger icon button) opens `#nav-drawer` with a `#nav-scrim` behind it; `internal/web/static/js/nav.js` manages the `.open` class, `aria-hidden`/`aria-expanded`, and closes on scrim click, the drawer's own ✕, or Escape. The tool list inside is generated from `.Tools` (already passed to every template by `internal/web/handlers.go`), so adding a tool to `registry.Tools` automatically adds it to the drawer.
- **Search bar**: `#tool-search` filters `window.MYTOOLKIT_TOOLS` — JSON computed once at `init()` from `registry.All()` (`internal/web/handlers.go`'s `toolsJSON`) and embedded per-page as a `template.JS` value (`{{.ToolsJSON}}`, not re-escaped) so the client never re-fetches it. Matches case-insensitively against each tool's `name` **and** `description` — the same text rendered in its homepage card and tool-page hero card, per the requirement to index "the title card section." Entirely client-side, zero network calls, symmetric with the JSON to TOON Converter's client-side philosophy but applied to navigation instead of a tool's core function.
- **Back-to-Home button**: `.back-home-btn`, shown in the top bar whenever `.ActiveSlug` (set by `internal/web/handlers.go`, empty on the homepage, the tool's slug otherwise) is non-empty.
- **Footer**: static, unconditional, on every page — developer/contact info.

**M3 component pass** (also added after the initial theming pass, referencing `m3.material.io`'s spacing/buttons/icon-buttons/checkbox/dialogs/lists/navigation-bar/navigation-drawer/radio-button/search/switch/text-fields component pages): `theme.css` gained a 4dp spacing scale (`--space-1`…`--space-10`), state-layer opacity tokens (`--state-hover`/`--state-focus`/`--state-press`, applied via `color-mix()` for button/list-item hover-focus-press backgrounds instead of the earlier `filter: brightness()` approximation), and a shape scale (`--shape-xs`…`--shape-full`). New CSS should reuse these tokens rather than hardcoding new spacing/radius values. Two concrete M3 component distinctions applied: **checkboxes** (`.options-row input[type="checkbox"]`, for selecting several options from a set — e.g. Password Generator's charset toggles) are visually and semantically distinct from **switches** (`<label class="switch">`, for a single standalone on/off setting — e.g. `base64`/`url-encode`'s "Decode" toggle); text fields (`textarea`/`input`/`select`) share one focus-ring treatment; `.nav-list-item` (leading icon, hover state layer, `.active` pill) is reused by both the drawer and the search-results dropdown rather than styled twice.

**Bugs this pass surfaced and fixed** (see `CHANGELOG.md` for the full list): `registry.Tool` had no `json:"..."` tags, so `json.Marshal` emitted capitalized Go field names (`Slug`, `Name`, …); the search bar's JS read lowercase (`t.slug`, `t.name`, …), so every search silently returned zero results (no thrown error) until caught by an end-to-end browser test — fixed with explicit lowercase tags plus a regression test (`internal/registry/registry_test.go`). Separately, no favicon was ever defined, so browsers' automatic `/favicon.ico` request 404'd on every page load (a benign but real console error) — fixed with `internal/web/static/icons/favicon.svg` and a `<link rel="icon">` in `layout.html`. A third, unrelated bug was also found and fixed in this same pass: the README's `sequenceDiagram` Mermaid block used unescaped `<name>`/`<slug>` placeholders in unquoted participant/message text, which GitHub's Mermaid renderer failed to parse (`Expecting '()', ...` error) — fixed by rewording the diagram to avoid angle brackets entirely, rather than relying on HTML-entity escaping which didn't reliably work in that renderer.

## Testing strategy

- **Location**: `internal/tools/<name>/<name>_test.go`, colocated with the pure logic package.
- **Style**: table-driven tests (`[]struct{ name string; input ...; want ...; wantErr bool }` + `t.Run`), the idiomatic Go pattern.
- **Mocking**: none needed — `internal/tools/*` packages are pure functions (bytes/strings in, bytes/string/error out).
- **Adapter tests**: thin `httptest.NewRecorder()`-based tests for `internal/httpapi/handlers/*_handler.go` verifying envelope/status codes, and cobra command tests for `internal/cli/*.go` verifying flag parsing. These are secondary to the pure-function tests.
- **Coverage target**: ~80%+ on `internal/tools/*`, tracked via `go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`.
- **Golden files**: for larger structured output (JSON tree, YAML formatting), use `testdata/` golden files per package.

## Docker, Helm and Makefile shape

**Dockerfile** (multistage):
1. `builder`: `golang:<latest-stable>` (or `-alpine`), `CGO_ENABLED=0 go build` producing a static binary (embedded templates/static assets, no runtime asset copy needed).
2. Final stage: `gcr.io/distroless/static-debian12` nonroot variant, `COPY --from=builder`, `USER nonroot`, `ENTRYPOINT ["/mytoolkit"]`, default `CMD ["serve"]`.
3. `EXPOSE 8080`.

**docker-compose.yml**: single `mytoolkit` service, built from the local Dockerfile, port `8080:8080`, environment variables loaded via `env_file: .env` (the user's local copy of `.env-example`) — see [Configuration](#configuration) for the full variable table.

**Helm chart** (`helm/mytoolkit/`), inspired by the structure at `gitlab.com/aeciopires/kube-pires/-/tree/master/helm-chart` (production-grade convention):

```
Chart.yaml
values.yaml            # image repo/tag, replicas, resources, ingress host, autoscaling min/max, metrics/serviceMonitor toggle
templates/
  deployment.yaml       # liveness/readiness probes -> /healthz, /readyz
  service.yaml
  ingress.yaml           # optional, values.ingress.enabled
  hpa.yaml                # optional, values.autoscaling.enabled
  serviceaccount.yaml
  configmap.yaml          # non-secret env config
  secret.yaml              # placeholder for future secret-backed config
  servicemonitor.yaml    # optional, values.metrics.serviceMonitor.enabled, scrapes /metrics
  _helpers.tpl
  NOTES.txt
tests/
  test-connection.yaml    # helm test hook hitting /healthz
```

- `helm/mytoolkit` is a standard Helm chart (`Chart.yaml`, `values.yaml`, `templates/`). Probes point at `/healthz`; pod annotations pre-declare `prometheus.io/scrape|port|path` for `/metrics` auto-discovery.
- Validate changes with `make helm-lint` (or `helm lint helm/mytoolkit`) and `helm template helm/mytoolkit` (re-render with `--set ingress.enabled=true --set autoscaling.enabled=true` to exercise the optional templates) before committing.
- `helm/mytoolkit/README.md` is **generated, not hand-written** — it's produced by `make helm-docs` (wraps [helm-docs](https://github.com/norwoodj/helm-docs)) from `helm/mytoolkit/README.md.gotmpl` plus the `# -- description` head-comments in `values.yaml`. If you add/rename/remove a `values.yaml` key, add a matching `# -- ...` comment directly above it and re-run `make helm-docs` rather than editing the chart's `README.md` by hand.
- `make helm-install` / `make helm-uninstall` wrap `helm upgrade --install` / `helm uninstall` against the `NAMESPACE` variable (defaults to the app name).
- Example of README.md.gotmpl https://github.com/aeciopires/my-world-cup-app/blob/main/charts/my-world-cup-app/README.md.gotmpl

**Makefile targets**:

```
build              # go build -o bin/mytoolkit ./cmd/mytoolkit
run                # go run ./cmd/mytoolkit serve
test               # go test ./...
test-verbose       # go test -v ./...
coverage           # go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out
lint               # golangci-lint run
fmt                # gofmt -s -w .
vet                # go vet ./...
check-tools        # verify required dev/runtime tools are installed, list what's missing (see below)
deps-check         # go mod verify && go mod tidy -diff / govulncheck (Go module dependency hygiene, not tool prerequisites)
docker-build       # docker build -t mytoolkit:latest .
docker-run         # docker run -p 8080:8080 mytoolkit:latest
compose-up / compose-down
helm-lint          # helm lint helm/mytoolkit
helm-template      # helm template helm/mytoolkit
kind-load          # kind load docker-image mytoolkit:latest --name kind-king-multinodes
helm-install-kind  # helm upgrade --install mytoolkit helm/mytoolkit --kube-context kind-king-multinodes
helm-test          # helm test mytoolkit --kube-context kind-king-multinodes
clean              # remove bin/, coverage.out
help               # self-documenting target list
```

**`check-tools` design**: iterates a fixed list of required CLI tools — `go`, `git`, `docker`, `docker compose` (plugin, not the legacy `docker-compose` binary), `helm`, `kubectl`, `kind`, `golangci-lint` — using `command -v <tool>` for each, and prints a per-tool `OK`/`MISSING` line. At the end it prints a summary list of any missing tools and their install-doc URL, and exits non-zero if anything is missing so it can also gate CI/other targets (e.g. `docker-build` can depend on `check-tools`). This directly satisfies the `TASK.md` requirement to "check dependencies" — it is distinct from `deps-check`, which validates the Go module graph (`go.mod`/`go.sum`), not the developer's local toolchain. Sketch:

```makefile
REQUIRED_TOOLS := go git docker helm kubectl kind golangci-lint

.PHONY: check-tools
check-tools:
	@missing=""; \
	for tool in $(REQUIRED_TOOLS); do \
		if command -v $$tool >/dev/null 2>&1; then \
			echo "[OK]      $$tool"; \
		else \
			echo "[MISSING] $$tool"; \
			missing="$$missing $$tool"; \
		fi; \
	done; \
	if command -v docker >/dev/null 2>&1 && ! docker compose version >/dev/null 2>&1; then \
		echo "[MISSING] docker compose (plugin)"; \
		missing="$$missing docker-compose-plugin"; \
	fi; \
	if [ -n "$$missing" ]; then \
		echo ""; echo "Missing tools:$$missing"; \
		exit 1; \
	fi; \
	echo ""; echo "All required tools are installed."
```

This is a design sketch for the implementation phase, not final Makefile source — the real target will be written and tested when the Makefile is authored (Phase 0, step 8 in [Suggested build sequencing](#suggested-build-sequencing)).

## Documentation structure

Every feature's documentation is written as plain markdown under `docs/` — never only inside its `PLAN_<FEATURE>.md`, which is the design doc, not the shipped reference. Per [Go module layout](#go-module-layout), each of the 12 features gets a fixed triplet of files: `docs/api/<slug>.md` (REST reference), `docs/cli/<slug>.md` (CLI reference), `docs/testing/<slug>.md` (unit test reference) — 36 files total, plus the shared `docs/environment-variables.md` covering the [Configuration](#configuration) table. A feature's Phase 1 work (see [Suggested build sequencing](#suggested-build-sequencing)) is not done until its triplet exists.

`README.md` must carry a **Documentation** section with a table linking every feature to its three doc files, so none of this reference material is orphaned outside of `docs/`:

```markdown
## Documentation

| Feature | API reference | CLI reference | Testing reference |
|---|---|---|---|
| JSON Tree Viewer | [docs/api/json-tree.md](docs/api/json-tree.md) | [docs/cli/json-tree.md](docs/cli/json-tree.md) | [docs/testing/json-tree.md](docs/testing/json-tree.md) |
| JSON Formatter | [docs/api/json-format.md](docs/api/json-format.md) | [docs/cli/json-format.md](docs/cli/json-format.md) | [docs/testing/json-format.md](docs/testing/json-format.md) |
| YAML Formatter | [docs/api/yaml-format.md](docs/api/yaml-format.md) | [docs/cli/yaml-format.md](docs/cli/yaml-format.md) | [docs/testing/yaml-format.md](docs/testing/yaml-format.md) |
| Password Generator | [docs/api/password-gen.md](docs/api/password-gen.md) | [docs/cli/password-gen.md](docs/cli/password-gen.md) | [docs/testing/password-gen.md](docs/testing/password-gen.md) |
| JWT Encode/Decode | [docs/api/jwt.md](docs/api/jwt.md) | [docs/cli/jwt.md](docs/cli/jwt.md) | [docs/testing/jwt.md](docs/testing/jwt.md) |
| QR Code Generator | [docs/api/qrcode.md](docs/api/qrcode.md) | [docs/cli/qrcode.md](docs/cli/qrcode.md) | [docs/testing/qrcode.md](docs/testing/qrcode.md) |
| Character, Word & Line Counter | [docs/api/text-count.md](docs/api/text-count.md) | [docs/cli/text-count.md](docs/cli/text-count.md) | [docs/testing/text-count.md](docs/testing/text-count.md) |
| URL Encode/Decode | [docs/api/url-encode.md](docs/api/url-encode.md) | [docs/cli/url-encode.md](docs/cli/url-encode.md) | [docs/testing/url-encode.md](docs/testing/url-encode.md) |
| Hash Generator | [docs/api/hash-gen.md](docs/api/hash-gen.md) | [docs/cli/hash-gen.md](docs/cli/hash-gen.md) | [docs/testing/hash-gen.md](docs/testing/hash-gen.md) |
| Base64 Encode/Decode | [docs/api/base64.md](docs/api/base64.md) | [docs/cli/base64.md](docs/cli/base64.md) | [docs/testing/base64.md](docs/testing/base64.md) |
| Case Converter | [docs/api/case-convert.md](docs/api/case-convert.md) | [docs/cli/case-convert.md](docs/cli/case-convert.md) | [docs/testing/case-convert.md](docs/testing/case-convert.md) |
| JSON to TOON Converter | [docs/api/json-toon.md](docs/api/json-toon.md) | [docs/cli/json-toon.md](docs/cli/json-toon.md) | [docs/testing/json-toon.md](docs/testing/json-toon.md) |
| YAML to JSON Converter | [docs/api/yaml-to-json.md](docs/api/yaml-to-json.md) | [docs/cli/yaml-to-json.md](docs/cli/yaml-to-json.md) | [docs/testing/yaml-to-json.md](docs/testing/yaml-to-json.md) |
| JSON to YAML Converter | [docs/api/json-to-yaml.md](docs/api/json-to-yaml.md) | [docs/cli/json-to-yaml.md](docs/cli/json-to-yaml.md) | [docs/testing/json-to-yaml.md](docs/testing/json-to-yaml.md) |
| Kubernetes YAML Validator | [docs/api/k8s-validate.md](docs/api/k8s-validate.md) | [docs/cli/k8s-validate.md](docs/cli/k8s-validate.md) | [docs/testing/k8s-validate.md](docs/testing/k8s-validate.md) |

See also: [Environment variables](docs/environment-variables.md).
```

This table is written once during Phase 2 ([Suggested build sequencing](#suggested-build-sequencing)) after all 11 doc triplets exist, and its links are re-checked for dead/renamed targets during Phase 5's per-feature verification pass.

## Version control hygiene (.gitignore)

The repository's current `.gitignore` is a generic Visual Studio/.NET template (429 lines, `bin/x64/`, `*.sln.docstates`, etc.) left over from scaffolding and irrelevant to a Go project — it does not even ignore compiled Go binaries. It is replaced outright, in Phase 0 ([Suggested build sequencing](#suggested-build-sequencing)), with a Go- and tooling-focused version so build artifacts, local secrets, and editor/OS cruft never get committed:

```gitignore
# Compiled binaries
/bin/
/mytoolkit
*.exe
*.test
*.dll
*.so
*.dylib

# Go build/test/coverage artifacts
*.out
coverage.out
coverage.html

# Go workspace/vendor (module-mode is used; vendoring is not, but ignore defensively)
/vendor/
go.work
go.work.sum

# Local environment overrides — .env-example stays tracked, .env never does
.env
.env.*
!.env-example

# Docker
*.tar

# Helm packaging/dependency artifacts
*.tgz
/helm/**/charts/
/helm/**/Chart.lock

# Local kubeconfig exports (e.g. from kind, Makefile helm-* targets)
kubeconfig*.yaml
*.kubeconfig

# Editor/IDE
.vscode/
.idea/
*.swp
*.swo

# OS
.DS_Store
Thumbs.db

# Logs
*.log
```

This list is revisited whenever a new tool/dependency is introduced that produces local artifacts (e.g. a linter cache directory) — the rule is the same as [Configuration](#configuration)'s: nothing that's a generated binary, secret, or machine-local file is ever committed, and `.gitignore` is updated in the same change that introduces the artifact-producing tool.

## Full repository tree

See [Go module layout](#go-module-layout) above for `cmd/`, `internal/`, `docs/`, `.skills/`, `helm/`; combined with `Dockerfile`, `docker-compose.yml`, `Makefile`, `go.mod`/`go.sum`, and the top-level markdown files, this is the complete tree to be created during implementation.

## Suggested build sequencing

**Phase 0 — shared scaffolding** (build once, before any feature):
1. `go mod init`, replace `.gitignore` with the Go-focused version from [Version control hygiene](#version-control-hygiene-gitignore), base skeleton, `cmd/mytoolkit/main.go` + `internal/cli/root.go` (empty root cobra command, `--help` works).
2. `internal/httpapi/router.go` + `internal/app/app.go` — chi router boots, `serve` subcommand starts an HTTP server; zerolog global logger initialized per [Logging](#logging) (JSON, `os.Stderr`, level from `MYTOOLKIT_LOG_LEVEL`).
3. `internal/httpapi/health.go` — `/healthz`, `/readyz`.
4. `internal/metrics/metrics.go` + `middleware/metrics.go` + `middleware/logging.go` — `/metrics` wired, empty registry; request-ID + structured request logging wired per [Logging](#logging).
5. `internal/web/templates/layout.html` + `static/css/theme.css` + `static/js/theme-toggle.js` + index page — dark/light toggle works on an empty shell, nav placeholder driven by `internal/registry`.
6. `internal/httpapi/response.go` — shared JSON envelope + error codes.
7. `internal/apperr`, `internal/textio`, `internal/config`, `internal/httpapi/handlers/generic.go`, `internal/cli/shared.go`, `internal/web/templates/partials/tool-panel.html`, `internal/web/static/js/tool-common.js` — the shared packages from [Shared code and configuration reuse](#shared-code-and-configuration-reuse), built and unit-tested before any feature consumes them.
8. Makefile core targets (`build`, `run`, `test`, `lint`) and a minimal Dockerfile — verify the empty shell builds/runs/dockerizes before adding features.

**Phase 1 — features, one at a time**, in increasing complexity/dependency order: Base64 → URL Encode/Decode → Hash Generator → Character/Word/Line Counter → Case Converter → JSON Formatter → YAML Formatter → JSON Tree Viewer → Password Generator → JWT Encode/Decode → QR Code Generator → JSON to TOON Converter (built last: it reuses JSON Tree Viewer's order-preserving decode technique and introduces the new client-side-only web pattern, so it benefits from every other tool's pattern already being proven first).

For each feature: `internal/tools/<name>` + tests → `internal/httpapi/handlers/<name>_handler.go` + route registration + `internal/registry` entry → `internal/cli/<name>.go` subcommand → `internal/web/templates/tools/<name>.html` + fetch()-based JS → `docs/api/<name>.md` + `docs/cli/<name>.md` + `docs/testing/<name>.md` → `.skills/<name>/SKILL.md`.

**Phase 2 — cross-cutting polish**: full `README.md` rewrite (architecture, Mermaid diagrams overall + per feature, directory structure, usage, environment variables table, and the [Documentation](#documentation-structure) table linking every feature's `docs/api|cli|testing/<slug>.md` triplet), `docs/environment-variables.md`, `CHANGELOG.md`, `ROADMAP.md`, `CLAUDE.md`.

**Phase 3 — packaging & deployment**: finalize Dockerfile and `docker-compose.yml`, author the full Helm chart, `helm lint`, deploy to `kind-king-multinodes`, run `helm test`, verify `/healthz`, `/metrics`, and a couple of tool endpoints via `kubectl port-forward`.

**Phase 4 — contribution docs**: `CONTRIBUTING.md`, modeled on the two-section structure of `raw.githubusercontent.com/aeciopires/learning-istio/refs/heads/main/CONTRIBUTING.md` — a "Contributing" section (fork/clone/branch/commit/PR/sync workflow with bash examples) and a "Tip" section (recommended editor/extensions), translated to English per the "all docs in English" requirement.

**Phase 5 — verification & bug-fixing pass**: after Phases 1–4 are complete, re-verify every `PLAN_<FEATURE>.md` specification against what was actually built, one feature at a time:
1. Run `go test ./internal/tools/<name>/...` and confirm every table-driven case passes.
2. Execute the exact CLI example from `docs/cli/<name>.md` against the built binary and confirm the output matches what's documented.
3. Call the REST endpoint with the exact example request from `docs/api/<name>.md` (`curl` or equivalent) and confirm the response matches what's documented (status code, envelope shape, values).
4. Exercise the feature's web page manually and confirm it behaves as described in the plan's Web UI section, in both light and dark theme, using the pill/circle button and color-token conventions from [Theming and visual design system](#theming-and-visual-design-system).
5. Confirm a Mermaid workflow diagram for the feature exists (embedded in `docs/api/<name>.md`, linked from `README.md`) showing its request lifecycle across CLI/REST/Web (input → validation → `internal/tools/<name>` call → output/error), and check it against the behavior actually observed in steps 1–4 — redraw it if it has drifted from the real implementation.
6. Any discrepancy found — a failing test, a CLI/REST response that doesn't match its own documented example, incorrect web behavior, a wrong Mermaid diagram — is treated as a bug and fixed in the implementation, not papered over by rewriting the doc to match broken behavior (the doc is only corrected when the doc itself, not the code, was wrong). A feature is not considered complete until this pass is clean.

This phase also produces and verifies the overall application Mermaid diagram for `README.md`'s Architecture section (the shared request path: CLI/Web/REST client → router/CLI parser → `internal/tools/<name>` → response, with metrics/logging middleware shown in the path), checked the same way against the actual shared scaffolding.

## Verification

- `go build ./...` and `go vet ./...` succeed after Phase 0.
- Each feature's `go test ./internal/tools/<name>/...` passes before its handler/CLI adapter is wired in.
- `make docker-build && make docker-run`, then `curl localhost:8080/healthz`, `curl localhost:8080/metrics`, and one `POST /api/v1/tools/<slug>` call succeed.
- `helm lint helm/mytoolkit` passes; `make kind-load && make helm-install && make helm-test` succeeds against the real `kind-kind-multinodes` context (cluster name `kind-multinodes` — see [Additional instructions](#additional-instructions-received-during-implementation)), followed by a manual `kubectl port-forward` check of `/healthz` and one tool endpoint.
- Phase 5's per-feature checklist (unit tests, CLI example, REST example, web behavior, Mermaid diagram accuracy) is clean for all twelve features, and every bug it surfaced has been fixed in code before the implementation is considered done. For JSON to TOON Converter specifically, this includes confirming the web page issues zero network requests for the live conversion (inspect the browser network tab) and that the JS and Go implementations agree on the shared fixture table from `PLAN_JSON_TOON_CONVERTER.md`.
- `README.md`'s [Documentation](#documentation-structure) table has one row per feature with working links to all 45 `docs/api|cli|testing/<slug>.md` files plus `docs/environment-variables.md`; no dead or missing links.

## Additional instructions received during implementation

Requirements added by the user after the initial planning pass, during the implementation session itself. Each is incorporated into the relevant section above; this section is the changelog of *why* those sections say what they say, kept for traceability.

- **Multi-arch (AMD64 + ARM64, Linux and macOS)**: the application and Docker image must run on both architectures. Addressed by keeping every dependency pure Go (no cgo) and building the [Dockerfile](#docker-helm-and-makefile-shape) with `--platform=$BUILDPLATFORM` on the builder stage plus `GOOS=$TARGETOS GOARCH=$TARGETARCH` cross-compilation, validated with `docker buildx build --platform linux/amd64,linux/arm64`. macOS isn't a container target (Docker on macOS runs Linux containers under the hood); the same cross-compiled `linux/arm64` image runs on Apple Silicon Docker Desktop.
- **Helm implementation must cover the `helm-docs`/Prometheus-annotation/install-uninstall specification** already present in this document's [Docker, Helm and Makefile shape](#docker-helm-and-makefile-shape) section (pod annotations `prometheus.io/scrape|port|path`, `helm/mytoolkit/README.md` generated via `make helm-docs` from `README.md.gotmpl` + `values.yaml`'s `# --` comments, `make helm-install`/`make helm-uninstall` against a `NAMESPACE` variable defaulting to the app name) — confirmed implemented and exercised against a real cluster.
- **README screenshots**: `README.md` carries a Screenshots section with images stored in `images/` at the repo root, captured from the real running application (headless Chrome), not mockups. The web layout's inline theme script also accepts a `?theme=light|dark` query-string override (in addition to `localStorage` and `prefers-color-scheme`) specifically so screenshots/demos can force a theme deterministically.
- **Source tree reorganization**: all Go/HTML/CSS/JS source (`cmd/`, `internal/`) was moved under `src/`, with `src/go.mod`/`src/go.sum` making `src/` its own Go module root — this keeps every existing `github.com/aeciopires/mytoolkit/internal/...` import path unchanged (no source-file edits needed) while separating application source from planning (`PLANS/`), documentation (`docs/`, `.skills/`), and deployment (`helm/`, `Dockerfile`, `docker-compose.yml`) at the repo root. The [Go module layout](#go-module-layout) tree and the Makefile (`cd $(SRC) && go ...`) reflect this.
- **`make docker-push`**: an interactive Makefile target that prompts for a Docker Hub username, password/access token (hidden input via `read -s`), and target repository, then builds and pushes a multi-arch image via `docker buildx build --platform ... --push`. The password is piped directly into `docker login --password-stdin` — never passed as a CLI argument (which would leak it via `ps`), never echoed, never written to disk — and `docker logout` runs afterward. Documented in `README.md`, `CONTRIBUTING.md`, and `CLAUDE.md`; this target requires an interactive terminal and must never be scripted or invoked non-interactively with embedded credentials.
- **`--version`/`-v` flag and a shared `VERSION` file**: the repo-root `VERSION` file is the single source of truth for the application version, read by both the Go build and the Docker build so they can never drift apart. `make build` embeds it via `-ldflags -X github.com/aeciopires/mytoolkit/internal/version.Version=$(VERSION)`; `make docker-build`/`docker-buildx`/`docker-push` pass it as a `--build-arg VERSION` (used both as the image tag and inside the Dockerfile's own `go build -ldflags` step). `cobra.Command.Version` on the root command exposes it as `mytoolkit --version`/`-v`, printed before any subcommand logic runs (including the default-to-`serve` behavior).
- **JSON to TOON Converter (12th tool)**: added per `PLAN_JSON_TOON_CONVERTER.md`, inspired by `scalevise.com/json-toon-converter`. Two requirements given together: (1) the web page's live conversion must be 100% client-side (no data sent to the server, verifiable by inspecting network requests), matching the reference site's privacy claim; (2) explicitly reaffirmed afterward, this tool's REST endpoint and CLI subcommand are **not optional** — they get a full Go implementation exactly like every other tool, identical in status to the other eleven. The resolution is two independent implementations of the same TOON-encoding algorithm: `internal/tools/jsontoon` (Go; backs REST + CLI) and `internal/web/static/js/json-toon.js` (vanilla JS; backs only the web page). This introduces a new reusable convention in `tool-common.js` — a `data-client-side` attribute on `.tool-panel` that opts a page out of the shared fetch-based `run()` wiring (copy/reset/download wiring still applies) — available to any future tool with the same no-network-call requirement. See [Shared code and configuration reuse](#shared-code-and-configuration-reuse).
- **YAML to JSON Converter and JSON to YAML Converter (13th and 14th tools)**: added per `PLAN_YAML_TO_JSON_CONVERTER.md`/`PLAN_JSON_TO_YAML_CONVERTER.md`, at the user's explicit request to use [`sigs.k8s.io/yaml`](https://github.com/kubernetes-sigs/yaml) — the library Kubernetes itself uses to accept YAML manifests as JSON. Both are fully generic (`handlers.Wrap`/`newTextToolCommand`, no client-side deviation, no bespoke wiring) — the interesting decisions were library-behavior ones, not architectural ones: `yaml-to-json` uses `YAMLToJSONStrict` (not the lenient `YAMLToJSON`) to reject duplicate YAML keys instead of silently keeping one; `json-to-yaml` pre-validates its input with `encoding/json.Unmarshal` before calling `sigs.k8s.io/yaml.JSONToYAML`, because that function parses its input as YAML internally and was verified to silently accept invalid JSON (trailing commas, unquoted keys, even a bare comment line) rather than reject it. Both tools inherit the library's YAML 1.1 boolean-resolution behavior (`yes`/`no`/`y`/`n`/`on`/`off` unquoted → JSON booleans, the "Norway problem") — verified directly and documented, not treated as a bug, since it's the library's own documented behavior (mirroring the same well-known gotcha already documented for `yaml-format`, see `PLAN_YAML_FORMATTER.md`). This is a deliberate second, different choice of YAML library within the same app: `yaml-format` uses `gopkg.in/yaml.v3` because it needs `yaml.Node` to preserve comments/anchors/key order for a same-format round trip; these two converters use `sigs.k8s.io/yaml` because a cross-format converter has no such round trip to preserve, and `sigs.k8s.io/yaml`'s JSON-shaped decode/encode is the right tool for that job.
- **Kubernetes YAML Validator (15th tool)**: added per `PLAN_K8S_YAML_VALIDATOR.md`, at the user's explicit request to use `sigs.k8s.io/yaml` again, this time to validate that YAML is shaped the way the Kubernetes API server requires (non-empty `apiVersion`/`kind`, object-shaped `metadata`), not just that it's syntactically valid YAML. This is the app's first tool combining **both** YAML libraries deliberately: `gopkg.in/yaml.v3` splits a `---`-separated multi-document stream into individual documents (reusing `yaml-format`'s already-proven technique, since `sigs.k8s.io/yaml` has no multi-document API of its own), then `sigs.k8s.io/yaml.YAMLToJSONStrict` validates each isolated document exactly as kubectl/the API server would decode it. It's also the app's second tool (after JSON Tree Viewer/Text Counter/Password Generator/JWT/QR Code) whose business logic returns a structured Go type instead of a plain string, requiring bespoke REST/CLI wiring — but notably its **web page needed zero bespoke JavaScript**: `tool-common.js`'s `run()` already had a `JSON.stringify(json.data, null, 2)` fallback for when a REST response has no `data.output` field (added defensively when the generic handler was built, but never actually exercised by a real tool until this one), and that fallback alone produces a perfectly readable report. Explicitly scoped to not validate against any specific resource's full schema (Deployment, Service, a CRD, ...) — that needs the full Kubernetes OpenAPI schema database, out of scope for a lightweight dependency-light utility; the web page and docs say so directly rather than implying more coverage than it has.
- **Header content centered to the page's content column**: the sticky `.topnav` bar's children were pinned to the far-left/far-right edges of the viewport on wide screens while `.content` below it was centered at `max-width: 960px` — visually misaligned, confirmed with a screenshot at 1440px width before fixing. Fixed with a `.topnav-inner` wrapper (same `max-width`/`margin: 0 auto`) around the header's children, added to every page since it lives in the shared `layout.html`. Explicitly verified this doesn't affect the nav drawer's overlap-avoidance (a separate concern the user flagged proactively): the drawer is `position: fixed` with a scrim, entirely unrelated to the header's flex layout, and stayed correct in before/after screenshots. See `.skills/web-ui-shell/SKILL.md`.
- **JWT Encode/Decode: multi-algorithm support**: at the user's explicit request to use `golang-jwt/jwt` (already a dependency, but only its HMAC methods were wired up) "to allow support to many algorithms," `jwttool.Encode`/`Decode` now support RSA (`RS256/384/512`), RSA-PSS (`PS256/384/512`), and ECDSA (`ES256/384/512`) and EdDSA, alongside the original HMAC (`HS256/384/512`) — 13 algorithms total, added to the web page as a `<select>` combobox per the request, with `HS256` kept first/default per the request's explicit "the default algorithm must be the [one] currently used." Required a new `key` parameter (PEM text) alongside the existing `secret` (HMAC-only) — see `PLAN_JWT_ENCODE_DECODE.md`'s Business logic section and `.skills/jwt/SKILL.md` for the secret-vs-key-by-algorithm-family design and how verification derives which one applies from the token's own `alg` header rather than a caller hint.
- **Observability stack: Grafana dashboard**: `docker-compose.yml`'s `prometheus`/`grafana` services (added ahead of this request, outside this document's original scope) previously had no dashboard — Grafana came up with an empty UI requiring manual setup. Added `observability/mytoolkit-dashboard.json` (per the request, saved directly in `observability/`, not a nested subfolder) covering every metric the app exposes: the three custom `mytoolkit_*` metrics from [Metrics design](#metrics-design), plus the Go runtime/process metrics that come for free via `promhttp.Handler()`'s default registry (goroutines, memory, GC, CPU, open FDs, network I/O). Every query was verified against a real running binary's `/metrics` output before being written, not assumed from typical `client_golang` conventions (e.g. `go_gc_duration_seconds`'s quantile labels are `0/0.25/0.5/0.75/1`, not `0.99`). Auto-provisioned via `observability/grafana/provisioning/{datasources,dashboards}/*.yml`, mounted into the `grafana` service — a fresh `docker compose up` needs zero manual data-source/dashboard setup. Verified end-to-end by actually running the stack, generating real traffic, and logging into Grafana's UI (not just validating the JSON). See `.skills/observability/SKILL.md` for two gotchas found and fixed this way: a table panel leaking Prometheus's own `instance`/`job` labels because its query had no `by (...)` aggregation, and a Docker single-file bind-mount inode issue where edits to the dashboard file didn't reach the container without restarting it.
- **`src/` renamed to `app/`**: at the user's explicit request. The directory itself (`mv src app`), the Makefile's `SRC` variable, and the Dockerfile's builder stage (`COPY app/...`, `WORKDIR` renamed from `/app/src` to `/build` to avoid the confusing appearance of `/app/app`) were updated and re-verified (`make build`/`test`/`vet`, plus a full `docker build`, all pass from the new location). No Go import paths changed — the module path is `github.com/aeciopires/mytoolkit` regardless of which host directory contains `go.mod`, so this was a pure filesystem/tooling rename, not a refactor. Every `src/`-path reference across `CLAUDE.md`, `README.md`, `CONTRIBUTING.md`, `ROADMAP.md`, this document, `docs/testing/*.md` (`cd src && go test` → `cd app && go test`), and every `.skills/*/SKILL.md` was swept and updated — except genuine historical record entries (e.g. this section's own "Source tree reorganization" bullet above, and `CHANGELOG.md`'s `[1.0.0]` entry describing the original move into `src/`), which describe what was true *at that point in time* and are deliberately left as-is rather than rewritten. `<img src="...">` HTML-attribute examples in QR Code's docs/skill were correctly left untouched — same substring, unrelated meaning.

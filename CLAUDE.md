<!-- TOC -->

- [CLAUDE.md](#claudemd)
  - [Project](#project)
  - [Repository layout](#repository-layout)
  - [Commands](#commands)
  - [Adding a new tool](#adding-a-new-tool)
  - [MCP Server](#mcp-server)
  - [Web UI shell](#web-ui-shell)
  - [Versioning](#versioning)
  - [Conventions](#conventions)

<!-- TOC -->

# CLAUDE.md

Guidance for Claude Code (and other agents) working in this repository.

## Project

MyToolkit is a Go application exposing 15 developer utilities as a web UI, a REST API, a CLI, and an MCP server — all four surfaces share one pure-function implementation per tool, with one documented exception (JSON to TOON Converter's web page, see Conventions below). See `PLANS/PLAN_ARCHITECTURE.md` for the full architecture rationale and `PLANS/PLAN_<FEATURE>.md` for each tool's design.

## Repository layout

Go/HTML/CSS/JS source lives under `app/` (its own Go module root — `app/go.mod`), separate from planning (`PLANS/`), documentation (`docs/`, `.skills/`), and deployment (`helm/`, `Dockerfile`, `docker-compose.yml`) files at the repo root.

```
app/
  cmd/mytoolkit/main.go       entrypoint; also carries swag's @title/@version/... general API annotations
  docs/                       swaggo/swag-generated OpenAPI spec (docs.go, swagger.json, swagger.yaml) — generated, do not hand-edit; regenerate with `make swagger-gen`
  internal/
    apperr/                   shared error type + OneOf[T] validator
    textio/                   shared --in/--out read/write helpers
    config/                   flag > env > default resolution
    response/                 shared JSON success/error envelope (leaf package)
    registry/                 tool metadata: slug, name, description
    cli/                      cobra commands; one file per tool, self-registers via init(); also carries swag @Router annotations per tool handler
    httpapi/                  chi router, health, generic REST handler wrapper, /swagger/* UI route
    mcp/                      MCP server: one file per tool, self-registers via init(), same pattern as cli/
    metrics/                  Prometheus collectors + usage ranking
    web/                      html/template pages, embedded CSS/JS
    tools/<name>/             pure business logic, one package per tool + its _test.go
```

Note: `app/docs/` (generated Swagger spec) is unrelated to the repo-root `docs/` (hand-written per-tool `api|cli|testing/<name>.md` reference docs) — same name, different purpose, don't confuse them. Likewise, the repo-root `mcp/` folder is documentation and MCP client config examples only (`mcp/README.md`, `mcp/examples/`) — the MCP server's actual Go code lives at `app/internal/mcp/`, since Go's `internal/` visibility rule means a root-level package can't import `app/internal/tools` directly.

## Commands

Run from `app/` (or use the Makefile targets at the repo root, which `cd` into `app/` for you):

```
cd app
go build ./...
go vet ./...
go test ./...
go run ./cmd/mytoolkit serve --port 8080
```

Makefile (repo root): `make build`, `make test`, `make run`, `make lint`, `make check-tools`, `make docker-build`, `make helm-lint`, `make swagger-gen`, etc. Run `make help` for the full list.

`make docker-push` interactively prompts for Docker Hub username/password-or-token/repository and pushes a multi-arch image. It requires a human at a terminal (hidden password prompt) and publishes to a real public registry — never invoke it non-interactively or on the user's behalf without explicit, in-the-moment confirmation.

## Adding a new tool

Follow the pattern of an existing simple tool (e.g. `base64` or `case-convert`):

1. `app/internal/tools/<name>/<name>.go` — pure function `func Do(input []byte, opts Options) (string, error)`, returning `*apperr.Error` for known failure modes. Colocate `<name>_test.go`.
2. `app/internal/registry/registry.go` — add a `Tool{Slug, Name, Emoji, Description}` entry.
3. `app/internal/cli/<name>.go` — `init()` registers both the cobra subcommand (`newTextToolCommand`, unless the tool doesn't fit the text-in/text-out shape — see Password Generator/JWT/QR Code/Text Counter/JSON Tree for bespoke wiring) and the REST handler (`handlers.Wrap`) via `registerToolHandler`.
4. `app/internal/web/templates/tools/<name>.html` — `{{define "content"}}` + `{{template "tool-panel" .}}`, plus an optional `{{define "tool-options"}}` block for extra form controls (`data-option name="..."` attributes are auto-collected into the REST request's `options` object by `tool-common.js`). If the tool's web page must never call the server (a hard product requirement, not a default), set `ClientSide: true` on its `registry.Tool` entry — this renders `data-client-side` on `.tool-panel`, which makes `tool-common.js` skip its fetch-based wiring so the page's own `{{define "extra-scripts"}}` inline script can own input → output conversion instead (see `json-toon` for the reference implementation).
5. `app/internal/mcp/<name>.go` — `init()` registers one (or more, see JWT) `sdkmcp.AddTool[In, Out]` call whose handler calls the same `internal/tools/<name>` function as the CLI/REST wiring above. See [MCP Server](#mcp-server) below and `.skills/mcp/SKILL.md`.
6. `docs/api/<name>.md`, `docs/cli/<name>.md`, `docs/testing/<name>.md`, `.skills/<name>/SKILL.md` — see any existing tool's docs for the expected shape. `docs/api/<name>.md` must include a `## Workflow` section with a Mermaid diagram of the request lifecycle (see any existing tool's doc for the pattern).
7. Add the tool to `README.md`'s feature list and Documentation table, and to `mcp/README.md`'s "Available tools" table.
8. Add `swaggo/swag` annotations (`@Summary`/`@Description`/`@Tags tools`/`@Accept`/`@Produce`/`@Param`/`@Success`/`@Failure`/`@Router`) above the tool's REST handler function — if it's wired via the generic `handlers.Wrap` (no named handler function exists yet), add a small named wrapper (`func <name>Handler() http.HandlerFunc { return handlers.Wrap(...) }`) to carry the annotation; see `.skills/swagger/SKILL.md`. Then run `make swagger-gen` and commit the regenerated `app/docs/`.

## MCP Server

`mytoolkit mcp` is a 4th surface over `internal/tools/<name>`, alongside web/REST/CLI — not a separate service. `app/internal/mcp/server.go` builds an `mcp.Server` (`github.com/modelcontextprotocol/go-sdk/mcp`) and every `app/internal/mcp/<name>.go` registers itself via `init()`, mirroring the `toolHandlers` self-registration pattern `internal/cli/serve.go` already uses for REST handlers. Each tool defines a small `In` struct (JSON schema auto-inferred from its `json`/`jsonschema` tags) mirroring its `Options` type, and a `handle<Name>` function calling straight into the same pure function CLI/REST call — no duplicated business logic, ever.

`app/internal/cli/mcp.go` wires the `mcp` subcommand, supporting two transports selected by `--transport`/`MYTOOLKIT_MCP_TRANSPORT` (`config.LoadMCP`, same flag > env > default precedence as `serve`): `stdio` (default — the standard way local MCP clients like Claude Desktop/Claude Code run a server) and `http` (`mcp.NewStreamableHTTPHandler`, for a shared/remote endpoint, `--port`/`MYTOOLKIT_MCP_PORT`, default `8081`). Logs already go to stderr only (see Conventions) — for `stdio` this is a hard correctness requirement, not just style: any stray stdout write corrupts the JSON-RPC stream.

Two per-tool exceptions worth knowing before touching `app/internal/mcp/`:
- `json-tree`'s `Out` type is `any`, not `jsontree.Node` — `jsontree.Node` is self-referential (`Children []Node`) and the SDK's schema inference (`google/jsonschema-go`) panics on cyclic struct types when building an *output* schema. `any` skips schema generation entirely (documented `AddTool` behavior) while the node is still returned as structured JSON content.
- `qrcode`'s handler returns `*sdkmcp.CallToolResult` directly with an `ImageContent` block (`Out` is `any`) instead of relying on the SDK's automatic JSON-of-`Out` content — the one binary-output MCP tool, matching the REST endpoint's own `image/png` exception.

Full docs (installation, client config examples, Mermaid workflow diagram): `mcp/README.md`. Tests: `app/internal/mcp/*_test.go` (per-tool `handle<Name>` unit tests) plus `app/internal/mcp/server_test.go` (a real client/server round trip over `mcp.NewInMemoryTransports()`, the MCP-surface equivalent of the REST layer's `httptest`-based handler tests).

**Metrics**: `app/internal/mcp/metrics.go`'s `metricsMiddleware`, registered once via `server.AddReceivingMiddleware` in `NewServer`, records `mytoolkit_mcp_requests_total{method,status}`/`mytoolkit_mcp_request_duration_seconds{method}` for every JSON-RPC method plus `mytoolkit_mcp_tool_calls_total{tool,status}`/`mytoolkit_mcp_tool_call_duration_seconds{tool}` for `tools/call` specifically and `mytoolkit_mcp_sessions_total` on `initialize` — one middleware covers all 16 tools, no per-handler instrumentation. Kept as a metric family separate from `mytoolkit_http_*`/`mytoolkit_tool_usage_total` (distinct client population, see Metrics design in `PLAN_ARCHITECTURE.md`). Only reachable via `GET /metrics` when `--transport http` (mounted on the same port as the MCP endpoint through one `http.ServeMux` in `internal/cli/mcp.go`) — `stdio` has no listening port to scrape. Wired into `observability/prometheus.yml` (job `mytoolkit-mcp`), `helm/mytoolkit/templates/mcp-deployment.yaml`'s Prometheus scrape annotations plus the optional `mcp-servicemonitor.yaml`, and a dedicated "MCP Server" row in `observability/mytoolkit-dashboard.json` — see `.skills/mcp/SKILL.md` and `.skills/observability/SKILL.md`.

## Web UI shell

`internal/web/templates/layout.html` (shared by every page) provides, once, for free:
- **Navigation drawer** (`internal/web/static/js/nav.js` + `.nav-drawer`/`.scrim` in `app.css`) — a hamburger button (`#nav-drawer-toggle`) opens a modal drawer listing every tool from `.Tools`, with a scrim that closes it on click, plus Escape-to-close. Follows the M3 "modal navigation drawer" pattern.
- **Search bar** (`#tool-search`, also wired in `nav.js`) — filters `window.MYTOOLKIT_TOOLS` (JSON-embedded server-side from `registry.All()`, see `internal/web/handlers.go`'s `toolsJSON`) client-side against each tool's `name`/`description`, live, no network call. Clicking a result or pressing Enter navigates to it.
- **Back-to-Home button** (`.back-home-btn`) — rendered whenever `.ActiveSlug` is non-empty (i.e. on any tool page, never on the homepage).
- **Footer** (`.site-footer`) — static developer/contact info, on every page.

`registry.Tool` has explicit `json:"..."` tags (lowercase) precisely because it's marshaled for the search bar's `window.MYTOOLKIT_TOOLS` *and* for `GET /api/v1/tools` — if you add a field to `Tool`, add a matching lowercase `json` tag, or you'll reintroduce the bug in "Conventions" below.

M3 design tokens (spacing, state-layer opacities, shape radii) live in `theme.css` as CSS custom properties (`--space-*`, `--state-*`, `--shape-*`) — reuse them in new component CSS rather than hardcoding `px`/`rem` values, to keep the app visually consistent with `https://m3.material.io/styles/spacing/overview` and the linked M3 component pages for buttons, icon-buttons, checkboxes, switches, radio buttons, text fields, lists, and navigation drawer/bar.

## Versioning

The repo-root `VERSION` file is the single source of truth for the application version. `make build` embeds it into the binary via `-ldflags -X .../internal/version.Version=$(VERSION)` (exposed via `mytoolkit --version`/`-v`); `make docker-build`/`docker-buildx`/`docker-push` pass it as a Docker `--build-arg VERSION` (used both to tag the image and to embed the same ldflag inside the container build); `make helm-docs` runs `helm-set-appversion` first, which `sed`-rewrites `helm/mytoolkit/Chart.yaml`'s `appVersion` field to match. Bump `VERSION` (and `CHANGELOG.md`) together when cutting a release — don't hardcode a version string anywhere else, including in `Chart.yaml` (`helm-set-appversion` will overwrite a hand-edited `appVersion` on the next `make helm-docs` anyway).

## Conventions

- `internal/tools/<name>` packages must never import `net/http`, `cobra`, or any other `internal/` package except `apperr`.
- Error codes and HTTP status are defined once via `apperr.New(status, code, message)` — never construct ad hoc error strings in handlers.
- Logs are structured JSON (zerolog) to stderr, always — never to stdout, since CLI tool output uses stdout.
- Prefer the generic `handlers.Wrap` / `newTextToolCommand` helpers; only write bespoke REST/CLI wiring when a tool's request/response shape genuinely doesn't fit (documented per-tool in `.skills/<name>/SKILL.md`).
- Every example in `docs/api|cli/<name>.md` must be copy-paste-verified against the running binary before being committed — don't hand-type expected output (including error messages, which are parser-dependent and easy to get subtly wrong). A prior pass shipped docs with a nonexistent request field, a stray unencoded-newline artifact, a hand-typed JSON format that didn't match `json.MarshalIndent`, and invented error message text — all found later by re-running the documented commands verbatim. Re-running every doc example against the real binary is exactly how those were caught; do this whenever you touch a tool's docs.
- A tool may ship a client-side JS mirror under `internal/web/static/js/<name>.js` (instead of calling the REST API from its own web page) only when a no-network-call guarantee is an explicit product requirement — the pure `internal/tools/<name>` Go package is still mandatory and backs REST/CLI exactly as normal. Such a tool must: (1) state the dual-implementation trade-off explicitly in its `.skills/<name>/SKILL.md`, (2) keep both implementations tested against one shared fixture table (a Go `_test.go` file plus a documented headless-browser parity check — see `docs/testing/json-toon.md`), and (3) set `ClientSide: true` on its `registry.Tool` entry to opt into the `data-client-side` convention rather than hand-rolling page-specific wiring.
- Any Go struct that gets `json.Marshal`ed for consumption by JS or an external client (e.g. `registry.Tool`) must have explicit lowercase `json:"..."` tags on every field. `registry.Tool` shipped without them for months — harmless while it was template-only, until the search bar started reading `t.slug`/`t.name` from the marshaled JSON and silently got `undefined` for every field (capitalized Go field names, no thrown error, just zero search results). Caught by an end-to-end browser test, not `go vet` or a unit test — see `internal/registry/registry_test.go`'s `TestToolJSONFieldsAreLowercase` for the regression test now guarding this.
- Use a **switch** (`<label class="switch"><input type="checkbox" ...></label>`, see `base64`/`url-encode`'s "Decode") for a single, standalone on/off setting; use a plain checkbox (`.options-row input[type="checkbox"]`, no extra class, see Password Generator's charset toggles) when several independent options are presented together as a set. This follows M3's checkbox-vs-switch guidance (`https://m3.material.io/components/switch/overview` vs `.../checkbox/overview`) — don't default everything to one or the other.
- New interactive UI (buttons, dialogs, lists, etc.) added anywhere in `internal/web` should be verified with a real browser (Playwright against the actual running binary, or at minimum a headless-Chrome screenshot), not just "the CSS looks right in the file" — two real bugs in this app's UI (a JSON-field-case mismatch breaking search, a missing favicon causing a console 404) were only caught this way, not by reading the source.
- The REST API is documented at `/swagger/index.html` (swaggo/swag). `swag` resolves a cross-package type in an annotation (e.g. `@Success 200 {object} pkg.Type`) only if the annotated file actually imports `pkg` — this is why the shared response DTOs it points at (`ToolSuccessResponse`, `ToolErrorResponse`, `ToolMeta`) live in `internal/cli` (imported everywhere handlers are registered) rather than `internal/httpapi` (not imported by `internal/cli`). See `.skills/swagger/SKILL.md` before adding or changing annotations.

---
name: mcp
description: Add or modify the MCP (Model Context Protocol) server (app/internal/mcp/, the `mytoolkit mcp` subcommand, mcp/README.md). Trigger on "add MCP tool", "mytoolkit mcp", "MCP server", "add a new tool" (MCP wiring is part of that checklist too).
---

# MCP Server

`mytoolkit mcp` is a 4th surface over `internal/tools/<name>`, alongside web/REST/CLI — not a separate service, not duplicated business logic. `app/internal/mcp/server.go` builds an `mcp.Server` (`github.com/modelcontextprotocol/go-sdk/mcp`); every `app/internal/mcp/<name>.go` registers itself via `init()` appending to a package-level `toolAdders` slice, the same self-registration pattern `internal/cli/serve.go` already uses for REST `toolHandlers`.

## One handler function per tool, testable directly

Each tool defines an `In` struct (JSON Schema auto-inferred by `google/jsonschema-go` from `json`/`jsonschema` struct tags — the `jsonschema` tag is a plain description string, not a `key=value` mini-language; there's no `jsonschema:"enum=..."` or `jsonschema:"default=..."` syntax, just prose) mirroring the tool's `Options`, plus a `handle<Name>` function with the exact signature `sdkmcp.ToolHandlerFor[In, Out]` expects:

```go
func handleBase64(_ context.Context, _ *sdkmcp.CallToolRequest, in base64In) (*sdkmcp.CallToolResult, textOut, error) {
	out, err := base64enc.Process([]byte(in.Input), base64enc.Options{...})
	if err != nil {
		return nil, textOut{}, toolErr(err)
	}
	return nil, textOut{Output: out}, nil
}
```

Because the request/request-context arguments accept `nil` (an interface and a pointer), every `handle<Name>` is directly unit-testable with no server/transport involved: `handleBase64(nil, nil, base64In{Input: "hi"})` — see `app/internal/mcp/*_test.go`. `common.go`'s `textOut{Output string}` is the shared MCP output shape for every tool that just returns transformed text, mirroring the REST envelope's `{"output": ...}`.

## Errors are plain `error` values — the SDK does the rest

A handler's returned `error` is automatically converted by the SDK into `CallToolResult{IsError: true, Content: [...text...]}` — don't build that envelope by hand. `errors.go`'s `toolErr(err)` just re-attaches the `apperr.Error` code the REST JSON envelope's `"code"` field also carries (`*apperr.Error.Error()` returns only the message, not the code) — everything else passes through unchanged.

## jwt splits into two MCP tools, not one with a mode field

Unlike the REST endpoint's single `/api/v1/tools/jwt` route with an options `mode` discriminator, `app/internal/mcp/jwt.go` registers `jwt_decode` and `jwt_encode` separately — two clean, independently-documented input shapes are a better fit for how an MCP client (or the LLM behind it) picks a tool than an overloaded field. Both still call straight into `jwttool.Decode`/`jwttool.Encode`, no new logic.

## qrcode returns image content, not text

`qrcode`'s handler skips the auto-JSON-content path entirely (`Out` is `any`) and returns `*sdkmcp.CallToolResult` directly with `Content: []sdkmcp.Content{&sdkmcp.ImageContent{Data: png, MIMEType: "image/png"}}` — the one binary-output MCP tool, matching the REST endpoint's own `image/png` exception (see `.skills/qrcode/SKILL.md`).

## json-tree's Out type is `any`, not jsontree.Node — cyclic schema panic

`jsontree.Node` is self-referential (`Children []Node`). Registering it as an `AddTool` `Out` type parameter makes `google/jsonschema-go` try to build an output JSON Schema for a cyclic Go struct, which **panics at server-construction time** (`cycle detected for type jsontree.Node`), not at call time — this was hit once already registering all 16 tools. The fix is `Out any`: per `AddTool`'s documented behavior, an `any` output type skips schema generation entirely, while the node is still returned as structured JSON content since `CallToolResult.Content` is populated from whatever value the handler returns.

## k8s-validate: a semantically-invalid document is not a tool error

Mirrors the REST handler exactly: a YAML document that parses fine but fails Kubernetes' own requirements (missing `apiVersion`, for example) is a **successful** call (`isError: false`) whose structured `Result.Valid` is `false` for that document. Only a hard YAML syntax error or "no documents at all" becomes a Go error / `isError: true`. Don't "fix" this into always erroring on `Valid == false` — it would break parity with the REST/CLI surfaces.

## password-gen: no MCP-side defaulting, matching REST (not CLI)

The REST handler applies zero defaults to `password.Options`'s booleans — the client must explicitly set at least one of `lowercase`/`uppercase`/`numbers`/`symbols` to `true` or the call fails with `NO_CHARSET_SELECTED`. Only the CLI applies convenience flag defaults (`--lowercase` defaulting to `true`, etc.), because `pflag` can distinguish "flag not passed" from "flag passed as false" — a JSON `bool` field can't make that distinction, so the MCP `In` struct (like the REST request) intentionally has no defaulting logic. Document required options in the tool's `Description`/field `jsonschema` tags instead of trying to fake CLI-style defaults.

## stdio transport: never write to stdout outside the protocol

`app/internal/cli/mcp.go` routes zerolog to stderr exactly like `serve.go` does — for `stdio` this isn't just the repo's usual logging convention, it's a hard correctness requirement. Any stray `fmt.Println`/log-to-stdout in a `handle<Name>` function (or anything it calls) corrupts the JSON-RPC stream and breaks every MCP client talking to the process. If you need to debug a handler, log via `zerolog/log` (stderr) or write to a file — never `os.Stdout`.

## Metrics: one middleware, not per-handler instrumentation

`app/internal/mcp/metrics.go`'s `metricsMiddleware` is registered once, in `server.go`'s `NewServer` via `server.AddReceivingMiddleware(metricsMiddleware)` — it wraps `sdkmcp.MethodHandler`, which sees **every** JSON-RPC method (`initialize`, `tools/list`, `tools/call`, notifications, ...), not just tool calls. This is deliberately *not* wired per-`handle<Name>` function: adding a 17th tool gets metrics for free, no `metrics.MCPToolCallsTotal.WithLabelValues(...)` call to remember inside the handler.

The middleware distinguishes success/error two ways, both required — a Go-level `err != nil` (protocol-level failure) is not the same thing as a tool-level failure: `sdkmcp.ToolHandlerFor` converts a handler's returned `error` into `CallToolResult{IsError: true, ...}` *without* an `err` at the middleware's level (see "Errors are plain error values" above), so `isToolError(result)` type-asserts to `*sdkmcp.CallToolResult` and checks `.IsError` explicitly — checking only `err != nil` would silently miscount every tool-level error (`INVALID_OPTION`, `NO_CHARSET_SELECTED`, etc.) as `status="success"`.

Metrics recorded (`internal/metrics/metrics.go`, prefixed `mytoolkit_mcp_` to stay a separate family from the REST/web surface's `mytoolkit_http_*`/`mytoolkit_tool_usage_total` — see `.skills/observability/SKILL.md`): `mytoolkit_mcp_requests_total{method,status}` / `mytoolkit_mcp_request_duration_seconds{method}` for every JSON-RPC method, `mytoolkit_mcp_tool_calls_total{tool,status}` / `mytoolkit_mcp_tool_call_duration_seconds{tool}` specifically for `tools/call` (tool name pulled from `req.GetParams().(*sdkmcp.CallToolParamsRaw).Name`), and `mytoolkit_mcp_sessions_total` (incremented on a successful `initialize`, no labels).

`/metrics` is only mounted for `--transport http` — `internal/cli/mcp.go`'s http branch builds an `http.ServeMux` (`/` → `sdkmcp.NewStreamableHTTPHandler`, `/metrics` → `promhttp.Handler()`) instead of passing the bare MCP handler straight to `http.ListenAndServe`. `stdio` has no listening port to hang a second path off of — this isn't a TODO, Prometheus's pull model has nothing to scrape from a local child process. Don't add a dedicated always-on metrics port for stdio "to be safe"; it wouldn't reflect the deployable scenario (docker-compose `mcp` profile / Helm `mcp.enabled`) this is actually wired into.

## Adding a new tool also means adding its MCP wiring

Per `CLAUDE.md`'s "Adding a new tool" checklist: create `app/internal/mcp/<name>.go` following one of the patterns above, add it to `mcp/README.md`'s "Available tools" table, and confirm `TestListToolsCoversEveryTool` in `app/internal/mcp/server_test.go` still passes (it asserts every expected tool name is present, the MCP-surface equivalent of `internal/httpapi/swagger_test.go`'s `TestSwaggerSpecCoversEveryTool`).

## Verification

```
cd app && go build ./... && go vet ./... && go test ./internal/mcp/... -v
./bin/mytoolkit mcp --help
```

For an actual protocol round trip (not just the in-memory-transport tests), drive the real binary over stdio with a throwaway MCP client using `mcp.CommandTransport{Command: exec.Command("./bin/mytoolkit", "mcp")}`, or the community `npx @modelcontextprotocol/inspector` if Node is available — see `mcp/README.md`'s "Examples" section for a captured transcript to compare against. For the `http` transport, `mytoolkit mcp --transport http --port 8081` plus a raw `curl` `initialize` request (see `mcp/README.md`) is enough to confirm the handler is wired correctly.

For the metrics themselves: `go test ./internal/mcp/... -run TestMetrics -v` covers the middleware in isolation (deltas via `prometheus/client_golang/prometheus/testutil.ToFloat64`, not absolute values — see `metrics_test.go`'s comment on why). To verify the full scrape path end-to-end, see `.skills/observability/SKILL.md`'s "Verifying changes" section (`docker compose --profile mcp up -d`, a real `initialize`/`tools/call` `curl` sequence, then querying Prometheus/Grafana directly rather than trusting the JSON).

Plan: `PLANS/PLAN_ARCHITECTURE.md`'s Dual mode section, and `CLAUDE.md`'s MCP Server section.

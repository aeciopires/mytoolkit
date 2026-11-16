---
name: observability
description: Modify the observability stack — metrics (Prometheus), logs (Loki via Grafana Alloy), traces (OpenTelemetry → Alloy → Tempo + Jaeger), the Grafana dashboards/data sources under observability/, the docker-compose.yml services, or the app-side instrumentation (internal/telemetry, internal/logging, internal/metrics, tracing middlewares). Trigger on "add a metric to the dashboard", "update Grafana dashboard", "add Prometheus scrape config", "add a span/attribute", "change log format", "Loki/Tempo/Jaeger/Alloy config".
---

# Observability stack

Three signals, one Compose stack, three dashboards:

```
mytoolkit ──/metrics scrape (OpenMetrics + trace_id exemplars)──────────► Prometheus ─┐
    │  JSON logs on stderr ─► Docker ─► Alloy (parse JSON) ─► Loki ──────────────────┤
    └─ OTLP/HTTP spans ─────────────► Alloy ─┬─► Tempo ─(span metrics, remote write)─►├─► Grafana
                                             └─► Jaeger ──────────────────────────────┘
```

## Files

```
docker-compose.yml                                 mytoolkit, mytoolkit-mcp (profile "mcp"), alloy, loki, tempo, jaeger, prometheus, grafana
observability/
  prometheus.yml                                   scrape jobs: mytoolkit, mytoolkit-mcp, + self-monitoring (prometheus, loki, tempo, alloy, jaeger, grafana)
  alloy/config.alloy                               Docker log discovery -> JSON parse -> level normalize -> Loki; OTLP receiver -> batch -> Tempo + Jaeger
  loki/loki.yaml                                   single-binary, filesystem, tsdb v13, structured metadata on, 7d retention
  tempo/tempo.yaml                                 single-binary, local storage, metrics generator (span-metrics, service-graphs) -> Prometheus remote write
  mytoolkit-dashboard.json                         "MyToolkit — Application Metrics" (uid mytoolkit-overview) — hand-formatted, edit textually
  mytoolkit-logs-dashboard.json                    "MyToolkit — Logs" (uid mytoolkit-logs)
  mytoolkit-traces-dashboard.json                  "MyToolkit — Traces" (uid mytoolkit-traces)
  grafana/provisioning/datasources/datasource.yml  Prometheus/Loki/Tempo/Jaeger with fixed uids and cross-links
  grafana/provisioning/dashboards/dashboard.yml    loads *.json from /var/lib/grafana/dashboards (folder "MyToolkit")
```

The dashboards stay at the top level of `observability/` (the original request asked for the dashboard "in the observability directory"); each is bind-mounted individually in `docker-compose.yml` — add a mount line when adding a dashboard.

## App-side instrumentation

| Signal | Code | Notes |
|---|---|---|
| Metrics | `internal/metrics` | `metrics.Handler()` serves `/metrics` with **OpenMetrics enabled** (the only format carrying exemplars; `promhttp.Handler()` doesn't). Record histograms with `metrics.Observe(ctx, obs, v)` so a sampled request attaches a `trace_id` exemplar. |
| Traces | `internal/telemetry`, `internal/httpapi/middleware/tracing.go`, `internal/mcp/tracing.go`, `internal/cli/tracing.go` | Opt-in via the standard `OTEL_*` env vars (`OTEL_EXPORTER_OTLP_ENDPOINT` turns it on; OTLP/HTTP only). REST: `mw.Tracing` (otelhttp, root of the chi router, span renamed to `METHOD /route/{pattern}` + `http.route`; `/metrics`, probes and `/static/` are skipped) and `mw.ToolSpan` (`tool <slug>` child, error status only for 5xx). MCP: `tracingMiddleware` (`mcp <method>` / `mcp tools/call <tool>`, `gen_ai.tool.name`, `mytoolkit.tool`). |
| Logs | `internal/logging` | `logging.Setup` makes **every** stderr line JSON: zerolog, stdlib `log` (incl. `http.Server.ErrorLog`), `log/slog`, OTel SDK errors. Common fields `level`, `service`, `version`, `time` (RFC 3339 nano), `message`; request lines add `trace_id`/`span_id`. Fatal `serve`/`mcp` errors are logged as JSON (`logFatal`) and not re-printed as text by `main`. |

Gotchas that each cost a debugging session (all regression-tested):

- **MCP spans must not inherit `ctx`'s span.** The go-sdk runs method handlers on a *session-scoped* context, so inheriting it put every call of a session into one giant trace. `tracingMiddleware` starts a new root unless the client propagated a parent: `params._meta.traceparent` first, then the HTTP `traceparent` header (`req.GetExtra().Header`). There is deliberately no `otelhttp` wrapper on the MCP HTTP handler.
- **Typed-nil params.** `notifications/initialized` arrives with a typed-nil `*InitializedParams` inside a non-nil `Params` interface; calling `GetMeta()` on it panicked and, since the SDK doesn't recover handler goroutines, **killed the whole MCP server**. `metaOf` checks with reflection, and `callRecovering` turns any handler panic into a JSON-RPC error.
- **otelhttp renames the span after the handler** using `r.Pattern`/the span-name formatter, overwriting a `SetName` done inside the handler — name spans through `WithSpanNameFormatter` (it's called again after routing). otelhttp does *not* set `http.route`; `mw.Tracing` does.
- **`slog.SetDefault` hijacks the stdlib logger** (redirects `log` output into the slog handler at INFO). `logging.Setup` calls it *before* `stdlog.SetOutput`.
- `zerolog.Logger.Write` emits events **without a `level` field**; stdlib lines go through `levelWriter` instead.

## Data source uids are fixed: `prometheus`, `loki`, `tempo`, `jaeger`

Every panel's `datasource.uid` must match `datasource.yml` exactly, or panels show "datasource not found" on a fresh `docker compose up`. Don't switch to `${DS_*}` template variables — those are for dashboards imported by hand through the UI; these are provisioned and must work with zero clicks. `$$` in `datasource.yml` escapes Grafana's provisioning env interpolation (`$${__value.raw}` is stored as `${__value.raw}`).

Cross-links configured there: Prometheus exemplars (`trace_id`) → Tempo; Loki `trace_id` structured metadata → Tempo (derived field, `matcherType: label`); Tempo/Jaeger → Loki (`tracesToLogsV2`, query `{service="<service.name>"} | trace_id="<id>"` — relies on the Compose service name equalling `OTEL_SERVICE_NAME`); Tempo → Prometheus (`tracesToMetrics`, service map).

## Logs pipeline (Alloy → Loki)

- Only containers of this Compose project are shipped (`COMPOSE_PROJECT` env → `discovery.relabel` keep rule). Labels: `service` (Compose service), `container`, `compose_project`, `level`, `tool`. `trace_id`/`span_id` are **structured metadata, never labels** (unbounded cardinality).
- Every stack component is configured to log JSON too: Prometheus `--log.format=json`, Loki `-log.format=json`, Tempo `server.log_format: json` (Tempo 3 has **no** `-log.format` flag — the container crash-loops with "flag provided but not defined"), Alloy `logging { format = "json" }`, Grafana `GF_LOG_CONSOLE_FORMAT=json`, Jaeger `--set=service.telemetry.logs.encoding=json` (its very first "application version" line is still plain text, printed before the logger exists). Exception: Tempo's embedded Prometheus remote-write agent (metrics generator WAL) always logs logfmt (`level=INFO ...`), whatever `log_format` says — a `stage.regex` fallback extracts `logfmt_level` so those lines still get a level. One `stage.json` parses everything else; `stage.template` normalizes level spellings (`INFO`, `warning`, `eror`, `dbug`, ...) to `error|warn|info|debug|fatal`, which the Logs dashboard's Level variable relies on.

## Traces pipeline

- Alloy fans the same spans out to Tempo and Jaeger, so one trace ID opens in both.
- Tempo 3.1's config differs from 2.x: e.g. `metrics_generator.traces_storage` and the `local-blocks` processor are rejected ("field ... not found") — check against the release's `example/docker-compose/single-binary/tempo.yaml`. `generate_native_histograms: classic` keeps `traces_spanmetrics_latency_bucket` as classic buckets for `histogram_quantile`.
- Span-metrics labels (verified from real data): `service`, `span_name`, `span_kind` (`SPAN_KIND_SERVER`/`SPAN_KIND_INTERNAL`), `status_code` (`STATUS_CODE_UNSET`/`STATUS_CODE_ERROR`).
- **Jaeger 2.23 serves only `api_v3`** (the legacy `/api/services`, `/api/traces` are gone). Grafana's Jaeger plugin uses v3 for health/services/trace-by-ID only with the `jaegerEnableGrpcEndpoint` feature toggle (set in Compose) and still uses the legacy API for **search** and dependencies (plugin source: "TODO: enable routing to gRPC when ready", jaeger#7594). So the Traces dashboard shows Jaeger traces **by ID** only; search lives in the Jaeger UI. Re-check when bumping either version.

## Grounding queries in real data

Every query was written against real output (`curl localhost:8080/metrics`, Loki/Tempo/Prometheus HTTP APIs), not from memory — e.g. `go_gc_duration_seconds` only has `quantile` labels `0`, `0.25`, `0.5`, `0.75`, `1` (no `0.99`). Custom app metrics (`internal/metrics/metrics.go`):
- REST/web: `mytoolkit_http_requests_total{tool,method,status}`, `mytoolkit_http_request_duration_seconds{tool,method}` (histogram, exemplars), `mytoolkit_tool_usage_total{tool}` — same data as `GET /api/v1/metrics/ranking`.
- MCP (recorded by `internal/mcp`'s `metricsMiddleware`): `mytoolkit_mcp_requests_total{method,status}`, `mytoolkit_mcp_request_duration_seconds{method}`, `mytoolkit_mcp_tool_calls_total{tool,status}`, `mytoolkit_mcp_tool_call_duration_seconds{tool}`, `mytoolkit_mcp_sessions_total` (counts `initialize` and the stateless protocol's `server/discover`). Kept separate from `mytoolkit_http_*` (distinct client population, see `PLAN_ARCHITECTURE.md`). Only scrapeable with `--transport http`.

Table panels: aggregate away Prometheus's own labels (`sum(...) by (tool)`), otherwise `instance`/`job`/`__name__` columns clutter the table — found by screenshotting, not by reading the JSON.

## Editing dashboards

- `mytoolkit-dashboard.json` is hand-formatted (one-line targets): edit it textually so diffs stay small; re-serializing it with `json.dump` rewrites ~2000 lines.
- The logs/traces dashboards are plain Grafana JSON (`schemaVersion` 39); edit them directly or in the Grafana UI and export.
- Single-file bind mounts track the host file's **inode**: an editor that saves via write-new-file-then-rename leaves the container on the old content, and neither `updateIntervalSeconds` nor `POST /api/admin/provisioning/dashboards/reload` picks the change up. Run `docker compose restart grafana`.

## Verifying changes

`python3 -m json.tool` is necessary but nowhere near sufficient. Run the stack (`make compose-up`, plus `docker compose --profile mcp up -d mytoolkit-mcp` for MCP panels), generate REST and MCP traffic (MCP: a real `initialize` → `notifications/initialized` → `tools/call` sequence, see `mcp/README.md`), then:

1. Confirm the data landed: Prometheus `/api/v1/query` and `/api/v1/query_exemplars`, Loki `/loki/api/v1/label/<name>/values` and `query_range`, Tempo `/api/search` + `/api/v2/traces/<id>`, Jaeger `/api/v3/services`.
2. Run **every panel's query through Grafana** (`POST /api/ds/query`, basic auth `admin:admin`) and require data for each — this caught the Jaeger search 404 and a trace-by-ID panel using the wrong `queryType` (`traceId`, not `traceql`). The Tempo `serviceMap` query type is resolved in the browser only (the backend answers "unsupported query type"), so:
3. Screenshot all three dashboards with Playwright (tall viewport — Grafana scrolls inside its own container, so `full_page` alone misses lower rows) and look at them.

If host ports are taken by another local stack, test with a throwaway override file that remaps `ports: !override [...]` rather than editing `docker-compose.yml`. Finish with `docker compose --profile mcp down -v` (Grafana's admin account locks after a few failed logins; `-v` resets it).

Plan: `PLANS/PLAN_ARCHITECTURE.md` (Metrics, logs and traces design section). No dedicated `PLAN_OBSERVABILITY.md` — this is infrastructure/tooling, not a user-facing tool.

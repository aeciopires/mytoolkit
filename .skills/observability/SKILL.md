---
name: observability
description: Modify the Prometheus + Grafana observability stack (observability/ directory, docker-compose.yml's prometheus/grafana services) or the Grafana dashboard JSON. Trigger on "add a metric to the dashboard", "update Grafana dashboard", "add Prometheus scrape config".
---

# Observability stack

`docker-compose.yml`'s `prometheus` and `grafana` services, plus everything under `observability/`:

```
observability/
  prometheus.yml                                  scrape config: job "mytoolkit" (mytoolkit:8080) and job "mytoolkit-mcp" (mytoolkit-mcp:8081, only up when the optional docker-compose "mcp" profile is running), 15s interval
  mytoolkit-dashboard.json                         the dashboard itself — deliberately at the top level, not nested, per the request that asked for it "in the observability directory"
  grafana/provisioning/datasources/datasource.yml  auto-provisions a Prometheus data source, fixed uid: "prometheus"
  grafana/provisioning/dashboards/dashboard.yml    tells Grafana to load *.json dashboards from /var/lib/grafana/dashboards
```

## Every panel's `datasource.uid` must be `"prometheus"`

This has to match `datasource.yml`'s `uid: prometheus` exactly, or panels show "datasource not found" on a fresh (no manual setup) `docker compose up`. Don't switch to a Grafana template variable (`${DS_PROMETHEUS}`) — that pattern is for dashboards a human imports through the UI and picks a data source for; this one is provisioned automatically and must work with zero clicks.

## Single-file bind mount gotcha

`docker-compose.yml` mounts the dashboard as a single file: `./observability/mytoolkit-dashboard.json:/var/lib/grafana/dashboards/mytoolkit-dashboard.json`. Docker single-file bind mounts track the host file's **inode**, not its path — an editor/tool that saves via write-new-file-then-rename (common, for atomicity) leaves the container looking at the old, now-unlinked inode. Verified directly: editing the file and waiting past `updateIntervalSeconds` (30s) and even forcing `POST /api/admin/provisioning/dashboards/reload` did **not** pick up a change; only `docker compose restart grafana` did (that recreates the mount). Whenever you edit `mytoolkit-dashboard.json` with the stack running, tell whoever's testing it to restart the `grafana` container, or the "verification" will silently be checking stale content.

## Grounding queries in real metric names

Every panel query in `mytoolkit-dashboard.json` was written against the *actual* `/metrics` output of a running binary (`curl localhost:PORT/metrics`), not from memory of what `client_golang`/`promhttp` usually expose — e.g. `go_gc_duration_seconds` only has `quantile` labels `0`, `0.25`, `0.5`, `0.75`, `1` (no `0.99`), which isn't obvious without checking. The custom app metrics are defined in `internal/metrics/metrics.go`, two families:
- REST/web surface: `mytoolkit_http_requests_total{tool,method,status}` (counter), `mytoolkit_http_request_duration_seconds{tool,method}` (histogram), `mytoolkit_tool_usage_total{tool}` (counter) — same data `GET /api/v1/metrics/ranking` derives from.
- MCP surface (recorded by `internal/mcp`'s `metricsMiddleware`, `server.AddReceivingMiddleware` — one hook covers every JSON-RPC method and tool, not per-handler instrumentation; see `.skills/mcp/SKILL.md`): `mytoolkit_mcp_requests_total{method,status}`, `mytoolkit_mcp_request_duration_seconds{method}`, `mytoolkit_mcp_tool_calls_total{tool,status}`, `mytoolkit_mcp_tool_call_duration_seconds{tool}` (histogram), `mytoolkit_mcp_sessions_total` (counter, no labels). Kept as a **separate metric family**, not merged into `mytoolkit_http_*`/`mytoolkit_tool_usage_total` — the MCP surface is a distinct client population (see `PLAN_ARCHITECTURE.md`'s usage-ranking scope note) and conflating them would misattribute usage between surfaces. Only reachable when the MCP process is `--transport http` (mounted on the same port as the MCP endpoint via one `http.ServeMux`, path `/metrics`) — `stdio` has no listening port, so Prometheus's pull model has nothing to scrape there; this is inherent to stdio, not a gap to "fix."

If a new custom metric is added to `internal/metrics`, add a panel for it here and re-verify against real `/metrics` output — don't guess label names.

## Table panels: aggregate away Prometheus's own labels

An *instant* query with no aggregation (e.g. bare `mytoolkit_tool_usage_total`) returns every label the metric carries, including Prometheus's own `instance`/`job`/`__name__` — as a table panel, that renders extra clutter columns no one wants. Wrap the query in `sum(...) by (the labels you actually want)` (see the "Tool Usage Ranking" panel's `sort_desc(sum(mytoolkit_tool_usage_total) by (tool))` — a real bug found and fixed by screenshotting the rendered dashboard, not by reading the JSON) to strip everything else.

## Verifying changes

Static JSON validity (`python3 -m json.tool`) is necessary but nowhere near sufficient — a syntactically valid dashboard can still reference the wrong datasource uid, an aggregation-free query cluttering a table, or a metric name that doesn't exist. Verify by actually running the stack: `docker compose up -d --build` (add `--profile mcp mytoolkit-mcp` to also bring up the MCP scrape target when testing `mytoolkit_mcp_*` panels), generate some traffic (`curl -X POST localhost:8080/api/v1/tools/<slug> -d '...'` a few times for REST/web panels; for MCP panels, a real `initialize` → `notifications/initialized` → `tools/call` JSON-RPC handshake against `localhost:8081/` — see `mcp/README.md`'s Examples section for the exact `curl` sequence), then confirm the data actually lands in Prometheus (`curl 'localhost:9090/api/v1/query?query=...'`) before checking Grafana — the Grafana API (`curl -u admin:admin localhost:3000/api/...`, HTTP Basic Auth works fine there) can query a panel's exact PromQL expression through the datasource proxy (`/api/datasources/proxy/uid/prometheus/api/v1/query?query=...`) without needing a browser, which is how the MCP panels were confirmed to render real data end-to-end. `docker compose down -v` afterward to reset Grafana's anonymous data volume (its default admin account has a login-attempt lockout after a handful of failures, encountered directly while testing this).

Plan: `PLANS/PLAN_ARCHITECTURE.md`'s Metrics design section. No dedicated `PLAN_OBSERVABILITY.md` — this is infrastructure/tooling, not a user-facing tool.

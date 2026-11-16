<!-- TOC -->

- [Environment Variables](#environment-variables)
  - [OpenTelemetry tracing](#opentelemetry-tracing)
  - [Logging](#logging)

<!-- TOC -->

# Environment Variables

Every runtime-configurable value follows precedence: **CLI flag > environment variable > built-in default**.

| Variable | CLI flag (`serve`) | Default | Description |
|---|---|---|---|
| `MYTOOLKIT_HOST` | `--host` | `0.0.0.0` | Interface the HTTP server binds to. Also used by `mytoolkit mcp --transport http` (via `--host`). |
| `MYTOOLKIT_PORT` | `--port` | `8080` | TCP port the HTTP server listens on. |
| `MYTOOLKIT_LOG_LEVEL` | `--log-level` | `info` | Log level: `debug`, `info`, `warn`, `error`. Also used by `mytoolkit mcp`. Logs are always JSON, see [Logging](#logging). |
| `MYTOOLKIT_MCP_TRANSPORT` | `--transport` (`mcp`) | `stdio` | `stdio` or `http`. See [../mcp/README.md](../mcp/README.md). |
| `MYTOOLKIT_MCP_PORT` | `--port` (`mcp`) | `8081` | TCP port `mytoolkit mcp --transport http` listens on. |
| `MYTOOLKIT_SHUTDOWN_DELAY` | — | `0` (none) | After SIGINT/SIGTERM, keep serving this long (Go duration, e.g. `5s`) with `/readyz` returning `503` before closing the listener and draining in-flight requests (up to 20s). Lets Kubernetes remove the pod from Service endpoints first; the Helm chart sets `5s`. Applies to `serve` and `mcp --transport http`. |

Copy `.env-example` to `.env` for local development / `docker-compose`:

```
cp .env-example .env
```

`docker-compose.yml` loads `.env` via `env_file`. The Helm chart exposes the same three `serve`-related keys under `values.yaml`'s `env:` map, and the two MCP-specific keys under `values.yaml`'s `mcp:` block (see [../helm/mytoolkit/values.yaml](../helm/mytoolkit/values.yaml)).

## OpenTelemetry tracing

`serve` and `mcp` export OpenTelemetry traces over **OTLP/HTTP** only when an OTLP endpoint is configured — otherwise tracing is off and costs nothing. Configuration uses the standard OpenTelemetry SDK variables (no `MYTOOLKIT_*` equivalent, no CLI flag):

| Variable | Default | Description |
|---|---|---|
| `OTEL_EXPORTER_OTLP_ENDPOINT` | unset (tracing off) | OTLP/HTTP base URL, e.g. `http://alloy:4318` (Grafana Alloy, an OpenTelemetry Collector, Tempo or Jaeger). Setting it turns tracing on. |
| `OTEL_EXPORTER_OTLP_TRACES_ENDPOINT` | unset | Full traces URL (e.g. `http://alloy:4318/v1/traces`); also turns tracing on. |
| `OTEL_SERVICE_NAME` | `mytoolkit` (`serve`) / `mytoolkit-mcp` (`mcp`) | `service.name` resource attribute. |
| `OTEL_RESOURCE_ATTRIBUTES` | unset | Extra resource attributes, e.g. `deployment.environment.name=prod`. |
| `OTEL_TRACES_SAMPLER` / `OTEL_TRACES_SAMPLER_ARG` | `parentbased_always_on` | Sampling, e.g. `parentbased_traceidratio` + `0.1`. |
| `OTEL_SDK_DISABLED` / `OTEL_TRACES_EXPORTER=none` | unset | Force tracing off even with an endpoint set. |

An incoming W3C `traceparent` header (REST, MCP over HTTP) or `params._meta.traceparent` (MCP, any transport) is honored. `docker-compose.yml` sets these for both app services; the Helm chart exposes them under `values.yaml`'s `tracing:` block.

## Logging

`serve` and `mcp` write **one JSON object per line to stderr** (never stdout — stdout carries CLI tool output and the MCP stdio JSON-RPC stream). Every line has `level`, `service`, `version`, `time` (RFC 3339, nanoseconds) and `message`; request lines add `tool`, `request_id`, `method`, `path`, `status`, `duration_ms` and, with tracing on, `trace_id`/`span_id`:

```json
{"level":"info","service":"mytoolkit","version":"1.2.0","request_id":"notebook/SMrz38YmWa-000001","tool":"base64","trace_id":"235fb3b24ca4bf832ef77b7d3906a19b","span_id":"9a8602dfa5fbe64a","method":"POST","path":"/api/v1/tools/base64","status":200,"duration_ms":0.187,"time":"2026-11-16T06:54:17.790699083-03:00","message":"request completed"}
```

Standard-library `log`, `log/slog` and OpenTelemetry SDK messages are routed into the same JSON format (with a `logger` field), and a fatal server error (e.g. port already in use) is logged as a JSON `error` line too. One-shot CLI tool commands keep their human-readable `Error: ...` message.


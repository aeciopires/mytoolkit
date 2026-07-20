<!-- TOC -->

- [Environment Variables](#environment-variables)

<!-- TOC -->

# Environment Variables

Every runtime-configurable value follows precedence: **CLI flag > environment variable > built-in default**.

| Variable | CLI flag (`serve`) | Default | Description |
|---|---|---|---|
| `MYTOOLKIT_HOST` | `--host` | `0.0.0.0` | Interface the HTTP server binds to. Also used by `mytoolkit mcp --transport http` (via `--host`). |
| `MYTOOLKIT_PORT` | `--port` | `8080` | TCP port the HTTP server listens on. |
| `MYTOOLKIT_LOG_LEVEL` | `--log-level` | `info` | zerolog level: `debug`, `info`, `warn`, `error`. Also used by `mytoolkit mcp`. |
| `MYTOOLKIT_MCP_TRANSPORT` | `--transport` (`mcp`) | `stdio` | `stdio` or `http`. See [../mcp/README.md](../mcp/README.md). |
| `MYTOOLKIT_MCP_PORT` | `--port` (`mcp`) | `8081` | TCP port `mytoolkit mcp --transport http` listens on. |

Copy `.env-example` to `.env` for local development / `docker-compose`:

```
cp .env-example .env
```

`docker-compose.yml` loads `.env` via `env_file`. The Helm chart exposes the same three `serve`-related keys under `values.yaml`'s `env:` map, and the two MCP-specific keys under `values.yaml`'s `mcp:` block (see [../helm/mytoolkit/values.yaml](../helm/mytoolkit/values.yaml)).

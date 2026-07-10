<!-- TOC -->

- [Environment Variables](#environment-variables)

<!-- TOC -->

# Environment Variables

Every runtime-configurable value follows precedence: **CLI flag > environment variable > built-in default**.

| Variable | CLI flag (`serve`) | Default | Description |
|---|---|---|---|
| `MYTOOLKIT_HOST` | `--host` | `0.0.0.0` | Interface the HTTP server binds to. |
| `MYTOOLKIT_PORT` | `--port` | `8080` | TCP port the HTTP server listens on. |
| `MYTOOLKIT_LOG_LEVEL` | `--log-level` | `info` | zerolog level: `debug`, `info`, `warn`, `error`. |

Copy `.env-example` to `.env` for local development / `docker-compose`:

```
cp .env-example .env
```

`docker-compose.yml` loads `.env` via `env_file`. The Helm chart exposes the same three keys under `values.yaml`'s `env:` map.

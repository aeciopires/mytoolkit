<!-- TOC -->

- [Roadmap](#roadmap)
  - [Done (v1.0.0)](#done-v100)
  - [Planned](#planned)

<!-- TOC -->

# Roadmap

## Done (v1.0.0)

MyToolkit is fully implemented, tested, and deployed.

- Application (``app/``) — Go module with 15 tools each with pure business logic + unit tests, a REST handler, a CLI subcommand, and a web page sharing one Material Design 3–styled layout with dark/light theming. Shared packages (``apperr``, ``textio``, ``config``, ``response``, ``registry``) avoid duplicating logic across the three surfaces. Structured JSON logging (zerolog), Prometheus metrics + usage ranking, ``/healthz/``,``/readyz``, and a ``--version/-v`` flag driven by a repo-root ``VERSION`` file.

- Packaging & docs: multi-stage/multi-arch Dockerfile, docker-compose.yml, a Makefile with check-tools, helm-docs, helm-install/uninstall, and a secure docker-push (credential prompt piped via --password-stdin, never logged). Full docs/api|cli|testing triplet + .skills/ per tool, rewritten README.md (architecture diagrams, screenshots, Documentation table), CHANGELOG.md, ROADMAP.md, CLAUDE.md, CONTRIBUTING.md. PLAN_ARCHITECTURE.md updated with a changelog of every requirement added mid-implementation.

## Planned

- Additional tools:
  - UUID generator
  - Diff JSON viewer
  - Diff YAML viewer
  - Cron expression parser
  - CSV to JSON Converter
  - CSV to YAML Converter
  - IAM Policy JSON to Terraform
  - Tiktokenizer
- CI pipeline (lint, test, build, and push a multi-arch image on tag) via GitHub Actions.
- Per-tool usage analytics dashboard on the homepage, backed by `/api/v1/metrics/ranking`.
- Configurable rate limiting for the REST API.
- i18n for the web UI.

Suggestions and contributions are welcome — see `CONTRIBUTING.md`.

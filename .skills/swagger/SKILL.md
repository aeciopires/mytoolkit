---
name: swagger
description: Add or modify Swagger/OpenAPI documentation (swaggo/swag annotations, app/docs/, the /swagger/index.html UI). Trigger on "document the REST API", "add swagger endpoint", "update swagger docs", "add a new tool" (annotations are part of that checklist too).
---

# Swagger / OpenAPI documentation

The REST API is documented at `GET /swagger/index.html` (swaggo/http-swagger), backed by a spec generated from `@`-comment annotations via `make swagger-gen` into `app/docs/` (`docs.go`, `swagger.json`, `swagger.yaml` — generated, never hand-edit; regenerate and commit the result after touching any annotation).

## The one thing that will bite you: cross-package type resolution

`swag` only resolves `@Success 200 {object} pkg.Type` (or `@Param`/`@Failure`) if **the file the annotation lives in actually imports `pkg`** — not just "the type is exported and in the module." This is why:

- The shared doc-only response DTOs (`ToolSuccessResponse`, `ToolErrorResponse`, `ToolMeta`, `ToolErrorBody`) live in **`internal/cli/swaggermodels.go`** (package `cli`), not `internal/httpapi` — every annotated handler is itself in `internal/cli`, and `internal/cli` doesn't import `internal/httpapi`. Moving these back to `internal/httpapi` will break `swag init` with `cannot find type definition: httpapi.ToolSuccessResponse` — this was hit and fixed once already, don't reintroduce it.
- These DTOs deliberately mirror (not reuse) `internal/response`'s real, unexported envelope types (`successResponse`/`errorResponse`). If the real envelope's shape changes, update both.
- For a tool with a genuinely custom response shape (JSON Tree Viewer, Text Counter, Kubernetes YAML Validator), the annotation uses swag's inline `object{field=Type}` composition syntax instead, e.g. `@Success 200 {object} object{success=bool,data=k8svalidate.Result,meta=ToolMeta}` — `k8svalidate.Result` resolves fine because `k8svalidate.go` already imports `internal/tools/k8svalidate` for real code, not just the annotation.

## Every tool needs a named handler function to annotate

9 of the 15 tools are wired via the generic `handlers.Wrap("slug", pkg.Fn)` called inline inside `init()` — there's no named function for `swag` to attach a `// godoc` comment to. Each of those got a tiny wrapper purely to carry the annotation:

```go
// base64Handler godoc
// @Summary Encode or decode Base64
// @Tags tools
// @Accept json
// @Produce json
// @Param request body object{input=string,options=object{decode=bool,variant=string,padding=bool}} true "..."
// @Success 200 {object} ToolSuccessResponse
// @Failure 400 {object} ToolErrorResponse
// @Router /api/v1/tools/base64 [post]
func base64Handler() http.HandlerFunc {
	return handlers.Wrap("base64", base64enc.Process)
}
```
`registerToolHandler("base64", base64Handler())` replaces the old inline `registerToolHandler("base64", handlers.Wrap(...))`. The other 6 tools (json-tree, text-count, password-gen, jwt, qrcode, k8s-validate) already had bespoke named handler functions (`jsonTreeHandler` etc.) — their annotations go directly above the existing function, no new wrapper needed.

**Adding a new tool**: if it's generic (`handlers.Wrap`), follow the wrapper pattern above. If it's bespoke, annotate the existing handler function directly. Either way, run `make swagger-gen` afterward and check `internal/httpapi/swagger_test.go`'s `TestSwaggerSpecCoversEveryTool` passes — it asserts every `registry.All()` slug has a matching Swagger path, specifically so a forgotten annotation fails a test instead of silently missing from the docs.

## One literal path per tool, not one templated route

chi's actual route table has a single `/api/v1/tools/{slug}` route (`dispatchTool` in `router.go` looks up the handler by slug at request time). The Swagger spec instead documents each tool as its own literal `@Router /api/v1/tools/<slug> [post]` path — OpenAPI paths don't have to mirror the server's internal routing tree 1:1, and per-tool paths let each one have its own accurate request schema/example/description, matching how `docs/api/<slug>.md` already documents them individually. Don't try to collapse these into one generic `/api/v1/tools/{slug}` Swagger path with a `slug` path parameter — it would lose exactly the per-tool detail that makes the generated docs useful.

## `/metrics` is deliberately undocumented

Prometheus's text exposition format isn't JSON and doesn't fit an OpenAPI schema — it's excluded from the Swagger spec on purpose, not an oversight. `/healthz`, `/readyz`, `GET /api/v1/tools`, and `GET /api/v1/metrics/ranking` (all real JSON endpoints) are annotated directly on their handlers in `internal/httpapi/health.go`/`router.go`, tagged `system` (vs. `tools` for the 15 tool endpoints).

## Regenerating and verifying

```
make swagger-gen   # cd app && go run github.com/swaggo/swag/cmd/swag@v1.16.6 init -g cmd/mytoolkit/main.go -o docs --parseDependency --parseInternal
make build && make test && make lint
```
`go run` pins an exact `swag` version (matching `go.mod`'s `github.com/swaggo/swag` requirement) so generation is reproducible without a separate global install. After regenerating, verify with more than the JSON diff — load `/swagger/index.html` in a real browser (or at least `curl localhost:8080/swagger/doc.json | python3 -m json.tool`) and confirm the endpoint you changed renders with the request/response shape you expect; a syntactically valid spec can still reference the wrong type or a stale example.

Plan: `PLANS/PLAN_ARCHITECTURE.md`'s REST API design section.

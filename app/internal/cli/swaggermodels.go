package cli

// This file exists purely so github.com/swaggo/swag has concrete, exported
// Go types in this package to point @Param/@Success/@Failure annotations
// at — swag only resolves a cross-package type reference (e.g.
// "somepkg.Type") if the annotated file actually imports that package, so
// these live in internal/cli itself rather than internal/httpapi, since
// httpapi isn't (and has no reason to be) imported by internal/cli. The
// app's real request/response envelope logic lives in
// internal/httpapi/handlers and internal/response — those types are
// intentionally unexported, since they have no reason to be part of that
// package's public API otherwise. Keep these two definitions structurally
// identical to the real envelope (see internal/response/response.go); if
// the real envelope's shape changes, update these too, or the generated
// Swagger spec will lie about the API.

// ToolSuccessResponse is the shared success envelope every tool endpoint
// returns on HTTP 200. Data's shape varies per tool (most return
// {"output": "..."}; a few — JSON Tree Viewer, Text Counter, Password
// Generator, Kubernetes YAML Validator — return their own structured
// shape instead, documented on their respective endpoints).
type ToolSuccessResponse struct {
	Success bool           `json:"success" example:"true"`
	Data    map[string]any `json:"data"`
	Meta    ToolMeta       `json:"meta"`
}

type ToolMeta struct {
	Tool       string  `json:"tool" example:"base64"`
	DurationMs float64 `json:"duration_ms" example:"0.08"`
}

// ToolErrorResponse is the shared error envelope every tool endpoint
// returns on a non-2xx status.
type ToolErrorResponse struct {
	Success bool          `json:"success" example:"false"`
	Error   ToolErrorBody `json:"error"`
}

type ToolErrorBody struct {
	Code    string `json:"code" example:"INVALID_OPTIONS"`
	Message string `json:"message" example:"invalid options object"`
}

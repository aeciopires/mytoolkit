package mcp

// textOut is the shared MCP output shape for tools that return a single
// transformed text result, mirroring the REST envelope's {"output": ...}.
type textOut struct {
	Output string `json:"output"`
}

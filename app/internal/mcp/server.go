// Package mcp is the 4th surface over internal/tools/<name>, alongside
// web, REST, and CLI: it exposes every tool as an MCP (Model Context
// Protocol) tool, using the same pure-function implementations the other
// three surfaces already call. Each <name>.go file registers itself via
// init(), mirroring the toolHandlers pattern internal/cli/serve.go uses
// for REST handlers.
package mcp

import (
	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// toolAdders accumulates one registration func per MCP tool, appended by
// each <name>.go's init(). jwt registers two (jwt_encode/jwt_decode), so
// this ends up with more entries than there are registry.Tool rows.
var toolAdders []func(*sdkmcp.Server)

func register(fn func(*sdkmcp.Server)) {
	toolAdders = append(toolAdders, fn)
}

// NewServer builds a fresh MCP server with every tool registered. It is
// cheap to call repeatedly (internal/tools/<name> functions are pure), so
// the streamable-HTTP transport calls it once per incoming connection
// while the stdio transport calls it once at process start.
func NewServer(version string) *sdkmcp.Server {
	server := sdkmcp.NewServer(&sdkmcp.Implementation{
		Name:    "mytoolkit",
		Title:   "MyToolkit",
		Version: version,
	}, nil)
	server.AddReceivingMiddleware(metricsMiddleware)
	for _, add := range toolAdders {
		add(server)
	}
	return server
}

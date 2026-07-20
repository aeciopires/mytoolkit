package mcp

import (
	"context"
	"encoding/json"
	"testing"

	sdkmcp "github.com/modelcontextprotocol/go-sdk/mcp"
)

// connect spins up a real client<->server MCP session over an in-memory
// transport pair, the SDK's equivalent of httptest.NewServer for protocol-
// level adapter tests (see internal/httpapi's handler tests for the same
// role on the REST surface).
func connect(t *testing.T) (*sdkmcp.ClientSession, func()) {
	t.Helper()
	ctx := context.Background()

	server := NewServer("test")
	clientTransport, serverTransport := sdkmcp.NewInMemoryTransports()

	serverSession, err := server.Connect(ctx, serverTransport, nil)
	if err != nil {
		t.Fatalf("server.Connect: %v", err)
	}

	client := sdkmcp.NewClient(&sdkmcp.Implementation{Name: "test-client", Version: "0.0.1"}, nil)
	clientSession, err := client.Connect(ctx, clientTransport, nil)
	if err != nil {
		t.Fatalf("client.Connect: %v", err)
	}

	return clientSession, func() {
		_ = clientSession.Close()
		_ = serverSession.Close()
	}
}

func TestListToolsCoversEveryTool(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()

	res, err := session.ListTools(context.Background(), nil)
	if err != nil {
		t.Fatalf("ListTools: %v", err)
	}

	want := []string{
		"base64", "json-format", "yaml-format", "password-gen",
		"jwt_decode", "jwt_encode", "qrcode", "text-count", "url-encode",
		"hash-gen", "case-convert", "json-toon", "yaml-to-json",
		"json-to-yaml", "k8s-validate", "json-tree",
	}
	if len(res.Tools) != len(want) {
		t.Fatalf("got %d tools, want %d: %+v", len(res.Tools), len(want), res.Tools)
	}
	got := map[string]bool{}
	for _, tool := range res.Tools {
		got[tool.Name] = true
		if tool.Description == "" {
			t.Errorf("tool %q has no description", tool.Name)
		}
	}
	for _, name := range want {
		if !got[name] {
			t.Errorf("missing tool %q", name)
		}
	}
}

func TestCallToolBase64RoundTrip(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()
	ctx := context.Background()

	encodeRes, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": "hello mcp"},
	})
	if err != nil {
		t.Fatalf("CallTool base64 encode: %v", err)
	}
	if encodeRes.IsError {
		t.Fatalf("unexpected error result: %+v", encodeRes.Content)
	}
	encoded := textFromResult(t, encodeRes)
	if encoded != "aGVsbG8gbWNw" {
		t.Fatalf("got %q, want %q", encoded, "aGVsbG8gbWNw")
	}

	decodeRes, err := session.CallTool(ctx, &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": encoded, "decode": true},
	})
	if err != nil {
		t.Fatalf("CallTool base64 decode: %v", err)
	}
	if decoded := textFromResult(t, decodeRes); decoded != "hello mcp" {
		t.Fatalf("got %q, want %q", decoded, "hello mcp")
	}
}

func TestCallToolBase64InvalidVariantIsToolError(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "base64",
		Arguments: map[string]any{"input": "hi", "variant": "bogus"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if !res.IsError {
		t.Fatalf("expected IsError=true for an invalid variant, got %+v", res)
	}
}

func TestCallToolQRCodeReturnsImageContent(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "qrcode",
		Arguments: map[string]any{"text": "https://example.com"},
	})
	if err != nil {
		t.Fatalf("CallTool qrcode: %v", err)
	}
	if res.IsError {
		t.Fatalf("unexpected error result: %+v", res.Content)
	}
	if len(res.Content) != 1 {
		t.Fatalf("got %d content blocks, want 1", len(res.Content))
	}
	img, ok := res.Content[0].(*sdkmcp.ImageContent)
	if !ok {
		t.Fatalf("got %T, want *mcp.ImageContent", res.Content[0])
	}
	if img.MIMEType != "image/png" {
		t.Errorf("got MIMEType %q, want image/png", img.MIMEType)
	}
	if len(img.Data) == 0 {
		t.Error("expected non-empty PNG data")
	}
	if !isPNG(img.Data) {
		t.Error("image data does not start with the PNG magic bytes")
	}
}

func TestCallToolK8sValidateInvalidDocumentIsNotAToolError(t *testing.T) {
	session, closeFn := connect(t)
	defer closeFn()

	res, err := session.CallTool(context.Background(), &sdkmcp.CallToolParams{
		Name:      "k8s-validate",
		Arguments: map[string]any{"input": "kind: Pod\n"},
	})
	if err != nil {
		t.Fatalf("CallTool: %v", err)
	}
	if res.IsError {
		t.Fatalf("a semantically-invalid-but-parseable document should not be a tool error, got %+v", res.Content)
	}

	var result struct {
		Valid bool `json:"valid"`
	}
	if err := json.Unmarshal(structuredJSON(t, res), &result); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	if result.Valid {
		t.Error("expected valid=false: document is missing apiVersion")
	}
}

func textFromResult(t *testing.T, res *sdkmcp.CallToolResult) string {
	t.Helper()
	var out struct {
		Output string `json:"output"`
	}
	if err := json.Unmarshal(structuredJSON(t, res), &out); err != nil {
		t.Fatalf("unmarshal structured content: %v", err)
	}
	return out.Output
}

func structuredJSON(t *testing.T, res *sdkmcp.CallToolResult) []byte {
	t.Helper()
	if res.StructuredContent != nil {
		b, err := json.Marshal(res.StructuredContent)
		if err != nil {
			t.Fatalf("marshal structured content: %v", err)
		}
		return b
	}
	if len(res.Content) != 1 {
		t.Fatalf("got %d content blocks, want 1", len(res.Content))
	}
	tc, ok := res.Content[0].(*sdkmcp.TextContent)
	if !ok {
		t.Fatalf("got %T, want *mcp.TextContent", res.Content[0])
	}
	return []byte(tc.Text)
}

func isPNG(b []byte) bool {
	sig := []byte{0x89, 'P', 'N', 'G', '\r', '\n', 0x1a, '\n'}
	if len(b) < len(sig) {
		return false
	}
	for i, c := range sig {
		if b[i] != c {
			return false
		}
	}
	return true
}

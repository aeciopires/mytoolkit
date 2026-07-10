// Package registry is the single source of truth for tool metadata, shared
// by the web nav, CLI help, GET /api/v1/tools, and metrics labels.
package registry

type Tool struct {
	Slug        string `json:"slug"`
	Name        string `json:"name"`
	Emoji       string `json:"emoji"`
	Description string `json:"description"`
	// ClientSide marks tools whose web page converts entirely in the
	// browser (no REST call from the page) — see json-toon and
	// PLAN_ARCHITECTURE.md's "data-client-side" convention.
	ClientSide bool `json:"client_side"`
}

var Tools = []Tool{
	{Slug: "json-tree", Name: "JSON Tree Viewer", Emoji: "🌳", Description: "Visualize JSON structures in a tree format for easier navigation and analysis."},
	{Slug: "json-format", Name: "JSON Formatter", Emoji: "📄", Description: "Format and organize JSON documents to improve readability."},
	{Slug: "yaml-format", Name: "YAML Formatter", Emoji: "📝", Description: "Format YAML files with consistent indentation."},
	{Slug: "password-gen", Name: "Password Generator", Emoji: "🔐", Description: "Generate strong, customizable passwords."},
	{Slug: "jwt", Name: "JWT Encode/Decode", Emoji: "🎫", Description: "Encode and decode JSON Web Tokens (JWT) for inspection and testing."},
	{Slug: "qrcode", Name: "QR Code Generator", Emoji: "📱", Description: "Generate QR codes from text, URLs, or unicode content."},
	{Slug: "text-count", Name: "Character, Word & Line Counter", Emoji: "📊", Description: "Count characters, words, and lines in any text."},
	{Slug: "url-encode", Name: "URL Encode/Decode", Emoji: "🌐", Description: "Encode and decode URLs according to web standards."},
	{Slug: "hash-gen", Name: "Hash Generator", Emoji: "🔒", Description: "Generate hashes using MD5, SHA-1, SHA-256, and SHA-512."},
	{Slug: "base64", Name: "Base64 Encode/Decode", Emoji: "🔤", Description: "Encode and decode data using Base64."},
	{Slug: "case-convert", Name: "Case Converter", Emoji: "🔡", Description: "Convert text between Sentence case, UPPER CASE, lower case, Title Case, Mixed Case, and Inverse Case."},
	{Slug: "json-toon", Name: "JSON to TOON Converter", Emoji: "🪶", Description: "Convert JSON into TOON to shrink LLM token usage. The web page converts entirely in your browser — nothing is sent to the server.", ClientSide: true},
	{Slug: "yaml-to-json", Name: "YAML to JSON Converter", Emoji: "🔃", Description: "Convert a YAML document to pretty-printed JSON."},
	{Slug: "json-to-yaml", Name: "JSON to YAML Converter", Emoji: "🔄", Description: "Convert a JSON document to YAML."},
	{Slug: "k8s-validate", Name: "Kubernetes YAML Validator", Emoji: "☸️", Description: "Validate that a YAML document (single or multi-document) has the fields the Kubernetes API requires: apiVersion, kind, and a well-formed metadata block."},
}

func All() []Tool {
	return Tools
}

func BySlug(slug string) (Tool, bool) {
	for _, t := range Tools {
		if t.Slug == slug {
			return t, true
		}
	}
	return Tool{}, false
}

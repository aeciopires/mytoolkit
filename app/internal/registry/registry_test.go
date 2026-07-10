package registry

import (
	"encoding/json"
	"testing"
)

// TestToolJSONFieldsAreLowercase locks in the registry.Tool JSON contract.
// Regression test: Tool previously had no json tags, so json.Marshal
// emitted capitalized Go field names (Slug, Name, ...). Both GET
// /api/v1/tools and the web layout's embedded window.MYTOOLKIT_TOOLS (used
// by the client-side search bar in nav.js, which reads t.slug/t.name/
// t.description/t.emoji) depend on lowercase keys — the mismatch silently
// broke search (no thrown error, just zero results) until caught by an
// end-to-end browser test, not a unit test. This test exists so the next
// regression is caught here instead.
func TestToolJSONFieldsAreLowercase(t *testing.T) {
	b, err := json.Marshal(Tool{
		Slug:        "example",
		Name:        "Example Tool",
		Emoji:       "🔧",
		Description: "An example.",
		ClientSide:  true,
	})
	if err != nil {
		t.Fatal(err)
	}

	var m map[string]any
	if err := json.Unmarshal(b, &m); err != nil {
		t.Fatal(err)
	}

	for _, key := range []string{"slug", "name", "emoji", "description", "client_side"} {
		if _, ok := m[key]; !ok {
			t.Errorf("expected lowercase JSON key %q, got keys: %v", key, m)
		}
	}
	for _, key := range []string{"Slug", "Name", "Emoji", "Description", "ClientSide"} {
		if _, ok := m[key]; ok {
			t.Errorf("unexpected capitalized JSON key %q leaked into output: %v", key, m)
		}
	}
}

func TestBySlug(t *testing.T) {
	if len(Tools) == 0 {
		t.Fatal("registry.Tools must not be empty")
	}
	want := Tools[0]
	got, ok := BySlug(want.Slug)
	if !ok || got != want {
		t.Errorf("BySlug(%q) = %+v, %v; want %+v, true", want.Slug, got, ok, want)
	}
	if _, ok := BySlug("does-not-exist"); ok {
		t.Error("BySlug() should return false for an unknown slug")
	}
}

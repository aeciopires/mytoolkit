---
name: web-ui-shell
description: Modify the shared web UI shell (navigation drawer, search bar, back-to-home button, footer, M3 design tokens) used by every page. Trigger on "add hamburger menu", "add nav drawer", "add search bar", "add footer", "M3 best practices", "material design components".
---

# Web UI shell

The shared chrome around every page lives in `internal/web/templates/layout.html` (executed once per request as the `"layout"` template, wrapping each page's `{{define "content"}}` block) plus two static assets: `internal/web/static/js/nav.js` and `internal/web/static/css/app.css`. Touch these three files (plus `theme.css` for tokens) — never re-implement nav/search/footer per tool page.

## Navigation drawer

M3 "modal navigation drawer" pattern: `#nav-drawer-toggle` (hamburger) opens `#nav-drawer`, backed by `#nav-scrim`. `nav.js` toggles the `.open` class on both and manages `aria-hidden`/`aria-expanded`. Closes on: hamburger re-click, `#nav-drawer-close` (✕), scrim click, or Escape. The tool list inside is built from `.Tools` (passed by every handler in `internal/web/handlers.go`) — adding a tool to `registry.Tools` automatically adds it to the drawer, no template change needed.

`.nav-drawer` is `position: fixed` (off-canvas via `transform: translateX(-100%)` when closed, slid in via `.open`) with `.scrim` dimming everything behind it — this is why it never overlaps or reflows the homepage/tool-page content underneath even though it visually sits on top while open: it's an overlay, not part of normal document flow. Keep it `position: fixed`; don't switch it to `position: absolute`/relative-in-flow to "simplify" the CSS, that would reintroduce actual layout overlap with page content instead of a controlled, scrim-backed overlay.

## Search bar

`#tool-search` filters `window.MYTOOLKIT_TOOLS` — a JSON blob embedded server-side (`internal/web/handlers.go`'s `toolsJSON` var, computed once at `init()` from `registry.All()`, injected via `{{.ToolsJSON}}` as a `template.JS` value so `html/template` doesn't re-escape it). Matching is case-insensitive substring match against `name` **and** `description` — i.e. it indexes the exact same text shown in each tool's homepage card and tool-page hero card, per the original request ("search bar indexing feature page and content about the title card section"). Entirely client-side, zero network calls.

**This is the exact mechanism that broke once already**: `registry.Tool` must have lowercase `json:"..."` tags on every field, or the JSON embedded into the page has capitalized keys (`Slug`, `Name`, ...) while `nav.js` reads lowercase (`t.slug`, `t.name`, ...) — no thrown error, just silently zero search results. `internal/registry/registry_test.go`'s `TestToolJSONFieldsAreLowercase` guards this; keep it passing.

## Header content is centered to the content column

`.topnav` (the sticky bar itself) stays full-bleed for its background/border, but its children (hamburger, brand, back-home button, search field, theme toggle) live inside a `.topnav-inner` wrapper (`max-width: 960px; margin: 0 auto;`) — the same width as `.content`. Without this wrapper, the header's content pins to the far-left/far-right edges of the viewport while the page body below it is centered, which looks misaligned on wide screens (this was a real, reported visual issue, not a hypothetical). This is a plain flex row, not `position: fixed`/`absolute`, so it cannot overlap page content by construction — don't confuse it with the nav drawer below.

## Back-to-Home button and footer

Both driven by `internal/web/handlers.go`'s data map, not per-page templates:
- `.ActiveSlug` is `""` on the homepage, `<slug>` on every tool page → `layout.html` shows `.back-home-btn` only `{{if .ActiveSlug}}`.
- `.site-footer` is unconditional, in `layout.html`, static content.

Never add a "back to home" link or footer markup inside an individual `tools/<name>.html` file — it's already there via the shared layout.

## M3 design tokens

`theme.css` defines the palette (`--color-*`, already existed) plus, added for this shell: a 4dp spacing scale (`--space-1`…`--space-10`, per `https://m3.material.io/styles/spacing/overview`), state-layer opacities (`--state-hover`/`--state-focus`/`--state-press`, per M3's interaction-states foundation, used with `color-mix()` for hover/focus/press backgrounds on buttons and list items), and a shape scale (`--shape-xs`…`--shape-full`). New CSS should reuse these tokens, not hardcode new `px`/`rem` spacing or radius values.

Component-specific notes:
- **Checkbox vs switch**: a plain `<input type="checkbox">` inside `.options-row` renders as an M3-style square checkbox (multi-select-from-a-set semantics — see Password Generator's charset toggles). Wrap a *single* standalone on/off setting in `<label class="switch">` instead (see `base64`/`url-encode`'s "Decode" toggle) — this is an M3 checkbox-vs-switch distinction, not a style preference; don't use one for the other.
- **Text fields**: focus state is a `box-shadow`/`border-color` ring in `--color-primary`, applied uniformly to `textarea`/`input[type=text]`/`input[type=number]`/`select`.
- **Lists**: `.nav-list-item` (56px-ish min-height, leading icon, hover state layer, `.active` pill highlight) is reused by both the drawer and the search-results dropdown — don't create a second list-item style.

## Verification

This is pure frontend (HTML/CSS/JS) — `go test` doesn't exercise it. Verify with a real browser (Playwright driving the actual running binary is what caught both real bugs found while building this: the JSON-case search bug above, and a missing-favicon console 404 fixed by adding `internal/web/static/icons/favicon.svg` + a `<link rel="icon">` in `layout.html`). At minimum: open the drawer, type a query into search and click a result, visit a tool page and click "Back to Home", and check the browser console/network tab for errors — don't just eyeball a screenshot.

Plan: `PLANS/PLAN_ARCHITECTURE.md`'s Theming and visual design system section.

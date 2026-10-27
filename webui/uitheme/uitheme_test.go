// uitheme_test.go — the theme-token + icon-pack golden battery (T-72's
// own acceptance: the token golden). Three consumers lock against TWO
// reference-generated goldens (testdiff/goldens/dump_theme.py +
// dump_icons.py, run against the UNMODIFIED reference at the pinned
// rev):
//
//  1. the Go tables      (theme.go / icons.go)   — direct canonical diff;
//  2. the CSS var dump    (../frontend/theme.css) — per-theme block
//     extraction, kebab→python key mapping, canonical diff;
//  3. the JS icon pack    (../frontend/icons.js)  — strict per-line
//     extraction, canonical diff.
//
// The canonical form is the generators' law (sort_keys, indent 2, no
// HTML escaping, trailing newline) — canon() below reproduces it; any
// byte drift fails the test with a per-token summary, never a silent
// pass.
package uitheme

import (
	"bytes"
	"encoding/json"
	"os"
	"regexp"
	"strings"
	"testing"
)

// canon marshals v in the generators' canonical JSON form: sorted keys,
// two-space indent, no HTML escaping, trailing newline — the exact shape
// Python's json.dumps(sort_keys=True, indent=2) + "\n" produces, so
// byte-comparison against the goldens is honest.
func canon(t *testing.T, v any) []byte {
	t.Helper()
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetEscapeHTML(false)
	enc.SetIndent("", "  ")
	if err := enc.Encode(v); err != nil {
		t.Fatalf("canon marshal: %v", err)
	}
	return buf.Bytes()
}

func loadGolden(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatalf("reading golden %s: %v (regenerate via testdiff/goldens/dump_theme.py|dump_icons.py)", name, err)
	}
	return b
}

// themeDoc is the theme golden's document shape.
type themeDoc struct {
	Schema int               `json:"schema"`
	Light  map[string]string `json:"light"`
	Dark   map[string]string `json:"dark"`
}

// iconsDoc is the icons golden's document shape.
type iconsDoc struct {
	Schema int               `json:"schema"`
	Icons  map[string]string `json:"icons"`
}

// diffTheme reports the first divergences between a got-doc and the
// golden, both already unmarshaled.
func diffTheme(t *testing.T, got, want *themeDoc) {
	for _, theme := range [...]struct {
		name string
		g, w map[string]string
	}{
		{"light", got.Light, want.Light},
		{"dark", got.Dark, want.Dark},
	} {
		if theme.g == nil || theme.w == nil {
			t.Errorf("%s table missing (unmarshal)", theme.name)
			continue
		}
		for k, v := range theme.w {
			if theme.g[k] != v {
				t.Errorf("%s.%s: has %q, golden has %q", theme.name, k, theme.g[k], v)
			}
		}
		for k := range theme.g {
			if _, ok := theme.w[k]; !ok {
				t.Errorf("%s.%s: an extra token the golden lacks", theme.name, k)
			}
		}
	}
}

// ---------------------------------------------------------------------------
// 1. The Go tables vs the theme golden.
// ---------------------------------------------------------------------------

func TestThemeGoTablesMatchGolden(t *testing.T) {
	// A MAP, not a struct: map keys marshal SORTED (dark, light, schema) —
	// the golden's own order. A struct would emit its declaration order
	// and diverge byte-wise with identical data.
	got := canon(t, map[string]any{"schema": 1, "light": Light, "dark": Dark})
	want := loadGolden(t, "theme_tokens.golden.json")
	if !bytes.Equal(got, want) {
		var g, w themeDoc
		if err := json.Unmarshal(got, &g); err != nil {
			t.Fatalf("drift, and the got-doc does not unmarshal: %v", err)
		}
		if err := json.Unmarshal(want, &w); err != nil {
			t.Fatalf("golden does not unmarshal (regenerate it): %v", err)
		}
		diffTheme(t, &g, &w)
		t.Fatalf("the Go theme tables drifted from the reference golden")
	}
	if len(Light) != 76 || len(Dark) != 76 {
		t.Fatalf("token counts: light=%d dark=%d, want 76/76", len(Light), len(Dark))
	}
}

// ---------------------------------------------------------------------------
// 2. The CSS var dump vs the same golden (the CSS twin).
// ---------------------------------------------------------------------------

// cssVarBlock extracts one block's custom-property declarations from
// theme.css. The parse is STRICT: inside a token block every non-blank
// line must be exactly `--name: #HEX;` — anything else fails with the
// offending line (the file is hand-maintained but golden-locked; a bad
// edit must never pass silently).
func cssVarBlock(t *testing.T, css, selector string) map[string]string {
	t.Helper()
	open := selector + " {"
	start := strings.Index(css, open)
	if start < 0 {
		t.Fatalf("theme.css: selector %q not found", selector)
	}
	start += len(open)
	end := strings.Index(css[start:], "\n}")
	if end < 0 {
		t.Fatalf("theme.css: block for %q not terminated", selector)
	}
	body := css[start : start+end]

	decl := regexp.MustCompile(`^\s*--([a-z0-9-]+):\s*(#[0-9A-Fa-f]{6});\s*$`)
	out := map[string]string{}
	for i, line := range strings.Split(body, "\n") {
		if strings.TrimSpace(line) == "" {
			continue
		}
		m := decl.FindStringSubmatch(line)
		if m == nil {
			t.Fatalf("theme.css %s: line %d is not a token declaration: %q", selector, i+1, line)
		}
		if _, dup := out[m[1]]; dup {
			t.Fatalf("theme.css %s: duplicate token --%s", selector, m[1])
		}
		out[m[1]] = m[2]
	}
	return out
}

// kebabToPython maps the CSS twin's kebab names back to the reference's
// table keys (sheet-alt → sheet_alt) — the deterministic inverse of the
// generator-side convention.
func kebabToPython(name string) string {
	return strings.ReplaceAll(name, "-", "_")
}

func TestThemeCSSDumpMatchesGolden(t *testing.T) {
	raw, err := os.ReadFile("../frontend/theme.css")
	if err != nil {
		t.Fatalf("reading ../frontend/theme.css: %v", err)
	}
	css := string(raw)

	toPythonKeys := func(vars map[string]string) map[string]string {
		out := make(map[string]string, len(vars))
		for k, v := range vars {
			out[kebabToPython(k)] = v
		}
		return out
	}

	got := canon(t, map[string]any{
		"schema": 1,
		"light":  toPythonKeys(cssVarBlock(t, css, ":root")),
		"dark":   toPythonKeys(cssVarBlock(t, css, `[data-theme="dark"]`)),
	})
	want := loadGolden(t, "theme_tokens.golden.json")
	if !bytes.Equal(got, want) {
		var g, w themeDoc
		if err := json.Unmarshal(got, &g); err != nil {
			t.Fatalf("drift, and the got-doc does not unmarshal: %v", err)
		}
		if err := json.Unmarshal(want, &w); err != nil {
			t.Fatalf("golden does not unmarshal (regenerate it): %v", err)
		}
		diffTheme(t, &g, &w)
		t.Fatalf("the CSS var dump drifted from the reference golden")
	}
}

// ---------------------------------------------------------------------------
// 3. The icon pack: the Go twin + the JS twin vs the icons golden.
// ---------------------------------------------------------------------------

func TestIconsGoPackMatchesGolden(t *testing.T) {
	got := canon(t, map[string]any{"schema": 1, "icons": Icons}) // map: sorted keys (icons, schema) — the golden's own order
	want := loadGolden(t, "icons.golden.json")
	if !bytes.Equal(got, want) {
		var g, w iconsDoc
		if err := json.Unmarshal(got, &g); err != nil {
			t.Fatalf("drift, and the got-doc does not unmarshal: %v", err)
		}
		if err := json.Unmarshal(want, &w); err != nil {
			t.Fatalf("golden does not unmarshal (regenerate it): %v", err)
		}
		for k, v := range w.Icons {
			if g.Icons[k] != v {
				t.Errorf("icons.%s: Go pack has %q, golden has %q", k, g.Icons[k], v)
			}
		}
		for k := range g.Icons {
			if _, ok := w.Icons[k]; !ok {
				t.Errorf("icons.%s: Go pack has an extra glyph the golden lacks", k)
			}
		}
		t.Fatalf("the Go icon pack drifted from the reference golden")
	}
	if len(Icons) != 28 {
		t.Fatalf("icon count: %d, want 28", len(Icons))
	}
}

// jsIconEntry matches one icons.js pack line: the file's single pack
// object, one entry per line, strict shape — a hand-maintained file
// locked by the golden, so a bad edit must never pass silently.
var jsIconEntry = regexp.MustCompile("^[[:space:]]*'([a-z0-9-]+)':[[:space:]]+`([^`]*)`,[[:space:]]*(//.*)?$")

func TestIconsJSPackMatchesGolden(t *testing.T) {
	raw, err := os.ReadFile("../frontend/icons.js")
	if err != nil {
		t.Fatalf("reading ../frontend/icons.js: %v", err)
	}
	pack := map[string]string{}
	inPack := false
	for i, line := range strings.Split(string(raw), "\n") {
		if strings.Contains(line, "const PACK = {") {
			inPack = true
			continue
		}
		if inPack && strings.HasPrefix(strings.TrimSpace(line), "};") {
			inPack = false
			continue
		}
		if !inPack || strings.TrimSpace(line) == "" || strings.HasPrefix(strings.TrimSpace(line), "//") {
			continue
		}
		m := jsIconEntry.FindStringSubmatch(line)
		if m == nil {
			t.Fatalf("icons.js line %d is not a pack entry: %q", i+1, line)
		}
		if _, dup := pack[m[1]]; dup {
			t.Fatalf("icons.js: duplicate glyph %q", m[1])
		}
		pack[m[1]] = m[2]
	}
	if len(pack) == 0 {
		t.Fatalf("icons.js: no pack entries extracted")
	}
	got := canon(t, map[string]any{"schema": 1, "icons": pack}) // map: sorted keys — the golden's own order
	want := loadGolden(t, "icons.golden.json")
	if !bytes.Equal(got, want) {
		var g, w iconsDoc
		if err := json.Unmarshal(got, &g); err != nil {
			t.Fatalf("drift, and the got-doc does not unmarshal: %v", err)
		}
		if err := json.Unmarshal(want, &w); err != nil {
			t.Fatalf("golden does not unmarshal (regenerate it): %v", err)
		}
		for k, v := range w.Icons {
			if g.Icons[k] != v {
				t.Errorf("icons.js %q: has %q, golden has %q", k, g.Icons[k], v)
			}
		}
		t.Fatalf("the JS icon pack drifted from the reference golden")
	}
}

// ---------------------------------------------------------------------------
// 4. The Go helpers' own laws.
// ---------------------------------------------------------------------------

func TestRGBA(t *testing.T) {
	cases := []struct {
		table map[string]string
		token string
		want  [4]uint8
	}{
		{Light, "window", [4]uint8{251, 248, 242, 255}},        // the cream window
		{Dark, "window", [4]uint8{34, 30, 46, 255}},            // #221E2E
		{Light, "progress_chunk", [4]uint8{66, 179, 107, 255}}, // #42B36B (green, SESSION-077)
	}
	for _, c := range cases {
		r, g, b, a := RGBA(c.table, c.token)
		if [4]uint8{r, g, b, a} != c.want {
			t.Errorf("RGBA(%s): got %d,%d,%d,%d want %v", c.token, r, g, b, a, c.want)
		}
	}
	// the honest-refusal law: unknown tokens → zero RGBA, never a guess.
	if r, g, b, a := RGBA(Light, "no_such_token"); r|g|b|a != 0 {
		t.Errorf("RGBA(unknown) = %d,%d,%d,%d, want all zero", r, g, b, a)
	}
}

func TestSVGLaws(t *testing.T) {
	// the tint substitution: '{color}' → the concrete tint (the _svg twin).
	doc := SVG(Icons, "stop", "#5F54B4")
	if !strings.Contains(doc, `fill="#5F54B4" stroke="none"`) {
		t.Errorf("stop glyph not tinted: %s", doc)
	}
	if !strings.HasPrefix(doc, `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"`) {
		t.Errorf("svg doc shape wrong: %s", doc)
	}
	// the graceful-degradation law: an unknown glyph is an EMPTY svg.
	if got := SVG(Icons, "no-such-glyph", "#FFF"); got != `<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 24 24"></svg>` {
		t.Errorf("unknown glyph must degrade to the empty svg, got: %s", got)
	}
	// the log level map's glyphs all exist in the pack (the desk's law:
	// ui.py:2144 — success:check, warning:triangle-alert, error:circle-x,
	// info:dot).
	for _, name := range []string{"check", "triangle-alert", "circle-x", "dot"} {
		if _, ok := Icons[name]; !ok {
			t.Errorf("the log level glyph %q is missing from the pack", name)
		}
	}
}

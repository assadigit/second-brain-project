// extract_test.go — T-32: the port of the reference's own TestWebExtract
// battery (app/tests/test_phase2.py — the characterization the Python
// app itself runs), plus the package's own law tests. The golden
// battery already byte-compares every field; these tests keep the
// reference's assertions alive verbatim (the acceptance contract).
package webextract

import (
	"strings"
	"testing"
)

// page loads one of the reference's own web fixtures through the
// bytes entry point (the reference's _page helper).
func page(t *testing.T, name string) Extracted {
	t.Helper()
	golden := loadGolden(t)
	fixtures := section(t, golden, "fixtures")
	raw, ok := fixtures[name]
	if !ok {
		t.Fatalf("fixture %q not in golden", name)
	}
	body := mustHex(t, raw.(map[string]any)["_body_hex"].(string))
	return ExtractFromBytes(body, "")
}

func TestReferenceArticle(t *testing.T) {
	p := page(t, "article.html")
	if p.Title != "The Practical Typography Guide" {
		t.Errorf("title = %q", p.Title)
	}
	if !strings.Contains(p.MetaDescription, "typography") {
		t.Errorf("meta description = %q", p.MetaDescription)
	}
	if !strings.Contains(p.Text, "twelve-column grid") {
		t.Errorf("text missing the grid passage: %q", p.Text[:min(80, len(p.Text))])
	}
	if p.IsJsShell || p.IsPaywall || !p.HasContent() {
		t.Errorf("flags: js=%v paywall=%v content=%v", p.IsJsShell, p.IsPaywall, p.HasContent())
	}
}

func TestReferenceLandingPage(t *testing.T) {
	p := page(t, "landing.html")
	if p.Title != "PaletteForge — color palettes that ship" {
		t.Errorf("title = %q", p.Title)
	}
	if !strings.Contains(p.MetaDescription, "40,000") {
		t.Errorf("meta description = %q", p.MetaDescription)
	}
	if !strings.Contains(p.Text, "WCAG-checked palettes") {
		t.Errorf("text missing the features list")
	}
	if p.IsJsShell {
		t.Errorf("landing must not be a js shell")
	}
}

func TestReferenceJsShell(t *testing.T) {
	p := page(t, "js_shell.html")
	if !p.IsJsShell {
		t.Errorf("js_shell.html must be a js shell")
	}
	// Title + meta description exist -> there IS content to classify
	// from (the model gets title/description only; fetch_status will
	// be 'partial').
	if !p.HasContent() {
		t.Errorf("a js shell with title+description still has content")
	}
}

func TestReferencePaywall(t *testing.T) {
	p := page(t, "paywall.html")
	if !p.IsPaywall {
		t.Errorf("paywall.html must be a paywall")
	}
	if !strings.Contains(strings.ToLower(p.Title), "supply-chain") {
		t.Errorf("title = %q", p.Title)
	}
	if !p.HasContent() {
		t.Errorf("a paywall with real title + description has content")
	}
}

func TestReferenceLatin1Bytes(t *testing.T) {
	p := page(t, "latin1.html")
	if !strings.Contains(p.Title, "Diseño") {
		t.Errorf("title = %q (accents lost?)", p.Title)
	}
	if !strings.Contains(p.Text, "tipográfica") {
		t.Errorf("text lost the accented passage")
	}
}

func TestReferenceEmptyHtml(t *testing.T) {
	p := Extract("")
	if !p.IsJsShell {
		t.Errorf("empty html is a js shell")
	}
	if p.HasContent() {
		t.Errorf("empty html has no content")
	}
}

func TestReferenceHugeTextIsCapped(t *testing.T) {
	huge := "<html><body>" + strings.Repeat("<p>word ", 200000) + "</p></body></html>"
	p := Extract(huge)
	if got := len([]rune(p.Text)); got > MaxTextChars {
		t.Errorf("text cap violated: %d runes", got)
	}
}

// --- the package's own law tests ------------------------------------------

func TestExtractedHasContentThresholds(t *testing.T) {
	cases := []struct {
		e    Extracted
		want bool
	}{
		{Extracted{}, false},
		{Extracted{Title: "t"}, true},
		{Extracted{MetaDescription: "d"}, true},
		{Extracted{Text: strings.Repeat("x", 79)}, false},
		{Extracted{Text: strings.Repeat("x", 80)}, true},
		{Extracted{Text: strings.Repeat("é", 80)}, true}, // runes, not bytes
		{Extracted{Text: strings.Repeat("é", 40)}, false},
	}
	for i, c := range cases {
		if got := c.e.HasContent(); got != c.want {
			t.Errorf("case %d: HasContent = %v, want %v", i, got, c.want)
		}
	}
}

func TestPaywallWindowIsFirst4000Runes(t *testing.T) {
	// marker inside the window → flagged
	early := "<p>" + strings.Repeat("filler ", 100) + "subscribe to continue reading</p>"
	if !Extract(early).IsPaywall {
		t.Errorf("marker inside the 4000-char window must flag")
	}
	// marker beyond the window (but under the 20000 cap) → not flagged
	late := "<p>" + strings.Repeat("filler words here ", 500) +
		"subscribe to continue reading</p>"
	if Extract(late).IsPaywall {
		t.Errorf("marker beyond the 4000-char window must not flag")
	}
}

func TestJsShellWordCountBoundary(t *testing.T) {
	// 119 word chars + a script tag → shell
	if p := Extract("<script></script>" + strings.Repeat("x", 119)); !p.IsJsShell {
		t.Errorf("119 word chars with a script is a js shell")
	}
	// 120 → not a shell
	if p := Extract("<script></script>" + strings.Repeat("x", 120)); p.IsJsShell {
		t.Errorf("120 word chars is not a js shell")
	}
	// tiny page without any script tag → not a shell
	if p := Extract(strings.Repeat("x", 50)); p.IsJsShell {
		t.Errorf("no script tag means no js shell")
	}
}

func TestScriptWordBoundary(t *testing.T) {
	// '<scripty>' — the \b after 'script' fails on a word char: no
	// script tag, so a tiny page is NOT a js shell.
	if p := Extract("<scripty>tiny</scripty>"); p.IsJsShell {
		t.Errorf("<scripty> must not count as a script tag")
	}
	// UPPERCASE counts (re.IGNORECASE)
	if p := Extract("<SCRIPT >tiny</SCRIPT>"); !p.IsJsShell {
		t.Errorf("<SCRIPT> counts as a script tag")
	}
	// a non-ASCII word char after "script" also defeats the boundary
	// (Python's \b is Unicode-aware; a bare ASCII \b would match).
	if p := Extract("<scripté>tiny</scripté>"); p.IsJsShell {
		t.Errorf("<scripté> must not count as a script tag")
	}
}

func TestCharsetBackfillLaw(t *testing.T) {
	// no meta charset + no header → utf-8 backfill
	p := ExtractFromBytes([]byte("<p>plain ascii body words</p>"), "")
	if p.CharsetUsed != "utf-8" {
		t.Errorf("charset_used = %q, want utf-8", p.CharsetUsed)
	}
	// unknown header charset that cannot decode → backfilled verbatim
	p = ExtractFromBytes([]byte("<p>plain ascii body words</p>"), "x-unknown")
	if p.CharsetUsed != "x-unknown" {
		t.Errorf("charset_used = %q, want x-unknown", p.CharsetUsed)
	}
	// the page's own meta charset wins over both
	p = ExtractFromBytes([]byte("<meta charset=\"koi8-r\"><p>ok</p>"), "")
	if p.CharsetUsed != "koi8-r" {
		t.Errorf("charset_used = %q, want koi8-r", p.CharsetUsed)
	}
}

func TestExtractNeverPanics(t *testing.T) {
	// a fuzz-ish sweep of hostile inputs — extraction never fails.
	hostile := []string{
		"<", "<<<<", "</", "</>", "<!", "<!-", "<!--", "<!---", "<![CDATA[",
		"<a href=", "<a=\"1\"", "<p a=\"", "\x00\x01\x02", "&", "&#", "&x",
		"&#99999999999999999999999999;", "&" + strings.Repeat("a", 100) + ";",
		"<title>", "<script>", "<textarea>", "<xmp>", strings.Repeat("<p>", 500),
		"<p>" + strings.Repeat("&amp;", 3000) + "</p>",
		"<div " + strings.Repeat("a='b' ", 2000) + ">",
		strings.Repeat("<!DOCTYPE html>", 200),
		"<svg><svg><svg>" + strings.Repeat("x", 50),
	}
	for _, h := range hostile {
		p := Extract(h) // must not panic
		_ = p.HasContent()
		_ = ExtractFromBytes([]byte(h), "") // must not panic
	}
}

func TestEmptyInputs(t *testing.T) {
	for _, s := range []string{"", " ", "\n\t "} {
		p := Extract(s)
		if !p.IsJsShell || p.HasContent() || p.Text != "" {
			t.Errorf("Extract(%q) = %+v", s, p)
		}
	}
	if p := ExtractFromBytes(nil, ""); !p.IsJsShell || p.HasContent() {
		t.Errorf("ExtractFromBytes(nil) = %+v", p)
	}
	if p := ExtractFromBytes([]byte("  \n "), ""); !p.IsJsShell {
		t.Errorf("whitespace body = %+v", p)
	}
}

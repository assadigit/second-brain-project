// verdict.go — the page verdict (chrome_tabs.py:2042-2135): is this
// tab the SITE, or Chrome's own answer to a failure? Pure — no
// socket, no browser; the exact reason strings surface in logs,
// reports and modals.
package chrometabs

import (
	"regexp"
	"strings"

	"github.com/assadigit/second-brain-project/internal/helpers"
)

// errCodeRe mirrors _ERR_CODE_RE: Chrome's own error pages carry an
// ERR_* code in the DOM (locale-independent — the code, not the
// prose, is printed for every locale).
var errCodeRe = regexp.MustCompile(`ERR_[A-Z0-9_]{2,}`)

// crashMarkers mirrors _CRASH_MARKERS (the renderer's own sad-tab
// pages).
var crashMarkers = []string{"aw, snap", "he's dead, jim", "err_crashed"}

// challengeTitleStrong mirrors _CHALLENGE_TITLE_STRONG: the
// interstitial's own name — decisive wherever it appears.
var challengeTitleStrong = []string{
	"just a moment", "checking your browser", "verify you are human",
	"attention required", "ddos protection",
	"enable javascript and cookies",
}

// challengeTitleWeak mirrors _CHALLENGE_TITLE_WEAK: real pages carry
// these words inside long titles; they only count on an
// interstitial's SHORT title.
var challengeTitleWeak = []string{
	"blocked", "access denied", "security check",
}

// challengeHTMLMarkers mirrors _CHALLENGE_HTML_MARKERS: the
// Cloudflare machinery's own script names — only a wall on a page
// that is otherwise THIN.
var challengeHTMLMarkers = []string{
	"challenge-platform", "cf-chl", "__cf_chl", "cf_chl_opt",
	"turnstile", "cdn-cgi/challenge", "cf-browser-verification",
}

const (
	// weakTitleMaxLen mirrors _WEAK_TITLE_MAX_LEN.
	weakTitleMaxLen = 60
	// weakTitleMaxWords mirrors _WEAK_TITLE_MAX_WORDS.
	weakTitleMaxWords = 4
	// challengeTextFloor mirrors _CHALLENGE_TEXT_FLOOR.
	challengeTextFloor = 500
	// minRealHTMLChars mirrors _MIN_REAL_HTML_BYTES (Python's len()
	// on str counts CHARACTERS — the floor is char-shaped).
	minRealHTMLChars = 200
)

// visibleTextLen mirrors _visible_text_len: how much VISIBLE text a
// DOM carries — script, style, noscript, template and comment blocks
// stripped, tags stripped, entities collapsed, whitespace squeezed.
// The challenge verdict's thinness signal. Pure, best-effort, never
// fails (an unmeasurable page is not a thin one).
//
// The reference's first pass rides a backreferencing regex
// (?is)<(script|style|noscript|template)\b[^>]*>.*?</\1\s*> — RE2
// has no backreferences, so the port scans the same grammar by hand:
// an open tag <name…> (the \b word boundary after the name), then
// the FIRST case-insensitive </name\s*> closes the block.
func visibleTextLen(html string) int {
	text := stripScriptStyleBlocks(html)
	text = stripComments(text)
	text = stripTags(text)
	text = stripEntities(text)
	return len([]rune(helpers.Strip(helpers.CollapseWhitespace(text))))
}

// stripScriptStyleBlocks removes <script|style|noscript|template …>
// … </name> blocks (case-insensitive, the close tag's name may carry
// trailing whitespace).
func stripScriptStyleBlocks(html string) string {
	var b strings.Builder
	i := 0
	for i < len(html) {
		lt := strings.IndexByte(html[i:], '<')
		if lt < 0 {
			b.WriteString(html[i:])
			break
		}
		start := i + lt
		rest := html[start+1:]
		name, ok := matchBlockName(rest)
		if !ok {
			// not a block open — keep the '<' and move on
			b.WriteString(html[i : start+1])
			i = start + 1
			continue
		}
		// find the open tag's '>'
		gt := strings.IndexByte(rest, '>')
		if gt < 0 {
			// no '>' — the regex's [^>]*> cannot match; keep as-is
			b.WriteString(html[i:])
			break
		}
		// find the first close </name\s*> after the open tag
		closeIdx := findCloseTag(html[start+1+gt+1:], name)
		if closeIdx < 0 {
			// unterminated block — the non-greedy regex needs its
			// close; the text stays
			b.WriteString(html[i : start+1])
			i = start + 1
			continue
		}
		b.WriteString(html[i:start]) // the text before the block
		b.WriteByte(' ')
		i = start + 1 + gt + 1 + closeIdx
	}
	return b.String()
}

// blockNames mirrors the regex's alternation set.
var blockNames = []string{"script", "style", "noscript", "template"}

// matchBlockName answers whether s starts with one of the block
// names followed by a word boundary (the regex's \b: the next char is
// not [A-Za-z0-9_]).
func matchBlockName(s string) (string, bool) {
	for _, name := range blockNames {
		if len(s) >= len(name) && strings.EqualFold(s[:len(name)], name) {
			after := s[len(name):]
			if after == "" {
				return name, true
			}
			c := after[0]
			if !(c >= 'a' && c <= 'z' || c >= 'A' && c <= 'Z' ||
				c >= '0' && c <= '9' || c == '_') {
				return name, true
			}
		}
	}
	return "", false
}

// findCloseTag finds the first case-insensitive "</name" + \s* + ">"
// in s; returns the index just past the close tag's '>' (-1 when
// absent).
func findCloseTag(s, name string) int {
	lower := strings.ToLower(s)
	want := "</" + name
	from := 0
	for {
		idx := strings.Index(lower[from:], want)
		if idx < 0 {
			return -1
		}
		at := from + idx + len(want)
		// \s* then '>'
		for at < len(s) && isPySpaceByte(s[at]) {
			at++
		}
		if at < len(s) && s[at] == '>' {
			return at + 1
		}
		from = from + idx + 1
	}
}

// isPySpaceByte answers Python \s for the ASCII range (the close-tag
// grammar's whitespace).
func isPySpaceByte(c byte) bool {
	switch c {
	case ' ', '\t', '\n', '\r', '\v', '\f':
		return true
	}
	return false
}

var (
	commentRe = regexp.MustCompile(`(?s)<!--.*?-->`)
	tagRe     = regexp.MustCompile(`(?s)<[^>]*>`)
	entityRe  = regexp.MustCompile(`&[a-zA-Z#0-9]{1,10};`)
)

func stripComments(s string) string { return commentRe.ReplaceAllString(s, " ") }
func stripTags(s string) string     { return tagRe.ReplaceAllString(s, " ") }
func stripEntities(s string) string { return entityRe.ReplaceAllString(s, " ") }

// PageIsReal mirrors page_is_real: True when the tab is showing the
// SITE's page. The laws, in order (the strongest sentence wins): the
// tab's own URL must be http(s) — a tab that ended on
// chrome-error:// or about: is Chrome's page, not the site's; the DOM
// must not carry an ERR_* code; it must not read as a crash page; it
// must not read as a bot challenge — a STRONG interstitial title
// phrase is decisive wherever it appears, a WEAK word only counts on
// a SHORT title (≤60 chars, ≤4 words), an HTML marker only counts on
// a page that is otherwise THIN (<500 visible characters). Returns
// (ok, reason) — reason is ” when ok. Pure.
func PageIsReal(pageHTML, href, title string) (bool, string) {
	href = strings.TrimSpace(href)
	lowered := helpers.PyLower(pageHTML)
	t := helpers.PyLower(helpers.Strip(title))
	proto, _, _ := pyURLParts(href)
	if proto != "http" && proto != "https" {
		shown := href
		if shown == "" {
			shown = "an empty URL"
		}
		return false, "the tab ended on " + helpers.PyRepr(shown) +
			" — Chrome's own page, not the site's"
	}
	if m := errCodeRe.FindString(pageHTML); m != "" {
		return false, "the tab showed Chrome's error page (" + m +
			") — the site never loaded"
	}
	for _, marker := range crashMarkers {
		if strings.Contains(lowered, marker) || strings.Contains(t, marker) {
			return false, "the tab crashed ('Aw, Snap' — the renderer died)"
		}
	}
	for _, marker := range challengeTitleStrong {
		if strings.Contains(t, marker) {
			return false, "the page stayed on a bot challenge (title: " +
				helpers.PyRepr(marker) + ")"
		}
	}
	if len([]rune(t)) <= weakTitleMaxLen && len(strings.Fields(t)) <= weakTitleMaxWords {
		for _, marker := range challengeTitleWeak {
			if strings.Contains(t, marker) {
				return false, "the page stayed on a bot challenge (title: " +
					helpers.PyRepr(marker) + ")"
			}
		}
	}
	if visibleTextLen(pageHTML) < challengeTextFloor {
		for _, marker := range challengeHTMLMarkers {
			if strings.Contains(lowered, marker) {
				return false, "the page stayed on a bot challenge (" + marker + ")"
			}
		}
	}
	return true, ""
}

// isSiteURL mirrors _is_site_url: an http(s) URL with a host — the
// load watch's own cheap check (a tab on about:blank or
// chrome-error:// is still LOADING as far as the door is concerned).
func isSiteURL(href string) bool {
	scheme, _, hasNetloc := pyURLParts(href)
	return (scheme == "http" || scheme == "https") && hasNetloc
}

// splitProbe mirrors _split_probe: 'complete|https://…' →
// ('complete', 'https://…') — the load watch reads readyState and the
// tab's URL in ONE DevTools call.
func splitProbe(probe string) (string, string) {
	if idx := strings.Index(probe, "|"); idx >= 0 {
		state := strings.TrimSpace(probe[:idx])
		href := strings.TrimSpace(probe[idx+1:])
		return state, href
	}
	return strings.TrimSpace(probe), ""
}

package websitepipeline

import (
	"os"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/assadigit/second-brain-project/internal/links"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// sitelinks.go — v0.64.0 ONE NOTE PER SITE: the consolidation's own
// writer (add_site_links_to_note + its parse/format helpers), the
// ownership probe (_note_lists_site_link), and a note's own placement
// read (_read_note_placement).

// LogFn mirrors the reference's optional log callback: log(msg, level).
type LogFn = func(msg, level string)

func noopLog(msg, level string) {}

// siteLinksLineRE mirrors _SITE_LINKS_LINE_RE:
// ^site_links:\s*(.*)$.
var siteLinksLineRE = regexp.MustCompile(`^site_links:[\s\v\x1c-\x1f\x85]*(.*)$`)

// ParseSiteLinksList mirrors _parse_site_links_list: one YAML-ish
// single-line list (site_links: [a, b]) → its items. The same
// string-scan the tags parse uses; canonical URLs never carry commas,
// so a comma split is the honest read.
func ParseSiteLinksList(raw string) []string {
	inner := strings.TrimSpace(raw)
	if strings.HasPrefix(inner, "[") && strings.HasSuffix(inner, "]") {
		inner = inner[1 : len(inner)-1]
	}
	var out []string
	for _, t := range strings.Split(inner, ",") {
		t = strings.TrimSpace(t)
		t = strings.Trim(t, `"`)
		t = strings.Trim(t, `'`)
		t = strings.TrimSpace(t)
		if t != "" {
			out = append(out, t)
		}
	}
	return out
}

// FmtSiteLinksList mirrors _fmt_site_links_list.
func FmtSiteLinksList(urls []string) string {
	return "site_links: [" + strings.Join(urls, ", ") + "]"
}

// SiteLinksResult mirrors add_site_links_to_note's return:
// {'appended': [urls actually added], 'already': n, 'written': bool}.
type SiteLinksResult struct {
	Appended []string
	Already  int
	Written  bool
}

// AddSiteLinksToNote mirrors add_site_links_to_note (v0.64.0 — ONE
// NOTE PER SITE, the append-only writer): every URL of the same site
// that is not already recorded rides TWO places — the frontmatter's
// site_links line (the machine's half: VaultIndex parses it) and the
// body's "## Links on this site" section (the owner's half). Every
// existing byte is preserved; idempotent; dry-run aware; an
// unexpected shape is skipped with a warning, never mangled. Never
// raises.
func AddSiteLinksToNote(path string, urls []string, log LogFn) SiteLinksResult {
	if log == nil {
		log = noopLog
	}
	out := SiteLinksResult{}
	if path == "" || len(urls) == 0 {
		return out
	}
	if _, err := os.Stat(path); err != nil {
		return out
	}
	var canon []string
	seen := map[string]bool{}
	for _, u := range urls {
		c := links.NormalizeWebsiteURL(u)
		if c == "" {
			c = strings.TrimSpace(u)
		}
		if c != "" && !seen[c] {
			seen[c] = true
			canon = append(canon, c)
		}
	}
	if len(canon) == 0 {
		return out
	}
	data, err := os.ReadFile(path)
	if err != nil {
		log("⚠️ One note per site: could not read "+baseName(path)+": "+err.Error()+
			" — the link is recorded in the ledger only", "warning")
		return out
	}
	content := decodeUTF8Replace(string(data))
	lines := strings.Split(content, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		log("⚠️ One note per site: "+baseName(path)+" carries no frontmatter block — left untouched (the link is recorded in the ledger only)", "warning")
		return out
	}
	// ---- the frontmatter half -------------------------------------------
	fmEnd := -1
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			fmEnd = i
			break
		}
	}
	if fmEnd < 0 {
		log("⚠️ One note per site: "+baseName(path)+"'s frontmatter never closes — left untouched", "warning")
		return out
	}
	var existing []string
	siteLineAt := -1
	sourceLineAt := -1
	for i := 1; i < fmEnd; i++ {
		if m := siteLinksLineRE.FindStringSubmatch(lines[i]); m != nil {
			siteLineAt = i
			existing = ParseSiteLinksList(m[1])
		}
		if strings.HasPrefix(strings.ToLower(strings.TrimSpace(lines[i])), "source:") {
			sourceLineAt = i
		}
	}
	have := map[string]bool{}
	for _, e := range existing {
		have[e] = true
	}
	var add []string
	for _, c := range canon {
		if !have[c] {
			add = append(add, c)
		}
	}
	if len(add) == 0 {
		out.Already = len(canon)
		return out
	}
	merged := append(append([]string{}, existing...), add...)
	newLine := FmtSiteLinksList(merged)
	if siteLineAt >= 0 {
		lines[siteLineAt] = newLine
	} else if sourceLineAt >= 0 {
		lines = insertAt(lines, sourceLineAt+1, newLine)
		fmEnd++
	} else {
		lines = insertAt(lines, 1, newLine)
		fmEnd++
	}
	// ---- the body half ---------------------------------------------------
	body := strings.Join(lines[fmEnd+1:], "\n")
	var newLinksMD []string
	for _, c := range add {
		newLinksMD = append(newLinksMD, "- ["+c+"]("+c+")")
	}
	newLinks := strings.Join(newLinksMD, "\n")
	if strings.Contains(body, SiteLinksHeading) {
		// extend the existing section: append after its last bullet
		bLines := strings.Split(body, "\n")
		headAt := -1
		for i, l := range bLines {
			if strings.TrimSpace(l) == SiteLinksHeading {
				headAt = i
				break
			}
		}
		if headAt >= 0 {
			lastBullet := headAt
			for i := headAt + 1; i < len(bLines); i++ {
				if strings.HasPrefix(strings.TrimSpace(bLines[i]), "- ") {
					lastBullet = i
				}
			}
			bLines = insertAt(bLines, lastBullet+1, newLinks)
			body = strings.Join(bLines, "\n")
		}
	} else {
		section := SiteLinksHeading + "\n" + newLinks + "\n"
		// insert before the closing footer block (the final '---' that
		// precedes the *Source: line); fall back to the very end.
		idx := strings.LastIndex(body, "\n---")
		if idx >= 0 && strings.Contains(body[idx:], "*Source:") {
			body = body[:idx+1] + section + body[idx+1:]
		} else {
			body = strings.TrimRight(strings.TrimRight(body, "\n")+"\n\n"+section, "\n") + "\n"
		}
	}
	newContent := strings.Join(lines[:fmEnd+1], "\n") + "\n" + body
	if err := storage.AtomicWriteText(path, newContent); err != nil {
		log("⚠️ One note per site: could not write "+baseName(path)+": "+err.Error()+
			" — the link is recorded in the ledger only", "warning")
		return out
	}
	out.Appended = add
	out.Written = true
	return out
}

// insertAt inserts v at index i (Python's lines.insert(i, v)).
func insertAt(lines []string, i int, v string) []string {
	out := make([]string, 0, len(lines)+1)
	out = append(out, lines[:i]...)
	out = append(out, v)
	out = append(out, lines[i:]...)
	return out
}

// NoteListsSiteLink mirrors _note_lists_site_link (v0.64.0): does the
// note's site_links: frontmatter list carry this canonical link? (ONE
// NOTE PER SITE's ownership proof.) Pure read; False on any doubt.
func NoteListsSiteLink(path, canonical string) bool {
	if path == "" || canonical == "" {
		return false
	}
	if _, err := os.Stat(path); err != nil {
		return false
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	head := decodeUTF8Replace(string(data))
	runes := []rune(head)
	if len(runes) > 4000 {
		head = string(runes[:4000])
	}
	// _SITE_LINKS_LINE_RE.search(head) — multiline search.
	for _, line := range strings.Split(head, "\n") {
		if m := siteLinksLineRE.FindStringSubmatch(line); m != nil {
			for _, c := range ParseSiteLinksList(m[1]) {
				if c == canonical {
					return true
				}
			}
			return false
		}
	}
	return false
}

// NotePlacement mirrors _read_note_placement's tuple: a note's own
// (category, subcategory, fetch_status) from its frontmatter, for the
// consolidation's ledger row. (”, ”, 'full') on any doubt —
// bookkeeping defaults, never failures.
func NotePlacement(path string) (cat, sub, status string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", "", "full"
	}
	head := decodeUTF8Replace(string(data))
	runes := []rune(head)
	if len(runes) > 4000 {
		head = string(runes[:4000])
	}
	lines := strings.Split(head, "\n")
	if len(lines) == 0 || strings.TrimSpace(lines[0]) != "---" {
		return "", "", "full"
	}
	for _, line := range lines[1:] {
		s := strings.TrimSpace(line)
		if s == "---" {
			break
		}
		low := strings.ToLower(s)
		if strings.HasPrefix(low, "category:") {
			cat = stripQuotes(strings.TrimSpace(s[len("category:"):]))
		} else if strings.HasPrefix(low, "subcategory:") {
			sub = stripQuotes(strings.TrimSpace(s[len("subcategory:"):]))
		} else if strings.HasPrefix(low, "fetch_status:") {
			status = stripQuotes(strings.TrimSpace(s[len("fetch_status:"):]))
		}
	}
	if status == "" {
		status = "full"
	}
	return cat, sub, status
}

// stripQuotes mirrors Python's .strip('"').strip("'") (strip ALL
// leading/trailing double quotes, then all single quotes).
func stripQuotes(s string) string {
	return strings.Trim(strings.Trim(s, `"`), `'`)
}

// baseName mirrors os.path.basename.
func baseName(path string) string {
	if i := strings.LastIndexByte(path, '/'); i >= 0 {
		return path[i+1:]
	}
	if i := strings.LastIndexByte(path, '\\'); i >= 0 {
		return path[i+1:]
	}
	return path
}

// decodeUTF8Replace mirrors open(..., encoding='utf-8',
// errors='replace') — invalid bytes become U+FFFD with Python's
// maximal-subpart placement: each maximal invalid subpart (the longest
// prefix that is still a prefix of some valid sequence) collapses into
// ONE U+FFFD (notestate's own law, pinned byte-exact by its golden;
// T-41's reality-wins fix — the previous version consumed only the
// lead byte, emitting two U+FFFDs where Python emits one for a
// truncated sequence like b"\xe2\x82"). A literal U+FFFD in the input
// stays exactly the bytes it was.
func decodeUTF8Replace(s string) string {
	if utf8.ValidString(s) {
		return s
	}
	var b strings.Builder
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		if r != utf8.RuneError || size != 1 {
			// Valid sequence — keep the original bytes.
			b.WriteString(s[i : i+size])
			i += size
			continue
		}
		b.WriteRune(utf8.RuneError)
		i += subpartLen(s[i:])
	}
	return b.String()
}

// subpartLen returns the length of the maximal invalid subpart starting
// at s[0]: the lead byte plus however many following bytes are legal
// continuations of a still-possibly-valid sequence (lead-specific range
// for the first continuation, 0x80-0xBF for the rest). C0/C1 (overlong
// 2-byte leads), 0x80-0xBF (bare continuations) and 0xF5-0xFF are
// invalid leads on their own.
func subpartLen(s string) int {
	lead := s[0]
	var expected int
	switch {
	case lead < 0xC2: // bare continuation, C0, C1
		return 1
	case lead < 0xE0: // C2-DF: one continuation
		expected = 1
	case lead < 0xF0: // E0-EF: two
		expected = 2
	case lead < 0xF5: // F0-F4: three
		expected = 3
	default: // F5-FF
		return 1
	}
	n := 1
	for i := 1; i <= expected && i < len(s); i++ {
		c := s[i]
		if c < 0x80 || c > 0xBF {
			break
		}
		// Lead-specific first-continuation ranges (anti-overlong and
		// anti-surrogate): E0 needs A0-BF, ED needs 80-9F, F0 needs
		// 90-BF, F4 needs 80-8F.
		if i == 1 {
			switch {
			case lead == 0xE0 && c < 0xA0:
				return n
			case lead == 0xED && c > 0x9F:
				return n
			case lead == 0xF0 && c < 0x90:
				return n
			case lead == 0xF4 && c > 0x8F:
				return n
			}
		}
		n++
	}
	return n
}

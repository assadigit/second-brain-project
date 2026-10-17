// noteparse.go — parse_note (directory_sync.py lines 216-383): the
// note-section parser (UI spec §5.5 + §11.3 — the JS twin lives in the
// Worker's src/directory.js; the grammar is kept identical on purpose).
// One website note → the drawer's structured fields; None when the
// text is not directory-shaped (no frontmatter / no source URL).
package directorysync

import (
	"regexp"
	"strings"
	"unicode"

	"github.com/assadigit/second-brain-project/internal/helpers"
	"github.com/assadigit/second-brain-project/internal/storage"
)

// The parser's regexes (mirrored one-for-one). NOTE the documented
// convergence: CPython's re \s/\S/\b are Unicode-aware on str; Go's
// regexp is ASCII-only. The note grammar's fixtures are
// ASCII-whitespace throughout (the pipeline writes them; the reference
// suite tests none), so the engines agree on every reachable input —
// a future golden that catches a divergence ports the specific regex
// by hand.
var (
	fmLineRe    = regexp.MustCompile(`^([A-Za-z0-9_]+):\s*(.*?)\s*$`)
	h1Re        = regexp.MustCompile(`^#\s+(.+?)\s*$`)
	tldrRe      = regexp.MustCompile(`(?i)^>\s*\*{0,2}\s*TL;DR\s*:?\*{0,2}\s*:?\s*(.*)$`)
	headingRe   = regexp.MustCompile(`^#{2,6}\s+(.+?)\s*$`)
	bulletRe    = regexp.MustCompile(`^\s*[-*•]\s*(.*)$`)
	strayRuleRe = regexp.MustCompile(`^---\s*\S`)
	signupRe    = regexp.MustCompile(`(?i)^sign[- ]?up required\s*:?\s*(yes|no)`)
	h1SearchRe  = regexp.MustCompile(`(?m)^#\s+(.+?)\s*$`)
	footerRe    = regexp.MustCompile(`(?i)^\s*\*Source:`)
	bannerRe    = regexp.MustCompile(`^\s*(?:>[^\n]*\n?)+`)
	blanksRe    = regexp.MustCompile(`\n{3,}`)
)

// banishTags mirrors _BANISH_TAGS: the banish tags that must never
// ride into public tags.
var banishTags = map[string]bool{
	"🗑️": true, "🗑": true, "delete": true, "auto_delete": true,
}

// ExtraSection mirrors one extra_sections entry: {title, bullets, text}.
type ExtraSection struct {
	Title   string
	Bullets []string
	Text    string
}

// ParsedSections mirrors parse_note's sections dict (the insertion
// order rides ToMap).
type ParsedSections struct {
	WhatItDoes       []string
	BestUsedFor      string
	PracticalDetails []string
	WatchOut         []string
	SimilarTools     string
}

// ParsedNote mirrors parse_note's return dict (the typed half); ToMap
// rebuilds the reference's exact dict shape in the reference's key
// order.
type ParsedNote struct {
	Name           string
	Source         string
	Summary        string
	Sections       ParsedSections
	ExtraSections  []ExtraSection
	SignupRequired string
	Pricing        string
	Featured       bool
	NeedsReview    bool
	Tags           []string
	AddedAt        string
	Structured     bool
	DetailMarkdown string
}

// unquote mirrors _unquote: strip one matching pair of quotes the note
// writers may add.
func unquote(value string) string {
	v := helpers.Strip(value)
	if len(v) >= 2 && v[0] == v[len(v)-1] && (v[0] == '"' || v[0] == '\'') {
		// quote chars are single bytes — byte compare is rune-safe here
		return v[1 : len(v)-1]
	}
	return v
}

// sectionKey mirrors _section_key: loose section-label match — strip
// #'s, bold markers, trailing colon, collapse whitespace, lowercase.
func sectionKey(heading string) string {
	key := stripHeadingMarks(heading)
	key = strings.TrimRight(key, ":")
	return pyCollapseSpace(helpers.PyLower(helpers.Strip(key)))
}

// stripHeadingMarks mirrors re.sub(r'[#*_`]', ”, heading).
func stripHeadingMarks(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch r {
		case '#', '*', '_', '`':
			continue
		}
		b.WriteRune(r)
	}
	return b.String()
}

// pyCollapseSpace mirrors re.sub(r'\s+', ' ', key) on the stripped
// string (ASCII \s — the documented convergence above).
func pyCollapseSpace(s string) string {
	var b strings.Builder
	prevSpace := false
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\v' || r == '\f' {
			if !prevSpace {
				b.WriteRune(' ')
				prevSpace = true
			}
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	return b.String()
}

// parseBullets mirrors _parse_bullets: bullet lines only, empty
// bullets dropped.
func parseBullets(lines []string) []string {
	var out []string
	for _, line := range lines {
		m := bulletRe.FindStringSubmatch(line)
		if m == nil {
			continue
		}
		text := helpers.Strip(m[1])
		if text != "" { // empty bullets dropped
			out = append(out, text)
		}
	}
	return out
}

// parseParagraph mirrors _parse_paragraph: ' '.join of the stripped
// non-blank lines.
func parseParagraph(lines []string) string {
	var kept []string
	for _, l := range lines {
		l = helpers.Strip(l)
		if l != "" {
			kept = append(kept, l)
		}
	}
	return strings.Join(kept, " ")
}

// parseTags mirrors _parse_tags: bracket strip, comma split, '#'
// prefix strip, the banish-tag filter (exact AND folded).
func parseTags(raw string) []string {
	s := helpers.Strip(raw)
	if strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]") && len(s) >= 2 {
		s = s[1 : len(s)-1]
	}
	var tags []string
	for _, t := range strings.Split(s, ",") {
		t = helpers.Strip(t)
		if t == "" {
			continue
		}
		t = strings.TrimLeft(t, "#")
		tags = append(tags, t)
	}
	var out []string
	for _, t := range tags {
		if t == "" || banishTags[t] || banishTags[helpers.PyLower(t)] {
			continue
		}
		out = append(out, t)
	}
	return out
}

// ParseNote mirrors parse_note(text). Returns nil when the text is
// not a directory-shaped note (no frontmatter / no source URL). The
// cleaning rules are the UI spec's §5.5: loose heading match, empty
// bullets dropped, stray '--- Name' artifacts cleaned, app chrome
// stripped, missing sections simply absent, unstructured notes keep
// their raw body as detail_markdown.
func ParseNote(text string) *ParsedNote {
	raw := text
	if !strings.HasPrefix(raw, "---") {
		return nil
	}
	lines := strings.Split(raw, "\n")
	close := -1
	for i := 1; i < len(lines); i++ {
		if helpers.Strip(lines[i]) == "---" {
			close = i
			break
		}
	}
	if close == -1 {
		return nil
	}
	fm := map[string]string{}
	for i := 1; i < close; i++ {
		m := fmLineRe.FindStringSubmatch(lines[i])
		if m != nil {
			fm[m[1]] = unquote(m[2])
		}
	}
	source := helpers.Strip(fm["source"])
	if source == "" {
		return nil
	}

	body := strings.Join(lines[close+1:], "\n")

	// Strip the app's ownership banner callout ("> [!info] Managed by
	// …"): the whole leading quote block goes when it carries the info
	// callout, otherwise it stays.
	body = bannerRe.ReplaceAllStringFunc(body, func(m string) string {
		if strings.Contains(helpers.PyLower(m), "[!info]") {
			return ""
		}
		return m
	})

	// Strip the trailing footer: '\n---\n*Source: …*' + the banish
	// block.
	if idx := strings.LastIndex(body, "\n---"); idx != -1 {
		if footerRe.MatchString(body[idx+4:]) {
			body = body[:idx]
		}
	}

	// Clean stray artifacts: '--- Descript' at the top, repeated
	// separators.
	bodyLines := strings.Split(body, "\n")
	first := 0
	for first = 0; first < len(bodyLines); first++ {
		t := helpers.Strip(bodyLines[first])
		if t == "" || strayRuleRe.MatchString(t) || t == "---" {
			continue
		}
		break
	}
	body = strings.Join(bodyLines[first:], "\n")
	body = helpers.Strip(blanksRe.ReplaceAllString(body, "\n\n"))

	// H1 + TL;DR.
	name := ""
	summary := ""
	var rest []string
	for _, line := range strings.Split(body, "\n") {
		if name == "" {
			if h1 := h1Re.FindStringSubmatch(line); h1 != nil {
				name = helpers.Strip(h1[1])
				continue
			}
			if tldr := tldrRe.FindStringSubmatch(line); tldr != nil {
				summary = helpers.Strip(strings.TrimRight(tldr[1], "*"))
				continue
			}
		} else if summary == "" {
			if tldr := tldrRe.FindStringSubmatch(line); tldr != nil {
				summary = helpers.Strip(strings.TrimRight(tldr[1], "*"))
				continue
			}
		}
		rest = append(rest, line)
	}
	if name == "" {
		if h1 := h1SearchRe.FindStringSubmatch(body); h1 != nil {
			name = helpers.Strip(h1[1])
		}
	}

	// Sections.
	sections := ParsedSections{}
	var extraSections []ExtraSection
	signupRequired := "unknown"

	type rawSection struct {
		title string
		key   string
		lines []string
	}
	var rawSections []rawSection
	var current *rawSection
	for _, line := range rest {
		if h := headingRe.FindStringSubmatch(line); h != nil {
			if current != nil {
				rawSections = append(rawSections, *current)
			}
			current = &rawSection{title: h[1], key: sectionKey(h[1])}
			continue
		}
		if current != nil {
			current.lines = append(current.lines, line)
		}
	}
	if current != nil {
		rawSections = append(rawSections, *current)
	}

	for _, sec := range rawSections {
		switch sec.key {
		case "what it does":
			sections.WhatItDoes = parseBullets(sec.lines)
		case "best used for":
			sections.BestUsedFor = parseParagraph(sec.lines)
		case "practical details":
			var bullets []string
			for _, line := range sec.lines {
				m := bulletRe.FindStringSubmatch(line)
				if m == nil {
					continue
				}
				text := helpers.Strip(m[1])
				if text == "" {
					continue
				}
				if su := signupRe.FindStringSubmatch(text); su != nil {
					// became a chip; \b's rune check hand-rolled (the
					// documented ASCII/\b convergence)
					if signupBoundaryOK(text, len(su[0])) {
						signupRequired = helpers.PyLower(su[1])
						continue
					}
				}
				bullets = append(bullets, text)
			}
			sections.PracticalDetails = bullets
		case "watch out":
			sections.WatchOut = parseBullets(sec.lines)
		case "similar tools":
			sections.SimilarTools = parseParagraph(sec.lines)
		default:
			if sec.key != "" {
				bullets := parseBullets(sec.lines)
				paragraph := parseParagraph(sec.lines)
				if len(bullets) > 0 || paragraph != "" {
					extraSections = append(extraSections, ExtraSection{
						Title:   helpers.Strip(sec.title),
						Bullets: bullets,
						Text:    ternary(len(bullets) > 0, "", paragraph),
					})
				}
			}
		}
	}

	fmLogin := helpers.PyLower(helpers.Strip(fm["login_required"]))
	if signupRequired == "unknown" && (fmLogin == "yes" || fmLogin == "no") {
		signupRequired = fmLogin
	}

	pricing := helpers.PyLower(helpers.Strip(pyOrStr(fm["pricing"], "unknown")))
	switch pricing {
	case "free", "freemium", "paid", "unknown":
	default:
		pricing = "unknown"
	}
	featured := helpers.PyLower(helpers.Strip(fm["featured"])) == "true"
	// v0.39.0 — Phase 4 (SPEC §9 P1): the manual "Needs review (maybe
	// not free anymore)" owner flag.
	needsReview := helpers.PyLower(helpers.Strip(fm["needs_review"])) == "true"

	structured := summary != "" || len(sections.WhatItDoes) > 0 ||
		sections.BestUsedFor != "" || len(sections.PracticalDetails) > 0 ||
		len(sections.WatchOut) > 0

	if summary == "—" {
		summary = ""
	}

	return &ParsedNote{
		Name:           name,
		Source:         source,
		Summary:        summary,
		Sections:       sections,
		ExtraSections:  extraSections,
		SignupRequired: signupRequired,
		Pricing:        pricing,
		Featured:       featured,
		NeedsReview:    needsReview,
		Tags:           parseTags(fm["tags"]),
		AddedAt:        helpers.Strip(fm["date_processed"]),
		Structured:     structured,
		DetailMarkdown: ternary(structured, "", body),
	}
}

// signupBoundaryOK mirrors the (yes|no)\b tail check with Python's
// Unicode \b semantics: the next rune must NOT be a word char
// (letter / digit / underscore).
func signupBoundaryOK(text string, end int) bool {
	if end >= len(text) {
		return true
	}
	r := []rune(text[end:])[0]
	return !isPyWordRune(r)
}

// isPyWordRune mirrors Python \w's letter/digit classes.
func isPyWordRune(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// ToMap rebuilds the reference's parse_note dict in the reference's
// key order (name, source, summary, sections, extra_sections,
// signup_required, pricing, featured, needs_review, tags, added_at,
// structured, detail_markdown).
func (p *ParsedNote) ToMap() *storage.OrderedMap {
	m := storage.NewOrderedMap()
	m.Set("name", p.Name)
	m.Set("source", p.Source)
	m.Set("summary", p.Summary)
	sec := storage.NewOrderedMap()
	sec.Set("what_it_does", strSliceAny(p.Sections.WhatItDoes))
	sec.Set("best_used_for", p.Sections.BestUsedFor)
	sec.Set("practical_details", strSliceAny(p.Sections.PracticalDetails))
	sec.Set("watch_out", strSliceAny(p.Sections.WatchOut))
	sec.Set("similar_tools", p.Sections.SimilarTools)
	m.Set("sections", sec)
	extras := make([]any, 0, len(p.ExtraSections))
	for _, e := range p.ExtraSections {
		em := storage.NewOrderedMap()
		em.Set("title", e.Title)
		em.Set("bullets", strSliceAny(e.Bullets))
		em.Set("text", e.Text)
		extras = append(extras, em)
	}
	m.Set("extra_sections", extras)
	m.Set("signup_required", p.SignupRequired)
	m.Set("pricing", p.Pricing)
	m.Set("featured", p.Featured)
	m.Set("needs_review", p.NeedsReview)
	m.Set("tags", strSliceAny(p.Tags))
	m.Set("added_at", p.AddedAt)
	m.Set("structured", p.Structured)
	m.Set("detail_markdown", p.DetailMarkdown)
	return m
}

func strSliceAny(s []string) []any {
	out := make([]any, len(s))
	for i, v := range s {
		out[i] = v
	}
	return out
}

// pyOrStr mirrors (v or fallback): an absent/empty value falls back.
func pyOrStr(v, fallback string) string {
	if v == "" {
		return fallback
	}
	return v
}

func ternary(b bool, t, f string) string {
	if b {
		return t
	}
	return f
}

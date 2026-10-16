package webextract

import (
	"regexp"
	"strings"
)

// parser.go — the port of the reference's _PageParser: an event
// tokenizer byte-faithful to CPython 3.12's html.parser in its
// convert_charrefs=True configuration (the only one the reference
// uses), plus the reference's own handlers (title, meta description,
// charset detection, invisible-depth tracking, block-tag newlines).
//
// This is T-32's zero-dependency alternative: the owner's verdict on
// golang.org/x/net/html did not arrive with SESSION-013's prompt, so
// the reference's own stdlib parser behavior is hand-rolled here
// instead (NEXT-SESSION §4's option b). The tokenizer is an isolated
// ~300-line state machine; an approved x/net/html swap later would
// touch only this file's scan internals — the handler laws stay.
//
// The ported state machine (html/parser.py at the pinned venv's
// CPython 3.12.15, goahead with end=1 — the reference feeds the whole
// document and closes, which is observably a single end-of-input
// pass):
//
//   - Non-raw mode scans for the next '<'; the whole span before it
//     arrives as ONE unescaped data event (the Python "&-near-end"
//     buffering heuristic is a two-phase feed optimization; the
//     close()-time flush it guards emits the same span this pass
//     emits).
//   - Raw mode (script/style/xmp/iframe/noembed/noframes/plaintext —
//     CDATA_CONTENT_ELEMENTS) scans for `</elem` followed by one of
//     \t\n\r\f / or > (case-insensitive, the set_cdata_mode lookahead)
//     and passes everything before it RAW (no charref conversion).
//   - RCDATA mode (textarea/title) is raw mode with escapable=true:
//     the data is unescaped, but tags inside are NOT parsed.
//   - Start tags ride tagfind_tolerant/attrfind_tolerant/locatetagend
//     semantics exactly (attribute names may contain quotes; the
//     lookbehind that lets a new attribute start after a closing
//     quote; the '/'-not-before-'>' guards; values unescaped after
//     quote stripping; names lowercased).
//   - End tags: `</>` is silently consumed, `</` + non-letter is a
//     bogus comment (dropped), a complete end tag fires the handler
//     and ALWAYS clears raw mode.
//   - Incomplete constructs at end-of-input: an incomplete start or
//     end tag discards the rest of the document; an unterminated
//     comment/declaration/PI is dropped; a lone trailing '<' (or the
//     exact string "</") becomes data.
//   - handle_startendtag is the default: start tag + end tag events.
//
// Not ported (dead state in the reference): _PageParser._script_depth
// and _skip_nl are written but never read; HTMLParser's position
// tracking (getpos/updatepos) is never consulted by the reference.

// The reference's tag sets.
var invisibleTags = map[string]bool{
	"script": true, "style": true, "noscript": true, "template": true,
	"svg": true, "head": true,
}

var blockTags = map[string]bool{
	"p": true, "div": true, "br": true, "li": true, "ul": true,
	"ol": true, "tr": true, "table": true, "section": true,
	"article": true, "header": true, "footer": true, "nav": true,
	"main": true, "aside": true, "h1": true, "h2": true, "h3": true,
	"h4": true, "h5": true, "h6": true, "blockquote": true,
	"pre": true, "figure": true, "figcaption": true, "hr": true,
}

// CPython 3.12's CDATA_CONTENT_ELEMENTS and RCDATA_CONTENT_ELEMENTS
// (scripting=False — noscript parses normally).
var cdataContentElements = map[string]bool{
	"script": true, "style": true, "xmp": true, "iframe": true,
	"noembed": true, "noframes": true,
}

var rcdataContentElements = map[string]bool{
	"textarea": true, "title": true,
}

// attr is one (name, value) pair; Python's None value and ” are the
// same thing for every consumer in the reference (dict.get + or ”).
type attr struct {
	Key   string
	Value string
}

type pageParser struct {
	// the reference's collected state
	title           string
	metaDescription string
	detectedCharset string
	inTitle         bool
	invisibleDepth  int
	chunks          []string

	// the tokenizer's raw/RCDATA state (set_cdata_mode)
	cdataElem string // lowercase element name; "" when not in raw mode
	escapable bool
}

// charsetInContentRe is re.search(r'charset=([^\s;]+)', content,
// re.IGNORECASE) with Python's \s class.
var charsetInContentRe = regexp.MustCompile(
	"(?i)charset=([^" + pySInner + ";]+)")

// ---------------------------------------------------------------------------
// The handlers (web_extract.py's _PageParser overrides)
// ---------------------------------------------------------------------------

func (p *pageParser) handleStartTag(tag string, attrs []attr) {
	if invisibleTags[tag] {
		p.invisibleDepth++
		return
	}
	if tag == "title" {
		p.inTitle = true
	}
	if tag == "meta" {
		d := attrsMap(attrs)
		name := asciiLower(d["name"])
		prop := asciiLower(d["property"])
		content := d["content"]
		if name == "description" && content != "" && p.metaDescription == "" {
			p.metaDescription = content
		} else if prop == "og:description" && content != "" && p.metaDescription == "" {
			p.metaDescription = content
		} else if d["charset"] != "" && p.detectedCharset == "" {
			p.detectedCharset = d["charset"]
		} else if asciiLower(d["http-equiv"]) == "content-type" &&
			content != "" && p.detectedCharset == "" {
			if m := charsetInContentRe.FindStringSubmatch(content); m != nil {
				p.detectedCharset = strings.Trim(m[1], "\"'")
			}
		}
	}
	if blockTags[tag] {
		p.chunks = append(p.chunks, "\n")
	}
}

func (p *pageParser) handleEndTag(tag string) {
	if tag == "title" {
		p.inTitle = false
	}
	if invisibleTags[tag] {
		if p.invisibleDepth > 0 {
			p.invisibleDepth--
		}
	}
	if blockTags[tag] {
		p.chunks = append(p.chunks, "\n")
	}
}

func (p *pageParser) handleData(data string) {
	if p.inTitle {
		p.title += data
		return
	}
	if p.invisibleDepth > 0 {
		return
	}
	// if data and data.strip('\n\t ') — ONLY those three characters.
	if data != "" && strings.Trim(data, "\n\t ") != "" {
		p.chunks = append(p.chunks, data)
	}
}

// text is the _PageParser.text property: join, collapse runs of
// spaces/tabs, collapse blank lines, strip (Python whitespace).
var (
	reSpaceRun = regexp.MustCompile(`[ \t]+`)
	reBlankRun = regexp.MustCompile("\n" + pySClass + "*\n+")
)

func (p *pageParser) text() string {
	raw := strings.Join(p.chunks, "")
	raw = reSpaceRun.ReplaceAllString(raw, " ")
	raw = reBlankRun.ReplaceAllString(raw, "\n")
	return pyStrip(raw)
}

func attrsMap(attrs []attr) map[string]string {
	d := make(map[string]string, len(attrs))
	for _, a := range attrs {
		d[a.Key] = a.Value
	}
	return d
}

// ---------------------------------------------------------------------------
// The tokenizer (html/parser.py, convert_charrefs=True, one pass)
// ---------------------------------------------------------------------------

func (p *pageParser) parseAll(raw string) {
	i := 0
	n := len(raw)
	for i < n {
		var j int
		if p.cdataElem == "" {
			j = strings.IndexByte(raw[i:], '<')
			if j < 0 {
				j = n
			} else {
				j += i
			}
		} else {
			j = p.findCdataEnd(raw, i)
			if j < 0 {
				// No more end tags in raw/RCDATA mode: the rest is data.
				p.emitDataSpan(raw[i:n])
				return
			}
		}
		if i < j {
			p.emitDataSpan(raw[i:j])
		}
		if j >= n {
			return
		}
		// raw[j] == '<': the goahead dispatch.
		var k int
		switch {
		case j+1 < n && isASCIILetter(raw[j+1]):
			k = p.parseStarttag(raw, j)
		case strings.HasPrefix(raw[j:], "</"):
			k = p.parseEndtag(raw, j)
		case strings.HasPrefix(raw[j:], "<!--"):
			k = p.parseComment(raw, j)
		case strings.HasPrefix(raw[j:], "<?"):
			k = p.parsePi(raw, j)
		case strings.HasPrefix(raw[j:], "<!"):
			k = p.parseHtmlDeclaration(raw, j)
		default:
			// a lone '<' (not followed by a letter): it is data
			p.handleData("<")
			k = j + 1
		}
		if k < 0 {
			// Incomplete construct at end of input (end=1 semantics).
			if strings.HasPrefix(raw[j:], "</") && j+2 == n {
				p.handleData("</")
			}
			// Every other incomplete shape is consumed silently — the
			// reference registers no comment/decl/PI handlers.
			k = n
		}
		i = k
	}
}

// emitDataSpan is goahead's data emission: unescaped in normal and
// RCDATA mode (convert_charrefs && _escapable), raw in raw mode.
func (p *pageParser) emitDataSpan(s string) {
	if p.cdataElem != "" && !p.escapable {
		p.handleData(s)
		return
	}
	p.handleData(pyUnescape(s))
}

// findCdataEnd is the interesting regex `</elem(?=[\t\n\r\f />])`
// (IGNORECASE|ASCII); -1 when no more end tags exist (the plaintext
// \Z regex never matches inside the buffer either).
func (p *pageParser) findCdataEnd(raw string, i int) int {
	if p.cdataElem == "plaintext" {
		return -1
	}
	target := "</" + p.cdataElem
	off := i
	for {
		k := indexCIASCII(raw, target, off)
		if k < 0 {
			return -1
		}
		after := k + len(target)
		if after >= len(raw) {
			return -1 // the lookahead needs a following byte
		}
		switch raw[after] {
		case '\t', '\n', '\r', '\f', ' ', '/', '>':
			return k
		}
		off = k + 1
	}
}

func (p *pageParser) parseStarttag(raw string, i int) int {
	endpos := locateTagEnd(raw, i+1)
	if endpos < 0 || raw[endpos-1] != '>' {
		return -1
	}

	// tagfind_tolerant: name, then separators ('/' only when not
	// before '>').
	k := i + 2 // raw[i+1] is the leading letter (starttagopen)
	for k < len(raw) && !isTagSep(raw[k]) {
		k++
	}
	nameEnd := k
	for k < len(raw) {
		c := raw[k]
		if isPyTagWS(c) {
			k++
			continue
		}
		if c == '/' && (k+1 >= len(raw) || raw[k+1] != '>') {
			k++
			continue
		}
		break
	}
	tag := asciiLower(raw[i+1 : nameEnd])

	// attrfind_tolerant loop, bounded by the tag end.
	var attrs []attr
	for k < endpos {
		if !attrLookbehind(raw[k-1]) || isTagSep(raw[k]) {
			break
		}
		k2 := k + 1
		for k2 < endpos && !isAttrNameStop(raw[k2]) {
			k2++
		}
		nameEnd2 := k
		// the optional value group: [\t\n\r\f ]*=[\t\n\r\f ]* value
		m := k2
		valStart, valEnd := -1, -1
		v := k2
		for v < endpos && isPyTagWS(raw[v]) {
			v++
		}
		if v < endpos && raw[v] == '=' {
			v++
			for v < endpos && isPyTagWS(raw[v]) {
				v++
			}
			if v < endpos && (raw[v] == '\'' || raw[v] == '"') {
				q := raw[v]
				e := v + 1
				for e < endpos && raw[e] != q {
					e++
				}
				if e < endpos {
					valStart, valEnd = v+1, e
					m = e + 1
				}
				// else: the value alternative fails, the whole group
				// backtracks to empty.
			} else {
				e := v
				for e < endpos && !isPyTagWS(raw[e]) && raw[e] != '>' {
					e++
				}
				valStart, valEnd = v, e
				m = e
			}
		}
		// trailing separators: (?:[\t\n\r\f ]|/(?!>))*
		for m < endpos {
			c := raw[m]
			if isPyTagWS(c) {
				m++
				continue
			}
			if c == '/' && (m+1 >= endpos || raw[m+1] != '>') {
				m++
				continue
			}
			break
		}
		value := ""
		if valStart >= 0 {
			value = raw[valStart:valEnd]
			if value != "" {
				value = pyUnescape(value)
			}
		}
		attrs = append(attrs, attr{Key: asciiLower(raw[nameEnd2:k2]), Value: value})
		k = m
	}

	// the residue check (unreachable in practice — the regexes stay in
	// lockstep — but ported for fidelity).
	end := pyStrip(raw[k:endpos])
	if end != ">" && end != "/>" {
		p.handleData(raw[i:endpos])
		return endpos
	}

	if end == "/>" {
		// handle_startendtag's default: both events.
		p.handleStartTag(tag, attrs)
		p.handleEndTag(tag)
	} else {
		p.handleStartTag(tag, attrs)
		if cdataContentElements[tag] || tag == "plaintext" {
			p.setCDataMode(tag, false)
		} else if rcdataContentElements[tag] {
			p.setCDataMode(tag, true)
		}
	}
	return endpos
}

func (p *pageParser) setCDataMode(elem string, escapable bool) {
	p.cdataElem = elem // already lowercased
	p.escapable = escapable
}

func (p *pageParser) parseEndtag(raw string, i int) int {
	// the fast check: no '>' anywhere after "</"
	gt := strings.IndexByte(raw[i+2:], '>')
	if gt < 0 {
		return -1
	}
	if !isASCIILetter(raw[i+2]) { // endtagopen: </ + letter
		if raw[i+2] == '>' {
			return i + 3 // "</>" is silently consumed
		}
		return parseBogusComment(raw, i)
	}
	j := locateTagEnd(raw, i+2)
	if j < 0 || raw[j-1] != '>' {
		return -1
	}
	// tagfind_tolerant from i+2 (the letter is guaranteed).
	k := i + 3
	for k < len(raw) && !isTagSep(raw[k]) {
		k++
	}
	p.handleEndTag(asciiLower(raw[i+2 : k]))
	// clear_cdata_mode: every complete end tag ends raw/RCDATA mode.
	p.cdataElem = ""
	p.escapable = true
	return j
}

func (p *pageParser) parseComment(raw string, i int) int {
	// commentclose: --!?> (search from i+4)
	if end, ok := findCommentClose(raw, i+4); ok {
		return end
	}
	// commentabruptclose: -?> anchored at i+4
	if i+4 < len(raw) {
		if raw[i+4] == '>' {
			return i + 5
		}
		if raw[i+4] == '-' && i+5 < len(raw) && raw[i+5] == '>' {
			return i + 6
		}
	}
	return -1
}

func (p *pageParser) parsePi(raw string, i int) int {
	pos := strings.IndexByte(raw[i+2:], '>')
	if pos < 0 {
		return -1
	}
	return i + 2 + pos + 1
}

func (p *pageParser) parseHtmlDeclaration(raw string, i int) int {
	if strings.HasPrefix(raw[i:], "<!--") { // handled by goahead; kept for fidelity
		return p.parseComment(raw, i)
	}
	if strings.HasPrefix(raw[i:], "<![CDATA[") {
		j := strings.Index(raw[i+9:], "]]>")
		if j < 0 {
			return -1
		}
		return i + 9 + j + 3 // unknown_decl — dropped
	}
	if len(raw) >= i+9 && asciiLower(raw[i:i+9]) == "<!doctype" {
		g := strings.IndexByte(raw[i+9:], '>')
		if g < 0 {
			return -1
		}
		return i + 9 + g + 1 // handle_decl — dropped
	}
	if strings.HasPrefix(raw[i:], "<![") {
		j := strings.IndexByte(raw[i+3:], '>')
		if j < 0 {
			return -1
		}
		return i + 3 + j + 1 // unknown_decl or comment — both dropped
	}
	return parseBogusComment(raw, i)
}

// parseBogusComment consumes <!…> or </…> to the first '>' (the
// comment itself is dropped — the reference registers no handler).
func parseBogusComment(raw string, i int) int {
	pos := strings.IndexByte(raw[i+2:], '>')
	if pos < 0 {
		return -1
	}
	return i + 2 + pos + 1
}

// findCommentClose finds the leftmost `-->` or `--!>`.
func findCommentClose(raw string, from int) (end int, ok bool) {
	k := from
	for {
		idx := strings.Index(raw[k:], "--")
		if idx < 0 {
			return 0, false
		}
		k += idx
		if k+2 < len(raw) && raw[k+2] == '>' {
			return k + 3, true
		}
		if k+3 < len(raw) && raw[k+2] == '!' && raw[k+3] == '>' {
			return k + 4, true
		}
		k++
	}
}

// locateTagEnd is locatetagend: how far a tag's name+attrs extend,
// starting at pos (i+1 for start tags, i+2 for end tags). Returns the
// match end; the caller checks raw[end-1] == '>' for completeness.
func locateTagEnd(raw string, pos int) int {
	n := len(raw)
	// tag name: [a-zA-Z][^\t\n\r\f />]*
	k := pos + 1
	for k < n && !isTagSep(raw[k]) {
		k++
	}
	// [\t\n\r\f /]* (ANY slash — locatetagend has no (?!>) guard here)
	for k < n && (isPyTagWS(raw[k]) || raw[k] == '/') {
		k++
	}
	// attribute loop
	for k < n {
		if !attrLookbehind(raw[k-1]) || isTagSep(raw[k]) {
			break
		}
		k2 := k + 1
		for k2 < n && !isAttrNameStop(raw[k2]) {
			k2++
		}
		m := k2
		v := k2
		for v < n && isPyTagWS(raw[v]) {
			v++
		}
		if v < n && raw[v] == '=' {
			v++
			for v < n && isPyTagWS(raw[v]) {
				v++
			}
			if v < n && (raw[v] == '\'' || raw[v] == '"') {
				q := raw[v]
				e := v + 1
				for e < n && raw[e] != q {
					e++
				}
				if e < n {
					m = e + 1
				}
				// else: group backtracks to empty (m stays k2)
			} else {
				for v < n && !isPyTagWS(raw[v]) && raw[v] != '>' {
					v++
				}
				m = v
			}
		}
		// [\t\n\r\f /]* (ANY slash)
		for m < n && (isPyTagWS(raw[m]) || raw[m] == '/') {
			m++
		}
		k = m
	}
	// `>?`
	if k < n && raw[k] == '>' {
		k++
	}
	return k
}

// ---------------------------------------------------------------------------
// The character classes (exactly CPython's, byte-wise)
// ---------------------------------------------------------------------------

func isASCIILetter(c byte) bool {
	return (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z')
}

// isPyTagWS is the regex whitespace class [\t\n\r\f ] — NO \v (the
// tag regexes list their classes explicitly).
func isPyTagWS(c byte) bool {
	switch c {
	case '\t', '\n', '\r', '\f', ' ':
		return true
	}
	return false
}

// isTagSep is the tag-name stop set [^\t\n\r\f />].
func isTagSep(c byte) bool {
	return isPyTagWS(c) || c == '/' || c == '>'
}

// isAttrNameStop is the attr-name rest stop set (adds '=').
func isAttrNameStop(c byte) bool {
	return isTagSep(c) || c == '='
}

// attrLookbehind is attrfind_tolerant's (?<=['"\t\n\r\f /]).
func attrLookbehind(c byte) bool {
	switch c {
	case '\'', '"', '\t', '\n', '\r', '\f', ' ', '/':
		return true
	}
	return false
}

// indexCIASCII finds needle (ASCII, starting "</") in s at or after
// from, case-insensitively over ASCII letters.
func indexCIASCII(s, needle string, from int) int {
	if from < 0 {
		from = 0
	}
	first := needle[0] // '<'
	for k := from; k+len(needle) <= len(s); k++ {
		if s[k] != first {
			continue
		}
		match := true
		for m := 1; m < len(needle); m++ {
			a, b := s[k+m], needle[m]
			if a >= 'A' && a <= 'Z' {
				a += 'a' - 'A'
			}
			if b >= 'A' && b <= 'Z' {
				b += 'a' - 'A'
			}
			if a != b {
				match = false
				break
			}
		}
		if match {
			return k
		}
	}
	return -1
}

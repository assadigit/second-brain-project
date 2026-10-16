package webextract

import "strings"

// extract.go — the port of web_extract.py's public surface: the
// Extracted record, extract(html) and extract_from_bytes(body,
// header_charset). Extraction never panics on bad input — a broken
// page is a partial result (the tokenizer is a total function).

// MaxTextChars caps the extracted body text (note-building / LLM
// excerpts).
const MaxTextChars = 20000

// jsShellMinText: a body with fewer alphanumeric characters AND at
// least one <script> tag is assumed to render client-side.
const jsShellMinText = 120

// paywallMarkers: lower-cased substring matches on the first 4000
// characters of the (truncated) text.
var paywallMarkers = []string{
	"subscribe to continue",
	"subscription required",
	"subscribe to read",
	"to continue reading",
	"create a free account to continue",
	"sign in to continue reading",
	"this article is for subscribers",
	"premium subscribers only",
	"unlock this article",
	"reader mode is not available", // some JS shells say this; treat as wall
}

// Extracted is what the extractor found on a page.
type Extracted struct {
	Title           string
	MetaDescription string
	Text            string
	IsJsShell       bool
	IsPaywall       bool
	CharsetUsed     string
}

// HasContent: enough signal (title / description / ≥80 chars of text)
// to classify and analyze the page.
func (e Extracted) HasContent() bool {
	return e.Title != "" || e.MetaDescription != "" || len([]rune(e.Text)) >= 80
}

// newExtracted applies the reference constructor's strips.
func newExtracted(title, metaDescription, text string, isJsShell, isPaywall bool, charsetUsed string) Extracted {
	return Extracted{
		Title:           pyStrip(title),
		MetaDescription: pyStrip(metaDescription),
		Text:            pyStrip(text),
		IsJsShell:       isJsShell,
		IsPaywall:       isPaywall,
		CharsetUsed:     charsetUsed,
	}
}

// Extract pulls title / meta description / visible text out of HTML.
// Never fails: unparseable input returns whatever was collected.
func Extract(html string) Extracted {
	if html == "" || pyStrip(html) == "" {
		return newExtracted("", "", "", true, false, "")
	}
	p := &pageParser{}
	p.parseAll(html)

	title := strings.Join(pySplitWords(p.title), " ")
	text := pyRunesSlice(p.text(), MaxTextChars)
	wordChars := pyWordCharsCount(text)

	// JS-only shell: the page HAS scripts but the served HTML has
	// almost no readable words (a truly empty static page is also "no
	// content").
	isJsShell := wordChars < jsShellMinText && pyHasScriptTag(html)

	lowText := pyLower(pyRunesSlice(text, 4000))
	isPaywall := false
	for _, marker := range paywallMarkers {
		if strings.Contains(lowText, marker) {
			isPaywall = true
			break
		}
	}

	return newExtracted(title, p.metaDescription, text, isJsShell, isPaywall,
		p.detectedCharset)
}

// ExtractFromBytes decodes body (header charset, then the page's own
// <meta charset> sniff, then UTF-8-with-replacement) and extracts.
// Non-UTF-8 pages keep their accented characters instead of turning
// into replacement-character walls.
func ExtractFromBytes(body []byte, headerCharset string) Extracted {
	if len(body) == 0 {
		return newExtracted("", "", "", true, false, "")
	}

	html := pyDecode(body, headerCharset)
	if html == nil {
		// Sniff <meta charset> from the raw bytes (ASCII-compatible
		// prefixes work for utf-8/windows-1252/iso-8859-* alike).
		if cs, ok := sniffCharset(body); ok {
			html = pyDecode(body, cs)
		}
	}
	if html == nil {
		s := pyUTF8Replace(body)
		html = &s
	}

	result := Extract(*html)
	if result.CharsetUsed == "" {
		if headerCharset != "" {
			result.CharsetUsed = headerCharset
		} else {
			result.CharsetUsed = "utf-8"
		}
	}
	return result
}

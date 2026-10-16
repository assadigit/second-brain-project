// tokenizer_test.go — T-32: focused event-level laws of the hand-rolled
// CPython html.parser port (the golden battery pins the observable
// Extracted fields; these tests pin the tokenizer's own rules so a
// regression names its law).
package webextract

import "testing"

func runParser(html string) *pageParser {
	p := &pageParser{}
	p.parseAll(html)
	return p
}

func TestTokenizerRawModeScansOnlyForItsEndTag(t *testing.T) {
	// a '<' inside script content is data, not markup
	p := runParser("<script>if (a<b) x();</script>ok")
	if p.text() != "ok" {
		t.Errorf("text = %q, want ok", p.text())
	}
	// </scriptfoo> does not close the script (the set_cdata_mode
	// lookahead needs whitespace, '/', or '>' after the name)
	p = runParser("<script>x</scriptfoo>y</script>z")
	if p.text() != "z" {
		t.Errorf("text = %q, want z", p.text())
	}
	// whitespace before '>' closes it
	p = runParser("<script>x</script >z")
	if p.text() != "z" {
		t.Errorf("text = %q, want z", p.text())
	}
	// case-insensitive close
	p = runParser("<script>x</SCRIPT>z")
	if p.text() != "z" {
		t.Errorf("text = %q, want z", p.text())
	}
}

func TestTokenizerSelfClosingScriptDoesNotEnterRawMode(t *testing.T) {
	// <script/> fires start+end events but NOT cdata mode — the text
	// after it parses normally (entities decoded).
	p := runParser("<script/>after &amp; this")
	if p.text() != "after & this" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerRcdataTitleUnescapesButKeepsMarkup(t *testing.T) {
	p := runParser("<title>Fish &amp; Chips</title>")
	if p.title != "Fish & Chips" {
		t.Errorf("title = %q", p.title)
	}
	// tags inside the title are DATA (RCDATA), not events
	p = runParser("<title>The <b>Bold</b> Title</title>")
	if p.title != "The <b>Bold</b> Title" {
		t.Errorf("title = %q", p.title)
	}
	// unterminated title: the close()-time flush still lands in title
	p = runParser("<title>never closed")
	if p.title != "never closed" {
		t.Errorf("title = %q", p.title)
	}
	// a second title element keeps accumulating (the reference's law)
	p = runParser("<title>First</title><title>Second</title>")
	if p.title != "FirstSecond" {
		t.Errorf("title = %q, want FirstSecond", p.title)
	}
}

func TestTokenizerXmpAndIframeAreRawButVisible(t *testing.T) {
	// xmp/iframe are raw-mode (no charref conversion) but NOT in the
	// reference's invisible set — their text is collected verbatim.
	p := runParser("<xmp>a &amp; <b>raw</b></xmp>")
	if p.text() != "a &amp; <b>raw</b>" {
		t.Errorf("text = %q", p.text())
	}
	p = runParser("<iframe>if &amp; raw</iframe>")
	if p.text() != "if &amp; raw" {
		t.Errorf("text = %q", p.text())
	}
	// textarea is RCDATA: unescaped, visible
	p = runParser("<textarea>a &amp; b</textarea>")
	if p.text() != "a & b" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerIncompleteConstructsDiscardTheRest(t *testing.T) {
	// an unterminated start tag at EOF discards everything from it on
	p := runParser("keep<div class=\"x")
	if p.text() != "keep" {
		t.Errorf("text = %q, want keep", p.text())
	}
	// an unterminated quoted attribute swallows to the closing quote…
	p = runParser(`keep<div a="x</p>later" >visible</div>`)
	if p.text() != "keep\nvisible" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerLoneAngleBracketIsData(t *testing.T) {
	p := runParser("a < b")
	if p.text() != "a < b" {
		t.Errorf("text = %q", p.text())
	}
	p = runParser("5 <3 hearts")
	if p.text() != "5 <3 hearts" {
		t.Errorf("text = %q", p.text())
	}
	p = runParser("ends with <")
	if p.text() != "ends with <" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerSilentAndBogusEndTags(t *testing.T) {
	// '</>' is silently consumed; '</ x>' and '</3>' are bogus comments
	for _, doc := range []string{"a</>b", "a</ x>b", "a</3>b"} {
		p := runParser(doc)
		if p.text() != "ab" {
			t.Errorf("%q: text = %q, want ab", doc, p.text())
		}
	}
	// a stray end tag of a BLOCK element still emits its newline
	p := runParser("a</div>b")
	if p.text() != "a\nb" {
		t.Errorf("text = %q, want a\\nb", p.text())
	}
}

func TestTokenizerStartEndTagFiresBothEvents(t *testing.T) {
	// <br/> fires start+end → two newlines, collapsing to one
	p := runParser("a<br/>b")
	if p.text() != "a\nb" {
		t.Errorf("text = %q", p.text())
	}
	// <br / > — the '/' is swallowed by tagfind's separators: a plain
	// start tag (one newline)
	p = runParser("a<br / >b")
	if p.text() != "a\nb" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerAttributeNamesMayContainQuotes(t *testing.T) {
	// `="2"` parses as an attribute NAME (locatetagend's lockstep) —
	// the tag is complete, the text after it survives.
	p := runParser(`keep<div a="1" ="2">tail`)
	if p.text() != "keep\ntail" {
		t.Errorf("text = %q", p.text())
	}
	// quoted attribute values are unescaped after quote-stripping
	p = runParser(`<meta name="description" content="&quot;q&quot;">`)
	if p.metaDescription != `"q"` {
		t.Errorf("meta = %q", p.metaDescription)
	}
	// unquoted attribute values are unescaped too
	p = runParser(`<meta name=description content=a&amp;b>`)
	if p.metaDescription != "a&b" {
		t.Errorf("meta = %q", p.metaDescription)
	}
}

func TestTokenizerMetaPrecedenceLaws(t *testing.T) {
	// first description wins (og before name, or name before og —
	// whichever appears first in the document)
	p := runParser(`<meta property="og:description" content="og">` +
		`<meta name="description" content="name">`)
	if p.metaDescription != "og" {
		t.Errorf("meta = %q, want og (first wins)", p.metaDescription)
	}
	p = runParser(`<meta name="description" content="name">` +
		`<meta property="og:description" content="og">`)
	if p.metaDescription != "name" {
		t.Errorf("meta = %q, want name (first wins)", p.metaDescription)
	}
	// duplicate attribute names: the LAST value wins (dict(attrs))
	p = runParser(`<meta name="description" content="first" content="second">`)
	if p.metaDescription != "second" {
		t.Errorf("meta = %q, want second (last wins)", p.metaDescription)
	}
	// the elif chain: a tag with name=description AND a charset does
	// NOT set the charset (the description branch consumed it)
	p = runParser(`<meta name="description" content="d" charset="utf-8">`)
	if p.metaDescription != "d" || p.detectedCharset != "" {
		t.Errorf("meta = %q charset = %q", p.metaDescription, p.detectedCharset)
	}
	// empty content falls through the chain to the charset branch
	p = runParser(`<meta name="description" content="" charset="utf-8">`)
	if p.detectedCharset != "utf-8" {
		t.Errorf("charset = %q (the elif chain must fall through)", p.detectedCharset)
	}
}

func TestTokenizerInvisibleDataBlockTagNewlines(t *testing.T) {
	// invisible subtrees drop their DATA but block tags inside still
	// emit newlines (only handle_data is depth-gated)
	p := runParser("<svg><div>hidden</div></svg>shown")
	if p.text() != "shown" {
		t.Errorf("text = %q, want shown", p.text())
	}
	// head is invisible but its title/meta are still collected
	p = runParser("<head><title>T</title>" +
		`<meta name="description" content="d"></head><body>ok</body>`)
	if p.title != "T" || p.metaDescription != "d" || p.text() != "ok" {
		t.Errorf("title=%q meta=%q text=%q", p.title, p.metaDescription, p.text())
	}
	// stray text directly in head is dropped
	p = runParser("<head>stray</head>ok")
	if p.text() != "ok" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerCommentAbruptCloseForms(t *testing.T) {
	// <!--> and <!---> close abruptly; content is dropped either way
	for _, doc := range []string{"a<!-->b", "a<!--->b", "a<!-- x -->b"} {
		p := runParser(doc)
		if p.text() != "ab" {
			t.Errorf("%q: text = %q, want ab", doc, p.text())
		}
	}
	// an unterminated comment at EOF is dropped
	p := runParser("a<!-- never closed")
	if p.text() != "a" {
		t.Errorf("text = %q", p.text())
	}
}

func TestTokenizerWhitespaceOnlyDataDropped(t *testing.T) {
	// data that strips to nothing against '\n\t ' is dropped — but
	// '\r' survives the check (not in the reference's strip set)
	p := runParser("<p>a\r\nb</p>")
	if p.text() != "a\r\nb" {
		t.Errorf("text = %q (\\r must survive)", p.text())
	}
	// a lone \r between two block newlines is still eaten by the
	// blank-line collapse (\n\s*\n+ matches "\n\r\n") — the
	// reference returns '' here (verified at the pinned rev).
	p = runParser("<p>\r</p>")
	if p.text() != "" {
		t.Errorf("text = %q, want empty (the collapse eats \\n\\r\\n)", p.text())
	}
	// nbsp is kept in the chunks but stripped by the final text strip
	p = runParser("<p>&nbsp;</p>")
	if p.text() != "" {
		t.Errorf("text = %q", p.text())
	}
}

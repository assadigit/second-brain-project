package webextract

import (
	_ "embed"
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"sync"
)

// entities.go — the port of Python's html.unescape, the exact function
// CPython 3.12's html.parser calls on data runs (convert_charrefs=True)
// and attribute values. Go's stdlib html.UnescapeString diverges from
// Python on eight observable shapes (numeric references to invalid
// codepoints — Python DROPS them to '', Go keeps the rune; '&#x;'
// without digits stays literal in Python and becomes U+FFFD in Go);
// the golden battery (webextract_pure.golden.json → unescape_battery)
// pins every one of them, so the port carries Python's algorithm
// itself, table and all.
//
// The named-reference table is GENERATED from the reference runtime's
// own html.entities.html5 (testdiff/goldens/dump_webextract.py) and
// embedded here — never hand-typed, pinned at sha256 level in the
// golden.

//go:embed entities_html5.json
var entitiesJSON []byte

var (
	entitiesOnce sync.Once
	entitiesMap  map[string]string
)

func entities() map[string]string {
	entitiesOnce.Do(func() {
		m := make(map[string]string, 2231)
		if err := json.Unmarshal(entitiesJSON, &m); err != nil {
			panic("webextract: embedded entities table is corrupt: " + err.Error())
		}
		entitiesMap = m
	})
	return entitiesMap
}

// pyInvalidCharrefs ports html/__init__.py's _invalid_charrefs: the
// numeric references that map specially (0x00 → U+FFFD, 0x0D → CR,
// 0x80-0x9F → the windows-1252 table per the HTML5 spec).
var pyInvalidCharrefs = map[uint64]string{
	0x00: "\uFFFD",
	0x0d: "\r",
	0x80: "€", 0x81: "\u0081", 0x82: "‚", 0x83: "ƒ",
	0x84: "„", 0x85: "…", 0x86: "†", 0x87: "‡",
	0x88: "ˆ", 0x89: "‰", 0x8a: "Š", 0x8b: "‹",
	0x8c: "Œ", 0x8d: "\u008d", 0x8e: "Ž", 0x8f: "\u008f",
	0x90: "\u0090", 0x91: "‘", 0x92: "’", 0x93: "“",
	0x94: "”", 0x95: "•", 0x96: "–", 0x97: "—",
	0x98: "˜", 0x99: "™", 0x9a: "š", 0x9b: "›",
	0x9c: "œ", 0x9d: "\u009d", 0x9e: "ž", 0x9f: "Ÿ",
}

// pyInvalidCodepoint ports _invalid_codepoints: control characters and
// noncharacters that Python's unescape DROPS to ”.
func pyInvalidCodepoint(num uint64) bool {
	if num >= 0x01 && num <= 0x08 {
		return true
	}
	if num == 0x0b {
		return true
	}
	if num >= 0x0e && num <= 0x1f {
		return true
	}
	if num >= 0x7f && num <= 0x9f {
		return true
	}
	if num >= 0xfdd0 && num <= 0xfdef {
		return true
	}
	if num <= 0x10ffff {
		lo := num & 0xffff
		if lo == 0xfffe || lo == 0xffff {
			return true
		}
	}
	return false
}

// charrefRe ports _charref:
// &(#[0-9]+;?|#[xX][0-9a-fA-F]+;?|[^\t\n\f <&#;]{1,32};?)
var charrefRe = regexp.MustCompile(
	`&(#[0-9]+;?|#[xX][0-9a-fA-F]+;?|[^\t\n\f <&#;]{1,32};?)`)

// pyUnescape is html.unescape(s).
func pyUnescape(s string) string {
	if !strings.ContainsRune(s, '&') {
		return s
	}
	return charrefRe.ReplaceAllStringFunc(s, func(m string) string {
		// m is "&" + group(1).
		return pyReplaceCharref(m[1:])
	})
}

// pyReplaceCharref ports _replace_charref (s = the matched reference
// content, semicolon included when present).
func pyReplaceCharref(s string) string {
	if s[0] == '#' {
		var num uint64
		var err error
		if s[1] == 'x' || s[1] == 'X' {
			num, err = strconv.ParseUint(strings.TrimRight(s[2:], ";"), 16, 64)
		} else {
			num, err = strconv.ParseUint(strings.TrimRight(s[1:], ";"), 10, 64)
		}
		if err != nil {
			// Python's int is unbounded; a Go parse overflow can only
			// mean a number beyond every limit → the surrogate/oversize
			// verdict below.
			num = 1 << 40
		}
		if v, ok := pyInvalidCharrefs[num]; ok {
			return v
		}
		if (num >= 0xD800 && num <= 0xDFFF) || num > 0x10FFFF {
			return "\uFFFD"
		}
		if pyInvalidCodepoint(num) {
			return ""
		}
		return string(rune(num))
	}
	// Named reference: exact match first (the table holds both the
	// 'name;' and legacy 'name' forms)…
	if v, ok := entities()[s]; ok {
		return v
	}
	// …then the longest matching prefix (≥2 runes), standard order.
	rs := []rune(s)
	for x := len(rs) - 1; x >= 2; x-- {
		if v, ok := entities()[string(rs[:x])]; ok {
			return v + string(rs[x:])
		}
	}
	return "&" + s
}

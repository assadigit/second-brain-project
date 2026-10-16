package webextract

import (
	"strings"
	"unicode"
)

// pystr.go — the Python string-law helpers. The reference's observable
// behavior leans on Python's str methods in a dozen places (strip,
// split, lower, regex \s); Go's equivalents differ at the edges
// (\x1c-\x1f are Python whitespace but not Go's; U+0130 lowercases to
// two runes in Python and one in Go). These helpers carry the exact
// Python semantics so the port's outputs stay byte-identical.

// pySInner is the inside of the character class equivalent to
// Python's \s (CPython: str.isspace() — ASCII whitespace + \x1c-\x1f
// + \x85 + \xa0 + the Unicode space separators).
const pySInner = `\t\n\v\f\r \x{1c}-\x{1f}\x{85}\x{a0}\x{1680}\x{2000}-\x{200a}\x{2028}\x{2029}\x{202f}\x{205f}\x{3000}`

// pySClass is the full class (with brackets).
const pySClass = "[" + pySInner + "]"

// pyIsSpace reports whether r is Python-whitespace (str.isspace()).
func pyIsSpace(r rune) bool {
	switch r {
	case ' ', '\t', '\n', '\v', '\f', '\r',
		0x1c, 0x1d, 0x1e, 0x1f, 0x85, 0xa0:
		return true
	}
	switch r {
	case 0x1680, 0x2028, 0x2029, 0x202f, 0x205f, 0x3000:
		return true
	}
	return r >= 0x2000 && r <= 0x200a
}

// pyStrip is str.strip() with no arguments.
func pyStrip(s string) string {
	return strings.TrimFunc(s, pyIsSpace)
}

// pySplitWords is str.split() with no arguments: split on runs of
// Python-whitespace, no empty fields.
func pySplitWords(s string) []string {
	var out []string
	i := 0
	rs := []rune(s)
	n := len(rs)
	for i < n {
		for i < n && pyIsSpace(rs[i]) {
			i++
		}
		if i >= n {
			break
		}
		j := i
		for j < n && !pyIsSpace(rs[j]) {
			j++
		}
		out = append(out, string(rs[i:j]))
		i = j
	}
	return out
}

// pyLower is str.lower(): rune-wise with the one multi-rune mapping
// (U+0130 İ → "i\u0307" — Python produces two runes, Go's unicode
// tables produce one).
func pyLower(s string) string {
	if !strings.Contains(s, "İ") {
		return strings.ToLower(s)
	}
	var b strings.Builder
	for _, r := range s {
		if r == 0x0130 {
			b.WriteString("i\u0307")
			continue
		}
		b.WriteRune(unicode.ToLower(r))
	}
	return b.String()
}

// asciiLower lowercases ASCII letters only. Python's tag/attribute
// name lowercasing (str.lower on names) is pure-ASCII-equivalent for
// every set-membership decision in this package: no member of the
// reference's tag sets contains a non-ASCII rune, and no non-ASCII
// rune can lower into a pure-ASCII member (the only Unicode lower
// mapping INTO an ASCII letter is the Kelvin sign U+212A → 'k', and
// no set member contains 'k'). Documented in doc.go.
func asciiLower(s string) string {
	b := []byte(s)
	for i, c := range b {
		if c >= 'A' && c <= 'Z' {
			b[i] = c + ('a' - 'A')
		}
	}
	return string(b)
}

// pyRunesSlice is Python string slicing s[:n] (by codepoint).
func pyRunesSlice(s string, n int) string {
	if n <= 0 {
		return ""
	}
	i := 0
	for _, r := range s {
		if i == n {
			return s[:i]
		}
		i += len(string(r))
	}
	return s
}

// pyWordChar reports whether r is a Python regex \w character
// ([a-zA-Z0-9_] plus Unicode letters/digits — str.isalnum() or '_').
func pyWordChar(r rune) bool {
	return r == '_' || unicode.IsLetter(r) || unicode.IsDigit(r)
}

// pyHasScriptTag is re.search(r'<script\b', html, re.IGNORECASE):
// a case-insensitive "<script" whose next character (if any) is NOT a
// Python word character (Python's \b is Unicode-aware; Go's regexp \b
// is ASCII-only — a literal '<scripté' must NOT count).
func pyHasScriptTag(html string) bool {
	lower := asciiLower(html)
	off := 0
	for {
		i := strings.Index(lower[off:], "<script")
		if i < 0 {
			return false
		}
		i += off
		rest := html[i+len("<script"):]
		if rest == "" {
			return true
		}
		rs := []rune(rest)
		if !pyWordChar(rs[0]) {
			return true
		}
		off = i + len("<script")
	}
}

// pyWordCharsCount is len(re.findall(r'[A-Za-z0-9]', text)).
func pyWordCharsCount(text string) int {
	n := 0
	for i := 0; i < len(text); i++ {
		c := text[i]
		if (c >= 'a' && c <= 'z') || (c >= 'A' && c <= 'Z') ||
			(c >= '0' && c <= '9') {
			n++
		}
	}
	return n
}

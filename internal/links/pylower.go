// pylower.go — Python str.lower() semantics for links (doc 05 §3 item 2).
//
// Python's str.lower() differs from Go's strings.ToLower in exactly two
// ways, both pinned by the generated golden (link_grammar.golden.json,
// section "py_lower"):
//  1. U+0130 (İ) lowercases to "i" + U+0307 (two runes) — the FULL
//     mapping; Go's unicode.ToLower yields just "i".
//  2. Final sigma: Greek capital sigma (Σ) lowercases to ς (U+03C2) when
//     it ends a "word" (preceded by a cased letter with only
//     case-ignorable characters between, and not followed by a cased
//     letter); otherwise σ.
//
// Everything else is the simple 1:1 mapping, where Go's unicode.ToLower
// already matches Python.
package links

import (
	"strings"
	"unicode"
)

// pyLower mirrors Python's str.lower(). Unexported for now; T-17 decides
// whether the helper library lifts it (doc 05 §3 item 2: "specialLower
// helper if needed" — the links golden needed it).
func pyLower(s string) string {
	if !strings.ContainsFunc(s, func(r rune) bool {
		return r == 0x130 || r == 0x3A3 || r == 0x3C2 || unicode.ToLower(r) != r
	}) {
		return s
	}
	var b strings.Builder
	b.Grow(len(s))
	rs := []rune(s)
	for i, r := range rs {
		switch {
		case r == 0x130: // İ → i + combining dot above
			b.WriteRune('i')
			b.WriteRune(0x307)
		case r == 0x3A3 && finalSigma(rs, i):
			b.WriteRune(0x3C2) // ς
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// finalSigma reports whether the capital sigma at rs[i] is word-final
// (SpecialCasing Final_Sigma): preceded by a sequence consisting of a
// cased letter then zero or more case-ignorable characters, and not
// followed by a sequence of zero or more case-ignorable characters then
// a cased letter.
func finalSigma(rs []rune, i int) bool {
	// look back: nearest non-case-ignorable char must be cased
	j := i - 1
	for j >= 0 && isCaseIgnorable(rs[j]) {
		j--
	}
	if j < 0 || !isCased(rs[j]) {
		return false
	}
	// look forward: nearest non-case-ignorable char must NOT be cased
	k := i + 1
	for k < len(rs) && isCaseIgnorable(rs[k]) {
		k++
	}
	if k < len(rs) && isCased(rs[k]) {
		return false
	}
	return true
}

// isCased mirrors the Unicode Cased property for the realistic subset:
// Ll, Lu, Lt categories (Other_Lowercase/Other_Uppercase never appear in
// the hosts/URL text this helper processes).
func isCased(r rune) bool {
	return unicode.IsLower(r) || unicode.IsUpper(r) || unicode.IsTitle(r)
}

// isCaseIgnorable mirrors Case_Ignorable: the Word_Break MidLetter /
// MidNumLet / Single_Quote characters plus the Mn, Me, Cf, Lm, Sk
// categories (the UAX#31 derivation).
func isCaseIgnorable(r rune) bool {
	if unicode.Is(unicode.Mn, r) || unicode.Is(unicode.Me, r) || unicode.Is(unicode.Cf, r) || unicode.Is(unicode.Lm, r) || unicode.Is(unicode.Sk, r) {
		return true
	}
	switch r {
	case '\u0027', '\u002E', '\u003A', '\u003B', '\u005F', '\u00B7',
		'\u2018', '\u2019', '\u2024', '\uFE52', '\uFF07', '\uFF0E':
		return true
	}
	return false
}

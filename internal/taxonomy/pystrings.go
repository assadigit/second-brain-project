// pystrings.go — Python string-semantics helpers for the taxonomy parser.
// Each is pinned byte-exact against the generated golden corpus
// (taxonomy_corpus.golden.json: splitlines / stripws / strip_chars).
// T-17's helpers package will centralize them (same promotion law as
// dryrun's formatComma and prompts' PyRepr).
package taxonomy

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

// pySplitLines mirrors str.splitlines(): split on \n, \r, \r\n, \v (\x0b),
// \f (\x0c), \x1c-\x1e, \x85, U+2028, U+2029. A trailing boundary yields
// no trailing empty line; consecutive boundaries yield empty lines; the
// empty string yields no lines at all.
func pySplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); {
		r, size := utf8.DecodeRuneInString(s[i:])
		switch r {
		case '\n', '\v', '\f', '\x1c', '\x1d', '\x1e', '\x85', '\u2028', '\u2029':
			lines = append(lines, s[start:i])
			i += size
			start = i
		case '\r':
			lines = append(lines, s[start:i])
			i += size
			if i < len(s) && s[i] == '\n' { // \r\n is ONE boundary
				i++
			}
			start = i
		default:
			i += size
		}
	}
	if start < len(s) {
		lines = append(lines, s[start:])
	}
	return lines
}

// pyIsSpace mirrors Python str.isspace()'s membership: Go's
// unicode.IsSpace (ASCII whitespace + Unicode White_Space) PLUS the
// \x1c-\x1f group separators Python also counts.
func pyIsSpace(r rune) bool {
	switch r {
	case '\x1c', '\x1d', '\x1e', '\x1f':
		return true
	}
	return unicode.IsSpace(r)
}

// pyStrip mirrors str.strip(); pyLstrip/pyRstrip the one-sided forms.
func pyStrip(s string) string  { return strings.TrimFunc(s, pyIsSpace) }
func pyLstrip(s string) string { return strings.TrimLeftFunc(s, pyIsSpace) }
func pyRstrip(s string) string { return strings.TrimRightFunc(s, pyIsSpace) }

// stripBOM removes a leading UTF-8 BOM (utf-8-sig decode).
func stripBOM(raw []byte) []byte {
	if len(raw) >= 3 && raw[0] == 0xEF && raw[1] == 0xBB && raw[2] == 0xBF {
		return raw[3:]
	}
	return raw
}

// pyUTF8DecodeError reproduces Python's UnicodeDecodeError message for
// the first invalid byte in a UTF-8 document: "'utf-8' codec can't decode
// byte 0xXX in position N: <reason>" with the reason being "invalid start
// byte" (the byte cannot lead a sequence), "invalid continuation byte"
// (a valid lead byte is followed by a non-continuation byte), or
// "unexpected end of data" (a multi-byte sequence is truncated by the end
// of input). Returned as a PLAIN error — Python raises it OUTSIDE the
// OSError/TaxonomyError contract and the port preserves that shape.
func pyUTF8DecodeError(raw []byte) error {
	for i := 0; i < len(raw); {
		b := raw[i]
		if b < utf8.RuneSelf {
			i++
			continue
		}
		var size int
		switch {
		case b >= 0xC2 && b <= 0xDF:
			size = 2
		case b >= 0xE0 && b <= 0xEF:
			size = 3
		case b >= 0xF0 && b <= 0xF4:
			size = 4
		default:
			// cannot lead a valid sequence
			return fmt.Errorf("'utf-8' codec can't decode byte 0x%02x in position %d: invalid start byte", b, i)
		}
		if i+size > len(raw) {
			return fmt.Errorf("'utf-8' codec can't decode byte 0x%02x in position %d: unexpected end of data", b, i)
		}
		for j := 1; j < size; j++ {
			c := raw[i+j]
			if c < 0x80 || c > 0xBF {
				return fmt.Errorf("'utf-8' codec can't decode byte 0x%02x in position %d: invalid continuation byte", c, i+j)
			}
		}
		i += size
	}
	return fmt.Errorf("'utf-8' codec can't decode: unreachable (validated)")
}

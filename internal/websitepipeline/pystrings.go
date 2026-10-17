package websitepipeline

import "os"

// pystrings.go — the Python string-shape helpers shared by this
// package (splitlines' full boundary set; the bounded head read in
// decoded characters).

// pySplitLines mirrors Python str.splitlines(): splits on \n, \r\n
// (one boundary), \r, \v, \f, \x1c-\x1e, \x85, \u2028, \u2029; keeps
// empty interior lines; no trailing empty line.
func pySplitLines(s string) []string {
	var out []string
	start := 0
	i := 0
	for i < len(s) {
		r, size := rune(s[i]), 1
		if s[i] >= 0x80 {
			for _, rr := range s[i:] {
				r, size = rr, len(string(rr))
				break
			}
		}
		switch {
		case r == '\r' && i+1 < len(s) && s[i+1] == '\n':
			out = append(out, s[start:i])
			i += 2
			start = i
		case r == '\n' || r == '\r' || r == '\v' || r == '\f' ||
			r == 0x1c || r == 0x1d || r == 0x1e || r == 0x85 ||
			r == 0x2028 || r == 0x2029:
			out = append(out, s[start:i])
			i += size
			start = i
		default:
			i += size
		}
	}
	if start < len(s) {
		out = append(out, s[start:])
	}
	return out
}

// readHeadRunes mirrors open(path, encoding='utf-8',
// errors='replace') followed by f.read(n) — the first n DECODED
// characters (Python's text-mode read counts characters, not bytes).
func readHeadRunes(path string, n int) (string, bool) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", false
	}
	s := decodeUTF8Replace(string(data))
	runes := []rune(s)
	if len(runes) > n {
		return string(runes[:n]), true
	}
	return s, true
}

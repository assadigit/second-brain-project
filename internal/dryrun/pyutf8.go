package dryrun

// pyUTF8Replace mirrors Python bytes.decode('utf-8', errors='replace'):
// invalid sequences are replaced by U+FFFD per MAXIMAL SUBPART — each
// maximal invalid subpart (Unicode 6 / PEP 383 semantics, CPython's
// decoder) yields exactly one replacement character, and decoding resumes
// at the byte that broke the sequence.
//
// This is load-bearing parity: previews of binary content (e.g. PNG
// banners) and of content cut mid-rune at the 160-byte boundary both hit
// this path, and the golden pins the exact output (the cjk-bytes and
// banner battery entries).
//
// Divergence note: Go's strings.ToValidUTF8 replaces each RUN of invalid
// bytes with a single replacement — b"\x80\x80" would give one U+FFFD
// where Python gives two. Hence this hand-rolled decoder.
func pyUTF8Replace(b []byte) string {
	var out []byte
	for i := 0; i < len(b); {
		c := b[i]
		if c < 0x80 {
			out = append(out, c)
			i++
			continue
		}
		var n int       // expected total sequence length
		var lo, hi byte // allowed range of the FIRST continuation byte
		switch {
		case c >= 0xC2 && c <= 0xDF:
			n, lo, hi = 2, 0x80, 0xBF
		case c == 0xE0:
			n, lo, hi = 3, 0xA0, 0xBF // excludes overlong encodings
		case c >= 0xE1 && c <= 0xEC:
			n, lo, hi = 3, 0x80, 0xBF
		case c == 0xED:
			n, lo, hi = 3, 0x80, 0x9F // excludes surrogates
		case c >= 0xEE && c <= 0xEF:
			n, lo, hi = 3, 0x80, 0xBF
		case c == 0xF0:
			n, lo, hi = 4, 0x90, 0xBF // excludes overlong encodings
		case c >= 0xF1 && c <= 0xF3:
			n, lo, hi = 4, 0x80, 0xBF
		case c == 0xF4:
			n, lo, hi = 4, 0x80, 0x8F // excludes > U+10FFFF
		default:
			// invalid start byte: 0x80–0xC1 (stray continuation / overlong
			// 2-byte prefix) and 0xF5–0xFF — one U+FFFD each.
			out = append(out, 0xEF, 0xBF, 0xBD)
			i++
			continue
		}
		// consume the maximal valid prefix of the sequence
		k := i + 1
		for ; k < i+n && k < len(b); k++ {
			l, h := byte(0x80), byte(0xBF)
			if k == i+1 {
				l, h = lo, hi
			}
			if b[k] < l || b[k] > h {
				break
			}
		}
		if k == i+n {
			out = append(out, b[i:i+n]...) // valid sequence, verbatim
			i = k
		} else {
			// maximal subpart b[i:k] → ONE U+FFFD; resume at the breaking
			// byte (or end of input if the sequence was truncated).
			out = append(out, 0xEF, 0xBF, 0xBD)
			i = k
		}
	}
	return string(out)
}

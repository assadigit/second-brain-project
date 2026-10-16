// Package helpers — the Python-semantics helper library (doc 05 §3
// items 2/3/4/6/7/12/13/16): the rune/rounding/ordering/strip helpers the
// engine packages need, plus the session's PROMOTED string helpers
// (formatComma from dryrun, pyLower from links, PyStr/PyRepr/
// PyOSErrorText from prompts, splitlines/strip from taxonomy — one
// canonical home, no duplication; the old sites keep thin delegations or
// re-pointed call sites).
//
// Every helper is pinned by the generated golden in testdata/ (see
// helpers_test.go): the corpus runs the pinned reference interpreter's
// own builtins (str[:n], str.rstrip(chars), round(x, n),
// dict.fromkeys, os.path.normcase, f"{n:,}", str.lower, str()).
package helpers

import (
	"errors"
	"fmt"
	"io/fs"
	"math"
	"math/big"
	"runtime"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"

	"encoding/json"
)

// ---------------------------------------------------------------------------
// Code-point slicing (doc 05 §3 item 3).
// ---------------------------------------------------------------------------

// CutRunes returns the first n CODE POINTS of s (Python s[:n]) — byte
// slicing would cut mid-rune (🗑️+VS16 is two code points; U+0130 one;
// CJK one each).
func CutRunes(s string, n int) string {
	if n <= 0 {
		return ""
	}
	count := 0
	for i := range s {
		if count == n {
			return s[:i]
		}
		count++
	}
	return s
}

// ---------------------------------------------------------------------------
// Character-set strip (doc 05 §3 items 6/7).
// ---------------------------------------------------------------------------

// RstripChars mirrors s.rstrip(chars): strips any combination of the
// SET's characters from the end — a character-SET, not a suffix. An
// empty set strips nothing.
func RstripChars(s, chars string) string {
	return strings.TrimRight(s, chars)
}

// StripChars mirrors s.strip(chars): both ends.
func StripChars(s, chars string) string {
	return strings.Trim(s, chars)
}

// LstripChars mirrors s.lstrip(chars): the start only.
func LstripChars(s, chars string) string {
	return strings.TrimLeft(s, chars)
}

// ---------------------------------------------------------------------------
// round(x, n) — banker's rounding (doc 05 §3 item 4).
// ---------------------------------------------------------------------------

// RoundHalfEven mirrors Python round(x, n) (n ≥ 0; negative n rounds to
// tens/hundreds/…): round-half-to-even evaluated on the EXACT decimal
// value of the binary double (big.Rat), then the result converted back
// to the nearest double. This is why 2.675 → 2.67 (the double is
// 2.674999…) while 0.125 → 0.12 (an exact half falls to the even side),
// and why Go's math.Round (half-away-from-zero) would diverge. A rounded
// value of zero keeps x's sign (Python: round(-0.5, 0) = -0.0).
func RoundHalfEven(x float64, n int) float64 {
	if math.IsNaN(x) || math.IsInf(x, 0) {
		return x
	}
	r := new(big.Rat).SetFloat64(x)
	if r == nil {
		return x
	}
	scale := new(big.Int).Exp(big.NewInt(10), big.NewInt(int64(absInt(n))), nil)
	scaleRat := new(big.Rat).SetInt(scale)
	scaled := new(big.Rat).Set(r)
	if n >= 0 {
		scaled.Mul(scaled, scaleRat)
	} else {
		scaled.Quo(scaled, scaleRat)
	}
	k := roundRatHalfEven(scaled)
	out := new(big.Rat).SetInt(k)
	if n >= 0 {
		out.Quo(out, scaleRat)
	} else {
		out.Mul(out, scaleRat)
	}
	f, _ := out.Float64()
	if f == 0 && (x < 0 || math.Signbit(x)) {
		return math.Copysign(0, -1)
	}
	return f
}

// roundRatHalfEven rounds a rational to the nearest integer, ties to
// even (floor-based, sign-correct).
func roundRatHalfEven(r *big.Rat) *big.Int {
	num, den := r.Num(), r.Denom()
	// floor division (big.Int.Quo truncates toward zero) + the FLOOR
	// remainder recomputed against the adjusted quotient (rem/den ∈ [0,1))
	q, _ := new(big.Int).QuoRem(num, den, new(big.Int))
	rem := new(big.Int).Sub(num, new(big.Int).Mul(q, den))
	if rem.Sign() < 0 {
		q.Sub(q, big.NewInt(1))
		rem = new(big.Int).Sub(num, new(big.Int).Mul(q, den))
	}
	twiceRem := new(big.Int).Abs(rem)
	twiceRem.Lsh(twiceRem, 1)
	cmp := twiceRem.Cmp(den)
	switch {
	case cmp < 0:
		return q
	case cmp > 0:
		return q.Add(q, big.NewInt(1))
	default: // exact half → even
		if q.Bit(0) == 0 {
			return q
		}
		return q.Add(q, big.NewInt(1))
	}
}

func absInt(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

// ---------------------------------------------------------------------------
// Ordered dedupe (doc 05 §3 item 16).
// ---------------------------------------------------------------------------

// OrderedDedupe mirrors dict.fromkeys(items): the first occurrence wins,
// order preserved (U+0130 ≠ i; ints and strings are comparables).
func OrderedDedupe[T comparable](items []T) []T {
	seen := make(map[T]struct{}, len(items))
	out := make([]T, 0, len(items))
	for _, it := range items {
		if _, dup := seen[it]; dup {
			continue
		}
		seen[it] = struct{}{}
		out = append(out, it)
	}
	return out
}

// ---------------------------------------------------------------------------
// os.path.normcase (doc 05 §3 item 13).
// ---------------------------------------------------------------------------

// Normcase mirrors os.path.normcase for the HOST platform: the identity
// on POSIX; slash-fold + FULL Unicode lower on Windows (per-platform
// parity, exactly like the reference).
func Normcase(s string) string {
	if runtimeIsWindows {
		return NormcaseNT(s)
	}
	return s
}

// NormcaseNT pins the Windows column (the production platform's
// behavior — note_state's moved-test comparison needs it regardless of
// the host): '/' → '\\', then str.lower with its FULL mappings (İ → i̇,
// final sigma).
func NormcaseNT(s string) string {
	return PyLower(strings.ReplaceAll(s, "/", "\\"))
}

// ---------------------------------------------------------------------------
// f"{n:,}" — the promoted dryrun.formatComma (doc 05 §3 item 12).
// ---------------------------------------------------------------------------

// FormatComma mirrors Python's f"{n:,}" — thousands separators in
// groups of three, negatives signed. Promoted from internal/dryrun
// (T-17's task-list note: promote, never duplicate).
func FormatComma(n int64) string {
	s := strconv.FormatInt(n, 10)
	neg := strings.HasPrefix(s, "-")
	if neg {
		s = s[1:]
	}
	if len(s) <= 3 {
		if neg {
			return "-" + s
		}
		return s
	}
	var parts []string
	for len(s) > 3 {
		parts = append([]string{s[len(s)-3:]}, parts...)
		s = s[:len(s)-3]
	}
	parts = append([]string{s}, parts...)
	out := strings.Join(parts, ",")
	if neg {
		return "-" + out
	}
	return out
}

// ---------------------------------------------------------------------------
// str.lower — the promoted links.pyLower (doc 05 §3 item 2).
// ---------------------------------------------------------------------------

// PyLower mirrors Python's str.lower(): the FULL case mapping — U+0130
// (İ) lowercases to "i" + U+0307, and capital sigma takes the final
// form (ς) when word-final. Everything else is Go's 1:1 unicode.ToLower.
func PyLower(s string) string {
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
	j := i - 1
	for j >= 0 && isCaseIgnorable(rs[j]) {
		j--
	}
	if j < 0 || !isCased(rs[j]) {
		return false
	}
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
// the text this helper processes).
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

// ---------------------------------------------------------------------------
// str() / repr() / OSError text — the promoted prompts helpers.
// ---------------------------------------------------------------------------

// PyStr mirrors Python str() for the value shapes parity code meets:
// string as-is; bool → "True"/"False"; nil → "None" (Python's own word —
// callers with special None grammar check nil BEFORE calling); ints in
// decimal; floats and JSON number literals via Python float repr
// (str(5) == "5" but str(5.0) == "5.0"; 1e15 positional, 1e-05
// scientific — storage.PyFloatRepr).
func PyStr(v any) string {
	switch vv := v.(type) {
	case nil:
		return "None"
	case string:
		return vv
	case bool:
		if vv {
			return "True"
		}
		return "False"
	case json.Number:
		lit := vv.String()
		if !strings.ContainsAny(lit, ".eE") {
			return lit // integral literal → Python int str()
		}
		f, err := vv.Float64()
		if err != nil {
			return lit
		}
		return PyFloatRepr(f)
	case float64:
		return PyFloatRepr(vv)
	case float32:
		return PyFloatRepr(float64(vv))
	case int:
		return strconv.Itoa(vv)
	case int8:
		return strconv.FormatInt(int64(vv), 10)
	case int16:
		return strconv.FormatInt(int64(vv), 10)
	case int32:
		return strconv.FormatInt(int64(vv), 10)
	case int64:
		return strconv.FormatInt(vv, 10)
	case uint:
		return strconv.FormatUint(uint64(vv), 10)
	case uint8:
		return strconv.FormatUint(uint64(vv), 10)
	case uint16:
		return strconv.FormatUint(uint64(vv), 10)
	case uint32:
		return strconv.FormatUint(uint64(vv), 10)
	case uint64:
		return strconv.FormatUint(vv, 10)
	default:
		return fmt.Sprintf("%v", v)
	}
}

// PyRepr mirrors Python repr() for strings: single quotes preferred,
// switching to double quotes only when the string contains ' and no ";
// \\ , \n, \r, \t escapes; \xNN for non-printable ASCII (including DEL)
// and non-printable Latin-1; \uNNNN / \UNNNNNNNN for non-printable
// runes; printable Unicode literal.
func PyRepr(s string) string {
	quote := byte('\'')
	if strings.ContainsRune(s, '\'') && !strings.ContainsRune(s, '"') {
		quote = '"'
	}
	var b strings.Builder
	b.WriteByte(quote)
	for _, r := range s {
		switch {
		case r == rune(quote):
			b.WriteByte('\\')
			b.WriteByte(quote)
		case r == '\\':
			b.WriteString(`\\`)
		case r == '\n':
			b.WriteString(`\n`)
		case r == '\r':
			b.WriteString(`\r`)
		case r == '\t':
			b.WriteString(`\t`)
		case r < 0x20 || r == 0x7f:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r < 0x7f:
			b.WriteRune(r)
		case unicode.IsPrint(r):
			b.WriteRune(r)
		case r < 0x100:
			fmt.Fprintf(&b, `\x%02x`, r)
		case r <= 0xffff:
			fmt.Fprintf(&b, `\u%04x`, r)
		default:
			fmt.Fprintf(&b, `\U%08x`, r)
		}
	}
	b.WriteByte(quote)
	return b.String()
}

// PyOSErrorText formats a read failure the way str(OSError) does in
// Python — "[Errno N] Message: 'PATH'" — for the common syscall shapes
// both a disk read (os.ReadFile) and an embed-FS lookup produce:
// missing → ENOENT; permission → EACCES; a directory read ("is a
// directory" — os and embed spell it the same) → EISDIR. Anything else
// falls back to the Go error text, still inside the caller's wrapper.
func PyOSErrorText(err error, path string) string {
	var pathErr *fs.PathError
	if errors.As(err, &pathErr) {
		switch {
		case errors.Is(err, fs.ErrNotExist):
			return "[Errno 2] No such file or directory: " + PyRepr(path)
		case errors.Is(err, fs.ErrPermission):
			return "[Errno 13] Permission denied: " + PyRepr(path)
		case strings.Contains(err.Error(), "is a directory"):
			return "[Errno 21] Is a directory: " + PyRepr(path)
		}
	}
	return err.Error()
}

// ---------------------------------------------------------------------------
// str.splitlines() / Unicode whitespace strip — the promoted taxonomy
// helpers.
// ---------------------------------------------------------------------------

// SplitLines mirrors str.splitlines(): split on \n, \r, \r\n, \v (\x0b),
// \f (\x0c), \x1c-\x1e, \x85, U+2028, U+2029. A trailing boundary yields
// no trailing empty line; consecutive boundaries yield empty lines; the
// empty string yields no lines at all.
func SplitLines(s string) []string {
	var lines []string
	start := 0
	for i := 0; i < len(s); {
		r, size := decodeRune(s[i:])
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

// pyIsSpace mirrors Python str.isspace() membership: Go's
// unicode.IsSpace plus the \x1c-\x1f group separators.
func pyIsSpace(r rune) bool {
	switch r {
	case '\x1c', '\x1d', '\x1e', '\x1f':
		return true
	}
	return unicode.IsSpace(r)
}

// Strip mirrors str.strip() with Python's Unicode whitespace set
// (U+00A0, U+2028, \x1c … strip); Lstrip/Rstrip the one-sided forms.
func Strip(s string) string  { return strings.TrimFunc(s, pyIsSpace) }
func Lstrip(s string) string { return strings.TrimLeftFunc(s, pyIsSpace) }
func Rstrip(s string) string { return strings.TrimRightFunc(s, pyIsSpace) }

// CollapseWhitespace mirrors Python re.sub(r'\s+', ' ', text) with its
// UNICODE \s class (promoted from notebuilder): any run of Python
// whitespace becomes one ASCII space.
func CollapseWhitespace(text string) string {
	var b strings.Builder
	inRun := false
	for _, r := range text {
		if pyIsSpace(r) {
			if !inRun {
				b.WriteByte(' ')
				inRun = true
			}
			continue
		}
		inRun = false
		b.WriteRune(r)
	}
	return b.String()
}

// ---------------------------------------------------------------------------
// Shared internals.
// ---------------------------------------------------------------------------

// pyFloatRepr reproduces Python's float repr (CPython float_repr_style
// short): the shortest digit string that round-trips, in positional
// notation while -3 <= decpt <= 16, otherwise scientific with a signed,
// at-least-two-digit exponent. Integral positional values keep ".0".
//
// Verified against the generated corpus golden (storage_corpus.golden.json
// float_repr): 1e15 → "1000000000000000.0", 1e16 → "1e+16",
// 1e-05 → "1e-05", 0.0001 → "0.0001", 0.5 → "0.5", 2.0 → "2.0".
// Go's own 'g' format would print 1e15 as "1e+15" — divergent.
func PyFloatRepr(f float64) string {
	// Python json.dumps spells non-finite floats Infinity/-Infinity/NaN
	// (never in real configs; kept exact for completeness).
	switch {
	case math.IsNaN(f):
		return "NaN"
	case math.IsInf(f, 1):
		return "Infinity"
	case math.IsInf(f, -1):
		return "-Infinity"
	}
	neg := false
	if f < 0 || (f == 0 && math.Signbit(f)) {
		neg = true
		f = -f
	}
	// Shortest round-trip digits + decimal exponent via 'e' format.
	sci := strconv.FormatFloat(f, 'e', -1, 64)
	eIdx := strings.IndexByte(sci, 'e')
	mant := sci[:eIdx]
	exp10, _ := strconv.Atoi(sci[eIdx+1:]) // value = d.ddd × 10^exp10
	digits := strings.Replace(mant, ".", "", 1)
	// decpt: position of the decimal point relative to the digit string
	// (value = 0.digits × 10^decpt).
	decpt := exp10 + 1

	var out string
	if decpt < -3 || decpt > 16 {
		// scientific: d[.ddd]e±XX (exponent = decpt-1, ≥ 2 digits)
		m := digits[:1]
		if len(digits) > 1 {
			m += "." + digits[1:]
		}
		sign := "+"
		e := decpt - 1
		if e < 0 {
			sign, e = "-", -e
		}
		es := strconv.Itoa(e)
		if len(es) < 2 {
			es = "0" + es
		}
		out = m + "e" + sign + es
	} else if decpt <= 0 {
		out = "0." + strings.Repeat("0", -decpt) + digits
	} else if decpt >= len(digits) {
		out = digits + strings.Repeat("0", decpt-len(digits)) + ".0"
	} else {
		out = digits[:decpt] + "." + digits[decpt:]
	}
	if neg {
		return "-" + out
	}
	return out
}

// runtimeIsWindows routes Normcase per-platform.
var runtimeIsWindows = runtime.GOOS == "windows"

// decodeRune is utf8.DecodeRuneInString (named for the call-site shape).
func decodeRune(s string) (rune, int) { return utf8.DecodeRuneInString(s) }

// ---------------------------------------------------------------------------
// UTF-8 replacement decode (Python open(..., errors='replace')).
// ---------------------------------------------------------------------------

// DecodeUTF8Replace mirrors decoding bytes as utf-8 with
// errors='replace' (Python's open(..., encoding='utf-8',
// errors='replace') and bytes.decode('utf-8', 'replace')) with the
// maximal-subpart rule the reference interpreter applies: each maximal
// invalid subpart (the longest prefix that is still a prefix of some
// valid sequence) collapses into ONE U+FFFD. Pinned byte-exact by the
// goldens' invalid-UTF-8 batteries (helpers', notestate's fingerprint
// corpus, notereads' read-bounded corpus):
//
//	b"\xff"→1×U+FFFD; b"\xc2A"→U+FFFD+"A"; b"a\xe2\x82"→"a"+1×U+FFFD
//	(the truncated 3-byte attempt is ONE subpart); b"\xc3\xc3A"→2×
//	U+FFFD+"A" (each C3 is its own failed attempt).
//
// (internal/notestate carries a local twin from T-21 — kept untouched
// this session; a future cleanup may re-point it here, with both
// goldens already pinning the same law.)
func DecodeUTF8Replace(data []byte) string {
	if utf8.Valid(data) {
		return string(data)
	}
	var b strings.Builder
	b.Grow(len(data))
	for len(data) > 0 {
		r, size := utf8.DecodeRune(data)
		if r != utf8.RuneError || size != 1 {
			// Valid sequence — keep the original bytes (a literal
			// U+FFFD in the input stays exactly the bytes it was).
			b.Write(data[:size])
			data = data[size:]
			continue
		}
		b.WriteRune(utf8.RuneError)
		data = data[subpartLen(data):]
	}
	return b.String()
}

// subpartLen returns the length of the maximal invalid subpart starting
// at data[0]: the lead byte plus however many following bytes are legal
// continuations of a still-possibly-valid sequence (lead-specific range
// for the first continuation, 0x80-0xBF for the rest). C0/C1 (overlong
// 2-byte leads), 0x80-0xBF (bare continuations) and 0xF5-0xFF are
// invalid leads on their own.
func subpartLen(data []byte) int {
	lead := data[0]
	var expected int
	switch {
	case lead < 0xC2: // bare continuation, C0, C1
		return 1
	case lead < 0xE0: // C2-DF: one continuation
		expected = 1
	case lead < 0xF0: // E0-EF: two
		expected = 2
	case lead < 0xF5: // F0-F4: three
		expected = 3
	default: // F5-FF
		return 1
	}
	n := 1
	for i := 1; i <= expected && i < len(data); i++ {
		c := data[i]
		if c < 0x80 || c > 0xBF {
			break
		}
		// Lead-specific first-continuation ranges (anti-overlong and
		// anti-surrogate): E0 needs A0-BF, ED needs 80-9F, F0 needs
		// 90-BF, F4 needs 80-8F.
		if i == 1 {
			switch {
			case lead == 0xE0 && c < 0xA0:
				return n
			case lead == 0xED && c > 0x9F:
				return n
			case lead == 0xF0 && c < 0x90:
				return n
			case lead == 0xF4 && c > 0x8F:
				return n
			}
		}
		n++
	}
	return n
}

// PyNewlines mirrors Python text mode's universal-newlines translation
// (open(..., 'r') with newline=None): \r\n and a lone \r both become
// \n; nothing else changes (\x85 NEL is NOT a file newline). Every
// reference read of a note file runs through this law, so the port's
// file reads must translate identically — pinned by the notereads
// read-bounded golden battery and the notestate CRLF fingerprint case.
func PyNewlines(s string) string {
	if !strings.Contains(s, "\r") {
		return s
	}
	s = strings.ReplaceAll(s, "\r\n", "\n")
	return strings.ReplaceAll(s, "\r", "\n")
}

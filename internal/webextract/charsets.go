package webextract

import (
	_ "embed"
	"encoding/base64"
	"encoding/json"
	"regexp"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/text/encoding"
	"golang.org/x/text/encoding/japanese"
	"golang.org/x/text/encoding/korean"
	"golang.org/x/text/encoding/simplifiedchinese"
	"golang.org/x/text/encoding/traditionalchinese"
)

// charsets.go — the port of the reference's `_try` decode ladder
// (body.decode(cs), STRICT — any error means "fall through to the
// sniff/replace path"). Python's codecs registry is the contract:
//
//   - the utf-8 family decodes STRICTLY (a single invalid byte fails,
//     unlike x/text's replacing decoders — the replace law only lives
//     in the final fallback);
//   - the single-byte charsets ride tables GENERATED from the
//     reference runtime's own codecs (testdiff/goldens/
//     dump_webextract.py → charsets_table.json) — x/text's own
//     charmaps diverge there (its ISO-8859-15 leaves the whole C1 row
//     unmapped; its Windows-1254 misses 0x8E/0x9E), so the port ships
//     the reference's tables instead;
//   - cp949 (the UHC-extended Korean codec — x/text has no
//     equivalent) rides its own generated table;
//   - gb2312 is STRICTER than GBK (lead 0xA1-0xF7, trail 0xA1-0xFE —
//     Python's euc-cn codec rejects the GBK extension zone), so the
//     grammar is checked before the GBK table decodes;
//   - the other DBCS charsets (gbk, gb18030, big5, shift_jis, euc-jp,
//     euc-kr) ride x/text's standard decoders behind a strict-reject
//     layer: x/text replaces invalid sequences with U+FFFD, so any
//     U+FFFD in the output means Python's strict decode would have
//     failed → nil (the ladder falls through). Documented edge: a
//     page that legitimately CONTAINS U+FFFD (only encodable in
//     gb18030) would be misjudged — recorded in doc.go;
//   - utf-16/utf-32 are hand-rolled with Python's exact rules (BOM
//     detection for the unspecified-endian names — native LE when
//     absent, the pinned venv and the Windows target are both LE —
//     odd lengths and unpaired surrogates fail strictly);
//   - unknown names return nil (Python's LookupError) — the sniff
//     path takes over, exactly as in the reference.
//
// Codec-name normalization mirrors codecs.lookup: lowercase, every
// non-[a-z0-9] character becomes '_'.

//go:embed charsets_table.json
var charsetsJSON []byte

//go:embed cp949_table.json
var cp949JSON []byte

//go:embed dbcs_accept.json
var dbcsAcceptJSON []byte

type decodeFn func(body []byte) *string

var (
	ladderOnce sync.Once
	ladderMap  map[string]decodeFn
)

// normalizeCodecName is codecs.lookup's normalization (lowercase,
// non-alphanumerics → '_').
func normalizeCodecName(name string) string {
	b := []byte(asciiLower(name))
	for i, c := range b {
		if !((c >= 'a' && c <= 'z') || (c >= '0' && c <= '9')) {
			b[i] = '_'
		}
	}
	return string(b)
}

// pyDecode is the reference's _try: one strict decode attempt; nil
// means "failed" (unknown codec or undecodable bytes).
func pyDecode(body []byte, charset string) *string {
	if charset == "" {
		return nil
	}
	ladderOnce.Do(buildLadder)
	fn, ok := ladderMap[normalizeCodecName(charset)]
	if !ok {
		return nil
	}
	return fn(body)
}

func buildLadder() {
	m := make(map[string]decodeFn)

	// --- the utf-8 family (strict) -----------------------------------
	utf8Strict := func(body []byte) *string {
		if !utf8.Valid(body) {
			return nil
		}
		s := string(body)
		return &s
	}
	for _, a := range []string{"utf_8", "utf8", "cp65001"} {
		m[a] = utf8Strict
	}

	// --- the ascii family (strict) -----------------------------------
	asciiStrict := func(body []byte) *string {
		for _, b := range body {
			if b >= 0x80 {
				return nil
			}
		}
		s := string(body)
		return &s
	}
	for _, a := range []string{"ascii", "us_ascii", "us", "ansi_x3_4_1968", "646"} {
		m[a] = asciiStrict
	}

	// --- the single-byte charsets (generated tables) -----------------
	tables := loadSBTables()
	singleByte := func(table *[256]int32) decodeFn {
		return func(body []byte) *string {
			var b strings.Builder
			b.Grow(len(body))
			for _, c := range body {
				v := table[c]
				if v < 0 {
					return nil // the unmapped byte: Python's strict error
				}
				b.WriteRune(rune(v))
			}
			s := b.String()
			return &s
		}
	}
	// canonical name (normalized) → aliases (normalized)
	sbAliases := map[string][]string{
		"iso_8859_1":   {"iso_8859_1", "iso8859_1", "latin_1", "latin1", "latin", "l1", "8859", "cp819"},
		"windows_1252": {"windows_1252", "cp1252"},
		"windows_1254": {"windows_1254", "cp1254"},
		"iso_8859_15":  {"iso_8859_15", "iso8859_15", "latin9"},
		"iso_8859_2":   {"iso_8859_2", "iso8859_2"},
		"iso_8859_9":   {"iso_8859_9", "iso8859_9"},
		"windows_1250": {"windows_1250", "cp1250"},
		"windows_1251": {"windows_1251", "cp1251"},
		"koi8_r":       {"koi8_r"},
		"mac_roman":    {"mac_roman", "macroman", "macintosh"},
		"cp850":        {"cp850", "850", "ibm850"},
		"tis_620":      {"tis_620"},
		"cp874":        {"cp874"},
	}
	for canonical, aliases := range sbAliases {
		table, ok := tables[canonical]
		if !ok {
			panic("webextract: generated charsets table is missing " + canonical)
		}
		t := table
		for _, a := range aliases {
			m[a] = singleByte(&t)
		}
	}

	// --- utf-16 family (Python's BOM rules, strict) ------------------
	m["utf_16"] = func(body []byte) *string {
		if len(body) >= 2 {
			if body[0] == 0xFF && body[1] == 0xFE {
				return decodeUTF16(body[2:], false)
			}
			if body[0] == 0xFE && body[1] == 0xFF {
				return decodeUTF16(body[2:], true)
			}
		}
		return decodeUTF16(body, false) // native LE (the pinned platforms)
	}
	m["utf_16le"] = func(body []byte) *string { return decodeUTF16(body, false) }
	m["utf_16be"] = func(body []byte) *string { return decodeUTF16(body, true) }

	// --- utf-32 family ------------------------------------------------
	m["utf_32"] = func(body []byte) *string {
		if len(body) >= 4 {
			if body[0] == 0xFF && body[1] == 0xFE && body[2] == 0x00 && body[3] == 0x00 {
				return decodeUTF32(body[4:], false)
			}
			if body[0] == 0x00 && body[1] == 0x00 && body[2] == 0xFE && body[3] == 0xFF {
				return decodeUTF32(body[4:], true)
			}
		}
		return decodeUTF32(body, false)
	}
	m["utf_32le"] = func(body []byte) *string { return decodeUTF32(body, false) }
	m["utf_32be"] = func(body []byte) *string { return decodeUTF32(body, true) }

	// --- the DBCS family (acceptance bitmaps gate x/text) --------------
	// x/text's decoders accept more than Python's codecs (its Big5 rides
	// the WHATWG extension leads, its ShiftJIS maps 0x80) — every decode
	// is gated on acceptance bitmaps GENERATED from Python's own codecs
	// (dbcs_accept.json): which singles decode, which (lead, trail)
	// pairs decode to one character. gb18030 additionally rides the FFFD
	// scan for its piecewise 4-byte gaps (x/text decodes those to
	// FFFD-resync, so any U+FFFD means Python errored); the other
	// charsets' tables never map to U+FFFD, so no scan is due.
	dbcsAccept := loadDbcsAccept()
	makeDBCS := func(cs string, enc encoding.Encoding, ffddScan bool) decodeFn {
		acc, ok := dbcsAccept[cs]
		if !ok {
			panic("webextract: dbcs acceptance table is missing " + cs)
		}
		return func(body []byte) *string {
			i := 0
			for i < len(body) {
				b := body[i]
				if b < 0x80 || acc.single[b] {
					i++
					continue
				}
				if i+1 < len(body) && acc.pairAt(b, body[i+1]) {
					i += 2
					continue
				}
				if cs == "euc-jp" && b == 0x8F && i+2 < len(body) &&
					acc.tripleAt(body[i+1], body[i+2]) {
					i += 3
					continue
				}
				if cs == "gb18030" && gb18030FourAt(body, i) {
					i += 4
					continue
				}
				return nil
			}
			out, err := enc.NewDecoder().Bytes(body)
			if err != nil {
				return nil
			}
			decoded := string(out)
			if ffddScan && strings.ContainsRune(decoded, 0xFFFD) {
				return nil
			}
			return &decoded
		}
	}
	m["gbk"] = makeDBCS("gbk", simplifiedchinese.GBK, false)
	m["gb18030"] = makeDBCS("gb18030", simplifiedchinese.GB18030, true)
	m["big5"] = makeDBCS("big5", traditionalchinese.Big5, false)
	m["big5_tw"] = m["big5"]
	m["csbig5"] = m["big5"]
	m["shift_jis"] = makeDBCS("shift_jis", japanese.ShiftJIS, false)
	m["sjis"] = m["shift_jis"]
	m["euc_jp"] = makeDBCS("euc-jp", japanese.EUCJP, false)
	m["eucjp"] = m["euc_jp"]
	m["euc_kr"] = makeDBCS("euc-kr", korean.EUCKR, false)
	// gb2312: Python's strict euc-cn acceptance, decoded through GBK's
	// standard table (the shared GB2312 region).
	m["gb2312"] = makeDBCS("gb2312", simplifiedchinese.GBK, false)
	// cp949: the UHC extension — the generated table.
	var cp949 [][]int32
	if err := json.Unmarshal(cp949JSON, &cp949); err != nil {
		panic("webextract: embedded cp949 table is corrupt: " + err.Error())
	}
	m["cp949"] = func(body []byte) *string {
		var b strings.Builder
		i := 0
		for i < len(body) {
			c := body[i]
			if c < 0x80 {
				b.WriteByte(c)
				i++
				continue
			}
			if c < 0x81 || c > 0xFD || i+1 >= len(body) {
				return nil
			}
			t := body[i+1]
			if t < 0x41 || t > 0xFE {
				return nil
			}
			v := cp949[c-0x81][t-0x41]
			if v < 0 {
				return nil
			}
			b.WriteRune(rune(v))
			i += 2
		}
		s := b.String()
		return &s
	}

	ladderMap = m
}

// loadSBTables loads the generated single-byte tables, normalizing
// the canonical names to the codec-lookup form.
func loadSBTables() map[string][256]int32 {
	var raw map[string][256]int32
	if err := json.Unmarshal(charsetsJSON, &raw); err != nil {
		panic("webextract: embedded charsets table is corrupt: " + err.Error())
	}
	tables := make(map[string][256]int32, len(raw))
	for k, v := range raw {
		tables[normalizeCodecName(k)] = v
	}
	return tables
}

// dbcsAcceptSet is one charset's acceptance bitmaps (base64-loaded).
type dbcsAcceptSet struct {
	single  [256]bool
	pair    [128][256]bool
	triples *[128][256]bool // euc-jp only
}

func (a *dbcsAcceptSet) pairAt(lead, trail byte) bool {
	if lead < 0x80 {
		return false
	}
	return a.pair[lead-0x80][trail]
}

func (a *dbcsAcceptSet) tripleAt(b2, b3 byte) bool {
	if a.triples == nil || b2 < 0x80 {
		return false
	}
	return a.triples[b2-0x80][b3]
}

// loadDbcsAccept loads the generated acceptance bitmaps.
func loadDbcsAccept() map[string]*dbcsAcceptSet {
	var raw map[string]struct {
		Singles string  `json:"singles"`
		Pairs   string  `json:"pairs"`
		Triples *string `json:"triples"`
	}
	if err := json.Unmarshal(dbcsAcceptJSON, &raw); err != nil {
		panic("webextract: embedded dbcs acceptance table is corrupt: " + err.Error())
	}
	out := make(map[string]*dbcsAcceptSet, len(raw))
	for cs, e := range raw {
		acc := &dbcsAcceptSet{}
		singles := decodeBitmap(e.Singles, 32)
		for i := 0; i < 256; i++ {
			if singles[i/8]&(1<<(i%8)) != 0 {
				acc.single[i] = true
			}
		}
		pairs := decodeBitmap(e.Pairs, 128*256/8)
		for lead := 0; lead < 128; lead++ {
			for t := 0; t < 256; t++ {
				off := (lead*256 + t) / 8
				bit := uint((lead*256 + t) % 8)
				if pairs[off]&(1<<bit) != 0 {
					acc.pair[lead][t] = true
				}
			}
		}
		if e.Triples != nil {
			triples := decodeBitmap(*e.Triples, 128*256/8)
			var t [128][256]bool
			for b2 := 0; b2 < 128; b2++ {
				for b3 := 0; b3 < 256; b3++ {
					off := (b2*256 + b3) / 8
					bit := uint((b2*256 + b3) % 8)
					if triples[off]&(1<<bit) != 0 {
						t[b2][b3] = true
					}
				}
			}
			acc.triples = &t
		}
		out[cs] = acc
	}
	return out
}

// decodeBitmap decodes one base64 bitmap, verifying its length.
func decodeBitmap(b64 string, wantBytes int) []byte {
	raw, err := base64.StdEncoding.DecodeString(b64)
	if err != nil || len(raw) != wantBytes {
		panic("webextract: corrupt bitmap in dbcs acceptance table")
	}
	return raw
}

// gb18030FourAt checks the 4-byte grammar (the linear region):
// (0x81-0xFE)(0x30-0x39)(0x81-0xFE)(0x30-0x39).
func gb18030FourAt(body []byte, i int) bool {
	if i+3 >= len(body) {
		return false
	}
	b1, b2, b3, b4 := body[i], body[i+1], body[i+2], body[i+3]
	return b1 >= 0x81 && b1 <= 0xFE &&
		b2 >= 0x30 && b2 <= 0x39 &&
		b3 >= 0x81 && b3 <= 0xFE &&
		b4 >= 0x30 && b4 <= 0x39
}

func decodeUTF16(body []byte, be bool) *string {
	if len(body)%2 != 0 {
		return nil
	}
	units := len(body) / 2
	var b strings.Builder
	b.Grow(units)
	unit := func(i int) uint16 {
		if be {
			return uint16(body[2*i])<<8 | uint16(body[2*i+1])
		}
		return uint16(body[2*i+1])<<8 | uint16(body[2*i])
	}
	i := 0
	for i < units {
		u := unit(i)
		switch {
		case u >= 0xD800 && u <= 0xDBFF:
			if i+1 >= units {
				return nil // unpaired high surrogate
			}
			u2 := unit(i + 1)
			if u2 < 0xDC00 || u2 > 0xDFFF {
				return nil
			}
			r := 0x10000 + (rune(u)-0xD800)<<10 + (rune(u2) - 0xDC00)
			b.WriteRune(r)
			i += 2
		case u >= 0xDC00 && u <= 0xDFFF:
			return nil // lone low surrogate
		default:
			b.WriteRune(rune(u))
			i++
		}
	}
	s := b.String()
	return &s
}

func decodeUTF32(body []byte, be bool) *string {
	if len(body)%4 != 0 {
		return nil
	}
	var b strings.Builder
	for i := 0; i < len(body); i += 4 {
		var v uint32
		if be {
			v = uint32(body[i])<<24 | uint32(body[i+1])<<16 |
				uint32(body[i+2])<<8 | uint32(body[i+3])
		} else {
			v = uint32(body[i+3])<<24 | uint32(body[i+2])<<16 |
				uint32(body[i+1])<<8 | uint32(body[i])
		}
		if (v >= 0xD800 && v <= 0xDFFF) || v > 0x10FFFF {
			return nil
		}
		b.WriteRune(rune(v))
	}
	s := b.String()
	return &s
}

// pyUTF8Replace is bytes.decode('utf-8', errors='replace'): one U+FFFD
// per maximal invalid subpart (Go's strings.ToValidUTF8 collapses
// whole invalid runs into a single U+FFFD — a different shape; the
// decode_law golden section pins this one).
func pyUTF8Replace(body []byte) string {
	var b strings.Builder
	b.Grow(len(body))
	i := 0
	n := len(body)
	for i < n {
		c := body[i]
		if c < 0x80 {
			b.WriteByte(c)
			i++
			continue
		}
		var need int
		var lo, hi byte // the first continuation's allowed range
		switch {
		case c >= 0xC2 && c <= 0xDF:
			need, lo, hi = 1, 0x80, 0xBF
		case c == 0xE0:
			need, lo, hi = 2, 0xA0, 0xBF
		case c >= 0xE1 && c <= 0xEC:
			need, lo, hi = 2, 0x80, 0xBF
		case c == 0xED:
			need, lo, hi = 2, 0x80, 0x9F
		case c >= 0xEE && c <= 0xEF:
			need, lo, hi = 2, 0x80, 0xBF
		case c == 0xF0:
			need, lo, hi = 3, 0x90, 0xBF
		case c >= 0xF1 && c <= 0xF3:
			need, lo, hi = 3, 0x80, 0xBF
		case c == 0xF4:
			need, lo, hi = 3, 0x80, 0x8F
		default: // 0xC0, 0xC1 (overlong), 0xF5-0xFF: invalid lead
			b.WriteRune(0xFFFD)
			i++
			continue
		}
		// consume the maximal valid subpart
		j := i + 1
		broken := false
		if j >= n || body[j] < lo || body[j] > hi {
			broken = true
		} else {
			j++
			for k := 1; k < need && !broken; k++ {
				if j >= n || body[j] < 0x80 || body[j] > 0xBF {
					broken = true
					break
				}
				j++
			}
		}
		if broken {
			// one replacement for the subpart consumed; resync at the
			// byte that broke it (or EOF)
			b.WriteRune(0xFFFD)
			i = j
			continue
		}
		// a complete valid sequence
		b.Write(body[i : i+need+1])
		i += need + 1
	}
	return b.String()
}

// sniffCharset is extract_from_bytes' meta sniff: the first 4096 bytes
// ASCII-ignore-decoded and lowercased, then
// re.search(r'charset=["\']?([a-z0-9_\-]+)', head).
var sniffRe = regexp.MustCompile(`charset=["']?([a-z0-9_\-]+)`)

func sniffCharset(body []byte) (string, bool) {
	limit := len(body)
	if limit > 4096 {
		limit = 4096
	}
	var b strings.Builder
	b.Grow(limit)
	for _, c := range body[:limit] {
		if c < 0x80 {
			if c >= 'A' && c <= 'Z' {
				c += 'a' - 'A'
			}
			b.WriteByte(c)
		}
	}
	m := sniffRe.FindStringSubmatch(b.String())
	if m == nil {
		return "", false
	}
	return m[1], true
}

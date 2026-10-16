// Package webextract ports core/web_extract.py — HTML extraction and
// the two partial-page heuristics (T-32, Phase 3): fetched HTML
// becomes title / meta description / visible text plus the JS-shell
// and paywall flags; extract_from_bytes rides the charset ladder
// (header charset → the page's own <meta charset> sniff → UTF-8 with
// replacement).
//
// Spec: docs/port-blueprint/module-specs/spec-core-web-extract.md;
// task T-32. internal/webfetch (SESSION-012) is the working example
// this port follows — the fourteenth golden generator
// (testdiff/goldens/dump_webextract.py) drives the unmodified
// reference and pins this package byte-for-byte.
//
// THE PACKAGE LAWS:
//   - Extraction never fails on bad input: a broken page is a partial
//     result (the tokenizer is a total function; Python's try/except
//     around feed/close has no Go counterpart to need).
//   - THE HAND-ROLLED TOKENIZER (the dependency verdict): the owner's
//     x/net/html verdict did not arrive with SESSION-013's prompt, so
//     the reference's own stdlib parser behavior (CPython 3.12's
//     html.parser, convert_charrefs=True) is ported by hand — the
//     zero-dependency alternative NEXT-SESSION §4 names. NO new
//     import was added; x/net/html remains one ADR away, and the
//     swap, if ever approved, is confined to parser.go's scan
//     internals (the handler laws and the goldens stay the referee).
//   - The entity table and the single-byte charset tables are
//     GENERATED from the reference runtime's own stdlib (html.entities.
//     html5, the codecs registry) and embedded — never hand-typed;
//     pinned at sha256 level in the pure golden.
//   - Python string semantics are carried exactly where they are
//     observable: str.strip/split/lower (including U+0130's two-rune
//     lowercase and the \x1c-\x1f whitespace row), regex \s, the
//     `<script\b` Unicode word boundary, rune-count slicing, the
//     20000/4000-char caps by codepoint.
//   - Tag and attribute NAME lowercasing is ASCII-only: no member of
//     the reference's tag sets contains a non-ASCII rune, and no
//     non-ASCII rune can lower into a pure-ASCII member (the only
//     Unicode lowercase INTO an ASCII letter is the Kelvin sign
//     U+212A → 'k', and no set member contains 'k') — documented, not
//     accidental.
//
// DOCUMENTED DIVERGENCES (each recorded per the port law):
//   - The DBCS family (gbk/gb18030/big5/shift_jis/euc-jp/euc-kr)
//     rides x/text's decoders behind a strict-reject layer: x/text
//     replaces invalid sequences with U+FFFD, so any U+FFFD in a
//     decode means Python's strict decode failed → the ladder falls
//     through (the sniff/replace path), matching the reference's
//     observable verdicts. The residual edge: a page that legitimately
//     CONTAINS U+FFFD (only representable in gb18030 among this
//     family) would be misjudged as undecodable — no test pins it,
//     Python's own table differences beyond the pinned probes are
//     likewise unpinned.
//   - gb2312 is the strict euc-cn grammar (lead 0xA1-0xF7, trail
//     0xA1-0xFE) decoded through GBK's standard table — the grammar
//     carries Python's strictness, the table the shared standard.
//   - utf-16/utf-32 without a BOM decode little-endian: Python uses
//     the platform's native order, and every platform this project
//     pins (the golden venv, the Windows target) is little-endian.
//   - The one-pass tokenizer: the reference feeds the whole document
//     then closes; the Python "&-near-end" buffering heuristic is a
//     two-phase feed optimization whose close()-time flush emits the
//     same span this single pass emits (the unescape battery pins the
//     observable result either way).
//   - Python's codecs registry is vaster than this ladder; unknown
//     names return nil (LookupError) and the sniff path takes over —
//     identical observable behavior for every name the golden battery
//     pins; names outside the ladder's table that PYTHON knows (e.g.
//     obscurer aliases) would diverge by falling to the sniff path
//     earlier — the realistic web charset set is covered.
package webextract

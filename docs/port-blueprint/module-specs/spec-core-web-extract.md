# Module Spec — `app/gitcurator/core/web_extract.py`

- **LOC**: 227 · **Go home**: `internal/webextract` · **Status**: port
- **Responsibility**: fetched HTML → title / meta description / visible text + the two partial
  heuristics (JS shell, paywall). Pure stdlib `html.parser`; never raises.

## Constants (`web_extract.py:`)

`MAX_TEXT_CHARS 20_000` · `_JS_SHELL_MIN_TEXT 120` · `_PAYWALL_MARKERS` (34-45: 'subscribe to
continue', 'subscription required', … 'reader mode is not available') ·
`_INVISIBLE_TAGS {script,style,noscript,template,svg,head}` · `_BLOCK_TAGS` (51-55).

## Symbols

| Symbol | Lines | Contract |
|---|---|---|
| `Extracted` | 58-83 | slots `(title, meta_description, text, is_js_shell, is_paywall, charset_used)`; `.has_content` = title or desc or `len(text) ≥ 80` |
| `_PageParser` | 86-160 | HTMLParser `convert_charrefs=True`; collects `<title>` text; `<meta name="description">` **then** `og:description` (name wins); charset from `<meta charset>` / `http-equiv` via `re.search(r'charset=([^\s;]+)', IGNORECASE)`; invisible-depth tracking; block tags emit `\n` |
| `extract(html)` | 163-195 | whitespace collapse (`[ \t]+→' '`, `\n\s*\n+→\n`), truncate 20k; `is_js_shell = word_chars < 120 AND '<script\b' in raw`; `is_paywall = any marker in first 4000 chars lowered`; parse errors keep collected data |
| `extract_from_bytes(body, header_charset='')` | 198-227 | header charset → `<meta charset>` sniff from first 4096 bytes (ASCII-ignore decode) → utf-8 replace; backfills `charset_used` |

## Go notes

- **Amendment (SESSION-013, T-32 landed — reality wins over specs)**: the
  owner's verdict on `golang.org/x/net/html` did NOT arrive with the session
  prompt (it had been surfaced since SESSION-011), so the port took
  NEXT-SESSION §4's option (b): a hand-rolled tokenizer byte-faithful to
  CPython 3.12's `html.parser` in its convert_charrefs=True configuration
  (the only one the reference uses) — zero new imports, no ADR needed. The
  tokenizer lives isolated in `internal/webextract/parser.go`; if the owner
  later approves an x/net/html ADR, the swap is confined to that file's scan
  internals (the handler laws and the goldens stay the referee).
- The original recommendation (kept for the record): use `golang.org/x/net/html`
  walking with the same conservative rules — title text concatenation,
  description precedence (name before og), invisible subtree skipping,
  block-tag newlines, charref decoding (parser default in Go too).
- Keep the regex-based JS-shell/paywall heuristics (they operate on raw HTML, not the DOM).
- The charset ladder (extract_from_bytes) rides tables GENERATED from the
  reference runtime's own stdlib (`html.entities.html5`, the codecs
  registry — embedded, sha-pinned by the golden) with x/text behind
  generated acceptance bitmaps for the DBCS family: x/text's decoders
  accept strictly more than Python's codecs (Big5 extension leads,
  Shift-JIS 0x80), so every DBCS decode is gated on Python's own
  acceptance sets (see `internal/webextract/doc.go` for the documented
  divergences, incl. the gb18030 FFFD-scan edge).

## Tests

The 5 shipped fixtures (`app/tests/fixtures/web/{article,landing,latin1,js_shell,paywall}.html`)
byte-compared on all Extracted fields; edge cases: no title, both metas, nested invisible,
truncation at 20k, latin-1 charset header, BOM.

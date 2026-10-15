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

- Use `golang.org/x/net/html` walking with the same conservative rules: title text
  concatenation, description precedence (name before og), invisible subtree skipping, block-tag
  newlines, charref decoding (parser default in Go too).
- Keep the regex-based JS-shell/paywall heuristics (they operate on raw HTML, not the DOM).

## Tests

The 5 shipped fixtures (`app/tests/fixtures/web/{article,landing,latin1,js_shell,paywall}.html`)
byte-compared on all Extracted fields; edge cases: no title, both metas, nested invisible,
truncation at 20k, latin-1 charset header, BOM.

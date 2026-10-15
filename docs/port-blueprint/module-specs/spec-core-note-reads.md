# Module Spec — `app/gitcurator/core/note_reads.py`

- **LOC**: 663 · **Go home**: `internal/notereads` · **Status**: port (with one documented fix)
- **Responsibility**: v0.71.0 THE ONE-OPEN LAW — shared read cache: RAM text layer (one bounded
  read per note per scan, FIFO budget) + persisted parse layer (unified frontmatter + body banish
  tag per note in cache.db, stat-fingerprinted) so the next scan of unchanged notes opens nothing.

## Symbols (`note_reads.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `DEFAULT_TEXT_BYTES` / `DEFAULT_RAM_BUDGET` | 60/64 | 200,000 / 32 MB |
| `_FM_STRING_KEYS` / `_REPOS_STRING_KEYS` | 71-79 | unified string keys; `('stars',)` read as string |
| `_LOCK_KEYS` / `_CURATED_KEYS` | 85-89 | `('locked','category_locked')` / `('curated',)` truthy-folded |
| `_PARSE_VERSION` | 100 | **4** — DB rows without `v==4` re-parse once |
| `_BANISH_KEYS` / `_BANISH_EMOJI` / `_BANISH_WORDS` | 105-120 | `('decommission','banish','blacklist')` / `'🗑️'` / 6 words (mirrors; real constants lazy-imported from website_pipeline and win) |
| `_BODY_TAG_STOPS` | 110 | `" \t\r\n,.;:!?)]}'\"»«…—>"` |
| `_read_bounded(path, max_bytes)` | 147-156 | the ONLY file-open site; utf-8 `errors='replace'`; any exception → `''` |
| `_stat_fingerprint(path)` | 159-167 | `(mtime_ns, size)` |
| `_truthy` | 170-175 | `true/yes/1` after quote-strip |
| `parse_frontmatter_unified(text)` | 178-243 | string-scan parser (no YAML lib): first line `---`, next `---`; flow tags `tags: [a, b]` + block tags (`- item`); returns the unified dict; never raises |
| `body_banish_tag_of(text)` | 246-294 | skips frontmatter (unterminated → `''`); fence toggling on ` ``` `/`~~~`; `#` token grammar (next char not space/`#`; prev char in space/`([`; token until stop char); matches emoji-substring or exact banish word |
| `NoteReadCache` | 297-592 | two layers + honesty counters (`opens`, `ram_hits`, `db_hits`, `db_writes`) |
| `text(path, max_bytes)` | 400-417 | hit requires same fingerprint **and** cached bound ≥ ask (serves slice); else one bounded read + FIFO remember |
| `frontmatter/banish_frontmatter/banish_tag` | 443-536 | shape adapters (inventory/banish shapes byte-identical with direct parsers) |
| `close(prune=True)` | 563-592 | prune dead-path rows; close connection |

## SQLite schema

```sql
note_reads (path TEXT PRIMARY KEY, mtime_ns INTEGER NOT NULL, size INTEGER NOT NULL,
  frontmatter_json TEXT NOT NULL DEFAULT '', banish_tag TEXT NOT NULL DEFAULT '',
  scanned_at REAL NOT NULL DEFAULT 0);
```

JSON carries `v: 4`; broken DB → `_db_broken=True` → RAM-only degradation. Connection: own
connection, busy_timeout 30000; all under one RLock (thread-safe, serializing).

## Known quirks → Go dispositions

1. **`_unified_from_json` drops `stars` on DB hits** (iterates only `_FM_STRING_KEYS`) —
   **FIX in Go** (restore `stars`), regression test, note in report (doc 09 §5).
2. RAM budget counted in **characters** (`len(str)`) — **FIX**: byte budget (internal only).
3. `_banish_tag` DB path is not version-gated (grammar stable) — preserve.
4. Larger ask re-opens once (honest) — preserve.

## Tests

Port `test_notereads`: one-open law across hygiene/banishment/inventory consumers; fingerprint
miss re-read; version-marker re-read once; FIFO eviction; broken-DB degradation; adapter shape
parity; plus the new stars regression test.

# Module Spec — `app/gitcurator/core/note_state.py`

- **LOC**: 842 · **Go home**: `internal/notestate` · **Status**: port
- **Responsibility**: SPEC §4.4 "Moves are corrections" — persistent per-note state + the
  run-start reconciliation state machine + the dismissed list + the recall-block markers.

## Symbols (`note_state.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `RECALL_START/END` | 57-58 | `<!-- gitcurator:recall:start/end -->` |
| `SKIPPED_FOLDER_MARKS` | 66-67 | `('_moc','_inbox','_missing','attachments','.obsidian','.trash')` — **substring** match on root path (`_skip_dir` 102-103) |
| `FOLDER_TO_CATEGORY` | 70 | reverse of CATEGORY_FOLDERS |
| `VAULT_GITHUB/VAULT_WEBSITES` | 77-78 | `'github'`/`'websites'` labels (identity = label) |
| `compute_fingerprint(content)` | 85-95 | strip recall block → `rstrip()` → sha256 hex (utf-8 replace) |
| `scan_vault(vault_path, resolver=None)` | 106-154 | os.walk (unsorted!); skip dirs; `.md` only; head read **800 chars**; `^source:\s*(.+)$`; source normalized (`normalize_url`) + `strip('"\'')`; rows `{path, source_url, folder, category, subcategory, unmanaged}` |
| `folder_category(folder)` | 157-175 | exact reverse-map → longest registered prefix → else `(folder, None)` = unmapped; `''`/`'_review'` → `(None,None)` |
| `NoteStateDB` | 182-436 | SQLite (schema below); `db_path=="cache.db"` → anchored `APP_DIR/cache.db`; `check_same_thread=False`, timeout 30, busy_timeout 30000, RLock |
| `record_baseline_if_empty` | 443-474 | records every managed note **only when vault has zero rows** |
| `classify_changes` | 477-578 | pure detector (state machine §below) |
| `apply_corrections` | 678-796 | writes frontmatter updates via bounded surgery; dry-run → no-op |
| `run_start_check` | 799-842 | baseline → classify → (live conn?) apply |
| `fm_replace_field` / `fm_swap_tag` / `_fm_bounds` | 604-662 | frontmatter surgery: replace-or-insert-before-closing-`---`; `tags: [..]` line only; **new tag prepended** |

## SQLite schema

```sql
note_state (id INTEGER PRIMARY KEY AUTOINCREMENT, vault TEXT NOT NULL,
  source_url TEXT NOT NULL, path TEXT NOT NULL, fingerprint TEXT NOT NULL,
  category TEXT, subcategory TEXT, locked INTEGER NOT NULL DEFAULT 0,
  first_seen TEXT NOT NULL, updated_at TEXT NOT NULL, UNIQUE(vault, source_url));
CREATE INDEX idx_note_state_vault ON note_state(vault);
corrections_log (id, ts, vault, source_url, from_category, to_category,
  from_path, to_path, note) + idx_corrections_vault;
note_state_dismissed (vault, source_url, last_path, dismissed_at,
  PRIMARY KEY (vault, source_url));
```

Timestamps: naive-local `datetime.now().isoformat(timespec='seconds')` → `YYYY-MM-DDTHH:MM:SS`.

## Classification state machine (`classify_changes`)

Per source on disk: duplicates (>1 file same source) → flag only, touch nothing. Single: no state
row → **unknown**; `normcase(disk) != normcase(row.path)` → **moved** (category read from the
folder, not frontmatter); same path + fingerprint mismatch → **edited** (report only); state rows
without disk presence → **deleted**; files without `source:` → **unmanaged**; known note in a
folder resolving to no category → **unmapped**.

`apply_corrections`: moved → skip-if-unmapped; else update `category` line (+ prepended tag swap),
`subcategory:` line (websites always — `""` when empty; GitHub only when sub exists),
`category_locked: true` inserted; write via atomic_write_text **only if changed**; `upsert(locked=1)`;
`log_correction`. deleted → `dismiss` (never re-added). Others report-only.

## Migration-sensitive semantics

- `os.walk` unsorted (scan order = winner of duplicate fingerprints etc.) — Go: lexical walk,
  differential cases pin winners (documented divergence class).
- `os.path.normcase` — Windows folds case+slashes, POSIX identity; port per-platform.
- First-800-char head read (code points).
- Insert-before-closing-`---` surgery must not touch unterminated frontmatter.

## Tests

Port: baseline-first-run silence; each classify branch; corrections apply (category line, tag
prepend, lock, subcategory rules); dismissed never re-added; dry-run no-op; surgery on
unterminated blocks. Add: Windows normcase matrix (cased paths).

> **SESSION-061 (the dated correction) — the SURGERY path's text-mode
> law, closed as the recorded CRLF round-trip differential (the
> SESSION-008 remainder).** Settled at byte level against the reference
> before any port change: every reference read of a note file is TEXT
> MODE — `open(path, 'r', encoding='utf-8', errors='replace')` — so
> universal newlines translate `\r\n` and lone `\r` to `\n` on READ,
> and `atomic_write_text` writes with `newline=''` (raw, no
> translation) — therefore a CRLF note that rides the
> moves-as-corrections surgery comes out ALL-LF (the CRLF is eaten by
> the read, never restored by the write), and every fingerprint the
> module stores (`record_note`'s own file read, the baseline,
> classify's current read, apply's row refresh) is the fingerprint of
> the TRANSLATED content. The port had FOUR raw reads
> (`os.ReadFile` + replace-decode, no translation) at exactly those
> sites, so the Go leaked `\r`'s into the surgery (mixed-endings
> output where the reference writes all-LF) and stored RAW
> fingerprints — a cross-binary `cache.db` disagreement: a
> Python-baselined CRLF row would read as perpetually edited to a
> raw-reading twin (the two apps share the DB by design, so the
> contract is real). The empirical proof, generated-never-hand-typed:
> `testdiff/goldens/dump_notestate.py` §5 re-rides the reference's own
> test_phase3 selections (the `_github_note` fixture + the moved/edited
> selections) over CRLF-written notes — the Windows norm — and
> recorded the reference's own bytes as
> `internal/notestate/testdata/notestate_crlf_roundtrip.golden.json`
> (byte-identical on regen; the reference's own asserts green twice:
> moved 1, applied 1, edited repo2, the output CR count ZERO);
> `internal/notestate/crlfsurgery_test.go` re-rides the same fixture
> bytes through the port and FAILED on its first run with the exact
> predicted divergences (the raw fingerprints `92a41fd6…` vs the
> reference's `9c45de47…`/`266ca069…`; the round-trip output carrying
> NINE `\r`'s where the reference writes none). The fix:
> `helpers.PyNewlines` after the replace-decode at the four read
> sites — the same law `readHead800` (SESSION-008),
> `StampCurated`/`RetireNoteFile` already ride. All green post-fix
> (the differential 1/1, the package 57, the battery 2397 + 89,
> race-clean). test_phase3 itself sits behind the PyQt6 import wall in
> the stdlib-only sandbox venv (it runs in the reference's own CI);
> the reference's own bytes through the generator are the pinned
> authority, and `test_notereads` (16/16, loadable) green before and
> after. The MULTIEDIT-TAB corruption struck a FOURTH time, in a new
> shape: a FAILED MultiEdit still re-expanded the whole `notestate.go`
> to spaces AND left the partial edits applied (the tool's atomicity
> claim is not trustworthy); caught by the standing `gofmt -l`,
> repaired by `gofmt -w` + a python patch on the real bytes, the
> post-repair diff verified to be exactly the four edit regions.

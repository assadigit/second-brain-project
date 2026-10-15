# Module Spec — `app/gitcurator/core/storage.py` + `core/dryrun.py`

- **LOC**: 182 + 304 · **Go home**: `internal/storage`, `internal/dryrun` · **Status**: port
- **Responsibility**: storage = atomic filesystem writes + safe filenames + the pure config-merge
  law; dryrun = the global rehearsal switch + withheld-operation log + shadow cache copy.

## storage.py (`storage.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `atomic_write_text(path, content, encoding='utf-8')` | 37-70 | (1) dry-run gate: record op `'atomic-write'` with content, return; (2) `makedirs(dirname(abspath))`; (3) `mkstemp(prefix='.tmp_', suffix=<ext or '.tmp'>, dir=same-dir)`; (4) write text-mode `newline=''` (no translation); (5) `flush` + `os.fsync`; (6) `os.replace` (atomic NTFS/POSIX); (7) on BaseException: best-effort unlink tmp, re-raise. **No chmod; no dir fsync.** |
| `atomic_write_bytes(path, data)` | 73-99 | same in `'wb'`; dry-run records bytes content |
| `safe_filename(name)` | 106-112 | `re.sub(r'[^a-zA-Z0-9\-_]+', '_', name or '')` — traversal-proof (Unicode → `_`) |
| `unique_path(path)` | 115-130 | existing → `_v1`, `_v2`… before extension (not race-free; callers re-check) |
| `build_note_filename(repo, category_key, tags)` | 133-138 | `"{safe(repo)}_{safe(cat.replace('/','_'))}_{safe(tags[0]) or 'misc'}.md"` |
| `merge_config(existing, updates)` | 145-177 | shallow copy; updates win; **nested dicts merged key-by-key**; unknown existing keys preserved; tuning keys (`timeout_per_repo`/`max_retries`/`delay_between_api_calls`) only defaulted when absent (60/3/0.5) |
| `write_config_file(path, config)` | 180-182 | `atomic_write_text(path, json.dumps(config, indent=2, ensure_ascii=False))` |

**Go notes**: `os.CreateTemp` + write + `Sync()` + `os.Rename`; JSON encoder with
`SetEscapeHTML(false)` + `SetIndent("", "  ")` (Python `ensure_ascii=False` parity — HTML-entity
escaping would corrupt byte-parity); merge law implemented over `map[string]any`.

## dryrun.py (`dryrun.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `enable()` / `disable()` | 74-90 | ON clears log + resets shadow; OFF keeps log, cleans shadow dir |
| `is_enabled()` | 93-95 | reads global without lock (benign) |
| `record(op, path, content=None, note="")` | 102-131 | appends `{time '%H:%M:%S', op, path, size (bytes), preview, note}`; bytes → utf-8-replace preview; newlines → `'⏎'`; preview cap 160 chars + `'…'`; never raises |
| `entries/entry_count/clear/summary` | 134-157 | read/clear/counts-per-op |
| `makedirs/write_text/append_text/remove/move` | 165-208 | stand-ins; ops `'makedirs'/'write'/'append'/'remove'/'move'`; move records on **dst** with `note=f'from {src}'`; `remove` raises OSError when disabled (like `os.remove`) |
| `shadow_cache_path(source_db)` | 215-258 | one copy per enable-session in `tempfile.mkdtemp(prefix='gitcurator-dryrun-')`; copy happens **outside the lock** (deadlock lesson); copies `-wal`/`-shm` sidecars; missing source → info entry; failed copy → warning + empty shadow |
| `render_report_markdown(title)` | 273-304 | counts sorted; table `| # | Time | Op | Target | Size | Preview/note |`; sizes `f"{n:,} B"`; `'|'`→`'\|'` |

**Concurrency**: `threading.Lock` around entries/shadow (banner thread writes markers
concurrently); single-batch-at-a-time assumption.

**Go notes**: `DryRun` struct with mutex; recorder entries carry the same fields; the shadow copy
must copy `cache.db`, `cache.db-wal`, `cache.db-shm` when present.

## Compatibility requirements

1. Temp files: prefix `.tmp_`, suffix = target extension, same directory (same-filesystem rename).
2. fsync before rename; failure cleanup; no extra chmod/fsync-dir (parity, even if suboptimal).
3. Dry-run report byte-comparable (differential case: same batch dry-run in Python vs Go →
   compare normalized reports).
4. `merge_config` semantics are load-bearing for the GUI/CLI save paths.

## Tests

- Atomicity: pre-existing content survives a failed write (inject error); readers never see
  partial content (torn-write probe).
- `safe_filename` traversal corpus (`../..`, `a/b`, `..`, unicode, empty).
- `merge_config` nested-merge + tuning-key + unknown-key cases (port from `test_core`).
- Dry-run: enable → write via stand-ins → entries recorded, disk untouched; shadow cache copies
  WAL sidecars; report renders.

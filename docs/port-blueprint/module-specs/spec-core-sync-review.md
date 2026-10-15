# Module Spec — `app/gitcurator/core/sync_review.py`

- **LOC**: 968 · **Go home**: `internal/syncreview` · **Status**: port
- **Responsibility**: the pure-decision half of Settings → Sync: the four-state cell model, the
  banner verdict, relative times, the chip grammar, and the DeferredLog SQLite ledger that
  remembers every review; plus read-only doors for the published index, vault-side items, and the
  git backup gap. Qt-free.

## Symbols (`sync_review.py:`)

| Symbol | Lines | Contract |
|---|---|---|
| `STATES ('behind','unknown','reviewed','match')` · `REASONS ('not_ready','private','duplicate','other')` + labels | 64-74 | the grammar |
| `cell_state(set_up, src, tgt, deferred_gap)` | 82-100 | not set up/unmeasured → `unknown`; gap = max(0, src−tgt); 0 → `match`; `deferred ≥ gap` → `reviewed`; else `behind` |
| `stage_cell(...)` | 114-153 | gap = first-seen order dedupe (`dict.fromkeys`) minus target set |
| `banner_verdict(cells)` | 156-234 | worst-first: Behind → "N items waiting" (named stages); Unknown → split **"not set up"** (needs='setup') vs **"not measured"** (needs='measure'); else "In sync · N deferred" / "In sync" |
| `cell_caption(stage)` | 237-309 | micro-copy per stage suffix (`:queue_to_vault`, `:vault_to_site`, `:vault_to_backup`) — v0.91.0 THE MICRO-COPY LAW |
| `format_added_ago(added_at, now)` | 340-368 | "just now"/"N m ago"/"N h ago"/"N d ago"/"N mo ago"/"N y ago"; `_DT_RE ^(\d{4})-(\d{2})-(\d{2})(?:[T ](\d{2}):(\d{2})(?::(\d{2}))?)?`; UTC; negative skew clamped; months = days//30 |
| `sort_oldest_first` / `mark_allowed(checked, reason)` | 384-405 | dated items by −days, undated last; ≥1 item AND a chip |
| `DeferredLog` | 437-708 | schema below; statuses `active/resolved/reopened`; `mark` (upsert, never deletes), `reopen` (only active→reopened, `reopen_count+1`), `resolve_seen` (present→resolved, absent→active), `resolve_gone`, `reviewed_ids`, `rows` (newest-first, case-folded search); stable IDs only (note path or URL); WAL, RLock, timeout 30, busy_timeout 30000; **UTC-aware** ISO timestamps |
| `fetch_published_index(worker_url, timeout=15)` | 719-751 | GET `/api/directory/index`, UA from cloudflare_sync, never raises, requires `{'sites': list}` |
| `vault_site_items(vault, log, cap=2000)` | 775-867 | mirrors directory_sync filters but keeps `note_path` + `added_at`; id = normalized URL |
| `vault_backup_gap(vault, log, cap=4000)` | 874-945 | no `.git` → every .md pending; else `git status --porcelain` (read-only), rename `old -> new` takes new path, skips `.obsidian/`+`.trash/`; id = repo-relative path |
| `vault_note_count(vault)` | 948-968 | — |

## Schema

```sql
sync_deferrals (id INTEGER PRIMARY KEY AUTOINCREMENT, stage TEXT NOT NULL,
  item_id TEXT NOT NULL, display_name TEXT DEFAULT '', reason TEXT NOT NULL,
  note TEXT DEFAULT '', status TEXT NOT NULL DEFAULT 'active',
  date_reviewed TEXT NOT NULL, date_resolved TEXT,
  reopen_count INTEGER NOT NULL DEFAULT 0, UNIQUE(stage, item_id));
CREATE INDEX idx_sync_deferrals_stage_item ON sync_deferrals(stage, item_id);
```

Lifecycle: mark → ACTIVE; reopen → REOPENED; item appears in target (or leaves gap) → RESOLVED
automatically (row kept forever; flip-back when the gap returns); NOTHING is ever deleted.

## Tests

Port `test_syncreview` (143): truth table, banner precedence incl. the setup/measure split, chip
grammar, ledger lifecycle + concurrency, relative clock, vault/git/published helpers, never-
delete law.

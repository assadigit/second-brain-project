# Module Spec — `gui/cache_db.py` + `gui/link_tracker.py` + `gui/vault_index.py` + `gui/link_helpers.py` + `gui/dead_links.py`

- **LOC**: 876 + 692 + 294 + 62 + 70 · **Go home**: `internal/store` (+ `repos.go`) and
  `internal/batch` (linktracker, vaultindex) · **Status**: port (grouped: the dedupe/persistence
  trio + their delegates)
- **Responsibility**: the GitHub-side SQLite sidecar, the 5-phase no-link-lost manifest, the
  in-memory URL→note index, and three tiny delegates.

## cache_db.py (`gui/cache_db.py:`)

**The locking law** (106-147): own connection `check_same_thread=False, timeout=30`;
`PRAGMA busy_timeout=30000`, `journal_mode=WAL`, `synchronous=NORMAL`; **every statement under
`threading.RLock`** ("sqlite connections with check_same_thread=False are NOT safe for
concurrent cursor use"). Default path anchored to APP_DIR (the GUI+CLI shared-cache fix).

**DDL** (149-288):

```sql
processed_repos (repo_id INTEGER PRIMARY KEY, url TEXT, owner TEXT, repo_name TEXT,
  processed_at TIMESTAMP, last_checked TIMESTAMP, note_path TEXT, category TEXT)
+ idx_processed_repos_url, idx_processed_repos_note_path
checkpoints (…)                       -- DEAD (created, never used) — not ported
failed_repos (id AUTOINCREMENT, url TEXT, error TEXT, failed_at TIMESTAMP,
  retry_count INTEGER DEFAULT 0, resolved BOOLEAN DEFAULT 0) + idx_failed_repos_url
decommissioned_repos (url TEXT PRIMARY KEY, reason TEXT, decommissioned_at TIMESTAMP,
  fail_count INTEGER NOT NULL DEFAULT 3)
repos_settled (url TEXT PRIMARY KEY, settled_at TIMESTAMP)
cache_meta (key TEXT PRIMARY KEY, value TEXT)
```

Migrations at open: `decommissioned_repos.fail_count` column add; one-time `notfound_strikes`
merge+DROP; meta guard `repos_settled_at`.

**Behavior laws**: all writes best-effort (never crash the batch on DB error);
`record_404` **fails CLOSED** (returns threshold on error — quarantine is fail-safe, 497-519);
`settle_repos` commits unconditionally (write-transaction law); `settle_existing_repos`
meta-guarded idempotent, clears the retry queue once; set reads (`get_dead_url_set`,
`get_settled_repo_set`) loaded once per batch; `failed_rows_truth(rows, github_has)` module
function resolves stale failed rows against the live vault.

## link_tracker.py (`gui/link_tracker.py:`)

- Manifest `<vault>/links_manifest.json`; **the vault (via VaultIndex) is the reconciliation
  ground truth** (v0.38.0).
- 5-phase per-link lifecycle: intake → `mark_processing` (**transient state NOT saved** — the
  v0.06 perf law) → mark_processed/recorded/failed/skipped/blocked → `verify` (full accounting:
  buckets github_processed/failed/skipped/pending, non_github_*, websites_*, blocked_recorded;
  `accounted+unaccounted+accounting_ok` must reconcile — "only X/Y accounted for — N escaped
  every bucket!" is a loud error; github notes verified as file-on-disk AND >100 bytes,
  websites >50; stale note_path healed via ground truth — a MOVE is not a loss).
- `get_all_clear` (486-523): only GITHUB rows gate the bot-queue mark-read; failed/processing/
  pending count as done when the vault holds the note.
- `get_reconciliation_urls`: heals pending rows whose note exists, re-points moved notes, tallies
  website/settled rows aside; persists the HEALED previous manifest without touching the live one.
- Atomic saves; dry-run recorded not written.

## vault_index.py (`gui/vault_index.py:`)

- `VaultIndex`: rebuilt per run; skip `_moc`, `_inbox`, `attachments`, `.obsidian`, `.trash`;
  **includes `_review`**; reads first **4000 chars** (for the `site_links:` line); `source:`
  normalized (configurable normalizer — websites pass `normalize_website_url`); site map
  (`site_note_for`) only for app-owned, non-`_review`, non-`failed` notes; first-come-wins.
- `find_obsidian_vaults()`: `%APPDATA%/obsidian/obsidian.json` registry + 11 common paths;
  `_safe_moc_name` Windows-illegal-char sanitizer.

## Delegates

`link_helpers.py` → `core.links` (3 one-liners); `dead_links.py` →
`DEAD_LINK_THRESHOLD 3`, `dead_link_threshold(config)` reads `notfound_strike_threshold`
clamped ≥2, errors → 3.

## Tests

DDL/behavior ports (processed/failed/decommissioned/settled flows; record_404 fail-closed;
settlement idempotence); manifest accounting law + reconciliation matrix; all-clear law;
VaultIndex skip/include sets + site-map gates + 4000-char read; threshold clamp.

# ADR-0003 — SQLite access strategy in Go (cache.db and siblings)

- **Status**: accepted
- **Date**: blueprint session (2026-10)

## Context

The Python app persists all state in **one SQLite file** `app/cache.db` (plus `sync_review.db` in the
same folder and the Telethon `session.session`):

- `cache.db` hosts tables owned by four independent stores (observed DDL):
  - `processed_repos`, `checkpoints` (dead), `failed_repos`, `decommissioned_repos`,
    `repos_settled`, `cache_meta` — `gui/cache_db.py:150-288` (`CacheDB`)
  - `note_state`, `corrections_log`, `note_state_dismissed` — `core/note_state.py:197-241`
  - `websites_processed`, `website_retry_queue`, `dismissed_urls`, `website_state_meta`,
    `websites_settled` — `core/website_state.py:94-132` (`WebsiteStateDB`)
  - `note_reads`, `embeddings`, `link_suggestions` — `core/note_reads.py:335-343`,
    `core/embeddings.py:273-283`, `core/linking.py:220-230`
- Every Python store opens **its own connection** with `check_same_thread=False`, `timeout=30`,
  `PRAGMA busy_timeout=30000`, and guards every statement with a per-instance `threading.RLock`
  (the "locking law"). `CacheDB` additionally sets `journal_mode=WAL`, `synchronous=NORMAL`
  (`gui/cache_db.py:106-147`). Note: `WebsiteStateDB`/`NoteStateDB` do **not** set WAL themselves but
  share the same file as `CacheDB` which does (WAL is a database-file-level persistent setting).
- Existing user data must keep working: the Go app will open the *same* `cache.db` created by the
  Python app (see `08-data-and-compatibility.md`).

## Problem

Choose a Go SQLite driver and concurrency model that preserves: WAL operation, cross-thread access,
30 s busy timeout semantics, and byte-compatible schema — with a simple Windows build (no CGO if
possible).

## Options considered

| Option | CGO | Notes |
|---|---|---|
| `github.com/mattn/go-sqlite3` | **Yes** (needs gcc/mingw on Windows for every build) | Battle-tested; CGO complicates the Windows toolchain and cross-compilation |
| `modernc.org/sqlite` (pure Go translation of SQLite) | No | Slightly slower (10–30% typical), fully self-contained builds |
| `crawshaw.io/sqlite` / `zombiezen.com/go/sqlite` | No (uses modernc) | Pool-oriented wrappers |

## Decision

1. **Driver: `modernc.org/sqlite`** via `database/sql` (`driverName "sqlite"`). Pure Go → trivial
   Windows builds, no toolchain beyond Go, static single exe. The app's SQLite workload is tiny
   (thousands of rows, dozens of statements per batch); the 10–30% driver overhead is irrelevant
   next to network/LLM latency.
2. **One connection per store, one `sync.RWMutex`-style guard per store**, mirroring the Python
   locking law: open connections with `?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)`
   (modernc driver pragma DSN parameters; exact syntax to be verified in implementation — see open
   questions in `14-risks-assumptions-open-questions.md`).
3. **Keep the exact table DDL** from the Python stores (schemas are pinned in the module specs).
   Add no migrations in Phase 1; the Go app **adopts** the existing database as-is (compatibility
   tests in `09-testing-and-parity.md` cover old-file opening).
4. Timestamps: Python writes naive-local `YYYY-MM-DDTHH:MM:SS` strings in most tables and UTC-aware
   ISO strings in `sync_deferrals`. **Go must write the same string shapes** (local-time formatting
   without offset for the legacy tables; see `05-python-to-go-mapping.md` §semantics).

## Consequences

- No CGO anywhere in the project (also required by ADR-0006's TLS-impersonation choice to keep the
  build simple).
- `checkpoints` table (dead schema, `gui/cache_db.py` creates but never uses) is *not* ported; opening
  an old DB simply leaves the table in place untouched.

## Evidence

- Locking law: `gui/cache_db.py:106-147` ("sqlite connections with check_same_thread=False are NOT
  safe for concurrent cursor use")
- WAL + busy_timeout: same file, and `core/sync_review.py:466-486` (DeferredLog explicitly sets WAL)
- Schemas: module specs `spec-gui-cache-db.md`, `spec-core-note-state.md`,
  `spec-core-website-state.md`, `spec-core-note-reads.md`, `spec-core-linking.md`

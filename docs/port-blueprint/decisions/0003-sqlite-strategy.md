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

## Amendments

### Am.1 — 2026-10-15 (SESSION-006, T-20): the declared-type conversion reality, the CAST read law, the verified DSN syntax, and the version pin

**Reality found while porting T-20** (the golden fixture test caught it pre-commit):

1. **The modernc driver converts declared-type TIMESTAMP reads.** `modernc.org/sqlite`
   (implemented at T-20 as **v1.36.1** — the newest version whose `go` directive, 1.21, keeps
   ADR-0002's `go 1.22` floor intact; v1.36.2+ require go ≥ 1.23, v1.41+ go ≥ 1.24) parses TEXT
   values from `DATE`/`DATETIME`/`TIMESTAMP`-declared columns into `time.Time` (a
   mattn/go-sqlite3-compatible heuristic; sqlite.go's parseTimeFormats eat every shape the app
   writes: `T`/space separators, optional fraction, optional offset). Through `database/sql` the
   value re-serializes into a DIFFERENT string shape — `"...+00:00"` becomes `"...Z"`, trailing
   fraction zeros drop (`".217680"` → `".21768"`). Python's `sqlite3` has NO declared-type
   conversion: these columns are plain `str`. Newer driver versions do NOT fix this —
   v1.60.1's `_texttotime` DSN parameter only guards UNDECLARED columns (and would bump the go
   floor to 1.26), so upgrading buys nothing here.
2. **THE CAST READ LAW**: every ported query that selects a timestamp-declared column wraps it
   in `CAST(col AS TEXT)` — an expression carries an empty declared type, so the raw bytes come
   back byte-identical to what Python wrote/read. CAST is applied ONLY to timestamp-declared
   columns (CAST on an INTEGER column stringifies the value). The law lives in
   `internal/store/doc.go`, is characterized by
   `internal/store.TestDriverDeclaredTimestampConversion`, and is enforced byte-exactly by the
   T-20 fixture golden (`internal/store/testdata/cache_fixture_content.golden.json`).
3. **DSN syntax verified** (doc 06 §1's `[verify syntax]` flag): the working form is
   `<path>?_pragma=busy_timeout(30000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)`
   — function-call pragma syntax as query parameters, applied per connection. Asserted by
   `TestDSNShape` + the pragma probes in `TestOpenPythonCreatedFixture` (journal_mode=wal,
   busy_timeout=30000, synchronous=1 — matching the reference's own open).
4. **Write-side stays string-typed**: Go writes timestamp columns as the `store.PyISOFormat*`
   helper strings (ADR-0003 decision 4 shapes) — the driver's `time.Time` parameter path
   (`formatTime` → `t.String()`, or `_time_format`) is never used. Cross-proof recorded in
   SESSION-006: Python's own `CacheDB` reads Go-written rows (raw bytes:
   `'2026-10-15T21:31:50.009129'` auto / `'2026-10-15T21:31:50'` seconds).

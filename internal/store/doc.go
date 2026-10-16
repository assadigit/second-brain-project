// Package store — the shared SQLite plumbing for every ledger (cache.db
// and siblings). New in Go (no single Python module maps here; it
// extracts the store law every Python store repeated), per ADR-0003 +
// doc 04 §3's store bullet; task T-20 (Phase 2). repos.go (task T-24)
// ports gui/cache_db.py's CacheDB into this package — the one store
// where THE CAST READ LAW is GENUINELY LIVE (its processed_at /
// last_checked / failed_at / decommissioned_at / settled_at columns are
// TIMESTAMP-declared; see repos.go's file doc).
//
// THE LOCKING LAW (ported from the reference's own store law,
// gui/cache_db.py:106-147 and core/sync_review.py:439-456):
//
//   - ONE connection per store instance (SetMaxOpenConns(1) — the exact
//     shape of Python's single sqlite3 connection with
//     check_same_thread=False), opened with the ADR-0003 DSN pragmas:
//     busy_timeout=30000, journal_mode=WAL, synchronous=NORMAL.
//   - A per-store guard (the embedded sync.RWMutex) for multi-statement
//     sequences, mirroring the reference's per-instance threading.RLock.
//   - close() is idempotent; call sites use defer (the v30 law).
//
// Divergences from the Python law, deliberate and documented:
//
//   - PRAGMAS UNIFIED: every Go store opens with WAL + synchronous=
//     NORMAL + busy_timeout via the DSN. In Python only CacheDB (and
//     sync_review's DeferredLog) set WAL; NoteStateDB / WebsiteStateDB /
//     note_reads relied on CacheDB having set it file-level first. The
//     Go side unifies (doc 04 §3's store bullet pins exactly these three
//     pragmas): a fresh DB created by any Go store is WAL from birth —
//     observable only for a file no CacheDB ever touched, which cannot
//     happen in the composed app (CacheDB opens at startup).
//   - RWMutex, not RLock: Go's mutex is NOT reentrant. Ported stores must
//     NOT nest acquisitions: public entry points take the guard, private
//     helpers run unlocked (the Go structural translation of Python's
//     recursive with self._lock blocks).
//   - Rows must be collected before nested statements on the same DB
//     (MaxOpenConns(1): an open *sql.Rows holds THE connection) — the Go
//     translation of Python's fetchall()-under-lock idiom.
//   - THE CAST READ LAW (ADR-0003 Am.1): the modernc driver (pinned
//     v1.36.1) converts TEXT values from DATE/DATETIME/TIMESTAMP-declared
//     columns into time.Time (a mattn-compatible declared-type heuristic
//     Python's sqlite3 never had), and that value re-serializes into a
//     DIFFERENT string shape ("+00:00" becomes "Z", trailing fraction
//     zeros drop). Python reads these columns as plain str. Therefore
//     EVERY ported query that selects a timestamp-declared column wraps
//     it in CAST(col AS TEXT) — an expression carries an empty declared
//     type, so the raw bytes come back byte-identical. CAST ONLY the
//     timestamp-declared columns: CAST on an INTEGER column stringifies
//     its value. Characterized by TestDriverDeclaredTimestampConversion;
//     the fixture test enforces the exact bytes against the golden.
//
// The dry-run seam stays at call sites exactly like Python
// (db_path=_dryrun.shadow_cache_path(...)): Open receives an
// already-shadowed path and needs no dryrun knowledge.
package store

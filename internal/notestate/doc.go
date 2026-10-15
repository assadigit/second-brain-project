// Package notestate ports core/note_state.py — persistent per-note state
// in cache.db (the §4.4 "Moves are corrections" engine).
//
// Spec: docs/port-blueprint/module-specs/spec-core-note-state.md;
// task T-21 (Phase 2).
//
// THE LOCKING LAW (store/doc.go, applied here): this DB embeds the shared
// store plumbing — ONE connection (MaxOpenConns(1)) opened with the
// ADR-0003 DSN pragmas, an embedded RWMutex that is NOT reentrant (public
// entry points acquire, private helpers run free), rows collected before
// any nested statement. close() is idempotent (inherited from store.DB).
//
// THE CAST READ LAW (ADR-0003 Am.1), checked and satisfied vacuously:
// the driver's declared-type conversion fires on DATE/DATETIME/TIMESTAMP/
// TIME-declared columns — the note-state schema declares EVERY column
// TEXT or INTEGER (see testdata/notestate_schema.golden.json, generated
// from the reference's own DDL), so reads return Python's plain-str bytes
// with no CAST. The schema-identity test pins this: changing any column
// to a TIMESTAMP declaration would change the DDL bytes and fail the
// test, forcing the CAST re-evaluation the law demands.
//
// THE DIVERGENCES from the Python law, deliberate and documented:
//
//   - Walk order: Python's os.walk yields filesystem order (unsorted);
//     the port walks lexically (filepath.WalkDir). The scan-order
//     effects (duplicate-path list order, which of two same-source
//     files wins) are a documented divergence class (module spec
//     "Migration-sensitive semantics"; differential cases pin winners).
//   - Unified pragmas: the Go open sets WAL + synchronous=NORMAL +
//     busy_timeout (store package law); the reference's NoteStateDB sets
//     only busy_timeout (journal_mode stays whatever the file had —
//     'delete' for a NoteStateDB-only file, WAL in the composed app
//     where CacheDB opened first). Observable only for a file no
//     CacheDB ever touched; the fixture golden records the reference's
//     own facts ('delete'/30000/2) and the fixture test asserts the
//     documented Go divergence instead (wal after a Go open).
//   - category/subcategory are *string (Python Optional[str]): nil ⇔
//     SQL NULL ⇔ "no category" (vault root / _review). The DB bytes
//     stay identical to Python's.
//   - Summary.MovedUnmapped is an always-present int (0); Python's
//     dict omits the key entirely in the dry-run early return —
//     unobservable through Go struct fields either way.
//   - Summary.Skipped carries the run_start_check no-DB marker
//     ('no note-state connection') that Python puts inside the applied
//     dict.
//
// The dry-run seam stays at call sites exactly like Python
// (db_path=_dryrun.shadow_cache_path(...)): New receives an
// already-shadowed path and needs no dryrun knowledge; the only
// dryrun-aware behavior is apply_corrections' rehearsal contract
// (record nothing, edit nothing).
package notestate

// Package notereads ports core/note_reads.py — v0.71.0 THE ONE-OPEN
// LAW: the scan's shared read cache.
//
// Spec: docs/port-blueprint/module-specs/spec-core-note-reads.md;
// task T-22 (Phase 2).
//
// One Scan click used to open every .md in the Websites vault up to
// FOUR times (the hygiene walk's bounded read, the banishment's
// frontmatter read, the banishment's body read, the inventory's head
// read). This package is the shared cache every one of those passes
// reads through:
//
//   - THE RAM TEXT LAYER — one bounded read per note per scan,
//     memoized by path + stat fingerprint (mtime_ns + size), served to
//     every consumer at whatever bound it asks for (a smaller bound
//     slices the cached text; a larger one re-reads once). A byte
//     budget keeps the layer honest on giant vaults (FIFO eviction,
//     never a crash).
//   - THE PARSED LAYER — the frontmatter (ONE unified parse: the union
//     of the keys the inventory door and the banishment door read) and
//     the body's inline banish tag, parsed ONCE per note and persisted
//     in the app's own cache.db keyed by the same stat fingerprint.
//     The NEXT scan of an unchanged note opens NOTHING.
//
// THE LOCKING LAW (store/doc.go, applied here): the cache's layers sit
// behind ONE mutex (the Go shape of Python's single RLock — NOT
// reentrant, so public entry points acquire and the private
// *Locked/tagOf helpers run free); the SQLite statements additionally
// ride the store plumbing — ONE connection (MaxOpenConns(1)) opened
// with the ADR-0003 DSN pragmas, the store's embedded RWMutex around
// every statement, rows collected before any nested statement (the
// prune pass). close is idempotent.
//
// THE CAST READ LAW (ADR-0003 Am.1), checked and satisfied vacuously:
// the note_reads schema declares every column TEXT, INTEGER or REAL
// (see testdata/notereads_schema.golden.json, generated from the
// reference's own DDL) — no DATE/DATETIME/TIMESTAMP declaration
// exists, so reads return Python's plain values with no CAST. The
// schema-identity test pins this: changing any column to a TIMESTAMP
// declaration would change the DDL bytes and fail the test, forcing
// the CAST re-evaluation the law demands. scanned_at is a REAL
// time.time() float (NOT a timestamp string) — ported as that shape
// (float64 seconds; typeof(scanned_at) == 'real' is asserted on both
// the Python-created fixture and Go-written rows).
//
// THE DIVERGENCES from the Python law, deliberate and documented (the
// spec's pre-classified fixes plus the store-package realities):
//
//   - THE STARS FIX (spec quirk 1): the reference's _unified_from_json
//     iterates only _FM_STRING_KEYS, so a DB-hit row's `stars` reads
//     ” even though the row carries it. The port restores the row's
//     own value (TestDBHitRowsCarryStars; the golden's
//     full_row_with_stars case is the quirk's evidence).
//   - THE BYTE BUDGET (spec quirk 2): the RAM text layer's budget is
//     accounted in BYTES (the reference counted characters — a
//     150,000-char text of é is 300,000 bytes). Internal only: the
//     observable eviction points differ on non-ASCII texts, pinned
//     deliberately in TestFIFOEvictionIsByteBudgeted against the
//     golden's character-accounted baseline.
//   - Unified pragmas: the Go open sets WAL + synchronous=NORMAL +
//     busy_timeout (store package law); the reference's NoteReadCache
//     sets only busy_timeout (journal_mode stays whatever the file
//     had — 'delete' for a lone NoteReadCache file). Observable only
//     for a file no other store ever touched; the fixture golden
//     records the reference's own facts ('delete'/30000/2) and the
//     fixture test asserts the documented Go divergence (wal).
//   - Hand-crafted rows outside the app's own JSON shapes self-heal
//     (decode failure reads as stale → one re-parse + re-store) where
//     the reference would str()-fold odd types or, for pathologically
//     mistyped banish/tags values, raise (the doors' try/except then
//     falls back to the direct parse). Unreachable for rows the app
//     itself writes — the reachable universe is byte-identical.
//   - readBounded reads at most 4n+7 bytes (provably the same first n
//     runes as Python's whole-file read(n)) — a bounded-memory
//     equivalence, not a behavior change.
//   - The DB-hit banish shape's key order is sorted (the reference
//     preserves encounter order); unobservable — hit rows are never
//     re-stored and dict comparisons are orderless.
//
// The dry-run seam stays at call sites exactly like Python (the scan
// job points the cache at the shadow cache.db under a rehearsal): New
// receives an already-shadowed path and needs no dryrun knowledge.
package notereads

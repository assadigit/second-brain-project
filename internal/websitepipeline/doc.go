// Package websitepipeline ports core/website_pipeline.py +
// core/website_state.py + core/website_directory.py — the Websites
// pipeline, its state ledger, and the directory-note regeneration.
//
// Spec: docs/port-blueprint/module-specs/spec-core-website-pipeline.md;
// tasks T-23 (state.go, Phase 2 — this file) and T-40/T-41 (Phase 4).
//
// THE LOCKING LAW (store/doc.go, applied here): this DB embeds the
// shared store plumbing — ONE connection (MaxOpenConns(1)) opened with
// the ADR-0003 DSN pragmas, an embedded RWMutex that is NOT reentrant
// (public entry points acquire, private helpers run free), rows
// collected before any nested statement. close() is idempotent
// (inherited from store.DB).
//
// THE CAST READ LAW (ADR-0003 Am.1), checked and satisfied vacuously:
// the driver's declared-type conversion fires on DATE/DATETIME/TIMESTAMP/
// TIME-declared columns — the website-state schema declares EVERY column
// TEXT or INTEGER (see testdata/websitestate_schema.golden.json,
// generated from the reference's own DDL), so reads return Python's
// plain-str bytes with no CAST. The schema-identity test pins this:
// changing any column to a TIMESTAMP declaration would change the DDL
// bytes and fail the test, forcing the CAST re-evaluation the law
// demands.
//
// THE DIVERGENCES from the Python law, deliberate and documented:
//
//   - Unified pragmas: the Go open sets WAL + synchronous=NORMAL +
//     busy_timeout (store package law); the reference's WebsiteStateDB
//     sets only busy_timeout (journal_mode stays whatever the file had —
//     'delete' for a WebsiteStateDB-only file, WAL in the composed app
//     where CacheDB opened first). Observable only for a file no
//     CacheDB ever touched; the fixture golden records the reference's
//     own facts ('delete'/30000/2) and the fixture test asserts the
//     documented Go divergence instead (wal after a Go open).
//   - settle_urls is NOT ported: the module spec records it as dead
//     code ("Dead: settle_urls (no callers)") — no production caller in
//     the reference; the one-spelling fold it exercised is pinned
//     through SettleExisting/SettleQueueHistory instead.
//   - dueRetries' Python datetime.fromisoformat accepts shapes Go's
//     time.Parse does not (any single-character date/time separator,
//     compact YYYYMMDDTHHMMSS). The port accepts the shapes the
//     reference family writes (naive seconds, space separator, optional
//     fraction, minute precision, date-only) plus the offset-carrying
//     and garbage classes the behavior golden pins; everything else
//     counts as due exactly like Python's parse-failure path. An
//     offset-carrying value counts as DUE in Python too (the naive-vs-
//     aware comparison raises TypeError, which due_retries catches and
//     treats as due) — pinned by the behavior golden's aware case.
//   - NormalizeLedgerKeys swallows statement errors and returns the
//     counts accumulated so far, exactly like the reference's blanket
//     try/except (rollback + partial report, never raises on the
//     caller's head).
//
// The dry-run seam stays at call sites exactly like Python
// (db_path=_dryrun.shadow_cache_path(...)): NewWebsiteStateDB receives
// an already-shadowed path and needs no dryrun knowledge.

// ============================================================================
// T-40 (SESSION-015) — the pipeline itself (pipeline*.go, notes.go,
// sitelinks.go, review.go, decommission.go, banish.go, master.go +
// the decomtable floor + the handdelivery/directorysync surfaces).
// The package laws above carry unchanged; these are T-40's own.
// ============================================================================
//
// THE LAYERING RECORD (the import-cycle the reference dissolves with
// Python's function-level imports): website_pipeline.py imports
// hand_delivery at module level while hand_delivery's stamps import
// website_pipeline's table helpers inside their functions. Go has no
// deferred imports, so the SHARED TABLE GRAMMAR (the parse + the
// status predicates + the canonical header) lives in the decomtable
// subpackage, below both:
//
//	decomtable      ← websitepipeline (the writers/scans stay here)
//	decomtable      ← handdelivery    (the stamps live there)
//	handdelivery    ← websitepipeline (the queue/consume calls)
//	directorysync   ← websitepipeline (the §11.1 banish-gate probe)
//
// WHAT T-40 PORTED vs WHAT THE LATER TASKS OWN:
//
//   - internal/handdelivery ships the pipeline's consumption surface
//     (the pure gestures, the queue read/write, Collect/Take, the
//     harvest stamp) — T-43 completes the module (the Chrome launch
//     family, the batch-level pickers, its own battery).
//   - internal/directorysync ships the banish-gate helpers (LoadBanish-
//     List / IsBanished / DefaultSharedHosts) — T-49 owns the rest.
//   - the v0.28.0 banned-domains DISK sweep (website_directory's
//     sweep_banned_notes) is a documented seam (sweepBannedNotesSeam)
//     until T-41 lands directorynote; the state-side purge (the DB
//     law) is enforced now.
//
// THE DIVERGENCES from the Python law, deliberate and documented
// (T-40's own list — T-23's list above carries unchanged):
//
//   - the third door's constructor ladder: the reference's
//     curl_cffi_available()/ensure_curl_cffi_installed() pair collapses
//     to "armed" — tls-client is ALWAYS compiled in (ADR-0006 §4's own
//     law), so the v0.47.0 auto-pip leg is obsolete by construction;
//     the "🤝 third door armed" line fires whenever impersonate_
//     fallback is on with a production fetcher.
//   - the unexpected-exception path: Python renders
//     f"{type(e).__name__}: {e}"; the Go catch-all renders the panic
//     value ("panic: …" never leaks — the recovery formats %v). The
//     pinned contract is the SHAPE (a failed result + the fallback
//     _review note), never the exception-class names.
//   - the auto-page re-verification (handdelivery's _auto_page_is_real
//     → chrome_tabs.page_is_real) degrades exactly as the reference's
//     own except-path defines ("a verdict that cannot be asked is not
//     a wall") until T-44 lands chrometabs.
//   - sanitize_tags' max_items parameter: notebuilder gained
//     SanitizeTagsMax (the reference's own API — the port had
//     hardcoded the default cap).
//   - Queue JSON: handdelivery's queue.json rides *storage.OrderedMap
//     metas (insertion-ordered) through storage.PyDumps — byte-parity
//     with json.dumps(indent=2, ensure_ascii=False); a 'consumed:
//     false' reset re-enqueue stays a boolean, exactly the wire shape.
//   - os.walk's pruned traversal (the banishment scans): filepath.
//     WalkDir with _inbox/dot-folder SkipDir — same deterministic
//     order (sorted dirs, files before subdirs at each level).
//
// THE STANDING LAWS, honored here: atomic saves through the house
// writer everywhere; the dry-run seam at every file mutation
// (dryrun.Remove/MkdirAll/WriteText); store-adjacent columns read
// only through the state layer (the CAST read law lives there);
// Python splitlines semantics via pySplitLines; text-mode bounded
// reads counted in DECODED characters (readHeadRunes); rune-wise
// truncations wherever Python slices codepoints.
package websitepipeline

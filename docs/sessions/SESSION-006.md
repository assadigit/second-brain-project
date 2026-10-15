# SESSION-006 — 2026-10-15 — Phase 2 opening (T-20) — the SQLite plumbing landed with the CAST read law; 234 tests race-clean; the environment was FULLY REBUILT again (second consecutive reset)

## Goal (from NEXT-SESSION.md v6 at start)

SESSION-006 = Phase 2 opening: T-20 first (`internal/store`: modernc DSN
pragmas — WAL, busy_timeout=30000, synchronous=NORMAL — per-store mutex,
naive/UTC timestamp helpers), then T-21…T-25 strictly in order as capacity
allows; tests first; goldens and Python-created cache.db fixtures generated
by running the unmodified reference; close cleanly at a natural boundary.

## What was done (task IDs, files, commits)

- **Environment restore (the SECOND full reset in a row)**: every artifact
  was gone at session start — both repo clones, the go toolchain, the
  venv (evidence: `ls` on all four recorded paths → No such file or
  directory; no `go` on PATH; only the Next.js scaffold snapshot in the
  mount). The owner's in-chat PAT re-authorized access; GitHub API
  identified the account (`assadigit`) and both repos. Cloned with the
  PAT passed EPHEMERALLY via `http.extraheader` (Basic
  `x-access-token:<PAT>` — the `token` scheme gets 401 on the git
  endpoint; verified with curl before cloning), so no remote URL, config
  file, or any artifact ever stored it. Work repo restored at
  `f1755c580357322abc603fb8ed8668dd1045a6ed` (SESSION-005 close;
  verified: `go vet` clean, **225/225 tests pass** on the rebuilt
  toolchain before any new work). GitCurator restored and pinned at
  `98f4fc71cfb862cee7225d864636c08062724348` — byte-exact match with
  `testdiff/PINNED-GITCURATOR-REV`. Toolchain rebuilt per
  `build/TOOLCHAIN.md`: go1.27.2 → `/home/z/.local/go` (downloaded from
  go.dev), venv `/home/z/venvs/gitcurator` (Python 3.12.15, stdlib-only).
  Symlinks recreated at both recorded paths.

- **T-20 `internal/store` (commit `0c65535`)** — M: the port's FIRST real
  dependency, `modernc.org/sqlite`, installed and **pinned at v1.36.1** —
  the newest version whose go directive (1.21) keeps ADR-0002's `go 1.22`
  floor intact (evidence: v1.36.2+ declare go 1.23.0, v1.41.0 declares
  1.24.0; the initially-fetched v1.60.1 had silently bumped go.mod to
  `go 1.26.0` — reverted, pinned compatible).
  - Generator `testdiff/goldens/dump_store.py` (the sixth): the
    timestamp-grammar corpus (fixed datetimes with INPUTS included — the
    auto-timespec quirk: fraction IFF microsecond != 0, zero-padded %06d;
    year 1 and 9999 edges; +00:00/+03:30/-05:00 offsets; byte-stable),
    THE PYTHON-CREATED FIXTURE `cache_python_created.db` (built via the
    reference's own CacheDB methods: add_processed incl. the INSERT OR
    REPLACE same-repo_id probe, update_last_checked, add_failed
    update-in-place + normalize_url keying + the
    resolved-history-never-touched law, record_404 counting,
    confirm_dead, settle_repos, set_meta; clean close → no sidecars)
    generated as an ATOMIC PAIR with its content golden
    (`fixture_sha256` ties them — regeneration rewrites both; the
    golden's timestamps are the reference's own now() strings), the
    schema golden (sqlite_master DDL + table_info), and the pragma facts
    the reference's own open sets (journal_mode=wal, busy_timeout=30000,
    synchronous=1). Stable goldens proven byte-identical across two
    consecutive runs. The generator stubs `gitcurator.gui.icons` for the
    PyQt-less stdlib venv — the reference's OWN CI pattern
    (tests/test_websitesqueuefix.py `_ensure_icons_importable`), app code
    unmodified.
  - Tests FIRST (9, `internal/store/store_test.go`): the grammar corpus;
    the Now-shape regexes + Local wall-clock self-consistency; **THE
    T-20 ACCEPTANCE** — open the Python-created fixture, WAL asserted,
    every table byte-exact vs the golden (Python's exact timestamp
    strings: microseconds for processed_at/failed_at/
    decommissioned_at, seconds for settled_at), Go writes survive
    close+reopen with the helper-shaped timestamps; **the DEFERRED
    sqlite-level shadow-cache assertions** (recorded in SESSION-004,
    landed here): port of test_phase0.py
    test_shadow_cache_reads_real_state_writes_are_discarded — the shadow
    sees real rows, writes land in the shadow only, the real cache stays
    untouched after disable; **THE SQLITE LAW concurrency port**
    (test_syncreview.py TestLedgerConcurrency shape): 4 handles × 20
    inserts on one file — busy_timeout + WAL give a clean wait, 80/80
    rows; close-idempotent (the v30 law); fresh-file creation; the DSN
    shape.
  - **THE GOLDEN-CAUGHT DIVERGENCE → ADR-0003 Am.1**: the modernc driver
    converts DATE/DATETIME/TIMESTAMP-declared TEXT reads into time.Time
    (mattn-compatible declared-type heuristic) and database/sql
    re-serializes them into DIFFERENT shapes — `"...+00:00"` becomes
    `"...Z"`, trailing fraction zeros drop (`".217680"` → `".21768"`).
    Python's sqlite3 has NO declared-type conversion (these columns are
    plain str). Investigation: newer drivers do NOT fix this — v1.60.1's
    `_texttotime` DSN parameter only guards UNDECLARED columns (and
    would bump the go floor to 1.26). **THE CAST READ LAW**: every
    ported query selecting a timestamp-declared column wraps it in
    `CAST(col AS TEXT)` — an expression carries an empty declared type,
    so the raw bytes come back byte-identical. CAST ONLY
    timestamp-declared columns (CAST on INTEGER stringifies: 111 →
    "111" — pinned by test). Characterized by
    TestDriverDeclaredTimestampConversion; enforced byte-exact by the
    fixture golden; documented in internal/store/doc.go.
  - **Cross-proof (bidirectional parity)**: Python's own CacheDB reads
    Go-written rows through a Go-opened, Go-written fixture copy —
    `is_repo_settled` finds the Go-settled URL, `get_meta` reads the
    Go-written value, `processed_row_for` returns the Go-written row,
    and raw SELECTs show the exact Python-shape bytes Go wrote
    (`'2026-10-15T21:31:50.009129'` auto / `'2026-10-15T21:31:50'`
    seconds).
  - Spec updates in-commit (reality wins): ADR-0003 Am.1 (the declared-
    type conversion reality, the CAST read law, the VERIFIED DSN syntax
    resolving doc 06's `[verify syntax]` flag, the v1.36.1 pin
    rationale); doc 06 §1 sqlite3 row updated; testdiff/README.md
    generator inventory records dump_store.py incl. the atomic-pair
    convention.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **Session goal — PASS**: Phase 2 opened, T-20 done properly with proof.
  - T-20 acceptance ("open a Python-created cache.db fixture; WAL mode
    asserted"): **PASS** — TestOpenPythonCreatedFixture green (WAL +
    busy_timeout=30000 + synchronous=1 asserted, all six tables
    byte-exact vs the generated golden, write probes durable across
    close+reopen); plus the cross-proof above.
- **Phase 2 exit gate — PARTIAL (in progress, 1 of 6 tasks)**: the
  moves-as-corrections suite (T-21) and the one-open law tests (T-22)
  are still ahead per the roadmap's strict order.
- **Full-suite evidence at close**: `go test -count=1 -race ./internal/...`
  → 10 packages ok, **234/234 tests PASS** (225 prior + 9 store: the
  seven Phase-1 packages unchanged + store), race detector clean;
  `go vet ./...` clean; `gofmt -l .` clean; `GOOS=windows go build ./...`
  clean; differential case `status-no-config` **PASS** (no CLI
  regression from the new dependency); generator determinism proven
  (stable goldens byte-identical across two runs).
- **Honest scope note**: T-21 (notestate, L) was assessed and NOT
  started — after the environment rebuild + T-20's unplanned
  driver-behavior investigation, remaining capacity could not complete
  an 842-line module + its 38-test characterization suite properly; the
  protocol's "one task done properly beats two half-done" applied.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/decisions/0003-sqlite-strategy.md` — Am.1
  (append-only): the modernc declared-type TIMESTAMP conversion reality;
  THE CAST READ LAW; the verified DSN pragma syntax; the v1.36.1 version
  pin (go-floor law); the write-side-stays-string-typed law. (Commit
  `0c65535`.)
- `docs/port-blueprint/06-dependency-migration.md` — the sqlite3 row's
  `[verify syntax]` flag resolved to the verified form + the CAST law +
  the pin. (Commit `0c65535`.)
- `testdiff/README.md` — dump_store.py recorded in the generator
  inventory, incl. the fixture/golden ATOMIC PAIR convention.
- `internal/store/doc.go` — the real package law replaces the T-00
  placeholder: the locking law, the unified-pragmas divergence, the
  RWMutex-not-reentrant rule, the collect-rows-before-nesting rule, and
  the CAST read law.

## Blockers & decisions needed

None blocking. Two honest records:

1. **The environment is now proven rebuildable twice over.** The
   SESSION-005 note ("treat the environment as rebuildable, not
   persistent") held: this session's start found a SECOND full wipe. The
   restore recipe (GitHub-API repo discovery via the in-chat PAT,
   ephemeral `http.extraheader` Basic auth, hash-level verification,
   toolchain rebuild per build/TOOLCHAIN.md) worked cleanly and is
   recorded here for the next session. Standing lesson unchanged: verify
   first with evidence; never store the PAT.
2. **Driver-behavior landmine catalogued, not improvised past**: the
   modernc declared-type conversion affects EVERY future store port
   (T-21 note_state has `corrections_log.corrected_at TIMESTAMP`; T-23
   website_state has settled_at/next_at TIMESTAMP; T-24 repos;
   sync_review's deferrals). The CAST read law (ADR-0003 Am.1) is the
   port-wide answer; T-21+ ports apply it to every timestamp-declared
   SELECT column.

## Owner actions requested

None required. Optional, standing: the Windows VM leg for T-70's
bootstrapper (non-blocking for Phases 2–5).

## Handoff

- Next session: **SESSION-007 = Phase 2 continuation — T-21
  (`internal/notestate`: scan_vault, classify/apply corrections,
  NoteStateDB; the moves-as-corrections suite; fingerprint parity with
  the recall block stripped; the NormcaseNT handoff)** per the roadmap's
  strict order.
- `NEXT-SESSION.md` rewritten: yes (v7 — the T-21 handoff).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies it
  verbatim.

# SESSION-015 — 2026-10-17 — Phase 4 OPENS: T-40 websitepipeline (the engine room) landed at 8098ac4 — 910 tests race-clean (+38); the state machine + the confirmation gate + THE DIFFERENTIAL GOLDEN; the tenth full environment wipe rebuilt before work

## Goal (from NEXT-SESSION.md at start)

SESSION-015 = Phase 4 OPENS: **T-40 `internal/websitepipeline`**
(XL) — the per-link state machine on top of T-23's state layer, per
spec-core-website-pipeline.md (READ THE SPEC BEFORE THE CODE). The
acceptance: state-machine table tests; banishment flows
(confirmed/declined/timeout); a differential golden-set case.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was fully wiped AGAIN (the tenth full reset)** —
  verified with evidence before anything else: the work repo, the
  GitCurator reference, the Go toolchain, and the venv were ALL
  absent. Rebuilt per docs/NEXT-SESSION.md §1's recipes: the remote
  discovered through the GitHub API with the in-chat PAT passed
  ephemerally as the Basic `x-access-token:<PAT>` header (never
  stored — no file, no env dump, no git config), second-brain-project
  re-cloned at `14e56f1` = origin/main (the SESSION-014 close), the
  reference checked out at
  `98f4fc71cfb862cee7225d864636c08062724348` (exact match with
  testdiff/PINNED-GITCURATOR-REV), symlinks recreated, go1.27.2
  reinstalled to /home/z/.local/go (the go.dev tarball), the
  stdlib-only Python 3.12.15 venv recreated. The go.mod pins verified
  (go 1.24.1 / sqlite v1.36.1 / tls-client v1.16.0 / x-net v0.48.0 /
  x-text v0.32.0 — all exact) and **the 872-test baseline re-verified
  green (race-clean)** on the rebuilt toolchain before any work.

- **T-40, commit `8098ac4` (38 new top-level tests; pushed the moment
  the battery went green)** — `internal/websitepipeline` + its
  dependency floors:
  - *The module-level machinery* (notes.go, sitelinks.go, review.go,
    decommission.go, banish.go, master.go, pipelineconst.go,
    pystrings.go): build_website_note (web-v2, the old
    core_offerings shape, the gist #snippet law, every cap ladder)
    + build_review_note; ONE NOTE PER SITE (the site_links parse/
    format/insert/replace/extend/footer/fallback writer,
    _note_lists_site_link, _read_note_placement); the _review scans
    (backlog with the graveyard + dismissed filters, the every-note
    scan, the reason reader, the half-fetched sweep); the master
    table's full grammar — the WRITERS (write_decommission_candidates
    fresh/append/dedupe/status-prefill, mark_urls_dead_in_table,
    _ensure_table_shape's trimmed-table restore, stamp_stored_rows,
    prune_decommission_table, refresh_master_table) and the SCANS
    (scan_decommission_table, scan_master_waiting_rows with the
    fetch/eyes classification + the settled ledger gate,
    scan_master_hand_rows, scan_master_redo_rows,
    note_is_properly_stored with the site-links ownership proof,
    harvest_hand_rows); THE BANISHMENT (the three doors — frontmatter
    tags flow+block, the body's inline tags with the stop-set grammar
    and fence/frontmatter skipping, the trash folder; the burial core
    _banish_url; banish_marked_notes with the per-door record rows +
    the skip law; banish_twins_in_other_vault; consume_decommission_
    table's passes 0-4 incl. the confirmation-gate hold and the
    v0.60.2 honesty law; the splits; settle_the_ledger with the
    healing-first order and the queue-history pass).
  - *The Pipeline class* (pipeline.go, pipeline_process.go,
    pipeline_run.go): the constructor's full gate sequence — the
    proxy/UA wraps (production-only), the CONFIRMATION GATE (the
    spoken count, the channel ask, every verdict's line), the
    graveyard consume riding the gate, the banishment pass + the
    twins sweep + THE TALLY (all three shapes), the blocked/self
    domain purges, the third-door + ladder arming lines (the
    always-compiled-in collapse, documented); process_link's state
    machine — the directory banish gate (the 60s-cached §11.1 probe),
    blocked/self domains, the dismissal doors (graveyard / banished /
    auto-verdict / plain, each with its named log line),
    already-in-vault vs upgrade vs the half-fetched hand exemption,
    the consolidation gate, the retry gate + the hand reborn, the
    fetch ladder with the failure taxonomy (no-heal auto-verdicts,
    the fourth-door hint, the archived rescue), classify/analyze with
    the corrective retries + the locked-row override + the site soft
    lock, the atomic writes, the per-link harvest, the four drivers
    (run with the settled gate + the hand-delivery pull + the
    duplicate report, run_due_retries with the honest splits,
    retry_review_backlog with the re-arm + the lost-row
    re-registration + the stale-placeholder cleanup,
    retry_master_waiting with the reborn + the honest verdict).
  - *The dependency floors* (the honest answer to the reference's
    function-level imports): `internal/websitepipeline/decomtable`
    (the table grammar floor — parse + the status predicates + the
    canonical header + TableCell; the layering record in its doc.go),
    `internal/handdelivery` (the consumption surface: the pure
    gestures, the queue read/write through *storage.OrderedMap +
    PyDumps byte-parity, SuggestedFilename, the README text,
    EnqueueHandDelivery with the canonical re-key heal,
    CollectDelivered, HandFetchResult, TakeHandDelivered with the
    redo-consume + the legacy-spelling heal, StatusIsHand,
    StampHandDeliveredRows — T-43 completes the module),
    `internal/directorysync/banishgate.go` (LoadBanishList /
    DefaultSharedHosts / IsBanished + normHost — T-49 completes the
    module).
  - *THE DIFFERENTIAL GOLDEN*:
    `internal/websitepipeline/testdata/websitepipeline_pure.golden.json`
    — 51 sections, 78,496 bytes, GENERATED by
    `testdiff/goldens/dump_websitepipeline.py` from the unmodified
    reference with every clock frozen (wp.datetime + ws.datetime
    rebound to a fixed subclass at 2026-10-17 12:00:00, hd.time
    stubbed — no per-run facts inside; the state DBs ride tempdirs
    and dump timestamp-free rows, the golden's baked tempdir paths
    normalized on both sides at compare time). The Go side replays
    every section and byte-compares through storage.PyDumps.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-40's task-list acceptance — PASS, all three stones**:
  1. *state-machine table tests* ✓ — pipeline_test.go: the gate table
     (fresh / directory-banish / blocked / self / the four dismissal
     doors / already-in-vault / upgrade / fetch-failure→review /
     no-heal→auto-verdict / walled→hint / archived→partial — every
     outcome one of the four kinds, every review/processed note
     verified on disk) + the detail battery (the destination log
     line, the retry row's attempts/last_error, the locked-row
     override, the consolidation, the settled gate, the duplicate
     report).
  2. *banishment flows (confirmed/declined/timeout)* ✓ —
     banishgate_test.go: confirmed deletes + reports + the loop
     closed (a re-paste never fetches), declined keeps + names the
     reason, timeout keeps + names the reason, no-channel defers,
     config opts back into auto, a broken channel defers safely, the
     bare-string verdict, the twins sweep (the human twin untouched —
     the sacred law), the same-vault guard, the gate holds ONLY the
     destructive door (the dead verdict still runs), the confirmed
     history row never re-asked, the tally dedupes across the doors,
     and the v0.60.2 law — a held run never claims removals.
  3. *a differential golden-set case* ✓ — the 51-section golden,
     byte-compared end-to-end (note builders, consolidation shapes,
     table grammar, status predicates, frontmatter/body parses, the
     scans, the hand queue incl. the consume/redo/heal wire shapes,
     note_is_properly_stored, the writers, the banishment flows
     confirmed + held, the twins, the splits, the settlement, the
     stamp/harvest/prune/refresh/sweep passes, the reason reader).
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 ./...` → **19 packages ok, 910/910 tests PASS, race
  detector clean, 0 failures** (was 872 at session start; +38 — the
  arithmetic closes exactly); go.mod/go.sum UNCHANGED (zero new
  dependencies — the authorized set untouched).
- **Honest notes**: (1) the SESSION-015 sandbox's own output renderer
  eats the two-character sequence `[h` (interpreted as a console
  escape) — a mid-session debugging detour chased a "corruption"
  that was never in the bytes; the code was correct all along and the
  detour cost only time (documented here so a future session doesn't
  repeat it: verify with %q dumps, never trust rendered text around
  bracket-h sequences); (2) the unexpected-exception path renders
  the panic value where Python renders `f"{type(e).__name__}: {e}"`
  — the pinned contract is the SHAPE (failed result + the fallback
  _review note), never the exception-class names (doc.go carries the
  record); (3) the auto-page re-verification (handdelivery's
  _auto_page_is_real) degrades exactly as the reference's own
  except-path defines until T-44 lands chrometabs; (4) the v0.28.0
  banned-domains DISK sweep is a documented no-op seam
  (sweepBannedNotesSeam) until T-41 lands directorynote — the
  DB-side purge (the never-collected law) is enforced now.

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-core-website-pipeline.md` —
  the Tests section carries the dated SESSION-015 note: the
  layering record (the decomtable floor — Go's no-deferred-imports
  answer to the reference's function-level import cycle), the
  dependency floors (handdelivery/directorysync surfaces shipped
  ahead of T-43/T-49), the frozen-clock golden contract, the
  clock-seam/SanitizeTagsMax/Delete additions to the neighboring
  packages, and the v0.28.0 disk-sweep seam. Reality-wins in the
  same commit as the port.
- `internal/websitepipeline/doc.go` — the package laws: T-23's full
  record preserved (moved verbatim from state.go's package comment,
  extended with T-40's sections — the layering, the divergences, the
  standing laws).
- `README.md` — the status paragraph (Phase 4 OPENED).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — fourth session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path). A one-line acknowledgement (or a
  rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The standing sandbox-reliability observation carries (tenth full
  wipe; push-after-every-task remained the load-bearing mitigation —
  T-40 was pushed the moment its battery went green).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-016 = Phase 4 continues — T-41
  `internal/websitepipeline/directorynote.go`** (M — the Website
  Directory regeneration) per spec-core-website-directory.md, and
  T-42 `internal/githubpipeline` + batch (L) if capacity remains.
- NEXT-SESSION.md rewritten: yes (v16).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.

# SESSION-010 — 2026-10-16 — Phase 2 CLOSE (T-23+T-24 re-ported, T-25 landed) — the phase is complete: 425 tests race-clean; SESSION-009's work was LOST (push-pending at wipe) and re-ported; the environment degraded MID-SESSION too (mode flips + toolchain wipe) and was repaired in place

## Goal (from NEXT-SESSION.md at start)

SESSION-010 = Phase 2 finish: T-25 (internal/batch/linktracker.go +
vaultindex.go) per the roadmap — but FIRST the prompt's own §1 demanded
verification, which found the environment fully wiped AND SESSION-009's
T-23/T-24 commits lost (never pushed before the wipe), so the session
became: rebuild, re-port T-23 + T-24, land T-25, close the phase.

## What was done (task IDs, files, commits)

- **The loss, proven at hash level**: origin/main sat at `df84157`
  (SESSION-008 close). The GitHub API answered **422** for both
  SESSION-009 commit SHAs (`39b4f58`, `2ddab80`) — no branch, no tag,
  no ref carried them; the local disk was wiped; therefore T-23/T-24
  were gone. The push-pending note in the lost NEXT-SESSION v9→v10
  rewrite proved prophetic. **The standing lesson this session
  institutionalized: PUSH AFTER EVERY TASK** — each of this session's
  three task commits was pushed the moment it landed.

- **Environment rebuild (the fifth full reset)**: work repo cloned via
  the owner's in-chat PAT passed ephemerally (Basic
  `x-access-token:<PAT>` header — never stored anywhere); GitCurator
  re-cloned and pinned at `98f4fc71cfb862cee7225d864636c08062724348`
  (exact match with `testdiff/PINNED-GITCURATOR-REV`); go1.27.2 →
  `/home/z/.local/go`; venv `/home/z/venvs/gitcurator` (Python
  3.12.15, stdlib-only); symlinks recreated; **the go.mod pin verified
  at blob level before any work** (`git rev-parse HEAD:go.mod` =
  `6721ebb8766573ec2209e6f536755822a905cee8`, the v1.36.1/go-1.22
  bytes); **the 314-test baseline re-verified green (race-clean) on
  the rebuilt toolchain before any new work.**

- **T-23 re-port, commit `0d02a15` (33 tests → 347)**:
  `internal/websitepipeline/state.go` — WebsiteStateDB (the five
  tables, adopt-as-is DDL), StateKey (THE ONE SPELLING at the ledger
  boundary), EnqueueRetry (attempts+1; next = now + 2×attempts days;
  first_failed_at survives; rune-truncated last_error), both
  settlements (SettleExisting with the honest unique-count quirk;
  SettleQueueHistory with the inserted-count honesty + the no-bulk-
  clear law), NormalizeLedgerKeys (the v0.64.0 healing pass — live
  note wins, more attempts wins, first by URL order; the Python
  error-swallow kept), purge/rearm/reset, DueRetriesAt (incl. the
  aware-datetime TypeError→due quirk and unparseable→due). Generator
  `testdiff/goldens/dump_websitestate.py` (the ninth): pure (the _key
  battery + the backoff delta table + the rune truncation) + the
  Python-created fixture ATOMIC PAIR + schema (CAST-law guard: all
  TEXT/INTEGER) + **behavior** goldens. Ported suites:
  test_phase2 TestWebsiteStateDB; test_onespelling
  TestTheStateBoundary + TestTheHealingPass; test_settledledger
  TestTheLedger (settle_urls is dead code per the module spec — the
  round-trip leg pins unsettle through SettleExisting);
  test_queuehistory TestTheSecondPass; test_masterretry
  TestResetRetryAttempts.

- **T-24 re-port, commit `05e67d9` (40 tests → 387)**:
  `internal/store/repos.go` — CacheDB (the retry queue's no-stacking /
  normalized-key / both-spellings laws; THE SETTLED REPOS LEDGER +
  settle_repos' unconditional commit; the ♻️ revive door through
  reset_dead_links; the 404 quarantine with **record_404 FAILING
  CLOSED**; the open-time migrations: fail_count column add +
  notfound_strikes merge-and-drop) + the dead_links threshold delegate
  + FailedRowsTruth (the pure truth-pass classifier). **THE CAST READ
  LAW IS GENUINELY LIVE HERE**: processed_at / last_checked /
  failed_at / decommissioned_at / settled_at are TIMESTAMP-declared
  (the T-20 schema golden pins it); GetDeadURLs and GetQuarantineStats
  select `CAST(decommissioned_at AS TEXT)` and the fixture test
  asserts the auto-timespec byte shape. Generator
  `testdiff/goldens/dump_cacherepos.py` (the tenth): pure
  (failed_rows_truth + the threshold clamp) + behavior (migrations
  with the legacy DDL recorded verbatim for byte-exact replay, the
  settlement, the revive door, the fail-closed probe — the reference's
  own answer: the threshold). The fixture acceptance REUSES T-20's
  Python-created cache fixture. Ported suites: test_quarantine
  QuarantineTests + MergedLineageTests; test_retrycry
  TestFailedRowsTruth + TestRetryQueueHonesty; test_reposettled
  TestTheLedger + TestTheReviveDoor; test_intakefix
  TestCacheDBMissingMethods.

- **T-25, commit `901644a` (38 tests → 425)**:
  `internal/batch/vaultindex.go` (VaultIndex: the skip sets with
  _review INCLUDED, the 4000-char read, the site_links grammar, the
  site-map registration gates — managed_by gitcurator only, failed
  excluded, _review excluded, first-come-wins; find_obsidian_vaults;
  SafeMocName) + `internal/batch/linktracker.go` (the 5-phase
  manifest: intake → the v0.06 transient-processing perf law →
  terminal marks → Verify with FULL ACCOUNTING — every bucket sums to
  the total or the loud "N escaped every bucket!" error — and the
  v0.38.0 ground-truth healing; GetAllClear (github rows only);
  GetReconciliationURLs with the v0.63.1 website-rows set-aside and
  the v0.63.3 settled-repos gate; **the Python json.dump(indent=2)
  byte-parity manifest writer** — insertion order, null fields,
  ensure_ascii escapes, surrogate pairs; atomic saves; the dry-run
  seam at the writer). Generator `testdiff/goldens/dump_linktracker.py`
  (the eleventh): manifest_bytes (the writer's byte-parity pin),
  SafeMocName, the env-steered obsidian discovery, and the behavior
  golden (skip sets, quote stripping, the 4000-char boundary, the
  site-map gates, add_url's incremental learn). Ported suites:
  test_reconfix TestReconciliation (the owner's 620-link scenario) +
  TestVerifyHealing + TestAllClearGroundTruth; test_retrycry
  TestReconciliationGithubOnly; test_reposettled
  TestTheReconciliationGate + TestGetAllClear (the settled probe
  through the DEFAULT-anchored CacheDB, redirected via
  GITCURATOR_HOME); test_onespelling TestTheVaultIndexKeys;
  test_lawfix TestTrashSkipLists (the vault_index leg); test_intakefix
  TestMissingRepoNote (the dedupe-through-the-note leg) + the _inbox
  skip.

- **THE MID-SESSION ENVIRONMENT EVENT (repaired in place)**: after the
  T-25 code was green, `git status` showed ~400 tracked files
  modified. Diagnosis at hash level: **content identical — only file
  modes flipped 100644→100755** (0 insertions, 0 deletions), and the
  same event **wiped /home/z/.local/go and /home/z/venvs** (the repos
  and .git survived). Repair: `git fsck` clean; modes restored from
  the index (`git checkout-index -f -a` — index-recorded executables
  preserved); toolchain + venv rebuilt; **the full battery re-run
  green on the restored tree before committing**. Recorded for the
  next session: the sandbox can degrade MID-session, not only at
  starts.

- **Generator-design lessons recorded for future authors**: (a) the
  one-spelling law bites probe authors too — raw UPDATE probes must
  target the CANONICAL key or they silently hit nothing (my first
  due_retries battery had this bug); (b) Python generators probing Qt
  modules need the icons stub (dump_store's pattern) and real
  signal-shaped loggers (`.emit`, not bare callbacks); (c) per-run
  paths must never leak into behavior goldens (basename-only facts);
  (d) rot-proofing: default-now probes need always-past / always-future
  / always-due rows, not dates near the generation instant.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **Phase 2 exit gate — PASS (all six tasks + every named item)**:
  1. *Opening a real Python-created cache.db fixture passes read/write
     probes* — FIVE Python-created fixtures now prove this family:
     T-20's cache_python_created.db (reused + extended by T-24's
     acceptance), T-21's notestate, T-22's notereads, T-23's
     websitestate — each opened by the port with every table
     byte-exact vs its generated golden and Go writes surviving
     close+reopen in the Python shapes.
  2. *Moves-as-corrections suite green* — T-21's 55-test suite
     (re-verified green in this session's 314-test baseline).
  3. *One-open law tests green* — T-22's suite + behavior golden
     (opens 4→0, db_hits 6; re-verified in the baseline).
  4. The task-list acceptances: T-23 backoff arithmetic (the pure
     golden's delta table: exactly 2.0 days per attempt, cap 3) +
     settle/heal passes (the behavior golden's full matrix);
     T-24 behavior tests + record_404 fail-closed (the reference's
     own closed-connection answer: the threshold); T-25 accounting
     tests (Verify's accounted/unaccounted/accounting_ok law asserted
     through the ported suites) + site-note registration gates (the
     behavior golden's site_notes: failed/_review/hand-written
     excluded, managed_by-only).
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l .`
  clean; `GOOS=windows go build ./...` clean; `go test -count=1 -race
  ./internal/...` → **14 packages ok, 425/425 tests PASS, race
  detector clean, 0 failures**; differential case `status-no-config`
  **PASS**; generator determinism proven for all three new generators
  (pure/behavior/schema goldens byte-identical across regenerations);
  the atomic-pair sha ties proven by the fixture tests.
- **Honest notes**: (1) T-23/T-24 are re-ports, not recoveries —
  SESSION-009's exact bytes are unrecoverable; the acceptances are
  met independently. (2) Verify's loud "N escaped every bucket!"
  branch is pinned through the accounting fields the ported suites
  assert (accounting_ok true paths); no Python suite crafts an
  unknown-status row to fire the error branch directly, and the port
  matches that coverage. (3) The mid-session mode-flip event cost
  capacity but no data.

## Spec updates made (file → what changed, why)

- `testdiff/README.md` — the generator inventory grew by three
  (dump_websitestate.py, dump_cacherepos.py, dump_linktracker.py —
  the ninth/tenth/eleventh), each with its golden shapes described.
- `internal/store/doc.go` — the package doc now records that repos.go
  is where the CAST read law is GENUINELY LIVE (the TIMESTAMP-declared
  columns + the two CAST-selecting methods).
- `internal/batch/doc.go` — the new package law doc (the manifest
  byte-shape law, the ground-truth law, the documented divergences:
  lexical walk order, strict-vs-replace reads, the sorted obsidian
  registry, the GITCURATOR_HOME test seam).
- No blueprint module-spec needed amending: spec-gui-cache-db.md and
  the website-pipeline spec's state section matched the pinned
  reference throughout (settle_urls' dead-code status, the
  unconditional commit, the migrations, the 4000-char read, the
  site-map gates — all already documented accurately).

## Blockers & decisions needed

None blocking. One standing observation for the owner: the sandbox's
reliability degraded further this session (a mid-session mode-flip +
partial wipe on top of the start-of-session full wipe). The
push-after-every-task discipline is now the load-bearing mitigation;
no action needed beyond keeping the PAT available in chat when
starting a session.

## Owner actions requested

None required. Optional, standing: the Windows VM leg for T-70's
bootstrapper (non-blocking for Phases 3–5).

## Handoff

- Next session: **SESSION-011 = Phase 3 "Network and LLM" opens** —
  T-30 first (`internal/llmclient`), per the roadmap's strict order.
- `NEXT-SESSION.md` rewritten: yes (v11 — the Phase-3 handoff).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies it
  verbatim.

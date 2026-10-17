# SESSION-017 — 2026-10-17 — Phase 4 continues: T-42 githubpipeline + batch landed at e92e4e7 — 984 tests race-clean (+46); THE GOLDEN (12 sections, GUI-layer import stubs) + the behavior battery; the notebuilder float reality-wins fix; the twelfth full environment wipe rebuilt before work

## Goal (from NEXT-SESSION.md at start)

SESSION-017 = Phase 4 continues: **T-42 `internal/githubpipeline` +
the GitHub side of `internal/batch`** (L) per spec-gui-processing.md —
the per-repo loop, the banner thread, the reports, the undo file.
Acceptance: batch summary goldens; 404-strike tests; rate-limit law
tests. T-43 handdelivery (L) if capacity remained — with the standing
instruction to close cleanly at a natural boundary.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was fully wiped AGAIN (the twelfth full reset)** —
  the work repo, the GitCurator reference, the Go toolchain, and the
  venv were all absent. Rebuilt per docs/NEXT-SESSION.md §1: the
  remote discovered through the GitHub API with the in-chat PAT
  (ephemerally as the Basic `x-access-token:<PAT>` header — never
  stored), second-brain-project re-cloned at `f95ce7e` = origin/main
  (the SESSION-016 close), the reference checked out at
  `98f4fc71cfb862cee7225d864636c08062724348` (exact match with
  testdiff/PINNED-GITCURATOR-REV), symlinks recreated, go1.27.2
  reinstalled (go.dev tarball), the stdlib-only Python 3.12.15 venv
  recreated. **A NEW hazard found and fixed this session: `git clone
  -c http.extraheader=…` PERSISTS the header into the new repo's
  .git/config — the PAT was sitting base64-decodable in BOTH clones'
  configs. Both were scrubbed (`git config --unset-all
  http.extraheader`) before the push; the recipe note below records
  it for every future rebuild.** The go.mod pins verified (go 1.24.1 /
  sqlite v1.36.1 / tls-client v1.16.0 / x-net v0.48.0 / x-text v0.32.0)
  and the 938-test baseline re-verified race-clean before work.

- **T-42, commit `e92e4e7` (46 new top-level tests; pushed the moment
  the battery went green)** — `internal/githubpipeline` (7 files) +
  `internal/batch/summary.go`:
  - *The GitHub API seam* (`ghclient.go`): the `RepoFetcher` interface
    (GetRateLimit / GetRepo / GetTopics / CountCommitsSince / GetReadme
    / GetLanguages / LatestRelease — the "fake repo" of the spec's test
    contract) + the production REST client (api.github.com over
    net/http, 15 s timeout, PyGithub's Accept header, standard TLS —
    the reference never threads netctx into the GitHub client, only
    into the banner), the GHError shape (status + str(e) — the
    403-with-"rate limit" substring law), the Link-header commit
    totalCount, the ordered-key languages decode, the base64
    UTF-8-ignore README decode.
  - *The per-repo loop* (`runner.go`): the gates in their exact order
    (non-GitHub → dead-set → dismissed → settled ledger → invalid →
    vault-index dedupe BEFORE any API call), the get_repo error paths
    (403+rate-limit: the 3605 s chunked wait + the THIS-repo retry
    incl. the post-wait 404; 401: the one-time token drop + the
    anonymous client swap; 404: the strike + the `_missing` note +
    the notfound_links.md append; other: the manifest failure — NO
    retry-queue row, the reference's own shape), the layer-2 SQLite
    dedupe with the self-heal, fork→parent, the enrichment, the
    banner goroutine joined 15 s, the v30 model re-read, the
    review/quality routing, the disk-full protocol (headless skip vs
    the GUI pause/resume), the bookkeeping chain, the delays.
  - *The meta floor* (`meta.go`): OrgReputation, CalculateCredibility
    (with Python's round(x,1) on the EXACT binary value),
    RecordMissingRepo (the idempotent `_missing/` placeholder +
    confirm_dead), BackfillMissingNotes.
  - *The banner* (`banner.go`): the v25 throttle (every 10→2 s /
    50→5 s), the 429 backoff ladder 5/10/15, the 24 h failed-marker
    cache, the >1000-byte + content-type gates, the atomic write,
    netctx's verify_ssl split (connectioncheck.OutboundTLSConfig).
  - *The LLM analysis* (`llmanalyze.go`): the 3-attempt ladder, the
    `<<<README_BEGIN>>>` injection guard, about_me.md context,
    models.analyze routing, the failure decision door
    (stop/retry/model/skip; the headless missing-model family +
    PickBestModel + ApplyModelChoice + LooksLikeConnectionError).
  - *The reports + undo* (`reports.go` + `undo.go`): the summary log
    (+ the summary_keep_last rotation), the final Markdown report,
    the batch snapshot + the new-files diff (os.walk's
    files-before-subdirs order, the substring root-skip quirk, the
    empty-snapshot falsiness quirk) + `_undo_last_batch.txt`, and the
    Finish completion chain with the aggregate lines.
  - *The scorecard* (`internal/batch/summary.go`): the v0.39.0
    `_build_batch_summary` — extend, never duplicate (T-25's
    LinkTracker + VaultIndex stay batch's).
  - *THE GOLDEN*: `testdata/githubpipeline.golden.json` (12 sections)
    GENERATED by `testdiff/goldens/dump_githubpipeline.py` from the
    unmodified reference @ 98f4fc7 with the reports clock frozen at
    2026-10-17 12:00:00 — **the reference's GUI-layer imports (PyQt6
    / PyGithub / ollama) are satisfied by minimal sys.modules stubs**
    (the pinned venv is stdlib-only; the reference code is unmodified;
    the reference's own mock.patch discipline — the star-import
    surfaces are empty, QThread/pyqtSignal are placeholders,
    GithubException carries the (status, data) shape). The reports'
    file formats BYTE-COMPARE; the prompt battery pins the injection
    guard + about_me + the ladder; the credibility / reputation /
    model-picker batteries replay.
  - *The behavior battery* (`runner_test.go` + `summary_test.go`): the
    gates (NO API call through every skip), the 404-strike flow (+
    idempotence), the rate-limit laws (the pause-until-reset, the
    continue-anyway warning, the %50 info line, the probe-error
    silence, the stop-during-wait), the 401 fallback, the cache dedupe
    + self-heal, the fork analysis, the review/quality routing, the
    disk-full trio, the delays, the stop door, the bookkeeping chain,
    the undo laws, the finish chain.

- **REALITY-WINS FIXES the task forced** (all with golden/behavior
  evidence, all in the same commit):
  1. **notebuilder's `cred_score` became a FLOAT**: the reference's
     `_calculate_credibility` returns `round(x, 1)` and every real note
     renders "credibility_score: 90.0/100" — the Go port's int field
     rendered "90/100" (a byte divergence on EVERY note; the corpus
     had only fed it ints, so the golden never caught it — T-42 is the
     module that computes the score). `NoteArgs.CredScore` → float64,
     the renderings go through PyFloatRepr, the notebuilder golden was
     REGENERATED with the production calling convention (float values,
     a 55.5 fraction pin, a 12500-stars comma-grouping pin).
  2. **The empty-attempt-2 comment lie**: llm.py's comment says an
     empty attempt 2 goes "straight to Attempt 3" but the raise
     escapes the except-handler into the OUTER failure door — the port
     follows the code (documented in doc.go).
  3. The primary-language tag merge keeps the reference's
     case-sensitive-membership quirk (the test reads the ORIGINAL
     spelling, the insert is the lowercase — "go" can land twice; the
     note's own sanitizer dedupes on the way into the frontmatter).
  4. The house JSON decoder's OrderedMap/json.Number/int64 shapes are
     handled at the ladder's result readers (the extract → the
     confidence coercion).

- **T-43 `internal/handdelivery` (L) — NOT STARTED, deliberately.**
  T-42's own scope grew (the notebuilder float fix + the golden
  regeneration + the GUI-stub generator machinery), and T-43 is a
  full L-sized task (core/hand_delivery.py:88-907). The one-task law
  applied: one task done properly with proof beats two half-done;
  closed cleanly at the natural boundary.

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-42's task-list acceptance — PASS**: *batch summary goldens;
  404-strike tests; rate-limit law tests* ✓ —
  `testdata/githubpipeline.golden.json` (12 sections) generated by
  `testdiff/goldens/dump_githubpipeline.py` against the pinned
  reference @ 98f4fc7 (the generator verifies the rev before running;
  frozen clocks; no per-run facts; the tempdir roots fold to `<TMP>`);
  the Go replay byte-compares the scorecard legs, the summary-log
  bytes + the rotation survivors under websites on/off, the
  final-report bytes (rich + minimal), the prompt battery, the
  ladder, and the three meta batteries; the behavior battery covers
  the 404-strike flow (the strike, the `_missing` note, the
  notfound-links line, the manifest skip, the idempotence) and the
  rate-limit laws (five legs incl. the stop-during-wait).
- **Full-suite evidence at close**: `go vet ./...` clean; `gofmt -l`
  clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 ./...` → **20 packages ok, 984/984 tests PASS, race
  detector clean, 0 failures** (was 938 at session start; +46 — the
  arithmetic closes exactly); go.mod/go.sum UNCHANGED (zero new
  dependencies — the authorized set untouched).
- **Honest notes**: (1) the golden was generated ONCE and never
  hand-edited — the divergences it exposed (the OrderedMap/int64
  number shapes, the reports-dir flow isolation) were fixed in the GO
  CODE and the GENERATOR respectively, never by weakening the golden;
  (2) the GUI worker SHELL (the mode wiring, the Ollama warmup family,
  the inputs fetchers, the interactive doors' GUI halves, the
  websites-phase wiring) is deliberately deferred to the GUI-shell task
  — the Runner exposes the seams (WebsitePhase, HygienePass,
  LLMDecisionFn, OnDiskFull, Finished) exactly as the spec's Go-mapping
  section demands, documented in doc.go + the spec's shipped record;
  (3) the clone-config PAT hazard was found and scrubbed in BOTH
  clones this session (see the recipe note below).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-gui-processing.md` — the Go
  mapping's Runner location amended (reality-wins: a batch-side Runner
  would import-cycle with T-25's LinkTracker — the Runner lives in
  githubpipeline, batch gains summary.go) + the dated SESSION-017
  shipped record (the files, the golden, the reality-wins fixes, the
  deferred shell).
- `docs/port-blueprint/traceability-matrix.md` — the D5 row widened to
  the whole module surface (the worker mixins + the loop + the
  scorecard).
- `README.md` — the status paragraph (T-42 + the float fix + 984
  race-clean + next: T-43).
- `internal/githubpipeline/doc.go` — the package record incl. the full
  divergence list (the comment lie, the non-object JSON door, the
  read-error degradation, the doc-probe reuse, the errno spelling, the
  int64/OrderedMap shapes).

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — sixth session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path). A one-line acknowledgement (or a
  rollback instruction) closes it.
- **The x/net/html keep-or-swap (standing, not a blocker)**: T-32
  shipped the hand-rolled tokenizer; nothing blocks on it.
- The standing sandbox-reliability observation carries (twelfth full
  wipe; push-after-every-task remained the load-bearing mitigation —
  T-42 was pushed the moment its battery went green).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session.

## Handoff

- Next session: **SESSION-018 = Phase 4 continues — T-43
  `internal/handdelivery`** (L) per spec-core-hand-delivery.md, then
  T-44 chrometabs if capacity remains.
- NEXT-SESSION.md rewritten: yes (v18).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.

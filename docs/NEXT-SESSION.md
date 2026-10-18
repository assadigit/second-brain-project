# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity (constants, links, storage+dryrun,
  prompts, taxonomy, notebuilder, config, helpers).
- **Phase 2 CLOSED (SESSION-006–010; all six tasks)**: T-20 `internal/
  store`, T-21 `internal/notestate`, T-22 `internal/notereads`, T-23
  `internal/websitepipeline/state.go`, T-24 `internal/store/repos.go`,
  T-25 `internal/batch` — 425 tests race-clean at the close.
- **Phase 3 CLOSED (SESSION-011–014; all five tasks)**: T-30 `internal/
  llmclient` (`f2073b5`), T-31 `internal/webfetch` (`4d93877`), T-32
  `internal/webextract` (`833b21b`), T-33 `internal/connectioncheck`
  (`e5846bd`), T-34 `internal/workerclient` (`6ffc4e9`) — 872 tests
  race-clean at the close.
- **Phase 4 CLOSED (SESSION-015–026)**: the task list T-40..T-50 ALL
  LANDED + the Phase-4 closer at `7aea86c` (the 30-link offline golden
  runner equivalent — the exit gate's all 10 differential stones green;
  1828 tests race-clean over 30 packages at the close).
- **T-60 LANDED ACROSS TWO SESSIONS (Phase 5)**: slice 1 — THE CLI
  SHELL at `143c590` (SESSION-027); slice 2 — THE COMMAND WIRINGS at
  `41d699e` (SESSION-028; the --init wizard with the repo radar,
  --test-connection over the ported battery, --detect-llm with the
  model menu, --hand-delivery complete, --chrome-retry's listing +
  dry-run rehearsal).
- **T-62 SLICE 1 LANDED (SESSION-029, `e2bc4cd`): THE RUN ENGINE'S
  DOORS** — the preflights (the PyGithub token check over the
  PreflightGithub seam, the TCP proxy reach over the NetDial seam,
  the provider branches with THE model-picker bridge wired — the
  not-pulled Ollama model opens the landed PromptModelMenu and
  persists through ApplyModelChoice; the GitHub-error continue-anyway
  door with the cron law), run_config_card (the boxed snapshot with
  the raw-comparison laws), the dry-run report writer
  (writeDryrunReport over the ported RenderReportMarkdown, the new
  dryrun.NowFn frozen-clock seam), and **THE PHASE-5 CLEAR RE-SPEC —
  ADR-0004 Am.3's bounded open item — RESOLVED ON EVIDENCE and
  implemented**: the all-clear verdict (GetAllClear gates the ✓/⏸️
  speech exactly like the reference's success+bot_source), the
  per-link ledger-state consumption, and the batch-complete mirror
  push as the ack (PhaseFiveClear + PushVaultMirrorAfterBatch over
  the new batch.VaultIndex.MirrorIndex() export + the cache's
  decommission events + the sha256 sorted-urls hash — the manager's
  own recipe — through workerclient.Sync.PushVaultMirror). The
  differential golden: **69,028 bytes, two-run byte-identical, 10 more
  sections** (the four cards, the github/proxy/llm preflight cases
  incl. TWO SAVED CONFIGS, the checks composition + abort, the
  dry-run report); the mutation teeth grew
  (TestGoldenCLIMutationPreflights). **1877 tests race-clean over 31
  packages (+11).**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-029 pushed the slice the moment the battery went
  green (`68e177c..e2bc4cd`). The recovery order for ANY mid-task death
  stays: (1) `git status` + `git log --oneline -5` + file timestamps
  to learn what ACTUALLY survived — never trust a handoff summary's
  claim over `git log`; (2) a preserved commit or draft is continued
  and audited (the full battery re-run), never blindly trusted, never
  restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **twenty-three full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth, SESSION-025's nineteenth, SESSION-026's twentieth,
  SESSION-027's twenty-first, SESSION-028's twenty-second,
  SESSION-029's twenty-third), plus SESSION-010's mid-session
  mode-flip, plus SESSION-021's CONTEXT death with the sandbox
  intact. The restore recipes: start-of-session (clone
  `assadigit/second-brain-project` with the PAT passed ephemerally as
  the Basic `x-access-token:<PAT>` header — locate the work repo via
  the GitHub API's authenticated /user/repos listing if the clone
  looks wrong (Currena is an UNRELATED sibling); never store the PAT —
  AND SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the PAT
  into the new repo's .git/config, so run `git config --unset-all
  http.extraheader` in BOTH clones right after cloning, then push
  with the ephemeral `-c` form only); GitCurator @ `98f4fc7` (verify
  against `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 →
  `/home/z/.local/go` (go.dev tarball, sha256 `ecadb99…`), venv
  `/home/z/venvs/gitcurator` (Python 3.12, stdlib-only), symlinks
  `/home/z/second-brain-project` and `/home/z/GitCurator`;
  mid-session mode-flip (`git checkout-index -f -a` restores modes
  from the index; `git fsck` to sanity-check). Verify the pins at
  every start: `git rev-parse HEAD:go.mod` — expect the v1.36.1/
  tls-client/1.24.1 bytes (do NOT bump sqlite; do NOT bump
  tls-client past v1.16.0). **Do NOT run `go mod download all`** (it
  churns go.sum; plain `go mod download` + `go build`/`go test` work
  without it). **Detached background processes DO NOT SURVIVE
  tool-call boundaries — run long suites in FOREGROUND CHUNKS per
  tool call.** **The renderer-escape + cd-discipline laws stay: the
  two-character pairs [h and [m are eaten on the OUTPUT path, the
  INPUT path, AND THE BASH-ARG path — write commit messages to a FILE
  and commit with -F, then verify the committed bytes with 'git log
  --format=%B' before pushing; note the DISPLAY of test diffs eats the
  pairs too (the golden replay compares JSON-decoded BYTES — display
  mangling never enters a comparison). AND the cd-discipline law: the
  shell RESETS its working directory between tool calls — EVERY git/go
  command carries its own cd; never commit outside
  /home/z/my-project/second-brain-project.** **The Write/Edit tool
  path REDACTS token-shaped strings at file-write time — never put a
  token-shaped string in a file you write, even a fake one (craft
  safe-shaped test tokens like `fake-github-token-…`; secret-scan the
  tree with a byte-level grep before every commit).**
  **SESSION-028's golden-side recipes stay: the env record's configs
  are sort_keys-ordered — craft any saved-order-sensitive config from
  an ALPHABETICAL literal; the scripted stdin folds — unfold before
  feeding; the dry-run switch must be RESTORED after a rehearsal.**
  **NEW (SESSION-029's, found by the differential itself): THE
  FIXED-LENGTH ROOTS LAW — a section whose output PADS a path to a
  fixed width (the run-config card's right-side padding) bakes the
  path LENGTH into the golden's bytes; the replay's buildEnv now uses
  roots whose lengths equal the generator's mkdtemp roots
  (`/tmp/gc-cli-app-goreplay` + `/tmp/gc-cli-vaults-goreplay` —
  "goreplay" = exactly the 8-rune suffix width tempfile.mkdtemp
  generates; keep that width when touching buildEnv); and the
  SHALLOW-COPY SNAPSHOT law — a golden section recording a config
  that the driven call MUTATES must copy.deepcopy BEFORE the call
  (apply_model_choice writes through the shared nested dicts).**

## 2. Next session: SESSION-030 — T-62 slice 2 — the engine + the drivers

- **T-62 (L) slice 2** — per spec-cli.md's cli_run.py section + the
  SESSION-029 shipped record (READ THE SPEC + THE LANDED SLICES BEFORE
  THE CODE — phase5.go's PhaseFiveClear/PushVaultMirrorAfterBatch and
  preflight.go are READY to be called by the engine's finish):
  1. **`run_batch_visual`** (cli_run.py 239-504): the synchronous
     Go shell over the ported engines (no Qt): the setup family
     (pipelines switches → the mode dispatch (direct/import) → the
     LinkTracker + intake + the banned/self drops → the cache open
     (the dryrun.ShadowCachePath in a rehearsal) → note-state → the
     dead set → the settled ledger → the repo-banish pass
     (repobanish.RunRepoBanishmentPass) → the vault index rebuild +
     `SetVaultIndex` → the backfill → the note-state start-of-run
     (notestate's run_start_check — CHECK whether it needs porting
     first; surface with evidence if it does) → the dismissed set →
     the undo snapshot); the warmup family (the Ollama list/timeout +
     autostart + warmup + the auto-switch/stand-in ladder with the
     model_prompt_callback bridge; the llamacpp re-probe; the cloud
     /models pre-flight); the GitHub loop over the ported
     `githubpipeline.Runner`; the websites phase over the ported
     `websitepipeline.Pipeline` (website_phase.py's wiring: the state
     (shadow in dry-run), settle_the_ledger, the websites VaultIndex,
     the llm router, the banish-confirm channel, split_dismissed,
     due_retries, the bar re-base, the backlog modes, the manifest
     marking, the directory rebuild, the inbox prune); the close +
     the ported Finish chain; the finish seals (vaultseal
     SealFromConfig → goodrepos PublishFromConfig → WebsitesSealFrom
     Config → ManualSealFromConfig — the reference's exact order, best
     effort); **the Phase-5 CLEAR call** (the landed PhaseFiveClear —
     the mirror push + the verdicts); the dry-run path (the switch
     flip/restore + the entry-count line + writeDryrunReport); the
     SIGINT handler (signal.Notify → runner.Stop) + the heartbeat (a
     100 ms ticker driving StatusLine.Tick — T-61's spinner under a
     real loop) + the exit-code law (0 iff the success flag).
  2. **The batch drivers**: cmd_auto (the Worker-queue fetch over
     `workerclient.GetPendingResult` + the `_bot_queue_job`
     CLASSIFICATION half — pending/in_vault/decommissioned/
     non_github/vault_index_count/pending_website_urls… — port it as
     its own coherent piece; the queue counts print; the dedup ground
     truth line; **the wrong-vault guard** (1374-1409, interactive
     y/N unless --yes, the non-interactive sync variant); the >10
     confirm (TTY-only, skipped by --yes); the run_batch_visual
     handoff with bot_source=True); cmd_sync's live beats (1b: the
     battery over the ported connectioncheck; 2: the landed button;
     3: the fetch + the non-interactive wrong-vault guard + the
     run); cmd_retry_failed (get_failed_urls → the run); the
     import-file mode (_fetch_from_import — the one-link grammar);
     **the chrome-retry live leg** (preflight → the real
     DeliverPagesViaChrome → run_batch_visual with non_github in the
     same run — the pieces are all landed now).
  3. If capacity runs short: land the engine shell properly with its
     own differential proof first (the drivers follow), record the
     remainder honestly.
- **The differential stone grows with the slice**: extend
  testdiff/goldens/dump_cli.py (the engine's seams the T-42/T-44/T-50
  goldens already proved; the batch dry-run report over the recorded
  entries; the wrong-vault guard + the >10 confirm over scripted
  stdin) and goldencli_test.go replays them — the two-run
  byte-identical proof + the mutation teeth stay part of every
  regeneration. MIND THE NEW LAWS: fixed-length roots, the
  deep-copy-before-mutation snapshot, the alphabetical literal for
  saved-order-sensitive configs, the unfold-before-feeding.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1);
  `store.PyISOFormat*` for timestamp writes; the dry-run seam at call
  sites; rows collected before nested statements; tests never
  weakened; reality wins over specs (fix the spec in the same commit);
  goldens are GENERATED by running the unmodified Python app, never
  hand-typed; no per-run facts in goldens; the frozen-clock seams
  (`clockNow` per package; handdelivery.Clock; dryrun.NowFn — the new
  one). T-41's: text-mode reads translate universal newlines before
  any cap or parse; invalid UTF-8 decodes with the maximal-subpart
  U+FFFD law; vault walks are os.walk's files-before-subdirs order.
  T-42's: the house JSON decoder's shapes; floats render through
  helpers.PyFloatRepr. T-43's: `pyBytesBlank` is the blank-body law;
  Go MAP ranges are RANDOM. T-44's: Go's `url.Parse` is NOT CPython's
  urlsplit; `cmd.Wait()` closes the pipe; scripted timing tests keep
  the reference's UNPATCHED cadences. T-45's: THE POINTER-THREADING
  LAW ([]*NoteRow); keep BOTH frontmatter parsers; Python text-mode
  read(n) counts CHARACTERS; the _windows_eyes law rides a stat seam.
  T-46's: the WALK-ORDER FOLD; bounded reads count runes; None doors
  type-narrow to the zero value; a revive-style table iteration rides
  sorted keys; dry-run asymmetries preserved. T-47's: THE REDACTION
  LAW; the token NEVER touches .git/config; FETCH_HEAD rewritten after
  every token-URL fetch; frozen GIT DATES make commit shas
  deterministic; the reference's mock-surface maps to instance seams.
  T-48's: the apiFn/tokenURLFn instance seams; the ReadDir-order
  convergence; atexit maps to an exported Cleanup; the base-dict
  freeze quirk; the frozen-dates sha trick. T-49's: the []-vs-null
  byte law; `storage.PyDumpsUTF8`; dictDoor presence checks + the
  double-wrapping trap; the silent unknown-type failure; the
  set-iteration convergence; the push_mirror env-inheritance law; the
  CPython re vs Go regexp convergence; the lstrip character-set
  quirk; the raw-tree banished round-trip; the importlib
  fixture-import generator trick. T-50's: the map-order law BITES THE
  LOGS; the []-vs-null law extends to .get() dumps; the house JSON
  decoder returns int64 for integral literals; str.splitlines'
  boundary set is wider than \n (pySplitLines is the twin);
  walk-order independence is earned by sorting; the clockNow seam
  covers every time.strftime write; the reference's exception-text
  rendering stays unpinnable (the SHAPE is the contract).
  SESSION-026's: the offline-golden-runner's quirks are Go notes; the
  generator that drives a reference TOOL imports it via importlib and
  stubs its cleanup; long suites run in foreground chunks.
  SESSION-027's: the gui.app generator door stubs to the REAL
  underlying classes; APP_DIR binds at cache_db import — patch
  constants BEFORE the import; the dead-level-color quirk; the
  optional-int flags resolve present-vs-absent AFTER parse; the
  Write-tool token redaction + the safe-shaped crafted-token law.
  SESSION-028's: the generator's fake PyQt6/PyGithub modules satisfy
  worker_jobs/WorkerLlmMixin's import guards so the reference's own
  picker runs; the scripted stdin folds — unfold before feeding; the
  env record's configs are sort_keys-ordered — craft any
  saved-order-sensitive config from an ALPHABETICAL literal; the
  battery's net doors inject through the exported connectioncheck
  seams; the dry-run switch must be RESTORED after a rehearsal; the
  reference's chrome-retry --dry-run never flips the switch — the
  port's fix is the recorded reality-wins. **SESSION-029's: the
  FIXED-LENGTH ROOTS law (padded output bakes path lengths — keep
  buildEnv's 8-rune-width roots); the deep-copy-before-mutation
  snapshot (a section's recorded config must precede any
  through-the-nested-dicts mutation); the deterministic fake clock
  (1000.0 → 1000.042 renders 42 ms through int((t1-t0)*1000)); the
  exception class-name tokens ride the seams (type("TimeoutError", …)
  — the SHAPE is the contract); a golden section capturing a RETURN
  VALUE cannot use the stdout-capture helper (the _cap0 lesson).**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-029.md` — the most recent report (the doors'
   evidence, the golden's growth, the Phase-5 re-spec's evidence
   chain, the two new house laws, the twenty-third-wipe rebuild).
2. `docs/port-blueprint/module-specs/spec-cli.md` — THE SPEC (the
   banner + the SESSION-027/028/029 shipped records + the cli_run.py
   section = T-62's contract + the Tests section = the acceptance).
3. `docs/port-blueprint/module-specs/spec-gui-processing.md` — the
   worker-shell spec (the _run_impl order IS the contract; the
   deliberately-deferred list names exactly what slice 2 builds).
4. The reference's run-engine sources (in order):
   `gitcurator/cli/cli_run.py` 239-504 (run_batch_visual — the
   signals' wiring 318-341, the finish seals 383-417, the Phase-5
   block 425-469, the SIGINT law 477-494);
   `gitcurator/gui/processing_worker.py` (`run` 199-237, `_run_impl`
   239-1889 — the setup, the warmup family 470-748, the post-loop
   1700-1889); `gitcurator/gui/worker/website_phase.py` in full;
   `gitcurator/gui/worker/inputs.py` (`_fetch_from_import`);
   `gitcurator/gui/worker_jobs.py` (`_bot_queue_job`'s classification
   half — the Option-B fetch's shape); `gitcurator/cli/cli.py`
   cmd_auto (1252-1440) + cmd_retry_failed + cmd_manual_batch +
   cmd_chrome_retry's live leg (1204-1245);
   `gitcurator/cli/cli_sync.py`'s live beats (161-381).
5. The landed slices (the patterns to extend): `internal/cli/*.go`
   (phase5.go's PhaseFiveClear/PushVaultMirrorAfterBatch +
   preflight.go's PreflightChecks are READY for the engine's finish;
   the seams' inventory), `testdiff/goldens/dump_cli.py` (the
   generator's seam inventory + the SESSION-029 laws in the
   sections), `internal/cli/goldencli_test.go` (the replay harness +
   the mutation teeth), and the engine surfaces:
   `internal/githubpipeline/runner.go` (the Runner + RunGitHubLoop +
   the WebsitePhase/ModelPromptFn seams), `internal/githubpipeline/
   undo.go` (the ported Finish chain), `internal/websitepipeline/`
   (the Pipeline + the four drivers), `internal/batch/linktracker.go`,
   `internal/workerclient/` (GetPendingResult + PushVaultMirror +
   helpers.go's builders), `internal/dryrun/` (the NowFn seam).

## 4. Known constraints / owner actions for SESSION-030

- **The go-directive acknowledgement (STILL WAITING — nineteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers
  x/net/html, approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB the persisted
  extraheader from .git/config after the clone — see §1); goldens are
  GENERATED by running its unmodified code (`PYTHONDONTWRITEBYTECODE=1`
  keeps it pristine).
- **Dependency law**: `modernc.org/sqlite` pinned at v1.36.1;
  `bogdanfinn/tls-client` pinned at v1.16.0; `golang.org/x/net`
  v0.48.0 + `golang.org/x/text` v0.32.0 (direct); ANY other add needs
  a new ADR — surface, don't improvise. (SESSION-029 added zero —
  sha256 + net ride stdlib.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections
  (ScanMasterRedoRows is the top item); the note_state SURGERY path's
  CRLF round-trip differential (Phase 4+); the Windows live legs
  (Phase 7 clean-machine checklist); a Windows VM without WebView2
  for T-70 (optional).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit — the Write-tool redaction is a safety NET, not
  a license); tests never weakened; reality wins over specs (fix the
  spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-030)

```
You are starting SESSION-030 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-three full wipes across
   SESSION-005..029 plus mid-session degradations; SESSION-021 added the
   CONTEXT-death class — trust git evidence over any lossy handoff summary:
   git status + git log --oneline -5 + file timestamps tell what ACTUALLY
   survived; a preserved commit or draft is continued and audited with the full
   battery, never blindly trusted, never restarted from zero). FIRST verify with
   evidence: work repo at /home/z/my-project/second-brain-project (symlinked
   /home/z/second-brain-project), read-only reference /home/z/my-project/GitCurator
   @ 98f4fc7 (symlinked /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 → /home/z/.local/go (go.dev tarball, sha256 ecbadb99…), venv
   /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, and
   go.mod with modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not
   bump any of these). Then check `git status` and `git log --oneline -3`:
   SESSION-029 landed T-62 slice 1 — THE RUN ENGINE'S DOORS — at e2bc4cd on top
   of SESSION-028's command wirings (the tree must be clean and origin/main at
   e2bc4cd; the baseline is 1877 tests race-clean over 31 packages — re-verify
   the suite before work if anything looks off). If anything is missing, rebuild
   per docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass
   it ephemerally with git -c http.extraheader="Authorization: Basic $(printf
   'x-access-token:%s' "$PAT" | base64)", the Basic scheme; never store it — AND
   SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the header into the new
   repo's .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; locate the work repo via the GitHub API's
   authenticated /user/repos listing when unsure — Currena is an UNRELATED
   sibling; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`). SESSION-029's operational finds (newest laws): the
   FIXED-LENGTH ROOTS law — a golden section whose output PADS a path to a fixed
   width bakes the path LENGTH into the golden's bytes, so the replay's buildEnv
   uses roots exactly as wide as the generator's mkdtemp roots ("/tmp/gc-cli-app-
   goreplay" + "/tmp/gc-cli-vaults-goreplay" — "goreplay" is exactly the 8-rune
   mkdtemp suffix width; keep it); the SHALLOW-COPY SNAPSHOT law — a golden
   section recording a config that the driven call MUTATES must copy.deepcopy
   BEFORE the call (apply_model_choice writes through shared nested dicts); the
   deterministic fake clock arithmetic (1000.0 → 1000.042 renders 42 ms through
   int((t1-t0)*1000)); exception class-name tokens ride the seams via
   type("TimeoutError", …) — the SHAPE is the contract; and a golden section
   capturing a RETURN VALUE cannot use the stdout-capture helper. The older
   laws stand: (a) THE CD-DISCIPLINE LAW — the shell resets its working
   directory between tool calls, so every git/go command carries its own cd (or
   go's -C flag); never commit outside /home/z/my-project/second-brain-project;
   (b) the renderer-eats-escape-pairs lesson — the two-character pairs [h and [m
   are eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path: avoid
   those spellings, write commit messages to a FILE and commit with -F, then
   verify the committed bytes with 'git log --format=%B' before pushing, and
   never trust a DISPLAY of escape-bearing bytes (grep the FILE bytes); (c) the
   Write/Edit tool REDACTS token-shaped strings at file-write time (craft
   safe-shaped test tokens, never ghp_-shaped ones; byte-level secret-scan
   before every commit); (d) detached background processes DO NOT SURVIVE
   tool-call boundaries — run long suites in FOREGROUND CHUNKS per tool call.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (T-62 slice 1 LANDED at e2bc4cd — the
   doors' record is in the status) → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records SESSION-029's
   close: the doors at e2bc4cd, 1877 tests race-clean over 31 packages, the
   environment recipes, the newest laws incl. the fixed-length-roots and
   snapshot ones) → docs/sessions/SESSION-029.md (the most recent report —
   the doors' evidence + the Phase-5 re-spec's evidence chain + the golden's
   growth).

3. THIS SESSION IS T-62 SLICE 2 — THE ENGINE + THE DRIVERS: per
   docs/port-blueprint/module-specs/spec-cli.md's cli_run.py section +
   spec-gui-processing.md's worker-shell section + the SESSION-029 shipped
   record (READ THE SPECS + THE LANDED SLICES BEFORE THE CODE — phase5.go's
   PhaseFiveClear/PushVaultMirrorAfterBatch and preflight.go's
   PreflightChecks are LANDED and READY for the engine to call):
   run_batch_visual's synchronous shell over the ported engines (the setup
   family: pipelines switches, the mode dispatch, the LinkTracker intake +
   the banned/self drops, the cache open with the dry-run shadow, note-state,
   the dead set, the settled ledger, the repo-banish pass, the vault index
   rebuild + SetVaultIndex, the backfill, the note-state start-of-run —
   check whether notestate's run_start_check needs porting and surface with
   evidence if it does, the dismissed set, the undo snapshot; the warmup
   family: the Ollama list/timeout + autostart + the auto-switch/stand-in
   ladder over the model_prompt_callback bridge, the llamacpp re-probe, the
   cloud /models pre-flight; the GitHub loop over githubpipeline.Runner; the
   websites phase over websitepipeline.Pipeline per website_phase.py's
   wiring; the close + the ported Finish chain; the finish seals in the
   reference's exact order — vaultseal.SealFromConfig → goodrepos.
   PublishFromConfig → WebsitesSealFromConfig → ManualSealFromConfig; the
   Phase-5 CLEAR call; the dry-run path — the switch flip/restore + the
   entry-count line + writeDryrunReport; the SIGINT handler + the 100 ms
   heartbeat ticker driving StatusLine.Tick + the exit-code law); then the
   batch drivers (cmd_auto over the Worker GetPendingResult fetch + the
   _bot_queue_job classification half ported as its own coherent piece +
   the queue counts print + the dedup ground-truth line + the wrong-vault
   guard (1374-1409) + the >10 confirm; cmd_sync's live beats 1b/2/3;
   cmd_retry_failed; the import-file mode; the chrome-retry LIVE leg —
   preflight → the real Chrome delivery → the batch processing in the same
   run). T-61's spinner under a real loop rides along (the heartbeat
   satisfies it). Extend the differential stone with the slice:
   testdiff/goldens/dump_cli.py grows the new sections (the batch dry-run
   report over the recorded entries; the wrong-vault guard + the >10 confirm
   over scripted stdin; whatever engine seams the T-42/T-44/T-50 goldens
   already proved) and internal/cli/goldencli_test.go replays them — the
   two-run byte-identical proof + the mutation teeth stay part of every
   regeneration; MIND THE NEW LAWS (fixed-length roots, deep-copy-before-
   mutation snapshots, alphabetical literals for saved-order-sensitive
   configs, unfold-before-feeding). If the task will not fit the session's
   capacity, land the engine shell properly with its own differential proof
   first and record the remainder honestly — one slice done properly with
   proof beats a rushed whole; close cleanly at a natural boundary when
   capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized set is exactly
   modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text
   v0.32.0 — do not bump any; x/net/html stays OUT unless the owner approves
   an ADR); surface blockers with evidence (hash-level, not display) instead
   of improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-62's
   own acceptance per the spec's Tests section — evaluate what this session's
   engine landed honestly), commit + push everything (ask me for the GitHub
   PAT if it is not already available — never store it anywhere), write
   docs/sessions/SESSION-030.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-031.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

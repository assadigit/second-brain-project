# SESSION-029 — 2026-10-18 — Phase 5: T-62 slice 1 — THE RUN ENGINE'S DOORS landed with the grown differential golden (the preflights + the run-config card + the dry-run report writer + THE PHASE-5 CLEAR RE-SPEC resolved on evidence) — 1877 tests race-clean; the twenty-third full wipe repaired before work

## Goal (from NEXT-SESSION.md at start)

SESSION-029 = **T-62 — the run engine** per spec-cli.md's cli_run.py
section + the SESSION-028 shipped record: run_batch_visual, the
Phase-5 clear gate's bounded re-spec, the preflights, the dry-run
report writer, the SIGINT/heartbeat/exit law, the chrome-retry live
leg, the batch drivers. The capacity clause: one slice done properly
with proof beats a rushed whole.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the twenty-third full wipe)**: neither
  repo, no Go, no venv, no symlinks. Rebuilt per NEXT-SESSION v29 §1
  before any work: the work repo located via the GitHub API's
  authenticated /user/repos listing (`assadigit/second-brain-project`,
  pushed 18:05Z — SESSION-028's close); both repos cloned with the
  ephemeral Basic `x-access-token` header and **scrubbed immediately**
  (0 extraheader bytes in both configs, verified by direct read);
  GitCurator checked out at `98f4fc71cfb8…` (matches
  testdiff/PINNED-GITCURATOR-REV byte-for-byte); go1.27.2 installed at
  /home/z/.local/go (sha256 `ecadb99091…`, the same digest every
  session records); the venv rebuilt (Python 3.12.15, stdlib-only);
  the symlinks re-created. go.mod verified UNCHANGED (v1.36.1 /
  tls-client v1.16.0 / x/net v0.48.0 / x/text v0.32.0 / go 1.24.1 —
  the authorized set exactly). The baseline re-verified BEFORE work:
  **1866/1866 race-clean over 31 packages** (`--- PASS` counted; `go
  mod download` only — never `all`).

- **Read-first, per the card**: the reference's cli_run.py in full
  (the fetch/mark-read family, _write_dryrun_report, run_batch_visual
  239-504 with the signal wiring + the finish seals + the Phase-5
  block + the SIGINT/heartbeat law, the preflights 565-719),
  cli.py's cmd_auto (1252-1440) + cmd_mark_read + cmd_retry_failed +
  cmd_manual_batch + cmd_chrome_retry's live leg, cli_sync.py in
  full, core/dryrun.py, the ProcessingWorker's _run_impl (the setup
  380-946, the warmup family, the post-loop 1700-1889),
  worker/website_phase.py in full, cloudflare_sync.py's
  get_pending_result + build_vault_mirror_entries +
  build_decommission_events, cloudflare_manager.py's run loop +
  on_batch_complete + _do_batch_push, the GUI's app.py + window.py
  (the cf_manager wiring), ADR-0004 Am.3 + spec-worker-contract §8.4,
  and the landed slices (internal/cli/*, dump_cli.py,
  goldencli_test.go, the runner/undo/reports surfaces, the
  workerclient endpoints + helpers).

- **T-62 slice 1 — THE RUN ENGINE'S DOORS, commit `e2bc4cd`** (15
  files):
  - *runconfig.go* — `RunConfigCard` (cli_run.py 524-562): the boxed
    run-configuration snapshot with the raw-comparison laws pinned by
    the golden: the provider branches compare `llm_provider` RAW
    against "ollama"/"llamacpp" (a "Ollama" spelling falls to the
    OpenAI-compatible row, byte-for-byte like the reference);
    `cloud_model`/proxy host+port render through `.get(key, default)`
    (a present null renders "None", an absent one the default); the
    token mask counts RUNES (`len()` over Python str).
  - *preflight.go* — `PreflightChecks` + the three preflights
    (565-719): `_preflight_github` over the new `CLI.PreflightGithub`
    seam (the default door rides the connectioncheck GitHub seam — the
    same GET /user the battery's check makes; the 401/transport/login
    mapping with the class-name token), `_preflight_proxy` over the
    new `CLI.NetDial` seam (socket.create_connection's twin; the
    guarded int() port; the ms measure; the class-name token), and
    `_preflight_llm` with **THE model-picker bridge wired** — the
    not-pulled Ollama model opens the landed `PromptModelMenu`
    (allow-decline, the preflight passes no status line) and the pick
    persists through `ApplyModelChoice`; the llamacpp probe → detect
    fallback → `ResolveLlamaCppModel` → the auto-fill + save branch
    with the stale-URL re-persist; the cloud line; the GitHub-error
    "continue anyway? (y/N)" door gated on a TTY (the cron law — a
    non-TTY run never asks, it aborts).
  - *phase5.go* — **THE PHASE-5 CLEAR RE-SPEC, implemented** (see the
    gate evaluation below for the evidence chain): `PhaseFiveClear`
    (the all-clear verdict speech over `tracker.GetAllClear()`, gated
    exactly like the reference's `if success and bot_source`; the ⏸️
    not-verified count over the manifest's failed/processing/pending
    GitHub rows; the honest no-manifest line) + `PushVaultMirrorAfterBatch`
    (the batch-complete mirror push — the ack: the entries from the
    new `batch.VaultIndex.MirrorIndex()` export, the decommission
    events from the cache's `GetAllDecommissioned`, the sha256 index
    hash over the sorted urls joined `|` — the manager's own recipe —
    through `workerclient.Sync.PushVaultMirror`; best-effort +
    failure-isolated; a disabled Worker is silent per the reference's
    own is_enabled early-return) + `writeDryrunReport` (215-236: the
    plain-open report under APP_DIR/reports/dry-runs/ over the ported
    `dryrun.RenderReportMarkdown`, the frozen clock).
  - *Engine touches (additive, no behavior change)*:
    `dryrun.NowFn` (the frozen-clock seam — the reference's generator
    freezes the module's datetime; both stamp sites — the entry times
    + the Generated line — now ride the seam) and
    `batch.VaultIndex.MirrorIndex()` (the url→path export in the
    mirror entries' shape, sorted-url order — the documented
    deterministic-wire divergence at `BuildVaultMirrorEntries`).

- **THE DIFFERENTIAL STONE GREW**: dump_cli.py drives 10 more sections
  — the four run-config cards, the four `_preflight_github` cases
  (the fake PyGithub door swapped per case: ok → "authenticated as
  octo-owner" / 401-rejected / GatewayTimeout-unreachable / no-token),
  the four `_preflight_proxy` cases (the fake socket
  create_connection + the deterministic fake clock — disabled /
  reachable 42 ms / string-port / TimeoutError-unreachable), the eight
  `_preflight_llm` cases (ollama ready / unreachable (ConnectionError
  class token) / empty / **the menu-pick WITH ITS SAVED CONFIG
  bytes** / llamacpp ready / **the auto-fill WITH ITS SAVED CONFIG** /
  not-found / cloud), the `preflight_checks` composition (the ok card +
  the github-rejected abort over a non-TTY), and the dry-run report
  (the reference's own `record()` over the frozen clock + the written
  file's path + body) → **goldencli.golden.json at 69,028 bytes,
  two-run byte-identical**. *goldencli_test.go* replays every new
  section byte-for-byte (both saved configs included).
  **TestGoldenCLIMutationPreflights** proves the teeth bite on the
  checks card, the LLM saved config, and the dry-run report body.

- **Two new house laws found by the differential itself**:
  1. **THE FIXED-LENGTH ROOTS LAW**: the run-config card's right-side
     padding bakes the PATH LENGTHS into the golden's bytes (the box
     pads to `avail - 18 - 1 - len(value)`), so the replay's temp
     roots must be the same WIDTH as the generator's mkdtemp roots —
     `buildEnv` now uses `/tmp/gc-cli-app-goreplay` +
     `/tmp/gc-cli-vaults-goreplay` ("goreplay" = exactly the 8-rune
     suffix width tempfile.mkdtemp generates).
  2. **The shallow-copy snapshot law**: a section recording a config
     that the driven call MUTATES must `copy.deepcopy` BEFORE the call
     — `_preflight_llm`'s `apply_model_choice` writes through the
     shared nested dicts (`dict(config)` is a shallow copy), which
     leaked the post-run model into the recorded input (caught live:
     the replay's decoded config opened the wrong branch).

- **Spec updates** (reality wins, in the same commit): spec-cli.md —
  the Phase-5 open item REPLACED by the resolved re-spec with its full
  evidence chain + the preflights' WIRED note + the run-config card +
  the dry-run report writer + the SESSION-029 shipped record naming
  the honest remainder; README.md — the leading status; the roadmap's
  Phase-5 progress note; traceability E9's evidence cell; the task
  list's T-62 row (the slice record).

- **Pushed**: `68e177c..e2bc4cd main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — 0 extraheader, 0 token bytes; the commit message
  written to a FILE and committed with -F, the committed bytes
  verified with `git log --format=%B` — only git's own trailing
  newline differs).

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-62's own acceptance (the spec's Tests section + the task card),
  evaluated honestly — slice 1 of 2 LANDED, the remainder recorded**:
  - flag-surface snapshot (help text) — LANDED (SESSION-027; unchanged
    this session — the not-yet cards stay for the engine doors).
  - wizard flow (fake stdin) — LANDED (SESSION-028; unchanged).
  - status card rendering — LANDED (SESSION-027; unchanged).
  - `--sync` door beats — beat 1a LANDED (SESSION-027); **the live
    beats (1b/2/3) remain slice 2's** (they ride the Worker-queue
    fetch + the engine).
  - wrong-vault guard — NOT YET (slice 2; rides cmd_auto's fetch
    classification).
  - **Phase-5 gating of msg-id persist + mark-read — RESOLVED AND
    LANDED (this session's bounded open item)**: the re-spec is
    recorded in spec-cli.md with the evidence chain (ADR-0004 Am.3's
    "mirror-push is the read acknowledgement" + spec-worker-contract
    §8.4's "Mirror-push is the ack — /api/vault_mirror/push remains
    the one and only way items leave the pending queue" + the
    reference's own on_batch_complete pattern) and implemented as
    `PhaseFiveClear` + `PushVaultMirrorAfterBatch` with their unit
    battery (the verdict laws, the gate, the three push outcomes, the
    decommission build, the sha256 hash recipe).
  - **dry-run report equality (differential case) — LANDED**: the
    golden's dryrun_report section pins the reference's own record() →
    render → write over the frozen clock; the Go replay re-records the
    same entries through the ported dryrun package (the new NowFn
    seam) and byte-compares the path + the body.
  - Ctrl+C exit code — NOT YET (slice 2; the engine's SIGINT law).
  - terminal paint (ANSI/NO_COLOR) — LANDED (SESSION-027; the
    spinner-under-a-real-loop leg rides slice 2's engine).
  - UTF-8 hardening — LANDED (SESSION-027; the Windows live test stays
    Phase 7's clean-VM checklist).
- **Full-suite evidence at the close**: `go vet ./...` clean; `gofmt
  -l` clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 ./...` → **31 packages ok, 1877/1877 top-level tests
  PASS, 0 failures, 0 skips, 0 data races** (was 1866 at session
  start; +11 — the preflight/phase5 unit battery (10 tests:
  TestRunConfigCardMask, TestRunConfigCardTruncation,
  TestPreflightGithubSeam, TestPreflightProxySeam,
  TestPreflightLLMOllama, TestPreflightLLMCloud,
  TestPreflightChecksComposition, TestPhaseFiveClearVerdicts,
  TestPushVaultMirrorAfterBatch, TestWriteDryrunReport) +
  TestGoldenCLIMutationPreflights — the arithmetic closes exactly);
  go.mod/go.sum UNCHANGED (zero new dependencies — sha256 + net ride
  stdlib).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-cli.md` — the Phase-5 CLEAR
  dated note replaced by the RESOLVED re-spec (the three-part evidence
  chain + the reality-wins record of the manager's dormant wiring at
  the pinned rev) + the preflights' WIRED note (the seams + the
  bridge + the cron law) + the dry-run writer note + the status line +
  the SESSION-029 shipped record (the remainder named honestly).
- `README.md` — the leading status: SESSION-029's record + the numbers
  + the Next line.
- `docs/port-blueprint/12-migration-roadmap.md` — the Phase-5
  paragraph carries the SESSION-029 progress note.
- `docs/port-blueprint/traceability-matrix.md` — E9's evidence cell:
  the golden's new size + T-62 slice 1's record.
- `docs/port-blueprint/13-implementation-task-list.md` — T-62's row:
  the slice record + the named remainder.

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — nineteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html keep-or-swap** (standing, not a blocker).
- **Deferred test debt (recorded with evidence, not blockers)**:
  test_noteredo's remaining un-ported selections; the note_state
  SURGERY path's CRLF round-trip differential; the Windows live legs
  (Phase 7 clean-machine checklist).
- **T-62's remainder is recorded, not blocked**: the engine shell +
  the drivers are slice 2 (SESSION-030's first item) — the capacity
  clause applied honestly (one slice with proof beats a rushed whole).

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it carried the twenty-third-wipe rebuild + this slice's push; the
   configs verified clean after every use).

## Handoff

- Next session: **SESSION-030 = T-62 slice 2 — THE ENGINE + THE
  DRIVERS**: run_batch_visual's synchronous shell (the setup family +
  the warmup family + the website-phase wiring + the Finish chain +
  the seals + the Phase-5 CLEAR + the dry-run path + the
  SIGINT/heartbeat + the exit law) + cmd_auto/cmd_sync's live beats
  over the Worker /api/pending fetch (+ the _bot_queue_job
  classification half) + the wrong-vault guard + the >10 confirm +
  cmd_retry_failed + the import-file mode + the chrome-retry live leg +
  T-61's spinner under a real loop — then T-63 (launchers).
- NEXT-SESSION.md rewritten: yes (v30).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.

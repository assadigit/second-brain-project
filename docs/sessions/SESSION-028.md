# SESSION-028 — 2026-10-18 — Phase 5: T-60's remainder — THE COMMAND WIRINGS landed with the grown differential golden (the wizard + the radar, the battery, detect-llm + the model menu, hand-delivery complete, chrome-retry's listing + dry-run) — 1866 tests race-clean; the twenty-second full wipe repaired before work

## Goal (from NEXT-SESSION.md at start)

SESSION-028 = **T-60's remainder — the command wirings** per
spec-cli.md's SESSION-027 shipped record: the `--init` wizard with the
repo radar, `--test-connection` over the ported battery, `--detect-llm`
with the model menu, the `--hand-delivery`/`--chrome-retry` wirings,
and the differential stone grown with every one of them. The capacity
clause: one wiring done properly with proof beats a rushed whole.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the twenty-second full wipe)**: neither
  repo, no Go, no venv, no symlinks. Rebuilt per NEXT-SESSION v28 §1
  before any work: the PAT verified against the GitHub API's
  authenticated /user/repos listing (the work repo located — Currena
  the unrelated sibling); both repos cloned with the ephemeral Basic
  `x-access-token` header and **scrubbed immediately** (0 extraheader
  bytes in both configs, verified by direct read); GitCurator checked
  out at `98f4fc7` (matches testdiff/PINNED-GITCURATOR-REV
  byte-for-byte); go1.27.2 installed at /home/z/.local/go (sha256
  `ecbadb99…`, the same digest every session records); the venv rebuilt
  (Python 3.12.15, stdlib-only); the symlinks re-created. go.mod
  verified UNCHANGED (v1.36.1 / tls-client v1.16.0 / x/net v0.48.0 /
  x/text v0.32.0 / go 1.24.1 — the authorized set exactly). The
  baseline re-verified BEFORE work: **1852/1852 race-clean over 31
  packages** (the card's number exactly; `go mod download` only — never
  `all`).

- **Read-first, per the card**: the reference's cmd_init (108-299) +
  cmd_test_connection (750-807) + cmd_detect_llm (814-976) +
  cmd_hand_delivery (1032-1088) + cmd_chrome_retry (1095-1245),
  cli_settings.py's model-menu half, cli_run.py's
  _section/_check_line + the preflights, worker_jobs._quick_detect_job,
  the landed slice (internal/cli/*, dump_cli.py, goldencli_test.go),
  and the reference's own test_detectset/test_connection patch surface
  (the seams' source of truth).

- **T-60 slice 2 — THE COMMAND WIRINGS, commit `41d699e`** (22 files):
  - *wizard.go* — cmd_init: the whole _ask sequence (the YOUR_
    placeholder empty-start law, the isdir retry loop, the api-id
    digit law, the provider branches — the llamacpp branch's probe
    offering the detected model —, the proxy block with the n-branch
    raw-keep law, the 2–10 clamp), THE REPO RADAR (RadarPending +
    ApplyDetections + ReportLines run REAL over the `RadarAll` seam;
    the try/except recovered as the skip warning), save_config, the
    saved box (the reference's `.ljust(64)`-on-the-painted-string
    quirk preserved byte-exactly — found by the golden), and the
    honest Option-B closing card (no login line ever; the keys
    collected for compatibility per ADR-0004 Am.3).
  - *testconn.go* — cmd_test_connection: the battery over
    `connectioncheck.RunLocalChecks` + `Summarize` with cli_run.py's
    own `_section`/`_check_line` renderers; the Worker leg IS the
    queue check; the exit-code law holds (warn → 0, error → 1).
  - *detectllm.go* — cmd_detect_llm: `runDetectJob` mirrors
    worker_jobs._quick_detect_job's composition SYNCHRONOUSLY over the
    CLI's probe seams (the reference rides a GUI thread + spinner; the
    log lines ride cli_print exactly as the drain loop printed them),
    the numbered menu (the only-one line, the ← current mark, the
    --yes auto-pick, the invalid-choice abort), set + save through
    `_clean_config_for_save`, the SET card.
  - *settings.go grew the model-menu half* — SmartModelPick (the
    githubpipeline.PickBestModel twin — the recommendation and the
    headless auto-pick always agree), PromptModelMenu (the numbered
    picker: the recommendation as entry 1, the embedding tag override,
    the q-decline, the five-try give-up, the non-TTY auto-pick),
    ApplyModelChoice (the provider-aware write + save) — the preflight's
    model-picker bridge, ready for T-62.
  - *hands.go* — cmd_hand_delivery COMPLETE (the walled listing over
    the StateDB adapter — the decomtable law's call-site adapter —,
    the enqueue with queue.json + README.txt + the open-Chrome ladder,
    the folder card) and cmd_chrome_retry (the listing: retry rows +
    the 🖐 hand rows + the redo set; the FULL dry-run rehearsal through
    the real DeliverPagesViaChrome with the dry-run switch flipped;
    the live leg behind the honest T-62 boundary card).
  - *cli.go* — the dispatch wires every implemented command; the
    not-yet cards retire for all but the T-62 batch-engine doors.

- **Engine seams (additive, no behavior change — the reference's own
  mock surface, exported for cross-package injection)**:
  connectioncheck.HTTPGetJSONFn / GitHubGetFn / GitHubReply (the
  reference's suites monkeypatch `_http_get_json` / `_github_get`);
  githubpipeline.IsEmbedModel (the `_is_embed_model` twin);
  handdelivery.FindChromeFn / SubprocessPopenFn / WebbrowserOpenFn
  (the open ladder's patchable seams).

- **THE DIFFERENTIAL STONE GREW**: dump_cli.py's payload now drives 9
  more sections — the battery card (the stubbed net doors + the
  stubbed live-telegram leg — the reference's own test pattern) + the
  structured rows + the worker /health cases (the reference's own
  `http_get` injection: match / older / unreachable / no-URL), the two
  detect cards with their SAVED CONFIG bytes, the model-menu matrix
  (the REAL WorkerLlmMixin picker behind fake PyQt6/PyGithub modules —
  the guards are import doors, the reference's own logic runs), the
  two wizard runs over scripted stdin (the stubbed radar + the probe
  seams) with their SAVED CONFIG bytes — **the wizard's own writes,
  byte-compared** —, the hand-delivery card + queue.json + README.txt
  (the frozen hd.time, the open ladder's no-browser answers), and the
  chrome-retry dry-run card → *goldencli.golden.json* (48,698 bytes,
  two-run byte-identical). *goldencli_test.go* replays every ported
  section: the battery's shared sections [1/4]..[3/4] byte-compared +
  every Option-B divergence asserted (the subtitle, the Worker section
  title, the verdict counts, the wizard's closing card — the
  reference's records kept, never silently dropped); the worker rows
  structurally byte-compared; both detect cards + their saved
  configs; the model-menu matrix; both wizard prefixes + saved
  configs; the hand-delivery card + both written files; the
  chrome-retry card. **TestGoldenCLIMutationWirings** proves the teeth
  bite on the wizard's saved config, the enqueue's queue.json, and
  the detect card.

- **REALITY-WINS FIND (fixed in-commit, recorded in the spec)**: the
  reference's `cmd_chrome_retry --dry-run` leg calls
  `deliver_pages_via_chrome` WITHOUT flipping the global dry-run
  switch (only the batch worker ever flips it) — the pinned rev's
  dry-run would launch the REAL Chrome while the card says "Chrome is
  never launched". The port flips the switch around the delivery and
  restores + clears it right after; the golden's reference side stubs
  the deliver (the rehearsal's engine — its bytes already pinned by
  T-44's chrometabs golden) so both sides answer the same card.

- **A wizard quirk surfaced by the golden and PRESERVED**: the
  reference's provider-ask default "ollama" self-normalizes to
  "llamacpp" when kept via Enter (`"llama" in "ollama"`) — the
  reference's own golden output proves it; the port's identical
  normalization reproduces it byte-for-byte. Plus the GO NOTES: the
  reference's int() crash edges (a non-decimal isdigit answer, a
  non-integer proxy port) crash the whole wizard — the port answers 0
  / the or-door default (the happy paths byte-identical).

- **Pushed**: `fc9cd20..41d699e main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — 0 extraheader, 0 token bytes; the commit message
  written to a FILE and committed with -F, the committed bytes
  verified with `git log --format=%B` — only git's own trailing
  newline differs).

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **T-60's own acceptance (the spec's Tests section), evaluated
  honestly — the command wirings all landed**:
  - flag-surface snapshot (help text) — LANDED (SESSION-027; the
    surface unchanged this session — only not-yet cards retired).
  - wizard flow (fake stdin) — **LANDED** (the two golden wizard runs
    with scripted stdin + the saved-config byte-comparisons + the
    unit battery).
  - status card rendering — LANDED (SESSION-027; unchanged).
  - `--sync` door beats — LANDED (SESSION-027's pure beats; the live
    beats remain T-62's).
  - wrong-vault guard — NOT YET (T-62's --auto engine; recorded).
  - Phase-5 gating of msg-id persist + mark-read — NOT YET (dropped
    per the Option-B re-scope; the all-clear gate's Phase-5 re-spec
    remains T-62's bounded open item — untouched here, as the law
    demands).
  - dry-run report equality — NOT YET (T-62's writer; the chrome-retry
    dry-run rehearsal lands now, the report writer is the batch's).
  - Ctrl+C exit code — NOT YET (T-62's run engine).
  - terminal paint (ANSI/NO_COLOR) — LANDED (SESSION-027).
  - UTF-8 hardening — LANDED (SESSION-027; the Windows live test
    stays Phase 7's clean-VM checklist).
- **Full-suite evidence at the close**: `go vet ./...` clean; `gofmt
  -l` clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 -p 2 ./...` → **31 packages ok, 1866/1866 top-level tests
  PASS, 0 failures, 0 skips, 0 data races** (was 1852 at session
  start; +14 — the new command battery: 13 unit tests +
  TestGoldenCLIMutationWirings — the arithmetic closes exactly);
  go.mod/go.sum UNCHANGED (zero new dependencies).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-cli.md` — the status line +
  the command contracts carry the SESSION-028 WIRED notes (--init,
  --test-connection with the Option-B divergence record,
  --detect-llm + the model-menu half, --hand-delivery,
  --chrome-retry) + the GO NOTES (the wizard's int() hardening) + the
  REALITY-WINS find (the chrome-retry dry-run switch) + the
  SESSION-028 shipped record at the end.
- `README.md` — the leading status: SESSION-028's record + the
  numbers + the Next line (T-62 first).
- `docs/port-blueprint/12-migration-roadmap.md` — the Phase-5
  paragraph carries the SESSION-028 progress note.
- `docs/port-blueprint/traceability-matrix.md` — E9's evidence cell:
  T-60 LANDED across two sessions.
- `docs/port-blueprint/13-implementation-task-list.md` — T-60's row:
  the two-slice record; only T-61's paint legs remain on the CLI
  books.

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — eighteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html keep-or-swap** (standing, not a blocker).
- **Deferred test debt (recorded with evidence, not blockers)**:
  test_noteredo's remaining un-ported selections; the note_state
  SURGERY path's CRLF round-trip differential; the Windows live legs
  (Phase 7 clean-machine checklist).
- **T-62's bounded open item stands**: the all-clear gate's Phase-5
  re-spec (ADR-0004 Am.3 bounded) — untouched here, as the law
  demands.

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it carried the twenty-second-wipe rebuild + this slice's push; the
   configs verified clean after every use).

## Handoff

- Next session: **SESSION-029 = T-62 — the run engine** (--sync's
  live beats + the --auto engine + the preflight with the
  PromptModelMenu bridge + the wrong-vault guard + the dry-run report
  writer + the SIGINT/exit-code law + the chrome-retry live leg; the
  all-clear gate's Phase-5 re-spec resolved THERE, never improvised),
  with T-61's remaining paint legs riding along.
- NEXT-SESSION.md rewritten: yes (v29).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.

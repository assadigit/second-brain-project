# SESSION-027 — 2026-10-18 — Phase 5 OPENED: T-60's first slice — THE CLI SHELL — landed with the differential golden (the paint box, the sync door's pure halves, the honest Option-B surface, the FULL status card) — 1852 tests race-clean; the twenty-first full wipe repaired before work

## Goal (from NEXT-SESSION.md at start)

SESSION-027 = **Phase 5 OPENS — T-60 `internal/cli`** (the engine parity
milestone's first task, per spec-cli.md with its Option-B re-scope notes).
The one-task law + the session prompt's capacity clause: land the first
coherent slice properly with its own differential proof and record the
remainder honestly.

## What was done (task IDs, files, commits — cite hashes)

- **The environment was WIPED (the twenty-first full wipe)**: neither
  repo, no Go, no venv, no symlinks. Rebuilt per NEXT-SESSION v27 §1
  before any work: the PAT verified against the GitHub API's
  authenticated /user/repos listing; both repos cloned with the
  ephemeral Basic `x-access-token` header and **scrubbed immediately**
  (verified by direct config read — 0 extraheader bytes; the
  work-repo location confirmed via the API listing, Currena the
  unrelated sibling); GitCurator checked out at `98f4fc7` (matches
  testdiff/PINNED-GITCURATOR-REV byte-for-byte); go1.27.2 installed at
  /home/z/.local/go (sha256 `ecbadb99…` — the same digest
  SESSION-025/026 recorded); the venv rebuilt (/home/z/venvs/gitcurator,
  Python 3.12.15, stdlib-only); the symlinks re-created. go.mod verified
  UNCHANGED (v1.36.1 / tls-client v1.16.0 / x/net v0.48.0 / x/text
  v0.32.0 / go 1.24.1 — the authorized set exactly). The baseline
  re-verified BEFORE work: **1828/1828 race-clean over 30 packages**
  (foreground chunks; the top-level count matched the card exactly).

- **Read-first, per the card**: the reference's CLI sources in full
  (cli.py 1,778 / cli_sync.py 381 / cli_settings.py 211 /
  cli_terminal.py 321 / cli_run.py 719 — the run engine read for the
  boundary decision, its legs deferred with T-62), the spec's banner +
  Tests section, the existing Phase-0 seed, the reference's own
  test_clisync.py battery (the door's laws), and the newest generator
  precedent (dump_goldenwebsites.py).

- **Capacity assessment (honest, per the prompt's clause)**: T-60 whole
  = ~3,500 reference LOC + the full test battery — more than one
  session. **The first coherent slice = THE CLI SHELL**, commit
  `143c590` (22 files, +3,638 lines):
  - *terminal.go* — the paint box verbatim: the C ANSI system (the
    NO_COLOR/TTY/--no-color law), level styles, term_width's
    COLUMNS/TTY/80 law, mask/_mask_token, rule, cli_print over the
    frozen-clock seam, the banner (the stale v0.09.5 quirk preserved),
    render_bar with banker's rounding, the StatusLine
    (start/tick/erase/log/stop), _strip_ansi, the Windows 65001 console
    hardening via syscall.NewLazyDLL (zero new deps; POSIX twins for
    TIOCGWINSZ + isatty).
  - *settings.go* — _config_path/load_config/save_config over the T-12
    config port, _ask (EOF answers the default), _mask_token,
    _clean_config_for_save.
  - *syncdoor.go* — THE SYNC DOOR's pure halves: GUI_INSTRUCTION,
    sync_failure_card, sync_missing_essentials (the isdir seam),
    sync_confirm_prompt (Enter/other/--yes/non-TTY/EOF) + cmd_sync's
    BEAT 1a wired (the no-config and missing-essentials failure cards
    differential-pinned; the engine beats land with T-62 behind an
    honest in-development card).
  - *cli.go* — the honest Option-B flag surface + the dispatch order +
    the friendly card + OUR --help: **the Telethon-only flags stay
    recognized and answer explicit refusal cards citing ADR-0004 Am.3
    (exit 1, never a silent no-op, never an unknown-flag crash) —
    T-60's honest-surface DECISION, recorded in the spec banner**; the
    not-yet commands answer in-development cards naming their landing
    task; the argparse laws (--help → stdout exit 0; invalid choice /
    unknown flag → exit 2; the optional ints resolve present-vs-absent
    after parse — a real bug caught by the refusal-card tests).
  - *status.go* — the FULL --status card: the 16 config rows, the
    taxonomy row, the sentinel rows over the connectioncheck port's
    seams, the Ollama/llama.cpp model checks over the llmclient seams,
    the CacheDB/NoteStateDB/WebsiteStateDB stats (the ported store,
    notestate, websitepipeline — all direct).
  - *dead.go* — --list-dead/--reset-dead (a missing config continues
    EMPTY per `load_config(path) or {}`).
  - *cfgutil.go* — the Python-dict access laws over the house
    OrderedMap (every falsy door byte-exact: 0 → "(not set)", "" →
    default).

- **THE DIFFERENTIAL STONE**: *testdiff/goldens/dump_cli.py* — the
  generator driving the reference's OWN CLI shapes with the house
  seams: **gui.icons stubbed with an empty module** (PyQt6-free import
  chain for the stdlib venv), **gitcurator.gui.app stubbed to the REAL
  CacheDB + dead_link_threshold** (the app.py door only re-exports
  them; its guards would exit), net_guard stubbed with fixed probe
  dicts, a fake ollama module feeding list_models_with_timeout,
  **APP_DIR patched BEFORE the cache_db import** (its APP_DIR binds at
  import time), the clock frozen at 2026-10-17 12:00:00 across
  cli_terminal/cache_db/note_state/website_state, both temp roots
  folded by REPLACE, the two-run byte-identical proof built in →
  *internal/cli/testdata/goldencli.golden.json* (23,528 bytes, 16
  sections incl. the env record the replay rebuilds).
  *internal/cli/goldencli_test.go* replays every ported section
  byte-for-byte (banner, paint shapes with colors ON and OFF, the
  cli_print lines, the essentials matrix, the button beats with their
  prompt bytes, the failure card, the FULL status card over the
  rebuilt crafted env — config + cache.db crafted through the ported
  store's own methods, the quarantine rows pinned to the reference's
  frozen stamps as fixture state —, status/sync no-config, sync
  bad-config, list-dead, reset-dead LAST); the reference's no-flag help
  card kept as the pre-Option-B RECORD — not byte-compared (our honest
  surface intentionally differs), with the divergence asserted; the
  mutation test bites. **The replay went green after exactly TWO
  one-token fixes** (the " ready" suffix on the Ollama row; the
  Enter-beat's scripted stdin) — the shapes held on the first
  serious run.

- **A new environment quirk recorded**: the Write/Edit tool path
  REDACTS token-shaped strings at file-write time (a crafted fake
  `ghp_…` test token became a placeholder in the generator source) —
  the display path ALSO redacts on read-back. The house law
  (no token-shaped strings in git even when fake) forced a safe-shaped
  crafted token (`fake-github-token-…`); the golden is internally
  consistent (both sides mask the same string). The
  renderer-eats-escape-pairs law confirmed AGAIN: the golden's ANSI
  bytes display mangled but compare byte-exact (the replay compares
  JSON-decoded bytes, never display).

- **Pushed**: `2cc07d1..143c590 main -> main` (the ephemeral PAT via
  the transient `-c http.extraheader` Basic form; the config verified
  CLEAN after — 0 extraheader, 0 token bytes; the commit message
  written to a FILE and committed with -F, the committed bytes
  verified with `git log --format=%B` — only git's own trailing
  newline differs).

## Gate evaluation (pass/fail/partial + EVIDENCE)

- **The Phase-5 entry gate: already earned** (Phase 4 CLOSED at
  `7aea86c`; verified at session start — the tree clean, origin/main
  at the SESSION-026 close, the baseline 1828/1828 re-verified).
- **T-60's own acceptance (the spec's Tests section), evaluated
  honestly — PARTIAL by design (slice 1 of the task)**:
  - flag-surface snapshot (help text) — **LANDED** (TestHelpSnapshot +
    TestHelpCardSnapshot pin OUR honest surface; the golden's
    help_card_reference section records the pre-Option-B reference;
    the divergence asserted, never silent).
  - status card rendering — **LANDED** (the golden's status_card
    section, byte-compared over the full crafted env; the Phase-0
    status-no-config differential case still PASSES through the
    reworked card — testdiff/harness.py green).
  - `--sync` door beats (Enter/other/`--yes`/non-TTY) — **LANDED**
    (sync_confirm_prompt pure + golden-pinned; cmd_sync beat 1a
    wired; beats 1b/2/3 = T-62's engine).
  - wrong-vault guard — **NOT YET** (T-62's --auto engine; recorded).
  - Phase-5 gating of msg-id persist + mark-read — **NOT YET** (the
    Option-B re-scope already dropped the msg-id cursor/mark-read
    steps; the all-clear gate's Phase-5 re-spec remains T-62's bounded
    open item per ADR-0004 Am.3 — never improvised here).
  - dry-run report equality — **NOT YET** (T-62's writer).
  - Ctrl+C exit code — **NOT YET** (T-62's run engine).
  - terminal paint (ANSI/NO_COLOR) — **LANDED** (the golden's
    paint_shapes section with colors ON and OFF, byte-compared;
    TestEnableColorsDoor + the StatusLine grammar tests).
  - UTF-8 hardening — **LANDED** (the Windows 65001 port +
    Go-native replace; the Windows-console live test stays Phase 7's
    clean-VM checklist per T-61's acceptance note).
- **Full-suite evidence at the close**: `go vet ./...` clean; `gofmt
  -l` clean; `GOOS=windows go build ./...` clean; `go test -race
  -count=1 -p 2 ./...` → **31 packages ok, 1852/1852 top-level tests
  PASS, 0 failures, 0 skips, 0 data races** (was 1828 at session
  start; +24 — the new internal/cli battery: TestGoldenCLI +
  TestGoldenCLIMutation + 22 unit tests — the arithmetic closes
  exactly); go.mod/go.sum UNCHANGED (zero new dependencies).

## Spec updates made (file → what changed, why)

- `docs/port-blueprint/module-specs/spec-cli.md` — the banner carries
  T-60's honest-surface DECISION (refusal cards, flags stay
  recognized) + the SESSION-027 shipped record (the slice's
  construction, the golden, the remainder list); **reality-wins fix**:
  the --status "never creates state" claim corrected (the CacheDB
  block opens — and on a fresh install CREATES — APP_DIR/cache.db;
  only the note-state/websites blocks are read-when-exists); the
  dead-level-color quirk recorded as a cli_terminal Go note.
- `README.md` — the leading status: Phase 5 OPENED (SESSION-027), the
  slice's record, "Next: T-60's remainder, then T-62".
- `docs/port-blueprint/12-migration-roadmap.md` — the Phase-5
  paragraph carries the progress note (the slice + the numbers).
- `docs/port-blueprint/traceability-matrix.md` — E9's evidence cell
  carries the slice + the golden.
- `docs/port-blueprint/13-implementation-task-list.md` — T-60's row
  records the slice landed + the remainder.

## Blockers & decisions needed

- **The go-directive acknowledgement (STILL WAITING — seventeenth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html keep-or-swap** (standing, not a blocker).
- **Deferred test debt (recorded with evidence, not blockers)**:
  test_noteredo's remaining un-ported selections (the redo-scan
  battery is the top item); the note_state SURGERY path's CRLF
  round-trip differential (Phase 4+); the Windows live legs (Phase 7
  clean-machine checklist).
- **T-62's bounded open item stands**: the all-clear gate's Phase-5
  re-spec (ADR-0004 Am.3 bounded) — untouched here, as the law
  demands.

## Owner actions requested

1. The go-directive acknowledgement (ADR-0002 Am.2) — a one-liner.
2. Standing: keep the PAT available in chat when starting a session
   (it carried the twenty-first-wipe rebuild + this slice's push; the
   configs verified clean after every use).

## Handoff

- Next session: **SESSION-028 = T-60's remainder — the command
  wirings** (the `--init` wizard with the repo radar +
  the llamacpp probe seam; `--test-connection` over the ported
  battery (connectioncheck.RunLocalChecks + Summarize + the
  _check_line/_section renderers); `--detect-llm` (probe → menu → set
  + save — cli_settings's model-menu half lands with it);
  `--hand-delivery` + `--chrome-retry` (the engines are ported — the
  CLI wiring follows); then T-62 if capacity remains (the --sync live
  beats + the --auto engine + the dry-run report writer + the
  SIGINT/exit-code law).**
- NEXT-SESSION.md rewritten: yes (v28).
- The next prompt: see `docs/NEXT-SESSION.md` §5 — the owner copies
  it verbatim.

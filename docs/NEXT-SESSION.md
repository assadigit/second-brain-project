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
  `41d699e` (SESSION-028) — the paint box, the settings halves, the
  sync door's pure halves + the honest Option-B flag surface, the FULL
  status card, list/reset-dead, the --init wizard with the repo radar,
  --test-connection, --detect-llm + the model menu, --hand-delivery,
  --chrome-retry's listing + dry-run rehearsal.
- **T-62 LANDED ACROSS TWO SESSIONS (Phase 5) — COMPLETE**: slice 1 —
  THE RUN ENGINE'S DOORS at `e2bc4cd` (SESSION-029: the preflights +
  run_config_card + the dry-run report writer + THE PHASE-5 CLEAR
  RE-SPEC resolved on evidence); slice 2 — THE ENGINE + THE DRIVERS at
  `6ccf184` + the close at `e0cef51` (SESSION-030: run_batch_visual's
  synchronous shell — the setup family, the warmup family, the GitHub
  loop, the websites phase, the Finish seals in the reference's exact
  order, the Phase-5 CLEAR, the dry-run path, the SIGINT drain + the
  100 ms heartbeat ticker, the exit law; the Worker-queue fetch + the
  _bot_queue_job classification half; cmd_auto's wrong-vault guard +
  the >10 confirm; cmd_sync's live beats; cmd_retry_failed; the
  import mode; the chrome-retry LIVE leg; `internal/platformintake`;
  the differential golden at 86,570 — the reference's OWN engine
  driven over a fake Qt event loop, the 29-line log sequence + the
  BATCH DONE card + the exit codes + the 12-entry report
  byte-compared, REAL mutation teeth, two-run byte-identical).
  **1911 tests race-clean over 32 packages** at the close.
- **THE PUSH DEBT IS CLEARED** (by SESSION-030's SECOND
  continuation, immediately after the close, with the owner's
  re-supplied ephemeral PAT): the pre-push byte-level secret scan ran
  first (the two-commit diff + the working tree + .git/config — all
  clean), then `git -c http.extraheader=… push origin main` landed
  `6ccf184..e175619`. Server-verified via authenticated ls-remote
  (`refs/heads/main` → `e175619616e3…`, byte-equal to local HEAD);
  the PAT never persisted (config re-checked: 0 extraheader bytes).
  The battery RE-VERIFIED at the pushed HEAD: **1911/1911 top-level
  tests race-clean over 32 packages** (`-count=1 -race`, exit 0,
  0 FAIL / 0 SKIP). SESSION-031 starts from a clean push state —
  after the environment check, verify `git status` shows up to date
  with 'origin/main' (the tip at or beyond the SESSION-030 post-close
  record `02d4ed8`; if the sandbox wiped the clone,
  re-clone and reconcile; never end a session push-pending — that is
  the PUSH-AFTER-EVERY-TASK law below, whose cautionary tale this
  debt became and then was paid).
- **Remaining on Phase 5's books**: T-61's remaining paint legs (the
  close-out audit — the paint box, the NO_COLOR/TTY law, the spinner
  under the heartbeat, and the 65001 hardening are all LANDED; the
  row's remainder is the Windows-console verification), T-63 (the
  launchers), T-64 (the differential corpus full run), T-65 (the
  `v0.92.0-go.1` tag + zip).
- **The environment recipe** (unchanged): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator;
  must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256 `ecadb99…`), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
  build/TOOLCHAIN.md, go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (the
  tls-client floor, ADR-0002 Am.2; never bump).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE STATE-CHAIN LAW** (SESSION-030): a golden replay of
    sequential sections is ONE state chain — an ISOLATED replay
    (e.g. a mutation test with a fresh buildEnv) must reproduce the
    payload's own pre-state (the `--reset-dead` run clears the dead
    set — the engine's banish pass otherwise records a missing-note
    backfill the golden never saw; the dry-run report writer's real
    `<app>/reports` dirs — the engine's makedirs rows otherwise say
    "create folder" instead of "folder exists") or the bytes diverge.
  - **THE MUTATION-TEETH LAW** (SESSION-030): a mutation test runs
    the REAL replay against the mutated golden and must FAIL, while
    the unmutated golden must still PASS (the harness guard) — a
    mutation test that only proves the pattern strings exist in the
    golden's text has NO teeth (the SESSION-030 close caught exactly
    that in the in-flight draft and replaced it).
  - **THE CONTEXT-DEATH CONTINUATION recipe** (SESSION-030, the
    SESSION-021 law proven in practice): git status + `git log
    --oneline -5` + file timestamps tell what ACTUALLY survived;
    audit the in-flight work (never blindly trust it — the
    continuation caught a toothless mutation test, a mangled
    comment, a stray DSN-named file, and two doc defects), repair,
    run the FULL battery, then land.
  - **THE DSN-EMPTY-PATH HAZARD** (SESSION-030): modernc.org/sqlite
    treats a DSN whose path is empty (`"?_pragma=…"`) as a LITERAL
    filename — it creates a file named `?_pragma=busy_timeout(30000)
    &_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)` in the
    process's cwd. Never pass "" as a db path;
    `NewWebsiteStateDB("cache.db")` is the AppDir-anchored default.
  - **THE FIXED-LENGTH ROOTS law** (SESSION-029): a golden section
    whose output PADS a path to a fixed width bakes the path LENGTH
    into the golden's bytes — the replay's buildEnv roots must be
    exactly as wide as the generator's mkdtemp roots
    (`/tmp/gc-cli-app-goreplay` + `/tmp/gc-cli-vaults-goreplay`;
    "goreplay" is the 8-rune suffix width; keep it).
  - **THE SHALLOW-COPY SNAPSHOT law** (SESSION-029): a golden
    section recording a config the driven call MUTATES must
    copy.deepcopy BEFORE the call.
  - The fake-clock arithmetic (1000.0 → 1000.042 renders 42 ms
    through `int((t1-t0)*1000)`); exception class-name tokens via
    `type("TimeoutError", …)` (the SHAPE is the contract); a golden
    section capturing a RETURN VALUE cannot use the stdout-capture
    helper.
  - **THE CD-DISCIPLINE LAW** — the shell resets its working
    directory between tool calls: every git/go command carries its
    own cd (or `go -C`); never commit outside the work repo.
  - **THE RENDERER-EATS-ESCAPE-PAIRS lesson** — the two-character
    pairs `[h` and `[m` are eaten on the OUTPUT path, the INPUT
    path, AND the BASH-ARG path: write commit messages to a FILE and
    commit with `-F`, verify with `git log --format=%B`, never trust
    a DISPLAY of escape-bearing bytes.
  - **THE WRITE-TOOL REDACTION lesson** — the Write/Edit tools
    REDACT token-shaped strings at file-write time (craft
    safe-shaped test tokens, never ghp_-shaped; byte-level
    secret-scan before every commit).
  - **THE BACKGROUND-PROCESS lesson** — detached processes do not
    survive tool-call boundaries; run long suites in FOREGROUND
    CHUNKS.
  - **PUSH AFTER EVERY TASK** (the SESSION-030 debt is exactly what
    this law prevents — ask the owner for the PAT the MOMENT a push
    is due; never store it).

## 2. Next session: SESSION-031 — T-61's close-out audit + T-63 the launchers

- **Verify the push state first** (§1): SESSION-030's debt was
  cleared by its second continuation — origin/main carries the
  SESSION-030 record through the post-close commit `02d4ed8` — so
  the session does NOT begin push-pending; confirm
  `git status` is in sync before work (and keep pushing after every
  task).
- **T-61 (S) — the terminal-paint close-out**: the paint box, the
  NO_COLOR/TTY/--no-color law, `term_width`'s COLUMNS law, the
  spinner + bar under the heartbeat, and the Windows 65001 console
  hardening (SetConsoleOutputCP/SetConsoleCP via syscall) are ALL
  LANDED (terminal.go, status.go, the engine's ticker). The
  remainder is the row's OWN verification: the Windows-console test
  leg — audit what spec-cli §cli_terminal.py + §Tests still demand
  (the ANSI/NO_COLOR paint tests, the UTF-8 hardening evidence per
  GOOS=windows), close the row with evidence or record the exact
  honest remainder (the clean-VM legs belong to T-70's Phase-7
  checklist, not here).
- **T-63 (S) — THE LAUNCHERS** per spec-cli's Shims + launchers
  section + doc 10 §2-6:
  1. **`cmd/gitcurator-cli`** — the main package (the arg wiring →
     `internal/cli`'s Run; the exit code passes through; stdin/stdout
     wired to the real TTY doors).
  2. **The `.bat` launchers** (ASCII-ONLY, CRLF line endings —
     assert BOTH in a test): `setlocal` → title → `cd /d "%~dp0"` →
     the walk-up `if not exist … cd /d "%~dp0\..\.."`
     (doc: harmless in Go — the assets are embedded) → run
     `GitCurator-cli.exe` DIRECTLY (no python discovery — the
     reference's venv-first ladder falls away) with the reference's
     own defaults (`GitCurator-CLI.bat` → `--cli --auto --yes`;
     `GitCurator-SYNC.bat` → `--cli --sync %*`; the Setup/Start
     equivalents) → `EXITCODE` echo + `pause`.
  3. **`gitcurator-cli.sh`** — `set -euo pipefail`, the walk-up,
     the exe, default `--cli --auto --yes`.
  4. **The `1-INSTALL`-equivalent checks** (doc 10 §6: git.exe on
     PATH — the same requirement as the Python release; print the
     next steps; NO registry writes, no PATH mutations).
  5. **The VERSION stamp** (doc 10 §3: stamped from the repo, the
     zip layout flattens the launchers to the root) — the build
     command shape lands (`GOFLAGS=-trimpath CGO_ENABLED=0 go build
     -o bin/GitCurator-cli.exe ./cmd/gitcurator-cli`); the actual
     release packaging is T-65.
  6. Tests: the .bat bytes (ASCII + CRLF assertions), the sh bytes,
     the cmd wiring (a smoke run through Run with a crafted config),
     GOOS=windows build of ./cmd/... clean.
- **The differential stone grows if the launchers have byte-level
  surfaces the reference owns** (the .bat set IS byte-anchored —
  extend dump_cli.py with a section that reads the reference's own
  .bat bytes and replay-compares the Go equivalents' bytes where the
  reference's grammar is shared; the venv-discovery lines DIVERGE by
  design — assert the divergence, never silently). MIND THE LAWS
  (fixed-length roots, deep-copy-before-mutation, unfold-before-
  feeding, the state-chain law, the mutation-teeth law).
- If capacity runs short: land `cmd/gitcurator-cli` + the .bat set
  with their byte tests properly first; record the remainder
  honestly. One slice done properly with proof beats a rushed whole.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-030.md` — the most recent report (T-62
   COMPLETE's evidence, the context-death continuation's audit, the
   state-chain + mutation-teeth laws, the push-PENDING record).
2. `docs/port-blueprint/module-specs/spec-cli.md` — THE SPEC (the
   SESSION-027..030 shipped records + the Shims+launchers section +
   the Tests section).
3. `docs/port-blueprint/10-build-release-packaging.md` §2-6 — the
   build commands, the zip layout, the launcher laws (ASCII/CRLF,
   the walk-up, the 1-INSTALL checks), the VERSION stamp.
4. The reference's launcher surfaces: `GitCurator-CLI.bat`,
   `GitCurator-SYNC.bat`, the Setup/Start `.bat`s,
   `gitcurator-cli.sh`, `1-INSTALL.bat` (in the repo root / scripts
   area — locate by `Glob`), and `gitcurator/cli/cli.py`'s main
   entry (the arg wiring the cmd package mirrors).
5. The landed CLI (the patterns to extend): `internal/cli/cli.go`
   (Run + the arg surface + the exit-code law), `terminal.go` +
   `status.go` (the paint + the spinner under the heartbeat — T-61's
   landed legs), `engine.go` (the shell the launchers drive), the
   golden harness (`testdiff/goldens/dump_cli.py` +
   `internal/cli/goldencli_test.go`).

## 4. Known constraints / owner actions for SESSION-031

- **Push state: CLEAN** — SESSION-030's push debt (`e0cef51` + the
  close commit, stranded local by the mid-session context death) was
  cleared by the session's second continuation with the owner's
  re-supplied ephemeral PAT: pushed, server-verified, and the battery
  re-verified at the pushed HEAD (1911/1911). The PAT was never
  stored and never echoed into any file or command log that
  persists. No first-action push is needed — keep the
  PUSH-AFTER-EVERY-TASK law.
- **The go-directive acknowledgement (STILL WAITING — twentieth
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2
  has the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers
  x/net/html, approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB the persisted
  extraheader from .git/config after the clone); goldens are
  GENERATED by running its unmodified code.
- **Dependency law**: `modernc.org/sqlite` v1.36.1;
  `bogdanfinn/tls-client` v1.16.0; `golang.org/x/net` v0.48.0 +
  `golang.org/x/text` v0.32.0 (direct); ANY other add needs a new
  ADR — surface, don't improvise. (SESSION-030 added zero.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections
  (ScanMasterRedoRows is the top item); the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).
- Standing: secrets / owner data never in git (byte-level
  secret-scan before every commit); tests never weakened; reality
  wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending
  (SESSION-030 is the cautionary tale: a context death between the
  task and its push leaves the PAT irrecoverable).

## 5. THE PROMPT TO COPY (SESSION-031)

```
You are starting SESSION-031 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-four full wipes across
   SESSION-005..030 plus mid-session degradations; SESSION-021 added the
   CONTEXT-death class and SESSION-030 PROVED the continuation recipe: git
   status + git log --oneline -5 + file timestamps tell what ACTUALLY survived;
   a preserved commit or in-flight draft is continued and AUDITED — never
   blindly trusted, never restarted from zero). FIRST verify with evidence:
   work repo at /home/z/my-project/second-brain-project (symlinked
   /home/z/second-brain-project), read-only reference /home/z/my-project/GitCurator
   @ 98f4fc7 (symlinked /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 → /home/z/.local/go (go.dev tarball, sha256 ecbadb99…), venv
   /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, and
   go.mod with modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not
   bump any of these). Then check `git status` and `git log --oneline -3`:
   SESSION-030 landed T-62 COMPLETE and PUSHED everything — the engine+drivers
   at 6ccf184, the differential golden's close at e0cef51, and the session's
   close commit e175619 (the push debt stranded by the mid-session context
   death was cleared by the session's SECOND continuation with the owner's
   re-supplied ephemeral PAT; server-verified; battery re-verified at the
   pushed HEAD: 1911 tests race-clean over 32 packages). origin/main carries
   SESSION-030's full record through the post-close commit 02d4ed8 — verify
   git status shows up to date with origin/main (local main's tip must be at
   02d4ed8 or newer). Use the PAT
   only ephemerally: git -c http.extraheader="Authorization: Basic $(printf
   'x-access-token:%s' "$PAT" | base64)" and never store it — AND SCRUB IT if
   you ever CLONE with it: `git clone -c http.extraheader=…` PERSISTS the
   header into the new clone's .git/config, so `git config --unset-all
   http.extraheader` in BOTH clones right after; Currena
   is an UNRELATED sibling; a mid-session mode-flip is repaired with
   `git checkout-index -f -a` after `git fsck`. Re-verify the suite before work
   if anything looks off. SESSION-030's operational laws (newest first): THE
   STATE-CHAIN LAW — a golden replay of sequential sections is ONE state chain
   and an ISOLATED replay (a mutation test with a fresh buildEnv) must
   reproduce the payload's own pre-state (the --reset-dead run clears the dead
   set — the engine's banish pass otherwise records a missing-note backfill
   the golden never saw; the dry-run writer's real <app>/reports dirs — the
   makedirs rows otherwise say "create folder" instead of "folder exists");
   THE MUTATION-TEETH LAW — a mutation test runs the REAL replay against the
   mutated golden and must FAIL while the unmutated golden still PASSES (a
   test that only proves the patterns exist in the golden's text has NO
   teeth); the DSN-empty-path hazard — modernc treats a "?_pragma…" DSN as a
   LITERAL filename, so never pass "" as a db path
   (NewWebsiteStateDB("cache.db") is the AppDir-anchored default). The older
   laws stand: FIXED-LENGTH ROOTS (the replay's roots stay exactly as wide as
   the generator's mkdtemp roots — /tmp/gc-cli-app-goreplay +
   /tmp/gc-cli-vaults-goreplay); SHALLOW-COPY SNAPSHOT (deep-copy a config
   the driven call mutates BEFORE the call); the fake-clock arithmetic
   (1000.0 → 1000.042 renders 42 ms via int((t1-t0)*1000)); exception
   class-name tokens via type("TimeoutError", …) — the SHAPE is the contract;
   a golden section capturing a RETURN VALUE cannot use the stdout-capture
   helper; (a) THE CD-DISCIPLINE LAW — every git/go command carries its own
   cd (or go's -C flag); never commit outside the work repo; (b) the
   renderer-eats-escape-pairs lesson — the two-character pairs  and 
   are eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path: write
   commit messages to a FILE and commit with -F, verify the committed bytes
   with 'git log --format=%B', never trust a DISPLAY of escape-bearing bytes;
   (c) the Write/Edit tool REDACTS token-shaped strings at file-write time
   (craft safe-shaped test tokens, never ghp_-shaped; byte-level secret-scan
   before every commit); (d) detached background processes DO NOT SURVIVE
   tool-call boundaries — run long suites in FOREGROUND CHUNKS per tool call.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (T-62 COMPLETE at e0cef51 + the close
   commit — the engine's record is in the status) → docs/SESSION-PROTOCOL.md
   (the working law) → docs/NEXT-SESSION.md (your state card — §1 records
   SESSION-030's close: T-62 COMPLETE, 1911 tests race-clean over 32
   packages, the push debt CLEARED by the second continuation, the newest laws
   incl. the state-chain +
   mutation-teeth ones) → docs/sessions/SESSION-030.md (the most recent
   report — the context-death continuation's audit + the golden's evidence).

3. THIS SESSION IS T-61's CLOSE-OUT AUDIT + T-63 THE LAUNCHERS (both S —
   one session fits both): per docs/port-blueprint/module-specs/spec-cli.md's
   Shims+launchers section + §Tests, doc 10 §2-6, and the SESSION-030 shipped
   record (READ THE SPECS + THE LANDED CLI BEFORE THE CODE). T-61: the paint
   box, the NO_COLOR/TTY law, the spinner under the heartbeat, and the 65001
   hardening are LANDED — audit what the row still demands (the
   Windows-console verification legs; the clean-VM legs belong to Phase 7,
   not here) and close the row with evidence or record the exact honest
   remainder. T-63: cmd/gitcurator-cli (the main package — the arg wiring →
   internal/cli's Run, the exit code passes through); the .bat launchers
   (ASCII-ONLY + CRLF, BOTH asserted in tests; setlocal → title → cd /d
   "%~dp0" → the walk-up if not exist → GitCurator-cli.exe DIRECTLY — no
   python discovery, the reference's own defaults: GitCurator-CLI.bat →
   --cli --auto --yes, GitCurator-SYNC.bat → --cli --sync %*, the
   Setup/Start equivalents → EXITCODE echo + pause); gitcurator-cli.sh
   (set -euo pipefail, the walk-up, the exe, the default); the
   1-INSTALL-equivalent checks (git.exe on PATH, the next-steps card, NO
   registry writes, no PATH mutations); the VERSION stamp + the build
   command shape (GOFLAGS=-trimpath CGO_ENABLED=0 go build -o
   bin/GitCurator-cli.exe ./cmd/gitcurator-cli — the release packaging is
   T-65). If the launchers have byte-level surfaces the reference owns (the
   .bat set IS byte-anchored), extend the differential stone: dump_cli.py
   reads the reference's own .bat bytes, the replay byte-compares the shared
   grammar and ASSERTS the venv-discovery divergence (never silent); MIND
   THE LAWS (fixed-length roots, deep-copy snapshots, unfold-before-feeding,
   the state-chain law, the mutation-teeth law). If capacity runs short,
   land cmd/gitcurator-cli + the .bat set with their byte tests properly
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

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-61's
   and T-63's own acceptance per the specs — evaluate what landed
   honestly), commit + push everything (ask me for the GitHub PAT the moment
   a push is due — never store it anywhere; never end the session
   push-pending), write docs/sessions/SESSION-031.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt (T-64 — the
   differential corpus full run — is the natural next task, then T-65 the
   tag), and end your reply with a plain-language report plus the prompt I
   should copy to start SESSION-032.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

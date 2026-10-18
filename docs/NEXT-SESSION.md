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
  (`e5846bd`), T-34 `internal/workerclient` (`6ffc6e9`) — 872 tests
  race-clean at the close.
- **PHASE 4 CLOSED (SESSION-015–026)**: the task list T-40..T-50 ALL
  LANDED (T-50 mirror at `555a72c`; T-51 removed by the Option-B
  stamp) — T-40 `8098ac4`, T-41 `f528c22`, T-42 `e92e4e7`, T-43
  `be688bd`, T-44 `71e5e34`, T-45 `5c0508e`, T-46 `992a3d1`, T-47
  `d3a5d27`, T-48 `7179a6c`, T-49 `e2c941a`, T-50 `555a72c` — AND
  **THE PHASE-4 CLOSER LANDED AT `7aea86c` (SESSION-026): the 30-link
  offline golden runner equivalent, the exit gate's last named
  artifact**. The generator (`testdiff/goldens/dump_goldenwebsites.py`)
  drives the UNMODIFIED reference's own `tools/run_golden_websites.py
  --offline` (CI's mode) over its own `tests/golden/websites.json`
  corpus — the fake LLM answering from the golden file, canned pages
  feeding the fetch, a throwaway vault, clocks frozen at 2026-10-17
  12:00:00, temp roots + the app dir folded by REPLACE — into
  `internal/websitepipeline/testdata/goldenwebsites.golden.json`
  (94,979 bytes; the generator runs its payload twice and refuses to
  write unless byte-identical); the Go replay
  (`internal/websitepipeline/goldenrunner_test.go`) runs the ported
  pipeline over the SAME corpus with the same seams and byte-compares
  the canned pages, the 90-call LLM choreography + the first three
  prompts, the per-link results, the 30 produced notes, the report,
  the console, and the exit-code law — **GREEN ON THE FIRST RUN, zero
  port fixes needed**; the mutation test (one golden byte) fails
  precisely. **The exit gate's all 10 named differential stones are
  green (the stone table: docs/sessions/SESSION-025.md §gate; the
  CLOSED stamp: 12-migration-roadmap.md §Phase 4; the README declares
  Phase 4 CLOSED). 1828/1828 tests race-clean over 30 packages.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-026 pushed the closer the moment it went green. The
  recovery order for ANY mid-task death stays: (1) `git status` +
  `git log --oneline -5` + file timestamps to learn what ACTUALLY
  survived — never trust a handoff summary's claim over `git log`;
  (2) a preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **twenty full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth, SESSION-025's nineteenth, SESSION-026's twentieth), plus
  SESSION-010's mid-session mode-flip, plus SESSION-021's CONTEXT
  death with the sandbox intact. The restore recipes: start-of-session
  (clone `assadigit/second-brain-project` with the PAT passed
  ephemerally as the Basic `x-access-token:<PAT>` header — locate the
  work repo via the GitHub API's authenticated /user/repos listing if
  the clone looks wrong (Currena is an UNRELATED sibling); never store
  the PAT — AND SCRUB IT: `git clone -c http.extraheader=…` PERSISTS
  the PAT into the new repo's .git/config, so run `git config
  --unset-all http.extraheader` in BOTH clones right after cloning,
  then push with the ephemeral `-c` form only); GitCurator @
  `98f4fc7` (verify against `testdiff/PINNED-GITCURATOR-REV`),
  go1.27.2 → `/home/z/.local/go` (go.dev tarball), venv `/home/z/
  venvs/gitcurator` (Python 3.12, stdlib-only), symlinks `/home/z/
  second-brain-project` and `/home/z/GitCurator`; mid-session
  mode-flip (`git checkout-index -f -a` restores modes from the index;
  `git fsck` to sanity-check). Verify the pins at every start: `git
  rev-parse HEAD:go.mod` — expect the v1.36.1/tls-client/1.24.1 bytes
  (do NOT bump sqlite; do NOT bump tls-client past v1.16.0). **Do NOT
  run `go mod download all`** (it churns go.sum; plain `go mod
  download` + `go build`/`go test` work without it). **NEW
  (SESSION-026): detached background processes DO NOT SURVIVE
  tool-call boundaries in this sandbox — the first two baseline
  attempts were reaped mid-module-download; run long suites in
  FOREGROUND CHUNKS per tool call (they fit comfortably inside the
  10-minute timeout with the warmed cache).** **The renderer-escape +
  cd-discipline laws stay: the two-character pairs [h and [m are eaten
  on the OUTPUT path, the INPUT path, AND THE BASH-ARG path — write
  commit messages to a FILE and commit with -F, then verify the
  committed bytes with 'git log --format=%B' before pushing; note the
  DISPLAY of test diffs eats the pairs too (grep the FILE bytes before
  believing a corruption claim). AND the cd-discipline law: the shell
  RESETS its working directory between tool calls — EVERY git command
  carries its own cd; never commit outside
  /home/z/my-project/second-brain-project.**

## 2. Next session: SESSION-027 — Phase 5 OPENS — T-60 `internal/cli` (the engine parity milestone)

- **T-60 `internal/cli` (L)** — the CLI surface per
  `docs/port-blueprint/module-specs/spec-cli.md` (**READ THE SPEC
  BEFORE THE CODE** — its banner records the Option-B re-scope
  ripples: the `--sync` CHECK and `--test-connection` live leg lose
  the Telegram part — the queue check rides the Worker only; the
  Telethon-only flags (`--mark-read`, `--from-id/--to-id`,
  `--offset-start/--count`, `--single-id`, `--init`'s interactive
  Telegram login offer) are accepted losses — T-60 decides the honest
  flag surface, never a silent no-op; the batch-end mark-read step
  falls away — mirror-push is the ack). The surface:
  `internal/cli` (the Phase-0 seed — the banner + the --status
  "no config" path — already exists) + `cmd/gitcurator-cli`:
  the full flag surface (init → login → status → test-connection →
  detect-llm → list-dead → reset-dead → hand-delivery → chrome-retry
  → mark-read → auto → sync → retry-failed → manual batch; no flag →
  the friendly help card), the wizard (`--init` — the repo radar
  before saving; collects the Telegram keys for compatibility but
  never offers the MTProto login), the status card (the config
  summary rows, the sentinel probes, the DB stats — never creates
  state), `--auto` + `--sync` (cli_sync.py's CHECK → THE BUTTON →
  SYNC doors, the wrong-vault guard, the cron law), the visual run
  engine's CLI-relevant halves (cli_run.py's dry-run report + the
  SIGINT/exit-code law — the QCoreApplication wiring maps to
  `internal/app` commands, no Qt anywhere), cli_settings.py (the
  model menu, the config save/mirror, `_mask_token`,
  `_clean_config_for_save`), cli_terminal.py (the C ANSI class +
  NO_COLOR/TTY law, `_harden_stdio`'s UTF-8, term_width/mask/rule/
  cli_print/banner/spinner/render_bar/StatusLine). **The mirror/tool
  legs — mirror_manual, add_recall_hooks, build_links, the status
  row — are T-60's books per the T-50 session's record.** The
  launchers' Go equivalents ride `cmd/gitcurator-cli` (ASCII/CRLF
  laws apply; the python-discovery walk-up becomes unnecessary but
  harmless).
- **The spec's Tests section is the acceptance**: flag-surface
  snapshot (help text), wizard flow (fake stdin), status card
  rendering, `--sync` door beats (Enter/other/`--yes`/non-TTY),
  wrong-vault guard, the Phase-5 gating law, dry-run report equality
  (differential case), Ctrl+C exit code, terminal paint
  (ANSI/NO_COLOR), UTF-8 hardening. A generated golden per the house
  pattern (a dump_cli.py driving the reference's own CLI shapes with
  frozen clocks + folded paths) is the differential stone.
- **T-62's bounded open item stands (do not improvise it)**: the
  all-clear gate's Phase-5 re-spec (what `tracker.get_all_clear()`
  now gates without the msg-id cursor/mark-read steps) — bounded by
  ADR-0004 Am.3; recorded in spec-cli.md's dated note and the T-50
  session's record.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; the dry-run seam at call sites; rows collected before
  nested statements; tests never weakened; reality wins over specs
  (fix the spec in the same commit); goldens are GENERATED by running
  the unmodified Python app, never hand-typed; no per-run facts in
  goldens; the frozen-clock seams (`clockNow` per package). T-41's:
  text-mode reads translate universal newlines before any cap or
  parse; invalid UTF-8 decodes with the maximal-subpart U+FFFD law;
  vault walks are os.walk's files-before-subdirs order. T-42's: the
  house JSON decoder's shapes; floats render through
  helpers.PyFloatRepr. T-43's: `pyBytesBlank` is the blank-body law;
  Go MAP ranges are RANDOM. T-44's: Go's `url.Parse` is NOT CPython's
  urlsplit; `cmd.Wait()` closes the pipe; scripted timing tests keep
  the reference's UNPATCHED cadences. T-45's: THE POINTER-THREADING
  LAW ([]*NoteRow); keep BOTH frontmatter parsers; Python text-mode
  read(n) counts CHARACTERS; the _windows_eyes law rides a stat seam.
  T-46's: the WALK-ORDER FOLD; bounded reads count runes; None doors
  type-narrow to the zero value; a revive-style table iteration rides
  sorted keys; dry-run asymmetries preserved. T-47's: THE REDACTION
  LAW; the token NEVER touches .git/config; FETCH_HEAD rewritten
  after every token-URL fetch; env scrubbing + HideWindow where the
  REFERENCE's own module scrubs; frozen GIT DATES make commit shas
  deterministic; the reference's mock-surface maps to instance seams.
  T-48's: the apiFn/tokenURLFn instance seams; the ReadDir-order
  convergence; atexit maps to an exported Cleanup; the base-dict
  freeze quirk; the frozen-dates sha trick. T-49's: the []-vs-null
  byte law; `storage.PyDumpsUTF8`; dictDoor presence checks + the
  double-wrapping trap; the silent unknown-type failure; the
  set-iteration convergence; the push_mirror env-inheritance law;
  the CPython re vs Go regexp convergence; the lstrip character-set
  quirk; the raw-tree banished round-trip; the importlib
  fixture-import generator trick. T-50's: the map-order law BITES THE
  LOGS; the []-vs-null law extends to .get() dumps; the house JSON
  decoder returns int64 for integral literals; str.splitlines'
  boundary set is wider than \n (pySplitLines is the twin);
  walk-order independence is earned by sorting; the clockNow seam
  covers every time.strftime write; the reference's exception-text
  rendering stays unpinnable (the SHAPE is the contract).
  **SESSION-026's additions: the offline-golden-runner's own quirks
  are now pinned Go notes (the w03 analysis prompt carries no
  'Website: ' line — a fake-LLM URL scan keyed on that line silently
  defaults; Python str.capitalize lowercases the REST); the generator
  that drives a reference TOOL (not a module) imports the tool via
  importlib and stubs its cleanup (rmtree records, never deletes —
  then the generator cleans up with the real shutil); detached
  background processes do not survive tool-call boundaries — run
  long suites in foreground chunks.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-026.md` — the most recent report (the
   Phase-4 closer's evidence, the gate's CLOSED stamp, the
   twentieth-wipe rebuild, the foreground-chunks operational find).
2. `docs/port-blueprint/module-specs/spec-cli.md` — THE SPEC (the
   full surface + the Option-B re-scope notes + the T-62 bounded
   open item + the Tests section = the acceptance).
3. The reference's CLI sources (in order): `gitcurator/cli/cli.py`
   (the command surface + dispatch), `gitcurator/cli/cli_sync.py`
   (the CHECK/BUTTON/SYNC doors), `gitcurator/cli/cli_run.py` (the
   visual run engine — the dry-run report + the SIGINT law),
   `gitcurator/cli/cli_settings.py`, `gitcurator/cli/cli_terminal.py`
   — plus the existing Phase-0 seed `internal/cli/cli.go` (the
   banner + the --status "no config" path, byte-exact, verified by
   testdiff/cases/status-no-config).
4. The working precedents: `testdiff/goldens/dump_goldenwebsites.py`
   (the newest generator — driving a reference TOOL via importlib +
   the cleanup stub + the two-run determinism proof built in),
   `internal/websitepipeline/goldenrunner_test.go` (the replay-harness
   pattern: the Go twins of the runner's own logic + t.Run
   subtests + the mutation-teeth habit).

## 4. Known constraints / owner actions for SESSION-027

- **The go-directive acknowledgement (STILL WAITING — sixteenth
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
  extraheader from .git/config after the clone — see §1); goldens are
  GENERATED by running its unmodified code (`PYTHONDONTWRITEBYTECODE=1`
  keeps it pristine).
- **Dependency law**: `modernc.org/sqlite` pinned at v1.36.1;
  `bogdanfinn/tls-client` pinned at v1.16.0; `golang.org/x/net`
  v0.48.0 + `golang.org/x/text` v0.32.0 (direct); ANY other add needs
  a new ADR — surface, don't improvise. (SESSION-026 added zero.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers — the CLI routing legs land WITH T-60 per the spec's own
  Tests section)**: test_noteredo's remaining un-ported selections —
  ScanMasterRedoRows (zero dedicated Go tests — the top item), the
  redo-consume laws beyond the golden section, the proper-note laws,
  the pipeline-redo legs; the GUI wiring legs stay the GUI worker
  shell task's books. Also standing: the note_state SURGERY path's
  CRLF round-trip differential (Phase 4+); the Windows live legs
  (Phase 7 clean-machine checklist); a Windows VM without WebView2
  for T-70 (optional).
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-027)

```
You are starting SESSION-027 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty full wipes across
   SESSION-005..026 plus mid-session degradations; SESSION-021 added the
   CONTEXT-death class — the working context ended AFTER the T-46 commit
   existed but BEFORE the push+close; the continuation recovered by trusting
   git evidence over the lossy handoff summary, auditing the preserved commit
   with the full battery, then pushing). FIRST verify with evidence: work repo
   at /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-
   project), read-only reference /home/z/my-project/GitCurator @ 98f4fc7
   (symlinked /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 → /home/z/.local/go, venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, and go.mod with
   modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not
   bump any of these). Then check `git status` and `git log --oneline -3`:
   SESSION-026 landed the Phase-4 closer at 7aea86c on top of SESSION-025's
   close 7301f19 (the tree must be clean and origin/main at the SESSION-026
   close; the baseline is 1828 tests race-clean over 30 packages — re-verify
   the suite before work if anything looks off). If anything is missing,
   rebuild per docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you
   ask — pass it ephemerally with git -c http.extraheader="Authorization:
   Basic $(printf 'x-access-token:%s' "$PAT" | base64)", the Basic scheme;
   never store it — AND SCRUB IT: `git clone -c http.extraheader=…` PERSISTS
   the header into the new repo's .git/config, so run `git config --unset-all
   http.extraheader` in BOTH clones right after cloning; a mid-session
   mode-flip is repaired with `git checkout-index -f -a` after `git fsck`;
   and if the sandbox or the context dies MID-TASK, learn what actually
   survived from git evidence — `git status` + `git log --oneline -5` + file
   timestamps — before restarting from zero: a preserved commit or draft is
   continued and audited, never blindly trusted, and a handoff summary's
   claim never outranks `git log`'s). SESSION-026's operational finds: the
   generator that drives a reference TOOL imports it via importlib and stubs
   its cleanup (rmtree records, never deletes); detached background processes
   DO NOT SURVIVE tool-call boundaries in this sandbox — run long test suites
   in FOREGROUND CHUNKS per tool call. The older laws stand: (a) THE
   CD-DISCIPLINE LAW — the shell resets its working directory between tool
   calls, so every git/go command carries its own cd (or go's -C flag); never
   commit outside /home/z/my-project/second-brain-project; (b) the
   renderer-eats-escape-pairs lesson — the two-character pairs [h and [m are
   eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path: avoid
   those spellings, or write commit messages to a FILE and commit with -F,
   then verify the committed bytes with 'git log --format=%B' before pushing.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (Phase 4 is CLOSED — the closer's
   record is in the status) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records the Phase-4 close: the
   closer at 7aea86c, 1828 tests race-clean over 30 packages, the environment
   recipes incl. the foreground-chunks find, the clone-config PAT-scrub
   hazard, the renderer + cd-discipline lessons, and the standing port laws
   incl. SESSION-026's additions) → docs/sessions/SESSION-026.md (the most
   recent report — the closer's evidence + the gate's CLOSED stamp).

3. THIS SESSION OPENS PHASE 5: T-60 internal/cli per
   docs/port-blueprint/module-specs/spec-cli.md — READ THE SPEC BEFORE THE
   CODE (its banner records the Option-B re-scope ripples: the Telethon-only
   flags are accepted losses — T-60 decides the honest flag surface, never a
   silent no-op; the queue check rides the Worker only; mirror-push is the
   ack). The mirror/tool legs — mirror_manual, add_recall_hooks,
   build_links, the status row — are T-60's books per the T-50 session's
   record. The existing Phase-0 seed (internal/cli/cli.go — the banner +
   the --status "no config" path, verified by testdiff/cases/
   status-no-config) is the floor. The spec's Tests section is the
   acceptance (flag-surface snapshot, wizard flow, status card, --sync door
   beats, wrong-vault guard, the Phase-5 gating law, dry-run report
   equality, Ctrl+C exit code, terminal paint, UTF-8 hardening); a generated
   golden per the house pattern (a dump_cli.py driving the reference's own
   CLI shapes with frozen clocks + folded paths) is the differential stone —
   testdiff/goldens/dump_goldenwebsites.py (driving a reference TOOL via
   importlib + the cleanup stub + the two-run determinism proof) is the
   newest precedent. T-62's bounded open item stands (the all-clear gate's
   Phase-5 re-spec — bounded by ADR-0004 Am.3; do not improvise it). If the
   task will not fit the session's capacity, land its first coherent slice
   properly with its own differential proof and record the remainder honestly
   — one slice done properly with proof beats a rushed whole; close cleanly
   at a natural boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
   Phase-5 entry is already earned — Phase 4 is CLOSED; evaluate T-60's own
   acceptance honestly), commit + push everything (ask me for the GitHub PAT
   if it is not already available — never store it anywhere), write
   docs/sessions/SESSION-027.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-028.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

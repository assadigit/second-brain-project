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
  SHELL at `143c590` (SESSION-027: the paint box, the settings
  halves, the sync door's pure halves + beat 1a, the honest Option-B
  flag surface + refusal cards, the FULL --status card,
  --list-dead/--reset-dead); slice 2 — THE COMMAND WIRINGS at
  `41d699e` (SESSION-028: the --init wizard with the repo radar over
  the RadarAll seam, --test-connection over the ported battery with
  the _section/_check_line renderers and every Option-B divergence
  asserted, --detect-llm with the model menu + cli_settings's
  model-menu half, --hand-delivery complete, --chrome-retry's listing
  + full dry-run rehearsal with the honest T-62 boundary card for the
  live leg; a reality-wins find fixed in-commit — the reference's
  chrome-retry --dry-run forgot to flip the dry-run switch). **The
  differential golden: testdiff/goldens/dump_cli.py → internal/cli/
  testdata/goldencli.golden.json (48,698 bytes, two-run
  byte-identical; 9 more sections incl. the wizards' saved-config
  bytes + the enqueue's writes) → goldencli_test.go replays every
  ported section byte-for-byte with the Option-B divergences asserted
  and TestGoldenCLIMutationWirings proving the teeth. 1866 tests
  race-clean over 31 packages (+14).**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-028 pushed the slice the moment the battery went
  green (`fc9cd20..41d699e`). The recovery order for ANY mid-task death
  stays: (1) `git status` + `git log --oneline -5` + file timestamps
  to learn what ACTUALLY survived — never trust a handoff summary's
  claim over `git log`; (2) a preserved commit or draft is continued
  and audited (the full battery re-run), never blindly trusted, never
  restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **twenty-two full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth, SESSION-025's nineteenth, SESSION-026's twentieth,
  SESSION-027's twenty-first, SESSION-028's twenty-second), plus
  SESSION-010's mid-session mode-flip, plus SESSION-021's CONTEXT
  death with the sandbox intact. The restore recipes: start-of-session
  (clone `assadigit/second-brain-project` with the PAT passed
  ephemerally as the Basic `x-access-token:<PAT>` header — locate the
  work repo via the GitHub API's authenticated /user/repos listing if
  the clone looks wrong (Currena is an UNRELATED sibling); never store
  the PAT — AND SCRUB IT: `git clone -c http.extraheader=…` PERSISTS
  the PAT into the new repo's .git/config, so run `git config
  --unset-all http.extraheader` in BOTH clones right after cloning,
  then push with the ephemeral `-c` form only); GitCurator @ `98f4fc7`
  (verify against `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 →
  `/home/z/.local/go` (go.dev tarball, sha256 `ecbadb99…`), venv
  `/home/z/venvs/gitcurator` (Python 3.12, stdlib-only), symlinks
  `/home/z/second-brain-project` and `/home/z/GitCurator`; mid-session
  mode-flip (`git checkout-index -f -a` restores modes from the index;
  `git fsck` to sanity-check). Verify the pins at every start: `git
  rev-parse HEAD:go.mod` — expect the v1.36.1/tls-client/1.24.1 bytes
  (do NOT bump sqlite; do NOT bump tls-client past v1.16.0). **Do NOT
  run `go mod download all`** (it churns go.sum; plain `go mod
  download` + `go build`/`go test` work without it). **Detached
  background processes DO NOT SURVIVE tool-call boundaries — run long
  suites in FOREGROUND CHUNKS per tool call.** **The renderer-escape +
  cd-discipline laws stay: the two-character pairs [h and [m are eaten
  on the OUTPUT path, the INPUT path, AND THE BASH-ARG path — write
  commit messages to a FILE and commit with -F, then verify the
  committed bytes with 'git log --format=%B' before pushing; note the
  DISPLAY of test diffs eats the pairs too (the golden replay compares
  JSON-decoded BYTES — display mangling never enters a comparison).
  AND the cd-discipline law: the shell RESETS its working directory
  between tool calls — EVERY git/go command carries its own cd; never
  commit outside /home/z/my-project/second-brain-project.** **The
  Write/Edit tool path REDACTS token-shaped strings at file-write time
  — never put a token-shaped string in a file you write, even a fake
  one (craft safe-shaped test tokens like `fake-github-token-…`;
  secret-scan the tree with a byte-level grep before every commit).**
  **NEW (SESSION-028, golden-side recipe): the golden generator's env
  record is dumped with json.dumps(sort_keys=True) — every RECORDED
  config loses its insertion order; any crafted config whose SAVED
  ORDER matters (the wizard's, detect's) must be written from an
  ALPHABETICAL literal in the generator so the replay's env-decoded
  file matches the reference's bytes; and the wizard's scripted stdin
  lives in the env record FOLDED — the replay must UNFOLD the script
  before feeding it.**

## 2. Next session: SESSION-029 — T-62 — the run engine

- **T-62 (L)** — the visual run engine + the batch doors, per
  spec-cli.md's cli_run.py section + the SESSION-028 shipped record:
  1. **`run_batch_visual`** (cli_run.py 239-504): wires the 8 worker
     signals (the Go port composes them synchronously over the ported
     githubpipeline/websitepipeline engines — no Qt), the batch
     drivers' wiring (cmd_auto 1252-1440, cmd_sync's live beats 1b/2/3
     over the Worker queue, cmd_manual_batch/--retry-failed/
     --import-file), the wrong-vault guard (1374-1409), the optional
     >10 confirm (skipped by --yes), the batch-end seals (GitHub seal
     → GoodRepos publish → websites seal → manual seal — all ported,
     wire them), **the Phase-5 CLEAR gate — its exact re-spec (what
     the all-clear now gates with the msg-id cursor and mark-read
     gone) is T-62's bounded open item per ADR-0004 Am.3 — resolve it
     HERE with the evidence, never improvised**.
  2. **The preflights** (cli_run.py 565-719): `_preflight_github`
     (the token check — the connectioncheck GitHub seam is the twin),
     `_preflight_proxy` (the TCP reach), `_preflight_llm` (the probe +
     THE interactive model picker — cli.PromptModelMenu + SmartModelPick
     + ApplyModelChoice are LANDED and tested; wire them into the
     preflight's flow), `preflight_checks` + `run_config_card`
     (524-562 — the boxed run-configuration snapshot; port it into the
     paint box).
  3. **The dry-run report writer** (`_write_dryrun_report`, 215-236 +
     dryrun.render_report_markdown) + the SIGINT handler + the
     100 ms heartbeat + the exit-code law (the batch's success flag).
  4. **The chrome-retry live leg** (the preflight → the real Chrome
     delivery → run_batch_visual processing the delivered pages in the
     same run) — the engines are ported; the wiring follows.
  5. **T-61's remaining paint legs** riding along: the spinner under a
     real loop (the StatusLine grammar is ported; drive it from the
     engine's phases), the Windows console live test stays Phase 7's
     clean-VM checklist.
- **The differential stone grows with the slice**: extend
  testdiff/goldens/dump_cli.py (the run-config card over crafted
  configs; the preflight cards over the seam answers; the batch dry-run
  report over the recorded dryrun entries — the reference side drives
  its own engine with the same seams the T-42/T-44/T-50 goldens
  already proved) and goldencli_test.go replays them — the two-run
  byte-identical proof + the mutation teeth stay part of every
  regeneration.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1);
  `store.PyISOFormat*` for timestamp writes; the dry-run seam at call
  sites; rows collected before nested statements; tests never
  weakened; reality wins over specs (fix the spec in the same commit);
  goldens are GENERATED by running the unmodified Python app, never
  hand-typed; no per-run facts in goldens; the frozen-clock seams
  (`clockNow` per package; handdelivery.Clock). T-41's: text-mode
  reads translate universal newlines before any cap or parse; invalid
  UTF-8 decodes with the maximal-subpart U+FFFD law; vault walks are
  os.walk's files-before-subdirs order. T-42's: the house JSON
  decoder's shapes; floats render through helpers.PyFloatRepr.
  T-43's: `pyBytesBlank` is the blank-body law; Go MAP ranges are
  RANDOM. T-44's: Go's `url.Parse` is NOT CPython's urlsplit;
  `cmd.Wait()` closes the pipe; scripted timing tests keep the
  reference's UNPATCHED cadences. T-45's: THE POINTER-THREADING LAW
  ([]*NoteRow); keep BOTH frontmatter parsers; Python text-mode
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
  CPython re vs Go regexp convergence; the lstrip character-set quirk;
  the raw-tree banished round-trip; the importlib fixture-import
  generator trick. T-50's: the map-order law BITES THE LOGS; the
  []-vs-null law extends to .get() dumps; the house JSON decoder
  returns int64 for integral literals; str.splitlines' boundary set is
  wider than \n (pySplitLines is the twin); walk-order independence is
  earned by sorting; the clockNow seam covers every time.strftime
  write; the reference's exception-text rendering stays unpinnable
  (the SHAPE is the contract). SESSION-026's: the offline-golden-
  runner's quirks are Go notes; the generator that drives a reference
  TOOL imports it via importlib and stubs its cleanup; long suites run
  in foreground chunks. SESSION-027's: the gui.app generator door stubs
  to the REAL underlying classes; APP_DIR binds at cache_db import —
  patch constants BEFORE the import; the dead-level-color quirk; the
  optional-int flags resolve present-vs-absent AFTER parse; the
  Write-tool token redaction + the safe-shaped crafted-token law.
  **SESSION-028's: the generator's fake PyQt6/PyGithub modules satisfy
  worker_jobs/WorkerLlmMixin's import guards so the reference's own
  picker runs; the scripted stdin folds — unfold before feeding; the
  env record's configs are sort_keys-ordered — craft any
  saved-order-sensitive config from an ALPHABETICAL literal; the
  battery's net doors inject through the exported connectioncheck
  seams (the reference's own monkeypatch surface); the dry-run switch
  must be RESTORED after a rehearsal (a leaked switch silently
  no-ops every later write in the same test binary); the reference's
  chrome-retry --dry-run never flips the switch — the port's fix is
  the recorded reality-wins.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-028.md` — the most recent report (the
   wirings' evidence, the golden's growth, the reality-wins find, the
   twenty-second-wipe rebuild).
2. `docs/port-blueprint/module-specs/spec-cli.md` — THE SPEC (the
   banner + the SESSION-027/028 shipped records + the cli_run.py
   section = T-62's contract + the Tests section = the acceptance).
3. The reference's run-engine sources (in order):
   `gitcurator/cli/cli_run.py` in full (run_batch_visual 239-504,
   the signals' wiring 318-341, the finish seals 383-417, the
   Phase-5 clear 425-469, the dry-run report 215-236, the SIGINT law
   477-494, the preflights 507-719); `gitcurator/cli/cli.py`
   cmd_auto (1252-1440) + cmd_manual_batch + cmd_retry_failed +
   cmd_sync's live beats in `gitcurator/cli/cli_sync.py`;
   `gitcurator/core/dryrun.py` (render_report_markdown); the GUI's
   batch-end seal calls in `gitcurator/gui/processing_worker.py`
   (the finish sequence — the ported vaultseal/goodrepos engines'
   call shapes).
4. The landed slices (the patterns to extend): `internal/cli/*.go`
   (the seams' inventory — the PromptModelMenu bridge is READY),
   `testdiff/goldens/dump_cli.py` (the generator's seam inventory +
   the SESSION-028 recipes in the docstring),
   `internal/cli/goldencli_test.go` (the replay harness + the fold/
   unfold + the mutation teeth), and the T-26 offline runner's
   golden (`testdiff/goldens/dump_goldenwebsites.py` +
   `internal/websitepipeline/goldenrunner_test.go`) — the batch
   engine's differential precedent.

## 4. Known constraints / owner actions for SESSION-029

- **The go-directive acknowledgement (STILL WAITING — eighteenth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers x/net/html,
  approve an ADR and a future session swaps `internal/webextract/
  parser.go`'s scan internals).
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
  a new ADR — surface, don't improvise. (SESSION-027/028 added zero —
  the Windows console calls ride syscall.NewLazyDLL.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections
  (ScanMasterRedoRows is the top item — note the CLI routing legs
  landed with T-60/T-62 per the spec's own Tests section); the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs (Phase 7 clean-machine checklist); a Windows
  VM without WebView2 for T-70 (optional).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit — the Write-tool redaction is a safety NET, not
  a license); tests never weakened; reality wins over specs (fix the
  spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-029)

```
You are starting SESSION-029 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-two full wipes across
   SESSION-005..028 plus mid-session degradations; SESSION-021 added the
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
   SESSION-028 landed T-60's remainder — THE COMMAND WIRINGS — at 41d699e on top
   of SESSION-027's CLI shell (the tree must be clean and origin/main at
   41d699e; the baseline is 1866 tests race-clean over 31 packages — re-verify
   the suite before work if anything looks off). If anything is missing, rebuild
   per docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass
   it ephemerally with git -c http.extraheader="Authorization: Basic $(printf
   'x-access-token:%s' "$PAT" | base64)", the Basic scheme; never store it — AND
   SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the header into the new
   repo's .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; locate the work repo via the GitHub API's
   authenticated /user/repos listing when unsure — Currena is an UNRELATED
   sibling; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`). SESSION-028's operational finds: the golden generator's
   env record is dumped sort_keys=True — every recorded config loses its
   insertion order, so craft any saved-order-sensitive config (the wizard's,
   detect's) from an ALPHABETICAL literal in the generator; the scripted stdin
   lives in the env record FOLDED — the replay must UNFOLD it before feeding;
   the dry-run switch must be RESTORED after a rehearsal (a leaked switch
   silently no-ops every later write in the same test binary). The older laws
   stand: (a) THE CD-DISCIPLINE LAW — the shell resets its working directory
   between tool calls, so every git/go command carries its own cd (or go's -C
   flag); never commit outside /home/z/my-project/second-brain-project; (b) the
   renderer-eats-escape-pairs lesson — the two-character pairs [h and [m are
   eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path: avoid those
   spellings, write commit messages to a FILE and commit with -F, then verify
   the committed bytes with 'git log --format=%B' before pushing, and never
   trust a DISPLAY of escape-bearing bytes (grep the FILE bytes); (c) the
   Write/Edit tool REDACTS token-shaped strings at file-write time (craft
   safe-shaped test tokens, never ghp_-shaped ones; byte-level secret-scan
   before every commit); (d) detached background processes DO NOT SURVIVE
   tool-call boundaries — run long suites in FOREGROUND CHUNKS per tool call.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (T-60 LANDED across two sessions — the
   slices' records are in the status) → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records SESSION-028's
   close: the command wirings at 41d699e, 1866 tests race-clean over 31
   packages, the environment recipes, the newest lessons incl. the
   sort_keys/fold/unfold and dry-run-switch-restore recipes) →
   docs/sessions/SESSION-028.md (the most recent report — the wirings'
   evidence + the golden's growth + the reality-wins find).

3. THIS SESSION IS T-62 — THE RUN ENGINE: per
   docs/port-blueprint/module-specs/spec-cli.md's cli_run.py section + the
   SESSION-028 shipped record (READ THE SPEC + THE LANDED SLICES BEFORE THE
   CODE): run_batch_visual (the batch drivers' wiring — --auto's engine,
   --sync's live beats over the Worker queue, --retry-failed,
   --import-file/manual batch; the wrong-vault guard; the >10 confirm; the
   batch-end seals — the vaultseal/goodrepos/website/manual engines are all
   ported, wire their calls), THE PHASE-5 CLEAR GATE — its exact re-spec with
   the msg-id cursor and mark-read gone is T-62's bounded open item per
   ADR-0004 Am.3 — resolve it HERE with the evidence, never improvised; the
   preflights (_preflight_github/_preflight_proxy/_preflight_llm —
   cli.PromptModelMenu + SmartModelPick + ApplyModelChoice are LANDED and
   tested, wire them into the preflight's flow; run_config_card); the dry-run
   report writer (dryrun.render_report_markdown); the SIGINT handler + the
   heartbeat + the exit-code law; the chrome-retry live leg (the preflight →
   the real Chrome delivery → the batch processing in the same run); T-61's
   remaining paint legs riding along (the spinner under a real loop). Extend
   the differential stone with the slice: testdiff/goldens/dump_cli.py grows
   the new sections (the run-config card, the preflight cards over the seam
   answers, the batch dry-run report over the recorded entries — the reference
   side drives its own engine with the seams the T-42/T-44/T-50 goldens
   already proved) and internal/cli/goldencli_test.go replays them — the
   two-run byte-identical proof + the mutation teeth stay part of every
   regeneration. If the task will not fit the session's capacity, land the
   first coherent engine slice properly with its own differential proof and
   record the remainder honestly — one slice done properly with proof beats a
   rushed whole; close cleanly at a natural boundary when capacity runs out.

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
   docs/sessions/SESSION-029.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-030.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

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
- **Phase 4 CLOSED (SESSION-015–026)**: the task list T-40..T-50 ALL
  LANDED + the Phase-4 closer at `7aea86c` (the 30-link offline golden
  runner equivalent — the exit gate's all 10 differential stones green;
  1828 tests race-clean over 30 packages at the close).
- **PHASE 5 OPENED (SESSION-027): T-60's first slice — THE CLI SHELL —
  LANDED at `143c590`**: the paint box (terminal.go), the settings
  halves (settings.go), THE SYNC DOOR's pure halves + beat 1a
  (syncdoor.go), the honest Option-B flag surface + dispatch + refusal
  cards for the Telethon accepted losses (cli.go — T-60's
  honest-surface DECISION recorded in the spec banner), the FULL
  --status card (status.go — 16 config rows + taxonomy + the sentinel
  rows + the Ollama/llama.cpp checks + the CacheDB/NoteStateDB/
  WebsiteStateDB stats), --list-dead/--reset-dead (dead.go), and the
  cfgutil or-door laws. **The differential golden:
  testdiff/goldens/dump_cli.py → internal/cli/testdata/
  goldencli.golden.json (two-run byte-identical) → goldencli_test.go
  byte-compares every ported section (the replay green after two
  one-token fixes; the mutation test bites; the reference's no-flag
  help card kept as the pre-Option-B RECORD — the divergence asserted,
  never silent). 1852 tests race-clean over 31 packages (+24).** The
  Phase-0 case status-no-config still PASSES through the reworked card.
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-027 pushed the slice the moment the battery went
  green (`2cc07d1..143c590`). The recovery order for ANY mid-task death
  stays: (1) `git status` + `git log --oneline -5` + file timestamps
  to learn what ACTUALLY survived — never trust a handoff summary's
  claim over `git log`; (2) a preserved commit or draft is continued
  and audited (the full battery re-run), never blindly trusted, never
  restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **twenty-one full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth, SESSION-025's nineteenth, SESSION-026's twentieth,
  SESSION-027's twenty-first), plus SESSION-010's mid-session
  mode-flip, plus SESSION-021's CONTEXT death with the sandbox intact.
  The restore recipes: start-of-session (clone `assadigit/
  second-brain-project` with the PAT passed ephemerally as the Basic
  `x-access-token:<PAT>` header — locate the work repo via the GitHub
  API's authenticated /user/repos listing if the clone looks wrong
  (Currena is an UNRELATED sibling); never store the PAT — AND SCRUB
  IT: `git clone -c http.extraheader=…` PERSISTS the PAT into the new
  repo's .git/config, so run `git config --unset-all http.extraheader`
  in BOTH clones right after cloning, then push with the ephemeral
  `-c` form only); GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball, sha256 `ecbadb99…`), venv `/home/z/venvs/gitcurator`
  (Python 3.12, stdlib-only), symlinks `/home/z/second-brain-project`
  and `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  Verify the pins at every start: `git rev-parse HEAD:go.mod` — expect
  the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do NOT bump
  tls-client past v1.16.0). **Do NOT run `go mod download all`** (it
  churns go.sum; plain `go mod download` + `go build`/`go test` work
  without it). **Detached background processes DO NOT SURVIVE
  tool-call boundaries — run long suites in FOREGROUND CHUNKS per tool
  call.** **The renderer-escape + cd-discipline laws stay: the
  two-character pairs [h and [m are eaten on the OUTPUT path, the INPUT
  path, AND THE BASH-ARG path — write commit messages to a FILE and
  commit with -F, then verify the committed bytes with 'git log
  --format=%B' before pushing; note the DISPLAY of test diffs eats the
  pairs too (the golden replay compares JSON-decoded BYTES — display
  mangling never enters a comparison). AND the cd-discipline law: the
  shell RESETS its working directory between tool calls — EVERY git
  command carries its own cd; never commit outside
  /home/z/my-project/second-brain-project.** **NEW (SESSION-027): the
  Write/Edit tool path REDACTS token-shaped strings at file-write time
  (a crafted fake ghp_… test token became a placeholder in a generator
  source; the display path redacts on read-back too) — never put a
  token-shaped string in a file you write, even a fake one (craft
  safe-shaped test tokens like `fake-github-token-…`; secret-scan the
  tree with a byte-level grep before every commit).**

## 2. Next session: SESSION-028 — T-60's remainder — the command wirings

- **T-60 remainder (L)** — per spec-cli.md's SESSION-027 shipped
  record, the list that remains:
  1. **`--init` wizard** (cli.py cmd_init): the _ask sequence (vault
     with the empty-start law + the isdir retry loop, the Telegram
     trio + bot + token compat collection — collected but never used
     per Option-B, the two other vaults, the LLM provider branches
     (ollama/llamacpp-with-probe-seam/cloud), the proxy block, the
     threshold clamp 2–10), **the repo radar BEFORE saving** (internal/
     reporadar is ported — the radar legs ride seams), save_config,
     the summary card, and NO interactive Telegram login offer
     (dropped per ADR-0004 Am.3 — the compat law: the wizard still
     COLLECTS the keys), the final "Setup complete" lines.
  2. **`--test-connection`** (cmd_test_connection): the battery wiring
     over the PORTED connectioncheck package (RunLocalChecks +
     Summarize + the CLI level mapping) + the _section/_check_line
     renderers (cli_run.py's own — port them into the paint box) —
     the Worker leg IS the queue check (Option-B; no Telegram leg).
  3. **`--detect-llm`** (cmd_detect_llm): probe → the model menu →
     set + save — **cli_settings's model-menu half lands with it**
     (smart_model_pick / prompt_model_menu / apply_model_choice; the
     menu is the numbered picker with the recommendation default).
  4. **`--hand-delivery` + `--chrome-retry`** (cmd_hand_delivery /
     cmd_chrome_retry): the engines are ported (internal/handdelivery,
     internal/chrometabs, the websitepipeline legs) — the CLI wiring
     follows (the walled-rows listing, the enqueue, the doors' honest
     cards; chrome-retry's run_batch_visual leg waits for T-62 — wire
     the listing + the T-62 boundary card honestly if the engine
     isn't there yet).
  5. **T-61's remaining paint legs** (the task list's S): the spinner
     animation under a real loop, the Windows console live test stays
     Phase 7's clean-VM checklist.
- **Then T-62 if capacity remains** (the --sync live beats + the
  --auto engine + the dry-run report writer + the SIGINT/exit-code
  law + the wrong-vault guard) — the T-62 bounded open item (the
  all-clear gate's Phase-5 re-spec, ADR-0004 Am.3) is resolved THERE,
  never improvised.
- **The differential stone grows with the slice**: extend
  testdiff/goldens/dump_cli.py (the wizard with scripted stdin + the
  radar stubbed; the battery with the net stubs on the reference side
  and the connectioncheck seams on the Go side; the detect-llm menu
  with the probe seams; the hand/chrome listings over crafted rows)
  and the replay in goldencli_test.go — the two-run byte-identical
  proof + the mutation teeth stay part of every regeneration.
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for any
  store-adjacent column; `store.PyISOFormat*` for timestamp writes;
  the dry-run seam at call sites; rows collected before nested
  statements; tests never weakened; reality wins over specs (fix the
  spec in the same commit); goldens are GENERATED by running the
  unmodified Python app, never hand-typed; no per-run facts in
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
  SESSION-026's: the offline-golden-runner's quirks are Go notes (the
  w03 analysis-prompt URL-scan miss; str.capitalize lowercases the
  REST); the generator that drives a reference TOOL imports it via
  importlib and stubs its cleanup (rmtree records, never deletes);
  long suites run in foreground chunks. **SESSION-027's: the gui.app
  door stubs to the REAL underlying classes (the import guards are
  environment shims, not logic); APP_DIR binds at cache_db import —
  patch constants BEFORE the import; the dead-level-color quirk
  (_LEVEL_STYLE freezes the pre-enable codes — the level colors are
  dead code in the reference, the port freezes them too); the
  optional-int flags resolve present-vs-absent AFTER parse; the
  Write-tool token redaction + the safe-shaped crafted-token law.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-027.md` — the most recent report (the
   slice's evidence, the golden's construction, the two one-token
   fixes, the twenty-first-wipe rebuild, the Write-tool redaction
   find).
2. `docs/port-blueprint/module-specs/spec-cli.md` — THE SPEC (the
   banner's honest-surface decision + the SESSION-027 shipped record
   with the remaining list + the Tests section = the acceptance).
3. The reference's CLI sources for the remainder (in order):
   `gitcurator/cli/cli.py` cmd_init (108-299) + cmd_test_connection
   (750-807) + cmd_detect_llm (814-976) + cmd_hand_delivery
   (1032-1088) + cmd_chrome_retry (1095-1245);
   `gitcurator/cli/cli_settings.py`'s model-menu half
   (smart_model_pick / prompt_model_menu / apply_model_choice);
   `gitcurator/cli/cli_run.py`'s _section/_check_line +
   preflight_checks (690-719).
4. The landed slice itself (the patterns to extend):
   `internal/cli/*.go` (terminal/settings/syncdoor/cli/status/dead +
   cfgutil's or-door laws), `testdiff/goldens/dump_cli.py` (the
   generator's seam inventory), `internal/cli/goldencli_test.go` (the
   replay harness + the fold/unfold + the mutation teeth).

## 4. Known constraints / owner actions for SESSION-028

- **The go-directive acknowledgement (STILL WAITING — seventeenth
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
  a new ADR — surface, don't improvise. (SESSION-027 added zero — the
  Windows console calls ride syscall.NewLazyDLL.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections —
  ScanMasterRedoRows (zero dedicated Go tests) is the top item; then
  the redo-consume laws beyond the golden section, the proper-note
  laws, the pipeline-redo legs (the CLI routing legs land WITH the
  T-60/T-62 work per the spec's own Tests section). Also standing:
  the note_state SURGERY path's CRLF round-trip differential (Phase
  4+); the Windows live legs (Phase 7 clean-machine checklist); a
  Windows VM without WebView2 for T-70 (optional).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit — the Write-tool redaction is a safety NET, not
  a license); tests never weakened; reality wins over specs (fix the
  spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-028)

```
You are starting SESSION-028 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-one full wipes across
   SESSION-005..027 plus mid-session degradations; SESSION-021 added the
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
   SESSION-027 landed T-60's first slice — THE CLI SHELL — at 143c590 on top of
   SESSION-026's close 2cc07d1 (the tree must be clean and origin/main at
   143c590; the baseline is 1852 tests race-clean over 31 packages — re-verify
   the suite before work if anything looks off). If anything is missing, rebuild
   per docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass
   it ephemerally with git -c http.extraheader="Authorization: Basic $(printf
   'x-access-token:%s' "$PAT" | base64)", the Basic scheme; never store it — AND
   SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the header into the new
   repo's .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; locate the work repo via the GitHub API's
   authenticated /user/repos listing when unsure — Currena is an UNRELATED
   sibling; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`). SESSION-027's operational finds: the gui.app generator
   door stubs to the REAL underlying classes (CacheDB/dead_link_threshold — the
   PyQt6 guards are environment shims, not logic); gitcurator.constants.APP_DIR
   binds at cache_db's import — patch it BEFORE the import; the Write/Edit tool
   REDACTS token-shaped strings at file-write time (craft safe-shaped test
   tokens, never ghp_-shaped ones; byte-level secret-scan before every commit);
   detached background processes DO NOT SURVIVE tool-call boundaries — run long
   suites in FOREGROUND CHUNKS per tool call. The older laws stand: (a) THE
   CD-DISCIPLINE LAW — the shell resets its working directory between tool
   calls, so every git/go command carries its own cd (or go's -C flag); never
   commit outside /home/z/my-project/second-brain-project; (b) the
   renderer-eats-escape-pairs lesson — the two-character pairs [h and [m are
   eaten on the OUTPUT path, the INPUT path, AND THE BASH-ARG path: avoid those
   spellings, write commit messages to a FILE and commit with -F, then verify
   the committed bytes with 'git log --format=%B' before pushing, and never
   trust a DISPLAY of escape-bearing bytes (grep the FILE bytes). PUSH AFTER
   EVERY TASK this session.

2. Read, in order: the repo README.md (Phase 5 OPENED — the slice's record is
   in the status) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-027's close: the
   CLI shell at 143c590, 1852 tests race-clean over 31 packages, the
   environment recipes, the newest lessons incl. the Write-tool token
   redaction) → docs/sessions/SESSION-027.md (the most recent report — the
   slice's evidence + the golden's construction).

3. THIS SESSION CONTINUES T-60: the command wirings per
   docs/port-blueprint/module-specs/spec-cli.md's SESSION-027 shipped record
   (READ THE SPEC + THE LANDED SLICE BEFORE THE CODE — the remaining list:
   the --init wizard with the repo radar (internal/reporadar is ported; the
   radar legs ride seams; the wizard COLLECTS the Telegram keys for
   compatibility but never offers the MTProto login per ADR-0004 Am.3),
   --test-connection over the PORTED battery (connectioncheck.RunLocalChecks +
   Summarize + the _section/_check_line renderers — the Worker leg IS the
   queue check; no Telegram leg), --detect-llm (probe → the model menu → set +
   save — cli_settings's model-menu half lands with it),
   --hand-delivery/--chrome-retry wiring (the engines are ported; the
   run_batch_visual legs wait for T-62 — honest boundary cards where the
   engine isn't there). Extend the differential stone with the slice:
   testdiff/goldens/dump_cli.py grows the new sections (the wizard with
   scripted stdin + the stubbed radar + the probe seams; the battery with the
   net stubs on the reference side and the connectioncheck seams on the Go
   side; the detect menu with the probe seams; the hand/chrome listings over
   crafted rows) and internal/cli/goldencli_test.go replays them — the
   two-run byte-identical proof + the mutation teeth stay part of every
   regeneration. T-62's bounded open item (the all-clear gate's Phase-5
   re-spec, ADR-0004 Am.3) is resolved IN T-62, never improvised. If the task
   will not fit the session's capacity, land the first coherent wiring
   properly with its own differential proof and record the remainder honestly
   — one wiring done properly with proof beats a rushed whole; close cleanly
   at a natural boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized set is exactly
   modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text
   v0.32.0 — do not bump any; x/net/html stays OUT unless the owner approves
   an ADR); surface blockers with evidence (hash-level, not display) instead
   of improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-60's
   own acceptance per the spec's Tests section — evaluate what this session's
   wirings landed honestly), commit + push everything (ask me for the GitHub
   PAT if it is not already available — never store it anywhere), write
   docs/sessions/SESSION-028.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-029.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

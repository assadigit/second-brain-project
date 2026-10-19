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
  `41d699e` (SESSION-028).
- **T-62 LANDED ACROSS TWO SESSIONS (Phase 5) — COMPLETE**: slice 1 at
  `e2bc4cd` (SESSION-029); slice 2 at `6ccf184` + the close at
  `e0cef51` (SESSION-030).
- **T-61 COMPLETE (SESSION-031's close-out audit, `d66cb2a`)**: the
  paint box + the NO_COLOR/TTY law + the spinner under the engine's
  heartbeat + the 65001 hardening all landed with tests
  (TestEnableColorsDoor + the paint unit battery + the golden's
  paint_shapes + engine_test.go's heartbeat proof) +
  GOOS=windows GOARCH=amd64 CGO_ENABLED=0 build + vet clean; the
  runtime Windows-console legs recorded as Phase 7's clean-machine
  checklist items (doc 10 §7; T-70's VM). **ADR-0007 Am.2 records the
  owner's re-instatement of the CLI-facing legs** (Am.1's cut
  superseded by the SESSION-027..031 instruction line).
- **T-63 COMPLETE (SESSION-031, `e89eaaa` + `abda6cf`)**: THE
  LAUNCHERS — cmd/gitcurator-cli (the --cli strip law ported from
  main.py's routing, the --version stamp door per doc 10 §5, the
  exit pass-through), VERSION (0.92.0-go.1), the launcher set in
  cmd/gitcurator-cli/ (GitCurator-CLI/SYNC/Setup/Start .bats +
  gitcurator-cli.sh + 1-INSTALL.bat — ASCII+CRLF at byte level, the
  walk-up with landmark VERSION, the exe run DIRECT with the
  reference's own --cli defaults, EXITCODE echo + pause, the git-only
  install check with NO registry writes and NO PATH mutations), and
  **THE DISPATCH FLIP** (the reality-wins find: the T-62 drivers had
  landed but Run's dispatch still answered not-yet cards for
  --auto/--sync/--retry-failed/--import-file — flipped to the real
  drivers; THE DISPATCH-COMPLETION LAW recorded). The differential
  stone grew the launchers section (the reference's OWN bytes recorded
  raw — shared grammar byte-compared, divergences asserted both ways,
  mutation teeth on three surfaces) at **101,777** bytes two-run
  byte-identical. **1922 tests race-clean over 33 packages** at the
  close.
- **Push state: CLEAN** — SESSION-031 pushed after every task
  (d66cb2a → e89eaaa → abda6cf → the close commit); verify `git
  status` is in sync at the start (the tip at or beyond the close
  commit).
- **Remaining on Phase 5's books**: T-64 (the differential corpus
  full run + fix loop), T-65 (the `v0.92.0-go.1` tag + zip artifact).
- **The environment recipe** (unchanged): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator;
  must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5`
  — 64 hex chars, verified byte-equal to go.dev's published JSON in
  SESSION-032; earlier cards carried a 63-char transcription of it),
  venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
  build/TOOLCHAIN.md, go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (the
  tls-client floor, ADR-0002 Am.2; never bump).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE DISPATCH-COMPLETION LAW** (SESSION-031): a command is only
    LANDED when Run's dispatch routes the argv to it — unit tests
    that call the cmd* methods directly keep the suite green while
    the user-facing door stays stale; the acceptance drives
    `Run(argv)`.
  - **THE STATE-CHAIN LAW** (SESSION-030): a golden replay of
    sequential sections is ONE state chain — an ISOLATED replay
    (e.g. a mutation test with a fresh buildEnv) must reproduce the
    payload's own pre-state (the `--reset-dead` run clears the dead
    set; the dry-run report writer's real `<app>/reports` dirs) or
    the bytes diverge.
  - **THE MUTATION-TEETH LAW** (SESSION-030): a mutation test runs
    the REAL replay against the mutated golden and must FAIL, while
    the unmutated golden must still PASS (the harness guard) — a
    mutation test that only proves the pattern strings exist has NO
    teeth.
  - **THE CONTEXT-DEATH CONTINUATION recipe** (SESSION-030/021): git
    status + `git log --oneline -5` + file timestamps tell what
    ACTUALLY survived; audit the in-flight work (never blindly trust
    it), repair, run the FULL battery, then land.
  - **THE DSN-EMPTY-PATH HAZARD** (SESSION-030): never pass "" as a
    db path (modernc treats a `?_pragma…` DSN as a LITERAL
    filename); `NewWebsiteStateDB("cache.db")` is the
    AppDir-anchored default.
  - **THE FIXED-LENGTH ROOTS law** (SESSION-029): the replay's
    buildEnv roots must be exactly as wide as the generator's
    mkdtemp roots (`/tmp/gc-cli-app-goreplay` +
    `/tmp/gc-cli-vaults-goreplay`; "goreplay" = the 8-rune suffix
    width; keep it).
  - **THE SHALLOW-COPY SNAPSHOT law** (SESSION-029): a golden section
    recording a config the driven call MUTATES must copy.deepcopy
    BEFORE the call.
  - The fake-clock arithmetic (1000.0 → 1000.042 renders 42 ms via
    `int((t1-t0)*1000)`); exception class-name tokens via
    `type("TimeoutError", …)` (the SHAPE is the contract); a golden
    section capturing a RETURN VALUE cannot use the stdout-capture
    helper.
  - **THE CD-DISCIPLINE LAW** — every git/go command carries its own
    cd (or go's -C flag); never commit outside the work repo.
  - **THE RENDERER-EATS-ESCAPE-PAIRS lesson** — write commit messages
    to a FILE and commit with -F, verify with `git log --format=%B`,
    never trust a DISPLAY of escape-bearing bytes.
  - **THE WRITE-TOOL REDACTION lesson** — the Write/Edit tools REDACT
    token-shaped strings at file-write time (craft safe-shaped test
    tokens, never ghp_-shaped; byte-level secret-scan before every
    commit).
  - **THE BACKGROUND-PROCESS lesson** — detached processes do not
    survive tool-call boundaries; run long suites in FOREGROUND
    CHUNKS.
  - **PUSH AFTER EVERY TASK** — ask the owner for the PAT the MOMENT
    a push is due; never store it.

## 2. Next session: SESSION-032 — T-64 the differential corpus full run, then T-65 the tag + zip

- **Verify the push state first** (§1): SESSION-031 pushed everything
  (the tree ends clean); confirm `git status` is in sync before work
  (and keep pushing after every task).
- **T-64 (M) — THE DIFFERENTIAL CORPUS FULL RUN + FIX LOOP** per doc
  09 §3 (testing-and-parity) + the task-list row: run the full
  differential corpus (the ten named stones' complete case lists —
  the doc 09 §3 case inventory is the authority; read it FIRST),
  archive the report, fix what diverges in the fix loop (reality wins
  — fix the spec in the same commit when the reference's own bytes
  contradict it). The CLI golden (goldencli, 101,777) is one of the
  stones — the corpus run drives ALL of them (doc 09's inventory)
  and lands the archived run report in the repo.
- **T-65 (XS) — THE `v0.92.0-go.1` TAG + ZIP ARTIFACT** per doc 10
  §2/§8 + the task-list row (only if T-64's capacity remains — one
  task properly beats two rushed):
  1. **The zip composition question to resolve per the specs first**:
     the deterministic-zip builder (mkzip, the build_zip.py laws:
     fixed timestamps 2026-09-29 00:00:00, sorted arcnames, 0o644,
     DEFLATE 9, tracked-files-only, the banned-secrets guard, the
     launcher flattening to the zip root, .bat CRLF+ASCII
     assertion) is T-80/Phase 7's row per traceability F4 — but
     T-65's acceptance wants "GitHub release with asset". Options:
     land a minimal mkzip slice early (the laws are already listed
     in doc 10 §2) or ask the owner. Surface the decision in the
     session report; do not improvise silently.
  2. The zip needs the CLI exe + the launchers (flattened to root) +
     VERSION + config.example.json + WINDOWS-QUICKSTART.md (doc 10
     §6). The GUI-side launchers (GitCurator.bat /
     GitCurator-DRY-RUN.bat Go equivalents — the DRY-RUN one runs
     `--cli --auto --yes --dry-run`) were recorded in SESSION-031 as
     T-65's books; write them with the same byte laws + the golden's
     launchers section extended to cover them if they are
     byte-anchored to the reference's own app-root .bats.
  3. The tag `v0.92.0-go.1` + the GitHub release with the zip asset
     (the owner's law: a release without its zip is not done); the
     sha256 recorded in the release notes.
- **The golden regeneration law** (standing): any dump_cli.py change
  regenerates the two-run byte-identical proof; MIND THE LAWS
  (fixed-length roots, deep-copy snapshots, the state-chain law, the
  mutation-teeth law, the dispatch-completion law).
- If capacity runs short: land T-64 properly first (the corpus run +
  the report + the fixes); T-65 is XS and fits a short session. One
  slice done properly with proof beats a rushed whole.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-031.md` — the most recent report (T-61 +
  T-63 COMPLETE; the dispatch flip; the launchers' differential).
2. `docs/port-blueprint/09-testing-and-parity.md` §3 — THE CORPUS
  CASE INVENTORY (T-64's authority).
3. `docs/port-blueprint/10-build-release-packaging.md` §2/§6/§8 — the
  build commands, the zip layout + laws, the release automation
  (T-65's authority; §10 the reference artifact).
4. `docs/port-blueprint/13-implementation-task-list.md` — the T-64/
  T-65 rows + the deferred test debt note.
5. The harness surfaces the corpus run drives: `testdiff/` (the ten
  stones' generators + the pinned rev file) + the replays under
  `internal/*/…_test.go`; `internal/cli/goldencli_test.go`'s
  launchers section (the pattern for any T-65 launcher additions).

## 4. Known constraints / owner actions for SESSION-032

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The T-65 zip-composition decision** (§2 above) — surface it with
  the spec citations; the owner may pre-decide (minimal mkzip slice
  vs Phase-7 deferral).
- **The go-directive acknowledgement (STILL WAITING — twenty-first
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers
  x/net/html, approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **ADR-0007 Am.2** (SESSION-031): the re-instatement record of the
  CLI-facing legs — the owner's veto is recorded as a standing
  option (nothing rides on it for T-64/T-65; the launchers + tag are
  the commanded books either way).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB the persisted
  extraheader from .git/config after the clone); goldens are
  GENERATED by running its unmodified code.
- **Dependency law**: `modernc.org/sqlite` v1.36.1;
  `bogdanfinn/tls-client` v1.16.0; `golang.org/x/net` v0.48.0 +
  `golang.org/x/text` v0.32.0 (direct); ANY other add needs a new
  ADR — surface, don't improvise. (SESSION-031 added zero.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections
  (ScanMasterRedoRows is the top item); the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live
  legs + the launchers' clean-VM smoke (Phase 7 checklist); a
  Windows VM without WebView2 for T-70 (optional).
- Standing: secrets / owner data never in git (byte-level
  secret-scan before every commit); tests never weakened; reality
  wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-032)

```
You are starting SESSION-032 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-five full wipes across
   SESSION-005..031 plus mid-session degradations; the continuation recipe is
   PROVEN: git status + git log --oneline -5 + file timestamps tell what
   ACTUALLY survived; a preserved commit or in-flight draft is continued and
   AUDITED — never blindly trusted, never restarted from zero). FIRST verify
   with evidence: work repo at /home/z/my-project/second-brain-project
   (symlinked /home/z/second-brain-project), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must
   match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go
   (go.dev tarball, sha256 ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39ead
   df76996045fe5), venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, and go.mod with modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text
   v0.32.0 direct, and the go directive at 1.24.1 (the tls-client-forced floor —
   ADR-0002 Am.2; do not bump any of these). Then check `git status` and `git
   log --oneline -3`: SESSION-031 landed T-61 COMPLETE (the close-out audit,
   d66cb2a) + T-63 COMPLETE THE LAUNCHERS (e89eaaa: cmd/gitcurator-cli + the
   .bat/.sh set + the VERSION stamp + THE DISPATCH FLIP; abda6cf: the
   differential stone's launchers section — the reference's own bytes at
   101,777, shared grammar byte-compared, divergences asserted, mutation teeth)
   and PUSHED everything after every task — 1922 tests race-clean over 33
   packages. Verify git status shows up to date with origin/main (the tip at or
   beyond SESSION-031's close commit). Use the PAT only ephemerally: git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)" and never store it — AND SCRUB IT if you ever CLONE with it: git
   clone -c http.extraheader=… PERSISTS the header into the new clone's
   .git/config, so git config --unset-all http.extraheader in BOTH clones right
   after; Currena is an UNRELATED sibling; a mid-session mode-flip is repaired
   with `git checkout-index -f -a` after `git fsck`. Re-verify the suite before
   work if anything looks off. The operational laws, newest first: THE
   DISPATCH-COMPLETION LAW — a command is only LANDED when Run's dispatch
   routes the argv to it (unit tests calling the cmd* methods directly keep the
   suite green while the user-facing door stays stale; the acceptance drives
   Run(argv)); THE STATE-CHAIN LAW — a golden replay of sequential sections is
   ONE state chain and an ISOLATED replay (a mutation test with a fresh
   buildEnv) must reproduce the payload's own pre-state (the --reset-dead run
   clears the dead set; the dry-run writer's real <app>/reports dirs);
   THE MUTATION-TEETH LAW — a mutation test runs the REAL replay against the
   mutated golden and must FAIL while the unmutated golden still PASSES; the
   DSN-empty-path hazard — never pass "" as a db path
   (NewWebsiteStateDB("cache.db") is the AppDir-anchored default). The older
   laws stand: FIXED-LENGTH ROOTS (the replay's roots stay exactly as wide as
   the generator's mkdtemp roots — /tmp/gc-cli-app-goreplay +
   /tmp/gc-cli-vaults-goreplay); SHALLOW-COPY SNAPSHOT (deep-copy a config
   the driven call mutates BEFORE the call); the fake-clock arithmetic
   (1000.0 → 1000.042 renders 42 ms via int((t1-t0)*1000)); exception
   class-name tokens via type("TimeoutError", …) — the SHAPE is the contract;
   a golden section capturing a RETURN VALUE cannot use the stdout-capture
   helper; (a) THE CD-DISCIPLINE LAW — every git/go command carries its own cd
   (or go's -C flag); never commit outside the work repo; (b) the
   renderer-eats-escape-pairs lesson — write commit messages to a FILE and
   commit with -F, verify the committed bytes with 'git log --format=%B', never
   trust a DISPLAY of escape-bearing bytes; (c) the Write/Edit tool REDACTS
   token-shaped strings at file-write time (craft safe-shaped test tokens,
   never ghp_-shaped; byte-level secret-scan before every commit); (d) detached
   background processes DO NOT SURVIVE tool-call boundaries — run long suites
   in FOREGROUND CHUNKS per tool call. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (T-61 + T-63 COMPLETE at
   abda6cf — the launchers' record is in the status) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md
   (your state card — §1 records SESSION-031's close: T-61 + T-63
   COMPLETE, 1922 tests race-clean over 33 packages, the dispatch
   flip, the newest law incl. the dispatch-completion one) →
   docs/sessions/SESSION-031.md (the most recent report).

3. THIS SESSION IS T-64 THE DIFFERENTIAL CORPUS FULL RUN (then T-65
   the tag + zip if capacity remains): per docs/port-blueprint/
   09-testing-and-parity.md §3's case inventory (READ IT FIRST — it
   is the corpus's authority) + the task-list row, run the FULL
   differential corpus over the ten landed stones (the generators
   under testdiff/ drive the reference @ 98f4fc7; every regeneration
   carries the two-run byte-identical proof), archive the run report
   in the repo, and fix what diverges in the fix loop (reality wins:
   fix the spec in the same commit when the reference's own bytes
   contradict it; tests are never weakened). Then, only if capacity
   remains, T-65 per doc 10 §2/§6/§8: resolve the zip-composition
   question per the specs FIRST (the deterministic mkzip is T-80's
   Phase-7 row per traceability F4, but T-65's acceptance wants the
   release asset — surface the minimal-mkzip-slice vs defer decision,
   do not improvise silently), write the GUI-side launcher
   equivalents (GitCurator.bat / GitCurator-DRY-RUN.bat — the
   DRY-RUN one runs --cli --auto --yes --dry-run; the same byte laws;
   extend the golden's launchers section if they are byte-anchored),
   and land the v0.92.0-go.1 tag + the GitHub release with the zip
   asset (the owner's law: a release without its zip is not done;
   sha256 in the notes). If capacity runs short: T-64 done properly
   with the archived report first — T-65 is XS and fits a short
   session; one slice done properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes); no new dependencies without
   an ADR (the authorized set is exactly modernc.org/sqlite v1.36.1,
   tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not bump
   any; x/net/html stays OUT unless the owner approves an ADR);
   surface blockers with evidence (hash-level, not display) instead of
   improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (T-64's — and T-65's if reached — own acceptance per the specs),
   commit + push everything (ask me for the GitHub PAT the moment a
   push is due — never store it anywhere; never end the session
   push-pending), write docs/sessions/SESSION-032.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt (Phase 6's
   entry — the GUI shell — is the natural next milestone once
   Phase 5's books are empty), and end your reply with a
   plain-language report plus the prompt I should copy to start
   SESSION-033.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

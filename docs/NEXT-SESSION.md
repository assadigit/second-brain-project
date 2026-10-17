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
- **Phase 4 — THE TASK LIST IS COMPLETE (SESSION-015–025; T-40..T-50
  all landed; T-51 removed by the Option-B stamp)**: T-40 `internal/
  websitepipeline` at `8098ac4`; T-41 `directorynote.go` at `f528c22`;
  T-42 `internal/githubpipeline` + `batch/summary.go` at `e92e4e7`;
  T-43 `internal/handdelivery` at `be688bd`; T-44 `internal/chrometabs`
  at `71e5e34`; T-45 `internal/vaultscan` at `5c0508e`; T-46
  `internal/vaulthygiene` + the repobanish completion at `992a3d1`;
  T-47 `internal/vaultseal` at `d3a5d27`; T-48 `internal/goodrepos` +
  `internal/reporadar` at `7179a6c`; T-49 `internal/directorysync` at
  `e2c941a`; **T-50 `internal/mirror` at `555a72c` (SESSION-025) —
  THE LIBRARY MIRROR + THE LINKING LAYER COMPLETE**: the whole merged
  group (mirror.py + mirror_keys.py + linking.py + embeddings.py +
  recall.py) — the containment guards, the RAW-source move matching,
  the recall block surgery + sanitizers, the embeddings surface over
  the loopback stub, THE SUGGESTION LIFECYCLE, the Suggestions note +
  the Related (auto) blocks in mirror copies only, the Phase-5↔6
  carry-over — with the ported test_phase5 + test_phase6 core
  selections (79) + THE GOLDEN (39 sections, the generator imports
  the reference's own fixtures via importlib, run twice
  byte-identical; it caught one live map-order bug in RunMirror's
  log loops before it could land). **1826/1826 tests race-clean over
  30 packages; T-50's acceptance — PASS.**
- **THE PHASE-4 EXIT GATE: 8 of 10 named differential stones landed;
  ONE artifact outstanding — the 30-link offline golden runner
  equivalent** (the reference's `tools/run_golden_websites.py
  --offline`: 30 entries from tests/golden/websites.json, a fake LLM
  answering from the golden file, canned pages feeding the fetch, the
  whole pipeline into a throwaway vault, the report + notes as the
  artifact — CI's mode). The machinery it exercises is fully ported
  (T-40's 51-section + T-42's 12-section goldens cover the parts);
  the end-to-end 30-entry corpus replay was never built on the Go
  side. **This is SESSION-026's first task — the Phase-4 closer.**
  The full stone-by-stone evidence table is in
  docs/sessions/SESSION-025.md's gate evaluation.
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-025 pushed T-50 the moment it went green. The
  recovery order for ANY mid-task death stays: (1) `git status` +
  `git log --oneline -5` + file timestamps to learn what ACTUALLY
  survived — never trust a handoff summary's claim over `git log`;
  (2) a preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **nineteen full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth, SESSION-024's
  eighteenth, SESSION-025's nineteenth), plus SESSION-010's
  mid-session mode-flip, plus SESSION-021's CONTEXT death with the
  sandbox intact. The restore recipes: start-of-session (clone
  `assadigit/second-brain-project` with the PAT passed ephemerally as
  the Basic `x-access-token:<PAT>` header — locate the work repo via
  the GitHub API's authenticated /user/repos listing if the clone
  looks wrong (Currena is an UNRELATED sibling); never store the PAT
  — AND SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the PAT
  into the new repo's .git/config, so run `git config --unset-all
  http.extraheader` in BOTH clones right after cloning, then push
  with the ephemeral `-c` form only); GitCurator @ `98f4fc7` (verify
  against `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/
  .local/go` (go.dev tarball), venv `/home/z/venvs/gitcurator`
  (Python 3.12, stdlib-only), symlinks `/home/z/second-brain-project`
  and `/home/z/GitCurator`; mid-session mode-flip (`git
  checkout-index -f -a` restores modes from the index; `git fsck` to
  sanity-check). Verify the pins at every start: `git rev-parse
  HEAD:go.mod` — expect the v1.36.1/tls-client/1.24.1 bytes (do NOT
  bump sqlite; do NOT bump tls-client past v1.16.0). **Do NOT run
  `go mod download all`** (it churns go.sum; `go build`/`go test`
  work without it). **The renderer-escape + cd-discipline laws stay:
  the two-character pairs [h and [m are eaten on the OUTPUT path, the
  INPUT path, AND THE BASH-ARG path — write commit messages to a FILE
  and commit with -F, then verify the committed bytes with 'git log
  --format=%B' before pushing; note the DISPLAY of test diffs eats
  the pairs too (grep the FILE bytes before believing a corruption
  claim). AND the cd-discipline law: the shell RESETS its working
  directory between tool calls — EVERY git command carries its own
  cd; never commit outside /home/z/my-project/second-brain-project.**

## 2. Next session: SESSION-026 — the Phase-4 closer (the 30-link offline golden runner equivalent), then Phase 5 OPENS (T-60 CLI parity) if capacity remains

- **First task: THE PHASE-4 CLOSER (M)** — the offline golden runner
  equivalent: a generator (`testdiff/goldens/dump_goldenwebsites.py`,
  the house pattern) that drives the UNMODIFIED reference's
  `tools/run_golden_websites.py --offline` over its own
  `tests/golden/websites.json` corpus (30 entries; the fake LLM
  answers from the golden file; the canned pages come from the
  runner's own `offline_fetch`) into a throwaway vault, dumping the
  produced notes + the report (clocks frozen, paths folded — every
  per-run fact normalized) into `internal/websitepipeline/testdata/`;
  then the Go replay: the ported pipeline over the SAME corpus (the
  same canned pages + fake-LLM seams) byte-compared against the
  golden. READ the reference's runner FIRST
  (`app/gitcurator/tools/run_golden_websites.py` — the offline_fetch
  + offline_llm shapes, the report renderer, the exit-code law), and
  the websites.json corpus itself. When it lands green + pushes:
  stamp the Phase-4 exit gate CLOSED in the roadmap (the stone table
  in SESSION-025.md is the evidence) and declare Phase 4 CLOSED.
- **Then: Phase 5 OPENS (if capacity remains — close cleanly at a
  natural boundary otherwise)** — T-60 `internal/cli`: the full flag
  surface, the wizard, the status/test/detect/hand/chrome/retry
  commands, per `docs/port-blueprint/module-specs/spec-cli.md` —
  READ THE SPEC BEFORE THE CODE. The CLI's mirror/tool legs (the
  mirror_manual, add_recall_hooks, build_links tools + the status
  row) are T-60's books (the T-50 session recorded them as such).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; the dry-run seam at call sites; rows collected before
  nested statements; tests never weakened; reality wins over specs
  (fix the spec in the same commit); goldens are GENERATED by
  running the unmodified Python app, never hand-typed; no per-run
  facts in goldens; the frozen-clock seams (`clockNow` per package —
  mirror's is the newest). T-41's: text-mode reads translate
  universal newlines before any cap or parse; invalid UTF-8 decodes
  with the maximal-subpart U+FFFD law; vault walks are os.walk's
  files-before-subdirs order. T-42's: the house JSON decoder's
  shapes; floats render through helpers.PyFloatRepr. T-43's:
  `pyBytesBlank` is the blank-body law; Go MAP ranges are RANDOM.
  T-44's: Go's `url.Parse` is NOT CPython's urlsplit; `cmd.Wait()`
  closes the pipe; scripted timing tests keep the reference's
  UNPATCHED cadences. T-45's: THE POINTER-THREADING LAW ([]*NoteRow);
  keep BOTH frontmatter parsers; Python text-mode read(n) counts
  CHARACTERS; the _windows_eyes law rides a stat seam. T-46's: the
  WALK-ORDER FOLD; bounded reads count runes; None doors type-narrow
  to the zero value; a revive-style table iteration rides sorted
  keys; dry-run asymmetries preserved. T-47's: THE REDACTION LAW;
  the token NEVER touches .git/config; FETCH_HEAD rewritten after
  every token-URL fetch; env scrubbing + HideWindow where the
  REFERENCE's own module scrubs; frozen GIT DATES make commit shas
  deterministic; the reference's mock-surface maps to instance
  seams. T-48's: the apiFn/tokenURLFn instance seams; the
  ReadDir-order convergence; atexit maps to an exported Cleanup; the
  base-dict freeze quirk; the frozen-dates sha trick. T-49's: the
  []-vs-null byte law; `storage.PyDumpsUTF8`; dictDoor presence
  checks + the double-wrapping trap; the silent unknown-type
  failure; the set-iteration convergence; the push_mirror
  env-inheritance law; the CPython re vs Go regexp convergence; the
  lstrip character-set quirk; the raw-tree banished round-trip; the
  importlib fixture-import generator trick. **T-50's additions: the
  map-order law BITES THE LOGS (plan.Trees is a map — iterate the
  fixed treeSpecs order wherever the reference rides dict order);
  the []-vs-null law extends to .get() dumps (never-set keys dump as
  null — emit nil, not the zero value); the house JSON decoder
  returns int64 for integral literals (float() doors must accept
  int64/int/json.Number); str.splitlines' boundary set is wider than
  \n (pySplitLines is the twin); walk-order independence is earned
  by sorting every order-dependent surface before pinning (the
  reference's own listdir order is filesystem-dependent); the
  clockNow seam covers every time.strftime write; the reference's
  exception-text rendering stays unpinnable (the SHAPE is the
  contract, per the T-40 law).**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-025.md` — the most recent report (the
   T-50 evidence, the 39-section golden + the 90 tests, the
   nineteenth-wipe rebuild, THE PHASE-4 CLOSE-OUT EVALUATION with the
   stone-by-stone evidence table — §'s gate evaluation is the
   Phase-4 closer's brief).
2. `app/gitcurator/tools/run_golden_websites.py` (in the reference @
   `98f4fc7`) + `app/tests/golden/websites.json` + the runner's
   offline_fetch/offline_llm shapes + the report renderer + the
   exit-code law — BEFORE writing the generator.
3. The working precedents: `testdiff/goldens/dump_mirror.py` (the
   newest generator — the importlib fixture-import trick, the fold
   semantics [a REPLACE, not a prefix test — the refusal messages
   EMBED paths], the frozen-clock patches), `internal/mirror/
   golden_test.go` (the normM comparison twin), `internal/
   websitepipeline/golden_test.go` (T-40's harness — the closer
   rides the same package).
4. `docs/port-blueprint/module-specs/spec-cli.md` — only when Phase
   5 actually opens (T-60).

## 4. Known constraints / owner actions for SESSION-026

- **The go-directive acknowledgement (STILL WAITING — fifteenth
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
  a new ADR — surface, don't improvise. (T-50 added zero.)
- **Deferred test debt on the books (recorded with evidence, not
  blockers — pick up when capacity allows, the redo-scan battery
  first)**: test_noteredo's remaining un-ported selections —
  ScanMasterRedoRows (zero dedicated Go tests — the top item), the
  redo-consume laws beyond the golden section, the proper-note laws,
  the pipeline-redo legs, the CLI chrome-retry routing legs; the GUI
  wiring legs stay the GUI worker shell task's books. Also standing:
  the note_state SURGERY path's CRLF round-trip differential
  (Phase 4+); the Windows live legs (Phase 7 clean-machine
  checklist); a Windows VM without WebView2 for T-70 (optional).
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-026)

```
You are starting SESSION-026 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (nineteen full wipes across
   SESSION-005..025 plus mid-session degradations; SESSION-021 added the
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
   SESSION-025 landed T-50 mirror at 555a72c (the tree must be clean and
   origin/main at the SESSION-025 close; the baseline is 1826 tests race-clean
   over 30 packages — re-verify the suite before work if anything looks off).
   If anything is missing, rebuild per docs/NEXT-SESSION.md §1 (clone via the
   PAT I give you when you ask — pass it ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s'
   "$PAT" | base64)", the Basic scheme; never store it — AND SCRUB IT: `git
   clone -c http.extraheader=…` PERSISTS the header into the new repo's
   .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; a mid-session mode-flip is repaired with `git
   checkout-index -f -a` after `git fsck`; and if the sandbox or the context
   dies MID-TASK, learn what actually survived from git evidence — `git
   status` + `git log --oneline -5` + file timestamps — before restarting
   from zero: a preserved commit or draft is continued and audited, never
   blindly trusted, and a handoff summary's claim never outranks `git
   log`'s). SESSION-025's operational finds: the fold in a golden generator
   must be a REPLACE not a prefix test (the refusal messages EMBED paths —
   a prefix fold silently passes them through); the reference's own
   os.walk/listdir order is filesystem-dependent — pin order-dependent
   surfaces SORTED; run_golden_websites.py --offline is the Phase-4 closer's
   subject. The older laws stand: (a) THE CD-DISCIPLINE LAW — the shell
   resets its working directory between tool calls, so every git/go command
   carries its own cd (or go's -C flag); never commit outside
   /home/z/my-project/second-brain-project; (b) the renderer-eats-escape-
   pairs lesson — the two-character pairs [h and [m are eaten on the OUTPUT
   path, the INPUT path, AND THE BASH-ARG path: avoid those spellings, or
   write commit messages to a FILE and commit with -F, then verify the
   committed bytes with 'git log --format=%B' before pushing. PUSH AFTER
   EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase-4
   state: T-40..T-50 ALL LANDED (T-50 at 555a72c, 1826 tests race-clean over
   30 packages), the exit gate 8-of-10 stones with ONE artifact outstanding,
   the environment recipes, the clone-config PAT-scrub hazard, the renderer +
   cd-discipline lessons, and the standing port laws incl. T-50's additions —
   the map-order law bites the logs, the []-vs-null law extends to .get()
   dumps, the house decoder's int64 shapes, the splitlines boundary set, the
   walk-order independence, the clockNow seam) → docs/sessions/SESSION-025.md
   (the most recent report — the T-50 evidence + THE PHASE-4 CLOSE-OUT
   EVALUATION with the stone-by-stone evidence table, which is your brief).

3. THIS SESSION IS THE PHASE-4 CLOSER: build the 30-link offline golden
   runner equivalent — the exit gate's last named artifact. READ FIRST:
   app/gitcurator/tools/run_golden_websites.py (in the reference @ 98f4fc7 —
   the --offline mode: a fake LLM answering from tests/golden/websites.json,
   the offline_fetch canned pages, the report renderer, the exit-code law)
   and the websites.json corpus itself. Then build per the house pattern: a
   generator (testdiff/goldens/dump_goldenwebsites.py) that drives the
   UNMODIFIED reference's --offline runner over its own corpus into a
   throwaway vault, dumping the produced notes + the report (clocks frozen,
   paths FOLDED BY REPLACE — the report embeds paths, no per-run facts
   anywhere) into internal/websitepipeline/testdata/; then the Go replay:
   the ported pipeline over the SAME corpus (the same canned pages + the
   fake-LLM seams) byte-compared against the golden. TEMP VAULTS ONLY, NO
   real network. When it lands green + pushes: stamp the Phase-4 exit gate
   CLOSED in the roadmap (cite SESSION-025.md's stone table as the evidence)
   and declare Phase 4 CLOSED in the README. THEN, if capacity remains:
   Phase 5 OPENS — T-60 internal/cli per docs/port-blueprint/module-specs/
   spec-cli.md (READ THE SPEC BEFORE THE CODE; the mirror/tool legs —
   mirror_manual, add_recall_hooks, build_links, the status row — are T-60's
   books per the T-50 session's record) — but close cleanly at a natural
   boundary if capacity runs out; one task done properly with proof beats a
   rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
   Phase-4 exit gate's final stamp — CLOSED if the closer lands, else the
   closer's own acceptance honestly), commit + push everything (ask me for
   the GitHub PAT if it is not already available — never store it anywhere),
   write docs/sessions/SESSION-026.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language report
   plus the prompt I should copy to start SESSION-027.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

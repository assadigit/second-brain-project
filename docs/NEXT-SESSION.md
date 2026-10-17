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
- **Phase 4 OPEN (SESSION-015–023)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  T-44 `internal/chrometabs` at `71e5e34` (THE FIFTH DOOR COMPLETE);
  T-45 `internal/vaultscan` at `5c0508e` (THE VAULT SCAN COMPLETE +
  the repos-banishment floor); T-46 `internal/vaulthygiene` + the
  COMPLETION of `internal/repobanish` at `992a3d1` (THE STRAY-NOTES
  DOOR + THE REPOS BANISHMENT COMPLETE); T-47 `internal/vaultseal`
  at `d3a5d27` (SESSION-022 — THE SEAL COMPLETE, + the
  directorysync belt floor); **T-48 `internal/goodrepos` +
  `internal/reporadar` landed at `7179a6c` (SESSION-023) — THE
  PUBLISHER + THE RADAR COMPLETE**: the curated vault published as
  its PUBLIC good-repos directory (the tiny frontmatter subset + the
  note parser with the display-name ladder url → aliases → filename
  stem, the emoji README writer with GitHub's anchor math + the
  plural doors + the stars-desc sort, the staging lifetime with the
  atexit → CleanupStaging convergence, THE CREATION GUARD, the fetch
  alignment + the skip-when-unchanged cycle, the one-time-token push
  with the defensive remote removes, THE REDACTION LAW + the
  FETCH_HEAD scrub, the config bridge + the CLI) + the read-only
  repo radar (the remote-URL grammar, git's content-addressed blob
  SHA, the diversified round-robin sampler, the candidate ranking
  with size NEVER an emptiness signal, THE DOMINANCE LADDER — ≥3
  matches AND ≥2× the runner-up else a report never a guess — the
  DETECTION ladder REMOTE → NAME → FINGERPRINT, the only-EMPTY-slots
  law, the quiet-startup gate, the never-raises doors) — with the
  ported test_goodrepos/test_reporadar selections (the reference's
  whole suites) + the local-bare-repos publish-flow battery (the
  tokenURLFn instance seam) + the safety-law battery + THE GOLDENS
  (25 + 12 sections, dump_goodrepos.py + dump_reporadar.py, paths/
  clock/git-dates frozen — the Go port produces the reference's
  exact byte-deterministic publish sha `a300991`).
  **1658/1658 tests race-clean over 29 packages; T-48's acceptance —
  PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-023 pushed T-48 the moment it went green. The
  recovery order for ANY mid-task death stays: (1) `git status` +
  `git log --oneline -5` + file timestamps to learn what ACTUALLY
  survived — never trust a handoff summary's claim over `git log`;
  (2) a preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE):
  **seventeen full wipes** now (SESSION-005..018 thirteen,
  SESSION-019's fourteenth mid-draft, SESSION-020's fifteenth,
  SESSION-022's sixteenth, SESSION-023's seventeenth — both repos,
  the toolchain, the venv, the symlinks), plus SESSION-010's
  mid-session mode-flip, plus SESSION-021's CONTEXT death with the
  sandbox intact (recovered by the audit-then-push order). The
  restore recipes: start-of-session (discover the remote via the
  GitHub API with the owner's in-chat PAT passed ephemerally as the
  Basic `x-access-token:<PAT>` header; clone `assadigit/
  second-brain-project` with `git -c http.extraheader=…` — the Basic
  scheme; never store it — AND SCRUB IT: `git clone -c
  http.extraheader=…` PERSISTS the PAT into the new repo's
  .git/config, so run `git config --unset-all http.extraheader` in
  BOTH clones right after cloning, then push with the ephemeral
  `-c` form only); GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball), venv `/home/z/venvs/gitcurator` (Python 3.12,
  stdlib-only), symlinks `/home/z/second-brain-project` and
  `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  Verify the pins at every start: `git rev-parse HEAD:go.mod` —
  expect the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do
  NOT bump tls-client past v1.16.0). **SESSION-023's operational
  additions, kept — the renderer-escape lesson WIDENED AGAIN: this
  sandbox's renderer eats the two-character console-escape pairs
  [h AND [m on the OUTPUT path, the INPUT path, AND THE BASH-ARG
  path (SESSION-023's first commit-message attempt lost the bracket-m
  of a mkdemp bracket spelling — caught by the card's own law:
  verify the committed bytes with 'git log --format=%B' before
  pushing; avoid bracket-adjacent-lowercase-m/h spellings in
  messages and heredocs, or write the message to a FILE and commit
  with -F, then verify). AND the cd-discipline law: the shell RESETS
  its working directory between tool calls — EVERY git command must
  carry its own cd (or go's -C flag); SESSION-023 had one
  wrong-repo commit (the sandbox harness repo at /home/z/my-project
  answered a bare 'git add -A') — reset with 'git reset --mixed
  HEAD~1' before any push; nothing leaked; never commit outside
  /home/z/my-project/second-brain-project.**

## 2. Next session: SESSION-024 — Phase 4 continues: T-49 directorysync (L), then T-50 mirror (M) if capacity remains

- **T-49 `internal/directorysync`** (L) per
  **`docs/port-blueprint/module-specs/spec-cloud-directory-sync.md`**
  — the REST of the module: build_index, publish, the command
  cycle, the config trio. Already landed: the belt floor (belt.go —
  the pattern builders + the shipped-config embed at
  assets/directorysync, T-47) and the banish gate
  (banishgate.go, T-40). The drift seal's publish door
  (`vaultseal.DriftSealDeps.PublishAfterBatch`) wires here — the
  runner's honest not-wired warning retires when it does.
  Acceptance per the task list: the index golden; the command-cycle
  tests; the `test_directorysync` ports.
- **T-50 `internal/mirror`** (M) if capacity remains: per
  spec-core-mirror + spec-core-linking — block-surgery idempotence,
  containment guards, the suggestion lifecycle.
- The Phase 4 queue behind them: T-51 telegram (L, dormant per
  ADR-0004 Am.3 — removed from Phase 4).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened;
  reality wins over specs (fix the spec in the same commit);
  goldens are GENERATED by running the unmodified Python app, never
  hand-typed; no per-run facts in goldens; the frozen-clock seams
  (`clockNow`/`handdelivery.Clock`/`repobanish.clockNow`/
  `vaultseal.clockNow`/`goodrepos.clockNow` now too) exist for
  exactly this. T-41's additions: text-mode reads translate
  universal newlines before any cap or parse (`helpers.PyNewlines`);
  invalid UTF-8 decodes with the maximal-subpart U+FFFD law; vault
  walks are os.walk's files-before-subdirs order, never WalkDir.
  T-42's: the house JSON decoder's shapes (storage.OrderedMap +
  json.Number / int64 scalars); floats render through
  helpers.PyFloatRepr. T-43's: `pyBytesBlank` is the blank-body law;
  Go MAP ranges are RANDOM — any Python dict-insertion-order tie
  rides a tracked order slice. T-44's: Go's `url.Parse` is NOT
  CPython's urlsplit (the chrometabs.pyurl shapes are the lenient
  reference); Go's `cmd.Wait()` closes the pipe (any read error +
  process-exited = the exit sentence); scripted timing tests keep
  the reference's UNPATCHED cadences. T-45's: THE
  POINTER-THREADING LAW ([]*NoteRow); keep BOTH frontmatter parsers
  (vault_scan's beside note_reads'); Python text-mode read(n) counts
  CHARACTERS; the _windows_eyes law rides a stat seam. T-46's: the
  WALK-ORDER FOLD; bounded reads count runes; never-raises wrappers
  are unreachable in Go — the honest error at the caller is the
  documented convergence; None doors type-narrow to the zero value;
  bare-string channel doors type-narrow to the dict shape; a
  revive-style table iteration rides sorted keys; dry-run
  asymmetries are preserved exactly. T-47's: THE REDACTION LAW; the
  token NEVER touches .git/config (token-less remote adds; one-time
  push URLs only); FETCH_HEAD rewritten after every token-URL fetch;
  env scrubbing + HideWindow for every git subprocess; CPython 3.12's
  fnmatch lives in internal/vaultseal/fnmatch.go; frozen GIT DATES
  make commit shas deterministic for goldens; the reference's
  mock-surface maps to instance seams; a GUI module's PURE half ports
  into the engine package it consumes. **T-48's additions: the
  instance seams are apiFn (the reference's own _api mock surface) +
  tokenURLFn (the _token_url builder — the door that lets the
  fetch/push ladder run against LOCAL BARE REPOS hermetically);
  Go's os.ReadDir sorted order is the documented convergence over
  CPython's readdir order when a walk feeds a sampler (the
  diversification law is order-insensitive; goldens generated from
  sorted-creation trees); atexit maps to an exported Cleanup the CLI
  defers; shutil.copy2 converges to a mode-preserving content copy;
  the publish-result base dict is FROZEN before ensure_repo (the raw
  repo_name quirk on a pushed first publish); configparser reads
  converge to a minimal INI scan; a push failure keeps the local
  commit but ok stays FALSE (exit 1) while an unconfigured-name skip
  is ok TRUE (exit 0); the frozen-dates sha trick pins the WHOLE
  publish path byte-identically (the a300991 proof).**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-023.md` — the most recent report (the
   T-48 evidence, the two goldens + the 108 tests, the
   seventeenth-wipe rebuild, the renderer/bash-arg + cd-discipline
   operational finds).
2. `docs/port-blueprint/module-specs/spec-cloud-directory-sync.md` —
   BEFORE T-49's code (the module's binding spec; note the T-47
   belt-floor record already in it).
3. The reference: `app/gitcurator/cloud/directory_sync.py` (1,826
   LOC — all of it) + its test suite (`app/tests/
   test_directorysync*.py` — read the selections before porting).
4. Working examples: `internal/directorysync/belt.go` +
   `banishgate.go` (the already-landed floors T-49 completes),
   `internal/vaultseal/drift.go` (THE DRIFT SEAL's pure half whose
   publish door — DriftSealDeps.PublishAfterBatch — wires in T-49),
   `internal/goodrepos` (T-48 — the newest git-subprocess law
   family: the apiFn/tokenURLFn seams, the local-bare-repos battery,
   the frozen-dates publish-sha golden), and
   `testdiff/goldens/dump_goodrepos.py` (the newest generator).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-024

- **The go-directive acknowledgement (STILL WAITING — twelfth
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
  a new ADR — surface, don't improvise. (T-48 added zero.)
- **T-49-specific safety laws**: temp vaults/trees only — never the
  owner's real data; no real network in tests (fake the Worker/HTTP
  surface the way the module's own tests do; local bare repos where
  a git remote is needed); the redaction + token laws carry from
  T-47/T-48 (the publish path talks to remotes too); tests never
  weakened.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves — including the drift seal's
  QThread wrapper and its publish wiring after T-49, test_fifthdoor
  §9b's end-of-delivery modal legs and test_vaultscan's GUI legs,
  the two-vault worker job (test_reposcan §8), the GUI scorecard
  models (test_reposcan §9) — the websites-phase wiring; the
  Runner's seams exist for exactly this; the Python-source contract
  tests of test_reporadar §6 join this task too); the note_state
  SURGERY path's CRLF round-trip differential (Phase 4+); the
  Windows live legs (Phase 7 clean-machine checklist); a Windows VM
  without WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-024)

```
You are starting SESSION-024 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (seventeen full wipes across
   SESSION-005..023 plus mid-session degradations; SESSION-021 added the
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
   SESSION-023 landed T-48 goodrepos + reporadar at 7179a6c (the tree must be
   clean and origin/main at the SESSION-023 close; the baseline is 1658 tests
   race-clean over 29 packages — re-verify the suite before work if anything
   looks off). If anything is missing, rebuild per docs/NEXT-SESSION.md §1
   (clone via the PAT I give you when you ask — pass it ephemerally with git
   -c http.extraheader="Authorization: Basic $(printf 'x-access-token:%s'
   "$PAT" | base64)", the Basic scheme; never store it — AND SCRUB IT: `git
   clone -c http.extraheader=…` PERSISTS the header into the new repo's
   .git/config, so run `git config --unset-all http.extraheader` in BOTH
   clones right after cloning; a mid-session mode-flip is repaired with `git
   checkout-index -f -a` after `git fsck`; and if the sandbox or the context
   dies MID-TASK, learn what actually survived from git evidence — `git
   status` + `git log --oneline -5` + file timestamps — before restarting
   from zero: a preserved commit or draft is continued and audited, never
   blindly trusted, and a handoff summary's claim never outranks `git
   log`'s). TWO operational laws from SESSION-023: (a) THE CD-DISCIPLINE
   LAW — the shell resets its working directory between tool calls, so every
   git/go command carries its own cd (or go's -C flag); a bare `git add -A`
   once hit the WRONG repo (the sandbox harness repo at /home/z/my-project)
   and needed a `git reset --mixed HEAD~1` before push — never commit
   outside /home/z/my-project/second-brain-project; (b) the renderer-eats-
   escape-pairs lesson WIDENED — the two-character pairs [h and [m are eaten
   on the OUTPUT path, the INPUT path, AND THE BASH-ARG path (a commit
   message lost the bracket-m of a mkdtemp bracket spelling): avoid those
   spellings, or write commit messages to a FILE and commit with -F, then
   verify the committed bytes with 'git log --format=%B' before pushing.
   PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4
   state with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at
   be688bd, T-44 at 71e5e34, T-45 at 5c0508e, T-46 at 992a3d1, T-47 at
   d3a5d27, and T-48 at 7179a6c, 1658 tests race-clean over 29 packages,
   the environment recipes, the clone-config PAT-scrub hazard, the renderer
   + cd-discipline lessons, and the standing port laws incl. T-48's
   additions — the apiFn/tokenURLFn instance seams, the ReadDir-order
   convergence, the atexit convergence, the base-dict freeze quirk, the
   frozen-dates publish-sha proof) → docs/sessions/SESSION-023.md (the most
   recent report — T-48's two goldens + the 108 tests, the
   seventeenth-wipe rebuild, the operational finds).

3. THIS SESSION IS T-49 (L): Phase 4 continues — internal/directorysync per
   docs/port-blueprint/module-specs/spec-cloud-directory-sync.md — READ THE
   SPEC BEFORE THE CODE. The REST of the module: build_index, publish, the
   command cycle, the config trio. Already landed: the belt floor (belt.go —
   the pattern builders + the shipped-config embed at
   assets/directorysync, T-47) and the banish gate (banishgate.go, T-40);
   the drift seal's publish door (vaultseal.DriftSealDeps.PublishAfterBatch)
   wires here — the runner's honest not-wired warning retires when it does.
   Acceptance per the task list: the index golden; the command-cycle tests;
   the test_directorysync ports. The working precedents: T-47's
   internal/vaultseal + T-48's internal/goodrepos (the git-subprocess law
   family — the env scrub, the redaction door, the timeout ladder, the
   instance seams, the local-bare-repos technique) and the house generator
   pattern (testdiff/goldens/dump_goodrepos.py is the newest example — the
   path + clock + git-dates folding). TEMP VAULTS/TREES ONLY, NO real
   network. If capacity remains after T-49 goes green + pushes: T-50
   internal/mirror (M) per spec-core-mirror + spec-core-linking — but close
   cleanly at a natural boundary if capacity runs out; one task done
   properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-49's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-024.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-025.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

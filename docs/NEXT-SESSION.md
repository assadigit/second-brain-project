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
- **Phase 4 OPEN (SESSION-015–022)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  T-44 `internal/chrometabs` at `71e5e34` (THE FIFTH DOOR COMPLETE);
  T-45 `internal/vaultscan` at `5c0508e` (THE VAULT SCAN COMPLETE +
  the repos-banishment floor); T-46 `internal/vaulthygiene` + the
  COMPLETION of `internal/repobanish` at `992a3d1` (THE STRAY-NOTES
  DOOR + THE REPOS BANISHMENT COMPLETE); **T-47 `internal/vaultseal`
  landed at `d3a5d27` (SESSION-022) — THE SEAL COMPLETE**: the whole
  module (the construction with the v0.83 creation guard + the v0.89
  local-only belt, the `_git` subprocess runner with env scrubbing +
  HideWindow + THE REDACTION LAW, ensure_repo [the nested-repo guard
  → git init -b main, the managed .gitignore blocks, the token-less
  remote, the 422 race re-check], the untrack law — `git rm -r
  --cached`, NEVER a disk delete — seal() with THE DIVERGENCE LADDER
  [plain push → fetch + FETCH_HEAD redaction → stale-rejection
  retry / rebase / unrelated-histories merge / the seal-rescue
  branch — main is NEVER force-pushed], resync() [write-tree delta
  counts, the commit-tree reconciliation, the empty-tree baseline],
  status(), the four config bridges, the standalone CLI, THE DRIFT
  SEAL's pure half [the plan + the runner with THE QUIET LAW; the
  publish door rides T-49's seam], CPython 3.12's fnmatch) + THE
  BELT FLOOR in `internal/directorysync` (the pattern builders + the
  shipped-config embed at assets/directorysync — T-49 completes the
  module) — with THE GOLDEN (16 sections, paths + clock + git-dates
  frozen, `dump_vaultseal.py`; it caught the universal-newlines find
  live) + the ported test_sealfix/test_resyncfix/test_localonly/
  test_driftseal selections + the safety-law battery. **1550/1550
  tests race-clean over 27 packages; T-47's acceptance — PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — SESSION-022 pushed T-47 the moment it went green. The
  recovery order for ANY mid-task death stays: (1) `git status` +
  `git log --oneline -5` + file timestamps to learn what ACTUALLY
  survived — never trust a handoff summary's claim over `git log`;
  (2) a preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
- Environment notes (STANDING — the sandbox is REBUILDABLE): **sixteen
  full wipes** now (SESSION-005…018 thirteen, SESSION-019's
  fourteenth mid-draft, SESSION-020's fifteenth, SESSION-022's
  sixteenth — both repos, the toolchain, the venv, the symlinks),
  plus SESSION-010's mid-session mode-flip, plus SESSION-021's
  CONTEXT death with the sandbox intact (recovered by the
  audit-then-push order). The restore recipes: start-of-session
  (discover the remote via the GitHub API with the owner's in-chat
  PAT passed ephemerally as the Basic `x-access-token:<PAT>` header;
  clone `assadigit/second-brain-project` with `git -c
  http.extraheader=…` — the Basic scheme; never store it — AND
  SCRUB IT: `git clone -c http.extraheader=…` PERSISTS the PAT into
  the new repo's .git/config, so run `git config --unset-all
  http.extraheader` in BOTH clones right after cloning, then push
  with the ephemeral `-c` form only); GitCurator @ `98f4fc7`
  (verify against `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 →
  `/home/z/.local/go` (go.dev tarball), venv `/home/z/venvs/
  gitcurator` (Python 3.12, stdlib-only), symlinks `/home/z/
  second-brain-project` and `/home/z/GitCurator`; mid-session
  mode-flip (`git checkout-index -f -a` restores modes from the
  index; `git fsck` to sanity-check). Verify the pins at every
  start: `git rev-parse HEAD:go.mod` — expect the
  v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do NOT bump
  tls-client past v1.16.0). **SESSION-015's debugging lesson, kept —
  widened by SESSION-022's own close: this sandbox's renderer EATS
  the two-character console-escape sequences `[h` AND `[m` — on the
  OUTPUT path (verify suspicious "corruption" with %q byte dumps
  before believing rendered text) AND ON THE INPUT PATH (SESSION-
  022's close-commit message lost a "[manual's" to an eaten "[m" —
  avoid bracket-adjacent-lowercase-m/h spellings inside heredocs and
  commit messages, or verify the committed bytes with
  `git log --format=%B` before pushing).**

## 2. Next session: SESSION-023 — Phase 4 continues: T-48 goodrepos + reporadar (M), then T-49 directorysync (L) if capacity remains

- **T-48 `internal/goodrepos` + `internal/reporadar`** (M) per
  **`docs/port-blueprint/module-specs/spec-integrations-goodrepos.md`**
  (READ THE SPEC AND ADR-0009 BEFORE THE CODE — both modules are
  ADR-0009's git-CLI-subprocess consumers; T-47's `_gitEnv`/redaction/
  timeout family in `internal/vaultseal/vaultseal.go` is the working
  precedent to follow, and `internal/vaultseal/fnmatch.go` is the
  house fnmatch if goodrepos' matching needs it). Acceptance per the
  task list: README/publish tests with fake GitHub; radar dominance
  tests.
- **T-49 `internal/directorysync`** (L) if capacity remains: the
  module's REST (build_index, publish, the command cycle, the config
  trio) per spec-cloud-directory-sync.md — the belt floor
  (belt.go) + the banish gate (banishgate.go) + the shipped-config
  embed are already landed; the drift seal's publish door
  (`vaultseal.DriftSealDeps.PublishAfterBatch`) wires here.
- The Phase 4 queue behind them: T-50 mirror (M), T-51 telegram (L,
  dormant per ADR-0004 Am.3).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened;
  reality wins over specs (fix the spec in the same commit);
  goldens are GENERATED by running the unmodified Python app, never
  hand-typed; no per-run facts in goldens; the frozen-clock seams
  (`clockNow`/`handdelivery.Clock`/`repobanish.clockNow`/
  **`vaultseal.clockNow`** now too) exist for exactly this.
  **T-41's additions: text-mode reads translate universal newlines
  before any cap or parse (`helpers.PyNewlines`) — T-47's golden
  caught this being missed in vaultseal's readGitignore, the freshest
  proof the law is load-bearing; invalid UTF-8 decodes with the
  maximal-subpart U+FFFD law; vault walks are os.walk's
  files-before-subdirs order, never WalkDir. T-42's additions: the
  house JSON decoder's shapes (storage.OrderedMap + json.Number /
  int64 scalars) must be handled wherever LLM/JSON output is read;
  floats render through helpers.PyFloatRepr. T-43's additions:
  `pyBytesBlank` is the blank-body law (Python bytes.strip is
  ASCII-only); Go MAP ranges are RANDOM — any Python
  dict-insertion-order tie must ride a tracked order slice.
  T-44's additions: Go's `url.Parse` is NOT CPython's urlsplit — the
  chrometabs.pyurl shapes are the house's lenient-parse reference;
  Go's `cmd.Wait()` closes the pipe (any read error + process-exited
  = the exit sentence, never wait for a literal EOF) — DIRECTLY
  load-bearing for git subprocesses; a scripted timing test must
  keep the reference's UNPATCHED cadences. T-45's additions: THE
  POINTER-THREADING LAW ([]*NoteRow); vault_scan's own frontmatter
  parser is preserved BESIDE note_reads' unified parse — keep both;
  Python text-mode read(n) counts CHARACTERS; the _windows_eyes law
  rides a stat seam (fsExists/fsIsDir/fsIsFile), not os patches.
  T-46's additions: the WALK-ORDER FOLD; the bounded read counts
  runes; never-raises wrappers are unreachable in Go — the honest
  error at the caller is the documented convergence; None doors
  type-narrow to the zero value; bare-string channel doors
  type-narrow to the dict shape (the T-40 precedent); the channel
  factory's raise maps to (nil, error); a revive-style table
  iteration rides sorted keys when determinism matters; dry-run
  asymmetries are preserved exactly. **T-47's additions: THE
  REDACTION LAW (any captured subprocess output that could carry a
  secret passes through the redaction door — goodrepos publishes to
  GitHub remotes too, so T-48 needs the same law); the token NEVER
  touches .git/config (token-less remote adds; one-time push URLs
  only); FETCH_HEAD gets rewritten after every token-URL fetch; env
  scrubbing + HideWindow for every git subprocess; CPython 3.12's
  fnmatch lives in internal/vaultseal/fnmatch.go (RE2 convergences
  documented); frozen GIT DATES make commit shas deterministic for
  goldens (GIT_AUTHOR_DATE/GIT_COMMITTER_DATE ride the subprocess
  env); the reference's mock-surface maps to instance seams
  (apiFn/pushFn/ensureFn fields + a package-level ctor seam); a GUI
  module's PURE half (backup_seal's drift seal) ports into the
  engine package it consumes, with the lazy-imported dependency as
  an injected seam until its task lands.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-022.md` — the most recent report (the T-47
   evidence, the golden + the 86 tests, the universal-newlines find,
   the sixteenth-wipe rebuild, the belt floor).
2. `docs/port-blueprint/module-specs/spec-integrations-goodrepos.md`
   and `docs/port-blueprint/decisions/0009-git-cli-subprocess.md` —
   BEFORE T-48's code.
3. The reference: `app/gitcurator/integrations/goodrepos.py` +
   `app/gitcurator/integrations/repo_radar.py` (all of each — the
   spec's section maps) + their test suites (`app/tests/
   test_goodrepos*.py`, `app/tests/test_reporadar.py` — read the
   suites' selections before porting).
4. Working examples: `internal/vaultseal` (T-47 — the git subprocess
   law family: the env scrub, the redaction door, the timeout
   ladder, the instance seams, the frozen-dates golden trick,
   fnmatch.go) and `internal/directorysync/belt.go` (the belt floor
   T-49 completes); `testdiff/goldens/dump_vaultseal.py` (the newest
   generator — the clock + git-dates freezing).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-023

- **The go-directive acknowledgement (STILL WAITING — eleventh
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
  a new ADR — surface, don't improvise. (T-47 added zero.)
- **T-48-specific safety laws**: temp repos only — never the owner's
  real vault; no real network in tests (fake the GitHub API surface
  the way test_sealfix/test_localonly do; goodrepos' own tests fake
  the remote with local bare repos — follow them); assert the token
  never appears in .git/config or any captured output; main is never
  force-pushed anywhere.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves — including the drift seal's QThread
  wrapper and its publish wiring after T-49, test_fifthdoor §9b's
  end-of-delivery modal legs and test_vaultscan's GUI legs, the
  two-vault worker job (test_reposcan §8), the GUI scorecard models
  (test_reposcan §9) — the websites-phase wiring; the Runner's seams
  exist for exactly this); the note_state SURGERY path's CRLF
  round-trip differential (Phase 4+); the Windows live legs (Phase 7
  clean-machine checklist); a Windows VM without WebView2 for T-70
  (optional).

## 5. THE PROMPT TO COPY (SESSION-023)

```
You are starting SESSION-023 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (sixteen full wipes across
   SESSION-005..022 plus mid-session degradations; SESSION-021 added the
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
   SESSION-022 landed T-47 vaultseal + the directorysync belt floor at d3a5d27
   (the tree must be clean and origin/main at the SESSION-022 close; the
   baseline is 1550 tests race-clean over 27 packages — re-verify the suite
   before work if anything looks off). If anything is missing, rebuild per
   docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass
   it ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it — AND SCRUB IT: `git clone -c
   http.extraheader=…` PERSISTS the header into the new repo's .git/config, so
   run `git config --unset-all http.extraheader` in BOTH clones right after
   cloning; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`; and if the sandbox or the context dies MID-TASK, learn
   what actually survived from git evidence — `git status` + `git log
   --oneline -5` + file timestamps — before restarting from zero: a preserved
   commit or draft is continued and audited, never blindly trusted, and a
   handoff summary's claim never outranks `git log`'s). PUSH AFTER EVERY TASK
   this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4 state
   with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at be688bd,
   T-44 at 71e5e34, T-45 at 5c0508e, T-46 at 992a3d1, and T-47 at d3a5d27,
   1550 tests race-clean over 27 packages, the environment recipes, the
   clone-config PAT-scrub hazard, the `[h` renderer lesson, and the standing
   port laws incl. T-47's additions — the redaction law, the token-never-in-
   .git/config law, the fnmatch home, the frozen-git-dates golden trick, the
   instance seams) → docs/sessions/SESSION-022.md (the most recent report —
   T-47's golden + the ported selections, the universal-newlines find the
   golden caught live, the belt floor in directorysync, the sixteenth-wipe
   rebuild).

3. THIS SESSION IS T-48 (M): Phase 4 continues — internal/goodrepos +
   internal/reporadar per docs/port-blueprint/module-specs/
   spec-integrations-goodrepos.md + ADR-0009 (docs/port-blueprint/decisions/
   0009-git-cli-subprocess.md) — READ THE SPEC AND THE ADR BEFORE THE CODE.
   ADR-0009's other git-CLI-subprocess consumers (integrations/goodrepos.py —
   the curated Obsidian vault published as its PUBLIC good-repos directory,
   the staging lifetime, the README/index writers, the publish ladder; and
   integrations/repo_radar.py — the owner's mirror-repo DETECTION that feeds
   manual_repo_name). The working precedent is T-47's internal/vaultseal: the
   _gitEnv scrub, the redaction door (the token never in .git/config, never
   in any captured output — goodrepos publishes to GitHub remotes too), the
   timeout ladder, the instance seams, fnmatch.go if the matching needs it,
   the frozen-git-dates golden trick. Acceptance: the ported selections of
   the reference's own suites (app/tests/test_goodrepos*.py +
   test_reporadar.py) — README/publish tests with fake GitHub (local bare
   repos as the remote exactly like the reference's own tests do), radar
   dominance tests — TEMP REPOS ONLY, NO real network. The house generator
   pattern applies if a golden is warranted — testdiff/goldens/
   dump_vaultseal.py is the newest example (the path + clock + git-dates
   folding). If capacity remains after T-48 goes green + pushes: T-49
   internal/directorysync per spec-cloud-directory-sync.md (build_index,
   publish, the command cycle, the config trio — the belt floor +
   banishgate + the shipped-config embed are already landed; the drift
   seal's publish door [vaultseal.DriftSealDeps.PublishAfterBatch] wires
   here) — but close cleanly at a natural boundary if capacity runs out;
   one task done properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-48's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-023.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-024.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

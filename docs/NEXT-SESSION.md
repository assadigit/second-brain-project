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
- **Phase 4 OPEN (SESSION-015–021)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  T-44 `internal/chrometabs` at `71e5e34` (THE FIFTH DOOR COMPLETE);
  T-45 `internal/vaultscan` at `5c0508e` (THE VAULT SCAN COMPLETE +
  the repos-banishment floor); **T-46 `internal/vaulthygiene` + the
  COMPLETION of `internal/repobanish` landed at `992a3d1`
  (SESSION-021) — THE STRAY-NOTES DOOR + THE REPOS BANISHMENT
  COMPLETE**: the whole hygiene module (the classifiers — empty stub,
  app-owned bodyless with the fence laws, the lock doors, the note-name
  shape, the identity fold; the [[wiki-link]] eyes + the head-identity
  harvest; THE ONE WALK — each .md opened exactly once through the
  shared notereads cache or the direct bounded read, the three
  fingerprints shape/name-twin/linked-ghost, the retirement to
  .trash/empty-stubs, the master-index scaffold removal, the legacy
  root reports, heal_machine_vaults) + ReviveBanishedRepos (THE ♻️
  DOOR BACK) + RunRepoBanishmentPass (the confirmation-gate
  orchestration with THE TALLY) — with THE GOLDEN (9 sections, paths +
  walk-order folded, `dump_vaulthygiene.py`) + the ported
  test_strayheal/test_vaulthygiene/test_repobanish selections +
  test_picklock's locked-bodyless law + test_vaultscan's one-open
  hygiene-sharing leg. **1464/1464 tests race-clean over 25 packages;
  T-46's acceptance (the ported selections of both modules' reference
  suites) — PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — and SESSION-021 proved why again: the working context DIED
  after the T-46 commit existed but before the push + close. The
  continuation recovered by trusting **git evidence over the (lossy)
  handoff summary** — `git log` proved the commit; the summary claimed
  no work existed — then audited the preserved commit with the full
  battery from scratch (vet/gofmt/windows-build clean, go.mod blob
  unchanged, secret scan clean, 1464/1464 race-clean) and only then
  pushed `e5d48dc..992a3d1`.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…018 (**thirteen**), plus
  SESSION-010's mid-session mode-flip, plus SESSION-019's FOURTEENTH
  (mid-draft: the toolchain wiped, the repos + the draft survived),
  plus SESSION-020's FIFTEENTH (a FULL wipe — both repos, the
  toolchain, the venv, the symlinks), plus **SESSION-021's new event
  class: a CONTEXT death with the sandbox intact** (the environment
  survived; the working context ended after-commit-before-push). The
  recovery order for ANY mid-task death: (1) `git status` + `git log
  --oneline -5` + file timestamps to learn what ACTUALLY survived —
  never trust a handoff summary's claim over `git log`; (2) a
  preserved commit or draft is continued and audited (the full
  battery re-run), never blindly trusted, never restarted from zero.
  The restore recipes: start-of-session (discover the remote via the
  GitHub API with the owner's in-chat PAT passed ephemerally as the
  Basic `x-access-token:<PAT>` header; clone `assadigit/
  second-brain-project` with `git -c http.extraheader=…` — the Basic
  scheme; never store it; **SCRUB AFTER EVERY CLONE: `git clone -c
  http.extraheader=…` PERSISTS the PAT into the new repo's
  .git/config — run `git config --unset-all http.extraheader` in BOTH
  clones right after cloning, then push with the ephemeral `-c` form
  only**); GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball), venv `/home/z/venvs/gitcurator` (Python 3.12,
  stdlib-only), symlinks `/home/z/second-brain-project` and
  `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  Verify the pins at every start: `git rev-parse HEAD:go.mod` —
  expect the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do
  NOT bump tls-client past v1.16.0). **SESSION-015's debugging
  lesson, kept: this sandbox's output renderer EATS the two-character
  sequence `[h` (a console escape) — verify suspicious "corruption"
  with %q byte dumps before believing rendered text.**

## 2. Next session: SESSION-022 — Phase 4 continues: T-47 vaultseal (L), then T-48 goodrepos + reporadar if capacity remains

- **T-47 `internal/vaultseal`** (L) per **`docs/port-blueprint/
  module-specs/spec-integrations-vaultseal.md` + ADR-0009
  (`docs/port-blueprint/decisions/0009-git-cli-subprocess.md`)**
  (READ THE SPEC AND THE ADR BEFORE THE CODE — the v21 card's
  "spec-core-vault-seal.md" filename was WRONG; these are the real
  names, and the task list's size is (L)). The git-CLI-subprocess
  seal of an Obsidian vault into a **private** GitHub mirror
  (`integrations/vault_seal.py`, 1,301 LOC): the construction (the
  v0.83 creation guard — a 404 is an ERROR spoken fast, opt-in
  creates PRIVATE; the local_only belt), `status()` (pure
  introspection, no network), `seal()` (ensure_repo: the
  nested-repo guard → `git init -b main`, the managed .gitignore
  block, the local-only identity, the token-less `remote add`, the
  401 fresh-token hint, the 422 creation race re-check;
  _apply_local_only: the SECOND managed block REBUILT from the
  current list + the untrack law — `git rm -r --cached`, NEVER a
  disk delete; `add -A` pathspec-limited; the unchanged skip; the
  commit-message shape; THE DIVERGENCE LADDER under `_push`: plain
  push → fetch + FETCH_HEAD token redaction → stale-rejection retry
  / rebase / unrelated-histories merge / `_rescue` branch — **main
  is NEVER force-pushed**; the never-raises wrapper), `resync()`
  (write-tree delta counts, the commit-tree reconciliation, the
  empty-tree baseline `4b825dc…`), the four config bridges (incl.
  `manual_seal_from_config`'s deliberately-NO-belt law — "a backup
  that filters the owner's own files is not a backup"), the
  standalone CLI. **Go laws**: `exec.Command("git", …)` with the
  exact arguments, env scrubbing (`GIT_TERMINAL_PROMPT=0`; strip
  `GIT_DIR`/`GIT_WORK_TREE`/`GIT_INDEX_FILE`/`GIT_CONFIG`/
  `GIT_CONFIG_NOSYSTEM`), HideWindow on Windows, token redaction in
  ANY captured output, **the token NEVER touches .git/config**,
  FETCH_HEAD rewrite; timeouts API 30 s / git 60 s / push+fetch
  180 s. Acceptance: the ported selections of the reference's
  real-git integration suites (`app/tests/test_driftseal.py` +
  `test_sealfix.py`) — first seal, unchanged skip, the divergence
  ladder's every branch, the belt untrack law (files stay on
  disk), resync delta counts, token-redaction assertions, the
  creation guard — **temp repos only, never the owner's vault; no
  real network** (fake the GitHub API surface the way the
  reference's own tests do).
- **T-48 `internal/goodrepos` + `reporadar`** (M) if capacity
  remains: `spec-integrations-goodrepos.md` (README/publish tests
  with fake GitHub; radar dominance tests).
- The Phase 4 queue behind them: T-49 directorysync (L; reuses
  T-34's rig directory doors — `banishgate.go` already staged),
  T-50 mirror (M), T-51 telegram (L, dormant per ADR-0004 Am.3).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened;
  reality wins over specs (fix the spec in the same commit);
  goldens are GENERATED by running the unmodified Python app, never
  hand-typed; no per-run facts in goldens; the frozen-clock seams
  (`clockNow`/`handdelivery.Clock`/`repobanish.clockNow`) exist for
  exactly this. **T-41's additions: text-mode reads translate
  universal newlines before any cap or parse (`helpers.PyNewlines`);
  invalid UTF-8 decodes with the maximal-subpart U+FFFD law; vault
  walks are os.walk's files-before-subdirs order, never WalkDir.
  T-42's additions: the house JSON decoder's shapes
  (storage.OrderedMap + json.Number / int64 scalars) must be
  handled wherever LLM/JSON output is read; floats render through
  helpers.PyFloatRepr. T-43's additions: `pyBytesBlank` is the
  blank-body law (Python bytes.strip is ASCII-only); Go MAP ranges
  are RANDOM — any Python dict-insertion-order tie must ride a
  tracked order slice. T-44's additions: Go's `url.Parse` is NOT
  CPython's urlsplit — the chrometabs.pyurl shapes are the house's
  lenient-parse reference; Go's `cmd.Wait()` closes the pipe (any
  read error + process-exited = the exit sentence, never wait for
  a literal EOF) — DIRECTLY load-bearing for T-47's git
  subprocesses; a scripted timing test must keep the reference's
  UNPATCHED cadences. T-45's additions: THE POINTER-THREADING LAW
  (Python dict aliasing across function boundaries = []*T in Go —
  struct copies break it silently); vault_scan's own frontmatter
  parser is preserved BESIDE note_reads' unified parse — keep both
  exactly as the reference does; Python text-mode read(n) counts
  CHARACTERS; the _windows_eyes law rides a stat seam
  (fsExists/fsIsDir/fsIsFile), not os patches. T-46's additions:
  the WALK-ORDER FOLD (the reference's os.listdir order vs the
  house ReadDir sorted order — decisions must be order-independent
  and goldens fold/compare sorted where order leaks); the bounded
  read counts runes; never-raises wrappers are unreachable in Go —
  the honest error at the caller is the documented convergence;
  None doors type-narrow to the zero value; bare-string channel
  doors type-narrow to the dict shape (the T-40 precedent); the
  channel factory's raise maps to (nil, error); a revive-style
  table iteration rides sorted keys when determinism matters;
  dry-run asymmetries are preserved exactly.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-021.md` — the most recent report (the T-46
   evidence, the golden + the ported selections, the context-death
   recovery and the audit-then-push order, the spec-name correction).
2. `docs/port-blueprint/module-specs/spec-integrations-vaultseal.md`
   and `docs/port-blueprint/decisions/0009-git-cli-subprocess.md` —
   BEFORE T-47's code.
3. The reference: `app/gitcurator/integrations/vault_seal.py` (all of
   it — the spec's section map: Construction 183-216, ensure_repo
   425-516, the ladder under _push 643-713, resync 820-969, status
   973-1009, the bridges + the CLI) + the test suites
   `app/tests/test_driftseal.py` + `app/tests/test_sealfix.py` (+ the
   spec's own Tests section).
4. Working examples: `internal/chrometabs` (T-44 — the subprocess
   law family: the launch ladder, the cmd.Wait() pipe-closing law,
   the exit-sentence reading) and `internal/handdelivery` (T-43 — the
   detached-process seam family); `testdiff/goldens/
   dump_vaulthygiene.py` (T-46 — the newest golden generator, the
   path + walk-order folding).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-022

- **The go-directive acknowledgement (STILL WAITING — tenth session
  carrying it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path).
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
  a new ADR — surface, don't improvise. (T-46 added zero.)
- **T-47-specific safety laws**: temp repos only — never the owner's
  real vault; no real network in tests (fake the GitHub API surface);
  assert the token never appears in `.git/config`, FETCH_HEAD, or any
  captured output; main is never force-pushed; the belt's `git rm
  --cached` never deletes from disk.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending
  (SESSION-021's context death is the freshest proof).
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves — including test_fifthdoor §9b's
  end-of-delivery modal legs and test_vaultscan's GUI legs: the Scan
  channel's Telegram round trip, the ScanPlanDialog checkbox
  battery, the ask-gate shapes, the two-vault worker job
  (test_reposcan §8), the GUI scorecard models (test_reposcan §9) —
  the websites-phase wiring; the Runner's seams exist for exactly
  this); the note_state SURGERY path's CRLF round-trip differential
  (Phase 4+); the Windows live legs (Phase 7 clean-machine
  checklist); a Windows VM without WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-022)

```
You are starting SESSION-022 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (fifteen full wipes across
   SESSION-005..020 plus mid-session degradations; SESSION-021 added a new
   event class — a CONTEXT death with the sandbox intact: the working context
   ended AFTER the T-46 commit existed but BEFORE the push+close; the
   continuation recovered by trusting git evidence over the lossy handoff
   summary, auditing the preserved commit with the full battery, then pushing).
   FIRST verify with evidence: work repo at /home/z/my-project/second-brain-
   project (symlinked /home/z/second-brain-project), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must
   match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/
   go, venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per build/
   TOOLCHAIN.md, and go.mod with modernc.org/sqlite at v1.36.1, bogdanfinn/
   tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0
   direct, and the go directive at 1.24.1 (the tls-client-forced floor —
   ADR-0002 Am.2; do not bump any of these). Then check `git status` and
   `git log --oneline -3`: SESSION-021 landed T-46 vaulthygiene + the repobanish
   completion at 992a3d1 (the tree must be clean and origin/main at the
   SESSION-021 close; the baseline is 1464 tests race-clean over 25 packages —
   re-verify the suite before work if anything looks off). If anything is
   missing, rebuild per docs/NEXT-SESSION.md §1 (clone via the PAT I give you
   when you ask — pass it ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it — AND SCRUB IT: `git clone -c
   http.extraheader=…` PERSISTS the header into the new repo's .git/config, so
   run `git config --unset-all http.extraheader` in BOTH clones right after
   cloning; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`; and if the sandbox or the context dies MID-TASK, learn
   what actually survived from git evidence — `git status` + `git log --oneline
   -5` + file timestamps — before restarting from zero: a preserved commit or
   draft is continued and audited, never blindly trusted, and a handoff
   summary's claim never outranks `git log`'s). PUSH AFTER EVERY TASK this
   session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4 state
   with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at be688bd,
   T-44 at 71e5e34, T-45 at 5c0508e, and T-46 at 992a3d1, 1464 tests
   race-clean over 25 packages, the SESSION-021 context-death recovery story,
   the environment recipes, the clone-config PAT-scrub hazard, the `[h`
   renderer lesson, and the standing port laws incl. T-46's walk-order-fold
   additions) → docs/sessions/SESSION-021.md (the most recent report — T-46's
   golden + the ported selections, the audit-then-push recovery, the spec-name
   correction for T-47).

3. THIS SESSION IS T-47 (L): Phase 4 continues — internal/vaultseal per
   docs/port-blueprint/module-specs/spec-integrations-vaultseal.md + ADR-0009
   (docs/port-blueprint/decisions/0009-git-cli-subprocess.md) — READ THE SPEC
   AND THE ADR BEFORE THE CODE. The git-CLI-subprocess seal of an Obsidian
   vault into a PRIVATE GitHub mirror (integrations/vault_seal.py, 1,301 LOC):
   the construction (the v0.83 creation guard — 404 is an ERROR spoken fast,
   opt-in creates PRIVATE; the local_only belt), status() (pure introspection,
   no network), seal() (ensure_repo: the nested-repo guard → git init -b main,
   the managed .gitignore block, the local-only identity, the token-less
   remote add, the 401 fresh-token hint, the 422 creation race re-check;
   _apply_local_only: the SECOND managed block REBUILT from the current list +
   the untrack law — git rm -r --cached, NEVER a disk delete; add -A
   pathspec-limited; the unchanged skip; the commit-message shape; THE
   DIVERGENCE LADDER under _push: plain push → fetch + FETCH_HEAD token
   redaction → stale-rejection retry / rebase / unrelated-histories merge /
   _rescue branch — main is NEVER force-pushed; the never-raises wrapper),
   resync() (write-tree delta counts, the commit-tree reconciliation, the
   empty-tree baseline 4b825dc…), the four config bridges (incl.
   manual_seal_from_config's deliberately-NO-belt law), the standalone CLI.
   Go laws: exec.Command("git", …) with the exact arguments, env scrubbing
   (GIT_TERMINAL_PROMPT=0; strip GIT_DIR/GIT_WORK_TREE/GIT_INDEX_FILE/
   GIT_CONFIG/GIT_CONFIG_NOSYSTEM), HideWindow on Windows, token redaction in
   ANY captured output, the token NEVER in .git/config, FETCH_HEAD rewrite,
   timeouts API 30 s / git 60 s / push+fetch 180 s. Acceptance: the ported
   selections of the reference's real-git integration suites (app/tests/
   test_driftseal.py + test_sealfix.py) — first seal, unchanged skip, the
   divergence ladder's every branch, the belt untrack law (files stay on
   disk), resync delta counts, token-redaction assertions, the creation guard
   — with TEMP REPOS ONLY (never the owner's vault) and NO real network (fake
   the GitHub API surface the way the reference's own tests do). The house
   generator pattern applies if a golden is warranted — testdiff/goldens/
   dump_vaulthygiene.py is the newest example (the path + walk-order
   folding). If capacity remains after T-47 goes green + pushes: T-48
   internal/goodrepos + reporadar per spec-integrations-goodrepos.md — but
   close cleanly at a natural boundary if capacity runs out; one task done
   properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-47's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-022.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-023.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

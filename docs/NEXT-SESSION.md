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
- **Phase 4 OPEN (SESSION-015–020)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  T-44 `internal/chrometabs` at `71e5e34` (THE FIFTH DOOR COMPLETE);
  **T-45 `internal/vaultscan` landed at `5c0508e` (SESSION-020) —
  THE VAULT SCAN COMPLETE: the radar (the inventory walk, the move
  candidates, the CPython urlparse shapes hand-rolled), ONE NOTE PER
  SITE (the consolidation magnet), the regex brain (declared home →
  the domain's settled siblings → the taxonomy's vocabulary, the
  repos twin through CATEGORY_FOLDERS), THE WEIGHED EVIDENCE (the
  folder census, the 0.0-1.0 scale with banker's rounding, the CLOSE
  CALL verdict, the alignment door), the chunked LLM filing, the
  validation door (the fresh-eyes absorption, the depth law, THE
  LANDABLE MOVE with the `_v` bend pre-computed as `as`), the
  picking door + the lock law, the census stamp, and the apply pass
  (verified landings, the spoken failure ledger, the deletions
  through the machinery's own skip doors, the repos ledger
  re-point); THE GOLDEN (21 sections, paths folded) + the ported
  test_vaultscan/test_censusstamp/test_picklock/test_weighhomes/
  test_reposcan selections (215 tests); the repos banishment FLOOR
  (the gate's eyes + the burial) landed with it in
  `internal/repobanish` — T-46 completes that module (revive + the
  orchestration + test_repobanish). THE POINTER-THREADING LAW is the
  session's reality-wins find (Go struct copies broke the
  reference's dict aliasing between attach_weighed_homes and
  _validate_moves — the chain threads []*NoteRow). 1347/1347 tests
  race-clean; T-45's acceptance (the ported vault-scan suites) —
  PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-45 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…018 (**thirteen**), plus
  SESSION-010's mid-session mode-flip, plus SESSION-019's
  FOURTEENTH (mid-draft: the toolchain wiped, the repos + the draft
  survived), plus **SESSION-020's FIFTEENTH: a FULL wipe — both
  repos, the toolchain, the venv, the symlinks all gone; repaired
  per the recipes (the PAT-passed clones with the extraheader
  scrubbed from BOTH configs immediately, go1.27.2 reinstalled, the
  venv recreated, the baseline re-verified 1132 green before
  work)**. The restore recipes: start-of-session (discover the
  remote via the GitHub API with the owner's in-chat PAT passed
  ephemerally as the Basic `x-access-token:<PAT>` header; clone
  `assadigit/second-brain-project` with `git -c http.extraheader=…`
  — the Basic scheme; never store it; **SCRUB AFTER EVERY CLONE:
  `git clone -c http.extraheader=…` PERSISTS the PAT into the new
  repo's .git/config — run `git config --unset-all http.extraheader`
  in BOTH clones right after cloning, then push with the ephemeral
  `-c` form only**); GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball), venv `/home/z/venvs/gitcurator` (Python 3.12,
  stdlib-only), symlinks `/home/z/second-brain-project` and
  `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  Verify the pins at every start: `git rev-parse HEAD:go.mod` —
  expect the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite;
  do NOT bump tls-client past v1.16.0). **SESSION-015's debugging
  lesson, kept: this sandbox's output renderer EATS the two-character
  sequence `[h` (a console escape) — verify suspicious "corruption"
  with %q byte dumps before believing rendered text.**
  **SESSION-019's lesson: if the sandbox dies mid-task, CHECK whether
  the draft survived before restarting from zero — `git status` +
  file timestamps tell; a preserved draft is continued, audited
  line-by-line against the reference, never blindly trusted.**
  **SESSION-020's lesson: the toolchain rebuild is routine (the
  go.dev tarball's sha256 pins it), but the FULL wipe also took the
  venv and symlinks — the §1 recipe list covers all four artifacts;
  re-verify the baseline suite BEFORE work, not after.**

## 2. Next session: SESSION-021 — Phase 4 continues: T-46 vaulthygiene + the repobanish completion (L), then T-47 vaultseal if capacity remains

- **T-46 `internal/vaulthygiene` + complete `internal/repobanish`**
  (L): core/vault_hygiene.py + core/repo_banish.py per
  spec-core-vault-hygiene.md + spec-core-repo-banish.md (READ THE
  SPECS BEFORE THE CODE). vaulthygiene: the machine-vault healing
  (is_app_owned_bodyless, is_locked_frontmatter, run_vault_hygiene,
  heal_machine_vaults — the one-open law's shared cache).
  repobanish: COMPLETE the module — revive_banished_repos (the ♻️
  door back), run_repo_banishment_pass (the confirmation-gate
  orchestration) — the gate's eyes + the burial + the note-finder
  already landed with T-45 (`internal/repobanish/repobanish.go`,
  battery in vaultscan's suites; the split is recorded in the
  package doc.go). Acceptance: the ported test_vaultscan-adjacent
  selections already green stay green + the ported test_repobanish
  + test_hygiene-law selections (the vault_hygiene legs of
  test_picklock — the locked-bodyless law — are T-46's own).
- **T-47 `internal/vaultseal`** (M) if capacity remains:
  spec-core-vault-seal.md.
- The Phase 4 queue behind them: T-48 goodrepos + reporadar, T-49
  directorysync (REUSES T-34's rig directory doors), T-50 mirror,
  T-51 telegram (dormant per ADR-0004 Am.3).
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
  a literal EOF); a scripted timing test must keep the reference's
  UNPATCHED cadences. T-45's additions: THE POINTER-THREADING LAW
  (Python dict aliasing across function boundaries = []*T in Go —
  struct copies break it silently); vault_scan's own frontmatter
  parser is preserved BESIDE note_reads' unified parse (flow tags
  REPLACE; the strip chain without a final strip) — keep both
  exactly as the reference does; Python text-mode read(n) counts
  CHARACTERS; the _windows_eyes law rides a stat seam
  (fsExists/fsIsDir/fsIsFile), not os patches.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-020.md` — the most recent report (the
   T-45 evidence, the fifteenth degradation, the
   pointer-threading find, the repobanish-floor split).
2. `docs/port-blueprint/module-specs/spec-core-vault-hygiene.md`
   and `docs/port-blueprint/module-specs/spec-core-repo-banish.md`
   — BEFORE T-46's code.
3. The reference: `app/gitcurator/core/vault_hygiene.py` +
   `app/gitcurator/core/repo_banish.py` (the not-yet-ported
   functions: revive_banished_repos at repo_banish.py:475-527,
   run_repo_banishment_pass at 529-660; all of vault_hygiene.py) +
   the test legs (find them via `app/tests/` — test_repobanish,
   and the hygiene legs: test_picklock's hygiene test,
   test_vaultscan's one-open hygiene-sharing test, plus whatever
   suite vault_hygiene's own specs name).
4. Working examples: `internal/vaultscan` (T-45 — the newest
   module: the pointer-threading law, the stat seam, the golden
   generator with the path-folding technique, the repobanish
   floor's call shape) and `internal/handdelivery` (T-43 — the
   plain generator + the PageIsRealFn seam).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-021

- **The go-directive acknowledgement (STILL WAITING — one line
  closes it)**: the main module rides `go 1.24.1`, forced by the
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
  a new ADR — surface, don't improvise. (T-45 added zero.)
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
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

## 5. THE PROMPT TO COPY (SESSION-021)

```
You are starting SESSION-021 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-018 — thirteen — plus a MID-session degradation
   in SESSION-010, plus SESSION-019's FOURTEENTH mid-draft toolchain wipe, plus
   SESSION-020's FIFTEENTH: a FULL wipe — both repos, the toolchain, the venv,
   the symlinks all gone, repaired per the recipes). FIRST verify with
   evidence: work repo at /home/z/my-project/second-brain-project (symlinked
   /home/z/second-brain-project), read-only reference /home/z/my-project/
   GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must match testdiff/
   PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go, venv
   /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per build/TOOLCHAIN.md,
   and go.mod with modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at
   v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and
   the go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do
   not bump any of these). Then check `git status` and `git log --oneline -3`:
   SESSION-020 landed T-45 vaultscan at 5c0508e (the tree must be clean and
   origin/main at the SESSION-020 close). If anything is missing, rebuild per
   docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass it
   ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it — AND SCRUB IT: `git clone -c
   http.extraheader=…` PERSISTS the header into the new repo's .git/config, so
   run `git config --unset-all http.extraheader` in BOTH clones right after
   cloning; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck` — and if the sandbox dies MID-TASK, check whether the draft
   survived before restarting from zero: a preserved draft is continued and
   audited, never blindly trusted) and re-verify the suite (1347 green) before
   work. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4 state
   with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at be688bd,
   T-44 at 71e5e34, and T-45 at 5c0508e, 1347 tests race-clean, the fifteenth
   degradation story, the clone-config PAT-scrub hazard, the environment
   recipes, the `[h` output-renderer lesson, the mid-task-death lesson, and
   T-45's pointer-threading lesson) → docs/sessions/SESSION-020.md (the most
   recent report — including T-45's golden + the five ported selections, the
   repos banishment floor split, and the two reality-wins finds).

3. THIS SESSION IS T-46 (L): Phase 4 continues — internal/vaulthygiene (all of
   core/vault_hygiene.py per docs/port-blueprint/module-specs/
   spec-core-vault-hygiene.md) + the COMPLETION of internal/repobanish
   (revive_banished_repos + run_repo_banishment_pass per
   spec-core-repo-banish.md — the gate's eyes, the burial, and the note-finder
   already landed with T-45 in internal/repobanish/repobanish.go; their
   behavior battery rides internal/vaultscan's suites and must stay green)
   (READ THE SPECS BEFORE THE CODE). Acceptance: the ported selections of the
   reference's own suites for both modules (test_repobanish; the vault_hygiene
   legs incl. test_picklock's locked-bodyless law and test_vaultscan's
   one-open hygiene-sharing test — check the hygiene specs' own test lists),
   with vaultscan's 215-test battery staying green throughout; the house
   generator pattern applies if a golden is warranted —
   testdiff/goldens/dump_vaultscan.py is the newest working example, including
   the path-folding technique. If capacity remains after T-46 goes green +
   pushes: T-47 internal/vaultseal per spec-core-vault-seal.md — but close
   cleanly at a natural boundary if capacity runs out; one task done properly
   with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-46's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-021.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-022.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

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
- **Phase 3 CLOSED (SESSION-011–014; all five tasks)**: T-30
  `internal/llmclient` (`f2073b5`), T-31 `internal/webfetch`
  (`4d93877`), T-32 `internal/webextract` (`833b21b`), T-33
  `internal/connectioncheck` (`e5846bd`), T-34 `internal/workerclient`
  (`6ffc4e9`) — 872 tests race-clean at the close.
- **Phase 4 OPEN (SESSION-015–017)**: T-40 `internal/websitepipeline`
  at `8098ac4` (the per-link state machine, THE BANISHMENT, the
  decomtable floor, the 51-section differential golden); T-41
  `internal/websitepipeline/directorynote.go` at `f528c22` (THE LAW
  sweep + the Website Directory regeneration, the output golden, three
  reality-wins primitive fixes); **T-42 `internal/githubpipeline` +
  `internal/batch/summary.go` landed at `e92e4e7` (SESSION-017) — the
  GUI processing worker's GitHub loop: the per-repo gate chain, the
  get_repo error paths (403+rate wait / 401 anonymous / the 404
  strike), the two-layer dedupe, fork→parent, the enrichment + the
  banner goroutine, the LLM 3-attempt ladder + the injection guard,
  the disk-full protocol, the reports + the undo file + the scorecard;
  THE GOLDEN (12 sections, frozen clocks, generated from the
  unmodified reference through minimal GUI-layer import stubs) + the
  behavior battery (the rate-limit laws, the 404-strike flow, the 401
  fallback); the reality-wins float fix in notebuilder (cred_score is
  the round(x,1) FLOAT — "90.0/100"). 984/984 tests race-clean;
  T-42's acceptance (batch summary goldens; 404-strike tests;
  rate-limit law tests) — PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-42 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…017 (**twelve so far**) PLUS
  SESSION-010's mid-session mode-flip event. The restore recipes:
  start-of-session (discover the remote via the GitHub API with the
  owner's in-chat PAT passed ephemerally as the Basic
  `x-access-token:<PAT>` header — SESSION-013…017 all had to do
  this from a bare sandbox; clone `assadigit/
  second-brain-project` with `git -c http.extraheader="Authorization:
  Basic $(printf 'x-access-token:%s' "$PAT" | base64)"` — the Basic
  scheme; never store it; GitCurator @ `98f4fc7` (verify against
  `testdiff/PINNED-GITCURATOR-REV`), go1.27.2 → `/home/z/.local/go`
  (go.dev tarball), venv `/home/z/venvs/gitcurator` (Python 3.12,
  stdlib-only), symlinks `/home/z/second-brain-project` and
  `/home/z/GitCurator`; mid-session mode-flip (`git checkout-index
  -f -a` restores modes from the index; `git fsck` to sanity-check).
  **SESSION-017's NEW hazard (SCRUB AFTER EVERY CLONE): `git clone -c
  http.extraheader=…` PERSISTS the header into the new repo's
  .git/config — the PAT sits base64-decodable in the file. After every
  clone: `git config --unset-all http.extraheader` in BOTH clones
  (the work repo AND the GitCurator reference), then push with the
  ephemeral `-c` form only.** Verify the pins at every start: `git
  rev-parse HEAD:go.mod` — expect the v1.36.1/tls-client/1.24.1 bytes
  (do NOT bump sqlite; do NOT bump tls-client past v1.16.0).
  **SESSION-015's debugging lesson, kept for the next sessions: this
  sandbox's output renderer EATS the two-character sequence `[h` (a
  console escape) — verify suspicious "corruption" with %q byte dumps
  before believing rendered text.**

## 2. Next session: SESSION-018 — Phase 4 continues: T-43 handdelivery (L), then T-44 chrometabs (L) if capacity remains

- **T-43 `internal/handdelivery`** (L): completes the module T-40
  floored — core/hand_delivery.py:88-907 per
  `spec-core-hand-delivery.md` (READ THE SPEC BEFORE THE CODE): the
  Chrome launch family, the queue/consume/harvest/stamps, the
  batch-level pickers. Acceptance per the task list: the ported
  `test_handdelivery`/`test_handharvest` cases (the reference's own
  suites). `internal/handdelivery` already exists as the T-40 floor
  (handdelivery.go with CollectDelivered + the Clock seam) — extend,
  never duplicate. If a golden is warranted, the house generator
  pattern applies (testdiff/goldens/ — dump_websitepipeline.py,
  dump_directorynote.py, and dump_githubpipeline.py are the working
  examples; dump_githubpipeline.py additionally shows the GUI-layer
  sys.modules stub technique for modules whose imports pull PyQt6 /
  PyGithub / ollama — hand_delivery.py is core/ and imports none of
  them, so the plain pattern should suffice).
- **T-44 `internal/chrometabs`** (L) if capacity remains:
  core/chrome_tabs.py:484-2563 per spec-core-chrome-tabs.md.
- The Phase 4 queue behind them: T-45 vaultscan, T-46 vaulthygiene +
  repobanish, T-47 vaultseal, T-48 goodrepos + reporadar, T-49
  directorysync (REUSES T-34's rig directory doors), T-50 mirror,
  T-51 telegram (dormant per ADR-0004 Am.3).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for any
  store-adjacent column; `store.PyISOFormat*` for timestamp writes;
  atomic saves; the dry-run seam at call sites; rows collected before
  nested statements; tests never weakened; reality wins over specs
  (fix the spec in the same commit); goldens are GENERATED by running
  the unmodified Python app, never hand-typed; no per-run facts in
  goldens; the frozen-clock seam (`clockNow`/`handdelivery.Clock`)
  exists for exactly this. **T-41's additions: text-mode reads
  translate universal newlines before any cap or parse
  (`helpers.PyNewlines`); invalid UTF-8 decodes with the
  maximal-subpart U+FFFD law; vault walks are os.walk's
  files-before-subdirs order, never WalkDir. T-42's additions: the
  house JSON decoder's shapes (storage.OrderedMap + json.Number /
  int64 scalars) must be handled wherever LLM/JSON output is read —
  llmclient.ExtractJSON never yields map[string]any directly; floats
  render through helpers.PyFloatRepr (credibility "90.0", not "90").**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-017.md` — the most recent report (the T-42
   evidence, the twelfth wipe + rebuild, the notebuilder float
   reality-wins fix, the GUI-layer stub generator technique, the
   clone-config PAT hazard).
2. `docs/port-blueprint/module-specs/spec-core-hand-delivery.md` —
   BEFORE T-43's code.
3. The reference: `app/gitcurator/core/hand_delivery.py:88-907` +
   the test legs (`app/tests/test_handdelivery.py`,
   `app/tests/test_handharvest.py`).
4. Working examples: `internal/websitepipeline` (the golden-generator
   pattern with frozen clocks), `internal/githubpipeline` (T-42 —
   the Runner seam pattern + the GUI-layer stub generator
   `testdiff/goldens/dump_githubpipeline.py`), and the existing
   `internal/handdelivery/handdelivery.go` (T-40's floor).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-018

- **The go-directive acknowledgement (STILL WAITING — one line closes
  it)**: the main module rides `go 1.24.1`, forced by the
  owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has the
  record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers x/net/html,
  approve an ADR and a future session swaps
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
  a new ADR — surface, don't improvise.
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves, the websites-phase wiring — the
  Runner's seams exist for exactly this); the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-018)

```
You are starting SESSION-018 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-017 — twelve so far — plus a MID-session
   degradation in SESSION-010; SESSION-013…017 all started from a BARE sandbox
   and had to discover the remote via the GitHub API with the PAT before
   cloning). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked
   /home/z/second-brain-project), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must
   match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go,
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, and go.mod with modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/
   x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these). Then
   check `git status` and `git log --oneline -3`: SESSION-017 landed T-42
   githubpipeline + batch at e92e4e7 (the tree must be clean and origin/main
   at the SESSION-017 close). If anything is missing, rebuild per
   docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass it
   ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it — AND SCRUB IT: `git clone -c
   http.extraheader=…` PERSISTS the header into the new repo's .git/config, so
   run `git config --unset-all http.extraheader` in BOTH clones right after
   cloning; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck`) and re-verify the suite (984 green) before work. PUSH AFTER
   EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4 state
   with T-40 at 8098ac4, T-41 at f528c22, and T-42 at e92e4e7, 984 tests
   race-clean, the twelfth full wipe + bare-sandbox rebuild story, the
   clone-config PAT-scrub hazard, the environment recipes, and the `[h`
   output-renderer lesson) → docs/sessions/SESSION-017.md (the most recent
   report — including T-42's golden through the GUI-layer import stubs, the
   notebuilder cred_score float reality-wins fix, and the house JSON-decoder
   shape laws that now carry forward).

3. THIS SESSION IS T-43 (L): Phase 4 continues — internal/handdelivery per
   docs/port-blueprint/module-specs/spec-core-hand-delivery.md (READ THE SPEC
   BEFORE THE CODE). Completes the module T-40 floored: core/hand_delivery.py
   88-907 — the Chrome launch family, the queue/consume/harvest/stamps, the
   batch-level pickers. internal/handdelivery already exists from T-40
   (handdelivery.go with CollectDelivered + the Clock seam) — extend, never
   duplicate. Acceptance: the ported test_handdelivery/test_handharvest cases
   (the reference's own suites; the house generator pattern applies if a
   golden is warranted — testdiff/goldens/dump_githubpipeline.py is the newest
   working example). If capacity remains after T-43 goes green + pushes: T-44
   internal/chrometabs (L) per spec-core-chrome-tabs.md — but close cleanly at
   a natural boundary if capacity runs out; one task done properly with proof
   beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-43's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-018.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-019.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

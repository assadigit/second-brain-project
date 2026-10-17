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
- **Phase 4 OPEN (SESSION-015–019)**: T-40 `internal/websitepipeline`
  at `8098ac4`; T-41 `internal/websitepipeline/directorynote.go` at
  `f528c22`; T-42 `internal/githubpipeline` + `internal/batch/
  summary.go` at `e92e4e7`; T-43 `internal/handdelivery` at `be688bd`;
  **T-44 `internal/chrometabs` landed at `71e5e34` (SESSION-019) —
  THE FIFTH DOOR COMPLETE: the hand-rolled RFC 6455 frame codec, the
  DevTools socket (handshake/call/ping-pong/close/timeout laws), the
  Chrome session ladders (launch → container → headless, adoption via
  DevToolsActivePort, the graceful Browser.close + tree kill), the
  real-profile rung (the attach link + chrome-attach.bat), the proxy
  legs with per-site route memory, the load watch (settle/restart/
  challenge-budget), page_is_real WIRED into the PageIsRealFn seam
  (the one SESSION-018 staged) with the deferred auto-door discard
  battery landed, the candidates, the delivery, the fetch driver
  honesty; THE GOLDEN (9 sections, the random mask bytes and
  handshake key folded out) + the ported test_fifthdoor selections
  (63 module tests + 2 pipeline legs — the full circle rides
  websitepipeline's EXTERNAL test package, the import-cycle breaker);
  pyurl.go hand-rolls CPython 3.12's lenient urlsplit INCLUDING the
  bracket law (Go's url.Parse rejects what Python accepts). 1132/1132
  tests race-clean; T-44's acceptance (the ported test_fifthdoor
  selections) — PASS.**
- **SESSION-009's lesson stays institutionalized: PUSH AFTER EVERY
  TASK** — T-44 was pushed the moment its battery went green.
- Environment notes (STANDING — the sandbox is REBUILDABLE): full
  wipes at the starts of SESSION-005…018 (**thirteen**), plus
  SESSION-010's mid-session mode-flip, plus **SESSION-019's
  FOURTEENTH degradation: mid-draft this time — the toolchain/venv/
  symlinks wiped while the repos AND the session's own in-progress
  draft survived; repaired per the recipes (draft preserved aside →
  fsck → `git checkout-index -f -a` → draft restored; the toolchain
  rebuilt)**. The restore recipes: start-of-session (discover the
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
  expect the v1.36.1/tls-client/1.24.1 bytes (do NOT bump sqlite; do
  NOT bump tls-client past v1.16.0). **SESSION-015's debugging
  lesson, kept: this sandbox's output renderer EATS the two-character
  sequence `[h` (a console escape) — verify suspicious "corruption"
  with %q byte dumps before believing rendered text.**
  **SESSION-019's lesson: if the sandbox dies mid-task, CHECK whether
  the draft survived before restarting from zero — `git status` +
  file timestamps tell; a preserved draft is continued, audited
  line-by-line against the reference, never blindly trusted.**

## 2. Next session: SESSION-020 — Phase 4 continues: T-45 vaultscan (L), then T-46 vaulthygiene + repobanish if capacity remains

- **T-45 `internal/vaultscan`** (L): core/vault_scan.py:548-3353
  per `spec-core-vault-scan.md` (READ THE SPEC BEFORE THE CODE) —
  the vault scan: the radar (what the vault holds), the weighing,
  the validation, and the apply pass. Acceptance per the task list:
  the ported vault-scan suites (the reference's own tests; the house
  generator pattern if a golden is warranted —
  `testdiff/goldens/dump_chrometabs.py` is the newest working
  example, including the fold-out technique for random bytes).
- **T-46 `internal/vaulthygiene` + `internal/repobanish`** (L) if
  capacity remains: spec-core-vault-hygiene.md +
  spec-core-repo-banish.md.
- The Phase 4 queue behind them: T-47 vaultseal, T-48 goodrepos +
  reporadar, T-49 directorysync (REUSES T-34's rig directory doors),
  T-50 mirror, T-51 telegram (dormant per ADR-0004 Am.3).
- Standing port laws carry: THE CAST READ LAW (ADR-0003 Am.1) for
  any store-adjacent column; `store.PyISOFormat*` for timestamp
  writes; atomic saves; the dry-run seam at call sites; rows
  collected before nested statements; tests never weakened; reality
  wins over specs (fix the spec in the same commit); goldens are
  GENERATED by running the unmodified Python app, never hand-typed;
  no per-run facts in goldens; the frozen-clock seams
  (`clockNow`/`handdelivery.Clock`) exist for exactly this.
  **T-41's additions: text-mode reads translate universal newlines
  before any cap or parse (`helpers.PyNewlines`); invalid UTF-8
  decodes with the maximal-subpart U+FFFD law; vault walks are
  os.walk's files-before-subdirs order, never WalkDir. T-42's
  additions: the house JSON decoder's shapes (storage.OrderedMap +
  json.Number / int64 scalars) must be handled wherever LLM/JSON
  output is read; floats render through helpers.PyFloatRepr. T-43's
  additions: `pyBytesBlank` is the blank-body law (Python bytes.strip
  is ASCII-only); Go MAP ranges are RANDOM — any Python
  dict-insertion-order tie must ride a tracked order slice. T-44's
  additions: Go's `url.Parse` is NOT CPython's urlsplit — the
  chrometabs.pyurl shapes are the house's lenient-parse reference;
  Go's `cmd.Wait()` closes the pipe (any read error + process-exited
  = the exit sentence, never wait for a literal EOF); a scripted
  timing test must keep the reference's UNPATCHED cadences when the
  script chunks are timed against them.**

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-019.md` — the most recent report (the T-44
   evidence, the fourteenth degradation + the draft-preserving
   repair, the two reality-wins finds, the wiring).
2. `docs/port-blueprint/module-specs/spec-core-vault-scan.md` —
   BEFORE T-45's code.
3. The reference: `app/gitcurator/core/vault_scan.py:548-3353` + the
   test legs (find them via `app/tests/` — the vault-scan suites).
4. Working examples: `internal/chrometabs` (T-44 — the newest
   module: the seams pattern, the golden generator with the
   fold-out technique, `pyurl.go`'s CPython-parity law) and
   `internal/handdelivery` (T-43 — the plain generator +
   the PageIsRealFn seam T-44 wired).
5. ADR-0003 (the CAST read law) before any store-adjacent column.

## 4. Known constraints / owner actions for SESSION-020

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
  a new ADR — surface, don't improvise. (T-44's CDP WebSocket stayed
  hand-rolled stdlib — the spec's old gorilla/websocket note is
  superseded by the SESSION-019 amendment.)
- Standing: secrets / owner data never in git; tests never weakened;
  reality wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.
- Deferred items on the books (not blockers): the GUI worker shell
  task (the mode wiring, the Ollama warmup, the inputs fetchers, the
  interactive doors' GUI halves — including test_fifthdoor §9b's
  end-of-delivery modal legs — the websites-phase wiring; the
  Runner's seams exist for exactly this); the note_state SURGERY
  path's CRLF round-trip differential (Phase 4+); the Windows live
  legs (Phase 7 clean-machine checklist); a Windows VM without
  WebView2 for T-70 (optional).

## 5. THE PROMPT TO COPY (SESSION-020)

```
You are starting SESSION-020 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (full wipes at the starts of
   SESSION-005 through SESSION-018 — thirteen — plus a MID-session degradation
   in SESSION-010, plus SESSION-019's FOURTEENTH: mid-draft, the toolchain
   wiped while the repos and the session's own draft survived). FIRST verify
   with evidence: work repo at /home/z/my-project/second-brain-project
   (symlinked /home/z/second-brain-project), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator; must
   match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 → /home/z/.local/go,
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, and go.mod with modernc.org/sqlite at v1.36.1,
   bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 + golang.org/
   x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these). Then
   check `git status` and `git log --oneline -3`: SESSION-019 landed T-44
   chrometabs at 71e5e34 (the tree must be clean and origin/main at the
   SESSION-019 close). If anything is missing, rebuild per
   docs/NEXT-SESSION.md §1 (clone via the PAT I give you when you ask — pass it
   ephemerally with git -c
   http.extraheader="Authorization: Basic $(printf 'x-access-token:%s' "$PAT" |
   base64)", the Basic scheme; never store it — AND SCRUB IT: `git clone -c
   http.extraheader=…` PERSISTS the header into the new repo's .git/config, so
   run `git config --unset-all http.extraheader` in BOTH clones right after
   cloning; a mid-session mode-flip is repaired with `git checkout-index -f -a`
   after `git fsck` — and if the sandbox dies MID-TASK, check whether the draft
   survived before restarting from zero: a preserved draft is continued and
   audited, never blindly trusted) and re-verify the suite (1132 green) before
   work. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records the Phase 4 state
   with T-40 at 8098ac4, T-41 at f528c22, T-42 at e92e4e7, T-43 at be688bd, and
   T-44 at 71e5e34, 1132 tests race-clean, the fourteenth degradation +
   draft-preserving repair story, the clone-config PAT-scrub hazard, the
   environment recipes, the `[h` output-renderer lesson, and the mid-task-death
   lesson) → docs/sessions/SESSION-019.md (the most recent report — including
   T-44's golden + the ported suites, the PageIsRealFn wiring, the pyurl
   CPython-parity law, and the two reality-wins finds).

3. THIS SESSION IS T-45 (L): Phase 4 continues — internal/vaultscan per
   docs/port-blueprint/module-specs/spec-core-vault-scan.md (READ THE SPEC
   BEFORE THE CODE). core/vault_scan.py:548-3353 — the vault scan: the radar,
   the weighing, the validation, the apply pass. Acceptance: the ported
   vault-scan suites (the reference's own suite; the house generator pattern
   applies if a golden is warranted — testdiff/goldens/dump_chrometabs.py is
   the newest working example, including the fold-out technique for random
   bytes). If capacity remains after T-45 goes green + pushes: T-46
   internal/vaulthygiene + internal/repobanish per spec-core-vault-hygiene.md
   + spec-core-repo-banish.md — but close cleanly at a natural boundary if
   capacity runs out; one task done properly with proof beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git; no new
   dependencies without an ADR (the authorized set is exactly modernc.org/
   sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; x/net/html stays OUT unless the owner approves an ADR); surface
   blockers with evidence (hash-level, not display) instead of improvising
   past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (Phase 4
   has no single exit gate — evaluate T-45's own acceptance honestly), commit
   + push everything (ask me for the GitHub PAT if it is not already available
   — never store it anywhere), write docs/sessions/SESSION-020.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-021.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

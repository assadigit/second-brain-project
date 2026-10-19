# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker
  HTTP API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity.
- **Phase 2 CLOSED** (SESSION-006–010): T-20…T-25 — 425 tests
  race-clean at the close.
- **Phase 3 CLOSED** (SESSION-011–014): T-30…T-34 — 872 tests
  race-clean at the close.
- **Phase 4 CLOSED** (SESSION-015–026): T-40…T-50 + the Phase-4 closer
  (the 30-link offline golden runner equivalent) — 1828 tests at the
  close.
- **Phase 5 CLOSED (SESSION-027–032)**: T-60 (the CLI shell + the
  command wirings), T-61 (the paint box + the console hardening — the
  close-out audit), T-62 (the run engine + the drivers), T-63 (the
  launchers + THE DISPATCH FLIP), **T-64 COMPLETE (SESSION-032,
  `42739b3` + `9dc79db`)** — the differential corpus full run: all 31
  generators × two passes GREEN (the two-run byte-identical proof +
  committed == regenerated + the four atomic pairs' own law), five
  generator-side determinism finds fixed, ZERO Go-port divergences,
  the report archived at
  `docs/poc-artifacts/t-64-corpus-full-run-report.md` — and
  **T-65 COMPLETE (SESSION-032, `a39cb3d`)** — the CLI parity release:
  build/mkzip (the minimal deterministic slice per doc 10 §2's laws +
  5 law tests), the GUI-side launcher pair at the repo root
  (GitCurator.bat with the honest GUI-milestone card +
  GitCurator-DRY-RUN.bat), the golden's launchers section grown to the
  reference's own app-root pair (105,092 payload bytes, two-run
  byte-identical, a fourth mutation surface), config.example.json +
  WINDOWS-QUICKSTART.md, **the tag `v0.92.0-go.1` at `a39cb3d` + the
  GitHub release 416838956 with
  `GitCurator-v0.92.0-go.1-windows.zip` (12 files, 15,988,399 bytes,
  sha256 `d2679a1de9d600e11d2e590016ed5359ca3fe0f841802289506f4cdd43f
  ecebb`, double-build identical, the asset round-trip verified)**.
  **1927 tests race-clean over 34 packages** at the close.
- **Push state: CLEAN** — SESSION-032 pushed after every task
  (`42739b3` → `9dc79db` → `a39cb3d` + the tag + the release); verify
  `git status` is in sync at the start (the tip at or beyond
  `a39cb3d`).
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
  - **THE T-65 ZIP-COMPOSITION RESOLUTION** (SESSION-032): the minimal
    deterministic mkzip slice landed at T-65 (the spec-supported
    reading: F5 assigns the flattening to T-65's packaging; the
    acceptance demands the asset; the owner's law — no zip, no
    release); T-80/Phase 7 keeps the FULL builder (the CI
    double-build ritual + the two-exe GUI composition + the release
    automation) per traceability F4.
  - **THE CORPUS-REGENERATION LAW** (SESSION-032, T-64): a golden that
    embeds per-run bytes (raw mkdtemp roots, file mtimes, epoch
    floats, set-iteration orders) can never regenerate byte-identical
    — fold the roots at the payload's end (the replays' fold regexes
    make the folded prefix a no-op), design fixtures so order-dependent
    winners have nothing to flip (same stem, different folder), and
    mask the documented per-run facts ONLY in the corpus driver's
    stability digest (the goldens stay raw — the replays prove each
    pair's consistency through the exact frozen values).
  - **THE DISPATCH-COMPLETION LAW** (SESSION-031): a command is only
    LANDED when Run's dispatch routes the argv to it — unit tests
    calling the cmd* methods directly keep the suite green while the
    user-facing door stays stale; the acceptance drives `Run(argv)`.
  - **THE STATE-CHAIN LAW** (SESSION-030): a golden replay of
    sequential sections is ONE state chain — an ISOLATED replay
    (e.g. a mutation test with a fresh buildEnv) must reproduce the
    payload's own pre-state.
  - **THE MUTATION-TEETH LAW** (SESSION-030): a mutation test runs
    the REAL replay against the mutated golden and must FAIL, while
    the unmutated golden must still PASS.
  - **THE CONTEXT-DEATH CONTINUATION recipe** (SESSION-030/021): git
    status + `git log --oneline -5` + file timestamps tell what
    ACTUALLY survived; audit the in-flight work (never blindly trust
    it), repair, run the FULL battery, then land. (SESSION-032 added:
    a full tool outage may just need waiting out — probe periodically,
    then continue exactly where you stopped.)
  - **THE DSN-EMPTY-PATH HAZARD** (SESSION-030): never pass "" as a
    db path; `NewWebsiteStateDB("cache.db")` is the AppDir-anchored
    default.
  - **THE FIXED-LENGTH ROOTS law** (SESSION-029): the replay's
    buildEnv roots must be exactly as wide as the generator's
    mkdtemp roots (`/tmp/gc-cli-app-goreplay` +
    `/tmp/gc-cli-vaults-goreplay`; keep the 8-rune suffix width).
  - **THE SHALLOW-COPY SNAPSHOT law** (SESSION-029): a golden section
    recording a config the driven call MUTATES must copy.deepcopy
    BEFORE the call.
  - The fake-clock arithmetic (1000.0 → 1000.042 renders 42 ms via
    `int((t1-t0)*1000)`); exception class-name tokens via
    `type("TimeoutError", …)` (the SHAPE is the contract); a golden
    section capturing a RETURN VALUE cannot use the stdout-capture
    helper.
  - **THE CD-DISCIPLINE LAW** — every git/go command carries its own
    cd (or go's -C flag); never commit outside the work repo; NEVER
    `git checkout -- <dir>` blindly after editing files in it (it
    wipes your edits — SESSION-032 paid this one: restore only the
    churned goldens you mean to restore).
  - **THE RENDERER-EATS-ESCAPE-PAIRS lesson** — write commit messages
    to a FILE and commit with -F, verify with `git log --format=%B`.
  - **THE WRITE-TOOL REDACTION lesson** — the Write/Edit tools REDACT
    token-shaped strings at file-write time (craft safe-shaped test
    tokens, never ghp_-shaped; byte-level secret-scan before every
    commit).
  - **THE BACKGROUND-PROCESS lesson** — detached processes do not
    survive tool-call boundaries; run long suites in FOREGROUND
    CHUNKS.
  - **PUSH AFTER EVERY TASK** — ask the owner for the PAT the MOMENT
    a push is due; never store it.

## 2. Next session: SESSION-033 — Phase 6's entry: PoC C (the Wails GUI shell gate), then T-71 if capacity remains

- **Verify the push state first** (§1): SESSION-032 pushed everything
  (the tree ends clean); confirm `git status` is in sync before work
  (and keep pushing after every task).
- **Phase 6's entry conditions** (roadmap §Phase 6): Phase 5 exit ✓
  (this session) + **PoC C: Wails** (ADR-0001's gate: a themed window
  + log desk + one job; startup ≤ 1.5 s; a 10k-line log flood stays
  interactive; the verdict recorded). **T-70 (M)** is the PoC row —
  read ADR-0001 + the roadmap's Phase-6 entry + doc 07 §B.4 (the
  performance criteria) FIRST. NOTE: the Wails CLI is a build-time
  dependency — the no-new-dependencies law applies to go.mod's RUNTIME
  set; the Wails toolchain question (its own module, pinned version,
  the npm/pnpm frontend step) is an ADR-0002-style record to surface
  in the session report BEFORE adding anything.
- **Then, only if capacity remains, T-71 (L)** per doc 04 §2/§5:
  `internal/app` commands + event bus + ask-gates + lifecycle/locks —
  the command contract tests + the lock/watchdog tests. One slice done
  properly with proof beats a rushed whole; closing cleanly at a
  natural boundary is always allowed.
- If capacity runs short: T-70 done properly with the verdict recorded
  first — it is Phase 6's gate.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-032.md` — the most recent report (T-64 +
  T-65 COMPLETE; Phase 5 CLOSED; the release).
2. `docs/port-blueprint/decisions/0001-*.md` (ADR-0001 — the GUI
  framework gate + PoC C's terms) + `docs/port-blueprint/
  12-migration-roadmap.md` §Phase 6 (the entry + the exit gate).
3. `docs/port-blueprint/07-*` §B.4 — the GUI performance criteria.
4. `docs/port-blueprint/module-specs/spec-gui-theme-icons.md` +
  `spec-gui-ui.md` — the theme tokens + the layout shell (T-70's
  themed window + T-72's foundation feed on these).
5. `docs/port-blueprint/13-implementation-task-list.md` — the T-70/T-71
  rows + Phase 6's full book.

## 4. Known constraints / owner actions for SESSION-033

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The Wails toolchain decision** (§2): surface it as an ADR-0002
  Am.3-style record (the pinned wails CLI version, the frontend build
  step, the CGO story for the GUI exe) BEFORE the first `wails`
  invocation — the owner may pre-decide.
- **The go-directive acknowledgement (STILL WAITING — twenty-second
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2 has
  the record + the rollback path).
- **The x/net/html question is a KEEP-OR-SWAP, not a blocker** (T-32
  shipped the hand-rolled tokenizer; if the owner prefers
  x/net/html, approve an ADR and a future session swaps
  `internal/webextract/parser.go`'s scan internals).
- **The release feedback loop**: the owner may report v0.92.0-go.1
  experiences (the zip, the launchers, the DRY-RUN on real data) —
  record them in the session report; runtime findings route to Phase
  7's checklist or immediate fixes per severity.
- **Deferred test debt on the books (recorded with evidence, not
  blockers)**: test_noteredo's remaining un-ported selections
  (ScanMasterRedoRows is the top item); the note_state SURGERY path's
  CRLF round-trip differential (Phase 4+); the Windows live legs +
  the launchers' clean-VM smoke (Phase 7 checklist; doc 10 §7); a
  Windows VM without WebView2 for T-70's GUI legs (optional).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB the persisted
  extraheader from .git/config after the clone); goldens are
  GENERATED by running its unmodified code.
- **Dependency law**: `modernc.org/sqlite` v1.36.1;
  `bogdanfinn/tls-client` v1.16.0; `golang.org/x/net` v0.48.0 +
  `golang.org/x/text` v0.32.0 (direct); ANY other add needs a new
  ADR — surface, don't improvise. (SESSION-032 added zero.)
- Standing: secrets / owner data never in git (byte-level
  secret-scan before every commit); tests never weakened; reality
  wins over specs (fix the spec in the same commit); surface
  blockers with evidence (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-033)

```
You are starting SESSION-033 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (twenty-six full wipes across
   SESSION-005..032 plus mid-session degradations — SESSION-032 itself survived
   a ~40-minute total tool outage by waiting it out and continuing where it
   stopped; the continuation recipe is PROVEN: git status + git log --oneline -5
   + file timestamps tell what ACTUALLY survived; a preserved commit or
   in-flight draft is continued and AUDITED — never blindly trusted, never
   restarted from zero). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 → /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON), venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, and go.mod with
   modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at v1.16.0,
   golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and the go
   directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do not bump
   any of these). Then check `git status` and `git log --oneline -3`: SESSION-032
   landed T-64 COMPLETE (the differential corpus full run at 42739b3 + 9dc79db:
   all 31 generators × two passes GREEN — the two-run byte-identical proof +
   committed == regenerated + the four atomic pairs' own law; five generator-side
   determinism finds fixed, ZERO Go-port divergences; the report archived at
   docs/poc-artifacts/t-64-corpus-full-run-report.md) + T-65 COMPLETE (a39cb3d:
   build/mkzip the minimal deterministic slice + the GUI-side launcher pair +
   the golden's grown launchers section + config.example.json +
   WINDOWS-QUICKSTART.md + THE RELEASE — tag v0.92.0-go.1, GitHub release
   416838956, GitCurator-v0.92.0-go.1-windows.zip 12 files / 15,988,399 bytes /
   sha256 d2679a1de9d600e11d2e590016ed5359ca3fe0f841802289506f4cdd43fecebb,
   double-build identical, the asset round-trip verified) and PUSHED everything
   after every task — 1927 tests race-clean over 34 packages. Verify git status
   shows up to date with origin/main (the tip at or beyond a39cb3d). Use the PAT
   only ephemerally: git -c http.extraheader="Authorization: Basic $(printf
   'x-access-token:%s' "$PAT" | base64)" and never store it — AND SCRUB IT if
   you ever CLONE with it (a clone with -c http.extraheader PERSISTS the header
   into the new clone's .git/config: git config --unset-all http.extraheader in
   BOTH clones right after; the GIT_CONFIG_COUNT env-var form avoids the
   persistence entirely). Re-verify the suite before work if anything looks off.
   The operational laws, newest first: THE T-65 ZIP-COMPOSITION RESOLUTION (the
   minimal mkzip slice landed at T-65 per F5 + the acceptance + the owner's
   no-zip-no-release law; T-80/Phase 7 keeps the full builder per F4 — do not
   re-litigate); THE CORPUS-REGENERATION LAW (a golden embedding per-run bytes —
   raw mkdtemp roots, file mtimes, epoch floats, set-iteration orders — can
   never regenerate byte-identical: fold the roots at payload end, design
   order-dependent fixtures so winners have nothing to flip, mask documented
   per-run facts ONLY in the corpus driver's stability digest); THE
   DISPATCH-COMPLETION LAW (a command is only LANDED when Run's dispatch routes
   the argv to it); THE STATE-CHAIN LAW (a golden replay of sequential sections
   is ONE state chain; an ISOLATED replay must reproduce the payload's own
   pre-state); THE MUTATION-TEETH LAW (the mutated golden must FAIL the real
   replay while the unmutated golden PASSES); the DSN-empty-path hazard (never
   pass "" as a db path; NewWebsiteStateDB("cache.db") is the AppDir-anchored
   default); FIXED-LENGTH ROOTS (the replay's roots stay exactly as wide as the
   generator's mkdtemp roots); SHALLOW-COPY SNAPSHOT (deep-copy a config the
   driven call mutates BEFORE the call); the fake-clock arithmetic (1000.0 →
   1000.042 renders 42 ms via int((t1-t0)*1000)); exception class-name tokens
   via type("TimeoutError", …) — the SHAPE is the contract; a golden section
   capturing a RETURN VALUE cannot use the stdout-capture helper; (a) THE
   CD-DISCIPLINE LAW — every git/go command carries its own cd; never commit
   outside the work repo; NEVER git checkout -- <dir> blindly after editing
   files in it (it wipes your edits — restore only what you mean to restore);
   (b) the renderer-eats-escape-pairs lesson — commit messages via -F FILE,
   verify with 'git log --format=%B'; (c) the Write/Edit tool REDACTS
   token-shaped strings at file-write time (craft safe-shaped test tokens;
   byte-level secret-scan before every commit); (d) detached background
   processes DO NOT SURVIVE tool-call boundaries — run long suites in FOREGROUND
   CHUNKS. PUSH AFTER EVERY TASK this session.

2. Read, in order: the repo README.md (Phase 5 CLOSED — T-64 + T-65 at a39cb3d)
   → docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-032's close: Phase 5 CLOSED, the release
   live, the newest laws) → docs/sessions/SESSION-032.md (the most recent
   report).

3. THIS SESSION IS PHASE 6'S ENTRY: T-70 PoC C THE WAILS GUI SHELL GATE (then
   T-71 internal/app only if capacity remains): per ADR-0001's PoC C terms +
   the roadmap §Phase 6 entry + doc 07 §B.4 (READ THEM FIRST — the gate's
   authority), build the PoC: the themed window (the spec-gui-theme-icons
   tokens), the log desk, ONE real job wired through — startup ≤ 1.5 s and a
   10k-line log flood staying interactive, the verdict recorded per ADR-0001's
   gate discipline. IMPORTANT BEFORE THE FIRST WAILS INVOCATION: the Wails
   toolchain is a build-time dependency — surface the pinned-version + frontend-
   step + CGO-story record (ADR-0002 Am.3-style) BEFORE adding anything; the
   runtime go.mod set stays exactly the authorized four. If capacity remains
   after T-70's verdict: T-71 per doc 04 §2/§5 (internal/app: commands + event
   bus + ask-gates + lifecycle/locks — the command contract tests + the
   lock/watchdog tests). If capacity runs short: T-70 done properly with the
   verdict recorded first — it is the gate; one slice done properly with proof
   beats a rushed whole.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized runtime set is
   exactly modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0,
   x/text v0.32.0 — do not bump any; x/net/html stays OUT unless the owner
   approves an ADR; the Wails build-toolchain record is this session's one
   expected addition, surfaced BEFORE use); surface blockers with evidence
   (hash-level, not display) instead of improvising past them.

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-70's
   own acceptance per ADR-0001 — and T-71's if reached), commit + push
   everything (ask me for the GitHub PAT the moment a push is due — never store
   it anywhere; never end the session push-pending), write
   docs/sessions/SESSION-033.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus the
   prompt I should copy to start SESSION-034.

One task done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

# NEXT-SESSION — the state card

> **Rewritten at every session close (per `SESSION-PROTOCOL.md`).**
> Owner: to start the next session, copy the prompt in §5 verbatim into a fresh AI chat.

## 1. Where the project stands

- **Phase 0 CLOSED** (SESSION-001/002; ADR-0001 Am.2, ADR-0006 Am.1+2,
  ADR-0004 Am.2+Am.3 — Option B stamped: the queue rides the Worker HTTP
  API only; gotd/td dormant in the Phase-8 register).
- **Phase 1 CLOSED** (SESSION-003–005): the eight foundation packages
  ported with generated-golden parity.
- **Phase 2 CLOSED** (SESSION-006–010): T-20…T-25.
- **Phase 3 CLOSED** (SESSION-011–014): T-30…T-34.
- **Phase 4 CLOSED** (SESSION-015–026): T-40…T-50 + the Phase-4 closer.
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity release
  `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip asset
  sha-verified + round-trip verified); the differential corpus full run
  (all generators × two passes GREEN).
- **Phase 6 — THE TASK LIST IS NOW EMPTY (SESSION-033–041)**: T-70
  COMPLETE (the entry gate), T-71 (`internal/app`), T-72 (the frontend
  foundation), T-73 (Settings A), T-74 (Settings B + the CLI-parity
  shell wiring), T-75 (the Sync view), T-76 (the Scan flow), T-77 (the
  main window chrome — the hero machine, the retroactive FIELD-TAG fix,
  the sentinel pass with the two reality fixes, the WAVs), and now
  **T-78 COMPLETE (SESSION-041): THE SIDE-BY-SIDE VERIFICATION VS THE
  PYTHON APP** — the committed machine `testdiff/side_by_side.py` +
  `testdiff/goside` proved the compatibility laws LIVE, 21/21 GREEN
  (evidence: `docs/poc-artifacts/t-78-side-by-side.md`): the app.lock
  mutual exclusion ACROSS BOTH BINARIES (the Python half drives THE
  REFERENCE'S OWN LOCK CODE extracted verbatim by pinned line range —
  the PyQt6 wall evidence-pinned: headless.py:34 → icons.py:45 →
  ModuleNotFoundError under the stdlib-only venv, and --cli/--headless
  both exit BEFORE the lock; the GUI entry is the reference's only
  lock-taking binary), both cross-refusals with the exact sentences,
  both stale healings, the bare-PID bytes law, **and ONE REAL
  DIVERGENCE FOUND AND FIXED** (the port's POSIX liveness read EPERM —
  a live not-ours process — as ALIVE while the reference's blanket
  except reads it DEAD; `internal/app/process_posix.go`'s killVerdict
  now mirrors the except exactly, pinned hermetically, both binaries
  take the PID-1 lock over identically live) + the same-data config
  laws (both apps write BYTE-IDENTICAL config.json through their real
  doors; the cross re-save stability — all four sha256 equal; the fill
  local-wins + fresh; the mirror never-loses + 0600). **The acceptance
  GREEN: lock test across both binaries ✓ — 2253 main-battery tests
  race-clean over 36 packages (2252 + 1) + 56 webui over 4.**
- **THE PHASE-EXIT FORK IS OPEN (the owner's call — §2)**: Phase 6's
  task list is empty; the remaining GUI-parity release demonstrations
  (AC-8's every-control coverage, AC-9's formal probes, AC-15/16/18's
  owner-machine numbers, AC-21's clean-VM leg) are assigned by the
  standing reading (T-70's gate doc + ADR-0001 Am.2) to Phase 7's
  clean-VM checklist — so the fork is (A) declare Phase 6 CLOSED →
  open Phase 7 (T-80) vs (B) add a T-79 (the More menu's
  honestly-disabled tail as named tasks) first. ONE decision resolves
  the phase-exit question AND the bookkeeping question.
- **Push state: CLEAN** — SESSION-041 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project **only when symlink(2) works — it is
  currently BLOCKED, use the real path**), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator
  — same caveat; must match testdiff/PINNED-GITCURATOR-REV), toolchain
  go1.27.2 → /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars, verified against go.dev's published JSON — BEWARE: the
  chat-prompt copies of this string keep losing a character, the same
  transcription corruption TOOLCHAIN.md Am.2 pinned; the committed pin
  is the authority), venv /home/z/venvs/gitcurator (Python 3.12,
  stdlib-only), the wails CLI v2.16.0 (`GOBIN=/home/z/.local/go/bin go
  install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time
  only), and the rootless webkit tree (~/.local/lib-wails — **ABSENT
  after the 36th wipe: SESSION-041 deliberately did NOT rebuild it (no
  wails build was needed for T-78's data-side legs); any GUI-touching
  session rebuilds it per webui/README.md + docs/poc-artifacts/
  t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4 — expect Am.4's
  hardlink fallback: symlink(2) was blocked at boot again**), the MAIN
  go.mod exactly modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client
  v1.16.0 + golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the
  go directive at 1.24.1 (never bump; ADR-0002 Am.2). The NESTED webui
  module rides go 1.25.0 (wails v2.16.0's floor) + the wails require +
  the transitive requires the direct internal/app import pulls
  (indirect via the parent replace — ADR-0002 Am.5; the MAIN go.mod
  stays byte-unchanged). **The downloads must run in FOREGROUND
  CHUNKS** — the sandbox kills backgrounded children when the call
  returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE EXTRACT-AND-EXEC SCOPE LAW** (SESSION-041): when the
    reference's own code sits behind an import wall (PyQt6) and no
    importable entry takes the path (probed: --cli/--headless exit
    before the lock), drive the reference's own BYTES by pinned
    line-range extraction — structural pins + the rev guard fail
    loudly on drift, the sha256 provenance recorded; never
    hand-transcribe, never stub the wall, never improvise past it
    silently.
  - **THE CROSS-BINARY DISAGREEMENT LAW** (SESSION-041): "same
    semantics" means the two binaries agree on EVERY input to the
    shared protocol — including the edge branches nobody hits on a
    healthy box (EPERM: Python's blanket except said dead, Go said
    alive — the mutual exclusion could break exactly when a PID was
    recycled to a stranger). Audit the EXCEPT LADDERS, not just the
    happy paths.
  - **THE SENTINEL-SEAT RACE LAW** (SESSION-040); **THE RETROACTIVE
    FIELD-TAG AUDIT** (SESSION-040); **THE NIL-OPTIONAL DEREF**
    (SESSION-040); **THE ONE-OWNER-PER-SURFACE law** (SESSION-040);
    **THE FIELD-TAG LAW** (SESSION-038); **THE DISPLAY-PIPELINE
    CORRUPTION LESSON** (SESSION-038 — paid twice more in 040 + once
    in 041: the MultiEdit tab corruption hit AGAIN writing Go source;
    `gofmt -w` restored the tabs since HEAD was gofmt-clean, and the
    Write tool kept tabs fine — settle every suspect whitespace with
    byte-level counts before "fixing"); **THE SYMLINK-BLOCK FALLBACK**
    (SESSION-037, TOOLCHAIN.md Am.4); **THE RENDER-CONTEXT LAW**
    (SESSION-037); **THE VLM RATE-LIMIT WINDOW** (SESSION-039); **THE
    PINNED-VERDICT-ROW LESSON** (SESSION-039); **THE ONE-SLICE-
    PROPERLY RECORD** (036–041); **THE LIVE-CONFIG LAW**; **THE
    GROUP-POINTER LAW**; **THE DISSOLUTION-PROBE RECORD** (ADR-0002
    Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP**; **THE
    SUITE-REVERIFICATION DIVIDEND** (paid three times now — 041's
    re-read caught the EPERM divergence in the lock twin); **THE
    WAILS-BUILD MODE FLIPS** (the wailsjs committed baseline IS 755; a
    stray `go build` in webui/ leaves a 22 MB `webui` binary — AND a
    stray `go build ./pkg/` at the repo root leaves the package's
    binary there: caught + removed in 041; build with `-o` into
    gitignored runs/bin/); **THE HERMETIC-SEAM LAW**; **THE
    RESUMED-CONTEXT RE-VERIFICATION LAW**; **THE MODE-FLIP
    DEGRADATION LAW**; **THE MODULE-BOUNDARY PROBE LAW**; **THE
    FRONTEND-MODULE LESSON**; **THE TIMER-CLAMP INSTRUMENT CAVEAT**;
    **THE TOOLCHAIN-ROUND-2/3 PROOF**; THE T-65 ZIP-COMPOSITION
    RESOLUTION; THE CORPUS-REGENERATION LAW; THE
    DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH laws;
    THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD;
    THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the fake-clock
    arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit.
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd — the sandbox's shell resets cwd AND PATH between calls;
    export PATH=/home/z/.local/go/bin:$PATH in the same command; the
    NESTED webui module needs its own cd; never commit outside the
    work repo; never checkout blindly); **THE RENDERER-EATS-ESCAPE-PAIRS
    lesson** (commit messages via -F FILE); **THE WRITE-TOOL
    REDACTION lesson** (byte-level secret-scan before every commit);
    **THE BACKGROUND-PROCESS lesson** (long suites + downloads in
    FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the
    moment a push is due; never store it).

## 2. Next session: SESSION-042 — the owner's phase-exit call FIRST, then Path A (T-80: Phase 7 opens — the deterministic zip builder) or Path B (T-79: the GUI doors addendum)

- **The fork is the FIRST item** — ask the owner in chat BEFORE writing
  any code (the call arrives with the pasted prompt or the owner's
  first reply):
  - **Path A — Phase 6 declared CLOSED; Phase 7 opens.** SESSION-042 =
    **T-80 (M): the deterministic zip builder + version stamping +
    manifest/icon** per doc 13's Phase-7 row (doc 10 §2/§5; acceptance:
    rebuild byte-identical — sha256 twice). The prior art: T-65's
    minimal mkzip slice (build/mkzip + the v0.92.0-go.1 release
    ritual) — T-80 is the FULL builder per doc 10's laws (the
    two-exe composition, the double-build reproducibility ritual, the
    release automation). Read doc 10 FIRST.
  - **Path B — a tasks addendum first.** SESSION-042 = **T-79: the GUI
    doors addendum** — the More menu's honestly-disabled tail as named
    tasks (the per-test dialogs: proxy/ollama/cloud/llamacpp/
    telegram-and-github; the linking tools: recall hooks, link
    suggestions, validate/verify/recategorize; the bot-queue tool
    dialogs: verify-all/export-all/retry-failed; the Chrome doors:
    hand-deliver, hand-rows scrape, tab-retry, attach; the
    decommission picker; the export zip) — write the addendum rows on
    doc 13 (the owner's list above), then implement the first slice.
- **If the owner's word is absent**, do NOT improvise the fork: the
  safe on-the-books work is the deferred test debt
  (test_noteredo's un-ported selections; the note_state surgery-path
  CRLF round-trip differential) while the call is requested.
- **Verify the push state first** (§1): the tip at or beyond
  SESSION-041's close.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-041.md` — the most recent report (T-78
   COMPLETE; the EPERM divergence + fix; the extract-and-exec scope;
   the phase-exit fork).
2. The fork's own reading (per §2): Path A →
   `docs/port-blueprint/10-build-release-packaging.md` (§2 the zip
   laws, §5 the builder, §8 the release workflow) + build/mkzip's
   existing laws + the T-65 row on doc 13. Path B → the More-menu
   disabled-tail list (SESSION-040's report §Blockers + doc 13's
   amendment area) + `spec-gui-ui.md`'s landing note.
3. `docs/port-blueprint/13-implementation-task-list.md` — Phase 7's
   book (T-80..T-83) or the T-79 addendum's home.
4. `docs/port-blueprint/15-acceptance-criteria.md` — the AC rows the
   chosen path serves (AC-21/22/23 for T-80; AC-8 for T-79).
5. As needed: `docs/port-blueprint/decisions/0008-config-data-compat.md`
   (Am.1 — the verification record; the reference for any
   same-data question that resurfaces).

## 4. Known constraints / owner actions for SESSION-042

- **THE PHASE-EXIT CALL (A vs B) — the owner's, due at session start**
  (§2). Everything else waits on it.
- **The webkit tree is ABSENT** (deliberately not rebuilt in 041) — a
  GUI-touching session (Path B's doors, or any wails build) rebuilds it
  per the committed recipe FIRST (expect Am.4's hardlink fallback:
  symlink(2) blocked at boot).
- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — thirty-second
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755 — the committed baseline IS 755 for the
  wailsjs files; only a CONTENT diff should remain) and never commit a
  stray binary (webui/webui, or a repo-root package binary — build
  with `-o` into gitignored paths).
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the owner;
  neither gates the fork.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7); the Backup page's Restore button
  (honestly disabled — rides the packaging phase); the More menu's
  disabled doors (the fork's Path B IS this item's home).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB any persisted
  extraheader — the GIT_CONFIG_COUNT env-var form avoids it entirely).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-042)

```
You are starting SESSION-042 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE-EXIT CALL COMES FIRST (the owner's — do not improvise it):
   Phase 6's task list is EMPTY (T-70..T-78 all COMPLETE; T-78 = the
   side-by-side verification vs the Python app landed GREEN in SESSION-041:
   the lock mutual exclusion across both binaries — the reference's own lock
   code driven live — plus the same-data config laws, 21/21, evidence at
   docs/poc-artifacts/t-78-side-by-side.md; ONE real divergence found + fixed:
   the POSIX EPERM liveness branch). The fork:
   (A) declare Phase 6 CLOSED and open Phase 7 — SESSION-042 = T-80 (M): the
   deterministic zip builder + version stamping + manifest/icon per doc 13's
   Phase-7 row (doc 10 §2/§5; acceptance: rebuild byte-identical, sha256
   twice; the prior art = T-65's minimal build/mkzip + the v0.92.0-go.1
   release ritual), OR
   (B) add the tasks addendum first — SESSION-042 = T-79: the GUI doors
   addendum (the More menu's honestly-disabled tail as named tasks: the
   per-test dialogs proxy/ollama/cloud/llamacpp/telegram-and-github, the
   linking tools recall-hooks/link-suggestions/validate/verify/recategorize,
   the bot-queue tool dialogs verify-all/export-all/retry-failed, the Chrome
   doors hand-deliver/hand-rows-scrape/tab-retry/attach, the decommission
   picker, the export zip) — write the addendum rows on doc 13, then
   implement the first slice.
   Ask me for the call in your FIRST reply and wait for it if it is not
   already in this prompt; if I am absent, do NOT improvise the fork — work
   the deferred test debt instead (test_noteredo's un-ported selections; the
   note_state surgery-path CRLF round-trip differential).

1. The environment is REBUILDABLE, not persistent (thirty-six wipe events
   across SESSION-005..041 — SESSION-041 absorbed the thirty-sixth full wipe
   BEFORE work: both repos, the toolchain, the venv, the module cache, AND
   the rootless webkit tree all gone; everything rebuilt from the committed
   recipes EXCEPT the webkit tree, deliberately deferred — no wails build
   was needed for T-78's data-side legs, so the tree is ABSENT until the
   next GUI-touching session rebuilds it per webui/README.md +
   docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4, Am.4's
   hardlink fallback expected: symlink(2) is blocked at BOOT, the 26 dev
   .so links land as hardlinks to the RESOLVED system targets — readlink -f
   the LINK'S TARGET against the system tree; the wails build + the gate leg
   the proof); NOTE the standing lessons: downloads and long suites must run
   in FOREGROUND CHUNKS (the sandbox kills backgrounded children when the
   call returns); this sandbox's tool-output rendering occasionally drops
   characters — settle every suspect reading with byte-level counts/hex
   before "fixing" anything (the MultiEdit tab corruption struck AGAIN in
   041 writing Go source — gofmt -w restored the tabs because HEAD was
   gofmt-clean; patch with python reading the real bytes); the VLM CLI
   rate-limits (429) under rapid calls — pause between the screenshot
   verifications; the continuation recipe is PROVEN at all levels: git
   status + git log --oneline -5 + file timestamps tell what ACTUALLY
   survived; a preserved commit or in-flight draft is continued and AUDITED —
   never blindly trusted, never restarted from zero; a resumed context
   re-verifies from scratch (the full battery + the hygiene scans) before
   committing; a wipe event may flip tracked files' MODES 644->755 while
   preserving content — diagnose via 'git diff --raw' and restore via chmod
   off the parsed list, NEVER git checkout; and the standing addendum: after
   every `wails build` the wailsjs bindings re-flip to 755 (the repo's
   committed baseline for the wailsjs files IS 755 — only a CONTENT diff
   should remain after a build), a stray `go build .` in webui/ leaves a
   22 MB binary named `webui`, and a stray `go build ./pkg/` at the repo
   root leaves the package's binary there — never commit either; build with
   -o into gitignored paths). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the /home/z/second-brain-project
   symlink is DOWN while symlink(2) is blocked — use the real path),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator — same caveat; must match testdiff/PINNED-GITCURATOR-REV),
   toolchain go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON; BEWARE the transcription
   corruption — the committed pin in build/TOOLCHAIN.md Am.2 is the authority),
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, the wails CLI v2.16.0 (GOBIN=/home/z/.local/go/bin go
   install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only).
   The MAIN go.mod: modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at
   v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and
   the go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2;
   do not bump any of these). The NESTED webui module rides go 1.25.0
   (wails v2.16.0's floor) + the wails require + the transitive requires the
   direct internal/app import pulls (indirect via the parent replace —
   ADR-0002 Am.5); the MAIN go.mod stays byte-unchanged. Then check `git
   status` and `git log --oneline -3`: SESSION-041 landed T-78 COMPLETE in
   one slice (the side-by-side machine: testdiff/side_by_side.py +
   testdiff/goside — 21/21 GREEN; the EPERM liveness fix in
   internal/app/process_posix.go pinned by TestKillVerdictMirrorsTheBlanketExcept;
   the raw-bytes law tightened in the lock tests; the docs: doc 13's row +
   doc 12's entry + ADR-0008 Am.1 + the archived evidence) and PUSHED
   everything at the close — 2253 main-battery tests race-clean over 36
   packages + 56 webui tests over 4. Verify git status shows up to date
   with origin/main (the tip at or beyond SESSION-041's close). Use the PAT
   only ephemerally: git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety (paid three times now: 040's nil-deref + race,
   041's EPERM divergence).

2. Read, in order: the repo README.md (Phase 6's task list EMPTY — T-78
   COMPLETE) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-041's close:
   T-78 COMPLETE, the newest laws — THE EXTRACT-AND-EXEC SCOPE LAW, THE
   CROSS-BINARY DISAGREEMENT LAW) → docs/sessions/SESSION-041.md (the most
   recent report) → then the chosen path's own reading per the state card
   §3 (Path A: docs/port-blueprint/10-build-release-packaging.md §2/§5/§8 +
   build/mkzip + doc 13's T-65/T-80 rows; Path B: the disabled-tail list +
   spec-gui-ui.md) → doc 15's relevant AC rows.

3. THIS SESSION IS THE OWNER'S PHASE-EXIT CALL FIRST (item 0 above), then
   EITHER T-80 (Path A: the deterministic zip builder + version stamping +
   manifest/icon — doc 10 §2/§5; acceptance: rebuild byte-identical, sha256
   twice) OR T-79 (Path B: the GUI doors addendum — the rows first, then
   the first slice). If capacity runs short after the call: one slice done
   properly with proof beats two half-done; close cleanly at a natural
   boundary.

4. Standing laws: reality wins over specs (fix the spec in the same commit);
   tests are never weakened; no secrets or owner data in git (byte-level
   secret-scan before every commit; never write token-shaped strings even as
   fakes); no new dependencies without an ADR (the authorized runtime set is
   exactly modernc.org/sqlite v1.36.1, tls-client v1.16.0, x/net v0.48.0,
   x/text v0.32.0 — do not bump any; the MAIN go.mod stays byte-unchanged;
   the wails require + the indirect transitive requires stay in the NESTED
   webui module per ADR-0002 Am.5); surface blockers with evidence
   (hash-level, not display) instead of improvising past them; probe before
   claiming any module-boundary rule (ADR-0002 Am.4); fake every
   network-touching seam in tests at construction (THE HERMETIC-SEAM law);
   when the reference's own code sits behind an import wall, drive the
   reference's own BYTES by pinned line-range extraction — structural pins
   + the rev guard, never hand-transcription, never a stubbed wall (THE
   EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the two binaries
   agree on EVERY input to the shared protocol, edge branches included —
   audit the except ladders (THE CROSS-BINARY DISAGREEMENT LAW); the payload
   structs' JSON tags mirror the Python dict keys and the frontend reads the
   tagged names (THE FIELD-TAG LAW — extend TestCrossedPayloadFieldTags with
   every new payload); a transition seat read+written by concurrent passes
   holds it under a mutex as ONE unit (THE SENTINEL-SEAT RACE LAW — run the
   battery ≥3× before trusting a new concurrent seam); every *T field read
   is settled with a nil check (THE NIL-OPTIONAL DEREF); one owner per
   frontend surface (duplicate EventsOn listeners double-paint); a modal's
   action buttons live OUTSIDE the scrollable body; after every wails build,
   check `git diff --raw` for mode flips (the wailsjs committed baseline IS
   755) and never commit a stray binary; every git/go command carries its
   own cd AND its own PATH=/home/z/.local/go/bin export (the sandbox shell
   resets both between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (the
   chosen path's own acceptance), commit + push everything (ask me for the
   GitHub PAT the moment a push is due — never store it anywhere; never end
   the session push-pending), write docs/sessions/SESSION-042.md, rewrite
   docs/NEXT-SESSION.md with the next session's prompt, and end your reply
   with a plain-language report plus the prompt I should copy to start
   SESSION-043.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

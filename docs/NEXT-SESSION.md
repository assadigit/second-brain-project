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
- **Phase 4 CLOSED (SESSION-015–026)**: T-40…T-50 + the Phase-4 closer.
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity release
  `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip asset
  sha-verified + round-trip verified); the differential corpus full run
  (all generators × two passes GREEN).
- **Phase 6 OPEN — THE OWNER'S PATH-B STAMP (SESSION-042)**: "the slow,
  more reliable and more rigorous method" — the More menu's
  honestly-disabled tail became the NAMED rows on doc 13. **T-79a
  COMPLETE (SESSION-042)** — the five per-test doors. **T-79b COMPLETE
  (SESSION-043)** — the five linking-tools doors. **T-79c COMPLETE
  (SESSION-044)** — the three bot-queue tool doors. **T-79d COMPLETE
  (SESSION-045)** — the four Chrome doors. **T-79e COMPLETE
  (SESSION-046)** — the decommission picker (the LAST tail row — the
  More menu's tail is now EMPTY; the one-ledger law LIVE: the picker
  writes the SAME Status cells a hand-edit writes, then enforces the
  burials immediately; ♻️ revived enforced by the same pass). **The
  one remaining row: T-79f — THE PHASE-6 CLOSURE ROW (SESSION-046's
  scoping decision, named on doc 13)**: the runner's automatic
  triggers — the end-of-run Chrome offer (`_maybe_offer_chrome_tab_
  retry`) + hero.py's caught-up hand scrape + `_maybe_redo_hand_notes`
  — the batch-finish/hero-machine surfaces riding the ALREADY-LANDED
  `startChromeTabRetry` runner (a follow-up wiring with its own pinned
  tests, not a fold-in). **Phase 6 CLOSES when T-79f lands; then Phase
  7 (T-80: the deterministic zip builder …) opens.** The export zip
  rides T-80's packaging composition per the standing assignment.
- **Push state: CLEAN** — SESSION-046 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (the /home/z/second-brain-project
  symlink is DOWN while symlink(2) is blocked — use the real path),
  read-only reference /home/z/my-project/GitCurator @ `98f4fc7`
  (must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars; NOTE: go.dev's ?mode=json has ROTATED to 1.27.3/1.26.10 and no
  longer lists 1.27.2 — verify the locally-computed hash against the
  COMMITTED pin in build/TOOLCHAIN.md Am.2 [itself verified against
  go.dev's published listing at SESSION-032]; the download stays HTTPS
  from go.dev), venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only —
  the plain `python3 -m venv` form when ensurepip is present, else the
  `--without-pip --copies` fallback), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **rebuilt at
  SESSION-046 [the FOURTEENTH rebuild]; if a wipe takes it, rebuild it
  FIRST when the session is GUI-touching** per webui/README.md +
  docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4
  [144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links;
  expect the Am.4 hardlink fallback while symlink(2) is blocked]),
  the MAIN go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (never bump;
  ADR-0002 Am.2). The NESTED webui module rides go 1.25.0 (wails
  v2.16.0's floor) + the wails require + the transitive requires the
  direct internal/app import pulls (indirect via the parent replace —
  ADR-0002 Am.5; the MAIN go.mod stays byte-unchanged). **The downloads
  must run in FOREGROUND CHUNKS** — the sandbox kills backgrounded
  children when the call returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE 42ND-WIPE RECIPE (SESSION-046)**: same as the 41st —
    clone both with the ephemeral GIT_CONFIG_COUNT PAT form, verify
    the rev pin, rebuild per the committed recipes, and re-verify
    the FULL battery (plain + race) before touching code; the
    re-verification pays again: 046's caught the 71-vs-73 webui
    arithmetic slip in SESSION-045's close docs (settled at the
    byte level — six shell tests landed in fd053ba; the five
    records corrected).
  - **THE REFERENCE'S-OWN-ARITHMETIC LAW (SESSION-046)**: when a
    new test's expectation disagrees with the engine over a
    reference behavior, re-derive the reference's own computation
    BEFORE "fixing" the engine — the deferral case's box says
    "backlog is clear" because the marking already filters the
    backlog scan (the table is the truth); my first expectation was
    wrong and the engine was right.
  - **THE VLM MISREAD LEDGER (FOUR incidents — 042/043/044/045;
    046 ran clean with both passes agreeing)**: ALWAYS settle a
    suspect verdict with the byte-level source check + the pinned
    tests; never fix code off a suspect VLM verdict.
  - **THE BINDING-IMPORT LAW (SESSION-044)**: every new wails-bound
    door must join the owning module's binding IMPORT statement —
    pinned by the import pins in the heroflow live tests.
  - **THE RACE-PASS DIVIDEND (SESSION-045)**; **THE
    EXTRACT-AND-EXEC SCOPE LAW** (SESSION-041); **THE
    CROSS-BINARY DISAGREEMENT LAW** (paid at 044 AND 045); **THE
    SENTINEL-SEAT RACE LAW** (SESSION-040); **THE RETROACTIVE
    FIELD-TAG AUDIT** (SESSION-040); **THE NIL-OPTIONAL DEREF**
    (SESSION-040); **THE ONE-OWNER-PER-SURFACE law** (SESSION-040);
    **THE FIELD-TAG LAW** (SESSION-038); **THE DISPLAY-PIPELINE
    CORRUPTION LESSON** (SESSION-038 — the Edit tool strikes
    SPACE-vs-TAB on Go source: gofmt -l after every edit batch,
    gofmt -w restores since HEAD is gofmt-clean; patch with python
    reading the real bytes — it struck twice more this session);
    **THE SYMLINK-BLOCK FALLBACK** (SESSION-037/044/045/046,
    TOOLCHAIN.md Am.4); **THE RENDER-CONTEXT LAW** (SESSION-037);
    **THE VLM RATE-LIMIT WINDOW** (SESSION-039); **THE
    PINNED-VERDICT-ROW LESSON** (SESSION-039); **THE
    ONE-SLICE-PROPERLY RECORD** (036–046); **THE LIVE-CONFIG
    LAW**; **THE GROUP-POINTER LAW**; **THE DISSOLUTION-PROBE
    RECORD** (ADR-0002 Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP**;
    **THE SUITE-REVERIFICATION DIVIDEND**; **THE WAILS-BUILD MODE
    FLIPS** (the wailsjs committed baseline IS 755; only a CONTENT
    diff should remain after a build; never commit a stray
    binary); **THE HERMETIC-SEAM LAW**; **THE
    RESUMED-CONTEXT RE-VERIFICATION LAW**; **THE MODE-FLIP
    DEGRADATION LAW**; **THE MODULE-BOUNDARY PROBE LAW**; **THE
    FRONTEND-MODULE LESSON**; THE T-65 ZIP-COMPOSITION RESOLUTION;
    THE CORPUS-REGENERATION LAW; THE DISPATCH-COMPLETION LAW; THE
    STATE-CHAIN + MUTATION-TEETH laws; THE CONTEXT-DEATH
    CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD; THE
    FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the fake-clock
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

## 2. Next session: SESSION-047 — T-79f: the Phase-6 closure row (the runner's automatic triggers)

- **SESSION-047 = T-79f (M)**: the three batch-finish/hero-machine
  triggers riding the ALREADY-LANDED `startChromeTabRetry` runner
  (T-79d): (a) the end-of-run Chrome offer
  (`_maybe_offer_chrome_tab_retry` — the batch-scorecard fanfare's
  offer modal over the session's offered set), (b) hero.py's
  caught-up hand scrape (the caught-up story's hand pass), (c)
  `_maybe_redo_hand_notes` (the redo pass). A follow-up wiring with
  its own pinned tests (the offered-set + the held-rows seats + the
  hero machine's caught-up story), not new engines. **Phase 6
  CLOSES when this row lands** — then the phase-exit gate
  evaluation per SESSION-PROTOCOL (honest, evidence-first) and
  Phase 7 (T-80) opens.
- Follow T-79a/b/c/d/e's landed shape (the engines over hermetic
  seams + the shell doors + the heroflow dispatch + THE
  BINDING-IMPORT LAW + the line grammar pinned hermetically + the
  evidence probe's click leg + THE FIELD-TAG law on every genuinely
  new payload).
- If capacity runs short: one slice done properly with proof beats
  two half-done; close cleanly at a natural boundary.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-046.md` — the most recent report (T-79e's
   landing; the 42nd-wipe repair; the arithmetic-slip repair; the
   scoping decision).
2. The reference at 98f4fc7: the fanfare chain in
   `gui/main_window/processing_control.py` (processing_finished's
   tail — `_maybe_deliver_hand_queue`, `_maybe_redo_hand_notes`,
   `_maybe_offer_chrome_tab_retry`) + hero.py's caught-up story.
3. The Go surfaces: `webui/heroui.go`'s `startChromeTabRetry` runner
   + the offered-set + held-rows seats (chromedoors.go's
   `chromeTabOffered`, the shell's `handRowsRows`) +
   `internal/app/herostate.go` (the hero machine's caught-up law) +
   the batch-finish handler (`onBatchFinished`) + the scorecard
   surfaces.
4. `docs/port-blueprint/13-implementation-task-list.md` — the
   T-79f row (NEW — SESSION-046's naming).
5. `docs/poc-artifacts/t-79e-decommission-doors.md` — the newest
   evidence pattern + the divergences' recording style.
6. As needed: `docs/poc-artifacts/t-79d-chrome-doors.md` (the
   runner's own record).

## 4. Known constraints / owner actions for SESSION-047

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree exists at SESSION-046's close** (the 14th
  rebuild) — if a wipe takes it, rebuild FIRST (the session is
  GUI-touching) per §1's recipe.
- **The T-79f naming is SESSION-046's scoping decision** — the owner
  may override it (fold into a different shape or rename); surface
  any override in the report + the task list in the same commit.
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — thirty-seventh
  session carrying it): the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755 — the committed baseline IS 755 for the
  wailsjs files; only a CONTENT diff should remain) and never commit
  a stray binary (webui/webui, or a repo-root package binary — build
  with `-o` into gitignored paths).
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the owner;
  neither gates the closure row.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7); the Backup page's Restore button
  (honestly disabled — rides T-80's packaging composition with the
  export zip).
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

## 5. THE PROMPT TO COPY (SESSION-047)

```
You are starting SESSION-047 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 6 is OPEN under the owner's Path-B stamp
   (SESSION-042: "the slow, more reliable and more rigorous method").
   T-79 + T-79a (the five per-test doors) + T-79b (the five
   linking-tools doors) + T-79c (the three bot-queue tool doors) +
   T-79d (the four Chrome doors) + T-79e (the decommission picker —
   the LAST tail row; the More menu's tail is EMPTY) are COMPLETE.
   SESSION-047 = T-79f (M): THE PHASE-6 CLOSURE ROW (SESSION-046's
   scoping decision, named on doc 13) — the runner's automatic
   triggers: (a) the end-of-run Chrome offer
   (_maybe_offer_chrome_tab_retry — the batch-scorecard fanfare's
   offer modal over the session's offered set), (b) hero.py's
   caught-up hand scrape, (c) _maybe_redo_hand_notes — the
   batch-finish/hero-machine surfaces riding the ALREADY-LANDED
   startChromeTabRetry runner (T-79d): a follow-up wiring with its
   own pinned tests (the offered-set + the held-rows seats + the
   hero machine's caught-up story), not new engines. PHASE 6 CLOSES
   WHEN THIS ROW LANDS — then the honest phase-exit gate evaluation
   per SESSION-PROTOCOL and Phase 7 (T-80: the deterministic zip
   builder …) opens. The owner may override the T-79f naming
   (surface any override in the report + the task list in the same
   commit). Follow T-79a/b/c/d/e's landed shape (the engines over
   hermetic seams + the shell doors + the heroflow dispatch + THE
   BINDING-IMPORT LAW [join the import statement, pinned by test] +
   the line grammar pinned hermetically + the evidence probe's click
   leg + THE FIELD-TAG law on every genuinely new payload). If
   capacity runs short: one slice done properly with proof beats two
   half-done.

1. The environment is REBUILDABLE, not persistent (forty-two wipe
   events across SESSION-005..046 — SESSION-046 opened on the
   forty-second: the WHOLE environment was gone, both repos included,
   all rebuilt from the committed recipes; the pre-work battery
   re-verification then caught ONE ARITHMETIC SLIP in SESSION-045's
   close docs — the recorded "71 webui (67+4)" is really 73 (67+6,
   six shell tests landed in fd053ba); the five records were corrected
   in 046's close commits — settle every discrepancy with repeated
   runs + the byte-level root cause BEFORE touching anything). NOTE
   the standing lessons: downloads and long suites must run in
   FOREGROUND CHUNKS (the sandbox kills backgrounded children when the
   call returns); this sandbox's tool-output rendering occasionally
   drops characters — settle every suspect reading with byte-level
   counts/hex before "fixing" anything (the Edit tool's space-vs-tab
   corruption struck in 042 AND 044 AND 046 writing Go source — gofmt
   -l after every edit batch, gofmt -w restores because HEAD is
   gofmt-clean; patch with python reading the real bytes); the VLM
   CLI rate-limits (429) under rapid calls — pause between the
   screenshot verifications — AND the VLM misreads enabled-vs-disabled
   states (FOUR incidents: 042/043/044/045 — each settled with the
   byte-level source check + the pinned tests; never "fix" code off a
   suspect VLM verdict; 046 ran clean with both passes agreeing);
   'ln -s' refuses AND kills the whole compound bash call at this
   boot — never put it in a compound command (the venv needs
   --without-pip --copies on ensurepip-less boots; the webkit tree's
   26 dev links need Am.4 hardlinks); the BINDING-IMPORT lesson
   (044): a wails door missing from the owning module's binding
   IMPORT statement throws a silent ReferenceError — the desk never
   speaks; join the import list and pin it by test; the
   REFERENCE'S-OWN-ARITHMETIC law (046): when a new test's
   expectation disagrees with the engine over a reference behavior,
   re-derive the reference's own computation BEFORE "fixing" the
   engine (the deferral case's box says "backlog is clear" because
   the marking already filters the backlog scan — the table is the
   truth). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the
   /home/z/second-brain-project symlink is DOWN while symlink(2) is
   blocked — use the real path), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (must match
   testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 ->
   /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64
   hex chars; go.dev's ?mode=json has ROTATED to 1.27.3 and no longer
   lists 1.27.2: verify the locally-computed hash against the COMMITTED
   pin in build/TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the
   transcription corruption), venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only; the plain 'python3 -m venv' form when ensurepip is
   present, else --without-pip --copies), the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — rebuilt at 046 [the
   FOURTEENTH rebuild]; if a wipe took it, rebuild FIRST per
   webui/README.md + docs/poc-artifacts/t-70-gate.md +
   build/TOOLCHAIN.md Am.3 + Am.4: 144 debs → dpkg -x → the 42 .pc
   prefix rewrites → the 26 dev links, hardlink fallback expected). The
   MAIN go.mod: modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at
   v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct,
   and the go directive at 1.24.1 (the tls-client-forced floor —
   ADR-0002 Am.2; do not bump any of these). The NESTED webui module
   rides go 1.25.0 (wails v2.16.0's floor) + the wails require + the
   transitive requires the direct internal/app import pulls (indirect
   via the parent replace — ADR-0002 Am.5); the MAIN go.mod stays
   byte-unchanged. Then check 'git status' and 'git log --oneline -3':
   SESSION-046 landed T-79e in its close commits and PUSHED everything —
   2348 main-battery tests race-clean over 36 packages (2337 + 11) + 77
   webui tests over 4 (73 + 4). Verify git status shows up to date with
   origin/main. Use the PAT only ephemerally: the GIT_CONFIG_COUNT
   env-var form avoids persistence entirely — and SCRUB IT if you ever
   CLONE with it. Re-verify the suite before work if anything looks
   off — the law pays in bugs, not just safety.

2. Read, in order: the repo README.md (SESSION-046's status block —
   T-79e) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-046's
   close) → docs/sessions/SESSION-046.md (the most recent report —
   the 42nd-wipe repair + the arithmetic-slip repair + the scoping
   decision) → the reference's fanfare chain (processing_control.py's
   processing_finished tail — _maybe_deliver_hand_queue,
   _maybe_redo_hand_notes, _maybe_offer_chrome_tab_retry) + hero.py's
   caught-up story → the Go surfaces (webui/heroui.go's
   startChromeTabRetry runner + the offered-set + held-rows seats +
   internal/app/herostate.go's caught-up law + the batch-finish
   handler + the scorecard surfaces) → doc 13's T-79f row → doc 15's
   phase-exit gates.

3. THIS SESSION = T-79f: the Phase-6 closure row — the runner's
   automatic triggers over the landed startChromeTabRetry runner.
   One slice done properly with proof beats two half-done; close
   cleanly at a natural boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes); no new dependencies without an
   ADR (the authorized runtime set is exactly modernc.org/sqlite
   v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; the MAIN go.mod stays byte-unchanged; the wails require +
   the indirect transitive requires stay in the NESTED webui module
   per ADR-0002 Am.5); surface blockers with evidence (hash-level, not
   display) instead of improvising past them; probe before claiming any
   module-boundary rule (ADR-0002 Am.4); fake every network-touching
   seam in tests at construction (THE HERMETIC-SEAM law); when the
   reference's own code sits behind an import wall, drive the
   reference's own BYTES by pinned line-range extraction — structural
   pins + the rev guard, never hand-transcription, never a stubbed wall
   (THE EXTRACT-AND-EXEC SCOPE LAW); "same semantics" means the two
   binaries agree on EVERY input to the shared protocol, edge branches
   included — audit the except ladders (THE CROSS-BINARY DISAGREEMENT
   LAW); the payload structs' JSON tags mirror the Python dict keys and
   the frontend reads the tagged names (THE FIELD-TAG LAW — extend the
   payload field-tags test with every new payload); a transition seat
   read+written by concurrent passes holds it under a mutex as ONE
   unit (THE SENTINEL-SEAT RACE LAW — run the battery ≥3× before
   trusting a new concurrent seam; every live-map mutation holds the
   Settings mutex, the SaveSettings discipline); every
   pointer/optional field read is settled with a nil check (THE
   NIL-OPTIONAL DEREF); one owner per frontend surface (duplicate
   EventsOn listeners double-paint); a modal's action buttons live
   OUTSIDE the scrollable body; every new wails door JOINS the owning
   module's binding IMPORT statement (THE BINDING-IMPORT LAW — pin it
   by test); after every wails build, check 'git diff --raw' for mode
   flips (the wailsjs committed baseline IS 755) and never commit a
   stray binary; gofmt -l after every Go edit batch (the Edit tool's
   space-vs-tab corruption — gofmt -w restores); every git/go command
   carries its own cd AND its own PATH=/home/z/.local/go/bin export
   (the sandbox shell resets both between calls; the NESTED webui
   module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (the chosen path's own acceptance — and the PHASE-6 EXIT question
   if T-79f lands: the phase-exit gate per the protocol's own laws,
   evidence-first), commit + push everything (ask me for the GitHub
   PAT the moment a push is due — never store it anywhere; never end
   the session push-pending), write docs/sessions/SESSION-047.md,
   rewrite docs/NEXT-SESSION.md with the next session's prompt, and
   end your reply with a plain-language report plus the prompt I
   should copy to start SESSION-048.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```

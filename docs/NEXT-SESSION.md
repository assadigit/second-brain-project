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
- **Phase 5 CLOSED (SESSION-027–032)**: T-60…T-65 — the CLI parity
  release `v0.92.0-go.1` (tag at `a39cb3d`, release 416838956, the zip
  asset sha-verified + round-trip verified); the differential corpus
  full run (all generators × two passes GREEN).
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70
  gate through the T-79a…T-79f Path-B rows — **T-79f (the Phase-6
  closure row) COMPLETE at SESSION-047** (the runner's automatic
  triggers: the end-of-run Chrome offer + hero.py's caught-up hand
  scrape + the redo pass, over the landed startChromeTabRetry runner;
  THE FLAG-CLEAR LAW in SyncRunBatch; the batch_summary facts riding
  BatchFinished; the six recorded divergences at
  docs/poc-artifacts/t-79f-fanfare-triggers.md). **The honest exit
  evaluation: all three roadmap gate criteria GREEN** (the golden
  paths' workflow tests; the performance criteria — first frame
  802/281 ms ≤ 1.5 s, the 10k flood 0 drops; AC-11's side-by-side
  lock DEMONSTRATED LIVE at T-78) **+ the Path-B rider satisfied** (the
  More menu's tail is EMPTY; the export-zip row rides T-80's packaging
  composition per the standing assignment). The GUI-parity RELEASE
  criteria (the Windows runtime legs + the clean-VM checklist) remain
  Phase 7's owner items exactly as ADR-0001 Am.2 registered.
- **Phase 7 OPEN — T-80 (the deterministic zip builder + version
  stamping + manifest/icon) is the next row**: the reproducible
  packaging (the zip's byte-identical rebuild, sha256 twice); the
  export zip + the Backup page's Restore button ride the same
  packaging composition (the standing assignment); doc 10 §2/§5 are
  the specs; then T-81 (the release workflow) + T-82 (the clean-VM
  checklist) + T-83 (the docs).
- **Push state: CLEAN** — SESSION-047 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (the /home/z/second-brain-project
  symlink is DOWN while symlink(2) is blocked — use the real path),
  read-only reference /home/z/my-project/GitCurator @ `98f4fc7`
  (must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars; NOTE: go.dev's ?mode=json has ROTATED past 1.27.2: verify
  the locally-computed hash against the COMMITTED pin in
  build/TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the
  transcription corruption), venv /home/z/venvs/gitcurator (Python 3.12,
  stdlib-only — the plain `python3 -m venv` form when ensurepip is
  present, else the `--without-pip --copies` fallback), the wails CLI
  v2.16.0 (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **rebuilt at
  SESSION-047 [the FIFTEENTH rebuild]; if a wipe takes it, rebuild it
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
  - **THE 43RD-WIPE RECIPE (SESSION-047)**: same as the 41st/42nd —
    clone both with the ephemeral GIT_CONFIG_COUNT PAT form, verify
    the rev pin, rebuild per the committed recipes, and re-verify the
    FULL battery (plain + race) before touching code.
  - **THE FLAG-CLEAR LAW (SESSION-047)**: SyncRunBatch clears the
    running flag + releases the lock BEFORE the finished signal
    crosses (the reference's processing_finished opens with
    `_batch_running = False`) — pinned by
    TestSyncRunBatchFlagClearLawAndFanfareFacts; the fanfare chain
    (T-79f) depends on it (its own re-run batch starts inside the
    crossing's handler).
  - **THE REFERENCE'S-OWN-ARITHMETIC LAW (SESSION-046)**; **THE
    42ND-WIPE RECIPE**; **THE VLM MISREAD LEDGER (FOUR incidents:
    042/043/044/045; 046 AND 047 ran clean with both passes
    agreeing)**; **THE BINDING-IMPORT LAW (SESSION-044)**; **THE
    RACE-PASS DIVIDEND (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE
    LAW** (SESSION-041); **THE CROSS-BINARY DISAGREEMENT LAW** (paid
    at 044, 045 AND 047 — the CLI twin's website summary + the
    flag-clear order); **THE SENTINEL-SEAT RACE LAW** (SESSION-040);
    **THE RETROACTIVE FIELD-TAG AUDIT** (SESSION-040); **THE
    NIL-OPTIONAL DEREF** (SESSION-040); **THE ONE-OWNER-PER-SURFACE
    law** (SESSION-040); **THE FIELD-TAG LAW** (SESSION-038);
    **THE DISPLAY-PIPELINE CORRUPTION LESSON** (SESSION-038 — the
    Edit tool strikes SPACE-vs-TAB on Go source: gofmt -l after every
    edit batch, gofmt -w restores since HEAD is gofmt-clean; patch
    with python reading the real bytes; it struck AGAIN at 047);
    **THE SYMLINK-BLOCK FALLBACK** (SESSION-037/044/045/046/047,
    TOOLCHAIN.md Am.4); **THE RENDER-CONTEXT LAW** (SESSION-037);
    **THE VLM RATE-LIMIT WINDOW** (SESSION-039 — pause between the
    passes); **THE PINNED-VERDICT-ROW LESSON** (SESSION-039); **THE
    ONE-SLICE-PROPERLY RECORD** (036–047); **THE LIVE-CONFIG LAW**;
    **THE GROUP-POINTER LAW**; **THE DISSOLUTION-PROBE RECORD**
    (ADR-0002 Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP**; **THE
    SUITE-REVERIFICATION DIVIDEND**; **THE WAILS-BUILD MODE FLIPS**
    (the wailsjs committed baseline IS 755; only a CONTENT diff should
    remain after a build; never commit a stray binary); **THE
    HERMETIC-SEAM LAW**; **THE RESUMED-CONTEXT RE-VERIFICATION LAW**;
    **THE MODE-FLIP DEGRADATION LAW**; **THE MODULE-BOUNDARY PROBE
    LAW**; **THE FRONTEND-MODULE LESSON**; THE T-65 ZIP-COMPOSITION
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

## 2. Next session: SESSION-048 — Phase 7's opening row: T-80 (the deterministic zip builder)

- **SESSION-048 = T-80 (M)**: the deterministic zip builder + version
  stamping + manifest/icon (doc 10 §2/§5): the reproducible packaging
  — the zip's byte-identical rebuild (the acceptance: sha256 equal on
  two builds from the same commit). The export zip + the Backup
  page's Restore button ride the same packaging composition (the
  standing assignment — SESSION-040's record + the honestly-disabled
  More-menu row + the Backup page's disabled Restore note).
- Doc 13's T-81/T-82/T-83 follow (the release workflow, the clean-VM
  checklist, the docs) — one slice done properly with proof beats two
  half-done; close cleanly at a natural boundary when capacity runs
  short.
- NOTE the boundary honestly: the GUI-parity RELEASE criteria (the
  Windows runtime legs + the clean-VM checklist) are the owner's
  verification items per ADR-0001 Am.2 — the sandbox has no Windows
  box (the standing constraint since SESSION-001); the cross-compile
  legs (`wails build -platform windows/amd64`, CGO-free) are the
  sandbox's proof.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-047.md` — the most recent report (T-79f's
   landing; the 43rd-wipe repair; THE PHASE-6 EXIT evaluation).
2. `docs/port-blueprint/10-build-release-packaging.md` §2/§5 (+ §7/§8
   for the T-81/T-82 boundary) — T-80's specs.
3. The reference at 98f4fc7: the deterministic zip's Python twin
   (the release packaging — `mkzip`/the release script + the zip
   manifest laws; find the exact module by `rg -l zip` over the
   reference's release/build scripts).
4. `build/mkzip/` — the Go zip helper already in the tree (the
   T-65-era zip-composition resolution; the CLI parity release's zip
   asset used it — SESSION-032's record).
5. `docs/poc-artifacts/t-79f-fanfare-triggers.md` — the newest
   evidence pattern + the divergences' recording style.
6. `docs/port-blueprint/15-acceptance-criteria.md` §6 (AC-21…AC-24)
   — the packaging gates T-80 feeds.

## 4. Known constraints / owner actions for SESSION-048

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree exists at SESSION-047's close** (the 15th
  rebuild) — if a wipe takes it, rebuild FIRST (the session is
  GUI-adjacent through the cross-compile legs) per §1's recipe.
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The batch-scorecard modal question (the owner's, NEW)**: the
  reference's Batch Complete modal + the stash/merge interplay is
  recorded as unlanded (T-79f's divergence #2 — its consumer was
  never a named row); the owner may name it a Phase-7 row or leave
  the bar+chime fanfare as the release shape.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — thirty-eighth
  session carrying it): the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary.
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the
  owner; neither gates the packaging row.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7); the Backup page's Restore button
  (honestly disabled — rides THIS row's packaging composition).
- **Single-writer law**: the fresh session is the ONLY writer; the
  overseer chat records decisions and audits — it does not implement.
- Python reference: `assadigit/GitCurator` @ `98f4fc7` — read-only;
  re-clone per §1 if the sandbox wiped it (and SCRUB any persisted
  extraheader — the GIT_CONFIG_COUNT env-var form avoids it
  entirely).
- Standing: secrets / owner data never in git (byte-level secret-scan
  before every commit); tests never weakened; reality wins over specs
  (fix the spec in the same commit); surface blockers with evidence
  (hash-level, not display).
- **Push after every task** — never end a session push-pending.

## 5. THE PROMPT TO COPY (SESSION-048)

```
You are starting SESSION-048 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 6 is CLOSED (SESSION-047's honest exit
   evaluation: the three roadmap gate criteria GREEN — the golden
   paths' workflow tests (sync → batch → seals; scan → confirm →
   apply; sync-view refresh), the performance criteria (first frame
   802/281 ms ≤ 1.5 s; the 10k flood 0 drops, handler max 24 ms/
   batch), AC-11's side-by-side lock DEMONSTRATED LIVE at T-78 —
   plus the Path-B rider: the More menu's tail is EMPTY, T-79a…
   T-79f all COMPLETE). The GUI-parity RELEASE criteria (the
   Windows runtime legs + the clean-VM checklist) remain Phase 7's
   owner items exactly as ADR-0001 Am.2 registered. SESSION-048 =
   T-80 (M): PHASE 7'S OPENING ROW — the deterministic zip builder +
   version stamping + manifest/icon (doc 10 §2/§5): the reproducible
   packaging, the zip's byte-identical rebuild (the acceptance:
   sha256 equal on two builds from the same commit). The export zip
   + the Backup page's Restore button ride the same packaging
   composition (the standing assignment — SESSION-040's record +
   the honestly-disabled More-menu row + the Backup page's disabled
   Restore note). T-81 (the release workflow), T-82 (the clean-VM
   checklist), T-83 (the docs) follow — one slice done properly
   with proof beats two half-done; close cleanly at a natural
   boundary when capacity runs short. NOTE the boundary honestly:
   the sandbox has no Windows box (the standing constraint since
   SESSION-001); the cross-compile legs (wails build -platform
   windows/amd64, CGO-free) are the sandbox's proof, the live
   Windows legs are the owner's checklist items.

1. The environment is REBUILDABLE, not persistent (forty-three wipe
   events across SESSION-005..047 — SESSION-047 opened on the
   forty-third: the WHOLE environment was gone again, both repos
   included, all rebuilt from the committed recipes). NOTE the
   standing lessons: downloads and long suites must run in
   FOREGROUND CHUNKS (the sandbox kills backgrounded children when the
   call returns); this sandbox's tool-output rendering occasionally
   drops characters — settle every suspect reading with byte-level
   counts/hex before "fixing" anything (the Edit tool's space-vs-tab
   corruption struck in 042 AND 044 AND 046 AND 047 writing Go source
   — gofmt -l after every edit batch, gofmt -w restores because HEAD
   is gofmt-clean; patch with python reading the real bytes); the VLM
   CLI rate-limits (429) under rapid calls — pause between the
   screenshot verifications — AND the VLM misreads enabled-vs-disabled
   states (FOUR incidents: 042/043/044/045 — each settled with the
   byte-level source check + the pinned tests; never "fix" code off a
   suspect VLM verdict; 046 AND 047 ran clean with both passes
   agreeing); 'ln -s' refuses AND kills the whole compound bash call at
   this boot — never put it in a compound command (the venv needs
   --without-pip --copies on ensurepip-less boots; the webkit tree's
   26 dev links need Am.4 hardlinks); the BINDING-IMPORT lesson
   (044): a wails door missing from the owning module's binding
   IMPORT statement throws a silent ReferenceError — join the import
   list and pin it by test; THE FLAG-CLEAR LAW (047): SyncRunBatch
   clears the running flag + releases the lock BEFORE the finished
   signal crosses — the fanfare chain depends on it; THE
   REFERENCE'S-OWN-ARITHMETIC law (046): when a new test's
   expectation disagrees with the engine over a reference behavior,
   re-derive the reference's own computation BEFORE "fixing" the
   engine. FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the
   /home/z/second-brain-project symlink is DOWN while symlink(2) is
   blocked — use the real path), read-only reference
   /home/z/my-project/GitCurator @ 98f4fc7 (must match
   testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 ->
   /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64
   hex chars; go.dev's ?mode=json has ROTATED past 1.27.2: verify the
   locally-computed hash against the COMMITTED pin in build/
   TOOLCHAIN.md Am.2 — the pin is the authority; BEWARE the
   transcription corruption), venv /home/z/venvs/gitcurator (Python 3.12,
   stdlib-only; the plain 'python3 -m venv' form when ensurepip is
   present, else --without-pip --copies), the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install
   github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only), and
   the rootless webkit tree (~/.local/lib-wails — rebuilt at 047 [the
   FIFTEENTH rebuild]; if a wipe took it, rebuild FIRST per
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
   SESSION-047 landed T-79f in its close commits and PUSHED everything —
   2356 main-battery tests race-clean over 36 packages (2348 + 8) + 84
   webui tests over 4 (77 + 7). Verify git status shows up to date with
   origin/main. Use the PAT only ephemerally: the GIT_CONFIG_COUNT
   env-var form avoids persistence entirely — and SCRUB IT if you ever
   CLONE with it. Re-verify the suite before work if anything looks
   off — the law pays in bugs, not just safety.

2. Read, in order: the repo README.md (SESSION-047's status block —
   T-79f + PHASE 6 CLOSED) → docs/SESSION-PROTOCOL.md (the working
   law) → docs/NEXT-SESSION.md (your state card — §1 records
   SESSION-047's close) → docs/sessions/SESSION-047.md (the most
   recent report — the 43rd-wipe repair + the landing + the honest
   phase-exit evaluation) → docs/port-blueprint/10-build-release-
   packaging.md §2/§5 (+ §7/§8 for the boundary) — T-80's specs →
   the reference's own release packaging at 98f4fc7 (the
   deterministic zip's Python twin — find the module by searching
   the reference for the zip/release script) → build/mkzip/ (the Go
   zip helper already in the tree — SESSION-032's CLI-parity zip
   asset rode it; the T-65 zip-composition resolution) → doc 13's
   T-80 row → doc 15 §6 (AC-21…AC-24).

3. THIS SESSION = T-80: the deterministic zip builder + version
   stamping + manifest/icon. Follow the landed shape: the engines
   main-battery-tested over hermetic seams + the pinned reference
   citations + the byte-level evidence. One slice done properly
   with proof beats two half-done; close cleanly at a natural
   boundary when capacity runs out.

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
   (the chosen path's own acceptance), commit + push everything (ask
   me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-048.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language
   report plus the prompt I should copy to start SESSION-049.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```

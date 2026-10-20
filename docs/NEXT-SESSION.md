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
- **Phase 6 CLOSED (SESSION-033–047)**: the Wails GUI from the T-70 gate
  through the T-79a…T-79f Path-B rows (the honest exit evaluation in
  SESSION-047's report: the three roadmap gate criteria GREEN + the
  Path-B rider satisfied).
- **Phase 7 OPEN — T-80 (the deterministic zip builder + version
  stamping + manifest/icon) COMPLETE at SESSION-048**: the full builder
  (the two-exe composition + the version-stamp + brand-icon guards),
  THE DOUBLE-BUILD RITUAL GREEN (AC-22: every sha256 pair identical,
  warm ×2 + a cold leg per exe; the zip `61b4cf2a…` — 13 files,
  56,908,881 bytes in), the manifest laws (supportedOS Win10 +
  longPathAware), the version stamping pinned (the -X-under-trimpath
  law found + fixed via the reference's own boot print), **the standing
  assignment landed with it** (the 📤 Export Backup ZIP row + the Backup
  page's Restore button LIVE; the v0.86.0 ledger rider on BOTH zips —
  T-71's placeholder note retired), and the evidence at
  `docs/poc-artifacts/t-80-reproducible-packaging.md`. **T-81 (the
  release workflow) is the next row**; then T-82 (the clean-VM
  checklist) + T-83 (the docs).
- **Push state: CLEAN** — SESSION-048 committed + pushed everything at
  the close (the ephemeral PAT; never stored, never written). Verify
  `git status` in sync at the start.
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (the /home/z/second-brain-project
  symlink is DOWN while symlink(2) is blocked — use the real path),
  read-only reference /home/z/my-project/GitCurator @ `98f4fc7`
  (must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars; NOTE: go.dev's ?mode=json has ROTATED past 1.27.2: verify the
  locally-computed hash against the COMMITTED pin in build/TOOLCHAIN.md
  Am.2 — the pin is the authority; BEWARE the transcription corruption),
  venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only — the plain
  `python3 -m venv` form when ensurepip is present, else the
  `--without-pip --copies` fallback), the wails CLI v2.16.0
  (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails — **the SIXTEENTH
  rebuild is STILL PENDING after the 44th wipe; rebuild it FIRST only
  when the session runs a LINUX GUI build** (the windows cross-compile
  legs are CGO-free — SESSION-048 needed no webkit) per webui/README.md
  + docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
  144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
  hardlink fallback expected), the MAIN go.mod exactly
  modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client v1.16.0 +
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the go
  directive at 1.24.1 (never bump; ADR-0002 Am.2). The NESTED webui
  module rides go 1.25.0 + the wails require + the transitive requires
  (indirect via the parent replace — ADR-0002 Am.5; the MAIN go.mod
  stays byte-unchanged). **The downloads must run in FOREGROUND
  CHUNKS** — the sandbox kills backgrounded children when the call
  returns.
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE 44TH-WIPE RECIPE (SESSION-048)**: same as the 41st–43rd —
    clone both with the ephemeral GIT_CONFIG_COUNT PAT form, verify the
    rev pin, rebuild per the committed recipes, and re-verify the FULL
    battery (plain + race) before touching code (this session: 2356 +
    84 pre-work, 2371 + 89 at the close — 15 + 5 new).
  - **THE -X-UNDER-TRIMPATH LAW (SESSION-048)**: the linker DROPS an
    unreferenced package-level var's `-X` bytes under `-trimpath` — any
    future -X stamp must be referenced in code (T-80's fix: the
    reference's own boot print); mkzip's version-stamp guard reads the
    bytes.
  - **THE VERIFY-THE-VERIFIER LESSON (SESSION-048)**: the tool-output
    renderer ate `[main]` into `ain]` and nearly caused a phantom "fix"
    of a CORRECT file — the hex dump proved the file right. Byte-level
    verification cuts BOTH ways: check the bytes AND distrust the
    rendering of the check itself (hex > repr > rg-pipes).
  - **THE DEFAULT-ICON FIND (SESSION-048)**: wails' build/
    appicon.png + icon.ico default to the TEMPLATE artwork — any new
    wails project (or a future re-clone) must re-land the brand icon
    (icon.ico byte-equal to assets/icons/gitcurator.ico; pinned by
    TestIconLaws).
  - **THE 43RD-WIPE RECIPE (SESSION-047)**; **THE FLAG-CLEAR LAW
    (SESSION-047)**; **THE REFERENCE'S-OWN-ARITHMETIC LAW
    (SESSION-046 — paid AGAIN at 048: the ledger rider's zip-ROOT
    arcname)**; **THE 42ND-WIPE RECIPE**; **THE VLM MISREAD LEDGER
    (FOUR incidents: 042/043/044/045; 046 AND 047 ran clean with both
    passes agreeing)**; **THE BINDING-IMPORT LAW (SESSION-044)**; **THE
    RACE-PASS DIVIDEND (SESSION-045)**; **THE EXTRACT-AND-EXEC SCOPE
    LAW** (SESSION-041); **THE CROSS-BINARY DISAGREEMENT LAW** (paid at
    044, 045 AND 047 AND 048 — the Worker UA stays the reference's own
    hardcoded string, never the stamped version); **THE
    SENTINEL-SEAT RACE LAW** (SESSION-040); **THE RETROACTIVE
    FIELD-TAG AUDIT** (SESSION-040); **THE NIL-OPTIONAL DEREF**
    (SESSION-040); **THE ONE-OWNER-PER-SURFACE law** (SESSION-040);
    **THE FIELD-TAG LAW** (SESSION-038); **THE DISPLAY-PIPELINE
    CORRUPTION LESSON** (SESSION-038 — the Edit tool strikes
    SPACE-vs-TAB on Go source: gofmt -l after every edit batch, gofmt -w
    restores since HEAD is gofmt-clean; patch with python reading the
    real bytes; it struck AGAIN at 048); **THE SYMLINK-BLOCK FALLBACK**
    (SESSION-037/044/045/046/047, TOOLCHAIN.md Am.4); **THE
    RENDER-CONTEXT LAW** (SESSION-037); **THE VLM RATE-LIMIT WINDOW**
    (SESSION-039 — pause between the passes); **THE
    PINNED-VERDICT-ROW LESSON** (SESSION-039); **THE
    ONE-SLICE-PROPERLY RECORD** (036–048); **THE LIVE-CONFIG LAW**;
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

## 2. Next session: SESSION-049 — Phase 7's second row: T-81 (the release workflow)

- **SESSION-049 = T-81 (S)**: the release workflow (doc 10 §8): tag push
  → build both exes → mkzip → attach `GitCurator-v<version>-windows.zip`
  to the GitHub release (the owner's law: a release without its zip is
  not done). The row's acceptance: "dry-run on a test tag". The
  workflow wraps `build/release.sh` (T-80's reusable core — the ritual
  already proves the identity); the existing `ci.yml` is UNTOUCHED per
  the safety rules (a NEW workflow file under a new path).
- T-82 (the clean-VM checklist execution + the AV scan note) + T-83 (the
  docs) follow — one slice done properly with proof beats two
  half-done; close cleanly at a natural boundary when capacity runs
  short.
- NOTE the boundary honestly: the sandbox's workflow dry-run proves the
  automation's own legs (the workflow file's syntax, the trigger, the
  build + zip + attach calls against a test tag); the LIVE Windows
  runtime legs remain the owner's T-82 checklist items.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-048.md` — the most recent report (T-80's
   landing + the 44th-wipe repair + the finds).
2. `docs/port-blueprint/10-build-release-packaging.md` §8 (+ the §2
   dated notes T-80 landed) — T-81's spec.
3. `build/release.sh` — the ritual the workflow wraps (T-80's core).
4. `docs/poc-artifacts/t-80-reproducible-packaging.md` — the evidence
   pattern + the reproduce-from-tag recipe.
5. `docs/port-blueprint/13-implementation-task-list.md` — the T-81 row +
   the T-80 COMPLETE record.
6. The repo's existing `.github/workflows/ci.yml` (READ-ONLY — the
   safety rule) + doc 15 §6 (AC-23: a GitHub release carries the zip +
   the sha256 + the parity report).

## 4. Known constraints / owner actions for SESSION-049

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- **The webkit tree is DOWN** (the 44th wipe took it; the SIXTEENTH
  rebuild is pending) — rebuild it FIRST only if the session runs a
  LINUX GUI build (the windows legs are CGO-free; SESSION-048 worked
  without it).
- **Formal releases happen only at roadmap milestones** — T-81's
  dry-run rides a TEST TAG (deleted after the proof), never a real
  version bump; the real `v0.9x.0-go.2` release moment is the owner's
  call at the phase gate.
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush.
- The vanilla-vs-vite call (the owner's): vanilla is landed and proven;
  a flip needs only an ADR + the build step.
- **The batch-scorecard modal question (the owner's)**: the reference's
  Batch Complete modal + the stash/merge interplay is recorded as
  unlanded (T-79f's divergence #2 — its consumer was never a named
  row); the owner may name it a Phase-7 row or leave the bar+chime
  fanfare as the release shape.
- **The v0.91.0 zip parity question (the owner's, minor, NEW)**: doc 10
  §10's "parity target for the Go zip ritual" = the DISCIPLINE (the
  double-build identity + tracked-only + the fixed stamps), which the
  ritual now matches; the compositions differ by design. No action
  needed unless the owner wants a different reading.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- The go-directive acknowledgement (STILL WAITING — thirty-ninth
  session carrying it): the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the wailsjs baseline IS 755) and never commit a stray binary.
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the owner;
  neither gates the packaging row.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential; the Windows
  live legs + the launchers' clean-VM smoke + the bootstrapper leg
  (Phase 7 checklist, doc 10 §7).
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

## 5. THE PROMPT TO COPY (SESSION-049)

```
You are starting SESSION-049 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

0. THE PHASE STATE: Phase 7 is OPEN — T-80 (the deterministic zip builder +
version stamping + manifest/icon) COMPLETE at SESSION-048: the full builder
(the two-exe composition + the version-stamp + brand-icon guards), THE
DOUBLE-BUILD RITUAL GREEN (AC-22: every sha256 pair identical, warm ×2 + a
cold-cache leg per exe; the zip 61b4cf2a… — 13 files, 56,908,881 bytes in),
the manifest laws (supportedOS Win10 + longPathAware), the version stamping
pinned (THE -X-UNDER-TRIMPATH LAW found + fixed via the reference's own boot
print), the standing assignment LANDED with it (the 📤 Export Backup ZIP
row + the Backup page's Restore button LIVE; the v0.86.0 ledger rider on
BOTH zips — T-71's placeholder note retired). SESSION-049 = T-81 (S): THE
RELEASE WORKFLOW (doc 10 §8): tag push → build both exes → mkzip → attach
GitCurator-v<version>-windows.zip to the GitHub release (the owner's law: a
release without its zip is not done). The row's acceptance: "dry-run on a
test tag". The workflow wraps build/release.sh (T-80's reusable core — the
ritual already proves the identity); the existing ci.yml is UNTOUCHED per
the safety rules (a NEW workflow file under a new path). T-82 (the
clean-VM checklist) + T-83 (the docs) follow — one slice done properly
with proof beats two half-done; close cleanly at a natural boundary when
capacity runs short. NOTE the boundary honestly: the sandbox's workflow
dry-run proves the automation's own legs (the file's syntax, the trigger,
the build + zip + attach calls against a TEST TAG — never a real version
bump; formal releases happen only at roadmap milestones); the live
Windows runtime legs remain the owner's T-82 checklist items.

1. The environment is REBUILDABLE, not persistent (forty-four wipe events
   across SESSION-005..048 — SESSION-048 opened on the forty-fourth: the
   WHOLE environment was gone again, both repos included, all rebuilt from
   the committed recipes). NOTE the standing lessons: downloads and long
   suites must run in FOREGROUND CHUNKS (the sandbox kills backgrounded
   children when the call returns); this sandbox's tool-output rendering
   occasionally drops characters — settle every suspect reading with
   byte-level counts/hex before "fixing" anything (the Edit tool's
   space-vs-tab corruption struck in 042 AND 044 AND 046 AND 047 AND 048
   writing Go source — gofmt -l after every edit batch, gofmt -w restores
   because HEAD is gofmt-clean; patch with python reading the real bytes;
   AND 048's sharper edge: the renderer ate '[main]' into 'ain]' ON THE
   VERIFICATION ITSELF — a hex dump proved the file correct; distrust the
   rendering of the check too); the VLM CLI rate-limits (429) under rapid
   calls — pause between the screenshot verifications — AND the VLM
   misreads enabled-vs-disabled states (FOUR incidents: 042/043/044/045 —
   each settled with the byte-level source check + the pinned tests;
   never "fix" code off a suspect VLM verdict; 046 AND 047 ran clean
   with both passes agreeing); 'ln -s' refuses AND kills the whole
   compound bash call at this boot — never put it in a compound command
   (the venv needs --without-pip --copies on ensurepip-less boots; the
   webkit tree's 26 dev links need Am.4 hardlinks); THE -X-UNDER-TRIMPATH
   LAW (048): the linker drops an unreferenced var's -X bytes under
   -trimpath — any future -X stamp must be referenced in code; THE
   DEFAULT-ICON FIND (048): wails' build/appicon.png + icon.ico default
   to the TEMPLATE artwork — any future wails project must re-land the
   brand icon; THE BINDING-IMPORT lesson (044): a wails door missing
   from the owning module's binding IMPORT statement throws a silent
   ReferenceError — join the import list and pin it by test; THE
   FLAG-CLEAR LAW (047): SyncRunBatch clears the running flag + releases
   the lock BEFORE the finished signal crosses — the fanfare chain
   depends on it; THE REFERENCE'S-OWN-ARITHMETIC law (046, paid again at
   048: the ledger rider's zip-ROOT arcname): when a new test's
   expectation disagrees with the engine over a reference behavior,
   re-derive the reference's own computation BEFORE "fixing" the engine.
   FIRST verify with evidence: work repo at
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
   the rootless webkit tree (~/.local/lib-wails — DOWN after the 44th
   wipe; the SIXTEENTH rebuild is pending: rebuild it FIRST only when
   the session runs a LINUX GUI build — the windows cross-compile legs
   are CGO-free, SESSION-048 needed no webkit — per webui/README.md +
   docs/poc-artifacts/t-70-gate.md + build/TOOLCHAIN.md Am.3 + Am.4:
   144 debs → dpkg -x → the 42 .pc prefix rewrites → the 26 dev links,
   hardlink fallback expected). The MAIN go.mod: modernc.org/sqlite at
   v1.36.1, bogdanfinn/tls-client at v1.16.0, golang.org/x/net v0.48.0 +
   golang.org/x/text v0.32.0 direct, and the go directive at 1.24.1 (the
   tls-client-forced floor — ADR-0002 Am.2; do not bump any of these).
   The NESTED webui module rides go 1.25.0 (wails v2.16.0's floor) + the
   wails require + the transitive requires the direct internal/app import
   pulls (indirect via the parent replace — ADR-0002 Am.5); the MAIN
   go.mod stays byte-unchanged. Then check 'git status' and 'git log
   --oneline -3': SESSION-048 landed T-80 in its close commits and
   PUSHED everything — 2371 main-battery tests race-clean over 36
   packages (2356 + 15) + 89 webui tests over 4 (84 + 5). Verify git
   status shows up to date with origin/main. Use the PAT only
   ephemerally: the GIT_CONFIG_COUNT env-var form avoids persistence
   entirely — and SCRUB IT if you ever CLONE with it. Re-verify the
   suite before work if anything looks off — the law pays in bugs, not
   just safety.

2. Read, in order: the repo README.md (SESSION-048's status block —
   T-80 COMPLETE) → docs/SESSION-PROTOCOL.md (the working law) →
   docs/NEXT-SESSION.md (your state card — §1 records SESSION-048's
   close) → docs/sessions/SESSION-048.md (the most recent report — the
   44th-wipe repair + the landing + the finds) → docs/port-blueprint/
   10-build-release-packaging.md §8 (+ the §2 dated notes T-80 landed) —
   T-81's spec → build/release.sh (the ritual the workflow wraps) →
   docs/poc-artifacts/t-80-reproducible-packaging.md (the evidence
   pattern) → doc 13's T-81 row → the repo's existing
   .github/workflows/ci.yml (READ-ONLY — the safety rule) → doc 15 §6
   (AC-23).

3. THIS SESSION = T-81: the release workflow. Follow the landed shape:
   the engines main-battery-tested over hermetic seams + the pinned
   reference citations + the byte-level evidence. The dry-run rides a
   TEST TAG (deleted after the proof) — never a real version bump.
   One slice done properly with proof beats two half-done; close cleanly
   at a natural boundary when capacity runs out.

4. Standing laws: reality wins over specs (fix the spec in the same
   commit); tests are never weakened; no secrets or owner data in git
   (byte-level secret-scan before every commit; never write
   token-shaped strings even as fakes); no new dependencies without an
   ADR (the authorized runtime set is exactly modernc.org/sqlite
   v1.36.1, tls-client v1.16.0, x/net v0.48.0, x/text v0.32.0 — do not
   bump any; the MAIN go.mod stays byte-unchanged; the wails require +
   the indirect transitive requires stay in the NESTED webui module per
   ADR-0002 Am.5); surface blockers with evidence (hash-level, not
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
   read+written by concurrent passes holds it under a mutex as ONE unit
   (THE SENTINEL-SEAT RACE LAW — run the battery ≥3× before trusting a
   new concurrent seam; every live-map mutation holds the Settings
   mutex, the SaveSettings discipline); every pointer/optional field
   read is settled with a nil check (THE NIL-OPTIONAL DEREF); one owner
   per frontend surface (duplicate EventsOn listeners double-paint); a
   modal's action buttons live OUTSIDE the scrollable body; every new
   wails door JOINS the owning module's binding IMPORT statement (THE
   BINDING-IMPORT LAW — pin it by test); after every wails build, check
   'git diff --raw' for mode flips (the wailsjs committed baseline IS
   755) and never commit a stray binary; gofmt -l after every Go edit
   batch (the Edit tool's space-vs-tab corruption — gofmt -w restores);
   every git/go command carries its own cd AND its own
   PATH=/home/z/.local/go/bin export (the sandbox shell resets both
   between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation
   (the chosen path's own acceptance), commit + push everything (ask
   me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-049.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language
   report plus the prompt I should copy to start SESSION-050.

One slice done properly with proof beats two half-done; close cleanly
at a natural boundary when capacity runs out.
```

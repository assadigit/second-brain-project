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
  (all 31 generators × two passes GREEN — now 33 with theme+icons).
- **Phase 6 OPEN (SESSION-033–040)** — T-70 COMPLETE (the entry gate),
  T-71 COMPLETE (`internal/app`, SESSION-034), T-72 COMPLETE (the
  frontend foundation, SESSION-035), T-73 COMPLETE (the Settings
  pages A, SESSION-036), T-74 COMPLETE (the Settings pages B + the
  CLI-parity shell wiring, SESSION-037), T-75 COMPLETE (the Sync
  view, SESSION-038), T-76 COMPLETE (the Scan flow, SESSION-039),
  and now **T-77 COMPLETE: THE MAIN WINDOW CHROME** (SESSION-040,
  two slices: `9cd7586` — internal/app/herostate.go the pure engines
  (HeroMachine — hero.py's GUI-only state machine: the four states,
  the click guards as data incl. the offline advisory + the
  Option-B-re-worded nothing box, the fetch routing + the caught-up
  law, the mirror, the stopping render, the exact tooltips, the >10
  gate, the progress formats, the sentinel paint laws) + **THE
  RETROACTIVE FIELD-TAG FIX** (the T-71 bus payloads crossed tagless
  — the progress bar NEVER filled and every finish read ❌; every
  crossed payload re-tagged + pinned by test) + the cooperative
  StopBatch door + the sentinel pass's full snapshot/offline-law
  emission (with TWO reality fixes the re-read caught: the
  nil-latency panic on the offline path + the sentinelOnline data
  race) + **the WAVs SHIPPED** (success/retry embedded, disk-first,
  the quiet batch gate); `c14fb8f` — the shell: webui/heroui.go the
  SIXTH bound struct (HeroClick the one click door with both stages,
  the 200 ms mirror, TestConnection the battery door, the More
  menu's live doors) + frontend/heroflow.js (the hero renders, the
  guard boxes, the sentinel paint, the battery modal with Start
  Syncing gated on all-ok, the More menu's verbatim list with the
  honest-disabled law, the batch chime)). **The acceptance GREEN:
  hero state machine tests + the startup cadence battery — 2252
  main-battery tests race-clean over 36 packages (2230 + 22) + 56
  webui tests over 4 (40 + 16), stable across three consecutive
  full -race runs**; the gate leg GREEN (first frame 908 ms ≤ 1.5 s;
  the 10k flood 555,555 lines/s, 0 drops, 26 ms max handler); the
  screenshot VLM-verified (the sentinels painting LIVE: Net: online
  green / AI: off — unreachable red / Proxy: off gray; the SYNC hero
  enabled; the honest placeholders retired). The 35th full wipe was
  absorbed before work (see the session report).
- **Push state: CLEAN** — SESSION-040 pushed after every slice;
  verify `git status` is in sync at the start (the tip at or beyond
  `c14fb8f`).
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
  only), and the rootless webkit tree (~/.local/lib-wails: apt
  --download-only + dpkg -x of libwebkit2gtk-4.1-dev + libgtk-3-dev,
  the .pc prefix rewrites, the 26 dev-symlink relinks — **if the
  sandbox blocks symlink(2) again, TOOLCHAIN.md Am.4's hardlink
  fallback: hardlink each dangling dev link to its RESOLVED system
  target** — NOTE the link's readlink target is the versioned .so,
  resolve THAT against the system tree; the block now bites at BOOT,
  a plain `ln -s` is refused outright), all per build/TOOLCHAIN.md
  Am.3 + Am.4; the MAIN go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (never
  bump; ADR-0002 Am.2). The NESTED webui module rides go 1.25.0
  (wails v2.16.0's floor) + the wails require + the transitive
  requires the direct internal/app import pulls (indirect via the
  parent replace — ADR-0002 Am.5; the MAIN go.mod stays
  byte-unchanged). **The downloads must run in FOREGROUND CHUNKS** —
  the sandbox kills backgrounded children when the call returns (the
  35th-wipe repair paid this once).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE SENTINEL-SEAT RACE LAW** (SESSION-040): a pass that reads a
    transition seat and writes it back (the offline verdict) must hold
    it under a mutex as ONE unit — the lifecycle's tick and the save
    law's 250 ms re-arm run concurrently; the race detector catches it
    only on re-runs (run the battery ≥3× before trusting a new
    concurrent seam).
  - **THE RETROACTIVE FIELD-TAG AUDIT** (SESSION-040): every payload
    that crosses to the frontend carries tags mirroring the keys the
    JS reads — the T-71 tagless shapes sat BROKEN (the bar never
    filled, ❌ always) through five sessions because nothing pinned
    the marshaled keys; TestCrossedPayloadFieldTags is the standing
    guard, extend it with every new payload.
  - **THE NIL-OPTIONAL DEREF** (SESSION-040): a probe's absent field
    (LatencyMs nil offline) dereferenced unconditionally panics into
    the crash-isolation and the feature silently never runs — settle
    every `*T` field read with a nil check.
  - **THE ONE-OWNER-PER-SURFACE law** (SESSION-040): when a second
    frontend module takes over a surface (the bar, the chime), the
    first owner's handlers must hand it over — duplicate EventsOn
    listeners double-paint.
  - **THE FIELD-TAG LAW** (SESSION-038): the payload structs' JSON
    tags mirror the Python dict keys verbatim (snake_case:
    `dot_tone`, `status_text`, `measured_at`…) — the frontend
    reads the TAGGED names, never the camelCase guesses.
  - **THE DISPLAY-PIPELINE CORRUPTION LESSON** (SESSION-038): this
    sandbox's tool-output rendering occasionally drops characters
    (`[hidden]` selectors reading as `idden]`, rg output mangling
    type names mid-token) — settle every suspect reading with
    byte-level counts/hex/od before "fixing" anything (paid TWICE
    more in 040: MultiEdit tab/space mismatches — patch with python
    reading the real bytes, never trust rendered whitespace).
  - **THE SYMLINK-BLOCK FALLBACK** (SESSION-037, TOOLCHAIN.md Am.4):
    a wipe rebuild on a symlink-blocking sandbox relinks the 26 dev
    .so links as hardlinks to the RESOLVED system targets (readlink
    -f the LINK'S TARGET against the system tree); the wails build
    over the tree is the proof. Now blocked at BOOT.
  - **THE RENDER-CONTEXT LAW** (SESSION-037): a pure renderer that
    reads config context takes the CONFIG as an argument.
  - **THE VLM RATE-LIMIT WINDOW** (SESSION-039): the vision CLI
    429s under rapid calls — batch the screenshot verifications with
    pauses, and crop+zoom for the small details.
  - **THE PINNED-VERDICT-ROW LESSON** (SESSION-039): a modal's
    action buttons must live OUTSIDE the scrollable body.
  - **THE ONE-SLICE-PROPERLY RECORD** (SESSION-036, paid again in
    037/038/039/040) — the wiring slice lands complete with its own
    proof before the pages.
  - **THE LIVE-CONFIG LAW** (SESSION-036); **THE GROUP-POINTER LAW**
    (SESSION-036); **THE DISSOLUTION-PROBE RECORD** (SESSION-035,
    ADR-0002 Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP** (SESSION-035);
    **THE SUITE-REVERIFICATION DIVIDEND** (SESSION-035, paid twice
    more in 040 — the nil-deref + the race); **THE WAILS-BUILD MODE
    FLIPS** (SESSION-035: the wailsjs committed baseline IS 755; a
    stray `go build .` in webui/ leaves a 22 MB `webui` binary);
    **THE HERMETIC-SEAM LAW** (SESSION-034); **THE RESUMED-CONTEXT
    RE-VERIFICATION LAW** (SESSION-034); **THE MODE-FLIP
    DEGRADATION LAW** (SESSION-033); **THE MODULE-BOUNDARY PROBE
    LAW** (SESSION-033, ADR-0002 Am.4); **THE FRONTEND-MODULE
    LESSON** (SESSION-033); **THE TIMER-CLAMP INSTRUMENT CAVEAT**
    (SESSION-033); **THE TOOLCHAIN-ROUND-2/3 PROOF** (SESSION-033/
    035); THE T-65 ZIP-COMPOSITION RESOLUTION; THE
    CORPUS-REGENERATION LAW; THE DISPATCH-COMPLETION LAW; THE
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

## 2. Next session: SESSION-041 — T-78 (S): the side-by-side verification vs the Python app (the lock mutual exclusion, the same-data laws) — the LAST Phase 6 task

- **Verify the push state first** (§1): confirm the tip is at or
  beyond `c14fb8f`.
- **T-78 per doc 13's row** (ADR-0008): the side-by-side verification
  — the Go app and the Python app run against the SAME data (the
  same app/ folder: config.json, cache.db, the vaults), proving the
  compatibility laws live: the app.lock mutual exclusion ACROSS THE
  TWO BINARIES (the Go AcquireAppLock vs the reference's app.py lock
  — the same file, the same PID-liveness law, the alive-refusal), the
  merge-never-rebuild config law on the reference's own config shape,
  the byte-comparable JSON (the SetEscapeHTML(false) law), the
  user-config fill/mirror. ACCEPTANCE: lock test across both
  binaries.
- **Scope honestly**: the reference's GUI needs PyQt6 — the venv is
  stdlib-only by law. Probe what the reference's lock path needs as an
  import (gui/app.py's lock lives at module import time — a PyQt6
  import chain). If the GUI import is unavoidable, surface it with
  evidence and scope the leg to what the law actually reads (the lock
  FILE's bytes + the PID law are testable binary-to-binary without
  Qt: the reference's own lock code is a small module — check whether
  a minimal harness can drive it, or whether the acceptance's "both
  binaries" means the CLI shell's lock for the Python side). Do NOT
  improvise past a real blocker — record it.
- **The bookkeeping question for the owner** (surfaced, not acted on):
  the More menu's honestly-disabled tail (the per-test dialogs, the
  linking tools, the bot-queue tool dialogs, the Chrome doors, the
  decommission picker, the export zip) has no named task on doc 13 —
  the owner's call: a tasks addendum vs folding into the Phase-7
  parity report.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-040.md` — the most recent report (T-77
   COMPLETE; the two reality fixes; the race law).
2. `docs/port-blueprint/decisions/0008-config-data-compat.md` — THE
   AUTHORITY for T-78 (the same-data laws, the byte-comparable JSON,
   the side-by-side install).
3. `docs/port-blueprint/08-data-and-compatibility.md` — the data
   compatibility plan (§ the side-by-side law).
4. The reference's lock + config code: `app/gitcurator/gui/app.py`
   (the app.lock acquisition at startup) + `app/gitcurator/constants.py`
   (APP_DIR) + `app/gitcurator/core/storage.py` (the JSON write laws).
5. The port's twins: `internal/app/lifecycle.go` (AcquireAppLock) +
   `internal/config/` (LoadAppConfig/Save — the merge + escape laws)
   + `internal/constants/`.
6. `docs/port-blueprint/13-implementation-task-list.md` — the T-78
   row + Phase 7's book (T-80..T-83).
7. `docs/port-blueprint/15-*.md` (the acceptance criteria doc) —
   Phase 6's exit gate wording.

## 4. Known constraints / owner actions for SESSION-041

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush; T-78
  proceeds on (b) either way.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails +
  the indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — thirty-first
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode
  flips (the bindings re-flip to 755 — the committed baseline IS 755
  for the wailsjs files; only a CONTENT diff should remain) and
  never commit a stray `webui/webui` binary.
- **If the sandbox blocks symlink(2) during a wipe rebuild**:
  TOOLCHAIN.md Am.4's hardlink fallback for the 26 dev links (now
  blocked at boot — expect it).
- The answering-window belt + the T-71 merged-verdict question
  (SESSION-039's report): two recorded divergences awaiting the
  owner; neither gates T-78.
- The More-menu disabled tail's bookkeeping question (§2) — surface
  it to the owner; do not improvise a task.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs + the launchers' clean-VM smoke + the
  bootstrapper leg (Phase 7 checklist, doc 10 §7); the Backup page's
  Restore button (honestly disabled — the safety-net restore +
  export-zip engines ride the packaging phase); the More menu's
  disabled doors (§2's bookkeeping question).
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

## 5. THE PROMPT TO COPY (SESSION-041)

```
You are starting SESSION-041 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (thirty-five wipe events
   across SESSION-005..040 — SESSION-040 absorbed the thirty-fifth full wipe
   BEFORE work: both repos, the toolchain, the venv, the module cache, AND the
   rootless webkit tree all gone, everything rebuilt from the committed recipes
   — the eighth webkit rebuild, Am.4's hardlink fallback again: symlink(2) is
   blocked at BOOT (a plain ln -s is refused outright), the 26 dev .so links
   land as hardlinks to the RESOLVED system targets (readlink -f the LINK'S
   TARGET — the versioned .so — against the system tree; the wails build + the
   gate leg the proof); NOTE the standing lessons: downloads and long suites
   must run in FOREGROUND CHUNKS (the sandbox kills backgrounded children when
   the call returns); this sandbox's tool-output rendering occasionally drops
   characters — settle every suspect reading with byte-level counts/hex before
   "fixing" anything (paid twice in 040 on MultiEdit tab/space mismatches —
   patch with python reading the real bytes); the VLM CLI rate-limits (429)
   under rapid calls — pause between the screenshot verifications; the
   continuation recipe is PROVEN at all levels: git status + git log --oneline
   -5 + file timestamps tell what ACTUALLY survived; a preserved commit or
   in-flight draft is continued and AUDITED — never blindly trusted, never
   restarted from zero; a resumed context re-verifies from scratch (the full
   battery + the hygiene scans) before committing; a wipe event may flip
   tracked files' MODES 644->755 while preserving content — diagnose via 'git
   diff --raw' and restore via chmod off the parsed list, NEVER git checkout;
   and the standing addendum: after every `wails build` the wailsjs bindings
   re-flip to 755 (the repo's committed baseline for the wailsjs files IS 755
   — only a CONTENT diff should remain after a build), and a stray `go build .`
   in webui/ leaves a 22 MB binary named `webui` — never commit it). FIRST
   verify with evidence: work repo at /home/z/my-project/second-brain-project
   (the /home/z/second-brain-project symlink is DOWN while symlink(2) is
   blocked — use the real path), read-only reference /home/z/my-project/GitCurator
   @ 98f4fc7 (symlinked /home/z/GitCurator — same caveat; must match
   testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 -> /home/z/.local/go
   (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON; BEWARE the transcription
   corruption — the committed pin in build/TOOLCHAIN.md Am.2 is the authority),
   venv /home/z/venvs/gitcurator (Python 3.12, stdlib-only) per
   build/TOOLCHAIN.md, the wails CLI v2.16.0 (GOBIN=/home/z/.local/go/bin go
   install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0 — build-time only),
   and the rootless webkit tree (~/.local/lib-wails per webui/README.md +
   docs/poc-artifacts/t-70-gate.md) per build/TOOLCHAIN.md Am.3 + Am.4. The
   MAIN go.mod: modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client at
   v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and
   the go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do
   not bump any of these). The NESTED webui module rides go 1.25.0 (wails
   v2.16.0's floor) + the wails require + the transitive requires the direct
   internal/app import pulls (indirect via the parent replace — ADR-0002
   Am.5); the MAIN go.mod stays byte-unchanged. Then check `git status` and
   `git log --oneline -3`: SESSION-040 landed T-77 COMPLETE in two slices plus
   the docs (9cd7586 — internal/app/herostate.go the hero machine's pure
   engines + THE RETROACTIVE FIELD-TAG FIX (the T-71 tagless bus payloads —
   the progress bar never filled, every finish read ❌; every crossed payload
   re-tagged + pinned by TestCrossedPayloadFieldTags) + the cooperative
   StopBatch door + the sentinel pass's full snapshot emission with TWO
   reality fixes (the nil-latency panic + the sentinelOnline data race — the
   seat now one unit under sentinelMu) + the WAVs shipped (success/retry
   embedded, disk-first); c14fb8f — the shell: webui/heroui.go the SIXTH
   bound struct (HeroClick the one click door, the 200 ms mirror,
   TestConnection the battery door, the More menu's live doors) +
   frontend/heroflow.js (the hero renders, the sentinel paint, the battery
   modal, the More menu's verbatim list with the honest-disabled law, the
   batch chime) + the docs commit riding the close) and PUSHED everything
   after every slice — 2252 main-battery tests race-clean over 36 packages +
   56 webui tests over 4, stable across three consecutive full -race runs.
   Verify git status shows up to date with origin/main (the tip at or beyond
   c14fb8f). Use the PAT only ephemerally: git -c
   http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety (SESSION-040's re-read caught a nil-deref panic AND
   a data race in shipped code).

2. Read, in order: the repo README.md (Phase 6 OPEN — T-77 at c14fb8f) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-040's close: T-77 COMPLETE, the newest
   laws — THE SENTINEL-SEAT RACE LAW, THE RETROACTIVE FIELD-TAG AUDIT, THE
   NIL-OPTIONAL DEREF, THE ONE-OWNER-PER-SURFACE law) →
   docs/sessions/SESSION-040.md (the most recent report) →
   docs/port-blueprint/decisions/0008-config-data-compat.md (THE AUTHORITY
   for T-78 — the same-data laws, the byte-comparable JSON, the
   side-by-side install) → docs/port-blueprint/08-data-and-compatibility.md
   (the side-by-side law) → the reference's gui/app.py (the app.lock
   acquisition) + constants.py + core/storage.py (the JSON write laws) → the
   port's twins: internal/app/lifecycle.go (AcquireAppLock) + internal/config
   (LoadAppConfig/Save) + internal/constants → doc 13's T-78 row + Phase 7's
   book.

3. THIS SESSION IS T-78 (S): THE SIDE-BY-SIDE VERIFICATION VS THE PYTHON APP
   per doc 13's row (ADR-0008): the Go app and the Python app against the
   SAME data (the same app/ folder: config.json, cache.db, the vaults),
   proving the compatibility laws live — the app.lock mutual exclusion
   ACROSS THE TWO BINARIES (the same file, the same PID-liveness law, the
   alive-refusal), the merge-never-rebuild config law on the reference's own
   config shape, the byte-comparable JSON (SetEscapeHTML(false)), the
   user-config fill/mirror. ACCEPTANCE: lock test across both binaries.
   Scope honestly: the reference's GUI needs PyQt6 — the venv is
   stdlib-only by law; probe what the reference's lock path needs as an
   import, and if the GUI import chain is unavoidable, surface it with
   evidence and scope the leg to what the law actually reads (the lock
   FILE's bytes + the PID law are testable binary-to-binary; check whether
   a minimal harness can drive the reference's own lock code). Do NOT
   improvise past a real blocker — record it. ALSO surface the owner's
   bookkeeping question (recorded in NEXT-SESSION §2): the More menu's
   honestly-disabled tail (the per-test dialogs, the linking tools, the
   Chrome doors, the export zip) has no named task on doc 13 — the owner's
   call: a tasks addendum vs folding into the Phase-7 parity report. If
   capacity runs short: one slice done properly with proof beats two
   half-done; close cleanly at a natural boundary. T-78 is the LAST Phase 6
   task — if it lands green AND capacity remains, the phase-exit question
   (opening Phase 7 per the protocol's gate laws) is the owner's to see in
   the report, not to improvise.

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
   network-touching seam in tests at construction (the HERMETIC-SEAM law);
   the payload structs' JSON tags mirror the Python dict keys and the
   frontend reads the tagged names (THE FIELD-TAG LAW — extend
   TestCrossedPayloadFieldTags with every new payload); a transition seat
   read+written by concurrent passes holds it under a mutex as ONE unit
   (THE SENTINEL-SEAT RACE LAW — run the battery ≥3× before trusting a new
   concurrent seam); every *T field read is settled with a nil check (THE
   NIL-OPTIONAL DEREF); one owner per frontend surface (duplicate EventsOn
   listeners double-paint); a modal's action buttons live OUTSIDE the
   scrollable body; after every wails build, check `git diff --raw` for
   mode flips (the wailsjs committed baseline IS 755) and never commit a
   stray webui/webui binary; every git/go command carries its own cd AND
   its own PATH=/home/z/.local/go/bin export (the sandbox shell resets
   both between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-78's
   own acceptance: lock test across both binaries), commit + push everything
   (ask me for the GitHub PAT the moment a push is due — never store it
   anywhere; never end the session push-pending), write
   docs/sessions/SESSION-041.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus
   the prompt I should copy to start SESSION-042.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

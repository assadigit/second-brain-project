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
- **Phase 6 OPEN (SESSION-033–037)** — T-70 COMPLETE (the entry gate),
  T-71 COMPLETE (`internal/app`, SESSION-034), T-72 COMPLETE (the
  frontend foundation, SESSION-035), T-73 COMPLETE (the Settings
  pages A, SESSION-036), and now **T-74 COMPLETE: THE SETTINGS PAGES
  B + THE CLI-PARITY SHELL WIRING** (SESSION-037, three slices: `a4619fe`
  — internal/app/batchwire.go, engine.go's runImpl composition
  re-hosted on the app surface (the warmup family, the inputs
  fetchers, the full intake, the banishment + settled ledger +
  backfill + the §4.4 passes, THE WEBSITES PHASE, THE FINISH CHAIN's
  seals + THE PHASE-5 CLEAR + the dry-run law) over the App's six new
  seams — THE DEBT T-73 CARRIED, CLOSED with 19 wiring tests; `6d479fb`
  — internal/app/settingsb.go, the pages' engines (THE QUEUE REPORT
  RENDERING from canned reports + the zinc ⏳/mint ✅ badge + the
  dashboard engines + the undo engine + the sources fetch over its
  seam + the backup ladders + the save groups) — 21 tests; `3091d76`
  — webui/settingsuib.go the third bound struct + frontend/
  settings.js's four pages at the reference's own sidebar order
  (Dashboard after Input, Bot, Sources, Backup — every label verbatim;
  the Restore button honestly disabled with its landing note; the
  SB_POC_OPEN_SETTINGS section probe) — 5 shell tests). **The
  acceptance GREEN: the wiring slice's own tests + the queue report
  rendering tests + the fast-sync laws (the already-tested pure
  halves) — 2060 main-battery tests race-clean over 35 packages
  (2020 + 19 + 21) + 28 webui tests over 4 (23 + 5).** The gate leg
  re-run GREEN (first frame 283/942 ms; the 10k flood 476,190
  lines/s, 0 drops); the screenshots VLM-verified (all TEN sidebar
  rows in the reference's order; the Backup page's groups with every
  button/checkbox/placeholder/status dot). The 32nd full wipe was
  absorbed before work (see build/TOOLCHAIN.md Am.4 — the
  symlink-syscall degradation + the hardlink fallback).
- **Push state: CLEAN** — SESSION-037 pushed after every task; verify
  `git status` is in sync at the start (the tip at or beyond
  `3091d76`).
- **The environment recipe** (GUI edition): work repo at
  /home/z/my-project/second-brain-project (symlinked
  /home/z/second-brain-project), read-only reference
  /home/z/my-project/GitCurator @ `98f4fc7` (symlinked /home/z/GitCurator;
  must match testdiff/PINNED-GITCURATOR-REV), toolchain go1.27.2 →
  /home/z/.local/go (go.dev tarball, sha256
  `ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5` — 64
  hex chars, verified against go.dev's published JSON), venv
  /home/z/venvs/gitcurator (Python 3.12, stdlib-only), the wails CLI
  v2.16.0 (`GOBIN=/home/z/.local/go/bin go install
  github.com/wailsapp/wails/v2/cmd/wails@v2.16.0` — build-time only),
  and the rootless webkit tree (~/.local/lib-wails: apt
  --download-only + dpkg -x of libwebkit2gtk-4.1-dev + libgtk-3-dev,
  the .pc prefix rewrites, the 26 dev-symlink relinks — **if the
  sandbox blocks symlink(2) again, TOOLCHAIN.md Am.4's hardlink
  fallback: hardlink each dangling dev link to its RESOLVED system
  target**), all per build/TOOLCHAIN.md Am.3 + Am.4; the MAIN go.mod
  exactly modernc.org/sqlite v1.36.1 + bogdanfinn/tls-client v1.16.0 +
  golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 + the go
  directive at 1.24.1 (never bump; ADR-0002 Am.2). The NESTED webui
  module rides go 1.25.0 (wails v2.16.0's floor) + the wails require +
  the transitive requires the direct internal/app import pulls
  (indirect via the parent replace — ADR-0002 Am.5; the MAIN go.mod
  stays byte-unchanged).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE SYMLINK-BLOCK FALLBACK** (SESSION-037, TOOLCHAIN.md Am.4):
    a wipe rebuild on a symlink-blocking sandbox relinks the 26 dev
    .so links as hardlinks to the RESOLVED system targets (readlink -f
    first); the wails build over the tree is the proof.
  - **THE RENDER-CONTEXT LAW** (SESSION-037): a pure renderer that
    reads config context (the websites switch, the vault path) takes
    the CONFIG as an argument — a dead local `websiteVault := ""`
    silently dropped the whole websites sub-table until the tests
    caught it; the canned-result tests must assert the context-fed
    lines.
  - **THE ONE-SLICE-PROPERLY RECORD** (SESSION-036) — paid again this
    session: the wiring slice landed complete with its own proof
    before the pages.
  - **THE LIVE-CONFIG LAW** (SESSION-036): the session-held map, the
    in-place saves; new pages build on it.
  - **THE GROUP-POINTER LAW** (SESSION-036): a page's save group
    carries ONLY its own keys; the merge never whitelist-rebuilds.
  - **THE DISSOLUTION-PROBE RECORD** (SESSION-035, ADR-0002 Am.5):
    probe, record, never improvise past a blocker.
  - **THE GOLDEN-MARSHAL-ORDER TRAP** (SESSION-035): Go maps marshal
    keys SORTED; structs emit DECLARATION order.
  - **THE SUITE-REVERIFICATION DIVIDEND** (SESSION-035): re-verifying
    the suite after a wipe catches REAL production races.
  - **THE WAILS-BUILD MODE FLIPS** (SESSION-035): every `wails build`
    re-flips the wailsjs bindings to 755; a stray `go build .` in
    webui/ leaves a 22 MB `webui` binary — check `git diff --raw`
    after every build, restore via chmod, never commit the binary.
  - **THE HERMETIC-SEAM LAW** (SESSION-034); **THE
    RESUMED-CONTEXT RE-VERIFICATION LAW** (SESSION-034); **THE
    MODE-FLIP DEGRADATION LAW** (SESSION-033); **THE
    MODULE-BOUNDARY PROBE LAW** (SESSION-033, ADR-0002 Am.4); **THE
    FRONTEND-MODULE LESSON** (SESSION-033: ES modules need
    `<script type="module">` + explicit `.js` extensions); **THE
    TIMER-CLAMP INSTRUMENT CAVEAT** (SESSION-033); **THE
    TOOLCHAIN-ROUND-2/3 PROOF** (SESSION-033/035); THE T-65
    ZIP-COMPOSITION RESOLUTION; THE CORPUS-REGENERATION LAW; THE
    DISPATCH-COMPLETION LAW; THE STATE-CHAIN + MUTATION-TEETH laws;
    THE CONTEXT-DEATH CONTINUATION recipe; THE DSN-EMPTY-PATH HAZARD;
    THE FIXED-LENGTH ROOTS; THE SHALLOW-COPY SNAPSHOT; the fake-clock
    arithmetic; the exception class-name tokens; the
    stdout-capture-helper limit.
  - **THE CD-DISCIPLINE LAW** (every git/go command carries its own
    cd — the sandbox's shell resets cwd between calls; never commit
    outside the work repo; never checkout blindly); **THE
    RENDERER-EATS-ESCAPE-PAIRS lesson** (commit messages via -F
    FILE); **THE WRITE-TOOL REDACTION lesson** (byte-level secret-scan
    before every commit); **THE BACKGROUND-PROCESS lesson** (long
    suites in FOREGROUND CHUNKS); **PUSH AFTER EVERY TASK** (ask for
    the PAT the moment a push is due; never store it).

## 2. Next session: SESSION-038 — T-75 (L): Sync view — the flow cards + the matrix + the differences modal

- **Verify the push state first** (§1): confirm the tip is at or beyond
  `3091d76`.
- **T-75 per doc 13's row**: the v0.87.0 GRAND CONVERGENCE page — ONE
  Sync page, two desks: THE VISUAL SYNC DISPLAY (the four-element
  flow: 🤖 bot → 📓 vault → ☁️ GitHub → 🌐 the site, read-only,
  refreshed on demand — the click is the consent for every network
  leg) ABOVE THE DEFERRED LEDGER (the stage-by-stage gap matrix —
  what is waiting, every gap fixed or explained); one Refresh
  measures both. Plus the differences modal (spec-gui-sync-views is
  the authority). ACCEPTANCE: the four-state rendering tests + the
  deferred-ledger UX tests.
- **The engine legs already live**: batchwire.go's review_retry +
  master_retry modes (the retry doors' engines), Commands.SyncFetchQueue
  (the queue stage's data), the drift-seal family (FastSync), and the
  queue check's classification counts (the Sync matrix's Queue →
  Vault cells feed off them — _bot_queue_counts' twin). The Sync
  page's UI (the sidebar row between Dashboard and Bot — the 'refresh'
  glyph is in the pack) wires to these.
- The Settings window is the established host (settingsui.go +
  settingsuib.go + settings.js); the Sync page joins SECTIONS at the
  reference's own order.

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-037.md` — the most recent report (T-74
   COMPLETE; the wiring re-host; the symlink-block fallback).
2. `docs/port-blueprint/module-specs/spec-gui-sync-views.md` — THE
   AUTHORITY for T-75 (the four-element flow + the gap matrix + the
   differences modal + the v0.87.0 convergence record).
3. `internal/app/batchwire.go` + `batchwire_test.go` — the retry
   modes' engines (review_retry / master_retry) the Sync page wires.
4. `internal/app/commands.go` (SyncFetchQueue / the drift-seal
   family) + `internal/app/queue.go` (the classification counts the
   matrix consumes).
5. `webui/settingsuib.go` + `frontend/settings.js` — the established
   page-binding + builder patterns (the third struct + the pages B's
   own builders; the Sync page rides the same shapes).
6. The reference's `gui/main_window/sync_view.py` + `sync_flow.py` +
   `sync_display.py` + `core/sync_review.py` — the convergence page's
   own source (read at /home/z/GitCurator).
7. `docs/port-blueprint/13-implementation-task-list.md` — the T-75
   row + Phase 6's book.

## 4. Known constraints / owner actions for SESSION-038

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush; T-75
  proceeds on (b) either way.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails + the
  indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — twenty-eighth
  session carrying it)**: the main module rides `go 1.24.1`, forced by
  the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755) and never commit a stray `webui/webui`
  binary. **A stray `go build .` in webui/ LEAVES the 22 MB binary —
  remove it before staging.**
- **If the sandbox blocks symlink(2) during a wipe rebuild**:
  TOOLCHAIN.md Am.4's hardlink fallback for the 26 dev links.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs + the launchers' clean-VM smoke + the
  bootstrapper leg (Phase 7 checklist, doc 10 §7); the sound WAV
  assets (not yet committed — the preview's missing-asset notice is
  the parity behavior until T-77/T-80 ship the files); **the Backup
  page's Restore button** (honestly disabled — the safety-net restore
  + export-zip engines ride the packaging phase, recorded in
  spec-gui-pages-b's dated note).
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

## 5. THE PROMPT TO COPY (SESSION-038)

```
You are starting SESSION-038 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (thirty-two wipe events across
   SESSION-005..037 — SESSION-037 absorbed the thirty-second full wipe BEFORE
   work: both repos, the toolchain, the venv, the module cache, AND the
   rootless webkit tree all gone, everything rebuilt from the committed recipes
   — the fifth webkit rebuild; NOTE the new Am.4: mid-rebuild the sandbox began
   BLOCKING symlink(2) outright, so the 26 dev .so links landed as HARDLINKS to
   the RESOLVED system runtime targets (readlink -f first — same bytes, same
   link resolution, the wails build + the gate leg the proof); the committed
   recipe stays the symlink form, Am.4 is the fallback record; the continuation
   recipe is PROVEN at all levels: git status + git log --oneline -5 + file
   timestamps tell what ACTUALLY survived; a preserved commit or in-flight
   draft is continued and AUDITED — never blindly trusted, never restarted
   from zero; a resumed context re-verifies from scratch (the full battery +
   the hygiene scans) before committing; a wipe event may flip tracked files'
   MODES 644->755 while preserving content — diagnose via 'git diff --raw' and
   restore via chmod off the parsed list, NEVER git checkout; and the standing
   addendum: after every `wails build` the wailsjs bindings re-flip to 755,
   and a stray `go build .` in webui/ leaves a 22 MB binary named `webui` —
   never commit it). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (symlinked /home/z/second-brain-project),
   read-only reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked
   /home/z/GitCurator; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
   ecbadb99091a3f46e31f5f934b068b1864eafa7995211b39eaddf76996045fe5 — 64 hex
   chars, verified against go.dev's published JSON), venv /home/z/venvs/gitcurator
   (Python 3.12, stdlib-only) per build/TOOLCHAIN.md, the wails CLI v2.16.0
   (GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0
   — build-time only), and the rootless webkit tree (~/.local/lib-wails per
   webui/README.md + docs/poc-artifacts/t-70-gate.md) per build/TOOLCHAIN.md
   Am.3 + Am.4. The MAIN go.mod: modernc.org/sqlite at v1.36.1, bogdanfinn/tls-client
   at v1.16.0, golang.org/x/net v0.48.0 + golang.org/x/text v0.32.0 direct, and
   the go directive at 1.24.1 (the tls-client-forced floor — ADR-0002 Am.2; do
   not bump any of these). The NESTED webui module rides go 1.25.0 (wails
   v2.16.0's floor) + the wails require + the transitive requires the direct
   internal/app import pulls (indirect via the parent replace — ADR-0002 Am.5);
   the MAIN go.mod stays byte-unchanged. Then check `git status` and
   `git log --oneline -3`: SESSION-037 landed T-74 COMPLETE in three slices
   (a4619fe — internal/app/batchwire.go THE CLI-PARITY SHELL WIRING, the debt
   T-73 carried closed: engine.go's runImpl composition re-hosted on the app
   surface over six new seams — the warmup family, the inputs fetchers, the
   full intake, the banishment + settled ledger + backfill + the §4.4 passes,
   THE WEBSITES PHASE, THE FINISH CHAIN's seals + THE PHASE-5 CLEAR + the
   dry-run law, with BatchOptions' new Mode + ImportFile; 19 wiring tests;
   6d479fb — internal/app/settingsb.go the pages' engines (THE QUEUE REPORT
   RENDERING from canned QueueReports + the zinc ⏳/mint ✅ badge + the
   dashboard engines + the undo engine + the sources fetch over its seam +
   the backup ladders + the save groups; 21 tests; 3091d76 — webui/
   settingsuib.go the third bound struct + frontend/settings.js's four pages
   at the reference's own sidebar order (Dashboard after Input, Bot, Sources,
   Backup — every label verbatim, the Restore button honestly disabled with
   its landing note, the SB_POC_OPEN_SETTINGS section probe; 5 shell tests)
   and PUSHED everything after every task — 2060 main-battery tests race-clean
   over 35 packages + 28 webui tests over 4. Verify git status shows up to date
   with origin/main (the tip at or beyond 3091d76). Use the PAT only
   ephemerally: git -c http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety.

2. Read, in order: the repo README.md (Phase 6 OPEN — T-74 at 3091d76) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-037's close: T-74 COMPLETE, the newest
   laws — THE SYMLINK-BLOCK FALLBACK, THE RENDER-CONTEXT LAW, the
   ONE-SLICE-PROPERLY record, the LIVE-CONFIG law, the GROUP-POINTER law) →
   docs/sessions/SESSION-037.md (the most recent report) → the T-75 spec
   (spec-gui-sync-views.md — THE AUTHORITY: the v0.87.0 GRAND CONVERGENCE
   page, ONE Sync page two desks — the four-element flow above the
   stage-by-stage gap matrix, one Refresh measures both) → internal/app/
   batchwire.go (the review_retry/master_retry engines the page wires) +
   commands.go + queue.go → webui/settingsuib.go + frontend/settings.js
   (the established page-binding + builder patterns) → the reference's
   gui/main_window/sync_view.py + sync_flow.py + sync_display.py +
   core/sync_review.py.

3. THIS SESSION IS T-75 (L): SYNC VIEW per doc 13's row — the flow cards +
   the matrix + the differences modal (spec-gui-sync-views): THE VISUAL
   SYNC DISPLAY (🤖 bot → 📓 vault → ☁️ GitHub → 🌐 the site — read-only,
   refreshed on demand, the click is the consent for every network leg)
   above THE DEFERRED LEDGER (the stage-by-stage gap matrix — what is
   waiting, every gap fixed or explained) + the differences modal. The
   page joins the Settings window's SECTIONS at the reference's own order
   (between Dashboard and Bot; the 'refresh' glyph is in the pack). The
   engine legs already live: batchwire.go's retry modes,
   Commands.SyncFetchQueue, the drift-seal family, the queue check's
   classification counts. ACCEPTANCE: the four-state rendering tests +
   the deferred-ledger UX tests. If capacity runs short: one slice done
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
   network-touching seam in tests at construction (the HERMETIC-SEAM law);
   after every wails build, check `git diff --raw` for mode flips and never
   commit a stray webui/webui binary; every git/go command carries its own cd
   (the sandbox shell resets cwd between calls).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-75's
   own acceptance: the four-state rendering tests + the deferred-ledger UX
   tests), commit + push everything (ask me for the GitHub PAT the moment a
   push is due — never store it anywhere; never end the session push-pending),
   write docs/sessions/SESSION-038.md, rewrite docs/NEXT-SESSION.md with the
   next session's prompt, and end your reply with a plain-language report plus
   the prompt I should copy to start SESSION-039.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

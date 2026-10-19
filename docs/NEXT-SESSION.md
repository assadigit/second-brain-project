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
- **Phase 6 OPEN (SESSION-033–038)** — T-70 COMPLETE (the entry gate),
  T-71 COMPLETE (`internal/app`, SESSION-034), T-72 COMPLETE (the
  frontend foundation, SESSION-035), T-73 COMPLETE (the Settings
  pages A, SESSION-036), T-74 COMPLETE (the Settings pages B + the
  CLI-parity shell wiring, SESSION-037), and now **T-75 COMPLETE:
  THE SYNC VIEW** (SESSION-038, three slices: `6eea74d` —
  internal/syncreview, core/sync_review.py PORTED — the placeholder
  seat T-71 recorded finally filled: the four-state model + the
  named verdict + the micro-copy captions + the DeferredLog over
  the SQLITE LAW + the three read-only doors, 101 tests;
  `04768e7` — internal/app/syncdisplay.go + syncview.go: THE VISUAL
  SYNC DISPLAY's pure model (the four cards verbatim, the §5
  arithmetic's every edge) + the gather + the Settings doors
  (RefreshSyncView — ONE payload, both desks) , 50 tests;
  `478101c` — the shell: the FOURTH bound struct + the Sync page at
  the reference's own sidebar order (the 2×2 flow grid, the
  four-state matrix, the differences modal with the disabled-until
  law, the first-show door, the fix doors), 5 shell tests). **The
  acceptance GREEN: the four-state rendering tests + the
  deferred-ledger UX tests — 2211 main-battery tests race-clean
  over 36 packages (2060 + 101 + 50) + 33 webui tests over 4
  (28 + 5).** The gate leg GREEN over the 6th webkit rebuild (first
  frame 292/730 ms ≤ 1.5 s; the 10k flood 588,235 lines/s, 0
  drops); the screenshots VLM-verified (the four cards' honest
  not-set grammar; the live measured-at line; the dashed unknown
  matrix). The 33rd full wipe was absorbed before work (see the
  session report — Am.4's hardlink fallback again, symlink(2)
  still blocked).
- **Push state: CLEAN** — SESSION-038 pushed after every slice;
  verify `git status` is in sync at the start (the tip at or beyond
  `478101c`).
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
  resolve THAT against the system tree), all per build/TOOLCHAIN.md
  Am.3 + Am.4; the MAIN go.mod exactly modernc.org/sqlite v1.36.1 +
  bogdanfinn/tls-client v1.16.0 + golang.org/x/net v0.48.0 +
  golang.org/x/text v0.32.0 + the go directive at 1.24.1 (never
  bump; ADR-0002 Am.2). The NESTED webui module rides go 1.25.0
  (wails v2.16.0's floor) + the wails require + the transitive
  requires the direct internal/app import pulls (indirect via the
  parent replace — ADR-0002 Am.5; the MAIN go.mod stays
  byte-unchanged).
- **The operational laws, newest first** (each was paid for in
  evidence — carry ALL of them):
  - **THE FIELD-TAG LAW** (SESSION-038): the payload structs' JSON
    tags mirror the Python dict keys verbatim (snake_case:
    `dot_tone`, `display_number`, `measured_at`…) — the frontend
    reads the TAGGED names, never the camelCase guesses; the first
    render's stuck "Not measured yet" note was this bug exactly.
  - **THE DISPLAY-PIPELINE CORRUPTION LESSON** (SESSION-038): this
    sandbox's tool-output rendering occasionally drops the two
    characters `[h` (the sha pin's lost character in chat prompts,
    `[hidden]` selectors reading as `idden]` are the same family) —
    settle every suspect reading with byte-level counts/hex before
    "fixing" anything.
  - **THE SYMLINK-BLOCK FALLBACK** (SESSION-037, TOOLCHAIN.md Am.4):
    a wipe rebuild on a symlink-blocking sandbox relinks the 26 dev
    .so links as hardlinks to the RESOLVED system targets (readlink
    -f the LINK'S TARGET against the system tree); the wails build
    over the tree is the proof.
  - **THE RENDER-CONTEXT LAW** (SESSION-037): a pure renderer that
    reads config context takes the CONFIG as an argument.
  - **THE ONE-SLICE-PROPERLY RECORD** (SESSION-036, paid again in
    037 and 038) — the wiring slice lands complete with its own
    proof before the pages.
  - **THE LIVE-CONFIG LAW** (SESSION-036); **THE GROUP-POINTER LAW**
    (SESSION-036); **THE DISSOLUTION-PROBE RECORD** (SESSION-035,
    ADR-0002 Am.5); **THE GOLDEN-MARSHAL-ORDER TRAP** (SESSION-035);
    **THE SUITE-REVERIFICATION DIVIDEND** (SESSION-035); **THE
    WAILS-BUILD MODE FLIPS** (SESSION-035: every `wails build`
    re-flips the wailsjs bindings to 755; a stray `go build .` in
    webui/ leaves a 22 MB `webui` binary — check `git diff --raw`
    after every build, restore via chmod, never commit the binary);
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
    cd — the sandbox's shell resets cwd between calls; the NESTED
    webui module needs its own cd; never commit outside the work
    repo; never checkout blindly); **THE RENDERER-EATS-ESCAPE-PAIRS
    lesson** (commit messages via -F FILE); **THE WRITE-TOOL
    REDACTION lesson** (byte-level secret-scan before every commit);
    **THE BACKGROUND-PROCESS lesson** (long suites in FOREGROUND
    CHUNKS); **PUSH AFTER EVERY TASK** (ask for the PAT the moment
    a push is due; never store it).

## 2. Next session: SESSION-039 — T-76 (M): the Scan flow — the picker → the plan modal (the masters + the arm-guard) → the scorecard

- **Verify the push state first** (§1): confirm the tip is at or
  beyond `478101c`.
- **T-76 per doc 13's row**: the Scan flow's three UI legs over the
  engine that already lives (T-71's scanjob.go — the _vault_scan_job
  port: the one-open cache, the scope picker's four miss variants,
  BOTH confirm doors, the picking arithmetic, the merged verdict):
  1. **THE VAULT PICKER** (vault_pick_dialog.py — its PURE model
     halves are ALREADY PORTED at T-75: `DetectScanVaults` lives in
     internal/app/syncdisplay.go, rows + statuses + the Personal
     read-only law; `pick_display_model` still to port);
  2. **THE PLAN MODAL** (scan_plan_dialog.py — the masters
     (select-all row states), the arm-guard, the plan list);
  3. **THE SCORECARD** (scan_result_dialog.py).
  ACCEPTANCE: workflow tests; selection semantics.
- The Settings window is the established host (the Scan flow is a
  MAIN-WINDOW flow in the reference — read spec-gui-dialogs +
  spec-gui-pages-a's vault_scan_ui section for the real host; the
  shell's Scan button + hero ride T-77, the modal may ride its own
  window or the settings overlay per the spec's own words).

## 3. Read-first list (in order)

1. `docs/sessions/SESSION-038.md` — the most recent report (T-75
   COMPLETE; the placeholder filled; the field-tag law).
2. `docs/port-blueprint/module-specs/spec-gui-dialogs.md` — THE
   AUTHORITY for the plan modal + the picker + the scorecard.
3. `docs/port-blueprint/module-specs/spec-gui-pages-a.md` — the
   vault_scan_ui section (the Scan button's own flow).
4. `internal/app/scanjob.go` + `scanjob_test.go` — the engine that
   already lives (T-71's port: the confirm doors, the scope picker,
   the picking arithmetic, the merged verdict).
5. `internal/app/syncdisplay.go` — `DetectScanVaults` (the picker's
   first pure half, already ported at T-75).
6. The reference's `gui/vault_pick_dialog.py` (the dialog + the
   second pure half) + `gui/scan_plan_dialog.py` +
   `gui/scan_result_dialog.py` (read at /home/z/my-project/GitCurator
   — the symlink is down while symlink(2) is blocked).
7. `docs/port-blueprint/13-implementation-task-list.md` — the T-76
   row + Phase 6's book.

## 4. Known constraints / owner actions for SESSION-039

- **Push state: CLEAN** — keep the PUSH-AFTER-EVERY-TASK law; ask for
  the PAT the moment a push is due (never stored, never written).
- The dissolution call (the owner's, ADR-0002 Am.5): (a) amend the
  authorized set vs (b) keep the nested boundary — no rush; T-76
  proceeds on (b) either way.
- The vanilla-vs-vite call (the owner's): vanilla is landed and
  proven; a flip needs only an ADR + the build step.
- **The authorized runtime set is UNCHANGED** (the four; the MAIN
  go.mod byte-unchanged — the nested webui module carries wails +
  the indirect transitive requires per Am.5).
- **The go-directive acknowledgement (STILL WAITING — twenty-ninth
  session carrying it)**: the main module rides `go 1.24.1`, forced
  by the owner-acknowledged tls-client v1.16.0 pin (ADR-0002 Am.2).
- The x/net/html question is a KEEP-OR-SWAP, not a blocker.
- **After every `wails build`**: check `git diff --raw` for mode flips
  (the bindings re-flip to 755) and never commit a stray `webui/webui`
  binary. **A stray `go build .` in webui/ LEAVES the 22 MB binary —
  remove it before staging.**
- **If the sandbox blocks symlink(2) during a wipe rebuild**:
  TOOLCHAIN.md Am.4's hardlink fallback for the 26 dev links.
- **The field-tag law** (SESSION-038): new payload structs' JSON tags
  mirror the Python dict keys; the frontend reads the tagged names.
- **Deferred test debt on the books** (evidence-recorded, not
  blockers): test_noteredo's remaining un-ported selections; the
  note_state SURGERY path's CRLF round-trip differential (Phase 4+);
  the Windows live legs + the launchers' clean-VM smoke + the
  bootstrapper leg (Phase 7 checklist, doc 10 §7); the sound WAV
  assets (not yet committed — the preview's missing-asset notice is
  the parity behavior until T-77/T-80 ship the files); the Backup
  page's Restore button (honestly disabled — the safety-net restore
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

## 5. THE PROMPT TO COPY (SESSION-039)

```
You are starting SESSION-039 of the second-brain-project — the multi-session Go
port of the GitCurator Python app (a personal second-brain library builder; full
specs exist). You are the implementation session; a separate overseer chat
(planner/mentor) records owner decisions and audits — do not expect it to write
code alongside you. You are the single writer on the repo.

1. The environment is REBUILDABLE, not persistent (thirty-three wipe events
   across SESSION-005..038 — SESSION-038 absorbed the thirty-third full wipe
   BEFORE work: both repos, the toolchain, the venv, the module cache, AND the
   rootless webkit tree all gone, everything rebuilt from the committed recipes
   — the sixth webkit rebuild, Am.4's hardlink fallback again: symlink(2) is
   STILL blocked, the 26 dev .so links land as hardlinks to the RESOLVED
   system targets (readlink -f the LINK'S TARGET — the versioned .so — against
   the system tree; the wails build + the gate leg the proof); NOTE the new
   lesson: this sandbox's tool-output rendering occasionally drops the two
   characters '[h' — the sha pin's lost character in chat prompts and
   '[hidden]' selectors reading as 'idden]' are the same family — settle every
   suspect reading with byte-level counts/hex before "fixing" anything; the
   continuation recipe is PROVEN at all levels: git status + git log --oneline
   -5 + file timestamps tell what ACTUALLY survived; a preserved commit or
   in-flight draft is continued and AUDITED — never blindly trusted, never
   restarted from zero; a resumed context re-verifies from scratch (the full
   battery + the hygiene scans) before committing; a wipe event may flip
   tracked files' MODES 644->755 while preserving content — diagnose via 'git
   diff --raw' and restore via chmod off the parsed list, NEVER git checkout;
   and the standing addendum: after every `wails build` the wailsjs bindings
   re-flip to 755, and a stray `go build .` in webui/ leaves a 22 MB binary
   named `webui` — never commit it). FIRST verify with evidence: work repo at
   /home/z/my-project/second-brain-project (the /home/z/second-brain-project
   symlink is DOWN while symlink(2) is blocked — use the real path), read-only
   reference /home/z/my-project/GitCurator @ 98f4fc7 (symlinked /home/z/GitCurator
   — same caveat; must match testdiff/PINNED-GITCURATOR-REV), toolchain
   go1.27.2 -> /home/z/.local/go (go.dev tarball, sha256
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
   `git log --oneline -3`: SESSION-038 landed T-75 COMPLETE in three slices
   (6eea74d — internal/syncreview core/sync_review.py PORTED, the placeholder
   seat T-71 recorded finally filled: the four-state model + the named
   verdict + the micro-copy captions + the DeferredLog over the SQLITE LAW +
   the three read-only doors, 101 tests; 04768e7 — internal/app/syncdisplay.go
   + syncview.go: THE VISUAL SYNC DISPLAY's pure model + the gather + the
   Settings doors (ONE Refresh payload for both desks), 50 tests; 478101c —
   webui/syncuib.go the fourth bound struct + frontend/settings.js's Sync
   page at the reference's own sidebar order (the 2×2 flow grid, the
   four-state matrix, the differences modal with the disabled-until law, the
   first-show door, the fix doors), 5 shell tests) and PUSHED everything after
   every slice — 2211 main-battery tests race-clean over 36 packages + 33
   webui tests over 4. Verify git status shows up to date with origin/main
   (the tip at or beyond 478101c). Use the PAT only ephemerally: git -c
   http.extraheader=
   "Authorization: Basic $(printf 'x-access-token:%s' "$PAT" | base64)" and
   never store it — AND SCRUB IT if you ever CLONE with it (the
   GIT_CONFIG_COUNT env-var form avoids the persistence entirely).
   Re-verify the suite before work if anything looks off — the law pays in
   bugs, not just safety.

2. Read, in order: the repo README.md (Phase 6 OPEN — T-75 at 478101c) →
   docs/SESSION-PROTOCOL.md (the working law) → docs/NEXT-SESSION.md (your
   state card — §1 records SESSION-038's close: T-75 COMPLETE, the newest
   laws — THE FIELD-TAG LAW, THE DISPLAY-PIPELINE CORRUPTION LESSON, the
   SYMLINK-BLOCK FALLBACK, the RENDER-CONTEXT LAW) → docs/sessions/
   SESSION-038.md (the most recent report) → the T-76 spec
   (spec-gui-dialogs.md — THE AUTHORITY: the vault picker + the plan modal
   (the masters + the arm-guard) + the scorecard) + spec-gui-pages-a.md's
   vault_scan_ui section → internal/app/scanjob.go (the engine that already
   lives — T-71's _vault_scan_job port) + internal/app/syncdisplay.go (the
   picker's first pure half: DetectScanVaults, already ported) → webui/
   settingsui*.go + frontend/settings.js (the established binding + builder
   patterns; THE FIELD-TAG LAW) → the reference's gui/vault_pick_dialog.py +
   gui/scan_plan_dialog.py + gui/scan_result_dialog.py.

3. THIS SESSION IS T-76 (M): THE SCAN FLOW per doc 13's row — the picker →
   the plan modal (the masters + the arm-guard) → the scorecard
   (spec-gui-dialogs + spec-gui-pages-a's vault_scan_ui): the three UI legs
   over the engine that already lives (scanjob.go's confirm doors + the
   scope picker + the picking arithmetic + the merged verdict). The picker's
   first pure half (DetectScanVaults) is already ported at T-75 — build on
   it, never re-invent. ACCEPTANCE: workflow tests; selection semantics. If
   capacity runs short: one slice done properly with proof beats two
   half-done; close cleanly at a natural boundary.

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
   frontend reads the tagged names (THE FIELD-TAG LAW); after every wails
   build, check `git diff --raw` for mode flips and never commit a stray
   webui/webui binary; every git/go command carries its own cd (the sandbox
   shell resets cwd between calls; the NESTED webui module needs its own cd).

5. Close the session per SESSION-PROTOCOL.md: honest gate evaluation (T-76's
   own acceptance: workflow tests + selection semantics), commit + push
   everything (ask me for the GitHub PAT the moment a push is due — never
   store it anywhere; never end the session push-pending), write
   docs/sessions/SESSION-039.md, rewrite docs/NEXT-SESSION.md with the next
   session's prompt, and end your reply with a plain-language report plus
   the prompt I should copy to start SESSION-040.

One slice done properly with proof beats two half-done; close cleanly at a
natural boundary when capacity runs out.
```

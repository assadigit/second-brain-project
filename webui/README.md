# webui/ — the Wails GUI shell (Phase 6)

The Wails frontend that replicates the documented PyQt6 design lives here
from Phase 6 onward (ADR-0001: workflow-preserving HTML replica; no visual
redesign — `docs/port-blueprint/07-ui-and-performance-assessment.md` Part D
marks the boundary).

## What lives here now (T-72, the frontend foundation)

The **production shell** — the themed window over the production
composition, binding `internal/app` DIRECTLY (the `guibridge/` interim
seam RETIRED at this landing, deleted with its tests):

- `main.go` + `shell.go` — the wails shell: the golden-locked token
  tables drive the window (uitheme.RGBA), the typed Bus fans the
  engines' events to this shell's ONE subscriber (LogLine → the
  deskbus's coalescing push; progress/status/batch events cross to the
  frontend as wails events), the startup passes (sentinel, llama
  autodetect, quiet heal, backlog, radar) speak their real lines onto
  the desk through the same transport, the single-instance app.lock is
  taken before the window (the alive-PID refusal exits 1 — the
  AppLock.Refused distinction added at T-72; every other lock failure
  CONTINUES WITHOUT the lock), and the shutdown drains the desk then
  runs the graces.
- `settingsui.go` — T-73's Settings window binding wrapper: the thin
  shell over internal/app's Settings surface (the six pages' engines)
  plus the native pickers (the import-file + vault-folder dialogs).
  Bound as the second wails struct; the generated bindings
  live in frontend/wailsjs/go/main/SettingsUI.js.
- `syncuib.go` — T-75's Sync page binding wrapper (the FOURTH bound
  struct): the ONE Refresh door (one payload, both desks — the flow
  cards + the stage matrix), the differences modal's doors (the
  stage data, Mark as intentional + Reopen with the recompute, the
  deferred-log reads), the fix doors (the queue-check leg that lands
  the session's queue state, SealNow, PublishNow), and the per-row
  Open note door.
- `scanui.go` — T-76's Scan flow binding wrapper (the FIFTH bound
  struct): the vault picker's data door, StartScan THE CTA's whole
  flow (the guards → the quiet cancel → the start line → the blocking
  job over the bus gate → the finished routing + the scorecard
  event), ScanAnswer the held gate's answer door, and the two
  evidence doors (the canned ask + the canned scorecard). The three
  modals' renderers live in `frontend/scanflow.js` (the 350 ms
  fresh-click door, the master boxes, the 12 s self-close; the Scan
  CTA's own machine — NOT the hero's).
- `heroui.go` — T-77's main window chrome binding wrapper (the SIXTH
  bound struct): HeroClick THE hero button's one click door (the
  guards → the boxes, the offline ask → HeroAnswerOffline, the >10
  confirm → HeroRunQueue, the fetch leg over SyncFetchQueue with the
  caught-up extras (the waiting-rows scan + the drift seal + the
  closing line), the process/import/stop legs), the 200 ms mirror
  loop, TestConnection (the battery door over
  connectioncheck.RunLocalChecks with the live cc:* events), and the
  More menu's live doors (the dashboard URL law, DarkMode, the retry
  batch modes, ChromeBoot). Bound as the sixth struct; the generated
  bindings live in frontend/wailsjs/go/main/HeroUI.js.
- `settingsuib.go` — T-74's pages B binding wrapper: the thin shell
  over internal/app's pages B engines + the Commands surface (the
  queue check + render + badge, Process All, the dashboard refresh +
  quarantine + undo, the sources fetch + process, the backup saves +
  ladders + job buttons + the FastSync door T-77's hero machine
  wires). Bound as the third wails struct; the generated bindings
  live in frontend/wailsjs/go/main/PagesBUI.js.
- `deskbus/` — THE COALESCED LOG's Go half (pure, headless-tested): the
  120 ms flush beat, the ≤250-row batches, the cap-1024 backpressure
  channel, the synchronous clear door, and (T-72's fix) the
  deterministic post-Stop refusal.
- `uitheme/` — the theme-token + icon-pack twins, golden-locked against
  the reference (testdiff/goldens/dump_theme.py + dump_icons.py →
  testdata/*.golden.json): the Go tables (76 LIGHT + 76 DARK tokens),
  the 28-glyph lucide-static pack, and the tests that also lock the
  CSS twin (frontend/theme.css) and the JS twin (frontend/icons.js)
  against the same goldens.
- `frontend/` — vanilla HTML/CSS/JS (no npm step; `wails.json`'s
  install/build commands are empty): `theme.css` (the FULL token
  tables, LIGHT + the `[data-theme="dark"]` plum night), `style.css`
  (the §2 type-scale + component classes + T-73's Settings-window
  laws), `icons.js` (the pack, tinted
  via currentColor; the reference's _LOG_LEVEL_ICONS law), the
  cache-mirror desk (cap 1000, filter re-render from cache, per-row
  tooltip), the layout shell (the sync_card CTA band — the 56px SYNC
  hero + Fast Sync + Scan + Test Connectivity, honestly disabled until
  T-77's hero button machine), the flood probe instruments, and
  `settings.js` (T-73: the Settings window's frontend half — the
  pages A rendered from the snapshot, the save-on-change law,
  the mini-dialog kit; T-74: the pages B — Dashboard, Bot, Sources,
  Backup — at the reference's own sidebar order with every label
  verbatim, the zinc/mint badge, the >10 confirm gates, the
  section-selecting evidence probe; the desk keeps living underneath).
- `frontend/heroflow.js` — T-77: the chrome's frontend half — the
  hero:state renderings (the label, the glyph, the disabled flag, the
  danger fill, the tooltips), the click door + the guard boxes + the
  offline question + the confirm gate, the progress row's live formats
  + the terminal marks + THE BATCH CHIME (<audio> — the QSoundEffect
  twin), ⚡ Fast Sync's busy render, the sentinel labels' paint (the
  dot + the word + the tooltip + the warn-now box), THE BATTERY MODAL
  (the four rows, the spinner, Start Syncing gated on all-ok), and
  THE MORE MENU (the reference's own action list verbatim — the live
  doors wired, the not-yet-landed engines honestly disabled with their
  landing notes; the Settings submenu carries Dark Mode).
- `fixture/` — the deterministic synthetic vault (no owner data).
- `build-support/` — the rootless-webkit evidence recipe: the LD_PRELOAD
  path shim + `run-gate.sh` (the Xvfb autorun leg — the log perf test's
  end-to-end evidence).

## The module laws (ADR-0002 Am.3 + Am.4 + Am.5)

- This is a **nested Go module**
  (`module github.com/assadigit/second-brain-project/webui`) requiring
  `github.com/wailsapp/wails/v2 v2.16.0` with a local `replace` to the
  parent — so the MAIN module's runtime set stays exactly the authorized
  four. Build with the pinned wails CLI (v2.16.0 — see
  `build/TOOLCHAIN.md` Am.3 for the full record).
- The nested module CAN import the parent's `internal/` packages (Am.4's
  boundary-probe correction) — and at T-72 it DOES: `shell.go` binds
  `internal/app` directly (the guibridge seam retired). The nested
  module's go.mod gained the transitive requires that import pulls
  (indirect, via the parent replace); the MAIN module's go.mod is
  byte-unchanged.
- **The T-72 dissolution probe (Am.5)**: importing wails into the MAIN
  module forces x/net v0.56.0 + x/text v0.39.0 + the 1.25.0 directive —
  bumps the authorized-pins law forbids. The boundary STAYS nested; the
  owner's call (amend the set vs keep the boundary) is recorded there.

## Build & run (the committed recipe)

```sh
# Linux dev leg (CGO + the rootless webkit tree — ADR-0002 Am.3's CGO story):
PKG_CONFIG_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu/pkgconfig \
LD_LIBRARY_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu \
wails build -tags webkit2_41

# the headless evidence run (Xvfb + autorun + the results JSON + screenshot):
cd build-support && ./run-gate.sh

# the Windows production leg (CGO-free cross-compile, pure-Go WebView2 loader):
wails build -platform windows/amd64
```

The T-70 gate's verdict + measurements: `docs/poc-artifacts/t-70-gate.md`.
The T-72 perf evidence (the log perf test through the production
transport): SESSION-035's report + `t70-gate-results.json` shapes.

The build output (`build/bin/`) never enters git; the exe icon is the wails
default — the brand-mark swap (`assets/icons/gitcurator.ico`) is the
packaging phase's (T-80) item.

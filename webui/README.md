# webui/ — the Wails GUI shell (Phase 6)

The Wails frontend that replicates the documented PyQt6 design lives here
from Phase 6 onward (ADR-0001: workflow-preserving HTML replica; no visual
redesign — `docs/port-blueprint/07-ui-and-performance-assessment.md` Part D
marks the boundary).

## What lives here now (T-70, the PoC C GUI-shell gate)

The **T-70 gate shell** — the themed window + the coalesced log desk + ONE
REAL JOB wired through (the note-state baseline over a generated fixture
vault, via the `guibridge/` seam in the parent module):

- `main.go` + `app.go` + `job.go` — the wails shell: the LIGHT-token window
  (1000×750, cream `BackgroundColour`), the desk bus wiring, the cold-start
  clock, the real job, the 10k flood leg, the honest measurement file.
- `deskbus/` — THE COALESCED LOG's Go half (pure, headless-tested): the
  120 ms flush beat, the ≤250-row batches, the cap-1024 backpressure
  channel, the synchronous clear door.
- `fixture/` — the deterministic synthetic vault (no owner data).
- `frontend/` — vanilla HTML/CSS/JS (no npm step; `wails.json`'s
  install/build commands are empty): the spec-gui-theme-icons LIGHT token
  table as CSS vars, the btn_kind grammar (one filled primary), the
  SegmentBar twin, the chips + search + clear doors, the cache-mirror desk
  (cap 1000, filter re-render from cache, per-row tooltip), the four
  lucide-static glyphs, and the flood probe (handler-time + task-gap).
- `build-support/` — the rootless-webkit evidence recipe: the LD_PRELOAD
  path shim (committed this time) + `run-gate.sh` (the Xvfb autorun leg).

## The module laws (ADR-0002 Am.3 + Am.4)

- This is a **nested Go module**
  (`module github.com/assadigit/second-brain-project/webui`) requiring
  `github.com/wailsapp/wails/v2 v2.16.0` with a local `replace` to the
  parent — so the MAIN module's runtime set stays exactly the authorized
  four. Build with the pinned wails CLI (v2.16.0 — see
  `build/TOOLCHAIN.md` Am.3 for the full record).
- The nested module CAN import the parent's `internal/` packages (Am.4's
  boundary-probe correction) — the `guibridge/` seam is kept for its
  architectural value (doc 04 §2's event-seam shape, main-battery tested),
  not as a forced boundary.
- **THE RETIREMENT LAW**: this module dissolves at T-72, when the
  production shell lands in the main module (`webui/bindings.go` importing
  `internal/app` directly; the wails require enters the main `go.mod` via
  its own ADR-0002 amendment) and `guibridge/` retires with it.

## Build & run (the gate's own recipe)

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

The gate's verdict + measurements: `docs/poc-artifacts/t-70-gate.md` (the
artifacts: `t-70-gate-results.json`, `t-70-gate-screenshot.png`).

The build output (`build/bin/`) never enters git; the exe icon is the wails
default — the brand-mark swap (`assets/icons/gitcurator.ico`) is the
packaging phase's (T-80) item.

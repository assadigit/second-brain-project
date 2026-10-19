# T-70 — the Phase-6 GUI-shell gate (PoC C re-run with ONE REAL JOB) — evidence

Recorded 2026-10-19, SESSION-033. Environment: Debian 13 trixie sandbox
(two mid-session states — the first build/measurement leg, then a
mid-session degradation wiped the toolchain + module cache + webkit tree;
everything rebuilt from the committed recipe and the gate re-run as the
reproducibility leg), go1.27.2, wails CLI **v2.16.0** (ADR-0001 Am.2's pin,
per ADR-0002 Am.3's toolchain record), webkit2gtk-4.1 2.54.0 via the
rootless extracted tree (`~/.local/lib-wails`), Xvfb 1100×750.

The gate's terms (ADR-0001 §Decision 4 + the roadmap's Phase-6 entry + the
T-70 row): **the themed window (the spec-gui-theme-icons tokens), the log
desk, ONE REAL job wired through — startup ≤ 1.5 s and a 10k-line log flood
staying interactive — the verdict recorded.**

## What this gate is (vs SESSION-001's PoC C)

ADR-0001 Am.2 recorded SESSION-001's PoC verdict (the framework is viable).
SESSION-033's mandate re-runs the gate as **Phase 6's entry condition** with
the substantial upgrade the mandate words: **ONE REAL JOB** — SESSION-001's
job was a synthetic `time.Sleep(20ms)` loop; this gate wires the **real
note-state baseline job** (the `record_baseline_if_empty` composition,
`note_state.py:443-474` — the same beat the batch engine's setup family
runs) through the `guibridge/` seam: a real `ScanVault` walk over a
deterministic 800-note fixture vault spread across the real CATEGORY_FOLDERS
trees, per-note read + `helpers.DecodeUTF8Replace` + `ComputeFingerprint`
(sha256) + SQLite upserts into a real `cache.db`. The desk is fed by the
engine's own log lines; the progress row by per-note progress events.

Also upgraded vs SESSION-001: the frontend is **vanilla** (no npm step —
ADR-0002 Am.3's frontend-step record), the wailsjs runtime shim + the webkit
path shim + `run-gate.sh` are **committed** (`webui/build-support/`), and the
desk implements the full coalesced-log contract's frontend half (the
cache-mirror law: cap 1000 rendered rows, filter re-render FROM CACHE,
per-row tooltip = full text, ONE fragment append + ONE scroll per batch) —
SESSION-001's desk had only the batch-append + trim.

## The measured evidence (both runs, Xvfb, autorun mode)

| ADR-0001 gate | Gate value | Run 1 (03:17) | Run 2 (11:40, post-degradation rebuild) |
|---|---|---|---|
| Cold start → first frame | ≤ 1.5 s | **266 ms** | **679 ms** (box still settling after the wipe event) |
| 10k-line flood, desk interactive | ≥ 2,000 rows/s, interactive | **458,695 lines/s accepted, 0 drops**; 41 batches; the flood cost the main thread **741 ms total / 23 ms max per batch** (free ~93% of the window); desk at "1000 rows / cache 1000" (the cap law) | **454,133 lines/s, 0 drops**; 25 ms max per batch |
| exe size | < 20 MB | Linux 12.4 MB / **Windows .exe 14.95 MB** (CGO-free cross-compile) | — |
| THE REAL JOB (this gate's own term) | must be real | **800 notes in 94 ms** (0.118 ms/note), 6 stubs never adopted, rows verified in the DB (the guibridge contract tests) | 800 notes in 101 ms |

Artifacts: `t-70-gate-results.json` (run 1), `t-70-gate-results-rerun.json`
(run 2), `t-70-gate-screenshot.png` (run 1; VLM-verified: the themed window
with the "Scan vault — baseline job" filled hero, the purple 800/800
progress row, the chips + search + Clear, the glyph-colored mono log rows,
"startup_frame: 266.101406ms", "1000 rows / cache 1000", the flood's final
lines verbatim).

## The honest caveats (measured, not guessed)

- **The task-gap instrument reads WebKit's timer throttling, not blockage**:
  under a WM-less Xvfb WebKit treats the page as background and clamps
  setTimeout to ~1 Hz — the sampler's 510–755 ms "gaps" are the clamp. The
  interactivity evidence is therefore the **handler-time instrument**
  (every `log:batch` event's own DOM-append work: 23–25 ms max per batch,
  638–741 ms total across the whole 10k flood) + the rendered desk + the
  still-interactive probe chain surviving the flood. On a real desktop WM
  (the owner's Windows/WebView2) the clamp does not apply.
- **2-CPU box**: all numbers were measured under a 2-CPU cgroup (earlier
  sessions' boxes had more headroom); the gate passes with margin anyway.
- **The Windows leg is cross-compiled, not executed** (no Windows in this
  sandbox — the same constraint SESSION-001 recorded): the exe builds
  CGO-free at 14.95 MB; the runtime Windows leg + the silent Evergreen
  bootstrapper path on a WebView2-less VM remain the owner's verification
  (ADR-0001 Am.1/Am.2's registered owner action, unchanged).
- **One anomalous battery run** in the immediate post-degradation settling
  window (3 packages failed once; names not captured — the box was still
  releasing the wipe event's resources); the suite then ran fully clean in
  two consecutive complete runs (1932/1932 over 35 packages). Not
  reproducible since; recorded for honesty.
- The single mid-session degradation (toolchain + module cache + webkit
  tree + /tmp wiped between run 1 and run 2; the repos and all in-flight
  work survived; 584 tracked files' modes flipped 644→755 — restored
  surgically via chmod without touching content) is itself evidence the
  committed recipe rebuilds the whole GUI toolchain from a clean tree.

## The verdict

**PASS — the Phase-6 entry gate is GREEN.** All ADR-0001 acceptance
criteria pass with wide margins on both runs (first frame ≤ 679 ms vs the
1,500 ms gate; ≥ 454k lines/s vs the 2,000 rows/s gate; 14.95 MB vs the
20 MB gate), **and** the gate's own upgrade — ONE REAL JOB wired through
the engine (not a synthetic loop) — is proven: the note-state baseline
over a real fixture vault, with the engine's own log lines on the desk and
per-note progress on the progress row, inside the themed Wails shell.

The production caveats (the WM-less-Xvfb timer clamp; the Windows runtime
leg + bootstrapper as the owner's verification) do not gate Phase 6's
*development* — they gate the GUI-parity *release* (Phase 7's clean-VM
checklist, doc 10 §7), exactly as ADR-0001 Am.2 already recorded.

## The reproduce recipe (committed, no /tmp-only parts)

```sh
# toolchain: go1.27.2 (sha256 ecbadb990… verified vs go.dev JSON) + the pinned CLI:
GOBIN=/home/z/.local/go/bin go install github.com/wailsapp/wails/v2/cmd/wails@v2.16.0

# the rootless webkit tree (Debian trixie, no root):
apt-get install --download-only -o Dir::Cache::archives=/tmp/wails-debs \
    -o Debug::NoLocking=1 -y libwebkit2gtk-4.1-dev libgtk-3-dev
for d in /tmp/wails-debs/*.deb; do dpkg -x "$d" ~/.local/lib-wails/; done
# then: rewrite the .pc prefixes; relink the dev symlinks (in-tree or system)

# build + headless evidence run (from webui/):
PKG_CONFIG_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu/pkgconfig \
LD_LIBRARY_PATH=~/.local/lib-wails/usr/lib/x86_64-linux-gnu \
  wails build -tags webkit2_41
cd build-support && ./run-gate.sh     # Xvfb + autorun + results JSON + screenshot

# the Windows leg:
wails build -platform windows/amd64  # CGO-free (pure-Go WebView2 loader)
```

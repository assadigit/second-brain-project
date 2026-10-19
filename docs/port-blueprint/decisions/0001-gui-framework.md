# ADR-0001 — GUI technology for the Go port

- **Status**: accepted (primary), with an explicit proof-of-concept gate before large-scale UI work (see ADR-0007 and the roadmap)
- **Date**: decided in the blueprint session (2026-10); repository evidence: GitCurator v0.91.0, HEAD 8fd4179
- **Deciders**: blueprint session (this document); final confirmation belongs to the implementation session's PoC

## Context

The Python app's GUI is **PyQt6** and is the single largest part of the codebase:

- ~25,600 LOC across `app/gitcurator/gui/` (123 production files total in the package; 41 are GUI files).
- A single `MainWindow` composed from **18 mixins** (`app/gitcurator/gui/main_window/window.py:93`), a
  **Settings master–detail dialog with 11 pages** (`gui/dialogs.py:73-101`), ~24 overflow-menu actions
  (`gui/main_window/ui.py:1849-1946`), 10+ specialized modal dialogs (scan plan, scan result, vault picker,
  connection test, differences modal, banish/LLM/disk-full/code modals).
- A rich but *fully documented* visual system: complete LIGHT/DARK token tables
  (`gui/theme.py:58-237`, reproduced in `07-ui-and-performance-assessment.md`), Lucide icon pack
  v0.544.0 (30 glyphs, `gui/icons.py:60-105`), Inter-Bold font asset, sounds, objectName-based QSS grammar.
- Heavy Qt-specific engineering to stay responsive: the coalesced log (120 ms flush, 250-row batches,
  `ui.py:65-86`), sentinel on a daemon thread (`window.py:544-594`), measured-height wordwrap hack
  (`gui/main_window/sync_flow.py:63-86`), 350 ms fresh-click input guard (`gui/scan_plan_dialog.py:768-872`).

The mission (`Port-investigation.md` §2, §7) requires: preserve workflow/functionality, **no visual
redesign in this phase**, Windows-first single-ish executable, honest assessment, and a documented
strategy rather than an implementation.

## Problem

Choose the GUI technology for the Go application that can (a) host the existing 11-page settings
workflow, dialogs, log desk, and theming without a *visual redesign*, (b) ship as a simple Windows
`.exe`, (c) keep future Linux/macOS feasible, and (d) actually fix the perceived sluggishness rather
than reproduce it.

## Options considered

| Option | Windows packaging | Workflow preservation | Assessment |
|---|---|---|---|
| **A. Wails v2/v3 (Go + OS WebView2)** | Single exe (~10–15 MB) + WebView2 runtime (preinstalled on Win10/11, bootstrappable) | Excellent — HTML/CSS reproduces the documented token tables, wordwrap, log view, Lucide SVGs, dark/light almost 1:1 | Primary recommendation. UI layer becomes TypeScript/HTML mirroring the documented design; all logic stays Go |
| **B. Fyne (pure Go, OpenGL)** | Single exe, zero runtime deps | Moderate — Material-styled widgets; the settings master-detail, rich colored log, wordwrapped cards and tooltips need re-approximation; visuals would drift (borderline "redesign") | Fallback if WebView2 is unacceptable |
| **C. Gio (pure Go, low-level)** | Single exe | Poor — no widget toolkit at this level of fidelity; months of UI framework work | Rejected |
| **D. Qt bindings for Go (therecipe/qt)** | Heavy CGO, Qt DLL deployment, project effectively unmaintained | Best visual fidelity in theory | Rejected: build complexity and abandonment risk defeat the "reliable Windows exe" goal |
| **E. GTK3/4 bindings (gotk3/gotk4)** | MSYS2 DLL soup on Windows | Moderate | Rejected: Windows distribution is notoriously fragile |
| **F. Electron / Tauri-style separate frontend** | Large runtime; Tauri = Rust backend | Same as A but splits the language | Rejected: mission mandates Go |
| **G. Terminal-first (TUI) only** | trivial | Not equivalent — the owner is a designer who lives in the GUI | Rejected as the end state (kept as Phase-1 CLI parity, see ADR-0007) |

## Decision

1. **Primary: Wails (Go backend + WebView2 frontend).** The already-existing, near-complete design
   documentation (token tables, type scale, objectName laws, per-page layouts recorded in
   `07-ui-and-performance-assessment.md` and the module specs) makes an HTML/CSS *replica* of the
   current UI feasible without inventing a new design — which satisfies "preserve functionality now,
   redesign later".
2. **Fallback: Fyne**, chosen only if the PoC proves WebView2 unavailable/unacceptable on the owner's
   Windows 10 machine.
3. **The Go backend is GUI-agnostic.** All domain logic (pipelines, vault scan, seals, sync, cache)
   lives in pure Go packages with no GUI imports — exactly mirroring the Python `core/` purity law
   (`SPEC.md` §2.8). The GUI is a thin shell over the same command surface the CLI uses.
4. **A small PoC is mandatory before large-scale UI work** (roadmap Phase 0): a Wails window with the
   cream/plum theme tokens, the log desk fed by a coalescing event bus, and one worker job — measured
   for startup time and log throughput.

## Consequences

- **Advantages**: HTML/CSS gives the wordwrap, eliding, tooltips, tables, dark/light theming and the
  Lucide icon pack nearly for free; Windows webview rendering is GPU-accelerated and smooth; the Go
  side gains true parallelism (no GIL) for vault walks, hashing and SQLite.
- **Disadvantages / risks**: adds a TypeScript/HTML surface the repo did not have; WebView2 runtime
  dependency must be documented (installed by default on updated Windows 10/11; the installer can
  bootstrap it); Wails v3 was still pre-release at decision time — the implementation session must
  pin a stable Wails version (v2.x) and record the choice.
- **Explicitly out of scope**: any new screens, IA changes, or visual redesign. The HTML UI must
  replicate the documented layout of the PyQt6 UI (page names, controls, states, dialogs, honesty
  laws), not improve upon it.

## Evidence

- `Port-investigation.md` §2 (GUI goal), §7 (assessment criteria)
- `app/gitcurator/gui/` inventory and Qt hacks: investigation reports 1-d/1-e (2026-10), see
  `port-blueprint/07-ui-and-performance-assessment.md`
- Theme token tables: `app/gitcurator/gui/theme.py:58-237`
- Window/dialog geometry laws: main window 1000×750, Settings 880×640
  (`gui/main_window/theme.py:199`, `gui/dialogs.py:113`)

## Verification required in the implementation session

1. Confirm Wails stable version + WebView2 availability on Windows 10 (the owner's OS) — `winget
   list Microsoft.EdgeWebView2Runtime` or registry probe; document the bootstrapper fallback.
2. PoC acceptance: cold start < 1.5 s to first frame; log desk sustains ≥ 2,000 rows/s appends with
   filtering; exe size < 20 MB.

## Amendment 1 — Owner Q&A confirmation (2026-10, post-blueprint session)

**Status change: Wails confirmed by the owner** (delegated decision after the WebView2 questionnaire;
Fyne remains the recorded fallback only if the PoC fails, not by owner preference).

Owner context recorded from the Q&A (verbatim signals):

- The owner's Windows 10 has **not received Windows Updates for months** (likely EOL) and **Edge
  was manually deleted** (Teams likely removed too) → **assume WebView2 is NOT installed on the
  owner machine.** A2 revised; R7 likelihood raised M→H — but the mitigation (Evergreen
  bootstrapper, below) is now the *expected* path, not a fallback. WebView2 installs independently
  of Edge and works on EOL Windows 10 (app component, not an OS update).
- Owner **accepts a one-time silent auto-install** of the official Microsoft bootstrapper (Q3) and
  is **always online at install time** (Q6) → Evergreen bootstrapper is the distribution path.
  "Fixed-version" runtime bundling into the zip is rejected (adds ~100 MB to every artifact).
- Owner wants **setup.exe AND portable zip** (Q5) → the installer chains the bootstrapper silently;
  the portable zip does a first-run check with a one-click silent fetch (ADR-0010 Am.1).
- **Cross-platform flexibility (macOS/Linux) is a registered non-blocking requirement** (Q3 note:
  "might need to use it in different OS… good to have"): Wails covers macOS via built-in WKWebView
  (no bootstrap) and Linux via WebKitGTK. v1 stays Windows-only; the `core/`-purity law already
  forbids Windows-only dependencies in `internal/` packages.
- Owner's UI aspiration (Q4): current design is "acceptable, but not the polished and professional
  grade UI I am looking for" → **replicate-now stands** (parity discipline), and the post-parity
  **professional redesign is upgraded from "maybe" to a planned Phase-8 priority**. The web stack
  is chosen partly because it has the highest polish ceiling and the cheapest redesign path.
- **Why not Fyne despite "absolute zero-dependency is tempting" (Q4):** the zero-dep benefit only
  materializes on the owner's Windows box, where the one-time silent install is already accepted
  and handled by the setup.exe; Fyne loses on both remaining axes — it replicates the current
  baseline *less* faithfully (Material drift) *and* has a lower ceiling for the professional-grade
  UI the owner ultimately wants. Decision: Wails.

PoC C (T-70) gains an explicit dimension: verify the silent bootstrapper path end-to-end on a
WebView2-less Windows VM (install → launch → themed window, measured).

## Amendment 2 — PoC verdict (T-70, SESSION-001, 2026-10-15)

**Status: accepted — Wails v2.16.0 (stable v2 line) confirmed by working PoC.**

Evidence (full detail: `docs/poc-artifacts/t-70-wails-poc.md`, screenshot
`t-70-wails-poc-screenshot.png`):

- **Themed window + log desk + one worker job — all proven live** on Linux
  (rootless webkit2gtk-4.1 tree, Xvfb, VLM-verified screenshot): GitCurator
  LIGHT tokens replicated (cream/plum), coalesced log bus with the
  120 ms/250-row/1024-cap numbers, sequential worker job with progress.
- **All three PoC acceptance criteria pass**: cold start → first frame
  **213–301 ms** (gate < 1.5 s); desk sustained a 10k-line flood at
  **625,000 lines/s accepted, 0 drops, UI interactive** (gate ≥ 2,000 rows/s);
  exe size **11.8 MB** Windows / 9.1 MB Linux (gate < 20 MB).
- **Windows cross-compile works from Linux** (plain `wails build -platform
  windows/amd64`, WebView2 loader is pure Go — no CGO).
- **Blocked-with-plan (owner action)**: the silent Evergreen-bootstrapper
  leg on a WebView2-less Windows VM — no Windows environment exists in the
  sandbox. The concrete plan (NSIS installer chaining the bootstrapper,
  the absence probes, the measurement checklist) is recorded in the
  artifact and NEXT-SESSION.md; it does not block Phase 1–5.
- PoC environment note: Fyne fallback remains unnecessary; the PoC found
  no Wails blocker on the Linux leg, and the Windows runtime leg is an
  owner-VM verification, not a library risk.

## Amendment 3 — The Phase-6 entry gate re-run verdict (T-70, SESSION-033, 2026-10-19)

**Status: PASS — the Phase-6 entry gate is GREEN (Wails v2.16.0 confirmed
again, now with ONE REAL JOB wired through the engine).**

The roadmap gates Phase 6's entry on "PoC C: Wails (ADR-0001 gate: themed
window + log desk + one job; startup ≤ 1.5 s; log flood smooth)". Am.2
(SESSION-001) had proven the framework with a synthetic worker loop; the
SESSION-033 mandate re-ran the gate as Phase 6's entry with the substantial
upgrade: **ONE REAL JOB** — the note-state baseline (the
`record_baseline_if_empty` composition, `note_state.py:443-474`) wired
through the `guibridge/` seam into the shell, running the real ScanVault
walk + per-note read + sha256 fingerprint + SQLite upserts over a
deterministic 800-note fixture vault.

Evidence (full table + caveats: `docs/poc-artifacts/t-70-gate.md`): first
frame **266 ms / 679 ms** (two runs — the second after a mid-session
degradation + full toolchain rebuild from the committed recipe; both far
under the 1.5 s gate); the 10k flood at **458,695 / 454,133 lines/s, 0
drops**, the flood's whole DOM-append cost to the main thread **23–25 ms
max per batch**; the real job **800 notes in 94–101 ms**; the Windows exe
cross-compiled CGO-free at **14.95 MB**.

Standing caveats (unchanged in kind from Am.2): the Windows runtime leg +
the silent Evergreen-bootstrapper verification on a WebView2-less VM remain
the owner's (Phase 7's clean-VM checklist); the WM-less-Xvfb timer clamp
makes the task-gap instrument read throttling — the interactivity evidence
is the handler-time instrument + the rendered desk (both recorded).

Decisions riding this verdict (recorded elsewhere, cited here):
- The Wails build-time toolchain + the nested `webui/` module +
  the vanilla frontend step: **ADR-0002 Am.3** (surfaced BEFORE the first
  wails invocation) + its **Am.4** boundary correction (the internal-rule
  claim disproved by probe).
- The seam's retirement law: `guibridge/` retires at T-72 when webui/
  dissolves into the main module.

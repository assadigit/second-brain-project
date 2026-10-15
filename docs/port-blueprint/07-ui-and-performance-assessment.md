# 07 — UI and Performance Assessment

## Part A — Existing GUI architecture (what the Go UI must replicate)

### A.1 Structure

- One main window (1000×750) + one non-modal Settings dialog (880×640, master-detail with 11
  pages + Lucide sidebar) — geometry laws in `gui/main_window/theme.py:199`, `gui/dialogs.py:113`.
- Main window chrome: top bar (logo, **Net/AI/Proxy sentinel labels**, settings + theme buttons) →
  CTA band (**SYNC** 56 px hero button, ⚡ Fast Sync, 🔍 Scan, 🔌 Test Connectivity) → progress row
  (`SegmentBar` dual-batch/result painting, `ui.py:94-173`) → Progress Logs card (filter chips
  all/errors/warnings/success + search + clear; `LogView`) → empty state.
- ~24 overflow actions in the More menu (`ui.py:1849-1946`).
- 10+ modal dialogs: connection test, scan plan (checkbox sections + masters + countdown),
  scan result scorecard (self-closing 12 s), vault picker, sync differences (tabs + chips),
  telegram code/2FA, LLM failure, disk full, batch confirm, quick-model picker, manual resolve,
  recategorize, restore.

### A.2 The complete visual language (already documented in-repo)

- **Theme tokens**: full LIGHT/DARK tables in `gui/theme.py:58-237` — reproduced in
  `module-specs/spec-gui-theme-icons.md` §1 (this is the CSS-variable source of truth for the Go
  UI).
- **Type scale**: settings_title 20/800 · sync_hero 24/800 · sync_cell_num 20/800 · plan_heading
  15/800 · card_heading 14/700 · cc_row_name 13/700 · base 13 · cc_item 12/11 · stat_value 14/700
  · micro-copy 11/10. Fonts: Segoe UI stack; Consolas mono (log); Inter-Bold registered from
  assets.
- **Icons**: Lucide v0.544.0, 30 glyphs, 24×24 viewBox, 2 px strokes, render-tinted, 2× DPR.
- **Buttons**: 8+2 variants (`primary/secondary/danger/ghost/icon/help/hero_*`) behind the
  `btn_kind` property grammar; **exactly one filled primary per view** law.
- **Status grammar**: sync dots (green/grey/red + word), cell washes (match/behind/reviewed/
  unknown), banner tones, pending badge zinc-not-red, pipeline glyphs.
- **Dialog grammar**: msg/question/card-modal families; fades 200 ms; zombie guards.
- **The 10 honesty laws** — `03-behavioral-specification.md` §3.

### A.3 GUI↔backend communication (the contract the Go UI preserves)

- All background work in QThreads; GUI updated **only** via queued signals (8 batch signals +
  job/section/result signals) — `gui/processing_worker.py:101-112`.
- Interactive ask-gates: `code_requested` (300 s), `scan_confirm_requested` (300 s),
  `llm_failed_signal` decision (600 s) — Event + signal round-trips.
- Timers: log flush 120 ms; hero mirror 200 ms; sentinel 60 s; watchdog 60 s; single-shots
  1200/1500/2000/2500/3500 ms at startup; post-save sentinel re-arm 250 ms; post-seal re-measure
  4,000 ms.
- The quiet-open law: nothing network at startup.

**Go mapping**: event bus + commands (doc 04 §2/§5); the Wails frontend subscribes and renders;
all cadences preserved.

## Part B — Performance findings

### B.1 Confirmed historical bottlenecks (fixed in Python — evidence in repo)

These are *proven* because the repo documents both the defect and the fix with measurements:

1. **The log** — was 1.67–1.7 s **per line** past the 1,000-entry cap (full document rebuild per
   line); now 8.7 µs/line with the 120 ms/250-row coalescing (`gui/main_window/ui.py:65-86`,
   CHANGELOG v0.77/v0.80; investigation 1-d).
2. **The sentinel** — a 2 s blocking `connect_ex` on the GUI thread every 60 s froze the window;
   moved to a daemon thread (`gui/main_window/window.py:338-350` comments).
3. **Queue-check classification** — 0.5–5 s GUI freeze per bot-queue check; moved to the worker
   thread (`gui/worker_jobs.py`, investigation 1-d).
4. **Manifest writes** — 2–3 full-file rewrites per link (~1,000 writes per 500-link batch);
   transient state no longer saved (`gui/link_tracker.py:145-157`).
5. **Scan modal master-click** — O(n²) live recount on 250 rows; silenced sweeps
   (`gui/scan_plan_dialog.py:631-649`).

### B.2 Confirmed current costs (by code inspection — labeled CONFIRMED)

1. **Python cold start**: PyQt6 + widget construction + 11 settings pages + fonts ≈ multiple
   seconds before first frame (the code fights this: Telethon import deferred, version stamp
   suppressed in CLI, `--cli` dispatched before GUI import — `app.py:31-38, 155-163`,
   `main.py:43-45`). Go + Wails startup is typically well under 1 s. **CONFIRMED (structural).**
2. **GIL-bound CPU work**: vault walks + note parsing + frontmatter regex + sha256 fingerprints +
   pure-Python cosine similarity over embeddings run on one core; the one-open cache exists
   precisely because re-reading 1,309 notes was slow (`core/note_reads.py`). Go's native
   parallelism and faster string/hash paths address this directly. **CONFIRMED (measurable via
   the Python timing_runner pattern).**
3. **Sequential per-item processing with politeness delays**: 0.5 s inter-item delay (+1.5 s large
   batches), 2 s per-domain delay, banner throttles — *by design* (rate limits). **Not a defect;
   Go must reproduce, not "fix".**
4. **LLM latency dominates wall-clock** (default 300 s timeouts, chunked scan filing). **Not
   fixable by the language.**

### B.3 Hypotheses (unmeasured — need baselines in implementation)

1. **Qt re-render costs** on the Sync view re-renders (the `lay.activate()` and measured-height
   hacks exist because paints beat layouts — `sync_flow.py:476-485`) — hypothesis: HTML rendering
   removes this class entirely.
2. **Settings-page construction at startup** (all 11 pages built upfront, `ui.py:179-2100`) —
   hypothesis: lazy page construction in the Go UI shaves startup further.
3. **SQLite per-statement RLock serialization** — hypothesis: negligible at current data scale;
   do not optimize preemptively.

### B.4 Performance acceptance criteria (measurable)

| Metric | Python baseline (to capture in Phase 0) | Go target |
|---|---|---|
| Cold start → first frame (GUI) | measure (est. 2–5 s) | **≤ 1.5 s** |
| CLI `--status` wall time | measure | ≤ 0.5 s |
| Log line ingest (desk) | 8.7 µs/line (Python, measured in-repo) | ≥ same; UI stays interactive at 1,000-line bursts |
| Full vault scan inventory (1,309 notes) | measure with timing rig | ≤ 50% of Python |
| Batch of 10 websites (fake fetch+LLM) | measure | wall time within +10% (network-bound anyway) |
| Sentinel pass (3 probes) | off-thread already | off-thread, < 250 ms each |
| Memory (idle GUI) | measure (PyQt6 typical 150–300 MB) | ≤ 200 MB |

Baseline procedure for the implementation session (Phase 0): run the Python app with
`QT_QPA_PLATFORM=offscreen` (or on the owner's machine), execute `tests/timing_runner.py`-style
timed cases for: startup, scan inventory, note_reads cache hit-rate, log flood (10k lines),
queue classification (1,000 fake links). Record numbers in the implementation report — do **not**
invent them.

## Part C — GUI framework comparison (summary of ADR-0001)

| Criterion | Wails (primary) | Fyne (fallback) | Qt-for-Go | GTK |
|---|---|---|---|---|
| Windows UX/visual quality | Native webview, full CSS control | Material-ish, visibly different | Qt fidelity | Inconsistent on Win |
| Packaging | exe + WebView2 (bootstrappable) | pure exe | CGO + Qt DLLs | DLL soup |
| Startup | fast | fast | slow-ish | slow |
| Wordwrap/log/tables/tooltips | HTML/CSS (trivial) | partial | full | partial |
| Theme replication from tokens | near-1:1 CSS vars | re-approximation | QSS (but binding is dead) | CSS-ish |
| Maintenance risk | active project | active | **unmaintained** | moderate |
| Accessibility | webview a11y + keyboard | growing | Qt | GTK |

**Recommendation**: Wails, workflow-preserving HTML replica of the documented design; Fyne only if
the WebView2 PoC fails. Full reasoning in `decisions/0001-gui-framework.md`.

## Part D — Migration-phase vs redesign-phase boundary

**In scope for the migration (must match today)**:

- Page set, names, and IA: main window (SYNC/Fast Sync/Scan/Test Connectivity, progress, log desk)
  + Settings dialog with the 11 pages in the same order.
- Every control, state, status dot, badge, chip, banner, countdown, and modal listed in the
  module specs (§GUI group) with the same labels and honesty behaviors.
- Theme token values (light/dark), type scale, Lucide icon set, sounds, window geometry.
- All keyboard behaviors that exist today (Tab focus with visible outlines, Enter submits code
  modal, Esc = safe cancel) — and no new ones.

**Explicitly deferred to a later redesign phase** (recorded, not built):

- Any layout improvements, new screens, information-architecture changes, new visual language.
- Responsive/resizable window redesign (current app is fixed-geometry by design).
- The owner's known irritants that are *design* issues (e.g. "table like one is confusing" was
  already addressed in v0.91.0; further reflows are redesign).
- Multi-vault dashboards, charts, or anything not present in the PyQt6 UI today.

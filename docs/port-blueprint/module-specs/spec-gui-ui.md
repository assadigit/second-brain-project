# Module Spec — `gui/main_window/ui.py`

- **LOC**: 3,032 · **Go home**: `ui/frontend/src/` (chrome + pages) + `internal/app` (log bus)
  · **Status**: rehost (layout + coalescing contract binding)
- **Responsibility**: all widget construction (main window chrome + the 11 Settings pages) and
  the **coalesced log engine** — the app's most performance-sensitive GUI code.

## Main window chrome (179-2100, order)

Inter fonts → window size (QSettings restore, min 760×380, default 900×400) → all 11 settings
pages → top bar (logo lockup; **Net/AI/Proxy sentinel labels**; settings + theme 34×34 buttons,
TabFocus) → `sync_card` CTA band (**SYNC `start_btn` 56 px**, ⚡ Fast Sync, 🔍 Scan, 🔌 Test
Connectivity) → progress row (`SegmentBar` + `progress_count`) → Progress Logs card (chips
all/errors/warnings/success + search + clear; `LogView`; empty state) → `more_btn` QMenu (~24
actions: tests, verify, retry, quarantine, 404, Chrome doors, export, dashboard, dark mode) →
`SettingsDialog(self)` → `_run_mirror_timer` 200 ms → `load_ui_config()` → theme apply.

## Settings pages (build sites)

Credentials (:281) · Proxy (:328) · Vault (:588; combo + pickers + pipeline switches +
blocked/self domains) · LLM (:974; provider radios, limits `llm_num_ctx`/`llm_max_output_tokens`,
engine cards) · Input (:1028) · Dashboard (:1128; stats, quarantine manager + threshold spinbox,
batch undo) · Sync (:1165; **flow cards + gap matrix + ONE Refresh**) · Bot (:1271; username,
token, Check Queue, pending badge, Process All/New, Mark All Read (hidden until verify),
queue_display) · Sources (:1320) · Backup (:1328; built by backup_seal) · Sound (:1431).
Save-on-change (`editingFinished`/`toggled` → page savers, 576-585).

## THE COALESCED LOG (65-86, 2374-2794) — the contract

- `log_message(msg, level)` **never touches the document**: appends `{seq, ts, level, msg}` to a
  pending queue (**8.7 µs/line measured**; the pre-v0.77 law was 1.67–1.7 s/line past the
  1,000-entry cap).
- A single-shot `_log_flush_timer` (interval **120 ms**) lands ≤ **250 rows** in ONE cursor edit
  block with ONE font/metrics pair (Consolas 12) and ONE tail scroll.
- Document mirrors the cache's rendered subset via seq bookkeeping (`_log_retired`,
  `_log_total`, `_log_visible_count`); head-trim retires oldest blocks in one edit
  (`LogView.trim_head`).
- `_LOG_CAP 1000` rendered rows; cache keeps `_all_log_entries` for filter re-renders.
- Row rendering: `_log_row_html` — prefix `[ts]` + level glyph (colored), message elided right
  at `viewport width − 18 (padding+border) − prefix width − 8 (scrollbar+safety)` with the
  full text on the row tooltip; per-row tooltip map re-keyed on trim.
- Filters: chips (all/errors/warnings/success) + search box re-render **from cache** (never
  re-parse the document); clear wipes cache+view+pending; caught-up sync flips the card to the
  empty state.
- `LogView` (`gui/main_window/log_view.py:34-142`): read-only NoWrap QTextEdit; `trim_head(n)`
  retires the oldest n rows in ONE cursor selection (microseconds whatever the pile — the
  pre-v0.77 law rebuilt the whole document per line, 1.7 s/line); per-row tooltip map re-keyed
  on trim; `apply_row_format` 2 px bottom margin; resizeEvent starts the 120 ms re-render
  debounce.
- LogView resize re-render debounced 120 ms (`log_view.py:125-128`).

## SegmentBar (94-173)

Dual-segment progress painting (batch portion + result portion), `bar_saved`/`bar_retry`
terminal colors, mono 15/700 count label; `setFormat` strings: "Fetching undone items…",
"{n} ready to process", "Import file ready", "Ready", "Processing…", "✅ Done", "❌ Failed".

## Misc contracts

- Last-sync stamp: config `last_sync_*` + QSettings `last_sync` = `"stamp|done|total"`
  (`_record_last_sync`/`_restore_last_sync`, 2967+; read on the GUI thread only).
- `_manifest_counts` reads `links_manifest.json` for the pipeline counter.
- Status-quantum: `QSettings`-restored size, min 760×380, default 900×400 (190-200).

## Go parity requirements

1. The log pipeline: producer-side queue → 120 ms flush → bounded 250-row DOM batch appends →
   1,000-row cap with head trimming → filter/search re-render from cache → per-row tooltip =
   full text. **Acceptance: 10,000-line flood with the UI staying interactive.**
2. Same chrome layout/CTA band/progress formats/chip filters/page order/save-on-change.
3. `More` menu action list ported verbatim (labels + behaviors).

## Tests

Log flood performance test (assert UI responsive + cap enforced + filter re-render correct);
page inventory snapshot (all controls present); SegmentBar format strings.

## Dated landing note — T-77 (2026-10-20, SESSION-040)

The main window chrome landed (internal/app/herostate.go — the pure
engines; webui/heroui.go + frontend/heroflow.js — the shell). The
port-level decisions, evidence-settled:

1. **The hero machine is PURE** (hero.py's GUI-only state, the house
   law): the guards cross as DATA (the vault box, the offline ask, the
   nothing-to-sync box, the >10 confirm), the shell executes them. The
   reference's `_confirm_batch` source word is preserved verbatim
   ("the bot queue" — bot_queue.py:721).
2. **The nothing-to-sync box is RE-WORDED for Option B** (ADR-0004
   Am.3): the reference's "enter the bot username + the Telegram API
   credentials" sentence becomes "Set the Worker's URL (or pair the
   app) in Settings → Bot" — the queue's config door under Option B.
   The import-file sentence stays verbatim (the ONE other input mode).
3. **THE RETROACTIVE FIELD-TAG FIX**: this spec's Go parity was built
   on a transport whose payloads crossed tagless from T-71 — the
   progress row NEVER filled during a real batch and every finish read
   ❌ Failed (the frontend's lowercase reads vs the Go field names).
   Every crossed payload now carries the tags (pinned by
   TestCrossedPayloadFieldTags); `needs_retry` joins
   finished_signal's payload (the retry chime's selector).
4. **The More menu's honest-disabled law**: the label list is verbatim
   (parity requirement 3); the engines that have not landed yet (the
   per-test dialogs, the linking tools, the Chrome doors, the export
   zip) render DISABLED with their landing notes — the same convention
   the CTA band rode pre-T-77. The per-test dialogs are NOT T-77's
   (the task named only the battery modal); the re-scoped battery
   carries the Worker leg (ADR-0004 Am.3's queue-reachability).
5. **The caught-up path's Chrome-door legs are deferred** (recorded
   divergence): the reference's caught-up sync also scrapes the 🖐 hand
   rows and redoes half-stored hand notes (hero.py:280-325 — the
   reference itself guards them as optional collaborators); the port
   lands the waiting-rows report + the drift seal + the closing line
   now, the hand-row legs ride the Chrome-doors GUI task.
6. **The sentinel paint lives in Go** (window.py's `_apply_*` laws as
   pure functions) — the frontend only renders; the paint vocabulary
   (incl. "Proxy: off / connected / unreachable") is verbatim.
7. **The sounds ship embedded, disk-first**: the reference's
   synthesized WAVs (license-free) ride the assets package; an owner's
   APP_DIR/assets/sounds copy always wins (the swap-the-chime law).

# Module Spec — Dialogs: `gui/dialogs.py`, `gui/scan_plan_dialog.py`, `gui/scan_result_dialog.py`, `gui/vault_pick_dialog.py`

- **LOC**: 406 + 930 + 618 + 420 · **Go home**: `ui/frontend/src/dialogs/*` · **Status**: rehost
  (workflow contracts below are binding; Qt mechanics replaced)
- **Responsibility**: the Settings master-detail shell + connection-test modal; the scan plan
  confirm door; the applied scorecard; the vault picker.

## dialogs.py

- `SettingsDialog`: **non-modal**, fixed 880×640, objectName `settings_dialog`; nav QListWidget
  190 px (Lucide icons + tooltips + hint) + QStackedWidget of the 11 pages; header re-parents
  the main More menu; TAB_INFO order: Credentials, Proxy, Vault, LLM, Input, Dashboard, Sync,
  Bot, Sources, Backup, Sound.
- `ConnectionTestDialog`: modal, min 560; 4 rows (📁 Vaults / 🧠 LLM / 🐙 GitHub / ✈️ Telegram),
  braille spinner 100 ms; per-row status + indented elided check lines (elide via the label's
  own polished font metrics, full detail on tooltip); **Start Syncing born disabled, enables
  only when all four verdicts ok**; on accept → `singleShot(0, main._on_hero_clicked)`.

## scan_plan_dialog.py (v0.62/0.73/0.77/0.80/0.82 layers)

- Pure `plan_display_model(plan)` (160-268) + `ScanPlanDialog` renderer.
- Sections: deletions (pale rose, 🗑️) / consolidations (teal, 🧲) / new folders (mint, 🌱) /
  moves (neutral, 🚚). Each: heading + **master "All" checkbox** + scrollable checkbox rows
  (`plan_row` 11 px, ONE line elided middle, full story on tooltip; `low_confidence` rows ride
  **PRE-UNCHECKED**).
- Scroll law: fixed height `min(sizeHint+4, 200)` px (the NO-FEEDBACK SCROLL LAW — never
  AdjustToContents+maximumHeight); `MAX_ROWS 500` freeze guard (counted-not-rendered beyond).
- Master sweep with `blockSignals(True)` + ONE recount (O(n)); Apply label live
  "🗂️ Apply plan (12 of 17)"; enable law: armed AND (picked>0 OR nothing to pick).
- **THE FRESH-CLICK DOOR**: app-level eventFilter swallows all input aimed at the dialog for
  **350 ms** after show; both verdict buttons born disabled + `setAutoDefault(False)`;
  `_on_keep/_on_apply` refuse unarmed answers; `done()` retires guard + countdown timer.
- Countdown 300 s (config `scan_confirm_timeout_s`) ticking "no answer in 300s = nothing
  done"; timeout ⇒ VERDICT_TIMEOUT + reject (safe defer); Esc/✕ = declined; selection (4
  section key sets) rides ONLY on 'confirmed'.
- Move rows speak the landing name "(as Same_v1.md)" (THE LANDABLE MOVE).
- `ask_scan_plan(plan, parent)` round shape.

## scan_result_dialog.py

- Pure `result_display_model(result)` + renderer; **landed counts, never the plan's promises**:
  `K of N` rows in warning tone when short; failed moves NAME files (cap 3); per-vault sections
  only when that vault changed; headings differentiate healed-only / applied-only / applied;
  footnote teaches recoverability; self-close **12 s** (`AUTO_CLOSE_S`), timer stopped on
  `finished`; OK primary.

## vault_pick_dialog.py

- Pure `detect_scan_vaults(config, with_counts)` (connection truth + honest note counts; count
  failures = "speaks no number, never a lie") + `pick_display_model` + renderer.
- Three rows (🌐 Websites / 🐙 GitHub Repos / ✍️ Personal); the two machine vaults tickable
  (**nothing pre-checked**); Personal shown read-only with the owner-owned explanation; live
  Scan-button arithmetic ("Scan this vault" / "Scan both vaults"); Cancel/Esc/✕ = safe cancel
  (returns None).

## Go parity requirements

1. Same modality, sizes, sections, row grammars, pre-check states, enable laws, countdowns,
   arm-guard concept (in web terms: ignore clicks for 350 ms after open), and safe-cancel laws.
2. Elide-middle single-line rows with full tooltip story.
3. Scorecard self-close 12 s with the same heading set.

## Tests (GUI workflow)

Picker: safe-cancel, nothing-pre-checked, live arithmetic. Plan: master sweep counts, arm-guard
(timed), timeout verdict, selection-only-on-confirm, low-confidence pre-uncheck, bent-name
display. Scorecard: K-of-N tones, named failures, self-close. Connection dialog: all-ok gate.

> **Dated note (2026-10-20, T-76/SESSION-039 — the landing record).**
> The three dialogs landed as two slices: the pure engines in
> `internal/app/scandialogs.go` (PickDisplayModelOf / PlanDisplayModelOf
> / **PlanPickState** — the picking semantics as a pure state machine,
> so the masters + the arm-guard's enable law + the four-key selection
> are main-battery-testable — / ResultDisplayModelOf) and the shell
> (webui/scanui.go the fifth bound struct + frontend/scanflow.js the
> three modals' renderers). Port-level decisions, all verified against
> the reference at 98f4fc7:
> 1. **The elide-middle twin**: Qt's pixel-metric ElideMiddle becomes a
>    char-budget `elideMiddle` (the full story on the tooltip) — the
>    browser has no middle-elide primitive; the one-line law + the
>    tooltip law are the parity.
> 2. **The repos-only double-section quirk PRESERVED**: the reference's
>    out dict carries `out['repos']` unconditionally and the finished
>    routing hands the raw dict to the scorecard, so a repos-only scan
>    renders the promoted section AND the repos sub-story section BOTH
>    (verified by running the reference's `result_display_model` over
>    the repos-only shape; pinned by the port's test with the probe
>    citation). A visual dedupe would be a behavior change — the
>    five-fix law forbids it here.
> 3. **The scan story extension**: the T-71 ScanResult grew the
>    reference's out-dict shape (Story/Repos/VaultsRan/Hygiene + the
>    vault seats; the planned counts speak PICKED units once a
>    confirmed selection rides). The never-read `AppliedCounts` field
>    was replaced by the real shape.
> 4. **The Scan CTA rides its own machine** (ui.py's scan_btn is wired
>    in vault_scan_ui.py, not hero.py): the button is live from T-76 —
>    the picker → the plan modal over the bus gate → the scorecard; the
>    SYNC / Fast Sync / Test Connectivity buttons stay T-77's hero
>    machine.
> 5. **The answering windows**: the modal's countdown ticks the
>    configured `scan_confirm_timeout_s` (the ask payload carries it);
>    the bus's own 300 s window is the belt under it (a dead modal
>    still defers safely). A configured window LONGER than 300 s would
>    be cut by the bus's belt — an edge the reference does not have
>    (its worker waits the same config value); recorded as a known
>    divergence for the owner's call.
> 6. **The evidence doors**: ProbeAskSample / ProbeScorecardSample
>    (canned fixture shapes, no owner data) + the SB_POC_OPEN_SCAN
>    probes (picker / plan / scorecard / flow) — the T-70 law
>    (screenshot-provable) applied to the modals.
> The ConnectionTestDialog's own row stays T-77's (the hero machine's
> "Start Syncing" door).

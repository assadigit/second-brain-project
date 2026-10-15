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

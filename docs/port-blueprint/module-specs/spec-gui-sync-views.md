# Module Spec — Sync views: `gui/main_window/sync_flow.py`, `gui/main_window/sync_view.py`, `gui/sync_display.py`

- **LOC**: 558 + 1,453 + 679 · **Go home**: pure halves → `internal/app` (display model +
  snapshot gather); rendering → `ui/frontend/src/sync/*` · **Status**: split port/rehost
- **Responsibility**: the Settings → Sync page — THE VISUAL SYNC DISPLAY (four-element flow
  cards) above THE DEFERRED LEDGER (6-cell gap matrix + differences modal), one Refresh for
  both. Paint + forward only; every decision in `core/sync_review`.

## sync_display.py (PURE — ports as Go logic)

- `ELEMENT_CHROME`: 🤖 Telegram Bot "source of input" / 📓 Obsidian Vault "local backup" /
  ☁️ GitHub Backups "cloud backup of Obsidian" / 🌐 Website Directory "main display and
  front-end" (the owner's own numbering; the Z-reading IS the pipeline).
- States `ok/not_set/error/never`; tone map; legend text.
- `classify_pending(pending)` (130-154): URL-router split repos/websites/unrouted;
  `repos+websites+unrouted == total` always.
- `collect_local_facts` (238-255): picker's vault walk reused; per-CONFIGURED-backup-repo rows
  with VaultSeal.status() git truth (unconfigured repo = silence — no row, no error); the
  Websites vault's public-ready count via the publish leg's own parser; read-only.
- `sync_display_model` (648-679): four cards with hero number + micro-copy label + stat rows +
  foot lines + hint; **the §5 arithmetic**: `V>P → "V−P need publishing"`; `V<P → "−N since the
  last publish (banished notes retire too)"` (never negative needs-publishing);
  never-published = "never"; unreachable worker degrades to unknown (never a fabricated count);
  calm line only when `waiting==0 AND any_measured`.
- `friendly_when`: "today 14:02" / date-time / raw / "never".

## sync_flow.py (the flow cards' Qt layer → rehost)

- `_SyncPage` emits `first_show` exactly once per session (the ONE automatic refresh).
- 2×2 grid: [bot] → [vault] / [github] → [website]; arrows are QLabel "→" with meaning
  tooltips; cards QFrame `sync_card` max width **300** (grid arithmetic documented at :53-58).
- Card interior (v0.91.0 THE HERO LAYOUT): head row (sync_dot + name) → **hero row**
  (`sync_hero` 24/800 + `sync_hero_label` 11 wordwrapped, height MEASURED via the label's own
  polished font metrics — the height-for-width law) → stat rows (label/value + tone + note) →
  foot lines → hint.
- Render states: placeholder ("not measured yet", grey dots, no fabricated counts), measuring
  ("· · ·" hero), error (cards KEEP last truth; note `measurement failed — {msg[:120]}`).
- Second click while measuring swallowed; wrong-test-name worker answers ignored; re-render ends
  with `lay.activate()` (deterministic close) — **web equivalent: ensure layout settled before
  paint** (CSS handles this natively; keep the semantic: no stacked generations).
- Refresh row: primary button + `sync_measured_note` ("measured at HH:MM:SS — the numbers are as
  fresh as this line"); `last_sync` QSettings `stamp|done|total`.

## sync_view.py (the matrix + modal → rehost)

- `_STAGE_DEFS`: 6 cells — repos/websites Queue→Vault (fix=check_queue), websites Vault→Website
  (fix=publish), repos/websites/personal Vault→Backup (fix=seal); `None` slots = muted "—".
- Matrix cells = QPushButton `sync_cell` (min 56×190) with `sync_cell_num` 20/800 +
  `sync_cell_icon` + wordwrapped caption (micro-copy law: every number is NAMED under itself);
  states via `cell_state` property washes (match=mint ✓ / behind=amber ⚠ clickable /
  reviewed=calm ⓘ clickable / unknown=dashed "—" with its own reason on the cell).
- Banner `sync_banner` + `banner_tone` + icon; worst-first verdict; legend; last-measured line.
- `_gather_sync_snapshot(config, queue_state, …)` runs INSIDE the worker (sentinel pattern):
  per-stage records incl. `reason`/`needs` ('setup' vs 'measure' — v0.91.0 THE NAMED VERDICT).
- `SyncDifferencesDialog`: tabs "Differences (n)" / "Deferred log (n)"; oldest-first checkbox
  rows (44 px, elide 320, "3 d ago" column 84, deferred chip, Open note button); Select all;
  four reason chips (44 px, exclusive) + note input; **"Mark as intentional" disabled until ≥1
  item AND a chip**; ONE inline confirmation line (never a second dialog); log tab = 6-column
  read-only table (search + Reopen on Active rows, cap 500) + the "nothing is ever deleted"
  note.
- Fix doors ride existing patterns (check queue / publish worker / seal workers); repos seal
  re-measure armed `singleShot(4000, refresh)`; quiet-open law (nav landing or Refresh triggers);
  one-pass-at-a-time worker guard; `_sync_recompute_from_ledger` instant repaint after modal
  changes (no network).

## Tests

Port `test_syncdisplay` (65): classify split; §5 arithmetic edges (shrink, never-published,
unreachable); honest read doors (get_pending_result three truths; fetch_published_index four);
first-show-once; consented worker; never-stacking re-render. Port `test_syncreview`-driven UX
cases: cell paint states, banner precedence, modal disabled-until, reopen flows.
